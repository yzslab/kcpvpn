package main

import (
	"github.com/yzslab/goipam"
)

func ip2long(ipAddr string) (uint32, error) {
	return goipam.IP2long(ipAddr)
}

func long2ip(ipLong uint32) string {
	return goipam.Long2ip(ipLong)
}