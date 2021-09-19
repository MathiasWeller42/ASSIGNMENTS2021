package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type UserInputStrategy interface {
	HandleIncomingFromUser() Transaction
}

type CommandLineUserInputStrategy struct {
}

func (inputStrategy *CommandLineUserInputStrategy) HandleIncomingFromUser() Transaction {

	//prompt user to type a message
	fmt.Println("Type an ID:")
	reader := bufio.NewReader(os.Stdin)
	id, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("User quit the program")
		os.Exit(0)
	}
	fmt.Println("Type account 1:")
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
	val, err := strconv.Atoi(amount)
	if err != nil {
		fmt.Println("Wrong conversion to int")
		val = 123
	}
	return *MakeTransaction(id, acc1, acc2, val)
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
