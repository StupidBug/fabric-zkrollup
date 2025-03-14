package state

import (
	"math/big"
	"testing"
)

func TestStateBalance(t *testing.T) {
	s := NewState()
	addr := "0x1234567890123456789012345678901234567890"

	// Test initial balance
	balance := s.GetBalance(addr)
	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected initial balance 0, got %s", balance.String())
	}

	// Test setting balance
	newBalance := big.NewInt(1000)
	s.SetBalance(addr, newBalance)

	// Test getting balance
	balance = s.GetBalance(addr)
	if balance.Cmp(newBalance) != 0 {
		t.Errorf("Expected balance %s, got %s", newBalance.String(), balance.String())
	}

	// Test balance is copied, not referenced
	newBalance.Add(newBalance, big.NewInt(500))
	balance = s.GetBalance(addr)
	if balance.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected balance not to change when source big.Int is modified")
	}
}

func TestStateNonce(t *testing.T) {
	s := NewState()
	addr := "0x1234567890123456789012345678901234567890"

	// Test initial nonce
	nonce := s.GetNonce(addr)
	if nonce != 0 {
		t.Errorf("Expected initial nonce 0, got %d", nonce)
	}

	// Test setting nonce
	s.SetNonce(addr, 1)

	// Test getting nonce
	nonce = s.GetNonce(addr)
	if nonce != 1 {
		t.Errorf("Expected nonce 1, got %d", nonce)
	}
}

func TestStateClone(t *testing.T) {
	s := NewState()
	addr1 := "0x1234567890123456789012345678901234567890"
	addr2 := "0x0987654321098765432109876543210987654321"

	// Set some initial state
	s.SetBalance(addr1, big.NewInt(1000))
	s.SetBalance(addr2, big.NewInt(2000))
	s.SetNonce(addr1, 1)
	s.SetNonce(addr2, 2)

	// Clone the state
	clone := s.Clone()

	// Verify balances are copied correctly
	for _, addr := range []string{addr1, addr2} {
		originalBalance := s.GetBalance(addr)
		clonedBalance := clone.GetBalance(addr)
		if originalBalance.Cmp(clonedBalance) != 0 {
			t.Errorf("Balance mismatch for %s: original %s, clone %s",
				addr, originalBalance.String(), clonedBalance.String())
		}
	}

	// Verify nonces are copied correctly
	for _, addr := range []string{addr1, addr2} {
		originalNonce := s.GetNonce(addr)
		clonedNonce := clone.GetNonce(addr)
		if originalNonce != clonedNonce {
			t.Errorf("Nonce mismatch for %s: original %d, clone %d",
				addr, originalNonce, clonedNonce)
		}
	}

	// Modify clone and verify original is unchanged
	clone.SetBalance(addr1, big.NewInt(3000))
	clone.SetNonce(addr1, 3)

	if s.GetBalance(addr1).Cmp(big.NewInt(1000)) != 0 {
		t.Error("Original balance changed after modifying clone")
	}
	if s.GetNonce(addr1) != 1 {
		t.Error("Original nonce changed after modifying clone")
	}
}

func TestStateConcurrency(t *testing.T) {
	s := NewState()
	addr := "0x1234567890123456789012345678901234567890"
	done := make(chan bool)

	// Start multiple goroutines to test concurrent access
	for i := 0; i < 10; i++ {
		go func(val int64) {
			s.SetBalance(addr, big.NewInt(val))
			s.SetNonce(addr, uint64(val))
			_ = s.GetBalance(addr)
			_ = s.GetNonce(addr)
			done <- true
		}(int64(i))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Final state should be valid (we don't test for specific values as they depend on timing)
	balance := s.GetBalance(addr)
	if balance == nil {
		t.Error("Balance should not be nil after concurrent operations")
	}
}
