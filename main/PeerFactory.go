package main

type PeerFactory interface {
	createOutBoundIPStrategy() OutboundIPStrategy
	createMessageSendingStrategy() MessageSendingStrategy
	createUserInputStrategy() UserInputStrategy
	createUriStrategy() UriStrategy
}

type RealPeerFactory struct {
}

func (realPeerFactory *RealPeerFactory) createOutboundIPStrategy() OutboundIPStrategy {
	return new(RealOutboundIPStrategy)
}

func (realPeerFactory *RealPeerFactory) createMessageSendingStrategy() MessageSendingStrategy {
	return new(RealMessageSendingStrategy)
}

func (realPeerFactory *RealPeerFactory) createUserInputStrategy() UserInputStrategy {
	return new(CommandLineUserInputStrategy)

}

func (realPeerFactory *RealPeerFactory) createUriStrategy() UriStrategy {
	return new(CommandLineUriStrategy)
}
