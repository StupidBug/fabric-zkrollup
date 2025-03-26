package types

// 区块相关请求和响应结构体

// BlockResponse represents a block response
type BlockResponse struct {
	Height           uint64                `json:"height"`
	Hash             string                `json:"hash"`
	PrevHash         string                `json:"prevHash"`
	MerkleRoot       string                `json:"merkleRoot"`
	StateRoot        string                `json:"stateRoot"`
	Timestamp        int64                 `json:"timestamp"`
	TransactionCount uint32                `json:"transactionCount"`
	Transactions     []TransactionResponse `json:"transactions"`
}

// BlocksData represents the data field in blocks response
type BlocksData struct {
	Blocks []BlockResponse `json:"blocks"`
}

// BlocksResponse represents the blocks query response
type BlocksResponse struct {
	Status string     `json:"status"`
	Data   BlocksData `json:"data"`
}
