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
	fmt.Println("Type account 1:") //TODO should ask for a secret key - however, we need a lookup for the corresponding secretkey
	acc1, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("Type account 2:")
	acc2, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("Type amount:")
	amount, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("The amount before conversion: ", amount)
	trimmed := strings.TrimSpace(amount)
	val, err := strconv.Atoi(trimmed)
	if err != nil {
		fmt.Println("Wrong conversion to int")
		val = 123
	}
	fmt.Println("Type private key:")
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
