package txpool

import (
	"sync"
	"zkrollup/internal/types/transaction"
)

// TxPool represents the transaction pool
type TxPool struct {
	mu           sync.RWMutex
	transactions []transaction.Transaction
}

// NewTxPool creates a new transaction pool
func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make([]transaction.Transaction, 0),
	}
}

// Add adds a transaction to the pool
func (p *TxPool) Add(tx transaction.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.transactions = append(p.transactions, tx)
}

// Remove removes a transaction from the pool
func (p *TxPool) Remove(hash [32]byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, t := range p.transactions {
		if t.Hash == hash {
			p.transactions = append(p.transactions[:i], p.transactions[i+1:]...)
			return
		}
	}
}

// Get returns a transaction by its hash
func (p *TxPool) Get(hash [32]byte) *transaction.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, t := range p.transactions {
		if t.Hash == hash {
			txCopy := t
			return &txCopy
		}
	}
	return nil
}

// GetAll returns all transactions in the pool
func (p *TxPool) GetAll() []transaction.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	txs := make([]transaction.Transaction, len(p.transactions))
	copy(txs, p.transactions)
	return txs
}

// Clear removes all transactions from the pool
func (p *TxPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.transactions = make([]transaction.Transaction, 0)
}

// Size returns the number of transactions in the pool
func (p *TxPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.transactions)
}
