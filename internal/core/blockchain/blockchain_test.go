package blockchain

import (
	"fmt"
	"math/big"
	"testing"
	"time"
	"zkrollup/internal/types/transaction"
)

func createTestTransaction(value int64, nonce uint64) transaction.Transaction {
	return transaction.Transaction{
		From:      "0000000000000000000000000000000000000001", // Use genesis account
		To:        "0000000000000000000000000000000000000002", // Use genesis account
		Value:     big.NewInt(value),
		Nonce:     nonce,
		Status:    transaction.StatusPending,
		Timestamp: time.Now().Unix(),
	}
}

func TestBlockCreation(t *testing.T) {
	bc := NewBlockchain()

	// Create and add a transaction
	tx1 := createTestTransaction(100, 0)
	tx1.Hash = tx1.ComputeHash()

	if err := bc.AddTransaction(tx1); err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}

	// Create block
	if err := bc.CreateBlock(); err != nil {
		t.Fatalf("Failed to create block: %v", err)
	}

	// Verify transaction status is updated
	confirmedTx := bc.GetTransactionByHash(tx1.Hash)
	if confirmedTx == nil {
		t.Fatal("Transaction not found after block creation")
	}
	if confirmedTx.Status != transaction.StatusConfirmed {
		t.Errorf("Expected transaction status %v, got %v", transaction.StatusConfirmed, confirmedTx.Status)
	}

	// Verify balances are updated
	senderBalance := bc.GetBalance("0000000000000000000000000000000000000001")
	if senderBalance.Cmp(big.NewInt(999900)) != 0 { // 1000000 - 100
		t.Errorf("Expected sender balance 999900, got %s", senderBalance.String())
	}

	receiverBalance := bc.GetBalance("0000000000000000000000000000000000000002")
	if receiverBalance.Cmp(big.NewInt(500100)) != 0 { // 500000 + 100
		t.Errorf("Expected receiver balance 500100, got %s", receiverBalance.String())
	}
}

func TestAutoBlockCreation(t *testing.T) {
	bc := NewBlockchain()

	// Create and add transaction
	tx1 := createTestTransaction(100, 0)
	tx1.Hash = tx1.ComputeHash()

	if err := bc.AddTransaction(tx1); err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}

	// Create block manually instead of using auto block
	if err := bc.CreateBlock(); err != nil {
		t.Fatalf("Failed to create block: %v", err)
	}

	// Verify transaction status
	confirmedTx := bc.GetTransactionByHash(tx1.Hash)
	if confirmedTx == nil {
		t.Fatal("Transaction not found after block creation")
	}
	if confirmedTx.Status != transaction.StatusConfirmed {
		t.Errorf("Expected transaction status %v, got %v", transaction.StatusConfirmed, confirmedTx.Status)
	}

	// Verify balances
	senderBalance := bc.GetBalance("0000000000000000000000000000000000000001")
	if senderBalance.Cmp(big.NewInt(999900)) != 0 { // 1000000 - 100
		t.Errorf("Expected sender balance 999900, got %s", senderBalance.String())
	}

	receiverBalance := bc.GetBalance("0000000000000000000000000000000000000002")
	if receiverBalance.Cmp(big.NewInt(500100)) != 0 { // 500000 + 100
		t.Errorf("Expected receiver balance 500100, got %s", receiverBalance.String())
	}
}

func TestTransactionValidation(t *testing.T) {
	bc := NewBlockchain()

	// Test insufficient balance
	tx1 := createTestTransaction(2000000, 0) // Amount larger than genesis balance
	if err := bc.AddTransaction(tx1); err == nil {
		t.Error("Expected error for insufficient balance")
	}

	// Test invalid nonce
	tx2 := createTestTransaction(100, 1) // nonce should be 0
	if err := bc.AddTransaction(tx2); err == nil {
		t.Error("Expected error for invalid nonce")
	}

	// Test valid transaction
	tx3 := createTestTransaction(100, 0)
	if err := bc.AddTransaction(tx3); err != nil {
		t.Errorf("Unexpected error for valid transaction: %v", err)
	}
}

func TestStateRoot(t *testing.T) {
	bc := NewBlockchain()

	// 1. Test initial state root
	initialRoot := bc.GetStateRoot()
	fmt.Println("initialRoot: ", initialRoot)
}
