package main

import (
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	usageFatal := func() {
		log.Fatalf("Usage: %s server|client --help", os.Args[0])
	}

	if len(os.Args) < 2 {
		usageFatal()
	}

	command := os.Args[1]

	StartLogRoutine()
	if command == "server" || command == "s" {
		err := createServerConfig(func(serverConfig *ServerConfig) {
			err := startServer(serverConfig)
			errorCheck(err)
		})
		errorCheck(err)
	} else if command == "client" || command == "c" {
		err := createClientConfig(func(clientConfig *ClientConfig) {
			err := startClient(clientConfig)
			if err != nil {
				log.Println(err)
			}
			if clientConfig.AutoReconnect == true {
				time.Sleep(1 * time.Second)
				if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
					log.Fatal(err)
				}
			}
		})
		errorCheck(err)
	} else {
		usageFatal()
	}
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}