package main

import (
	"fmt"
	"net"
)

type OutboundIPStrategy interface {
	GetOutboundIP() string
}

type RealOutboundIPStrategy struct {
}

// Get preferred outbound IP of this machine
// From SO: https://stackoverflow.com/a/37382208 (thank you, Simon :D)
func (outboundIPStrategy *RealOutboundIPStrategy) GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Mistake while trying to get own IP")
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

type FixedOutboundIPStrategy struct {
	ip string
}

func MakeFixedOutboundIPStrategy(ip string) *FixedOutboundIPStrategy {
	outboundIPStrategy := new(FixedOutboundIPStrategy)
	outboundIPStrategy.ip = ip
	return outboundIPStrategy
}

func (outboundIPStrategy *FixedOutboundIPStrategy) GetOutboundIP() string {
	return outboundIPStrategy.ip
}
