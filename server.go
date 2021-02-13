package main

import (
	"fmt"
	"github.com/xtaci/tcpraw"
	"github.com/yzslab/kcp-go"
	"log"
	"os"
	"os/exec"
	"time"
)

func startServer(config *ServerConfig) error {
	defer config.GetIP4AM().Close()

	config.CommonConfig.PrintSummary()

	block := createBlock(&config.CommonConfig)
	listenAddressAndPort := fmt.Sprintf("%s:%d", config.GetIP(), config.GetPort())

	var listener *kcp.Listener
	var err error
	if config.EnableTCPSimulation {
		if conn, tcpErr := tcpraw.Listen("tcp", listenAddressAndPort); err == nil {
			listener, err = kcp.ServeConn(block, config.GetDatashard(), config.GetParityshard(), conn)
		} else {
			return tcpErr
		}
	} else {
		listener, err = kcp.ListenWithOptions(listenAddressAndPort, block, config.GetDatashard(), config.GetParityshard())
	}
	if err != nil {
		return err
	}

	defer listener.Close()
	log.Printf("listening on %s", listenAddressAndPort)

	if err := listener.SetDSCP(config.GetDSCP()); err != nil {
		return fmt.Errorf("SetDSCP: %s", err)
	}
	if err := listener.SetReadBuffer(config.GetSocketBufferSize()); err != nil {
		return fmt.Errorf("SetReadBuffer: %s", err)
	}
	if err := listener.SetWriteBuffer(config.GetSocketBufferSize()); err != nil {
		return fmt.Errorf("SetWriteBuffer: %s", err)
	}

	connectedHookInvokeChannel, connectedHookHandlerStopChannel := startConnectedHookHandler()
	defer close(connectedHookHandlerStopChannel)
	config.ConnectedHookInvokeChannel = connectedHookInvokeChannel

	log.Println("waiting for connections...")
	for {
		session, err := listener.AcceptKCP()
		if err != nil {
			return err
		}

		log.Println("connection opened:", session.LocalAddr(), "->", session.RemoteAddr())

		session.SetStreamMode(true)
		session.SetWriteDelay(false)
		session.SetNoDelay(config.GetNodelay(), config.GetInterval(), config.GetResend(), config.GetNoCongestion())
		session.SetMtu(int(config.GetUDPMTU()))
		session.SetWindowSize(config.GetSendWindowSize(), config.GetReceiveWindowSize())
		session.SetACKNoDelay(config.GetAckNodelay())
		if config.Datashard != 0 && config.Parityshard != 0 {
			session.SetRapidFec(config.EnableRapidFec)
			session.SetRapidFecMinInterval(time.Duration(config.GetInterval()) * time.Millisecond)
		}

		client, err := NewVPNClient(session, fmt.Sprintf("%s%d", config.GetVNINamePrefix(), config.IncreaseConnectionCounter()), config)

		var state State
		state = client

		go func() {
			defer client.Close()

			_, err := IterateState(state)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func startConnectedHookHandler() (chan *VPNClient, chan struct{}) {
	vpnClientChan := make(chan *VPNClient)
	closeChan := make(chan struct{})

	go func() {
		var vpnClient *VPNClient

		isFileExists := func(path string) bool {
			info, err := os.Stat(path)
			if err != nil {
				return false
			}
			if info.IsDir() {
				return false
			}
			return true
		}

		runHook := func(hookPath string, vpnClient *VPNClient) error {
			vpnClient.log(fmt.Sprintf("hook: %s", hookPath))
			cmd := exec.Command(hookPath)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("KV_CLIENT_ID=%s", vpnClient.clientId),
				fmt.Sprintf("KV_VNI_INTERFACE_NAME=%s", vpnClient.vniName),
				fmt.Sprintf("KV_CLIENT_IP=%s", long2ip(vpnClient.peerIP)),
				fmt.Sprintf("KV_CLIENT_IP_MODE=%d", vpnClient.ClientIPMode),
				fmt.Sprintf("KV_REMOTE_ADDR=%s", vpnClient.remoteAddr),
			)
			err := cmd.Run()
			return err
		}

		runDefaultHook := func(vpnClient *VPNClient) error {
			defaultHookPath := fmt.Sprintf("%s/on_connected", vpnClient.serverConfig.HookDirectory)
			if isFileExists(defaultHookPath) {
				return runHook(defaultHookPath, vpnClient)
			}
			return nil
		}

	HandlerLoop:
		for {
			select {
			case vpnClient = <-vpnClientChan:
				if vpnClient.serverConfig.HookDirectory == "" {
					continue
				}
				var err error
				if vpnClient.clientIdLength == 0 {
					err = runDefaultHook(vpnClient)
				} else {
					hookPath := fmt.Sprintf("%s/on_%s_connected", vpnClient.serverConfig.HookDirectory, vpnClient.clientId)
					if isFileExists(hookPath) {
						err = runHook(hookPath, vpnClient)
					} else {
						err = runDefaultHook(vpnClient)
					}
				}
				if err != nil {
					vpnClient.log(err.Error())
				}
				break
			case <-closeChan:
				break HandlerLoop
			}
		}
	}()

	return vpnClientChan, closeChan
}
