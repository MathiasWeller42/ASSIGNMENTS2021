package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestShouldSendMessage(t *testing.T) {
	msg := "testbesked"

	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(msg)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()

	peer1 := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	go peer1.HandleIncomingFromUser()
	time.Sleep(2 * time.Second)
	if messageSendingStrategy.messagesSent[msg] {
		t.Error("Message was not present in map")
	} else if len(messageSendingStrategy.messagesSent) < 1 {
		t.Errorf("No messages in map")
	} else {
		fmt.Println("Test 0 passed")
	}

}

func TestShouldReturnCorrectURI(t *testing.T) {
	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy("")
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	uri := peer1.GetURI()
	if uri != "123:123" {
		t.Errorf("The uri was not correct")
	} else {
		fmt.Println("Test 1 passed")
	}
}

func TestShouldMakeNewNetworkOnInvalidIP(t *testing.T) {
	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy("")
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	out_conn := peer1.ConnectToNetwork(peer1.GetURI())
	if len(peer1.connections) != 0 {
		t.Errorf("Network too large, should have been 0")
	}
	if out_conn != nil {
		t.Errorf("Connected to a network when it should not have")
	} else {
		fmt.Println("Test 2 passed.")
	}
}

func TestShouldStartListeningSuccessfully(t *testing.T) {
	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy("")
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)
	listener := peer1.PrintOwnURI()
	defer listener.Close()

	if peer1.ip != "localhost" {
		t.Errorf("Wrong outbound IP")
	}
	intval, _ := strconv.Atoi(peer1.port)
	if intval < 0 || intval > 65535 {
		t.Errorf("Illegal port")
	} else {
		fmt.Println("Test 3 passed")
	}
}

func TestShouldConnectToExistingNetwork(t *testing.T) {
	fixedUriStrategy1 := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy("")
	realOutboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy1, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	peer1.ConnectToNetwork(peer1.GetURI())
	listener := peer1.PrintOwnURI()
	defer listener.Close()
	go peer1.TakeNewConnection(listener)

	fixedUriStrategy2 := MakeFixedUriStrategy(peer1.ip, peer1.port)
	peer2 := MakePeer(fixedUriStrategy2, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	out_conn := peer2.ConnectToNetwork(peer2.GetURI())

	if out_conn == nil {
		t.Errorf("Connection did not work")
	}
	defer out_conn.Close()
	time.Sleep(time.Second)
	if len(peer1.connections) != 1 {
		t.Errorf("Peer1's network should include peer2")
	} else if len(peer2.connections) != 1 {
		t.Errorf("Peer2's network should include peer1")
	} else {
		fmt.Println("Test 4 passed.")
	}
}
