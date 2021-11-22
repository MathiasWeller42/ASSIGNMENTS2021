package main

import (
	"fmt"
	"strings"
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

func (l *Ledger) Transaction(t *SignedTransaction) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	from := strings.TrimRight(t.From, "\r\n")
	to := strings.TrimRight(t.To, "\r\n")
	//check whether accounts exist, otherwise create them
	fromBalance, from_exists := l.Accounts[from]
	if !from_exists {
		fmt.Println("Adding new account 'from'")
		l.Accounts[from] = 0
		fromBalance = 0
	}
	_, to_exists := l.Accounts[to]
	if !to_exists {
		fmt.Println("Adding new account 'to'")
		l.Accounts[to] = 0
	}

	if fromBalance >= t.Amount {
		l.Accounts[from] -= t.Amount
		l.Accounts[to] += (t.Amount - 1)
		return true
	} else {
		fmt.Println("Oh no, fromBalance is ", fromBalance, ", while the amount to pay is ", t.Amount, ", from account:", from)
		return false
	}
}

func (l *Ledger) Print() {
	l.lock.Lock()
	defer l.lock.Unlock()

	fmt.Println("Ledger state:")
	for acc, balance := range l.Accounts {
		fmt.Println("Account", acc[:3], "has balance", balance, "AU")
	}
}

func (l *Ledger) GiveRewardForStake(publicKey string, noOfTransactions int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	realPK := strings.TrimRight(publicKey, "\r\n")
	reward := 10 + noOfTransactions
	fmt.Println("Awarding the block creator", reward, "AU for that amount of transactions in block + 10")

	l.Accounts[realPK] += reward
}

func (l *Ledger) AddAccount(newAcc string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	realAcc := strings.TrimRight(newAcc, "\r\n")
	_, from_exists := l.Accounts[realAcc]
	if !from_exists {
		l.Accounts[realAcc] = 0
		return true
	}
	return false
}

func (l *Ledger) AddGenesisAccount(newAcc string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	realAcc := strings.TrimRight(newAcc, "\r\n")
	_, from_exists := l.Accounts[realAcc]
	if !from_exists {
		l.Accounts[realAcc] = 1000000
		return true
	}
	return false
}
