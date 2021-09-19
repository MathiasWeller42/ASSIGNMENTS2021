package main

import (
	"fmt"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	//check whether accounts exist, otherwise create them
	_, from_exists := l.Accounts[t.From]
	if !from_exists {
		l.Accounts[t.From] = 0
	}
	_, to_exists := l.Accounts[t.To]
	if !to_exists {
		l.Accounts[t.To] = 0
	}

	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

func (l *Ledger) Print() {
	l.lock.Lock()
	defer l.lock.Unlock()

	fmt.Println("Ledger state:")
	for acc, balance := range l.Accounts {
		fmt.Println("Account", acc, "has balance", balance)
	}
}
