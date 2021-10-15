package main

import "fmt"

type MessageSendingStrategy interface {
	SendMessageToAllPeers(message SignedTransaction, peer *Peer)
}

type RealMessageSendingStrategy struct {
}

func (realMessageSendingStrategy *RealMessageSendingStrategy) SendMessageToAllPeers(message SignedTransaction, peer *Peer) {
	for _, connection := range peer.connections {
		fmt.Println("Sending", message, "to", connection)
		peer.SendMessage(connection, message)
	}
}

type StubbedMessageSendingStrategy struct {
	messagesSent MessagesSent
}

func (stubbedMessageSendingStrategy *StubbedMessageSendingStrategy) SendMessageToAllPeers(message SignedTransaction, peer *Peer) {
	stubbedMessageSendingStrategy.messagesSent[message.ID] = true
}

func MakeStubbedMessageSendingStrategy() *StubbedMessageSendingStrategy {
	messageSendingStrategy := new(StubbedMessageSendingStrategy)
	messageSendingStrategy.messagesSent = make(map[string]bool)
	return messageSendingStrategy
}
