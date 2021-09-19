package main

import (
	"math/rand"
	"strconv"
	"time"
)

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

func MakeTransaction(from string, to string, amount int) *Transaction {
	transaction := new(Transaction)
	rand.Seed(time.Now().UnixNano())
	integer := rand.Int()
	transaction.ID = strconv.Itoa(integer)
	transaction.From = from
	transaction.To = to
	transaction.Amount = amount
	return transaction
}
