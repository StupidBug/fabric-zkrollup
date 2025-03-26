package mock

import "github.com/StupidBug/fabric-zkrollup/pkg/zk"

func MockBalance() []zk.Account {
	return []zk.Account{
		{
			Address: "0000000000000000000000000000000000000001",
			Balance: 1000000,
			Nonce:   0,
		},
		{
			Address: "0000000000000000000000000000000000000002",
			Balance: 500000,
			Nonce:   0,
		},
		{
			Address: "0000000000000000000000000000000000000003",
			Balance: 300000,
			Nonce:   0,
		},
	}
}
