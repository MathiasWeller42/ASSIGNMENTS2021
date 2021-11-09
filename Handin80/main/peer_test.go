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

func TestShouldSendVerifiedTransactionCorrectlyToAnotherPeer(t *testing.T) {
	peer1, listener1 := createPeerSeq("fa", "fa")
	defer listener1.Close()

	peer2, listener := createPeer(peer1.ip, peer1.port)
	defer listener.Close()
	newRsa := MakeRSA(2000)
	publicKey := (newRsa.n).String()
	secretKey := (newRsa.d).String()

	peer3, listener3 := createPeerWithTransaction(peer1.ip, peer1.port, publicKey, "bob", 200, secretKey)
	defer listener3.Close()
	time.Sleep(11 * time.Second)
	if !(peer3.ledger.Accounts[publicKey] == 800) {
		fmt.Println("ledger3 account not right, has value", peer1.ledger.Accounts[publicKey])
	} else if !(peer2.ledger.Accounts[publicKey] == 800) {
		fmt.Println("ledger2 account not right")
	} else if !(peer1.ledger.Accounts[publicKey] == 800) {
		fmt.Println("ledger1 account not right")
	} else {
		fmt.Println("Test passed, the transaction was succesfully sent and verified at all peers")
	}

}

func TestShouldNotSendUnVerifiedTransactionCorrectlyToAnotherPeer(t *testing.T) {
	peer1, listener1 := createPeerSeq("fa", "fa")
	defer listener1.Close()

	peer2, listener := createPeer(peer1.ip, peer1.port)
	defer listener.Close()
	newRsa := MakeRSA(2000)
	publicKey := (newRsa.n).String()
	secretKey := "wrong"

	peer3, listener3 := createPeerWithTransaction(peer1.ip, peer1.port, publicKey, "bob", 200, secretKey)
	defer listener3.Close()
	time.Sleep(11 * time.Second)
	if !(peer3.ledger.Accounts[publicKey] == 0) {
		fmt.Println("ledger3 account not right, has value", peer1.ledger.Accounts[publicKey])
	} else if !(peer2.ledger.Accounts[publicKey] == 0) {
		fmt.Println("ledger2 account not right")
	} else if !(peer1.ledger.Accounts[publicKey] == 0) {
		fmt.Println("ledger1 account not right")
	} else {
		fmt.Println("Test passed, the tranasction was not sent and thus was not verified")
	}

}

func TestShouldSendVerified1000TransactionCorrectlyToAnotherPeer(t *testing.T) {
	peer1, listener1 := createPeerSeq("fa", "fa") //This peer becomes the sequencer
	defer listener1.Close()

	peer2, listener2 := createPeer(peer1.ip, peer1.port)
	defer listener2.Close()
	newRsa := MakeRSA(2000)
	publicKey := (newRsa.n).String()
	secretKey := (newRsa.d).String()

	peer3, listener3 := createPeerWithTransactionSendXtimes(peer1.ip, peer1.port, publicKey, "B", 10, secretKey, 100) //Peer3 attemps to send 100x10 to "B" from "publickey"
	defer listener3.Close()
	peer4, listener4 := createPeerWithTransactionSendXtimes(peer1.ip, peer1.port, publicKey, "C", 10, secretKey, 100) //Peer4 attemps to send 100x10 to "C" from "publickey"
	defer listener4.Close()

	//wait a while to make sure everyone is done connecting and sending transactions (especially with the 10 second block times)
	time.Sleep(21 * time.Second)

	//check that A has balance 0 everywhere
	if !(peer1.ledger.Accounts[publicKey] == 0 && peer2.ledger.Accounts[publicKey] == 0 && peer3.ledger.Accounts[publicKey] == 0 && peer4.ledger.Accounts[publicKey] == 0) {
		fmt.Println("ledger1 account not right, has value", peer1.ledger.Accounts[publicKey])
		fmt.Println("ledger2 account not right, has value", peer2.ledger.Accounts[publicKey])
		fmt.Println("ledger3 account not right, has value", peer3.ledger.Accounts[publicKey])
		fmt.Println("ledger4 account not right, has value", peer4.ledger.Accounts[publicKey])

		//Check that account B has the same value for all peers
	} else if !(peer1.ledger.Accounts["B"] == peer2.ledger.Accounts["B"] && peer2.ledger.Accounts["B"] == peer3.ledger.Accounts["B"] && peer3.ledger.Accounts["B"] == peer4.ledger.Accounts["B"]) {
		fmt.Println("The B account has not been updated the same at all peers")
		fmt.Println("ledger1 B account not right, has value", peer1.ledger.Accounts["B"])
		fmt.Println("ledger2 B account not right, has value", peer2.ledger.Accounts["B"])
		fmt.Println("ledger3 B account not right, has value", peer3.ledger.Accounts["B"])
		fmt.Println("ledger4 B account not right, has value", peer4.ledger.Accounts["B"])

		//Check that account C has the same value for all peers
	} else if !(peer1.ledger.Accounts["C"] == peer2.ledger.Accounts["C"] && peer2.ledger.Accounts["C"] == peer3.ledger.Accounts["C"] && peer3.ledger.Accounts["C"] == peer4.ledger.Accounts["C"]) {
		fmt.Println("The C account has not been updated the same at all peers")
		fmt.Println("ledger1 C account not right, has value", peer1.ledger.Accounts["C"])
		fmt.Println("ledger2 C account not right, has value", peer2.ledger.Accounts["C"])
		fmt.Println("ledger3 C account not right, has value", peer3.ledger.Accounts["C"])
		fmt.Println("ledger4 C account not right, has value", peer4.ledger.Accounts["C"])

		//Check that the sum of B and C is 3000, since the 1000 coins from A should be distributed on B and C, and both accounts started with 1000 coins.
	} else if (peer1.ledger.Accounts["C"] + peer1.ledger.Accounts["B"]) != 3000 {
		fmt.Println("The collected balance of B and C is incorrect, was", peer1.ledger.Accounts["C"]+peer1.ledger.Accounts["B"], "expected", 3000)

		//Hurray everything worked!! :D
	} else {
		fmt.Println("Test passed, the transaction was succesfully sent and verified at all peers")
		fmt.Println("B has final value", peer1.ledger.Accounts["B"])
		fmt.Println("C has final value", peer1.ledger.Accounts["C"])
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

func createPeer(ip string, port string) (*Peer, net.Listener) { //This peer is seq
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

func createPeerSeq(ip string, port string) (*Peer, net.Listener) { //This peer is seq - this method exists as to make it easier to merge with next assignment
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
	if peer.seq {
		go peer.HandleBlocks()
	}
	go takeNewConnectionsHelp(peer, listener)

	return peer, listener
}

func createPeerWithTransaction(ip string, port string, from string, to string, amount int, secret string) (*Peer, net.Listener) {
	transaction := MakeSignedTransaction(from, to, amount, secret)

	fixedUriStrategy := MakeFixedUriStrategy(ip, port)
	fixedInputStrategy := MakeFixedInputStrategy(*transaction)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := new(RealMessageSendingStrategy)
	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	uri := peer.GetURI()

	peer.JoinNetwork(uri)

	listener := peer.StartListeningForConnections()

	peer.AddSelfToConnectionsURI()

	ownURI := peer.ip + ":" + peer.port
	peer.BroadcastPresence(ownURI)

	//go peer.HandleIncomingFromUser()
	go send1message(peer)
	go peer.SendMessages()
	go takeNewConnectionsHelp(peer, listener)

	return peer, listener
}

func takeNewConnectionsHelp(peer *Peer, listener net.Listener) {
	for {
		peer.TakeNewConnection(listener)
	}
}

func send1message(peer *Peer) {
	msg := peer.userInputStrategy.HandleIncomingFromUser()
	peer.outbound <- msg
}

func createPeerWithTransactionSendXtimes(ip string, port string, from string, to string, amount int, secret string, x int) (*Peer, net.Listener) {
	transaction := MakeSignedTransaction(from, to, amount, secret)

	fixedUriStrategy := MakeFixedUriStrategy(ip, port)
	fixedInputStrategy, _ := MakeFixedInputStrategy2(*transaction, from, secret)
	fixedOutboundIPStrategy := MakeFixedOutboundIPStrategy("localhost")
	messageSendingStrategy := new(RealMessageSendingStrategy)
	peer := MakePeer(fixedUriStrategy, fixedInputStrategy, fixedOutboundIPStrategy, messageSendingStrategy)

	uri := peer.GetURI()

	peer.JoinNetwork(uri)

	listener := peer.StartListeningForConnections()

	peer.AddSelfToConnectionsURI()

	ownURI := peer.ip + ":" + peer.port
	peer.BroadcastPresence(ownURI)

	//go peer.HandleIncomingFromUser()
	go sendXmessage(peer, x)
	go peer.SendMessages()
	go takeNewConnectionsHelp(peer, listener)

	return peer, listener
}

func sendXmessage(peer *Peer, x int) {
	time.Sleep(1 * time.Second)
	for i := 0; i < x; i++ {
		msg := peer.userInputStrategy.HandleIncomingFromUser()
		peer.outbound <- msg
	}
}

/*
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
*/
