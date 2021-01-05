package main

import (
	"errors"
	"io"
)

type MessageType uint8

const (
	MessageClientInformation MessageType = iota
	MessageClientSetIPValueAccepted
	MessageClientSetIPValueRejected
	MessageServerAssignedIPValue
	MessageServerAssignableIPExhausted
	MessageServerConfiguration
	MessageClientReady
	MessageServerReady
)

type ClientIPMode uint8

const (
	ClientIPModeServerAssign ClientIPMode = iota
	ClientIPModeClientSet
	ClientIPModeOther
)

func buildControlMessage(message []byte) []byte {
	buffer := make([]byte, len(message)+2)
	buffer[0] = 0b11010101
	var i int
	messageLength := len(message)
	for i = 0; i < messageLength; i++ {
		buffer[1+i] = message[i]
	}
	buffer[1+i] = buffer[0]
	return buffer
}

func retrieveControlMessage(message []byte) (int, []byte, error) {
	messageLength := len(message)
	if messageLength <= 2 {
		return -1, nil, errors.New("invalid message length")
	}
	if message[0] != 0b11010101 && message[messageLength-1] != 0b11010101 {
		return -1, nil, errors.New("invalid message tag")
	}
	return messageLength - 2, message[1 : messageLength-1], nil
}

func writeControlMessage(writer io.Writer, buffer []byte) (int, error) {
	return writer.Write(buildControlMessage(buffer))
}

func readControlMessageWithProvidedBuffer(reader io.Reader, buffer []byte) (int, []byte, error) {
	n, err := reader.Read(buffer)
	if err != nil {
		return -1, nil, err
	}
	return retrieveControlMessage(buffer[:n])
}

func readControlMessage(reader io.Reader, messageLength uint) (int, []byte, error) {
	buffer := make([]byte, messageLength+2)
	return readControlMessageWithProvidedBuffer(reader, buffer)
}
