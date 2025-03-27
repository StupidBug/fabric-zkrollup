package mock

import (
	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
)

const AccountsNum = 20

// MockBalance 返回用于测试的初始账户状态
func MockBalance() []types.Account {
	accounts := []types.Account{}
	// 添加更多测试账户，用于压力测试
	for i := 1; i < AccountsNum; i++ {
		address := ""
		if i < 10 {
			address = "000000000000000000000000000000000000000" + string(rune('0'+i))
		} else {
			address = "00000000000000000000000000000000000000" + string(rune('0'+(i/10))) + string(rune('0'+(i%10)))
		}
		accounts = append(accounts, types.Account{
			Address: address,
			Balance: 500000,
			Nonce:   0,
		})
	}
	return accounts
}

// MockBalance 返回用于测试的初始账户状态
func MockTransaction() []types.Transaction {
	var transactions []types.Transaction
	for i := 0; i < 512; i++ {
		transactions = append(transactions,
			types.Transaction{
				From:   "0000000000000000000000000000000000000001",
				To:     "0000000000000000000000000000000000000002",
				Amount: 1,
				Nonce:  i,
			},
		)
	}
	return transactions
}

// TestAccount 表示一个测试账户，包含私钥和公钥信息
type TestAccount struct {
	Address    string
	PrivateKey string
	PublicKeyX string
	PublicKeyY string
}
