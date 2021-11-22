package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Connections = []net.Conn                    //storing all peer connections
type MessagesSent = map[string]TransactionStruct //storing the messages sent

type ConnectionsURI = []string

type Block = []string //second to last entry is the ID, last is the delimiter

type TransactionStruct struct {
	transaction SignedTransaction
	sent        bool
}

type BlocksSent = map[string]bool

type Peer struct {
	outbound               chan SignedTransaction //The channel used to handle incoming messages, funelling them to a separate method to handle broadcast and printing
	messagesSent           MessagesSent           //Map of the messages this peer has already sent and printed to user
	messagesSentMutex      *sync.Mutex            //Mutex for handling concurrency when inserting into the messagesSent map
	connections            Connections            //Map containing all the active connections for this peer
	connectionsMutex       *sync.Mutex            //Mutex for handling concurrency when reading from and writing to the connections map.
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
	nextBlock              Block
	nextBlockMutex         *sync.Mutex
	genesisLedger          *Ledger
	seed                   int
	hardness               big.Int //Dis big boi be big tuf
	genesisBlock           Block
	blocksSent             BlocksSent
	blocksSentMutex        *sync.Mutex
	slotLength             float64
	slotNumber             int
	connectionThreshold    int
	blockTree              *BlockTree
	systemRunning          bool
	winners                map[int][]string
}

func MakePeer(uri UriStrategy, user UserInputStrategy, outbound OutboundIPStrategy, message MessageSendingStrategy) *Peer {
	//Initialize all fields
	peer := new(Peer)
	peer.outbound = make(chan SignedTransaction)
	peer.connectionsMutex = &sync.Mutex{}
	peer.connections = make([]net.Conn, 0)
	peer.messagesSent = make(map[string]TransactionStruct)
	peer.messagesSentMutex = &sync.Mutex{}
	peer.uriStrategy = uri
	peer.userInputStrategy = user
	peer.outboundIPStrategy = outbound
	peer.messageSendingStrategy = message
	peer.ledger = MakeLedger()
	peer.connectionsURI = make([]string, 0)
	peer.connectionsURIMutex = &sync.Mutex{}
	peer.rsa = MakeRSA(2000)
	peer.nextBlock = make([]string, 0)
	peer.nextBlockMutex = &sync.Mutex{}
	peer.genesisLedger = MakeLedger()
	peer.hardness = *big.NewInt(0)
	peer.genesisBlock = make([]string, 0)
	peer.seed = 0
	peer.blocksSent = make(map[string]bool)
	peer.blocksSentMutex = &sync.Mutex{}
	peer.slotLength = 10 // <-- Change slotlength here!
	peer.slotNumber = 0
	peer.connectionThreshold = 2 // <-- Change # of peers here!
	peer.blockTree = nil         //This will ALWAYS point to the genesis block in the tree (after HandleGenesisBlock())
	peer.systemRunning = false
	peer.winners = make(map[int][]string)
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

	peer.AddNewSkUser()

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
	//the other may or may not listen, but we do not know, so we add it to be sure
	peer.AppendToConnections(in_conn)

	//send own connectionsURI in case the new peer is brand new
	peer.SendConnectionsURI(in_conn)

	//handle input from the new connection (and send all previous messages to new?)
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
		fmt.Println("No peer found, starting new  peer to peer network")
		go peer.SendGenesisBlockEventually()
		return nil
	} else {
		peer.AppendToConnections(out_conn)
		//receive the peer's connectionsURI list before anything else
		peer.connectionsURIMutex.Lock()
		newConnectionsURI := peer.ReceiveConnectionsURI(out_conn)
		peer.connectionsURI = newConnectionsURI
		peer.connectionsURIMutex.Unlock()

		go peer.HandleIncomingMessagesFromPeer(out_conn)

		//connect to the 10 peers before yourself in the list
		peer.ConnectToFirst10PeersInConnectionsURI(peer.connectionsURI, uri)
		return out_conn
	}
}

func (peer *Peer) SendGenesisBlockEventually() {
	peer.genesisBlock = peer.MakeGenesisBlock()
	for {
		time.Sleep(1)
		peer.connectionsURIMutex.Lock()
		if len(peer.connectionsURI) >= peer.connectionThreshold {
			peer.connectionsURIMutex.Unlock()
			fmt.Println("Enough peers in system, send genesis block")
			marshalled := peer.MarshalBlock(peer.genesisBlock)
			peer.SendBlockToAllPeers(marshalled)
			return
		}
		peer.connectionsURIMutex.Unlock()
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

func (peer *Peer) CreateGenesisLedger() {
	peer.genesisLedger = MakeLedger()
	publicKeys := peer.genesisBlock[:10]
	for _, key := range publicKeys {
		peer.genesisLedger.AddGenesisAccount(key)
	}
}

func (peer *Peer) HandleGenesisBlock() {
	publicKeys := peer.genesisBlock[:10]
	for _, key := range publicKeys {
		peer.ledger.AddGenesisAccount(key)
		peer.genesisLedger.AddGenesisAccount(key)
	}
	peer.ledger.Print()
	peer.genesisLedger.Print()

	peer.seed, _ = strconv.Atoi(peer.genesisBlock[10])
	peer.hardness = *(ConvertStringToBigInt(peer.genesisBlock[11]))

	fmt.Println("Received seed", peer.seed)
	fmt.Println("Received hardness", peer.hardness)

	//initialize blockTree
	genesisNode := MakeBlockTreeNode("vk", 0, "draw", peer.genesisBlock, "signature")
	peer.blockTree = MakeBlockTree(genesisNode)

	go peer.HandleLottery()
}

func (peer *Peer) ReceiveConnectionsURI(coming_from net.Conn) ConnectionsURI {
	reader := bufio.NewReader(coming_from)
	marshalled1, err := reader.ReadBytes(']')
	if err != nil {
		fmt.Println("Lost connection to peer")
		panic(-1)
	}
	connectionsURI := peer.DemarshalConnectionsURI(marshalled1)

	return connectionsURI
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

func (peer *Peer) SendBlockToAllPeers(marshalledBlock []byte) {
	//fmt.Println("sendBlockToAllPeers was called")
	for _, connection := range peer.connections {
		//fmt.Println("Sending block to", connection)
		peer.SendBlock(connection, marshalledBlock)
	}
}

func (peer *Peer) SendBlock(connection net.Conn, marshalledBlock []byte) {
	//send the marshalled block to the connection
	//fmt.Println("Sendblock was called")
	_, err := connection.Write(marshalledBlock)
	if err != nil {
		fmt.Println("Tried to send to a lost connection")
		//delete the missing connection
		peer.DeleteFromConnections(connection)
	}
}

func (peer *Peer) SendMessages() {
	for {
		//get a message from the channel and check if it has been sent before
		message := <-peer.outbound
		peer.messagesSentMutex.Lock()
		if !peer.messagesSent[message.ID].sent {

			transactionStruct := new(TransactionStruct)
			transactionStruct.sent = true
			transactionStruct.transaction = message
			peer.messagesSent[message.ID] = *transactionStruct
			peer.messagesSentMutex.Unlock()

			peer.nextBlockMutex.Lock()
			peer.nextBlock = append(peer.nextBlock, message.ID)
			peer.nextBlockMutex.Unlock()

			//send the message out to all peers in the network
			peer.messageSendingStrategy.SendMessageToAllPeers(message, peer)

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
	for !peer.systemRunning {
		time.Sleep(1)
	}
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
				//not a new presence, let's see if it is a block
				demarshalled, err := peer.DemarshalBlock(marshalled)
				if err != nil {
					//this was not a block, but a connectionsURI
					fmt.Println("received a connectionsURI")
					continue
				} else {
					//this was a block
					fmt.Println("received (probably) a block")
					if demarshalled[len(demarshalled)-1] == "yeet" {
						//fmt.Println("Demarshalled block") TODO: verify somehow
						peer.blocksSentMutex.Lock()
						if !peer.blocksSent[demarshalled[len(demarshalled)-2]] {
							//fmt.Println("Got a previously unseen block")
							peer.blocksSent[demarshalled[len(demarshalled)-2]] = true
							go peer.SendBlockToAllPeers(marshalled)
							if peer.slotNumber == 0 {
								peer.genesisBlock = demarshalled[:len(demarshalled)-2]
								peer.HandleGenesisBlock()
								peer.systemRunning = true
								peer.slotNumber += 1
							} else if peer.VerifyWinningBlock(*peer.rsa, demarshalled, peer.seed) { //check at vedkommende har vundet lotteriet TODO fjernede peer
								fmt.Println("Verified a winning block, adding to tree")

								prevHash := demarshalled[len(demarshalled)-4]
								if peer.blockTree.GetLongestChainLeaf().Node.OwnBlockHash == prevHash {
									success := peer.UpdateLedgerWithBlock(demarshalled)
									if success {
										fmt.Println("All transactions valid, adding block to tree")
										peer.AddBlockToTree(demarshalled)
									} else {
										fmt.Println("One or more invalid transactions, performing rollback")
										peer.UpdateLedgerOnRollback()
									}
								} else {
									fmt.Println("") //We should have checked whether some of the transactions are invalid
									peer.AddBlockToTree(demarshalled)
								}

							} else {
								fmt.Println("Did not verify a block winning")
							}
						} else {
							fmt.Println("Got a block that's seen before")
						}
						peer.blocksSentMutex.Unlock()
					} else {
						fmt.Println("Rejected a block")
					}
				}
			}
		} else {
			//demarshalled a transaction - adding message to channel
			fmt.Println("Received a transaction, sending to all")
			peer.outbound <- msg
		}
	}
}

func (peer *Peer) AddBlockToTree(block Block) {
	sigma := block[len(block)-3]
	prevHash := block[len(block)-4]
	draw := block[len(block)-5]
	slotNumber, _ := strconv.Atoi(block[len(block)-6])
	publicKey := block[len(block)-7]
	//blockConst := block[len(block)-8] //The string "BLOCK"
	blockTransactions := block[:len(block)-7]

	blockTreeNode := MakeBlockTreeNode(publicKey, slotNumber, draw, blockTransactions, sigma)
	nodeAsLeaf := MakeBlockTree(blockTreeNode)
	peer.AddChildAndRollbackIfNecessary(nodeAsLeaf, prevHash)
	fmt.Println("peer.blockTree after addChild: ", peer.blockTree)
	peer.blockTree.PrintTree()
}

func (peer *Peer) VerifyWinningBlock(rsa RSA, block Block, seed int) bool {
	//Verify that sigma = (BLOCK, slot, (U,M), h) under vk - check
	//Verify that Draw = (LOTTERY, seed, slot) under vk
	//Verify that numTickets(vk) * Hash(Draw) >= hardness
	sigma := block[len(block)-3]
	hash := block[len(block)-4]
	draw := block[len(block)-5]
	slotNumber, _ := strconv.Atoi(block[len(block)-6])
	publicKey := block[len(block)-7]
	//blockConst := block[len(block)-8] //The string "BLOCK"
	blockTransactions := block[:len(block)-8]

	sigmaCheck := rsa.VerifyBlockSignature(slotNumber, blockTransactions, hash, sigma, publicKey)
	drawCheck := rsa.VerifyDraw(draw, slotNumber, seed, publicKey)
	if !sigmaCheck {
		fmt.Println("Sigmacheck failed")
		return false
	} else if !drawCheck {
		fmt.Println("Drawcheck failed")
		return false
	}

	toHash := "LOTTERY:" + strconv.Itoa(peer.seed) + ":" + strconv.Itoa(slotNumber) + ":" + publicKey + ":" + draw
	drawHash := Hash(toHash)
	peer.ledger.lock.Lock()
	dolladollabills := big.NewInt(int64(peer.ledger.Accounts[publicKey]))
	peer.ledger.lock.Unlock()
	x := big.NewInt(0)
	x.Mul(dolladollabills, drawHash)
	if x.Cmp(&(peer.hardness)) == 1 {
		peer.ledger.GiveRewardForStake(publicKey, len(blockTransactions))
		fmt.Println("Sigmacheck, drawcheck and hardness success")
		return true
	} else {
		fmt.Println("ticket was SMALLER than hardness")
		return false
	}
}

func (peer *Peer) UpdateLedger(transaction *SignedTransaction) bool {
	success := true
	if transaction.Amount >= 1 && peer.rsa.VerifyTransaction(*transaction) {
		transactionSuccess := peer.ledger.Transaction(transaction)
		if transactionSuccess {
			fmt.Println("Message successfully put in ledger")
			success = true
		} else {
			success = false
			fmt.Println("Invalid transaction, brings account below 0", transaction)
		}
	} else {
		success = false
		fmt.Println("Invalid transaction, did not verify or amount < 0", transaction)
	}
	return success
}

func (peer *Peer) AddChildAndRollbackIfNecessary(newBlockTree *BlockTree, prevhash string) {
	currentLeaf := peer.blockTree.GetLongestChainLeaf()
	if currentLeaf.Node.OwnBlockHash == prevhash {
		currentLeaf.AddChild(newBlockTree)
		//fmt.Println("Rollback not necessary, prev was longest")
		return
	} else {
		peer.blockTree.AddChildAt(newBlockTree, prevhash)
		longest := peer.blockTree.GetLongestChainLeaf()
		if longest != currentLeaf {
			fmt.Println("Rollback was necessary")
			peer.UpdateLedgerOnRollback()
		}
	}
}

func (peer *Peer) UpdateLedgerOnRollback() {
	blocks := peer.blockTree.GetLongestChainOfBlocksAsSlice()
	peer.ledger = peer.genesisLedger
	peer.UpdateLedgerWithSliceOfBlocks(blocks)
	peer.CreateGenesisLedger() //this is because Go does not allow deep copies... *grumble*
}

func (peer *Peer) UpdateLedgerWithSliceOfBlocks(blocks []Block) {
	for _, block := range blocks {
		peer.UpdateLedgerWithBlock(block)
	}
}

func (peer *Peer) UpdateLedgerWithBlock(block Block) bool {
	totalSuccess := true
	for _, transactionID := range block {
		if transactionID == "BLOCK" {
			peer.ledger.Print()
			return totalSuccess
		}
		transactionStruct := peer.messagesSent[transactionID]
		success := peer.UpdateLedger(&(transactionStruct.transaction))
		if success {
			fmt.Println("Ledger was succesfully updated with transaction from block")
		} else {
			totalSuccess = false
			fmt.Println("updating the ledger failed")
			peer.ledger.Print()
		}
	}
	peer.ledger.Print()
	return totalSuccess
}

func (peer *Peer) HandleLottery() {
	slotLength := peer.slotLength
	t := time.NewTicker(time.Duration(slotLength) * time.Second)
	for now := range t.C {
		fmt.Println(" ")
		fmt.Println("Time:", now)
		won, draw := peer.EnterLottery(peer.slotNumber, peer.seed, peer.rsa.n, peer.rsa.d)
		fmt.Println("Slotnumber is:", peer.slotNumber)
		if won {
			fmt.Println("Won slot: ", peer.slotNumber)
			peer.HandleWinning(draw)
		}
		peer.slotNumber += 1
	}
}

func (peer *Peer) EnterLottery(slot int, seed int, n big.Int, d big.Int) (bool, string) {
	toSign := "LOTTERY:" + strconv.Itoa(seed) + ":" + strconv.Itoa(slot)
	draw := peer.rsa.FullSign(toSign, n, d)
	nString := ConvertBigIntToString(&n)
	toHash := "LOTTERY:" + strconv.Itoa(seed) + ":" + strconv.Itoa(slot) + ":" + nString + ":" + ConvertBigIntToString(draw) //entry into lottery
	hashed := Hash(toHash)
	numTickets := big.NewInt(int64(peer.genesisLedger.Accounts[nString]))
	val := big.NewInt(0) //Val(vk, slot, Draw) = accountBalance(vk) * Hash(LOTTERY, Seed, slotnumber, vk, draw), where draw = Sig_sk(LOTTERY, slot)
	val = val.Mul(numTickets, hashed)
	if val.Cmp(&(peer.hardness)) == 1 { //If val >= hardness
		fmt.Println("I WON the lottery in slot: ", slot)
		return true, (ConvertBigIntToString(draw))
	} else {
		fmt.Println("I lost the lottery in slot: ", slot)
		return false, ""
	}
}

func (peer *Peer) HandleWinning(draw string) {
	//append block
	//send out block
	//sends message (BLOCK, vk, slotnumber, Draw, (U,M), hash, sigma=signature of (BLOCK, slot, (U,M), hash))
	peer.nextBlockMutex.Lock()
	block := peer.nextBlock //TODO  - Bør vi tjekke at beskederne i nextBlock ikke er modtaget fra en anden
	blockData := peer.nextBlock
	peer.nextBlock = make([]string, 0)
	peer.nextBlockMutex.Unlock()

	block = append(block, "BLOCK")                            //BLOCK
	block = append(block, ConvertBigIntToString(&peer.rsa.n)) //Public key / VK
	block = append(block, strconv.Itoa(peer.slotNumber))      //Slotnumber
	block = append(block, draw)                               //Draw
	prevBlockHash := peer.getPrevBlockHash()
	block = append(block, prevBlockHash) //Hash
	signature := peer.rsa.CreateBlockSignature(peer.slotNumber, blockData, prevBlockHash)
	block = append(block, signature) //Sigma

	marshalled := peer.MarshalBlock(block)
	peer.SendBlockToAllPeers(marshalled)

}

func (peer *Peer) getPrevBlockHash() string {
	leafTree := peer.blockTree.GetLongestChainLeaf()
	leafNode := leafTree.Node
	return leafNode.OwnBlockHash
}

func (peer *Peer) MakeGenesisBlock() Block {
	block := make([]string, 0)

	key1 := "99220599159528886888088184316939863466036751390102525224276426598372453374970490581931623644823947730183615834970415935110413997957190268925746447986875045619423695530354351164666747197504575160571344765059782114834542464872778036174724267595394424527864340941278010670086469948102379662561982997267164040169642087775263921101619527508747168787150255148601076426931391934490878646150913881272213119308249212668240473054293497445413288931672488505288383846988700744069430536032867719784270610575963716882208428391419536387867514792667272504971046836440153259514944143565853403498339935380048998940846417"
	block = append(block, key1)
	key2 := "85599412879953205917443492336639751512406088883204488437353959378513522416955024172185792293853309086614045424001070998033302348930881810677829310464668788989108423815275401435146750989817827599726082819178311301126706125959538642880756008147465621540684475755420770483545469382177500313066980466028571318051567830522706186098394127053090432052404629932589506015468201259035009870260298981986568167423584202889539017410056236371005761663036714192956723654474436989600592958683850034049398561016771652916907371747535483041614739176853531168948855088182175940234786700494940589098849136266341141384155161"
	block = append(block, key2)
	key3 := "94938669199553053778857680890888139261052515031742833094394381264005413538787479617048724551658947047884930786878545592643341156546341629355279941159238226952308048437799632118321605345240931468890113418679236203299986663053672098258710173978852424933276878715147922334893146561517950444275025644591250775616495714369301297920912844517777635723745293144772908702674201382610931091114936463191905169309350953666108512426366528125453362350408077328372072616595844746964708723653945189478227150662157838762184542167781442195750766257212784191074842194010540524247230081017641149812905846481771040387362473"
	block = append(block, key3)
	key4 := "86452583383266791348634781602421438878698534614119180980968848675882051691784360216437184392986096403545804536088141253936277211321077867008928351117078687752098828115507303231608265820845626805071434231628014647390168434917827717171761542393062938452743852651638618758631815804918108293956355994913095700264201221183679675084999195848794748227364469628612623985152910407381138458527275020855508562816993941567564049116010190501316109460348488259049824369447635049113942952733472839266362177569036225918382937248188277295907694169924415652794045804910426213945443411813621594154348105518818184294346249"
	block = append(block, key4)
	key5 := "81594135348859355889822650216374879117537833325687870348641933990660498172029614843858960160893543004971705050678775284039487254233047107626982118266040228151804438920763314751390066459108316561422333596643117831355245977759296418702636935930016592216482955679278425203161695636496952978681695659108476606655855138949339170964644615516779164639460908768229783291708712591905418163269629394026883090207088407617292418141057959266029941628814934364965311043536457231536747940040915132263532817583457081113622394066389240032727746638150825261278748778368829057825231315703499329107889498851283808455800053"
	block = append(block, key5)
	key6 := "68712351107036437585558705394329932588021453857860002869669357991132586124756852552170712566827012427858101219618342474924223484377006108583647716433567632020410463771552291860494123411766514967952437791952149536595761761768628201669701017186368206738007739214514369690280938959709348367114444332007867429293050870824281366539775693622799966913656661963258346342063688423549783151910273001303698713413753940866803337598653489349374695693083666618952340366801511166307844828148472432093460461261574440793050106855026669923821302882500596809281593084599607212884045166225929286154650189492119251332062599"
	block = append(block, key6)
	key7 := "91273955433896845510861081477211261393472951931840460054284580008859771006155212340140914787944390979120878930868609996748058338377907640267677073912279817976565958447357541072604121964619900931316478594350118528701705448279141426514958466636086268075062290138629777662471099352388822521900394668942098551444827506570358387824412636648945544583564346168203783434336720700601682369105551602717253212251522431865494980399642085639650473281351181921205615546964230228380781298938116442898519184131549614047592945564571080985964305710217520577386421357275728766359526997545444234524045366917234420371495547"
	block = append(block, key7)
	key8 := "93794657355404229319431128515003855336007553691660833870546161779998773890778963918610615793046674621435648393754855465140820703534244156001995273674589296398055231116947487910019524886278710170329803131021053795853655945796124400948014542456370545523186541835638621833119654376852910734839180322005666717956119166548684174165313913116800090123914818313442355435582269314510251304868551601140242896427931023264834388541950156747332715777556673901039316230552233797602769052653802383292709024448454683898134884115648887478942523963719887462236797537388653795357302782037638846557126523009545287101796433"
	block = append(block, key8)
	key9 := "84128649689229141748476650346323678486437898780555562239274803331023504201661721263725702164844226716062121140282102659840393631295501503873123285321391312474481459683725818683741140427611468317891812252761172109050348451968978555634590357968074943405603291955327118302707346701747767186913466136936866428063285478486512344926095099958608452772818854082739431085262847212088429662485943109069166647077422997592232691907767336984413736921170670999915528283134193422301868063392769631810144899315337178298078695275092232062924257321021504333397583129567164544715249346168224769592996908555919759005032463"
	block = append(block, key9)
	key10 := "70621195703417734770343058091658851655361561261137761610884679258670223399663358264109783765671832663936755207486891476183124001771723250631352914064206920117635941024601491293081835825560660689267333988928709259887345879848794348613877724159584075736686419675997111138048781709667966034965470003107432057065409218721592297689386038371608402840880519345471603774550343846256833649393555592852268617261622997674448275202937119980781626022142094513891745877971273140692436832862996188468524311487549280326916150648541763048073345791515098905458134163062720372875194208748523538910421681658158526277952333"
	block = append(block, key10)

	rand.Seed(time.Now().UnixNano())
	seed := rand.Int()
	block = append(block, strconv.Itoa(seed))

	//Calculating hardness
	x := big.NewInt(0)
	x.Exp(big.NewInt(2), big.NewInt(271), nil) //For standard setting (10 peers) hardness should be 30*(2^271)
	x.Mul(x, big.NewInt(2))
	hardness := ConvertBigIntToString(x)

	block = append(block, hardness)
	return block
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

func (peer *Peer) MarshalBlock(block Block) []byte {
	//Signér
	//fmt.Println("Marshalling, signing and appending delimiter to block")
	blockHash := ConvertBigIntToString(Hash(strings.Join(block, ":")))
	block = append(block, blockHash) //This appends the ID to the block (to use in messagesSent)
	block = append(block, "yeet")    //Add delimiter
	bytes, err := json.Marshal(block)
	if err != nil {
		fmt.Println("Marshalling block failed")
	}
	return bytes
}

func (peer *Peer) DemarshalBlock(bytes []byte) (Block, error) {
	var block Block
	err := json.Unmarshal(bytes, &block)
	return block, err
}

func (peer *Peer) GetConnections() []net.Conn {
	return peer.connections
}

func (peer *Peer) AddNewSkUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Setup account:")
	fmt.Println("Make new account = y/yes, preexisting account = n/no, genesis account = 1-10:")
	decision, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	trimmedDecision := strings.TrimRight(decision, "\r\n")
	if trimmedDecision == "y" || trimmedDecision == "yes" {
		newRsa := MakeRSA(2000)
		publicKey := ConvertBigIntToString(&newRsa.n)
		secretKey := ConvertBigIntToString(&newRsa.d)
		success := peer.ledger.AddAccount(publicKey)
		if success {
			fmt.Println("Successfully created new account, this is your secret Key:")
			fmt.Println(secretKey) //notice that both secret and public key are formatted as strings corresponding to the value inside the BigInt, and NOT bytes translated into string from the bigInt.
			fmt.Println("And this is your public name:")
			fmt.Println(publicKey)
		}
	} else if trimmedDecision == "n" || trimmedDecision == "no" {
		fmt.Println("You have chosen to use a preexisting account. Enter the public key:")
		publicKey, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("User quit the program")
			os.Exit(0)
		}
		fmt.Println("This is what you entered:", publicKey)
		fmt.Println("Now enter your secretKey:")
		secretKey, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("User quit the program")
			os.Exit(0)
		}
		fmt.Println("This is what you entered:", secretKey)
		peer.rsa = MakeRSAWithKeys(publicKey, secretKey)
		fmt.Println("You can now make transactions using your public and secret key.")
		fmt.Println("---------------------------------------------------------------------------------------------------------")
	} else if trimmedDecision == "1" {
		peer.rsa = MakeRSAWithKeys("99220599159528886888088184316939863466036751390102525224276426598372453374970490581931623644823947730183615834970415935110413997957190268925746447986875045619423695530354351164666747197504575160571344765059782114834542464872778036174724267595394424527864340941278010670086469948102379662561982997267164040169642087775263921101619527508747168787150255148601076426931391934490878646150913881272213119308249212668240473054293497445413288931672488505288383846988700744069430536032867719784270610575963716882208428391419536387867514792667272504971046836440153259514944143565853403498339935380048998940846417", "66147066106352591258725456211293242310691167593401683482850951065581635583313660387954415763215965153455743889980277290073609331971460179283830965324583363746282463686902900776444498131669716773714229843373188076556361643248518690783149511730262949685242893960852007113390979965401586441707988664844762745508117637359502895110544591891447889581986360912541538288351271783269725834355782179249158683359773911535600994935883950084494585598131890180600348693060660169070167869890819212774506630114016303382665745688834796844837019895605694433162313497206230710528213636589702842149997136746634853051713867")
		fmt.Println("You now use GenKey 1!")
	} else if trimmedDecision == "2" {
		peer.rsa = MakeRSAWithKeys("85599412879953205917443492336639751512406088883204488437353959378513522416955024172185792293853309086614045424001070998033302348930881810677829310464668788989108423815275401435146750989817827599726082819178311301126706125959538642880756008147465621540684475755420770483545469382177500313066980466028571318051567830522706186098394127053090432052404629932589506015468201259035009870260298981986568167423584202889539017410056236371005761663036714192956723654474436989600592958683850034049398561016771652916907371747535483041614739176853531168948855088182175940234786700494940589098849136266341141384155161", "57066275253302137278295661557759834341604059255469658958235972919009014944636682781457194862568872724409363616000713998688868232620587873785219540309779192659405615876850267623431167326545218399817388546118874200751137417306359095253837338764977081027122983836947180322363646254785000208711320310685701811559846153402328139644121775388847982512106893024434013483758676091513272361707236043468694775351691990371111047305227069747581372080012677826929336064868485779336318206441324807451529044360430322327105045986377923610860674048722625908251886363769200610782611957129406730807694517303759791724147883")
		fmt.Println("You now use GenKey 2!")
	} else if trimmedDecision == "3" {
		peer.rsa = MakeRSAWithKeys("94938669199553053778857680890888139261052515031742833094394381264005413538787479617048724551658947047884930786878545592643341156546341629355279941159238226952308048437799632118321605345240931468890113418679236203299986663053672098258710173978852424933276878715147922334893146561517950444275025644591250775616495714369301297920912844517777635723745293144772908702674201382610931091114936463191905169309350953666108512426366528125453362350408077328372072616595844746964708723653945189478227150662157838762184542167781442195750766257212784191074842194010540524247230081017641149812905846481771040387362473", "63292446133035369185905120593925426174035010021161888729596254176003609025858319744699149701105964698589953857919030395095560771030894419570186627439492151301538698958533088078881070230160620979260075612452824135533324442035781398839140115985901616622184585810098614889928764374345300296183350429727487513247907826811332243853037976223252383517850387930734758706368641950566574540183286993467902730480342892768574995901365843993484181587657388877051442097120415616997954645795326298701270004588959437930315518222758790083847871266845126016838669344583388972271861405121067179022340927614093613385620747")
		fmt.Println("You now use GenKey 3!")
	} else if trimmedDecision == "4" {
		peer.rsa = MakeRSAWithKeys("86452583383266791348634781602421438878698534614119180980968848675882051691784360216437184392986096403545804536088141253936277211321077867008928351117078687752098828115507303231608265820845626805071434231628014647390168434917827717171761542393062938452743852651638618758631815804918108293956355994913095700264201221183679675084999195848794748227364469628612623985152910407381138458527275020855508562816993941567564049116010190501316109460348488259049824369447635049113942952733472839266362177569036225918382937248188277295907694169924415652794045804910426213945443411813621594154348105518818184294346249", "57635055588844527565756521068280959252465689742746120653979232450588034461189573477624789595324064269030536357392094169290851474214051911339285567411385791834732552077004868821072177213897084536714289487752009764926778956611885144781174361595375292301829235101092412505754543869945405529304237329942051377748136662340932984463101236766809521029118372315113914738854726914491432436095678435744369554205214824941855772892744662093065231091360991479412976539166537893744224448305815567518191356700827230606962081724820505633499922128887121521862665873647845984096294415483390872767806744329110111292872707")
		fmt.Println("You now use GenKey 4!")
	} else if trimmedDecision == "5" {
		peer.rsa = MakeRSAWithKeys("81594135348859355889822650216374879117537833325687870348641933990660498172029614843858960160893543004971705050678775284039487254233047107626982118266040228151804438920763314751390066459108316561422333596643117831355245977759296418702636935930016592216482955679278425203161695636496952978681695659108476606655855138949339170964644615516779164639460908768229783291708712591905418163269629394026883090207088407617292418141057959266029941628814934364965311043536457231536747940040915132263532817583457081113622394066389240032727746638150825261278748778368829057825231315703499329107889498851283808455800053", "54396090232572903926548433477583252745025222217125246899094622660440332114686409895905973440595695336647803367119183522692991502822031405084654745510693485434536292613842209834260044306072211040948222397762078554236830651839530945801757957286677728144321970452852283468774463757664635319121130439405638970909854179038382504969020111450170410451417348364551804794839643872324455490156426963773821964652385519393075073646193260017039391994317289516698863165631189739683305288120447200424391498376646686085100647822777706297051468474204958854675868534108069068727135675736578916481170092283193250304079147")
		fmt.Println("You now use GenKey 5!")
	} else if trimmedDecision == "6" {
		peer.rsa = MakeRSAWithKeys("68712351107036437585558705394329932588021453857860002869669357991132586124756852552170712566827012427858101219618342474924223484377006108583647716433567632020410463771552291860494123411766514967952437791952149536595761761768628201669701017186368206738007739214514369690280938959709348367114444332007867429293050870824281366539775693622799966913656661963258346342063688423549783151910273001303698713413753940866803337598653489349374695693083666618952340366801511166307844828148472432093460461261574440793050106855026669923821302882500596809281593084599607212884045166225929286154650189492119251332062599", "45808234071357625057039136929553288392014302571906668579779571994088390749837901701447141711218008285238734146412228316616148989584670739055765144289045088013606975847701527906996082274511009978634958527968099691063841174512418801113134011457578804492005159476342913126853959306472898911409629554671900562891662517910310816976048692788121342291618014156450442211850894825573035111466101918246333531693242849867852990861822107357625185920032273742245112863764196731704101713955660553343809048635685453501018233883060616462317688347186989653622928858407676200919168956518687164103793355239275219789957227")
		fmt.Println("You now use GenKey 6!")
	} else if trimmedDecision == "7" {
		peer.rsa = MakeRSAWithKeys("91273955433896845510861081477211261393472951931840460054284580008859771006155212340140914787944390979120878930868609996748058338377907640267677073912279817976565958447357541072604121964619900931316478594350118528701705448279141426514958466636086268075062290138629777662471099352388822521900394668942098551444827506570358387824412636648945544583564346168203783434336720700601682369105551602717253212251522431865494980399642085639650473281351181921205615546964230228380781298938116442898519184131549614047592945564571080985964305710217520577386421357275728766359526997545444234524045366917234420371495547", "60849303622597897007240720984807507595648634621226973369523053339239847337436808226760609858629593986080585953912406664498705558918605093511784715941519878651043972298238360715069414643079933954210985729566745685801136965519427617676638977757390845383374860092419851774980732901592548347933596445961386267208512158936772511903756932088854615177782343747050024542617812991958723910140318864835765370320427471027760923989513917970228497157097274228721872030968193465764382197457306647061271751006469270808021101220186242685745944137832107706962543049504351247550396764180885509402107674673111483343461787")
		fmt.Println("You now use GenKey 7!")
	} else if trimmedDecision == "8" {
		peer.rsa = MakeRSAWithKeys("93794657355404229319431128515003855336007553691660833870546161779998773890778963918610615793046674621435648393754855465140820703534244156001995273674589296398055231116947487910019524886278710170329803131021053795853655945796124400948014542456370545523186541835638621833119654376852910734839180322005666717956119166548684174165313913116800090123914818313442355435582269314510251304868551601140242896427931023264834388541950156747332715777556673901039316230552233797602769052653802383292709024448454683898134884115648887478942523963719887462236797537388653795357302782037638846557126523009545287101796433", "62529771570269486212954085676669236890671702461107222580364107853332515927185975945740410528697783080957098929169903643427213802356162770667996849116392864265370154077964991940013016590852473446886535420680702530569103963864082933965343028304247030348791027890425747888746436251235273823226120214670431563924686410594852264011717447911867965421792457711609768721398045113825769696270387407334328842613191390579920555424156978849748835588041653519311851435698539337440204744765491196153239257941728858976517501464047535513082183750917382887638010506211749109922606527099218176875462650114570015548353827")
		fmt.Println("You now use GenKey 8!")
	} else if trimmedDecision == "9" {
		peer.rsa = MakeRSAWithKeys("84128649689229141748476650346323678486437898780555562239274803331023504201661721263725702164844226716062121140282102659840393631295501503873123285321391312474481459683725818683741140427611468317891812252761172109050348451968978555634590357968074943405603291955327118302707346701747767186913466136936866428063285478486512344926095099958608452772818854082739431085262847212088429662485943109069166647077422997592232691907767336984413736921170670999915528283134193422301868063392769631810144899315337178298078695275092232062924257321021504333397583129567164544715249346168224769592996908555919759005032463", "56085766459486094498984433564215785657625265853703708159516535554015669467774480842483801443229484477374747426854735106560262420863667669248748856880927541649654306455817212455827426951740978878594541501840781406033565634645985703756393571978716628937068861303551412201804897801165178124608977424624565383660090020108279414641760356014618805194026664668950447058660780645842250453168342552799489026854312248658849391273762450578714474116884341495389211345085680851007877682232233392631306652138191781996971108044388337365555941608247827596381515834264967330548744937537617112889038432355010227570468939")
		fmt.Println("You now use GenKey 9!")
	} else {
		peer.rsa = MakeRSAWithKeys("70621195703417734770343058091658851655361561261137761610884679258670223399663358264109783765671832663936755207486891476183124001771723250631352914064206920117635941024601491293081835825560660689267333988928709259887345879848794348613877724159584075736686419675997111138048781709667966034965470003107432057065409218721592297689386038371608402840880519345471603774550343846256833649393555592852268617261622997674448275202937119980781626022142094513891745877971273140692436832862996188468524311487549280326916150648541763048073345791515098905458134163062720372875194208748523538910421681658158526277952333", "47080797135611823180228705394439234436907707507425174407256452839113482266442238842739855843781221775957836804991260984122082667847815500420901942709471280078423960683067660862054557217040440459511555992619139506591563919899196232409251816106389383824457613117331407425365854473111977356643646668738276832537633132845787805744042592646044493003311754064041832749213975665596294833811787512683785544903252943781383044290811821447018358747441605357665865259282762476668907306423923701622064148912823436616682105409323617355598360462566986128445001798326379207490257344784110525958297045948387829570574059")
		fmt.Println("You now use GenKey 10! (Maybe you asked for that, maybe I'm just nice)")
	}
}
