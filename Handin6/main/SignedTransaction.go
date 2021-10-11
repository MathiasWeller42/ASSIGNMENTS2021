package main

import (
	"math/rand"
	"strconv"
	"time"
)

type SignedTransaction struct {
	ID        string
	From      string
	To        string
	Signature string
	Amount    int
}

func MakeSignedTransaction(from string, to string, amount int, keyD string) *SignedTransaction {
	transaction := new(SignedTransaction)
	rand.Seed(time.Now().UnixNano())
	integer := rand.Int()
	transaction.ID = strconv.Itoa(integer)
	transaction.From = from
	transaction.To = to
	transaction.Amount = amount
	keyN := transaction.From
	rsa := MakeRSA(2000)
	rsa.FullSignTransaction(transaction, keyN, keyD)
	return transaction
}
