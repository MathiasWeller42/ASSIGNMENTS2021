package main

type MessageSendingStrategy interface {
	SendMessageToAllPeers(message Transaction, peer *Peer)
}

type RealMessageSendingStrategy struct {
}

func (realMessageSendingStrategy *RealMessageSendingStrategy) SendMessageToAllPeers(message Transaction, peer *Peer) {
	for _, connection := range peer.Connections {
		peer.SendMessage(connection, message)
	}
}

type StubbedMessageSendingStrategy struct {
	messagesSent MessagesSent
}

func (stubbedMessageSendingStrategy *StubbedMessageSendingStrategy) SendMessageToAllPeers(message Transaction, peer *Peer) {
	stubbedMessageSendingStrategy.messagesSent[message] = true
}

func MakeStubbedMessageSendingStrategy() *StubbedMessageSendingStrategy {
	messageSendingStrategy := new(StubbedMessageSendingStrategy)
	messageSendingStrategy.messagesSent = make(map[Transaction]bool)
	return messageSendingStrategy
}
