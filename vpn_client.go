package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xtaci/smux"
	"github.com/yzslab/go-tuntap"
	"io"
	"log"
	"net"
)

type VPNClientStatus uint8

const (
	VPN_CLIENT_STATUS_HANDSHAKING VPNClientStatus = iota
	VPN_CLIENT_STATUS_NORMAL
	VPN_CLIENT_STATUS_CLOSED
)

type VPNClient struct {
	HasState
	session              net.Conn
	controlMessageStream io.ReadWriteCloser
	vniStream            io.ReadWriteCloser
	vniName              string
	serverConfig         *ServerConfig
	vni                  go_tuntap.VirtualNetworkInterface
	smuxSession          *smux.Session
	state                VPNClientStatus
	peerIP               uint32
	isVNIPersistent      bool
	stop                 chan struct{}
	remoteAddr           string
	clientIdLength       uint8
	clientId             string
	clientExpectMTU      go_tuntap.VirtualNetworkInterfaceMTU
	ClientIPMode
}

func NewVPNClient(session net.Conn, vniName string, serverConfig *ServerConfig) (*VPNClient, error) {
	vpnClient := &VPNClient{
		session:      session,
		vniName:      vniName,
		serverConfig: serverConfig,
		state:        VPN_CLIENT_STATUS_HANDSHAKING,
		remoteAddr:   session.RemoteAddr().String(),
	}
	vpnClient.setNextState(vpnClient.initialState)
	return vpnClient, nil
}

func (c *VPNClient) initialState(arguments interface{}) (interface{}, error) {
	c.setNextState(c.createSMux)
	return nil, nil
}

func (c *VPNClient) createSMux(arguments interface{}) (interface{}, error) {
	c.log("creating smux session")
	mux, err := createSMux(c.session, c.serverConfig.GetSMuxBufferSize(), c.serverConfig.GetKeepaliveInterval())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c.smuxSession = mux
	c.log("smux session created")
	c.setNextState(c.acceptControlMessageStream)
	return nil, nil
}

func (c *VPNClient) acceptControlMessageStream(arguments interface{}) (interface{}, error) {
	c.log("accepting control message stream")
	controlMessageStream, err := c.smuxSession.Accept()
	if err != nil {
		return nil, err
	}
	c.controlMessageStream = controlMessageStream
	c.log("control message stream accepted")
	c.setNextState(c.receiveClientInformation)
	return nil, nil
}

func (c *VPNClient) receiveClientInformation(arguments interface{}) (interface{}, error) {
	_, controlMessage, err := readControlMessage(c.controlMessageStream, 32)
	if err != nil {
		return nil, err
	}
	c.log("retrieving client information")

	bufferPosition := 0
	if messageType := MessageType(controlMessage[bufferPosition]); messageType != MessageClientInformation {
		return nil, errors.New("receiveClientInformation(): not a client information message")
	}
	bufferPosition++

	c.clientIdLength = controlMessage[bufferPosition]
	if c.clientIdLength > 16 {
		return nil, errors.New("the length of client id must not > 16")
	}
	bufferPosition++

	if err := c.validateClientIdValue(c.clientIdLength, controlMessage[bufferPosition:]); err != nil {
		return nil, err
	}

	c.clientId = string(controlMessage[bufferPosition : bufferPosition+int(c.clientIdLength)])
	bufferPosition += int(c.clientIdLength)
	if c.clientIdLength > 0 {
		c.log(fmt.Sprintf("client id: %s", c.clientId))
	}

	c.clientExpectMTU = go_tuntap.VirtualNetworkInterfaceMTU(binary.LittleEndian.Uint16(controlMessage[bufferPosition:]))
	bufferPosition += 2
	c.log(fmt.Sprintf("clien expect mtu: %d", c.clientExpectMTU))

	c.ClientIPMode = ClientIPMode(controlMessage[bufferPosition])
	bufferPosition++

	switch c.ClientIPMode {
	case ClientIPModeClientSet:
		c.log("mode ClientIPModeClientSet")
		c.setNextState(c.validateClientSetIPValue)
		return controlMessage[bufferPosition:], nil
	case ClientIPModeServerAssign:
		c.log("mode ClientIPModeServerAssign")
		c.setNextState(c.assignIP)
		break
	case ClientIPModeOther:
		c.log("mode ClientIPModeOther")
		c.setNextState(c.sendServerConfiguration)
		break
	default:
		return nil, errors.New(fmt.Sprintf("handleClientIPMode(): unsupported client ip mode: %d", c.ClientIPMode))
	}
	return nil, nil
}

func (c *VPNClient) validateClientIdValue(length uint8, buffer []byte) error {
	for i := uint8(0); i < length; i++ {
		if !(buffer[i] > 'a' && buffer[i] < 'z' || buffer[i] > 'A' && buffer[i] < 'Z' || buffer[i] > '0' || buffer[i] < '9' || buffer[i] == '_' || buffer[i] == '-') {
			return errors.New("only a-zA-Z0-9_- is allowed in client id")
		}
	}
	return nil
}

func (c *VPNClient) validateClientSetIPValue(arguments interface{}) (interface{}, error) {
	buffer := arguments.([]byte)
	clientSetIPValue, err := retrieveClientSetIPValue(buffer)
	if err != nil {
		return nil, err
	}
	c.log(fmt.Sprintf("client set ip: %s", long2ip(clientSetIPValue)))
	if c.serverConfig.GetIP4AM().AssignSpecificIP(clientSetIPValue) == false {
		c.setNextState(c.sendClientSetIPValueRejected)
		return nil, nil
	}
	c.peerIP = clientSetIPValue
	c.setNextState(c.sendClientSetIPValueAccepted)
	return nil, nil
}

func (c *VPNClient) assignIP(arguments interface{}) (interface{}, error) {
	assignedIP := c.serverConfig.GetIP4AM().Assign()
	if assignedIP < 0 {
		c.setNextState(c.sendServerAssignableIPExhausted)
		return nil, nil
	}
	c.peerIP = uint32(assignedIP)
	c.log(fmt.Sprintf("server assigned ip: %s", long2ip(c.peerIP)))
	c.setNextState(c.sendServerAssignedIPValue)
	return nil, nil
}

func retrieveClientSetIPValue(message []byte) (uint32, error) {
	return binary.LittleEndian.Uint32(message), nil
}

func (c *VPNClient) sendServerAssignedIPValue(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 5+ServerConfigurationSize)
	buffer[0] = uint8(MessageServerAssignedIPValue)
	binary.LittleEndian.PutUint32(buffer[1:], c.peerIP)
	err := c.putVPNConfiguration(buffer[5:])
	if err != nil {
		return nil, err
	}
	c.log("sending server assigned ip value")
	_, err = writeControlMessage(c.controlMessageStream, buffer)
	c.log("server assigned ip sent")
	if err != nil {
		return nil, err
	}
	c.setNextState(c.receiveClientReady)
	return nil, nil
}

func (c *VPNClient) sendServerAssignableIPExhausted(arguments interface{}) (interface{}, error) {
	c.setNextState(nil)
	var err error
	buffer := make([]byte, 1)
	buffer[0] = uint8(MessageServerAssignableIPExhausted)
	c.log("sending server assignable ip exhausted value")
	_, err = writeControlMessage(c.controlMessageStream, buffer)
	c.log("server assignable ip exhausted sent")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *VPNClient) sendClientSetIPValueAccepted(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 1+ServerConfigurationSize)
	buffer[0] = uint8(MessageClientSetIPValueAccepted)
	err := c.putVPNConfiguration(buffer[1:])
	if err != nil {
		return nil, err
	}
	c.log("sending client set ip value accepted")
	_, err = writeControlMessage(c.controlMessageStream, buffer)
	c.log("client set ip value accepted sent")
	if err != nil {
		return nil, err
	}
	c.setNextState(c.receiveClientReady)
	return nil, nil
}

func (c *VPNClient) sendClientSetIPValueRejected(arguments interface{}) (interface{}, error) {
	c.setNextState(nil)
	var err error
	buffer := make([]byte, 1)
	buffer[0] = uint8(MessageClientSetIPValueRejected)
	c.log("sending client set ip value rejected")
	_, err = writeControlMessage(c.controlMessageStream, buffer)
	c.log("client set ip value rejected sent")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *VPNClient) sendServerConfiguration(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, ServerConfigurationSize+1)
	buffer[0] = uint8(MessageServerConfiguration)
	c.putVPNConfiguration(buffer[1:])
	c.setNextState(c.receiveClientReady)
	return writeControlMessage(c.controlMessageStream, buffer)
}

func (c *VPNClient) putVPNConfiguration(buffer []byte) error {
	buffer[0] = uint8(c.serverConfig.GetVNIMode())
	binary.LittleEndian.PutUint16(buffer[1:], uint16(c.serverConfig.GetVNIMTU()))
	binary.LittleEndian.PutUint16(buffer[3:], uint16(c.serverConfig.FullFrameMTU))
	binary.LittleEndian.PutUint32(buffer[5:], c.serverConfig.GetLocalIP())
	binary.LittleEndian.PutUint32(buffer[9:], c.serverConfig.GetIP4Netmask())
	return nil
}

func (c *VPNClient) receiveClientReady(arguments interface{}) (interface{}, error) {
	buffer := make([]byte, 8)
	c.log("receiving client ready")
	_, controlMessage, err := readControlMessageWithProvidedBuffer(c.controlMessageStream, buffer)
	if err != nil {
		return nil, err
	}
	if controlMessage[0] != uint8(MessageClientReady) {
		return nil, errors.New("receiveClientReady(): not a client ready message")
	}
	c.log("client ready received")
	c.setNextState(c.sendServerReady)
	return nil, nil
}

func (c *VPNClient) sendServerReady(arguments interface{}) (interface{}, error) {
	c.log("sending server ready")
	defer func() {
		c.log("server ready sent")
	}()
	c.setNextState(c.createVNI)
	return writeControlMessage(c.controlMessageStream, []byte{
		uint8(MessageServerReady),
	})
}

func (c *VPNClient) createVNI(arguments interface{}) (interface{}, error) {
	// use smallest mtu
	mtu := c.serverConfig.GetVNIMTU()
	if c.clientExpectMTU > 0 && c.clientExpectMTU < c.serverConfig.GetVNIMTU() {
		mtu = c.clientExpectMTU
	}
	c.log(fmt.Sprintf("vni mtu: %d", mtu))

	c.log(fmt.Sprintf("vni name: %s", c.vniName))
	localIP := c.serverConfig.GetLocalIP()
	// do not set local ip in tap mode
	if c.serverConfig.GetVNIMode() == go_tuntap.TAP {
		localIP = 0
	}
	vni, err := CreateVNI(&VNIConfig{
		Mode:         c.serverConfig.GetVNIMode(),
		Name:         c.vniName,
		MTU:          mtu,
		LocalIP:      localIP,
		PeerIP:       c.peerIP,
		IsPersistent: c.isVNIPersistent,
	}, c.log)
	if err != nil {
		return nil, err
	}

	c.vni = vni

	if vni.GetMode() == go_tuntap.TAP && c.serverConfig.BRCtl4Go != nil {
		c.log(fmt.Sprintf("add interface %s to bridge", vni.GetName()))
		err := c.serverConfig.BRCtl4Go.AddInterface(vni.GetName())
		if err != nil {
			return nil, err
		}
	}

	c.setNextState(c.handle)
	return nil, nil
}

func (c *VPNClient) handle(arguments interface{}) (interface{}, error) {
	defer c.log(fmt.Sprintf("handler for %s ended", c.remoteAddr))
	c.setNextState(nil)

	c.log("invoking connected hooks")
	c.serverConfig.ConnectedHookInvokeChannel <- c
	c.log("connected hooks invoked")

	c.state = VPN_CLIENT_STATUS_NORMAL
	c.stop = make(chan struct{})

	c.log("accepting vni stream")
	vniStream, err := c.smuxSession.Accept()
	if err != nil {
		return nil, err
	}
	c.log("vni stream accepted")
	c.vniStream = vniStream

	c.log("net<--->vni exchange handler starting")
	exchangeEvent := startReadWriterExchange(vniStream, c.vni, uint(c.serverConfig.FullFrameMTU), c.log)
	c.log("net<--->vni exchanging...")

	c.log("control message handler starting")
	controlMessageEvent := c.startControlMessageHandler()
	c.log("control message handler started")

	c.log("waiting for handler's events...")
	select {
	case <-c.stop:
	case <-exchangeEvent:
	case <-controlMessageEvent:
	}
	return nil, nil
}

func (c *VPNClient) startControlMessageHandler() chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer func() {
			close(ch)
			c.log("control message handler ended")
		}()

		controlMessageStream := c.controlMessageStream
		var err error

		controlMessageBuffer := make([]byte, ControlMessageBufferSize)
		if err != nil {
			c.log(err.Error())
			return
		}
		// var n int
		// var controlMessage []byte

		for {
			// TODO: implement control message
			_, _, err = readControlMessageWithProvidedBuffer(controlMessageStream, controlMessageBuffer)
			if err != nil {
				c.log(err.Error())
				return
			}
		}
	}()
	return ch
}

func (c *VPNClient) SendStopSignal() error {
	if c.stop == nil {
		return errors.New("stop in current status isn't unsupported")
	}
	close(c.stop)
	c.stop = nil
	return nil
}

func (c *VPNClient) Close() error {
	// strange block on close vni stream, so trying to close smux session first...
	if c.smuxSession != nil {
		_ = c.smuxSession.Close()
		c.smuxSession = nil
		c.log("smux session closed")
	}
	if c.controlMessageStream != nil {
		_ = c.controlMessageStream.Close()
		c.controlMessageStream = nil
		c.log("control message stream closed")
	}
	if c.vniStream != nil {
		c.log("closing vni stream")
		_ = c.vniStream.Close()
		c.log("vni stream closed")
	}
	if c.session != nil {
		_ = c.session.Close()
		c.session = nil
		c.log("net connection closed")
	}
	if c.vni != nil {
		_ = c.vni.Close()
		c.vni = nil
		c.log("vni closed")
	}
	if c.peerIP != 0 {
		c.serverConfig.GetIP4AM().Release(c.peerIP)
		c.log(fmt.Sprintf("ip %s released", long2ip(c.peerIP)))
		c.peerIP = 0
	}
	c.state = VPN_CLIENT_STATUS_CLOSED
	return nil
}

func (c *VPNClient) log(message string) {
	SendLog(fmt.Sprintf("[%s]: %s", c.remoteAddr, message))
}
