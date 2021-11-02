package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Connections = []net.Conn       //storing all peer connections
type MessagesSent = map[string]bool //storing the messages sent

type ConnectionsURI = []string

type Peer struct {
	outbound               chan SignedTransaction //The channel used to handle incoming messages, funelling them to a separate method to handle broadcast and printing
	messagesSent           MessagesSent           //Map of the messages this peer has already sent and printed to user
	messagesSentMutex      *sync.Mutex            //Mutex for handling concurrency when inserting into the messagesSent map
	connections            Connections            //Map containing all the active connections for this peer
	connectionsMutex       *sync.Mutex            //Mutex for handling concurrency when reading from og writing to the connections map.
	uriStrategy            UriStrategy            //Strategy for getting the URI to which it tries to connect
	userInputStrategy      UserInputStrategy
	outboundIPStrategy     OutboundIPStrategy
	messageSendingStrategy MessageSendingStrategy
	port                   string //outbound port (for taking new connections)
	ip                     string //outbound ip
	ledger                 *Ledger
	connectionsURI         ConnectionsURI //Holds the URIs of all peers currently present in the network.
	connectionsURIMutex    *sync.Mutex    //Mutex for connectionsURI
	rsa                    *RSA           //RSA object to do verification and signing
	seq                    bool           //bool to tell us if we're the sequencer or not
	seqRSA                 *RSA           //this only exists for the sequencer
	seqPk                  string         //the Public key for the sequencer (everyone has this)
}

func MakePeer(uri UriStrategy, user UserInputStrategy, outbound OutboundIPStrategy, message MessageSendingStrategy) *Peer {
	//Initialize all fields
	peer := new(Peer)
	peer.outbound = make(chan SignedTransaction)
	peer.connectionsMutex = &sync.Mutex{}
	peer.connections = make([]net.Conn, 0)
	peer.messagesSent = make(map[string]bool)
	peer.messagesSentMutex = &sync.Mutex{}
	peer.uriStrategy = uri
	peer.userInputStrategy = user
	peer.outboundIPStrategy = outbound
	peer.messageSendingStrategy = message
	peer.ledger = MakeLedger()
	peer.connectionsURI = make([]string, 0)
	peer.connectionsURIMutex = &sync.Mutex{}
	peer.rsa = MakeRSA(2000)
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
	//ask for IP and port of an existing peer via user input or other strategy
	otherURI := peer.GetURI()

	//connect to the given IP and port via TCP
	conn := peer.JoinNetwork(otherURI)
	if conn != nil {
		defer conn.Close()
	}

	//listen for connections on own ip and port to which other peers can connect, the listener object is passed to takeNewConnection
	listener := peer.StartListeningForConnections()
	defer listener.Close()

	//add yourself to end of connectionsURI list which was received in the joinNetwork call
	peer.AddSelfToConnectionsURI()

	//broadcast new presence in network so everyone can append you to connectionsURI
	ownURI := peer.ip + ":" + peer.port
	peer.BroadcastPresence(ownURI)

	//take input from the user (for testing purposes)
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
	//the other may or may not listen, but we do not know, so we add it to be sure
	peer.AppendToConnections(in_conn)

	//send own connectionsURI in case the new peer is brand new
	peer.SendConnectionsURI(in_conn)

	//send seqPk
	peer.SendSeqPk(in_conn)

	//handle input from the new connection
	go peer.HandleIncomingMessagesFromPeer(in_conn)
}

func (peer *Peer) BroadcastPresence(uri string) {
	//add a delimiter to make it easier to read on the other side
	uriToSend := uri + "uri]"
	peer.connectionsMutex.Lock()
	defer peer.connectionsMutex.Unlock()

	//send the presence to all connections
	for _, conn := range peer.connections {
		_, err := fmt.Fprint(conn, uriToSend)
		if err != nil {
			//delete the missing connection
			peer.DeleteFromConnections(conn)
		}
	}
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
		fmt.Println("No peer found, starting new  peer to peer network - you are the sequencer")
		peer.seq = true
		peer.seqRSA = MakeRSA(2000)
		n := ConvertBigIntToString(&(peer.seqRSA.n))
		e := ConvertBigIntToString(peer.seqRSA.e)
		peer.seqPk = n + ":" + e
		fmt.Println("This is seqPK:", peer.seqPk)
		return nil
	} else {
		peer.seq = false
		peer.seqRSA = nil
		peer.AppendToConnections(out_conn)
		//receive the peer's connectionsURI list before anything else
		peer.connectionsURIMutex.Lock()
		peer.connectionsURI = peer.ReceiveConnectionsURI(out_conn)
		peer.connectionsURIMutex.Unlock()
		//receive seqPk
		peer.seqPk = peer.ReceiveSeqPk(out_conn)
		//handle messages
		go peer.HandleIncomingMessagesFromPeer(out_conn)

		//connect to the 10 peers before yourself in the list
		peer.ConnectToFirst10PeersInConnectionsURI(peer.connectionsURI, uri)
		return out_conn
	}
}

func (peer *Peer) ConnectToPeer(uri string) {
	out_conn, err := net.Dial("tcp", uri)
	if err != nil {
		return
	} else {
		peer.AppendToConnections(out_conn)
		go peer.HandleIncomingMessagesFromPeer(out_conn)
	}
}

func (peer *Peer) ReceiveConnectionsURI(coming_from net.Conn) ConnectionsURI {
	reader := bufio.NewReader(coming_from)
	marshalled, err := reader.ReadBytes(']')
	if err != nil {
		fmt.Println("Lost connection to peer")
		panic(-1)
	}
	connectionsURI := peer.DemarshalConnectionsURI(marshalled)
	return connectionsURI
}

func (peer *Peer) ReceiveSeqPk(coming_from net.Conn) string {
	reader := bufio.NewReader(coming_from)
	bytes, err := reader.ReadBytes(']')
	if err != nil {
		fmt.Println("Lost connection to peer")
		panic(-1)
	}
	asString := string(bytes)
	seqPk := asString[:len(asString)-1]
	fmt.Println("received this seqPk" + seqPk)
	return seqPk
}

func (peer *Peer) ConnectToFirst10PeersInConnectionsURI(connectionsURI ConnectionsURI, olduri string) {
	peer.connectionsURIMutex.Lock()
	defer peer.connectionsURIMutex.Unlock()
	index := len(connectionsURI) - 1
	i := 0
	for i < 10 && index >= 0 {
		uri := connectionsURI[index]
		if uri != olduri {
			peer.ConnectToPeer(uri)
		}
		i++
		index--
	}
}

func (peer *Peer) SendMessages() {
	for {
		//get a message from the channel and check if it has been sent before
		message := <-peer.outbound
		peer.messagesSentMutex.Lock()
		if !peer.messagesSent[message.ID] {
			//if this message has not been sent before, print it to the user (for testing purposes) and update ledger
			success := peer.UpdateLedger(&message) //Update ledger - returns true if successful otherwise false
			//print the ledger for manual testing purposes
			peer.ledger.Print()
			peer.messagesSent[message.ID] = true
			peer.messagesSentMutex.Unlock()
			//send the message out to all peers in the network
			if success {
				fmt.Println("Transaction was validated and added to ledger, now broadcasting it")
				peer.messageSendingStrategy.SendMessageToAllPeers(message, peer)
			} else {
				fmt.Println("Did not send an invalid transaction")
			}

		} else {
			peer.messagesSentMutex.Unlock()
		}
	}
}

func (peer *Peer) SendMessage(connection net.Conn, message SignedTransaction) {
	//send the message to the connection
	marshalled := peer.MarshalTransaction(message)
	_, err := connection.Write(marshalled)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(connection)
	}
}

func (peer *Peer) SendConnectionsURI(conn net.Conn) {
	marshalled := peer.MarshalConnectionsURI(peer.connectionsURI)
	_, err := conn.Write(marshalled)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(conn)
	}
}

func (peer *Peer) SendSeqPk(conn net.Conn) {
	toBeSent := []byte(peer.seqPk + "]")
	_, err := conn.Write(toBeSent)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(conn)
	}
	fmt.Println("sent this seqPk: " + peer.seqPk)
}

func (peer *Peer) AppendToConnections(conn net.Conn) {
	peer.connectionsMutex.Lock()
	peer.connections = append(peer.GetConnections(), conn)
	peer.connectionsMutex.Unlock()
}

func (peer *Peer) AppendToConnectionsURI(uri string) bool {
	peer.connectionsURIMutex.Lock()
	defer peer.connectionsURIMutex.Unlock()
	//only add to connectionsURI if the URI is not already in there
	shouldAdd := !peer.contains(peer.connectionsURI, uri)
	if shouldAdd {
		peer.connectionsURI = append(peer.connectionsURI, uri)
	}
	//for broadcasting presence we need to know whether it was new or not
	return shouldAdd
}

//taken from StackOverflow https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func (peer *Peer) contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

//only used for manual testing
func (peer *Peer) HandleIncomingFromUser() {
	peer.AddNewSkUser()
	for {
		msg := peer.userInputStrategy.HandleIncomingFromUser()
		peer.outbound <- msg
	}
}

func (peer *Peer) HandleIncomingMessagesFromPeer(connection net.Conn) {
	defer connection.Close()
	//take messages from the peer
	reader := bufio.NewReader(connection)
	for {
		marshalled, err := reader.ReadBytes(']')
		if err != nil {
			fmt.Println("Lost connection to peer")
			return
		}
		msg, err := peer.DemarshalTransaction(marshalled)
		if err != nil {
			//Tried to demarshall something that was not a transaction, trying to read as it as a presence (URI)
			asString := string(marshalled)
			if strings.Contains(asString, "uri]") {
				//it is not a transaction, but a URI presence
				uriString := asString[:len(asString)-4]
				//add it to connectionsURI and if it was new, keep broadcasting
				continueBroadcasting := peer.AppendToConnectionsURI(uriString)
				fmt.Println("Added new URI, list now has length:", len(peer.connectionsURI))
				if continueBroadcasting {
					peer.BroadcastPresence(uriString)
				}
			} else {
				//This was a connectionsURI list so ignore it
			}
		} else {
			//demarshalled a transaction - adding message to channel
			peer.outbound <- msg
		}
	}
}

func (peer *Peer) UpdateLedger(transaction *SignedTransaction) bool {
	var success bool
	if transaction.Amount >= 0 && peer.rsa.VerifyTransaction(*transaction) {
		peer.ledger.Transaction(transaction)
		fmt.Println("Message successfully put in ledger")
		success = true
	} else {
		success = false
		fmt.Println("Invalid transaction", transaction)
	}
	return success
}

func (peer *Peer) MarshalTransaction(transaction SignedTransaction) []byte {
	bytes, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println("Marshaling transaction failed")
	}
	//add extra ']' as delimiter
	bytes = append(bytes, ']')
	return bytes
}

func (peer *Peer) DemarshalTransaction(bytes []byte) (SignedTransaction, error) {
	var transaction SignedTransaction
	//delete the extra ']'
	bytes = bytes[:len(bytes)-1]
	err := json.Unmarshal(bytes, &transaction)
	return transaction, err
}

func (peer *Peer) MarshalConnectionsURI(connectionsURI ConnectionsURI) []byte {
	peer.connectionsURIMutex.Lock()
	bytes, err := json.Marshal(connectionsURI)
	peer.connectionsURIMutex.Unlock()
	if err != nil {
		fmt.Println("Marshaling connectionsURI failed")
	}
	return bytes
}

func (peer *Peer) DemarshalConnectionsURI(bytes []byte) ConnectionsURI {
	var connectionsURI ConnectionsURI
	err := json.Unmarshal(bytes, &connectionsURI)
	if err != nil {
		fmt.Println("Demarshaling connectionsURI failed", err)
	}
	return connectionsURI
}

func (peer *Peer) GetConnections() []net.Conn {
	return peer.connections
}

func (peer *Peer) AddNewSkUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Would you like to make a new account - answer y/n?")
	decision, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	trimmedDecision := strings.TrimRight(decision, "\r\n")
	if trimmedDecision == "y" || trimmedDecision == "yes" {
		newRsa := MakeRSA(2000)
		publicKey := (newRsa.n).String()
		secretKey := (newRsa.d).String()
		success := peer.ledger.AddAccount(publicKey)
		if success {
			fmt.Println("Successfully created new account, this is your secret Key:")
			fmt.Println(secretKey) //notice that both secret and public key are formatted as strings corresponding to the value inside the BigInt, and NOT bytes translated into string from the bigInt.
			fmt.Println("And this is your public name:")
			fmt.Println(publicKey)
		}
	} else {
		fmt.Println("You have chosen to use a preexisting account.")
	}
	fmt.Println("You can now make transactions using your public and secret key.")
	fmt.Println("---------------------------------------------------------------------------------------------------------")
}
