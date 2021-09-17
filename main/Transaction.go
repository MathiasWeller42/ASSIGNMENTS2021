package main

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

func MakeTransaction(id string, from string, to string, amount int) *Transaction {
	transaction := new(Transaction)
	transaction.ID = id
	transaction.From = from
	transaction.To = to
	transaction.Amount = amount
	return transaction
}
