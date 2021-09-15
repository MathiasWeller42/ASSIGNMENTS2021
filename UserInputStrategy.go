package main

import (
	"bufio"
	"fmt"
	"os"
)

type UserInputStrategy interface {
	HandleIncomingFromUser() string
}

type CommandLineUserInputStrategy struct {
}

func (inputStrategy *CommandLineUserInputStrategy) HandleIncomingFromUser() string {

	//prompt user to type a message
	fmt.Println("Type a message:")
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	return msg
}

type FixedInputStrategy struct {
	currentInput string
}

func (inputStrategy *FixedInputStrategy) HandleIncomingFromUser() string {
	return inputStrategy.currentInput
}

func MakeFixedInputStrategy(input string) *FixedInputStrategy {
	inputStrategy := new(FixedInputStrategy)
	inputStrategy.currentInput = input
	return inputStrategy
}
