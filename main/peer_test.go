package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestShouldReadMarshalledValuesCorrectlyFromConn(t *testing.T) {
	peer1, _, listener, listener2 := connectTwoPeers(t)
	defer listener.Close()
	defer listener2.Close()
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
	peer1, peer2, listener, listener2 := connectTwoPeers(t)
	defer listener.Close()
	defer listener2.Close()
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

/*
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
*/
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
	transaction := MakeTransaction("acc1", "acc2", 100)
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

	transaction := MakeTransaction("Mathias", "Rasmus", 100)
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

func TestConnurishouldhavelen10after10peers(t *testing.T) {
	peer1, _, listener, listener2 := connectTwoPeers(t)
	defer listener.Close()
	defer listener2.Close()
	for i := 0; i < 9; i++ {
		go connectNewPeer(peer1, t)
	}
	time.Sleep(4 * time.Second)
	conns := peer1.connectionsURI

	if len(conns) == 10 {
		fmt.Println("Test passed")
	} else {
		t.Error("Not all peers are added to conn uri, list has length:", len(conns))
	}
}

func peerFixture() *Peer {
	transaction := MakeTransaction("acc1", "acc2", 100)

	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()

	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)
	return peer
}

func connectTwoPeers(t *testing.T) (*Peer, *Peer, net.Listener, net.Listener) {
	transaction := MakeTransaction("acc1", "acc2", 100)
	fixedUriStrategy1 := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	realOutboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer1 := MakePeer(fixedUriStrategy1, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	peer1.JoinNetwork(peer1.GetURI())
	listener := peer1.StartListeningForConnections()
	//defer listener.Close()
	go peer1.TakeNewConnection(listener)

	fixedUriStrategy2 := MakeFixedUriStrategy(peer1.ip, peer1.port)
	peer2 := MakePeer(fixedUriStrategy2, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)

	peer2.JoinNetwork(peer2.GetURI())
	listener2 := peer2.StartListeningForConnections()
	defer listener2.Close()
	go peer2.TakeNewConnection(listener)

	uri1 := peer1.GetURI()
	peer1.AddSelfToConnectionsURI()
	peer1.BroadcastPresence(uri1)

	uri2 := peer2.GetURI()
	peer2.AddSelfToConnectionsURI()
	peer2.BroadcastPresence(uri2)

	return peer1, peer2, listener, listener2

}

func connectNewPeer(peer *Peer, t *testing.T) {
	transaction := MakeTransaction("acc1", "acc2", 100)
	fixedUriStrategy1 := MakeFixedUriStrategy(peer.ip, peer.port)
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	realOutboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	newPeer := MakePeer(fixedUriStrategy1, fixedInputStrategy, realOutboundIPStrategy, messageSendingStrategy)
	newPeer.JoinNetwork(newPeer.GetURI())
	listener := newPeer.StartListeningForConnections()
	defer listener.Close()
	go newPeer.TakeNewConnection(listener)

	uri := newPeer.GetURI()
	newPeer.AddSelfToConnectionsURI()
	newPeer.BroadcastPresence(uri)
	for {
		go newPeer.TakeNewConnection(listener)
	}

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
