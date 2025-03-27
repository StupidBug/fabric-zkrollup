package zk

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/mock"
)

func TestRollup(t *testing.T) {
	// 创建固定的账户状态
	accounts := mock.MockBalance()
	transactions := mock.MockTransaction()
	fmt.Printf("accounts: %#v\n", accounts)

	// 计算旧状态根
	old_merkleRoot := ComputeAccountMerkleRoot(accounts)
	fmt.Printf("old_merkleRoot: %v\n", old_merkleRoot)

	input := ProofInput{
		OldStateRoot: old_merkleRoot,
		Accounts:     accounts,
		Transactions: transactions,
	}

	// 生成证明
	output, err := GenerateProof(input)
	if err != nil {
		fmt.Printf("Failed to generate proof: %v\n", err)
		return
	}

	// 序列化输出
	outputBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Printf("Failed to marshal proof output: %v\n", err)
		return
	}

	// 验证证明
	err = VerifyProof(string(outputBytes))
	if err != nil {
		fmt.Printf("Failed to verify proof: %v\n", err)
		return
	}

	fmt.Println("Proof verification succeeded!")
}

func TestStateRoot(t *testing.T) {
	accounts := []types.Account{
		types.Account{Address: "0000000000000000000000000000000000000001", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000002", Balance: 500004, Nonce: 6},
		types.Account{Address: "0000000000000000000000000000000000000003", Balance: 499996, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000004", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000005", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000006", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000007", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000008", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000009", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000010", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000011", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000012", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000013", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000014", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000015", Balance: 500004, Nonce: 6},
		types.Account{Address: "0000000000000000000000000000000000000016", Balance: 499996, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000017", Balance: 500000, Nonce: 10},
		types.Account{Address: "0000000000000000000000000000000000000018", Balance: 500004, Nonce: 6},
		types.Account{Address: "0000000000000000000000000000000000000019", Balance: 499996, Nonce: 10},
	}
	transactions := []types.Transaction{
		types.Transaction{From: "0000000000000000000000000000000000000002", To: "0000000000000000000000000000000000000003", Amount: 1, Nonce: 6},
		types.Transaction{From: "0000000000000000000000000000000000000002", To: "0000000000000000000000000000000000000003", Amount: 1, Nonce: 7},
		types.Transaction{From: "0000000000000000000000000000000000000002", To: "0000000000000000000000000000000000000003", Amount: 1, Nonce: 8},
		types.Transaction{From: "0000000000000000000000000000000000000002", To: "0000000000000000000000000000000000000003", Amount: 1, Nonce: 9},
	}
	input := ProofInput{
		OldStateRoot: ComputeAccountMerkleRoot(accounts),
		Accounts:     accounts,
		Transactions: transactions,
	}
	fmt.Printf("old state root: %s\n", input.OldStateRoot)
	// 生成证明
	output, err := GenerateProof(input)
	if err != nil {
		fmt.Printf("Failed to generate proof: %v\n", err)
		return
	}
	fmt.Printf("new state root: %s\n", output.NewStateRoot)
}
