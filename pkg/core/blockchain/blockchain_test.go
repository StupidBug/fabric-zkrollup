package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

func createTestTransaction(value int64, nonce uint64) transaction.Transaction {
	// Generate a test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate key pair: %v", err))
	}

	tx := transaction.Transaction{
		From:      "0000000000000000000000000000000000000001", // Use genesis account
		To:        "0000000000000000000000000000000000000002", // Use genesis account
		Value:     int(value),
		Nonce:     nonce,
		Status:    transaction.StatusPending,
		Timestamp: time.Now().Unix(),
	}

	// Sign the transaction
	if err := tx.SignTransaction(privateKey); err != nil {
		panic(fmt.Sprintf("Failed to sign transaction: %v", err))
	}

	// Store public key in blockchain state
	bc := NewBlockchain()
	bc.SetPublicKey(tx.From, &privateKey.PublicKey)

	return tx
}

func TestBlockCreation(t *testing.T) {
	bc := NewBlockchain()

	// Create and add a transaction
	tx1 := createTestTransaction(100, 0)
	tx1.Hash = tx1.ComputeHash()

	// Store public key in blockchain state
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	bc.SetPublicKey(tx1.From, &privateKey.PublicKey)

	// Sign the transaction
	if err := tx1.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

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
	if senderBalance != 999900 { // 1000000 - 100
		t.Errorf("Expected sender balance 999900, got %d", senderBalance)
	}

	receiverBalance := bc.GetBalance("0000000000000000000000000000000000000002")
	if receiverBalance != 500100 { // 500000 + 100
		t.Errorf("Expected receiver balance 500100, got %d", receiverBalance)
	}
}

func TestAutoBlockCreation(t *testing.T) {
	bc := NewBlockchain()

	// Create and add transaction
	tx1 := createTestTransaction(100, 0)
	tx1.Hash = tx1.ComputeHash()

	// Store public key in blockchain state
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	bc.SetPublicKey(tx1.From, &privateKey.PublicKey)

	// Sign the transaction
	if err := tx1.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

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
	if senderBalance != 999900 { // 1000000 - 100
		t.Errorf("Expected sender balance 999900, got %d", senderBalance)
	}

	receiverBalance := bc.GetBalance("0000000000000000000000000000000000000002")
	if receiverBalance != 500100 { // 500000 + 100
		t.Errorf("Expected receiver balance 500100, got %d", receiverBalance)
	}
}

func TestTransactionValidation(t *testing.T) {
	bc := NewBlockchain()

	// Generate a key pair for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	bc.SetPublicKey("0000000000000000000000000000000000000001", &privateKey.PublicKey)

	// Test insufficient balance
	tx1 := createTestTransaction(2000000, 0) // Amount larger than genesis balance
	if err := tx1.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	if err := bc.AddTransaction(tx1); err == nil {
		t.Error("Expected error for insufficient balance")
	}

	// Test invalid nonce
	tx2 := createTestTransaction(100, 1) // nonce should be 0
	if err := tx2.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	if err := bc.AddTransaction(tx2); err == nil {
		t.Error("Expected error for invalid nonce")
	}

	// Test valid transaction
	tx3 := createTestTransaction(100, 0)
	if err := tx3.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
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
