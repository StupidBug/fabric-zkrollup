package blockchain

import (
	"sort"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/block"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"
)

func (bc *Blockchain) PopTransactions() []transaction.Transaction {
	transactions := bc.txPool.Pop(maxTransactions)
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].From != transactions[j].From {
			return transactions[i].From < transactions[j].From
		}
		return transactions[i].Nonce < transactions[j].Nonce
	})
	return transactions
}

func (bc *Blockchain) GetTransactionForProof(block *block.Block) []types.Transaction {
	var transactions []types.Transaction
	for _, tx := range block.Transactions {
		transactions = append(transactions, types.Transaction{
			From:   tx.From,
			To:     tx.To,
			Amount: tx.Value,
			Nonce:  int(tx.Nonce),
		})
	}
	return transactions
}

func (bc *Blockchain) GetAccountsForProof() []types.Account {
	var accounts []types.Account
	allAccounts := bc.state.GetAllAccounts()
	for addr, acc := range allAccounts {
		accounts = append(accounts, types.Account{
			Address: addr,
			Balance: acc.Balance,
			Nonce:   int(acc.Nonce),
		})
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Address < accounts[j].Address
	})
	return accounts
}
