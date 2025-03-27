package zk

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/mock"
)

func BenchmarkGenerateProof(b *testing.B) {
	b.StopTimer()
	accounts := mock.MockBalance()
	transactions := mock.MockTransaction()
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
}

func BenchmarkValidateProof(b *testing.B) {
	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	accounts := mock.MockBalance()
	old_merkleRoot := ComputeAccountMerkleRoot(accounts)
	input := ProofInput{
		OldStateRoot: old_merkleRoot,
		Accounts:     accounts,
		Transactions: mock.MockTransaction(),
	}
	proof, _ := GenerateProof(input)
	outputBytes, err := json.Marshal(proof)
	if err != nil {
		return
	}
	b.StartTimer()
	// // 验证证明
	err = VerifyProof(string(outputBytes))
	if err != nil {
		panic(err)
	}
}
