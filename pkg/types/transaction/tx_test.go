package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"testing"
	"time"
)

func TestTransactionHash(t *testing.T) {
	// Create a test transaction
	tx := Transaction{
		From:      "0x1234567890123456789012345678901234567890",
		To:        "0x0987654321098765432109876543210987654321",
		Value:     1000,
		Nonce:     1,
		Status:    StatusPending,
		Timestamp: time.Now().Unix(),
	}

	// Compute hash
	hash := tx.ComputeHash()

	// Verify hash is not empty
	emptyHash := [32]byte{}
	if hash == emptyHash {
		t.Error("Expected non-empty hash")
	}

	// Verify same transaction produces same hash
	hash2 := tx.ComputeHash()
	if hash != hash2 {
		t.Error("Expected same hash for same transaction")
	}

	// Verify different transactions produce different hashes
	tx2 := tx
	tx2.Value = 2000
	hash3 := tx2.ComputeHash()
	if hash == hash3 {
		t.Error("Expected different hash for different transaction")
	}
}

func TestTransactionStatus(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPending, "pending"},
		{StatusConfirmed, "confirmed"},
		{StatusFailed, "failed"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("Status.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestTransactionString(t *testing.T) {
	tx := Transaction{
		From:      "sender",
		To:        "receiver",
		Value:     1000,
		Nonce:     1,
		Status:    StatusPending,
		Timestamp: time.Now().Unix(),
	}

	// Compute hash
	tx.Hash = tx.ComputeHash()

	// Get string representation
	str := tx.String()

	// Verify string contains important transaction details
	if len(str) == 0 {
		t.Error("Expected non-empty string representation")
	}

	expectedFields := []string{
		"Hash",
		"sender",
		"receiver",
		"1000",
		"1",
		"pending",
	}

	for _, field := range expectedFields {
		if !contains(str, field) {
			t.Errorf("Expected string representation to contain %s", field)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTransactionSignature(t *testing.T) {
	// Generate a test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a test transaction
	tx := Transaction{
		From:      "sender",
		To:        "receiver",
		Value:     1000,
		Nonce:     1,
		Status:    StatusPending,
		Timestamp: time.Now().Unix(),
	}

	// Sign the transaction
	if err := tx.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify signature exists
	if tx.Signature.R == nil || tx.Signature.S == nil {
		t.Error("Signature components should not be nil")
	}

	// Verify signature with correct public key
	if !tx.VerifySignature(&privateKey.PublicKey) {
		t.Error("Signature verification failed with correct public key")
	}

	// Generate another key pair for negative test
	wrongKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Verify signature with wrong public key
	if tx.VerifySignature(&wrongKey.PublicKey) {
		t.Error("Signature verification should fail with wrong public key")
	}

	// Test signature with modified transaction data
	tx.Value = 2000
	if tx.VerifySignature(&privateKey.PublicKey) {
		t.Error("Signature verification should fail with modified transaction data")
	}
}

func TestSignatureConsistency(t *testing.T) {
	// Generate a test key pair
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Create two identical transactions
	tx1 := Transaction{
		From:      "sender",
		To:        "receiver",
		Value:     1000,
		Nonce:     1,
		Status:    StatusPending,
		Timestamp: time.Now().Unix(),
	}

	tx2 := tx1 // Create a copy

	// Sign both transactions
	if err := tx1.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign first transaction: %v", err)
	}

	if err := tx2.SignTransaction(privateKey); err != nil {
		t.Fatalf("Failed to sign second transaction: %v", err)
	}

	// Verify both signatures
	if !tx1.VerifySignature(&privateKey.PublicKey) {
		t.Error("First transaction signature verification failed")
	}

	if !tx2.VerifySignature(&privateKey.PublicKey) {
		t.Error("Second transaction signature verification failed")
	}

	// Compare signatures
	if tx1.Signature.R.Cmp(tx2.Signature.R) == 0 && tx1.Signature.S.Cmp(tx2.Signature.S) == 0 {
		t.Error("Signatures should be different for same transaction (due to random k in ECDSA)")
	}
}

func TestInvalidSignatures(t *testing.T) {
	tx := Transaction{
		From:      "sender",
		To:        "receiver",
		Value:     1000,
		Nonce:     1,
		Status:    StatusPending,
		Timestamp: time.Now().Unix(),
	}

	// Test with nil signature components
	if tx.Signature.R != nil || tx.Signature.S != nil {
		t.Error("New transaction should have nil signature components")
	}

	// Test with zero signature components
	tx.Signature.R = big.NewInt(0)
	tx.Signature.S = big.NewInt(0)

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if tx.VerifySignature(&privateKey.PublicKey) {
		t.Error("Verification should fail with zero signature components")
	}

	// Test with negative signature components
	tx.Signature.R = big.NewInt(-1)
	tx.Signature.S = big.NewInt(-1)
	if tx.VerifySignature(&privateKey.PublicKey) {
		t.Error("Verification should fail with negative signature components")
	}
}
