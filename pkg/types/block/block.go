package block

import (
	"crypto/sha256"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

// Block represents a block in the blockchain
type Block struct {
	Header       Header
	Transactions []transaction.Transaction
}

// Header contains the header information of a block
type Header struct {
	Version          uint32    // Block version
	PrevHash         [32]byte  // Hash of the previous block
	MerkleRoot       [32]byte  // Merkle root of transactions
	StateRoot        string    // Merkle root of global state
	Timestamp        time.Time // Block timestamp
	Height           uint64    // Block height
	TransactionCount uint32    // Number of transactions in the block
}

// ComputeHash calculates the hash of a block header
func (h *Header) ComputeHash() [32]byte {
	var data []byte
	data = append(data, byte(h.Version))
	data = append(data, h.PrevHash[:]...)
	data = append(data, h.MerkleRoot[:]...)
	data = append(data, []byte(h.StateRoot)...)
	data = append(data, []byte(h.Timestamp.String())...)
	data = append(data, byte(h.Height))
	data = append(data, byte(h.TransactionCount))
	return sha256.Sum256(data)
}

// ComputeHash computes the hash of the block
func (b *Block) ComputeHash() [32]byte {
	// Compute hash based on header fields
	data := append(b.Header.PrevHash[:], b.Header.MerkleRoot[:]...)
	timeBytes := []byte(b.Header.Timestamp.String())
	data = append(data, timeBytes...)
	data = append(data, byte(b.Header.Height))
	return sha256.Sum256(data)
}
