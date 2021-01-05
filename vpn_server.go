package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xtaci/smux"
	go_tuntap "github.com/yzslab/go-tuntap"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type VPNServer struct {
	HasState
	session              net.Conn
	controlMessageStream io.ReadWriteCloser
	vni                  go_tuntap.VirtualNetworkInterface
	clientConfig         *ClientConfig
	smuxSession          *smux.Session
	status               VPNClientStatus
	isVNIPersistent      bool
	serverExpectMTU      go_tuntap.VirtualNetworkInterfaceMTU
}

func NewVPNServer(session net.Conn, clientConfig *ClientConfig, isVNIPersistent bool) (*VPNServer, error) {
	v := &VPNServer{
		session:         session,
		clientConfig:    clientConfig,
		isVNIPersistent: isVNIPersistent,
	}
	v.setNextState(v.initialState)
	return v, nil
}

func (v *VPNServer) initialState(arguments interface{}) (interface{}, error) {
	v.setNextState(v.createSMux)
	return nil, nil
}

func (v *VPNServer) createSMux(arguments interface{}) (interface{}, error) {
	v.log("creating smux session")
	mux, err := createSMux(v.session, v.clientConfig.GetSMuxBufferSize(), v.clientConfig.GetKeepaliveInterval())
	if err != nil {
		return nil, err
	}
	v.smuxSession = mux
	v.log("smux session created")
	v.setNextState(v.openControlMessageStream)
	return nil, nil
}

func (v *VPNServer) openControlMessageStream(arguments interface{}) (interface{}, error) {
	v.log("opening control message stream")
	controlMessageStream, err := v.smuxSession.Open()
	if err != nil {
		return nil, err
	}
	v.controlMessageStream = controlMessageStream
	v.log("control message stream opened")
	v.setNextState(v.sendClientInformation)
	return nil, nil
}

func (v *VPNServer) sendClientInformation(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 32)
	bufferPosition := 0

	// message type 1 byte
	buffer[bufferPosition] = uint8(MessageClientInformation)
	bufferPosition++

	// client id length and client id value 17 bytes max
	buffer[bufferPosition] = v.clientConfig.ClientIdLength
	bufferPosition++
	copy(buffer[bufferPosition:], v.clientConfig.ClientId)
	bufferPosition += int(v.clientConfig.ClientIdLength)

	// mtu 2 bytes
	binary.LittleEndian.PutUint16(buffer[bufferPosition:], uint16(v.clientConfig.GetVNIMTU()))
	bufferPosition += 2

	// client ip mode 1 byte
	buffer[bufferPosition] = uint8(v.clientConfig.ClientIPMode)
	bufferPosition++
	switch v.clientConfig.ClientIPMode {
	case ClientIPModeServerAssign:
		v.setNextState(v.receiveIPAssignment)
		break
	case ClientIPModeClientSet:
		v.setNextState(v.receiveIPValidation)
		break
	case ClientIPModeOther:
		v.setNextState(v.receiveServerConfiguration)
	default:
		log.Fatalf("unknown client ip mode: %d", v.clientConfig.ClientIPMode)
	}

	// local ip 4 bytes
	binary.LittleEndian.PutUint32(buffer[bufferPosition:], v.clientConfig.GetLocalIP())

	v.log("sending client information")
	defer v.log("client information sent")
	return writeControlMessage(v.controlMessageStream, buffer)
}

func (v *VPNServer) receiveIPAssignment(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 5+ServerConfigurationSize+2)
	v.log("receiving ip assignment")
	_, controlMessage, err := readControlMessageWithProvidedBuffer(v.controlMessageStream, buffer)
	if err != nil {
		return nil, err
	}
	if controlMessage[0] != uint8(MessageServerAssignedIPValue) {
		return nil, errors.New("receiveIPAssignment(): not a ip assignment message")
	}
	v.clientConfig.SetLocalIP(binary.LittleEndian.Uint32(controlMessage[1:]))
	v.log(fmt.Sprintf("ip assignment received: %s", long2ip(v.clientConfig.GetLocalIP())))
	v.setNextState(v.retrieveServerConfiguration)
	return controlMessage[5:], nil
}

func (v *VPNServer) receiveIPValidation(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 1+ServerConfigurationSize+2)
	_, controlMessage, err := readControlMessageWithProvidedBuffer(v.controlMessageStream, buffer)
	if err != nil {
		return nil, err
	}
	if controlMessage[0] != uint8(MessageClientSetIPValueAccepted) {
		return nil, errors.New("receiveIPValidation(): client set ip value not accepted")
	}
	v.log("client set ip value accepted")
	v.setNextState(v.retrieveServerConfiguration)
	return controlMessage[1:], nil
}

func (v *VPNServer) receiveServerConfiguration(arguments interface{}) (interface{}, error) {
	_, controlMessage, err := readControlMessage(v.controlMessageStream, uint(ServerConfigurationSize+1))
	if err != nil {
		return nil, err
	}
	if controlMessage[0] != uint8(MessageServerConfiguration) {
		return nil, errors.New("receiveServerConfiguration(): not a server configuration message")
	}
	v.setNextState(v.retrieveServerConfiguration)
	return controlMessage[1:], nil
}

func (v *VPNServer) retrieveServerConfiguration(arguments interface{}) (interface{}, error) {
	buffer, ok := arguments.([]byte)
	if ok == false {
		return nil, errors.New("retrieveServerConfiguration(): arguments not a type of []byte")
	}
	v.clientConfig.SetVNIMode(go_tuntap.VirtualNetworkInterfaceMode(buffer[0]))
	v.log(fmt.Sprintf("vni mode: %d", v.clientConfig.GetVNIMode()))
	v.serverExpectMTU = go_tuntap.VirtualNetworkInterfaceMTU(binary.LittleEndian.Uint16(buffer[1:]))
	v.log(fmt.Sprintf("server expect mtu: %d", v.serverExpectMTU))
	v.clientConfig.FullFrameMTU = go_tuntap.VirtualNetworkInterfaceMTU(binary.LittleEndian.Uint16(buffer[3:]))
	v.log(fmt.Sprintf("full frame mtu: %d", v.clientConfig.FullFrameMTU))
	v.clientConfig.SetPeerIP(binary.LittleEndian.Uint32(buffer[5:]))
	v.log(fmt.Sprintf("peer ip: %s", long2ip(v.clientConfig.GetPeerIP())))

	netmask := binary.LittleEndian.Uint32(buffer[9:])
	if (v.clientConfig.ClientIPMode != ClientIPModeOther || v.clientConfig.GetIP4Netmask() == 0) && netmask > 0 {
		v.clientConfig.SetIP4Netmask(netmask)
		v.log(fmt.Sprintf("netmask: %s", long2ip(v.clientConfig.GetIP4Netmask())))
	}
	v.setNextState(v.sendClientReady)
	return nil, nil
}

func (v *VPNServer) sendClientReady(arguments interface{}) (interface{}, error) {
	v.setNextState(v.receiveServerReady)
	defer v.log("client ready sent")
	v.log("sending client ready")
	return writeControlMessage(v.controlMessageStream, []byte{
		uint8(MessageClientReady),
	})
}

func (v *VPNServer) receiveServerReady(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 8)
	v.log("receiving server ready")
	_, controlMessage, err := readControlMessageWithProvidedBuffer(v.controlMessageStream, buffer)
	if err != nil {
		return nil, err
	}
	if controlMessage[0] != uint8(MessageServerReady) {
		return nil, errors.New("receiveClientReady(): not a server ready message")
	}
	v.log("server ready received")
	v.setNextState(v.createVNI)
	return nil, nil
}

func (v *VPNServer) createVNI(arguments interface{}) (interface{}, error) {
	// use smallest mtu
	mtu := v.serverExpectMTU
	if v.clientConfig.GetVNIMTU() > 0 && v.clientConfig.GetVNIMTU() < mtu {
		mtu = v.clientConfig.GetVNIMTU()
	}
	v.log(fmt.Sprintf("vni mtu: %d", mtu))

	bridgeVNI := false
	localIP := v.clientConfig.GetLocalIP()
	if v.clientConfig.GetVNIMode() == go_tuntap.TAP && v.clientConfig.BRCtl4Go != nil {
		bridgeVNI = true
		localIP = 0
	}

	vni, err := CreateVNI(&VNIConfig{
		Mode:         v.clientConfig.GetVNIMode(),
		Name:         v.clientConfig.GetVNIName(),
		MTU:          mtu,
		LocalIP:      localIP,
		Netmask:      v.clientConfig.GetIP4Netmask(),
		PeerIP:       v.clientConfig.GetPeerIP(),
		IsPersistent: v.isVNIPersistent,
	}, v.log)
	if err != nil {
		return nil, err
	}
	v.vni = vni

	if bridgeVNI {
		v.log(fmt.Sprintf("add interface %s to bridge", vni.GetName()))
		err := v.clientConfig.BRCtl4Go.AddInterface(v.vni.GetName())
		if err != nil {
			return nil, err
		}
	}

	v.setNextState(v.handle)
	return nil, nil
}

func (v *VPNServer) handle(parameters interface{}) (interface{}, error) {
	defer v.log("handler ended")
	v.setNextState(nil)
	// defer v.Close()

	vniStream, err := v.smuxSession.Open()
	if err != nil {
		return nil, err
	}
	defer vniStream.Close()

	if v.clientConfig.OnConnectedHook != "" {
		v.log(fmt.Sprintf("hook: %s", v.clientConfig.OnConnectedHook))
		cmd := exec.Command(v.clientConfig.OnConnectedHook)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("KV_CLIENT_ID=%s", v.clientConfig.ClientId),
			fmt.Sprintf("KV_VNI_INTERFACE_NAME=%s", v.clientConfig.VNIName),
			fmt.Sprintf("KV_CLIENT_IP=%s", long2ip(v.clientConfig.LocalIP)),
			fmt.Sprintf("KV_CLIENT_IP_MODE=%d", v.clientConfig.ClientIPMode),
			fmt.Sprintf("KV_REMOTE_ADDR=%s", long2ip(v.clientConfig.PeerIP)),
		)
		err := cmd.Run()
		if err != nil {
			v.log(err.Error())
		}
	}

	v.log("net<--->vni exchange handler starting")
	exchangeEvent := startReadWriterExchange(vniStream, v.vni, uint(v.clientConfig.FullFrameMTU), v.log)
	v.log("net<--->vni exchanging...")

	v.log("control message handler starting")
	controlMessageEvent := v.startControlMessageHandler()
	v.log("control message handler started")

	sigs := make(chan os.Signal)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	v.log("waiting for handler's events...")
	select {
	case <-sigs:
		v.clientConfig.AutoReconnect = false
		break
	case <-exchangeEvent:
	case <-controlMessageEvent:
	}
	signal.Stop(sigs)
	signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	return nil, nil
}

func (v *VPNServer) startControlMessageHandler() chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer func() {
			close(ch)
			v.log("control message handler ended")
		}()
		controlMessageStream := v.controlMessageStream
		controlMessageBuffer := make([]byte, ControlMessageBufferSize)
		var err error
		// var n int
		// var controlMessage []byte

		for {
			// TODO: implement control message
			_, _, err = readControlMessageWithProvidedBuffer(controlMessageStream, controlMessageBuffer)
			if err != nil {
				return
			}
		}
	}()
	return ch
}

func (v *VPNServer) Close() error {
	if v.controlMessageStream != nil {
		_ = v.controlMessageStream.Close()
		v.controlMessageStream = nil
		v.log("control message stream closed")
	}
	if v.smuxSession != nil {
		_ = v.smuxSession.Close()
		v.smuxSession = nil
		v.log("smux session closed")
	}
	if v.session != nil {
		_ = v.session.Close()
		v.session = nil
		v.log("net connection closed")
	}
	if v.vni != nil {
		_ = v.vni.Close()
		v.vni = nil
		v.log("vni closed")
	}
	return nil
}

func (v *VPNServer) log(message string) {
	SendLog(message)
}
