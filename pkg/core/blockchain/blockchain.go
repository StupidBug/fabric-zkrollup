package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/chaincode"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/txpool"
	"github.com/StupidBug/fabric-zkrollup/pkg/crypto"
	"github.com/StupidBug/fabric-zkrollup/pkg/mock"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/block"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/state"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
	"github.com/StupidBug/fabric-zkrollup/pkg/utils.go"
	"github.com/StupidBug/fabric-zkrollup/pkg/zk"
)

const (
	maxTransactions = 128
	debug           = true
)

// Blockchain represents the blockchain
type Blockchain struct {
	mu          sync.RWMutex // protects blocks and autoBlock
	blocks      []*block.Block
	state       *state.State
	txPool      *txpool.TxPool
	merkleTree  *crypto.MerkleTree // 当前区块的 Merkle 树
	createBlock chan struct{}
	packageCh   chan *block.Block
	autoBlock   bool
}

// NewBlockchain creates a new blockchain instance
func NewBlockchain() *Blockchain {
	callbackCh := make(chan struct{})
	packageCh := make(chan *block.Block)
	bc := &Blockchain{
		blocks:      make([]*block.Block, 0),
		state:       state.NewState(),
		txPool:      txpool.NewTxPool(),
		merkleTree:  crypto.NewMerkleTree(nil),
		autoBlock:   false,
		createBlock: callbackCh,
		packageCh:   packageCh,
	}
	go bc.PackageWoker()

	var accounts []types.Account
	if debug {
		accounts = mock.MockBalance()
	} else {
		accounts = chaincode.GetAllTokenBalances()
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
	// log.Printf("Found public key for sender %s: X=%s, Y=%s", tx.From,
	// 	senderPubKey.X.String(), senderPubKey.Y.String())

	// Verify signature
	txHash := tx.ComputeHash()
	// log.Printf("Transaction hash for verification: %x", txHash)
	// log.Printf("Signature values - R: %s, S: %s", tx.Signature.R.String(), tx.Signature.S.String())

	if !tx.VerifySignature(senderPubKey) {
		log.Printf("Signature verification failed for transaction %x", txHash)
		return fmt.Errorf("invalid signature")
	}

	// Get current balance and nonce - acquire read lock
	bc.mu.RLock()
	senderBalance := bc.state.GetBalance(tx.From)
	// expectedNonce := bc.state.GetNonce(tx.From)
	bc.mu.RUnlock()

	// Validate balance and nonce
	if senderBalance < tx.Value {
		log.Printf("Insufficient balance - Required: %d, Available: %d",
			tx.Value, senderBalance)
		return fmt.Errorf("insufficient balance")
	}

	// if tx.Nonce != expectedNonce {
	// 	log.Printf("Invalid nonce - Expected: %d, Got: %d",
	// 		expectedNonce, tx.Nonce)
	// 	return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, tx.Nonce)
	// }

	// Add to transaction pool - acquire write lock
	bc.mu.Lock()
	bc.txPool.Add(tx)
	bc.mu.Unlock()

	if bc.txPool.Size() >= maxTransactions {
		bc.createBlock <- struct{}{}
	}

	// log.Printf("Added transaction %s to pool", tx.String())
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
	transactions := bc.PopTransactions()
	if len(transactions) == 0 {
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
	bc.packageCh <- block

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
		ticker := time.NewTicker(5 * time.Second) // Create block every 1 second
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
				}
				log.Printf("Successfully created new block with %d transactions", bc.GetLatestBlock().Header.TransactionCount)
			} else {
				log.Printf("No transactions in pool during auto block creation")
			}
		}

		// Start a goroutine to monitor transaction pool
		go func() {
			for {
				log.Printf("start create block when reach max transaction")
				<-bc.createBlock
				checkAndCreateBlock()
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

func (bc *Blockchain) PackageWoker() {
	for block := range bc.packageCh {
		bc.applyTransactions(block)
		bc.blocks = append(bc.blocks, block)
	}
}

// applyTransactions applies a list of transactions and updates the state tree
func (bc *Blockchain) applyTransactions(block *block.Block) error {
	// 获取所有账户状态
	accounts := bc.GetAccountsForProof()
	transactions := bc.GetTransactionForProof(block)

	oldStateRoot := bc.GetStateRoot()
	// 准备证明输入
	input := zk.ProofInput{
		OldStateRoot: oldStateRoot,
		Accounts:     accounts,
		Transactions: transactions,
	}
	utils.PrintSlice(input.Accounts)
	utils.PrintSlice(input.Transactions)

	// 生成证明
	output, err := zk.GenerateProof(input)
	if err != nil {
		panic(fmt.Errorf("failed to generate ZK proof [transactions: %d]: %w", len(input.Transactions), err))
		return fmt.Errorf("failed to generate ZK proof [req: %#v]: %v", input, err)
	}

	if debug {
	} else {
		chaincode.JsonVerify(output)
	}

	// 更新交易状态
	for i := range block.Transactions {
		tx := &block.Transactions[i]
		// 更新交易状态
		tx.Status = transaction.StatusConfirmed
	}

	// 更新账户状态
	for _, account := range output.NewAccounts {
		// 更新发送方余额和nonce
		bc.state.SetBalance(account.Address, account.Balance)
		bc.state.SetNonce(account.Address, uint64(account.Nonce))
	}
	// Update state root
	block.Header.StateRoot = output.NewStateRoot

	fmt.Println("output new stateRoot: ", output.NewStateRoot)
	return nil
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

// GetAllBlocks returns all blocks with their transactions
func (bc *Blockchain) GetAllBlocks() []*block.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocks := make([]*block.Block, len(bc.blocks))
	copy(blocks, bc.blocks)
	return blocks
}
