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

type ConnectionsURI = []string

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
	ledger                 *Ledger
	connectionsURI         ConnectionsURI
	connectionsURIMutex    *sync.Mutex
}

func MakePeer(uri UriStrategy, user UserInputStrategy, outbound OutboundIPStrategy, message MessageSendingStrategy) *Peer {
	//initialize message channel, message map, connections map and connections mutex
	peer := new(Peer)
	peer.outbound = make(chan Transaction)
	peer.connectionsMutex = &sync.Mutex{}
	peer.connections = make([]net.Conn, 0)
	peer.messagesSent = make(map[Transaction]bool)
	peer.messagesSentMutex = &sync.Mutex{}
	peer.uriStrategy = uri
	peer.userInputStrategy = user
	peer.outboundIPStrategy = outbound
	peer.messageSendingStrategy = message
	peer.ledger = MakeLedger()
	peer.connectionsURI = make([]string, 0)
	peer.connectionsURIMutex = &sync.Mutex{}
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
	out_conn := peer.JoinNetwork(uri)
	if out_conn != nil {
		defer out_conn.Close()
	}

	//print own IP and port
	listener := peer.StartListeningForConnections()
	defer listener.Close()

	//broadcast new presence in network
	//peer.AddSelfToConnectionsURI()
	//peer.BroadcastPresence() //broadcast own presence!!!

	//take input from the user
	go peer.HandleIncomingFromUser()

	//set up a thread to send outbound messages
	go peer.SendMessages()

	//listen for connections from other peers
	for {
		peer.TakeNewConnection(listener)
	}
}

func (peer *Peer) TakeNewConnection(listener net.Listener) {
	in_conn, err := listener.Accept()
	fmt.Println("Connection accepted on IP: ", listener.Addr().String())
	if err != nil {
		fmt.Println("New peer connection failed")
		return
	}

	//add the new connection to connections
	fmt.Println("Adding this connection to slice:", in_conn.RemoteAddr().String())
	peer.AppendToConnections(in_conn)
	//time.Sleep(1 * time.Second) //!!! trying to allow other connection time to catch up
	//peer.SendConnectionsURI(in_conn)

	//handle input from the new connection and send all previous messages to new
	go peer.HandleIncomingFromPeer(in_conn)
	//go peer.SendAllPreviousMessagesToPeer(in_conn)
}

func (peer *Peer) GetURI() string {
	return peer.uriStrategy.GetURI()
}

func (peer *Peer) AddSelfToConnectionsURI() {
	peer.AppendToConnectionsURI(peer.ip + ":" + peer.port)
}

func (peer *Peer) StartListeningForConnections() net.Listener {
	peer.ip = peer.outboundIPStrategy.GetOutboundIP()
	listener, _ := net.Listen("tcp", peer.ip+":")
	_, own_port, _ := net.SplitHostPort(listener.Addr().String())
	peer.port = own_port
	fmt.Println("Taking connections on " + peer.ip + ":" + own_port)
	return listener
}

func (peer *Peer) JoinNetwork(uri string) net.Conn {
	//connect to the given uri via TCP
	fmt.Println("Connecting to uri: ", uri)
	out_conn, err := net.Dial("tcp", uri)
	if err != nil {
		fmt.Println("No peer found, starting new  peer to peer network")
		return nil
	} else {
		fmt.Println("Connected to peer in network, you can now send and receive from the network")
		peer.AppendToConnections(out_conn) //!!!should this be here? - tjek opgavebeskrivelse

		/*peer.connectionsURIMutex.Lock()
		peer.connectionsURI = peer.ReceiveConnectionsURI(out_conn)
		peer.connectionsURIMutex.Unlock()
		fmt.Println("Received connectionsURI")*/

		//peer.ConnectToFirst10PeersInConnectionsURI(peer.connectionsURI)

		go peer.HandleIncomingFromPeer(out_conn) //!!!same as above - should this be here?
		return out_conn
	}
}

func (peer *Peer) ConnectToPeer(uri string) {
	out_conn, err := net.Dial("tcp", uri)
	if err != nil {
		fmt.Println("No peer found on uri: ", uri)
	} else {
		peer.AppendToConnections(out_conn)
		go peer.HandleIncomingFromPeer(out_conn)
	}
}

func (peer *Peer) ReceiveConnectionsURI(coming_from net.Conn) ConnectionsURI {
	reader := bufio.NewReader(coming_from)
	marshalled, err := reader.ReadBytes(']') //!!! Correct delimiter (this seems to be correct)
	fmt.Println("Received the still marshalled connections list:", marshalled, " now calling demarshal:")
	if err != nil {
		fmt.Println("Lost connection to Peer")
		panic(-1)
	}
	fmt.Println("Received connections")
	connectionsURI := peer.DemarshalConnectionsURI(marshalled)
	return connectionsURI
}

func (peer *Peer) ConnectToFirst10PeersInConnectionsURI(connectionsURI ConnectionsURI) {
	peer.connectionsURIMutex.Lock()
	defer peer.connectionsURIMutex.Unlock()
	index := len(connectionsURI) - 1
	i := 0
	for i < 10 && index >= 0 {
		uri := connectionsURI[index]
		peer.ConnectToPeer(uri)
		i++
		index--
	}
}

func (peer *Peer) SendAllPreviousMessagesToPeer(conn net.Conn) {
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
			fmt.Println("Message put in ledger and sent: ", message)
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
	fmt.Println("Sending this message:", message, "To this:", connection.RemoteAddr().String())
	marshalled := peer.MarshalTransaction(message)
	_, err := connection.Write(marshalled)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(connection)
	}
}

func (peer *Peer) SendConnectionsURI(conn net.Conn) {
	fmt.Println("Now marshalling this connections array: ", peer.connections, "and sending to ", conn.RemoteAddr().String())
	marshalled := peer.MarshalConnectionsURI(peer.connectionsURI)
	_, err := conn.Write(marshalled)
	fmt.Println("Sent", len(peer.connectionsURI), " connections to new peer")
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(conn)
	}
}

func (peer *Peer) AppendToConnections(conn net.Conn) {
	peer.connectionsMutex.Lock()
	peer.connections = append(peer.GetConnections(), conn)
	peer.connectionsMutex.Unlock()
}

func (peer *Peer) AppendToConnectionsURI(uri string) {
	peer.connectionsURIMutex.Lock()
	peer.connectionsURI = append(peer.connectionsURI, uri)
	peer.connectionsURIMutex.Unlock()
}

func (peer *Peer) DeleteFromConnections(conn net.Conn) {
	peer.connectionsMutex.Lock()
	defer peer.connectionsMutex.Unlock()
	for index, connection := range peer.GetConnections() {
		if connection == conn {
			peer.connections = peer.RemoveConnection(peer.GetConnections(), index)
			break
		}
	}
}

func (peer *Peer) DeleteFromConnectionsURI(uri string) {
	peer.connectionsURIMutex.Lock()
	defer peer.connectionsURIMutex.Unlock()
	for index, connection := range peer.connectionsURI {
		if connection == uri {
			peer.connectionsURI = peer.RemoveURI(peer.connectionsURI, index)
			break
		}
	}
}

func (peer *Peer) RemoveConnection(slice []net.Conn, s int) []net.Conn {
	return append(slice[:s], slice[s+1:]...)
}

func (peer *Peer) RemoveURI(slice []string, s int) []string {
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
	reader := bufio.NewReader(connection)
	for {
		fmt.Println("Waiting for input...")
		marshalled, err := reader.ReadBytes(']') //!!! Delim
		if err != nil {
			fmt.Println("Lost connection to peer")
			return
		}
		fmt.Println("Got input, demarshalling")
		msg, err := peer.DemarshalTransaction(marshalled)
		if err != nil {
			fmt.Println("Tried to demarshall something that was not a transaction, ignoring")
		} else {
			//add message to channel
			peer.outbound <- msg
		}
	}
}

func (peer *Peer) UpdateLedger(transaction *Transaction) {
	peer.ledger.Transaction(transaction)
}

func (peer *Peer) MarshalTransaction(transaction Transaction) []byte {
	bytes, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println("Marshaling transaction failed")
	}
	return bytes
}

func (peer *Peer) DemarshalTransaction(bytes []byte) (Transaction, error) {
	var transaction Transaction
	fmt.Println("Trying to demarshal this transaction:", bytes)
	err := json.Unmarshal(bytes, &transaction)
	if err != nil {
		fmt.Println("Demarshaling transaction failed")
	}
	return transaction, err
}

func (peer *Peer) MarshalConnectionsURI(connectionsURI ConnectionsURI) []byte {
	peer.connectionsURIMutex.Lock()
	fmt.Println("Marshalling performed on this array:", connectionsURI, " end of array")
	bytes, err := json.Marshal(connectionsURI)
	peer.connectionsURIMutex.Unlock()
	if err != nil {
		fmt.Println("Marshaling connectionsURI failed")
	}
	fmt.Println(bytes)
	return bytes
}

func (peer *Peer) DemarshalConnectionsURI(bytes []byte) ConnectionsURI {
	var connectionsURI ConnectionsURI
	err := json.Unmarshal(bytes, &connectionsURI)
	if err != nil {
		fmt.Println("Demarshaling connectionsURI failed", err)
	}
	fmt.Println(connectionsURI)
	return connectionsURI
}

func (peer *Peer) GetConnections() []net.Conn {
	return peer.connections
}

//!!!should be removed
/*
func (peer *Peer) MarshalASingleConnection(connections net.Conn) []byte {
	peer.connectionsMutex.Lock()
	fmt.Println("Marshalling performed on this array: end of array")
	bytes, err := json.Marshal(connections)
	peer.connectionsMutex.Unlock()
	if err != nil {
		fmt.Println("Marshaling failed")
	}
	fmt.Println(bytes)
	return bytes
}

func (peer *Peer) DemarshalASingleConnection(bytes []byte) net.Conn {
	var connections net.Conn
	err := json.Unmarshal(bytes, &connections)
	if err != nil {
		fmt.Println("Demarshaling failed", err)
	}
	fmt.Println(connections)
	return connections
}
*/
