package crypto

import (
	"bytes"
	"crypto/sha256"
	"sync"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  [32]byte
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root  *MerkleNode
	Leafs []*MerkleNode // 保存所有叶子节点，用于增量更新
}

// StateNode represents a node in the state Merkle tree
type StateNode struct {
	Key   []byte     // Address or intermediate node key
	Value []byte     // Account state data or intermediate hash
	Hash  [32]byte   // Node hash
	Left  *StateNode // Left child
	Right *StateNode // Right child
}

// StateTree represents the global state Merkle tree
type StateTree struct {
	mu   sync.RWMutex
	root *StateNode
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		node.Hash = sha256.Sum256(data)
	} else {
		prevHashes := append(left.Hash[:], right.Hash[:]...)
		node.Hash = sha256.Sum256(prevHashes)
	}

	node.Left = left
	node.Right = right

	return &node
}

// NewMerkleTree creates a new Merkle tree from a list of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{
			Root: &MerkleNode{
				Hash: [32]byte{},
			},
			Leafs: make([]*MerkleNode, 0),
		}
	}

	var leafs []*MerkleNode

	// Create leaf nodes
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		leafs = append(leafs, node)
	}

	nodes := make([]*MerkleNode, len(leafs))
	copy(nodes, leafs)

	// If we have only one node, return it as the root
	if len(nodes) == 1 {
		return &MerkleTree{
			Root:  nodes[0],
			Leafs: leafs,
		}
	}

	// Build tree by pairing nodes
	for len(nodes) > 1 {
		levelSize := (len(nodes) + 1) / 2 * 2 // 确保偶数个节点
		level := make([]*MerkleNode, 0, levelSize/2)

		for i := 0; i < len(nodes); i += 2 {
			var right *MerkleNode
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = nodes[i] // 如果是奇数个节点，复制最后一个
			}
			node := NewMerkleNode(nodes[i], right, nil)
			level = append(level, node)
		}

		nodes = level
	}

	return &MerkleTree{
		Root:  nodes[0],
		Leafs: leafs,
	}
}

// AddNode adds a new leaf node to the tree
func (m *MerkleTree) AddNode(data []byte) {
	newLeaf := NewMerkleNode(nil, nil, data)
	m.Leafs = append(m.Leafs, newLeaf)

	// 重建树
	nodes := make([]*MerkleNode, len(m.Leafs))
	copy(nodes, m.Leafs)

	// 构建新的树结构
	for len(nodes) > 1 {
		levelSize := (len(nodes) + 1) / 2 * 2
		level := make([]*MerkleNode, 0, levelSize/2)

		for i := 0; i < len(nodes); i += 2 {
			var right *MerkleNode
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = nodes[i]
			}
			node := NewMerkleNode(nodes[i], right, nil)
			level = append(level, node)
		}

		nodes = level
	}

	m.Root = nodes[0]
}

// GetRoot returns the Merkle root hash
func (m *MerkleTree) GetRoot() [32]byte {
	return m.Root.Hash
}

// CreateMerkleTreeFromTransactions creates a Merkle tree from a list of transactions
func CreateMerkleTreeFromTransactions(transactions []transaction.Transaction) *MerkleTree {
	if len(transactions) == 0 {
		return NewMerkleTree(nil)
	}

	var txHashes [][]byte
	for _, tx := range transactions {
		hash := tx.ComputeHash()
		txHashes = append(txHashes, hash[:])
	}

	return NewMerkleTree(txHashes)
}

// AddTransaction adds a transaction to an existing Merkle tree
func (m *MerkleTree) AddTransaction(transaction *transaction.Transaction) {
	hash := transaction.ComputeHash()
	m.AddNode(hash[:])
}

// VerifyTransactionMerkleRoot verifies if a list of transactions matches a given Merkle root
func VerifyTransactionMerkleRoot(transactions []transaction.Transaction, root [32]byte) bool {
	tree := CreateMerkleTreeFromTransactions(transactions)
	return tree.GetRoot() == root
}

// NewStateTree creates a new state tree
func NewStateTree() *StateTree {
	return &StateTree{
		root: nil,
	}
}

// Update updates or inserts a key-value pair in the state tree
func (st *StateTree) Update(key, value []byte) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.root == nil {
		st.root = &StateNode{
			Key:   key,
			Value: value,
			Hash:  sha256.Sum256(append(key, value...)),
		}
		return
	}

	st.root = st.updateNode(st.root, key, value)
}

// updateNode recursively updates the tree
func (st *StateTree) updateNode(node *StateNode, key, value []byte) *StateNode {
	if node == nil {
		return &StateNode{
			Key:   key,
			Value: value,
			Hash:  sha256.Sum256(append(key, value...)),
		}
	}

	// Compare keys to determine which branch to take
	if bytes.Compare(key, node.Key) < 0 {
		node.Left = st.updateNode(node.Left, key, value)
	} else if bytes.Compare(key, node.Key) > 0 {
		node.Right = st.updateNode(node.Right, key, value)
	} else {
		// Update existing node
		node.Value = value
	}

	// Recalculate hash
	var data []byte
	if node.Left != nil {
		data = append(data, node.Left.Hash[:]...)
	}
	if node.Right != nil {
		data = append(data, node.Right.Hash[:]...)
	}
	data = append(data, node.Key...)
	data = append(data, node.Value...)
	node.Hash = sha256.Sum256(data)

	return node
}

// Get retrieves a value from the state tree
func (st *StateTree) Get(key []byte) ([]byte, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	node := st.root
	for node != nil {
		cmp := bytes.Compare(key, node.Key)
		if cmp == 0 {
			return node.Value, true
		} else if cmp < 0 {
			node = node.Left
		} else {
			node = node.Right
		}
	}
	return nil, false
}

// GetRoot returns the state root hash
func (st *StateTree) GetRoot() [32]byte {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.root == nil {
		return [32]byte{}
	}
	return st.root.Hash
}

// VerifyProof verifies a Merkle proof for a given key-value pair
func (st *StateTree) VerifyProof(key, value []byte, proof [][32]byte) bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.root == nil {
		return false
	}

	hash := sha256.Sum256(append(key, value...))
	for _, proofElement := range proof {
		data := append(hash[:], proofElement[:]...)
		hash = sha256.Sum256(data)
	}

	return hash == st.root.Hash
}
