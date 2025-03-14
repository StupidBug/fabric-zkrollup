package state

import (
	"crypto/ecdsa"
	"sync"
)

// AccountState represents the state of an account
type AccountState struct {
	Balance int
	Nonce   uint64
}

// State represents the current state of the blockchain
type State struct {
	mu       sync.RWMutex
	balances map[string]int              // Address -> Balance mapping
	nonces   map[string]uint64           // Address -> Nonce mapping
	pubKeys  map[string]*ecdsa.PublicKey // Address -> Public Key mapping
}

// NewState creates a new state instance
func NewState() *State {
	return &State{
		balances: make(map[string]int),
		nonces:   make(map[string]uint64),
		pubKeys:  make(map[string]*ecdsa.PublicKey),
	}
}

// GetBalance returns the balance of an address
func (s *State) GetBalance(address string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if balance, exists := s.balances[address]; exists {
		return balance
	}
	return 0
}

// SetBalance sets the balance for an address
func (s *State) SetBalance(address string, balance int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.balances[address] = balance
}

// GetNonce returns the nonce of an address
func (s *State) GetNonce(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.nonces[address]
}

// SetNonce sets the nonce for an address
func (s *State) SetNonce(address string, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nonces[address] = nonce
}

// GetPublicKey returns the public key for an address
func (s *State) GetPublicKey(address string) *ecdsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.pubKeys[address]
}

// SetPublicKey sets the public key for an address
func (s *State) SetPublicKey(address string, pubKey *ecdsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pubKeys[address] = pubKey
}

// Clone creates a deep copy of the state
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewState()
	for addr, balance := range s.balances {
		newState.balances[addr] = balance
	}
	for addr, nonce := range s.nonces {
		newState.nonces[addr] = nonce
	}
	for addr, pubKey := range s.pubKeys {
		newState.pubKeys[addr] = pubKey
	}
	return newState
}

// GetAllAccounts returns all accounts and their states
func (s *State) GetAllAccounts() map[string]*AccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make(map[string]*AccountState)
	for addr, balance := range s.balances {
		accounts[addr] = &AccountState{
			Balance: balance,
			Nonce:   s.nonces[addr],
		}
	}
	return accounts
}
