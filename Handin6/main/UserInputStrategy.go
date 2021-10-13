package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UserInputStrategy interface {
	HandleIncomingFromUser() SignedTransaction
}

type CommandLineUserInputStrategy struct {
}

func (inputStrategy *CommandLineUserInputStrategy) HandleIncomingFromUser() SignedTransaction {

	//prompt user to type a message
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Make a new transaction:")
	fmt.Println("Type 'From' account (public key):")
	acc1, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("Type 'To' account (public key):")
	acc2, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("Type amount (more than 0):")
	amount, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("The amount before conversion: ", amount)
	trimmed := strings.TrimSpace(amount)
	val, err := strconv.Atoi(trimmed)
	if err != nil {
		fmt.Println("Wrong conversion to int, setting the value to -1 (invalid message)")
		val = -1
	}
	fmt.Println("Type your secret key:")
	privateKey, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	return *MakeSignedTransaction(acc1, acc2, val, privateKey)
}

type FixedInputStrategy struct {
	currentInput SignedTransaction
}

func (inputStrategy *FixedInputStrategy) HandleIncomingFromUser() SignedTransaction {
	return inputStrategy.currentInput
}

func MakeFixedInputStrategy(input SignedTransaction) *FixedInputStrategy {
	inputStrategy := new(FixedInputStrategy)
	inputStrategy.currentInput = input
	return inputStrategy
}
