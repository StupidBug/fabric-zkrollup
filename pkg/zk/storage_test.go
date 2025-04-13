package zk

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	// 创建初始存证数据
	initialEvidence := []string{
		"initial1: 这是初始存证数据1",
		"initial2: 这是初始存证数据2",
	}

	// 计算初始状态根
	oldStateRoot := ComputeStorageMerkleRoot(initialEvidence)

	// 创建100条新的存证数据
	evidence := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		evidence[i] = fmt.Sprintf("evidence%d: 这是第%d条测试存证数据，包含一些随机内容：%d", i+1, i+1, i*12345)
	}

	// 创建输入参数
	input := StorageProofInput{
		OldStateRoot: oldStateRoot, // 使用初始存证数据计算的状态根
		Evidence:     evidence,
	}

	// 生成证明
	output, err := Package(input)
	if err != nil {
		fmt.Printf("生成证明失败: %v\n", err)
		return
	}

	// 打印输出结果
	fmt.Printf("初始存证数据:\n")
	for i, ev := range initialEvidence {
		fmt.Printf("  存证%d: %s\n", i+1, ev)
	}
	fmt.Printf("初始状态根: %s\n", oldStateRoot)
	fmt.Printf("\n新存证数据数量: %d\n", len(output.Evidence))
	fmt.Printf("批次根: %s\n", output.BatchRoot)
	fmt.Printf("新状态根: %s\n", output.NewStateRoot)

	// 序列化输出
	outputBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Printf("序列化输出失败: %v\n", err)
		return
	}

	// 验证证明
	err = VerifyStorageProof(string(outputBytes))
	if err != nil {
		fmt.Printf("验证证明失败: %v\n", err)
		return
	}

	fmt.Println("存证场景测试成功！")
}
