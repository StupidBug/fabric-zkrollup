package types

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

// Address represents a 20-byte address
type Address [20]byte

// String returns the hex string representation of the address
func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

// Transaction status constants
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
)

// Transaction represents a basic transaction in the system
type Transaction struct {
	From      Address  // Sender address
	To        Address  // Receiver address
	Value     *big.Int // Transaction amount
	Nonce     uint64   // Transaction nonce
	Signature []byte   // Transaction signature
	Hash      [32]byte // Transaction hash
	Status    string   // Transaction status
}

// ComputeHash computes the hash of the transaction
func (tx *Transaction) ComputeHash() [32]byte {
	// TODO: Implement proper transaction serialization
	data := append(tx.From[:], tx.To[:]...)
	data = append(data, tx.Value.Bytes()...)
	data = append(data, big.NewInt(int64(tx.Nonce)).Bytes()...)
	return sha256.Sum256(data)
}

// Block represents a block in the blockchain
type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}

// BlockHeader contains the header information of a block
type BlockHeader struct {
	Version          uint32    // Block version
	PrevHash         [32]byte  // Hash of the previous block
	MerkleRoot       [32]byte  // Merkle root of transactions
	Timestamp        time.Time // Block timestamp
	Height           uint64    // Block height
	TransactionCount uint32    // Number of transactions in the block
}

// ComputeHash calculates the hash of a block header
func (h *BlockHeader) ComputeHash() [32]byte {
	var data []byte
	data = append(data, byte(h.Version))
	data = append(data, h.PrevHash[:]...)
	data = append(data, h.MerkleRoot[:]...)
	data = append(data, []byte(h.Timestamp.String())...)
	data = append(data, byte(h.Height))
	data = append(data, byte(h.TransactionCount))
	return sha256.Sum256(data)
}

// ComputeHash computes the hash of the block header
func (b *Block) ComputeHash() [32]byte {
	// TODO: Implement proper block header serialization
	data := append(b.Header.PrevHash[:], b.Header.MerkleRoot[:]...)
	timeBytes := big.NewInt(b.Header.Timestamp.Unix()).Bytes()
	data = append(data, timeBytes...)
	heightBytes := big.NewInt(int64(b.Header.Height)).Bytes()
	data = append(data, heightBytes...)
	return sha256.Sum256(data)
}

// CalculateMerkleRoot calculates the Merkle root of the block's transactions
func (b *Block) CalculateMerkleRoot() [32]byte {
	if len(b.Transactions) == 0 {
		return [32]byte{}
	}

	var hashes [][]byte
	for _, tx := range b.Transactions {
		hash := tx.Hash[:]
		hashes = append(hashes, hash)
	}

	// If there's only one transaction, use its hash as the root
	if len(hashes) == 1 {
		var root [32]byte
		copy(root[:], hashes[0])
		return root
	}

	// Build the Merkle tree
	for len(hashes) > 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var nextLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			var data []byte
			data = append(data, hashes[i]...)
			data = append(data, hashes[i+1]...)
			hash := sha256.Sum256(data)
			nextLevel = append(nextLevel, hash[:])
		}
		hashes = nextLevel
	}

	var root [32]byte
	copy(root[:], hashes[0])
	return root
}

// AccountState represents the state of an account
type AccountState struct {
	Balance *big.Int
	Nonce   uint64
}

// State represents the current state of the system
type State struct {
	Accounts map[Address]*AccountState
	Height   uint64
}
