package main

import (
	"fmt"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     *sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	ledger.lock = &sync.Mutex{}
	return ledger
}

func (l *Ledger) Transaction(t *SignedTransaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	//check whether accounts exist, otherwise create them
	_, from_exists := l.Accounts[t.From]
	if !from_exists {
		l.Accounts[t.From] = 1000
	}
	_, to_exists := l.Accounts[t.To]
	if !to_exists {
		l.Accounts[t.To] = 1000
	}

	//check that from account has enough money, otherwise do nothing
	if l.Accounts[t.From] < t.Amount {
		fmt.Println("Rejected transaction that makes from account negative")
		return
	}

	//move the money
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

func (l *Ledger) AddAccount(newAcc string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	_, from_exists := l.Accounts[newAcc]
	if !from_exists {
		l.Accounts[newAcc] = 1000
		return true
	}
	return false
}
