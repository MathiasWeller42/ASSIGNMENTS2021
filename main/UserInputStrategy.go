package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UserInputStrategy interface {
	HandleIncomingFromUser() Transaction
}

type CommandLineUserInputStrategy struct {
}

func (inputStrategy *CommandLineUserInputStrategy) HandleIncomingFromUser() Transaction {

	//prompt user to type a message
	reader := bufio.NewReader(os.Stdin)
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
	fmt.Println("The amount before conversion: ", amount)
	trimmed := strings.TrimSpace(amount)
	val, err := strconv.Atoi(trimmed)
	if err != nil {
		fmt.Println("Wrong conversion to int")
		val = 123
	}
	return *MakeTransaction(acc1, acc2, val)
}

type FixedInputStrategy struct {
	currentInput Transaction
}

func (inputStrategy *FixedInputStrategy) HandleIncomingFromUser() Transaction {
	return inputStrategy.currentInput
}

func MakeFixedInputStrategy(input Transaction) *FixedInputStrategy {
	inputStrategy := new(FixedInputStrategy)
	inputStrategy.currentInput = input
	return inputStrategy
}
