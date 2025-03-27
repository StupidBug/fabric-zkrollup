package zk

import (
	"encoding/json"
	"testing"

	"github.com/StupidBug/fabric-zkrollup/pkg/mock"
)

var (
	accounts     = mock.MockBalance()
	transactions = mock.MockTransaction()
)

func BenchmarkGenerateProof(b *testing.B) {
	old_merkleRoot := ComputeAccountMerkleRoot(accounts)
	input := ProofInput{
		OldStateRoot: old_merkleRoot,
		Accounts:     accounts,
		Transactions: transactions,
	}

	b.StartTimer()
	// 生成证明
	_, err := GenerateProof(input)
	if err != nil {
		b.Fatalf("Failed to generate proof: %s\n", err.Error())
	}
	b.StopTimer()
}

func BenchmarkValidateProof(b *testing.B) {
	old_merkleRoot := ComputeAccountMerkleRoot(accounts)
	input := ProofInput{
		OldStateRoot: old_merkleRoot,
		Accounts:     accounts,
		Transactions: transactions,
	}
	proof, _ := GenerateProof(input)
	outputBytes, err := json.Marshal(proof)
	if err != nil {
		return
	}
	b.StartTimer()
	// // 验证证明
	_ = VerifyProof(string(outputBytes))
	b.StopTimer()
}
