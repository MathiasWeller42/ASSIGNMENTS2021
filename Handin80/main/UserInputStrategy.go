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

//Main Strategy which takes input from command line and creates a new SignedTransaction from it
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

//FixedInputStrategy takes a transaction as input to its constructor and outputs this transaction when HandleIncomingFromUser() is called
type FixedInputStrategy struct {
	currentInput SignedTransaction
}

func MakeFixedInputStrategy(input SignedTransaction) *FixedInputStrategy {
	inputStrategy := new(FixedInputStrategy)
	inputStrategy.currentInput = input
	return inputStrategy
}

func (inputStrategy *FixedInputStrategy) HandleIncomingFromUser() SignedTransaction {
	return inputStrategy.currentInput
}

//FixedInputStrategy2 creates a new transaction with the given parameters each time HandleIncomingFromUser() is called
type FixedInputStrategy2 struct {
	currentInput SignedTransaction
	publicKey    string
	secretKey    string
	amount       int
	to           string
}

func MakeFixedInputStrategy2(input SignedTransaction, publicKey string, secretKey string) (*FixedInputStrategy2, string) {
	inputStrategy := new(FixedInputStrategy2)
	inputStrategy.currentInput = input
	inputStrategy.publicKey = publicKey
	inputStrategy.secretKey = secretKey
	inputStrategy.amount = input.Amount
	inputStrategy.to = input.To
	return inputStrategy, inputStrategy.publicKey
}

func (inputStrategy *FixedInputStrategy2) HandleIncomingFromUser() SignedTransaction {
	transaction := MakeSignedTransaction(inputStrategy.publicKey, inputStrategy.to, inputStrategy.amount, inputStrategy.secretKey)
	inputStrategy.currentInput = *transaction
	return inputStrategy.currentInput
}
