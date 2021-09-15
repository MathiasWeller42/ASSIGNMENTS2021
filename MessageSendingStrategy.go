package main

type MessageSendingStrategy interface {
	SendMessageToAllPeers(message string, peer *Peer)
}

type RealMessageSendingStrategy struct {
}

func (realMessageSendingStrategy *RealMessageSendingStrategy) SendMessageToAllPeers(message string, peer *Peer) {
	for connection, _ := range peer.connections {
		peer.SendMessage(connection, message)
	}
}

type StubbedMessageSendingStrategy struct {
	messagesSent MessagesSent
}

func (stubbedMessageSendingStrategy *StubbedMessageSendingStrategy) SendMessageToAllPeers(message string, peer *Peer) {
	stubbedMessageSendingStrategy.messagesSent[message] = true
}

func MakeStubbedMessageSendingStrategy() *StubbedMessageSendingStrategy {
	messageSendingStrategy := new(StubbedMessageSendingStrategy)
	return messageSendingStrategy
}
