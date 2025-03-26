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
