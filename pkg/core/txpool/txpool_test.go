package txpool

import (
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

func createTestTransaction(value int64, nonce uint64) transaction.Transaction {
	return transaction.Transaction{
		From:      "sender",
		To:        "receiver",
		Value:     int(value),
		Nonce:     nonce,
		Status:    transaction.StatusPending,
		Timestamp: time.Now().Unix(),
	}
}

func TestTxPoolAdd(t *testing.T) {
	pool := NewTxPool()
	tx1 := createTestTransaction(1000, 0)
	tx1.Hash = tx1.ComputeHash()

	// Test adding transaction
	pool.Add(tx1)
	if pool.Size() != 1 {
		t.Errorf("Expected pool size 1, got %d", pool.Size())
	}

	// Test getting added transaction
	tx := pool.Get(tx1.Hash)
	if tx == nil {
		t.Error("Expected to get transaction from pool")
	}
	if tx.Hash != tx1.Hash {
		t.Error("Transaction hash mismatch")
	}
}

func TestTxPoolRemove(t *testing.T) {
	pool := NewTxPool()
	tx1 := createTestTransaction(1000, 0)
	tx1.Hash = tx1.ComputeHash()
	tx2 := createTestTransaction(2000, 1)
	tx2.Hash = tx2.ComputeHash()

	// Add transactions
	pool.Add(tx1)
	pool.Add(tx2)
	if pool.Size() != 2 {
		t.Errorf("Expected pool size 2, got %d", pool.Size())
	}

	// Remove transaction
	pool.Remove(tx1.Hash)
	if pool.Size() != 1 {
		t.Errorf("Expected pool size 1, got %d", pool.Size())
	}

	// Verify correct transaction was removed
	tx := pool.Get(tx1.Hash)
	if tx != nil {
		t.Error("Expected transaction to be removed")
	}
	tx = pool.Get(tx2.Hash)
	if tx == nil {
		t.Error("Expected transaction to remain in pool")
	}
}

func TestTxPoolGetAll(t *testing.T) {
	pool := NewTxPool()
	tx1 := createTestTransaction(1000, 0)
	tx1.Hash = tx1.ComputeHash()
	tx2 := createTestTransaction(2000, 1)
	tx2.Hash = tx2.ComputeHash()

	// Add transactions
	pool.Add(tx1)
	pool.Add(tx2)

	// Get all transactions
	txs := pool.GetAll()
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}

	// Verify transactions are copied
	txs[0].Value = 3000
	poolTx := pool.Get(tx1.Hash)
	if poolTx.Value != 1000 {
		t.Error("Pool transaction was modified when modifying GetAll result")
	}
}

func TestTxPoolClear(t *testing.T) {
	pool := NewTxPool()
	tx1 := createTestTransaction(1000, 0)
	tx1.Hash = tx1.ComputeHash()
	tx2 := createTestTransaction(2000, 1)
	tx2.Hash = tx2.ComputeHash()

	// Add transactions
	pool.Add(tx1)
	pool.Add(tx2)

	// Clear pool
	pool.Clear()
	if pool.Size() != 0 {
		t.Errorf("Expected empty pool, got size %d", pool.Size())
	}

	// Verify transactions are removed
	tx := pool.Get(tx1.Hash)
	if tx != nil {
		t.Error("Expected transaction to be removed after clear")
	}
}

func TestTxPoolConcurrency(t *testing.T) {
	pool := NewTxPool()
	done := make(chan bool)
	const numGoroutines = 10

	// Start multiple goroutines to test concurrent access
	for i := 0; i < numGoroutines; i++ {
		go func(val int64) {
			tx := createTestTransaction(val, uint64(val))
			tx.Hash = tx.ComputeHash()
			pool.Add(tx)
			_ = pool.Get(tx.Hash)
			pool.Remove(tx.Hash)
			_ = pool.GetAll()
			done <- true
		}(int64(i))
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Final state should be valid
	txs := pool.GetAll()
	if len(txs) > numGoroutines {
		t.Errorf("Pool size %d exceeds maximum possible size %d", len(txs), numGoroutines)
	}
}
