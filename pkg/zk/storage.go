package zk

type StorageRollup interface {
	Package(*StorageProofInput) *StorageProofOutput // 打包
	ConputeInitialStatRoot([]string) string         // 初始状态根
}

// 输入参数结构体
type StorageProofInput struct {
	OldStateRoot string   // 旧状态根
	Evidence     []string // 存证
}

// 输出参数结构体
type StorageProofOutput struct {
	OldStateRoot string
	BatchRoot    string
	NewStateRoot string
	Evidence     []string    // 添加新的账户状态字段
	Proof        interface{} // 使用interface{}来存储proof
	Vk           interface{} // 使用interface{}来存储vk
}
