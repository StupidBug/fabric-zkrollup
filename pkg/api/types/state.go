package types

// 状态相关请求和响应结构体

// StateRootResponse represents a state root query response
type StateRootResponse struct {
	StateRoot string `json:"stateRoot"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
}
