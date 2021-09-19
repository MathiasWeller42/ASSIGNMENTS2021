package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Connections = map[net.Conn]interface{} //storing all peer connections
type MessagesSent = map[Transaction]bool    //storing the messages sent

type Peer struct {
	outbound               chan Transaction //The channel used to handle incoming messages, funelling them to a separate method to handle broadcast and printing
	messagesSent           MessagesSent     //Map of the messages this peer has already sent and printed to user
	messagesSentMutex      *sync.Mutex      //Mutex for handling concurrency when inserting into the messagesSent map
	connections            Connections      //Map containing all the active connections for this peer
	connectionsMutex       *sync.Mutex      //Mutex for handling concurrency when reading from og writing to the connections map.
	uriStrategy            UriStrategy      //Strategy for getting the URI to which it tries to connect
	userInputStrategy      UserInputStrategy
	outboundIPStrategy     OutboundIPStrategy
	messageSendingStrategy MessageSendingStrategy
	port                   string //outbound port (for taking new connections)
	ip                     string //outbound ip
}

func MakePeer(uri UriStrategy, user UserInputStrategy, outbound OutboundIPStrategy, message MessageSendingStrategy) *Peer {
	//initialize message channel, message map, connections map and connections mutex
	peer := new(Peer)
	peer.outbound = make(chan Transaction)
	peer.connectionsMutex = &sync.Mutex{}
	peer.connections = make(map[net.Conn]interface{})
	peer.messagesSent = make(map[Transaction]bool)
	peer.messagesSentMutex = &sync.Mutex{}
	peer.port = ""
	peer.uriStrategy = uri
	peer.userInputStrategy = user
	peer.outboundIPStrategy = outbound
	peer.messageSendingStrategy = message
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
	listener := peer.PrintOwnURI()
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
	peer.connectionsMutex.Lock()
	peer.connections[in_conn] = nil
	peer.connectionsMutex.Unlock()
	//handle input from the new connection and send all previous messages to new
	go peer.HandleIncomingFromPeer(in_conn)
	go peer.SendAllPrevious(in_conn)
}

func (peer *Peer) GetURI() string {
	return peer.uriStrategy.GetURI()
}

func (peer *Peer) PrintOwnURI() net.Listener {
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
		peer.connections[out_conn] = nil
		peer.connectionsMutex.Unlock()
		go peer.HandleIncomingFromPeer(out_conn)
		return out_conn
	}
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
			fmt.Println("Transaction, from: ", message.From, "to: ", message.To, "Amount: ", message.Amount)
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
		peer.connectionsMutex.Lock()
		delete(peer.connections, connection)
		peer.connectionsMutex.Unlock()
	}
}

func (peer *Peer) HandleIncomingFromUser() {
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
		marshalled, err := reader.ReadBytes(']') //delim
		if err != nil {
			fmt.Println("Lost connection to Peer")
			return
		}

		msg, err := peer.DemarshalTransaction(marshalled)
		if err != nil {
			fmt.Println("Mistake in demarshalling, ignoring transaction: ", msg, "error: ", err)
		} else {
			//add message to channel
			peer.outbound <- msg
		}
	}
}

func (peer *Peer) MarshalTransaction(transaction Transaction) []byte {
	bytes, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println("Marshaling transaction failed")
	}
	//add extra ']' as delimiter
	bytes = append(bytes, ']')
	return bytes
}

func (peer *Peer) DemarshalTransaction(bytes []byte) (Transaction, error) {
	var transaction Transaction
	//delete the extra ']'
	bytes = bytes[:len(bytes)-1]
	fmt.Println("Trying to demarshal this transaction:", string(bytes))
	err := json.Unmarshal(bytes, &transaction)
	if err != nil {
		fmt.Println("Demarshaling transaction failed", err)
	}
	return transaction, err
}
