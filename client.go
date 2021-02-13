package main

import (
	"fmt"
	"github.com/xtaci/tcpraw"
	"github.com/yzslab/kcp-go"
	"log"
	"time"
)

func startClient(config *ClientConfig) error {
	config.CommonConfig.PrintSummary()

	block := createBlock(&config.CommonConfig)

	remoteAddr := fmt.Sprintf("%s:%d", config.GetIP(), config.GetPort())
	log.Printf("connecting to %s", remoteAddr)

	var session *kcp.UDPSession
	var err error
	if config.EnableTCPSimulation {
		conn, err := tcpraw.Dial("tcp", remoteAddr)
		if err != nil {
			return err
		}
		session, err = kcp.NewConn(remoteAddr, block, config.GetDatashard(), config.GetParityshard(), conn)
	} else {
		session, err = kcp.DialWithOptions(remoteAddr, block, config.GetDatashard(), config.GetParityshard())
	}
	if err != nil {
		return err
	}

	defer session.Close()
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

	server, err := NewVPNServer(session, config, config.IsVNIPersistent())
	if err != nil {
		return err
	}
	defer server.Close()

	_, err = IterateState(server)
	return err
}
