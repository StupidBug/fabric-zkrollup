package types

// 交易相关请求和响应结构体

// PublicKey represents the public key in API request/response
type PublicKey struct {
	X string `json:"x" binding:"required"`
	Y string `json:"y" binding:"required"`
}

// Signature represents the signature in API request/response
type Signature struct {
	R string `json:"r" binding:"required"`
	S string `json:"s" binding:"required"`
}

// TransactionRequest represents a transaction submission request
type TransactionRequest struct {
	From      string    `json:"from" binding:"required"`
	To        string    `json:"to" binding:"required"`
	Value     string    `json:"value" binding:"required"`
	Nonce     string    `json:"nonce" binding:"required"`
	Signature Signature `json:"signature" binding:"required"`
	PublicKey PublicKey `json:"publicKey" binding:"required"`
}

// TransactionResponse represents a transaction response
type TransactionResponse struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     string    `json:"value"`
	Nonce     uint64    `json:"nonce"`
	Status    string    `json:"status"`
	Timestamp int64     `json:"timestamp"`
	Signature Signature `json:"signature,omitempty"`
	PublicKey PublicKey `json:"publicKey,omitempty"`
}

// TransactionRequest represents a transaction request/response with numeric status
type TransactionWithNumStatus struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     int       `json:"value"`
	Nonce     uint64    `json:"nonce"`
	Status    int       `json:"status"`
	Timestamp int64     `json:"timestamp"`
	Signature Signature `json:"signature,omitempty"`
	PublicKey PublicKey `json:"publicKey,omitempty"`
}
