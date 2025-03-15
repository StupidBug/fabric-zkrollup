package crypto

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestMerkleTreeEmpty(t *testing.T) {
	// Test empty tree
	tree := NewMerkleTree(nil)
	if tree == nil {
		t.Fatal("Expected non-nil tree for empty data")
	}

	root := tree.GetRoot()
	emptyHash := [32]byte{}
	if root != emptyHash {
		t.Error("Expected empty hash for empty tree")
	}
}

func TestMerkleTreeSingleNode(t *testing.T) {
	// Test tree with single node
	data := [][]byte{[]byte("test data")}
	tree := NewMerkleTree(data)

	expectedHash := sha256.Sum256(data[0])
	if tree.GetRoot() != expectedHash {
		t.Error("Root hash doesn't match expected hash for single node")
	}
}

func TestMerkleTreeMultipleNodes(t *testing.T) {
	// Test tree with multiple nodes
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}
	tree := NewMerkleTree(data)

	// Verify root is not empty
	if tree.GetRoot() == [32]byte{} {
		t.Error("Expected non-empty root hash for multiple nodes")
	}

	// Verify tree structure
	if tree.Root.Left == nil || tree.Root.Right == nil {
		t.Error("Expected non-nil children for root node")
	}
}

func TestMerkleTreeOddNodes(t *testing.T) {
	// Test tree with odd number of nodes
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
	}
	tree := NewMerkleTree(data)

	// Last node should be duplicated
	lastHash := sha256.Sum256(data[2])
	if !bytes.Equal(tree.Root.Right.Left.Hash[:], lastHash[:]) ||
		!bytes.Equal(tree.Root.Right.Right.Hash[:], lastHash[:]) {
		t.Error("Last node was not properly duplicated for odd number of nodes")
	}
}

func TestMerkleTreeConsistency(t *testing.T) {
	// Test that same data produces same tree
	data1 := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
	}
	data2 := make([][]byte, len(data1))
	copy(data2, data1)

	tree1 := NewMerkleTree(data1)
	tree2 := NewMerkleTree(data2)

	if tree1.GetRoot() != tree2.GetRoot() {
		t.Error("Same data produced different Merkle roots")
	}
}

func TestMerkleNodeCreation(t *testing.T) {
	// Test leaf node creation
	data := []byte("leaf data")
	leafNode := NewMerkleNode(nil, nil, data)
	expectedHash := sha256.Sum256(data)
	if leafNode.Hash != expectedHash {
		t.Error("Leaf node hash doesn't match expected hash")
	}

	// Test internal node creation
	left := NewMerkleNode(nil, nil, []byte("left"))
	right := NewMerkleNode(nil, nil, []byte("right"))
	parent := NewMerkleNode(left, right, nil)

	// Verify parent hash is combination of children
	combined := append(left.Hash[:], right.Hash[:]...)
	expectedParentHash := sha256.Sum256(combined)
	if parent.Hash != expectedParentHash {
		t.Error("Parent node hash doesn't match expected hash")
	}
}

func TestMerkleTreeModification(t *testing.T) {
	// Create initial tree
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
	}
	tree1 := NewMerkleTree(data)
	originalRoot := tree1.GetRoot()

	// Modify data and create new tree
	data[1] = []byte("modified")
	tree2 := NewMerkleTree(data)
	modifiedRoot := tree2.GetRoot()

	// Verify roots are different
	if originalRoot == modifiedRoot {
		t.Error("Modified data produced same Merkle root")
	}
}

func TestStateTree(t *testing.T) {
	tree := NewStateTree()

	// Test empty tree
	if root := tree.GetRoot(); root != [32]byte{} {
		t.Error("Empty tree should have empty root")
	}

	// Test single update
	addr1 := []byte("address1")
	value1 := []byte("value1")
	tree.Update(addr1, value1)

	root1 := tree.GetRoot()
	if root1 == [32]byte{} {
		t.Error("Root should not be empty after update")
	}

	// Test value retrieval
	if val, exists := tree.Get(addr1); !exists {
		t.Error("Value should exist in tree")
	} else if string(val) != "value1" {
		t.Errorf("Expected value1, got %s", string(val))
	}

	// Test multiple updates
	addr2 := []byte("address2")
	value2 := []byte("value2")
	tree.Update(addr2, value2)

	root2 := tree.GetRoot()
	if root2 == root1 {
		t.Error("Root should change after new update")
	}

	// Test updating existing key
	tree.Update(addr1, []byte("newvalue1"))
	if val, _ := tree.Get(addr1); string(val) != "newvalue1" {
		t.Error("Value should be updated")
	}

	// Test non-existent key
	if _, exists := tree.Get([]byte("nonexistent")); exists {
		t.Error("Should not find non-existent key")
	}
}

func TestStateTreeConcurrency(t *testing.T) {
	tree := NewStateTree()
	done := make(chan bool)
	const numGoroutines = 10

	// Start multiple goroutines to test concurrent access
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			key := []byte(fmt.Sprintf("key%d", i))
			value := []byte(fmt.Sprintf("value%d", i))
			tree.Update(key, value)
			_, _ = tree.Get(key)
			_ = tree.GetRoot()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
