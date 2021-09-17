package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

//!!! refactor connections to be list
type Connections = []net.Conn            //storing all peer connections
type MessagesSent = map[Transaction]bool //storing the messages sent

type Peer struct {
	outbound               chan Transaction //The channel used to handle incoming messages, funelling them to a separate method to handle broadcast and printing
	messagesSent           MessagesSent     //Map of the messages this peer has already sent and printed to user
	messagesSentMutex      *sync.Mutex      //Mutex for handling concurrency when inserting into the messagesSent map
	Connections            Connections      //Map containing all the active connections for this peer
	connectionsMutex       *sync.Mutex      //Mutex for handling concurrency when reading from og writing to the connections map.
	uriStrategy            UriStrategy      //Strategy for getting the URI to which it tries to connect
	userInputStrategy      UserInputStrategy
	outboundIPStrategy     OutboundIPStrategy
	messageSendingStrategy MessageSendingStrategy
	port                   string //outbound port (for taking new connections)
	ip                     string //outbound ip
	ledger                 *Ledger
}

func MakePeer(uri UriStrategy, user UserInputStrategy, outbound OutboundIPStrategy, message MessageSendingStrategy) *Peer {
	//initialize message channel, message map, connections map and connections mutex
	peer := new(Peer)
	peer.outbound = make(chan Transaction)
	peer.connectionsMutex = &sync.Mutex{}
	peer.Connections = make([]net.Conn, 0)
	peer.messagesSent = make(map[Transaction]bool)
	peer.messagesSentMutex = &sync.Mutex{}
	peer.uriStrategy = uri
	peer.userInputStrategy = user
	peer.outboundIPStrategy = outbound
	peer.messageSendingStrategy = message
	peer.ledger = MakeLedger()
	return peer
}

func main() {
	//Intialize strategy and peer
	commandLineUriStrategy := new(CommandLineUriStrategy) // Strategy to get the URI from user command-line input
	commandLineUserInputStrategy := new(CommandLineUserInputStrategy)
	outboundIPStrategy := new(RealOutboundIPStrategy)
	messageSendingStrategy := new(RealMessageSendingStrategy)
	peer := MakePeer(commandLineUriStrategy, commandLineUserInputStrategy, outboundIPStrategy, messageSendingStrategy)
	peer.run()
}

func (peer *Peer) run() {
	//ask for IP and port of an existing peer
	uri := peer.GetURI()

	//connect to the given IP and port via TCP
	out_conn := peer.ConnectToNetwork(uri)
	if out_conn != nil {
		defer out_conn.Close()
	}

	//print own IP and port
	listener := peer.StartListeningForConnections()
	defer listener.Close()

	//take input from the user
	go peer.HandleIncomingFromUser()

	//set up a thread to send outbound messages
	go peer.SendMessages()

	//listen for connections from other pères
	for {
		peer.TakeNewConnection(listener)
	}
}

func (peer *Peer) TakeNewConnection(listener net.Listener) {
	in_conn, err := listener.Accept()
	fmt.Println("Connection accepted on IP: ", listener.Addr().String())
	if err != nil {
		fmt.Println("New peer connection failed")
		panic(-1)
	}
	//add the new connection to connections
	peer.AppendToConnections(in_conn)
	peer.SendConnections(in_conn)
	//handle input from the new connection and send all previous messages to new
	go peer.HandleIncomingFromPeer(in_conn)
	go peer.SendAllPrevious(in_conn)
}

func (peer *Peer) GetURI() string {
	return peer.uriStrategy.GetURI()
}

func (peer *Peer) StartListeningForConnections() net.Listener {
	peer.ip = peer.outboundIPStrategy.GetOutboundIP()
	listener, _ := net.Listen("tcp", peer.ip+":")
	_, own_port, _ := net.SplitHostPort(listener.Addr().String())
	peer.port = own_port
	fmt.Println("Taking connections on " + peer.ip + ":" + own_port)
	return listener
}

func (peer *Peer) ConnectToNetwork(uri string) net.Conn {
	//connect to the given uri via TCP
	fmt.Println("Connecting to uri: ", uri)
	out_conn, err := net.Dial("tcp", uri)
	if err != nil {
		fmt.Println("No peer found, starting new  peer to peer network")
		return nil
	} else {
		fmt.Println("Connected to peer in network, you can now send and receive from the network")
		peer.connectionsMutex.Lock()
		peer.Connections = peer.ReceiveConnections(out_conn)
		peer.connectionsMutex.Unlock()

		//make connections to 10 newest peers!!!
		//broadcast own presence!!!

		go peer.HandleIncomingFromPeer(out_conn)
		return out_conn
	}
}

func (peer *Peer) ReceiveConnections(coming_from net.Conn) []net.Conn {
	reader := bufio.NewReader(coming_from)
	marshalled, err := reader.ReadBytes(']') //!!! Correct delimiter
	fmt.Println(marshalled)
	if err != nil {
		fmt.Println("Lost connection to Peer")
		return make([]net.Conn, 0)
	}
	connections := peer.DemarshalConnections(marshalled)
	fmt.Println("Received connections")
	return connections
}

func (peer *Peer) SendAllPrevious(conn net.Conn) {
	//send all old messages in the messagesSent map to a new connection
	peer.messagesSentMutex.Lock()
	defer peer.messagesSentMutex.Unlock()
	fmt.Println("Sending this many previous messages to new peer:", len(peer.messagesSent))
	i := 0
	for message := range peer.messagesSent {
		fmt.Println("Sending message number ", i)
		i++
		peer.SendMessage(conn, message)
	}
}

func (peer *Peer) SendMessages() {
	for {
		//get a message from the channel and check if it has been sent before
		message := <-peer.outbound
		peer.messagesSentMutex.Lock()
		if !peer.messagesSent[message] {
			//if this message has not been sent before, print it to the user
			fmt.Println("Message put in ledger: ", message)
			peer.UpdateLedger(&message) //Update ledger
			peer.messagesSent[message] = true
			peer.messagesSentMutex.Unlock()
			//send the message out to all peers in the network
			peer.messageSendingStrategy.SendMessageToAllPeers(message, peer)
		} else {
			peer.messagesSentMutex.Unlock()
		}
	}
}

func (peer *Peer) SendMessage(connection net.Conn, message Transaction) {
	//send the message to the connection
	marshalled := peer.MarshalTransaction(message)
	_, err := connection.Write(marshalled)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(connection)
	}
}

func (peer *Peer) SendTransaction(conn net.Conn, transaction Transaction) {
	marshalled := peer.MarshalTransaction(transaction)
	_, err := conn.Write(marshalled)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(conn)
	}
}

func (peer *Peer) SendConnections(conn net.Conn) {
	fmt.Println(peer.Connections)
	marshalled := peer.MarshalConnections(peer.GetConnections())
	_, err := conn.Write(marshalled)
	fmt.Println("Sent connections to new peer")
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(conn)
	}
}

func (peer *Peer) AppendToConnections(conn net.Conn) {
	peer.connectionsMutex.Lock()
	peer.Connections = append(peer.GetConnections(), conn)
	peer.connectionsMutex.Unlock()
}

func (peer *Peer) DeleteFromConnections(conn net.Conn) {
	peer.connectionsMutex.Lock()
	defer peer.connectionsMutex.Unlock()
	for index, connection := range peer.GetConnections() {
		if connection == conn {
			peer.Connections = peer.Remove(peer.GetConnections(), index)
			break
		}
	}
}

func (peer *Peer) Remove(slice []net.Conn, s int) []net.Conn {
	return append(slice[:s], slice[s+1:]...)
}

func (peer *Peer) HandleIncomingFromUser() { //!!! not necessary any longer, except for testing
	// kinda still necessary though, we haven't refactored enough yet...
	for {
		msg := peer.userInputStrategy.HandleIncomingFromUser()
		peer.outbound <- msg
	}
}

func (peer *Peer) HandleIncomingFromPeer(connection net.Conn) {
	defer connection.Close()
	//take messages from the peer
	for {
		reader := bufio.NewReader(connection)
		marshalled, err := reader.ReadBytes(']') //!!! Delim
		if err != nil {
			fmt.Println("Lost connection to Peer")
			return
		}
		msg := peer.DemarshalTransaction(marshalled)
		//add message to channel
		peer.outbound <- msg
	}
}

func (peer *Peer) UpdateLedger(transaction *Transaction) {
	peer.ledger.Transaction(transaction)
}

func (peer *Peer) MarshalTransaction(transaction Transaction) []byte {
	bytes, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println("Marshaling failed")
	}
	//bytes = append(bytes, "\n"...)
	return bytes
}

func (peer *Peer) DemarshalTransaction(bytes []byte) Transaction {
	var transaction Transaction
	//bytes = append(bytes, "]"...)
	err := json.Unmarshal(bytes, &transaction)
	if err != nil {
		fmt.Println("Demarshaling failed")
	}
	return transaction
}

func (peer *Peer) MarshalConnections(connections []net.Conn) []byte {
	peer.connectionsMutex.Lock()
	bytes, err := json.Marshal(peer.GetConnections())
	peer.connectionsMutex.Unlock()
	if err != nil {
		fmt.Println("Marshaling failed")
	}
	//bytes = append(bytes, "\n"...)
	fmt.Println(bytes)
	return bytes
}

func (peer *Peer) DemarshalConnections(bytes []byte) []net.Conn {
	var connections Connections
	//bytes = append(bytes, "]"...)
	err := json.Unmarshal(bytes, &connections)
	if err != nil {
		fmt.Println("Demarshaling failed", err)
	}
	fmt.Println(connections)
	return connections
}

func (peer *Peer) GetConnections() []net.Conn {
	return peer.Connections
}
