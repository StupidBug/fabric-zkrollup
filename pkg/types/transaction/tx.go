package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Status represents the status of a transaction
type Status int

const (
	StatusPending Status = iota
	StatusConfirmed
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusConfirmed:
		return "confirmed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Signature represents an ECDSA signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Hash      [32]byte  // Hash of the transaction
	From      string    // Sender's address
	To        string    // Recipient's address
	Value     int       // Amount to transfer
	Nonce     uint64    // Transaction nonce
	Status    Status    // Transaction status
	Timestamp int64     // Transaction timestamp
	Signature Signature // Transaction signature
}

// ComputeHash calculates the hash of a transaction
func (tx *Transaction) ComputeHash() [32]byte {
	data := []byte(fmt.Sprintf("%s%s%d%d", tx.From, tx.To, tx.Value, tx.Nonce))
	return sha256.Sum256(data)
}

// SignTransaction signs the transaction with the given private key
func (tx *Transaction) SignTransaction(privateKey *ecdsa.PrivateKey) error {
	hash := tx.ComputeHash()

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	tx.Signature.R = r
	tx.Signature.S = s

	return nil
}

// VerifySignature verifies the transaction signature
func (tx *Transaction) VerifySignature(publicKey *ecdsa.PublicKey) bool {
	if tx.Signature.R == nil || tx.Signature.S == nil {
		return false
	}
	hash := tx.ComputeHash()
	return ecdsa.Verify(publicKey, hash[:], tx.Signature.R, tx.Signature.S)
}

// String returns a string representation of the transaction
func (tx *Transaction) String() string {
	return fmt.Sprintf("Transaction{Hash: %s, From: %s, To: %s, Value: %d, Nonce: %d, Status: %s}",
		hex.EncodeToString(tx.Hash[:]),
		tx.From,
		tx.To,
		tx.Value,
		tx.Nonce,
		tx.Status)
}
