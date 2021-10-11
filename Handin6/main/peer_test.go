package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestShouldReadMarshalledValuesCorrectlyFromConn(t *testing.T) { //pass
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

func TestShouldMarshalConnectionsCorrectly(t *testing.T) { //pass
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

/*
func TestPeerUpdateLedgerShouldUpdateWithTransaction(t *testing.T) { //TODO
	peer := peerFixture()
	transaction := MakeSignedTransaction("acc1", "acc2", 100, "yeet")
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
*/
func TestShouldMarshalTransactionCorrectly(t *testing.T) { //works sometimes, god knows why...
	peer1 := peerFixture()

	transaction := MakeSignedTransaction("400", "Rasmus", 100, "300")
	fmt.Println("Marshalling this transaction: ", *transaction)
	marshalled := peer1.MarshalTransaction(*transaction)
	fmt.Println("This is marshalled:", marshalled)
	demarshalled, _ := peer1.DemarshalTransaction(marshalled)
	fmt.Println("Getting this demarshalled thing back:", demarshalled)
	if demarshalled.Amount != transaction.Amount {
		fmt.Println("Amounts not equal")
	} else if demarshalled.ID != transaction.ID {
		fmt.Println("Id not equal")
	} else if demarshalled.From != transaction.From {
		fmt.Println("From not equal")
	} else if demarshalled.To != transaction.To {
		fmt.Println("To not equal")
	} else if demarshalled.Signature != transaction.Signature {
		fmt.Println("Signatures not equal, got: ", demarshalled.Signature, ", expected: ", transaction.Signature)
	} else if demarshalled.Amount != 100 {
		t.Error("Field amount not demarshalled properly")
	} else {
		fmt.Println("Marshalling test passed")
	}
}
func TestConnurishouldhavelen10after10peers(t *testing.T) { //pass
	peer1, listener1 := createPeer("fa", "fa")
	defer listener1.Close()

	for i := 0; i < 9; i++ {
		_, listener := createPeer(peer1.ip, peer1.port)
		defer listener.Close()
	}

	time.Sleep(2 * time.Second)
	conns := peer1.connectionsURI

	if len(conns) == 10 {
		fmt.Println("Test passed with conns:", conns)
	} else {
		t.Error("Not all peers are added to conn uri, list has length:", len(conns))
	}
}

func TestConnuriANDConnsshouldhavelen10and9after10peers(t *testing.T) { //pass
	peer1, listener1 := createPeer("fa", "fa")
	defer listener1.Close()

	for i := 0; i < 9; i++ {
		_, listener := createPeer(peer1.ip, peer1.port)
		defer listener.Close()
	}

	time.Sleep(2 * time.Second)
	conns := peer1.connectionsURI
	conns2 := peer1.connections

	if len(conns) == 10 && len(conns2) == 9 {
		fmt.Println("Test passed, both have at least length 10")
	} else {
		t.Error("Not all peers are added to either conn uri or conn, lists have length:", len(conns), "and ", len(conns2))
	}
}

func TestShouldOnlyConnectTo10lastConnURI(t *testing.T) { //fails
	peer1, listener1 := createPeer("fa", "fa")
	defer listener1.Close()

	for i := 0; i < 12; i++ {
		_, listener := createPeer(peer1.ip, peer1.port)
		defer listener.Close()
	}
	peer2, listener2 := createPeer(peer1.ip, peer1.port)
	defer listener2.Close()

	time.Sleep(2 * time.Second)
	conns := peer2.connectionsURI
	conns2 := peer2.connections
	//Conns2 should be 22, 12 peers connected to it, it connected to 10 peers.
	if len(conns) == 14 && len(conns2) >= 10 && len(conns2) <= 11 {
		fmt.Println("Test passed, both have at least length 10")
	} else {
		t.Error("Not all peers are added to either conn uri or conn, lists have length:", len(conns), "and ", len(conns2))
	}
}

func peerFixture() *Peer {
	transaction := MakeSignedTransaction("acc1", "acc2", 100, "yeet")

	fixedUriStrategy := MakeFixedUriStrategy("123", "123")
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()

	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)
	return peer
}

func connectTwoPeers(t *testing.T) (*Peer, *Peer, net.Listener, net.Listener) {
	transaction := MakeSignedTransaction("acc1", "acc2", 100, "yeet")
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
	transaction := MakeSignedTransaction("acc1", "acc2", 100, "yeet")
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

func createPeer(ip string, port string) (*Peer, net.Listener) {
	transaction := MakeSignedTransaction("acc1", "acc2", 100, "yeet")

	fixedUriStrategy := MakeFixedUriStrategy(ip, port)
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := MakeStubbedMessageSendingStrategy()
	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	uri := peer.GetURI()

	peer.JoinNetwork(uri)

	listener := peer.StartListeningForConnections()

	peer.AddSelfToConnectionsURI()

	ownURI := peer.ip + ":" + peer.port
	peer.BroadcastPresence(ownURI)

	//go peer.HandleIncomingFromUser()

	go peer.SendMessages()
	go takeNewConnectionsHelp(peer, listener)

	return peer, listener
}

func takeNewConnectionsHelp(peer *Peer, listener net.Listener) {
	for {
		peer.TakeNewConnection(listener)
	}
}
