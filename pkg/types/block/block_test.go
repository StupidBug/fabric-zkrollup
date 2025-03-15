package block

import (
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/crypto"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

func createTestTransaction(from, to string, value int64, nonce uint64) transaction.Transaction {
	transaction := transaction.Transaction{
		From:      from,
		To:        to,
		Value:     int(value),
		Nonce:     nonce,
		Status:    transaction.StatusPending,
		Timestamp: time.Now().Unix(),
	}
	transaction.Hash = transaction.ComputeHash()
	return transaction
}

func TestBlockHeader(t *testing.T) {
	header := Header{
		Version:          1,
		PrevHash:         [32]byte{1, 2, 3},
		MerkleRoot:       [32]byte{4, 5, 6},
		Timestamp:        time.Now(),
		Height:           10,
		TransactionCount: 2,
	}

	hash1 := header.ComputeHash()
	hash2 := header.ComputeHash()

	// 相同的区块头应该产生相同的哈希
	if hash1 != hash2 {
		t.Error("Same header should produce same hash")
	}

	// 修改区块头应该产生不同的哈希
	header.Height = 11
	hash3 := header.ComputeHash()
	if hash1 == hash3 {
		t.Error("Different headers should produce different hashes")
	}
}

func TestBlockHash(t *testing.T) {
	block := &Block{
		Header: Header{
			Version:          1,
			PrevHash:         [32]byte{1, 2, 3},
			MerkleRoot:       [32]byte{4, 5, 6},
			Timestamp:        time.Now(),
			Height:           10,
			TransactionCount: 2,
		},
		Transactions: []transaction.Transaction{
			createTestTransaction("sender1", "receiver1", 100, 0),
			createTestTransaction("sender2", "receiver2", 200, 0),
		},
	}

	hash1 := block.ComputeHash()
	hash2 := block.ComputeHash()

	// 相同的区块应该产生相同的哈希
	if hash1 != hash2 {
		t.Error("Same block should produce same hash")
	}

	// 修改区块应该产生不同的哈希
	block.Header.Height = 11
	hash3 := block.ComputeHash()
	if hash1 == hash3 {
		t.Error("Different blocks should produce different hashes")
	}
}

func TestBlockMerkleRoot(t *testing.T) {
	// 创建测试交易
	transactions := []transaction.Transaction{
		createTestTransaction("sender1", "receiver1", 100, 0),
		createTestTransaction("sender2", "receiver2", 200, 0),
	}

	// 创建区块
	block := &Block{
		Header: Header{
			Version:          1,
			PrevHash:         [32]byte{},
			Timestamp:        time.Now(),
			Height:           1,
			TransactionCount: uint32(len(transactions)),
		},
		Transactions: transactions,
	}

	// 计算 Merkle 根
	merkleTree := crypto.CreateMerkleTreeFromTransactions(transactions)
	block.Header.MerkleRoot = merkleTree.GetRoot()

	// 验证 Merkle 根不为空
	emptyHash := [32]byte{}
	if block.Header.MerkleRoot == emptyHash {
		t.Error("Merkle root should not be empty")
	}

	// 验证交易列表匹配 Merkle 根
	if !crypto.VerifyTransactionMerkleRoot(transactions, block.Header.MerkleRoot) {
		t.Error("Transaction list should match Merkle root")
	}

	// 验证不同交易产生不同的 Merkle 根
	transactions2 := []transaction.Transaction{
		createTestTransaction("sender1", "receiver1", 300, 0), // 修改了金额
		createTestTransaction("sender2", "receiver2", 200, 0),
	}
	merkleTree2 := crypto.CreateMerkleTreeFromTransactions(transactions2)
	if merkleTree2.GetRoot() == block.Header.MerkleRoot {
		t.Error("Different transactions should produce different Merkle roots")
	}
}

func TestEmptyBlock(t *testing.T) {
	block := &Block{
		Header: Header{
			Version:          1,
			PrevHash:         [32]byte{},
			MerkleRoot:       [32]byte{},
			Timestamp:        time.Now(),
			Height:           0,
			TransactionCount: 0,
		},
		Transactions: []transaction.Transaction{},
	}

	// 验证空区块的哈希计算
	hash := block.ComputeHash()
	if hash == [32]byte{} {
		t.Error("Empty block should still have a valid hash")
	}

	// 验证空区块的 Merkle 根
	merkleTree := crypto.CreateMerkleTreeFromTransactions(block.Transactions)
	if merkleTree.GetRoot() != [32]byte{} {
		t.Error("Empty block should have empty Merkle root")
	}
}
