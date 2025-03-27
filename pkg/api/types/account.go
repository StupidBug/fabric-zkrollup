package types

// 账户相关请求和响应结构体

// BalanceResponse represents a balance query response
type BalanceResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// NonceResponse represents a nonce query response
type NonceResponse struct {
	Address string `json:"address"`
	Nonce   int    `json:"nonce"`
}

// Account 表示账户状态
type Account struct {
	Address string // 电路外：账户地址为string类型
	Balance int    // 电路外：账户余额为int类型
	Nonce   int    // 电路外：nonce为int类型
}

// Transaction 表示交易
type Transaction struct {
	From   string // 电路外：发送者地址为string类型
	To     string // 电路外：接收者地址为string类型
	Amount int    // 电路外：转账金额为int类型
	Nonce  int    // 电路外：交易nonce为int类型
}
