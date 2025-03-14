package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	// "zkrollup/internal/chaincode"
	"zkrollup/internal/core/txpool"
	"zkrollup/internal/crypto"
	"zkrollup/internal/types"
	"zkrollup/internal/types/block"
	"zkrollup/internal/types/state"
	"zkrollup/internal/types/transaction"
	"zkrollup/internal/zk"
)

// Blockchain represents the blockchain
type Blockchain struct {
	mu         sync.RWMutex // protects blocks and autoBlock
	blocks     []*block.Block
	state      *state.State
	txPool     *txpool.TxPool
	merkleTree *crypto.MerkleTree // 当前区块的 Merkle 树
	autoBlock  bool
}

// NewBlockchain creates a new blockchain instance
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		blocks:     make([]*block.Block, 0),
		state:      state.NewState(),
		txPool:     txpool.NewTxPool(),
		merkleTree: crypto.NewMerkleTree(nil),
		autoBlock:  false,
	}

	accounts := []zk.Account{
		{
			Address: "0000000000000000000000000000000000000001",
			Balance: 1000000,
			Nonce:   0,
		},
		{
			Address: "0000000000000000000000000000000000000002",
			Balance: 500000,
			Nonce:   0,
		},
		{
			Address: "0000000000000000000000000000000000000003",
			Balance: 300000,
			Nonce:   0,
		},
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Address < accounts[j].Address
	})

	// Set initial balances in state
	for _, account := range accounts {
		bc.state.SetBalance(account.Address, account.Balance)
	}

	// Compute initial state root using zk package
	stateRoot := zk.ComputeAccountMerkleRoot(accounts)

	// Create genesis block
	genesisBlock := &block.Block{
		Header: block.Header{
			Version:          1,
			PrevHash:         [32]byte{},
			MerkleRoot:       [32]byte{},
			StateRoot:        stateRoot,
			Timestamp:        time.Now(),
			Height:           0,
			TransactionCount: 0,
		},
		Transactions: []transaction.Transaction{},
	}

	// Add genesis block to blockchain
	bc.blocks = append(bc.blocks, genesisBlock)
	log.Printf("Genesis block created with state root: %s", stateRoot)

	return bc
}

// AddTransaction adds a transaction to the transaction pool
func (bc *Blockchain) AddTransaction(tx transaction.Transaction) error {
	// Verify transaction signature first
	if tx.Signature.R == nil || tx.Signature.S == nil {
		log.Printf("Transaction missing signature - R: %v, S: %v", tx.Signature.R, tx.Signature.S)
		return fmt.Errorf("missing signature")
	}

	// Verify signature values are in valid range
	if tx.Signature.R.Sign() <= 0 || tx.Signature.S.Sign() <= 0 {
		log.Printf("Invalid signature values - R: %s, S: %s", tx.Signature.R.String(), tx.Signature.S.String())
		return fmt.Errorf("invalid signature values")
	}

	// Get sender's public key - acquire read lock
	bc.mu.RLock()
	senderPubKey := bc.state.GetPublicKey(tx.From)
	bc.mu.RUnlock()

	if senderPubKey == nil {
		log.Printf("Public key not found for sender %s", tx.From)
		return fmt.Errorf("public key not found for sender %s", tx.From)
	}
	log.Printf("Found public key for sender %s: X=%s, Y=%s", tx.From,
		senderPubKey.X.String(), senderPubKey.Y.String())

	// Verify signature
	txHash := tx.ComputeHash()
	log.Printf("Transaction hash for verification: %x", txHash)
	log.Printf("Signature values - R: %s, S: %s", tx.Signature.R.String(), tx.Signature.S.String())

	if !tx.VerifySignature(senderPubKey) {
		log.Printf("Signature verification failed for transaction %x", txHash)
		return fmt.Errorf("invalid signature")
	}

	// Get current balance and nonce - acquire read lock
	bc.mu.RLock()
	senderBalance := bc.state.GetBalance(tx.From)
	expectedNonce := bc.state.GetNonce(tx.From)
	bc.mu.RUnlock()

	// Validate balance and nonce
	if senderBalance < tx.Value {
		log.Printf("Insufficient balance - Required: %d, Available: %d",
			tx.Value, senderBalance)
		return fmt.Errorf("insufficient balance")
	}

	if tx.Nonce != expectedNonce {
		log.Printf("Invalid nonce - Expected: %d, Got: %d",
			expectedNonce, tx.Nonce)
		return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, tx.Nonce)
	}

	// Add to transaction pool - acquire write lock
	bc.mu.Lock()
	bc.txPool.Add(tx)
	bc.mu.Unlock()

	log.Printf("Added transaction %s to pool", tx.String())
	return nil
}

// GetTransactionByHash returns a transaction by its hash
func (bc *Blockchain) GetTransactionByHash(hash [32]byte) *transaction.Transaction {
	// First check the transaction pool
	if tx := bc.txPool.Get(hash); tx != nil {
		return tx
	}

	// Then check the blocks
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == hash {
				txCopy := tx
				return &txCopy
			}
		}
	}

	return nil
}

// GetBalance returns the balance of an address
func (bc *Blockchain) GetBalance(address string) int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.state.GetBalance(address)
}

// SetBalance is now private and only used during genesis block creation
func (bc *Blockchain) setBalance(address string, balance int) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Only allow setting balance when creating genesis block
	if len(bc.blocks) > 0 {
		return fmt.Errorf("cannot set balance after genesis block: balance can only be modified through transactions")
	}

	// Convert string address to types.Address
	var addr types.Address
	decoded, err := hex.DecodeString(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %v", err)
	}
	copy(addr[:], decoded)

	// Set balance in state
	bc.state.SetBalance(address, balance)
	log.Printf("Set balance for address %s to %d", address, balance)
	return nil
}

// GetNonce returns the nonce of an address
func (bc *Blockchain) GetNonce(address string) uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.state.GetNonce(address)
}

// GetStateRoot returns the current state root
func (bc *Blockchain) GetStateRoot() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.blocks) == 0 {
		return ""
	}
	return bc.blocks[len(bc.blocks)-1].Header.StateRoot
}

// GetLatestBlock returns the latest block in the chain
func (bc *Blockchain) GetLatestBlock() *block.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.blocks) == 0 {
		return nil
	}
	return bc.blocks[len(bc.blocks)-1]
}

// CreateBlock creates a new block with transactions from the pool
func (bc *Blockchain) CreateBlock() error {
	// Get pending transactions without any lock
	transactions := bc.txPool.GetAll()
	if len(transactions) == 0 {
		log.Printf("No transactions in pool to create block")
		return fmt.Errorf("no transactions to create block")
	}

	log.Printf("Creating new block with %d transactions", len(transactions))

	// Get previous block hash with read lock
	bc.mu.RLock()
	prevHash := [32]byte{}
	blockHeight := uint64(len(bc.blocks))
	if len(bc.blocks) > 0 {
		prevHash = bc.blocks[len(bc.blocks)-1].ComputeHash()
	}
	bc.mu.RUnlock()

	// Create new block (no lock needed)
	block := &block.Block{
		Header: block.Header{
			Version:          1,
			PrevHash:         prevHash,
			Timestamp:        time.Now(),
			Height:           blockHeight,
			TransactionCount: uint32(len(transactions)),
		},
		Transactions: transactions,
	}

	// Calculate Merkle root (no lock needed)
	merkleTree := crypto.CreateMerkleTreeFromTransactions(transactions)
	block.Header.MerkleRoot = merkleTree.GetRoot()
	log.Printf("Calculated Merkle root: %x", block.Header.MerkleRoot)

	log.Printf("applyTransactions: %d", len(transactions))
	// Apply transactions and update state
	if newStateRoot, err := bc.applyTransactions(block); err != nil {
		return fmt.Errorf("failed to apply transactions: %v", err)
	} else {
		// Update state root
		block.Header.StateRoot = newStateRoot
	}

	// Add block to chain
	bc.blocks = append(bc.blocks, block)

	// Clear processed transactions from pool
	bc.txPool.Clear()

	return nil
}

// VerifyBlock verifies a block's integrity
func (bc *Blockchain) VerifyBlock(block *block.Block) error {
	// Verify block header
	if block.Header.Height == 0 {
		// Genesis block has no previous hash
		if block.Header.PrevHash != [32]byte{} {
			return fmt.Errorf("genesis block must have empty previous hash")
		}
	} else {
		// Non-genesis blocks must have valid previous hash
		bc.mu.RLock()
		prevBlock := bc.blocks[block.Header.Height-1]
		bc.mu.RUnlock()
		if block.Header.PrevHash != prevBlock.Header.ComputeHash() {
			return fmt.Errorf("invalid previous block hash")
		}
	}

	// Verify transaction count
	if uint32(len(block.Transactions)) != block.Header.TransactionCount {
		return fmt.Errorf("transaction count mismatch")
	}

	// Verify Merkle root
	if !crypto.VerifyTransactionMerkleRoot(block.Transactions, block.Header.MerkleRoot) {
		return fmt.Errorf("merkle root mismatch")
	}

	return nil
}

// StartAutoBlock starts automatic block creation
func (bc *Blockchain) StartAutoBlock() {
	bc.mu.Lock()
	if bc.autoBlock {
		bc.mu.Unlock()
		return
	}
	bc.autoBlock = true
	bc.mu.Unlock()

	go func() {
		log.Println("Starting auto block creation goroutine")
		ticker := time.NewTicker(1 * time.Second) // Create block every 1 second
		defer ticker.Stop()

		checkAndCreateBlock := func() {
			// Check if auto block is still enabled
			bc.mu.RLock()
			autoBlockEnabled := bc.autoBlock
			bc.mu.RUnlock()

			if !autoBlockEnabled {
				log.Println("Auto block creation stopped")
				return
			}

			// Check pool size without holding the lock
			poolSize := bc.txPool.Size()
			if poolSize > 0 {
				log.Printf("Auto block creation triggered with %d transactions in pool", poolSize)
				if err := bc.CreateBlock(); err != nil {
					log.Printf("Error creating block: %v", err)
				} else {
					log.Printf("Successfully created new block with %d transactions", poolSize)
				}
			} else {
				log.Printf("No transactions in pool during auto block creation")
			}
		}

		// Start a goroutine to monitor transaction pool
		go func() {
			for {
				time.Sleep(100 * time.Millisecond) // Check every 100ms

				bc.mu.RLock()
				autoBlockEnabled := bc.autoBlock
				bc.mu.RUnlock()

				if !autoBlockEnabled {
					return
				}

				if bc.txPool.Size() > 1 {
					checkAndCreateBlock()
				}
			}
		}()

		// Main loop for timed block creation
		for {
			select {
			case <-ticker.C:
				checkAndCreateBlock()
			}
		}
	}()
	log.Println("Auto block creation started")
}

// StopAutoBlock stops automatic block creation
func (bc *Blockchain) StopAutoBlock() {
	bc.mu.Lock()
	bc.autoBlock = false
	bc.mu.Unlock()
}

// validateTransaction validates a transaction
func (bc *Blockchain) validateTransaction(transaction *transaction.Transaction) error {
	// Note: This function assumes the caller holds appropriate locks
	senderBalance := bc.state.GetBalance(transaction.From)
	log.Printf("Validating transaction - Sender: %s, Balance: %d, Transfer Amount: %d",
		transaction.From, senderBalance, transaction.Value)

	if senderBalance < transaction.Value {
		log.Printf("Insufficient balance - Required: %d, Available: %d",
			transaction.Value, senderBalance)
		return fmt.Errorf("insufficient balance")
	}

	// Check nonce
	expectedNonce := bc.state.GetNonce(transaction.From)
	if transaction.Nonce != expectedNonce {
		log.Printf("Invalid nonce - Expected: %d, Got: %d",
			expectedNonce, transaction.Nonce)
		return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, transaction.Nonce)
	}

	return nil
}

// applyTransactions applies a list of transactions and updates the state tree
func (bc *Blockchain) applyTransactions(block *block.Block) (string, error) {
	// 准备输入数据
	var accounts []zk.Account

	// 获取所有账户状态
	allAccounts := bc.state.GetAllAccounts()
	for addr, acc := range allAccounts {
		accounts = append(accounts, zk.Account{
			Address: addr,
			Balance: acc.Balance,
			Nonce:   int(acc.Nonce),
		})
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Address < accounts[j].Address
	})

	// 准备交易数据
	var transactions []zk.Transaction
	for _, tx := range block.Transactions {
		transactions = append(transactions, zk.Transaction{
			From:   tx.From,
			To:     tx.To,
			Amount: tx.Value,
			Nonce:  int(tx.Nonce),
		})
	}

	oldStateRoot := bc.GetStateRoot()
	// 准备证明输入
	input := zk.ProofInput{
		OldStateRoot: oldStateRoot,
		Accounts:     accounts,
		Transactions: transactions,
	}

	fmt.Printf("input: %#v\n", input)
	// 生成证明
	output, err := zk.GenerateProof(input)
	if err != nil {
		return "", fmt.Errorf("failed to generate ZK proof [req: %#v]: %v", input, err)
	}

	// chaincode.JsonVerify(output)

	// 更新账户状态
	for i := range block.Transactions {
		tx := &block.Transactions[i]
		// 更新发送方余额和nonce
		fromBalance := bc.state.GetBalance(tx.From)
		bc.state.SetBalance(tx.From, fromBalance-tx.Value)
		bc.state.SetNonce(tx.From, tx.Nonce+1)

		// 更新接收方余额
		toBalance := bc.state.GetBalance(tx.To)
		bc.state.SetBalance(tx.To, toBalance+tx.Value)

		// 更新交易状态
		tx.Status = transaction.StatusConfirmed
	}
	fmt.Println("output.NewStateRoot: ", output.NewStateRoot)

	return output.NewStateRoot, nil
}

// ResetState resets the blockchain state
func (bc *Blockchain) ResetState() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Reset blocks
	bc.blocks = make([]*block.Block, 0)

	// Reset state
	bc.state = state.NewState()

	// Reset transaction pool
	bc.txPool = txpool.NewTxPool()

	// Reset Merkle tree
	bc.merkleTree = crypto.NewMerkleTree(nil)

	log.Println("Blockchain state has been reset")
}

// GetPublicKey returns the public key for an address
func (bc *Blockchain) GetPublicKey(address string) *ecdsa.PublicKey {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.state.GetPublicKey(address)
}

// SetPublicKey sets the public key for an address
func (bc *Blockchain) SetPublicKey(address string, pubKey *ecdsa.PublicKey) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.state.SetPublicKey(address, pubKey)
}

// GetTransactionPool returns all transactions in the pool
func (bc *Blockchain) GetTransactionPool() []transaction.Transaction {
	return bc.txPool.GetAll()
}

// GetHeight returns the current blockchain height
func (bc *Blockchain) GetHeight() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return uint64(len(bc.blocks))
}

// GetBlock returns a block at the specified height
func (bc *Blockchain) GetBlock(height uint64) (*block.Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.blocks)) {
		return nil, fmt.Errorf("block not found at height %d", height)
	}

	return bc.blocks[height], nil
}

// GetBlocks returns all blocks in the blockchain
func (bc *Blockchain) GetBlocks() []*block.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocks := make([]*block.Block, len(bc.blocks))
	copy(blocks, bc.blocks)
	return blocks
}
