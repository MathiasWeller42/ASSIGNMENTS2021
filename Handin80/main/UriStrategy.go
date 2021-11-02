package main

import (
	"bufio"
	"fmt"
	"os"
)

type UriStrategy interface {
	GetURI() string
}

type CommandLineUriStrategy struct {
}

func (uriStrategy *CommandLineUriStrategy) GetURI() string {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Please enter an IP address:")
	scanner.Scan()
	ip := scanner.Text()

	fmt.Println("Please enter port: ")
	scanner.Scan()
	port := scanner.Text()

	uri := ip + ":" + port
	return uri
}

type FixedUriStrategy struct {
	uri string
}

func MakeFixedUriStrategy(ip, port string) *FixedUriStrategy {
	uriStrategy := new(FixedUriStrategy)
	uriStrategy.uri = ip + ":" + port
	return uriStrategy
}

func (uriStrategy *FixedUriStrategy) GetURI() string {
	return uriStrategy.uri
}
