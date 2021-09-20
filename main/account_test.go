package main

import (
	"fmt"
	"testing"
)

func TestShouldCreateLedger(t *testing.T) {
	ledger := MakeLedger()

	if len(ledger.Accounts) != 0 {
		t.Errorf("Accounts should be empty initially")
	} else {
		fmt.Println("Ledger test TestShouldCreateLedger passed")
	}
}

func TestShouldCreateTransaction(t *testing.T) {
	transaction := MakeTransaction("from", "to", 0)

	if transaction.ID != "id" {
		t.Errorf("ID not initialized correctly")
	}
	if transaction.From != "from" {
		t.Errorf("From not initialized correctly")
	}
	if transaction.To != "to" {
		t.Errorf("To not initialized correctly")
	}
	if transaction.Amount != 0 {
		t.Errorf("Amount not initialized correctly")
	} else {
		fmt.Println("Ledger test TestShouldCreateTransaction passed")
	}
}

func TestShouldAddAccountToLedger(t *testing.T) {
	ledger := MakeLedger()
	ledger.Accounts["acc1"] = 200
	ledger.Accounts["acc2"] = 450

	if len(ledger.Accounts) != 2 {
		t.Errorf("Accounts not added or added too many")
	}
	if ledger.Accounts["acc1"] != 200 {
		t.Errorf("acc1 does not exist or has wrong balance")
	}
	if ledger.Accounts["acc2"] != 450 {
		t.Errorf("acc2 does not exist or has wrong balance")
	} else {
		fmt.Println("Ledger test TestShouldAddAccountToLedger passed")
	}
}

func TestShouldMoveMoneyFromAccountToAccountOnTransaction(t *testing.T) {
	ledger := MakeLedger()
	ledger.Accounts["acc1"] = 200
	ledger.Accounts["acc2"] = 200
	transaction := MakeTransaction("acc1", "acc2", 100)
	ledger.Transaction(transaction)
	accountBalancesCorrect := ledger.Accounts["acc1"] == 100 && ledger.Accounts["acc2"] == 300

	if !accountBalancesCorrect {
		t.Errorf("Account balances should be correct")
	} else {
		fmt.Println("Ledger test TestShouldMoveMoneyFromAccountToAccountOnTransaction passed")
	}
}

func TestShouldMoveMoneyOnNewAccounts(t *testing.T) {
	ledger := MakeLedger()
	transaction := MakeTransaction("acc1", "acc2", 100)
	ledger.Transaction(transaction)
	accountBalancesCorrect := ledger.Accounts["acc1"] == -100 && ledger.Accounts["acc2"] == 100

	if !accountBalancesCorrect {
		t.Errorf("Account balances should be correct")
	} else {
		fmt.Println("Ledger test TestShouldMoveMoneyOnNewAccounts passed")
	}
}
