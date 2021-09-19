package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestShouldReadMarshalledValuesCorrectlyFromConn(t *testing.T) {
	peer1, _ := connectTwoPeers(t)
	time.Sleep(1 * time.Second)
	realConn := peer1.connectionsURI
	fmt.Println("Testing on this array:", realConn, "with length: ", len(realConn))
	marshalledConn := peer1.MarshalConnectionsURI(realConn)
	fmt.Println("Marshalled ConnectionsURI: ", marshalledConn)
	demarshalledConn := peer1.DemarshalConnectionsURI(marshalledConn)
	fmt.Println("Demarshalled ConnectionsURI: ", demarshalledConn)
	marshalledCorrectly := testEq(demarshalledConn, realConn)
	if !marshalledCorrectly {
		t.Error("Expected,", realConn, "Got", demarshalledConn)
	} else {
		fmt.Println("TestShouldReadMarshalledValuesCorrectlyFromConn passed")
	}
}
func TestShouldMarshalConnectionsURIOverNetWork(t *testing.T) {
	peer1, peer2 := connectTwoPeers(t)
	time.Sleep(1 * time.Second)
	realConn := peer1.connectionsURI
	fmt.Println("Testing on this array:", realConn, "with length: ", len(realConn))
	demarshalledConn := peer2.connectionsURI
	marshalledCorrectly := testEq(demarshalledConn, realConn)
	fmt.Println("Peer1:", realConn, "Peer 2:", demarshalledConn)
	if !marshalledCorrectly {
		t.Error("Expected,", realConn, "Got", demarshalledConn)
	} else {
		fmt.Println("TestShouldMarshalConnectionsURIOverNetWork passed")
	}
}

func TestShouldReadMarshalledConn(t *testing.T) {
	peer1, _ := connectTwoPeers(t)
	time.Sleep(1 * time.Second)
	realConn := peer1.connections[0]

	fmt.Println("Testing on this array:", realConn, "with length: , len(realConn)")
	fmt.Println("This is the type of the conn object: ", reflect.TypeOf(realConn))
	remoteAddressString := realConn.RemoteAddr().String()
	fmt.Println("This is the remoteAddr: ", remoteAddressString)
	marshalledConn := peer1.MarshalASingleConnection(realConn)
	demarshalledConn := peer1.DemarshalASingleConnection(marshalledConn)

	fmt.Println("demarshalled connection: ", demarshalledConn)

	sentCorrectly := realConn.RemoteAddr() == demarshalledConn.RemoteAddr()
	if !sentCorrectly {
		t.Error("Expected,", realConn, "Got", demarshalledConn)
	} else {
		fmt.Println("TestShouldReadMarshalledConn passed")
	}
}

func TestShouldMarshalConnectionsCorrectly(t *testing.T) {
	/*peer1, _ := connectTwoPeers(t)
	realConn := peer1.GetConnections()*/
	peer1 := peerFixture()
	realConn := []string{"yo", "yo"}
	fmt.Println("Realconn:", realConn, "End")
	marshalledConn := peer1.MarshalConnectionsURI(realConn)
	demarshalledConn := peer1.DemarshalConnectionsURI(marshalledConn)
	marshallingCorrect := testEq(realConn, demarshalledConn)
	fmt.Println(realConn)
	time.Sleep(1 * time.Second)
	if !marshallingCorrect {
		t.Errorf("Arrays should be equal")
	} else {
		fmt.Println("Peer test TestShouldMarshalConnectionsCorrectly passed")
	}
}

func TestPeerUpdateLedgerShouldUpdateWithTransaction(t *testing.T) {
	peer := peerFixture()
	transaction := MakeTransaction("transID", "acc1", "acc2", 100)
	peer.ledger.Accounts["acc1"] = 200
	peer.ledger.Accounts["acc2"] = 200
	peer.UpdateLedger(transaction)
	accountBalancesCorrect := peer.ledger.Accounts["acc1"] == 100 && peer.ledger.Accounts["acc2"] == 300

	if !accountBalancesCorrect {
		t.Errorf("Account balances should be correct")
	} else {
		fmt.Println("Peer test TestPeerUpdateLedgerShouldUpdateWithTransaction passed")
	}

}

func TestShouldMarshalTransactionCorrectly(t *testing.T) {
	peer1 := peerFixture()

	transaction := MakeTransaction("1234", "Mathias", "Rasmus", 100)
	fmt.Println("Marshalling this transaction: ", transaction)
	marshalled := peer1.MarshalTransaction(*transaction)
	fmt.Println("This is marshalled:", marshalled)
	demarshalled, _ := peer1.DemarshalTransaction(marshalled)
	fmt.Println("Getting this demarshalled thing back:", demarshalled)
	if (*transaction) != demarshalled {
		t.Error("")
	} else if demarshalled.Amount != 100 {
		t.Error("Field amount not demarshalled properly")

	} else {
		fmt.Println("Marshalling test passed")
	}
}

/*
func TestShouldSendMessage(t *testing.T) {
	msg := "testbesked"

	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(msg)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()

	peer1 := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	go peer1.HandleIncomingFromUser()
	go peer1.SendMessages()
	time.Sleep(2 * time.Second)
	if !messageSendingStrategy.messagesSent[msg] {
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
	listener := peer1.StartListeningForConnections()
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
	listener := peer1.StartListeningForConnections()
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
*/
func peerFixture() *Peer {
	transaction := MakeTransaction("id", "acc1", "acc2", 100)

	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()

	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)
	return peer

}

func connectTwoPeers(t *testing.T) (*Peer, *Peer) {
	transaction := MakeTransaction("id", "acc1", "acc2", 100)
	fixedUriStrategy1 := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	realOutboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy1, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	peer1.JoinNetwork(peer1.GetURI())
	listener := peer1.StartListeningForConnections()
	defer listener.Close()
	go peer1.TakeNewConnection(listener)

	fixedUriStrategy2 := MakeFixedUriStrategy(peer1.ip, peer1.port)
	peer2 := MakePeer(fixedUriStrategy2, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	peer2.JoinNetwork(peer2.GetURI())

	return peer1, peer2

}

func connectNewPeer(peer *Peer, t *testing.T) *Peer {
	transaction := MakeTransaction("id", "acc1", "acc2", 100)
	fixedUriStrategy1 := MakeFixedUriStrategy(peer.ip, peer.port)
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	realOutboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	newPeer := MakePeer(fixedUriStrategy1, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)
	newPeer.JoinNetwork(newPeer.GetURI())

	return newPeer
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
