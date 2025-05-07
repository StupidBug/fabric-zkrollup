package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
)

const (
	baseURL       = "http://localhost:8080/api/v1"
	maxRetries    = 5
	maxWorkers    = 10
	taskPerWorker = 10
	retryInterval = 3 * time.Second
)

// 发送交易的工作线程
func sendTransactionWorker() (string, error) {
	// 创建一个HTTP客户端，设置更长的超时时间和连接池
	client := http.DefaultClient

	// 构造交易请求
	req := &types.ProofStorageReq{
		Evidence: []string{fmt.Sprintf("%d: %d", rand.IntN(1000), rand.IntN(1000))},
	}

	// 发送交易
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := client.Post(baseURL+"/proof/storage", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(fmt.Errorf("send trans failed: %w", err))
	}

	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		panic(string(body))
	}

	var result types.ProofStorageResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return "", err
	}
	resp.Body.Close()
	time.Sleep(100 * time.Millisecond)
	// fmt.Printf("工作线程 %d 发送交易成功: %s %#v\n", idx, result.Hash, txReq)

	return result.Hashs[0], nil
}

// 获取交易状态
func getTransactionStatus(hash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/proof/status?hash=%s", baseURL, hash))
	if err != nil {
		return "", fmt.Errorf("获取存证状态失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.ProofStatusResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析存证状态响应失败: %v", err)
	}

	return result.Status, nil
}

// 工具函数 - 等待交易确认
func waitForConfirmation(hash string) bool {
	var retryCount int
	time.Sleep(3 * time.Second)
	for retryCount < maxRetries {
		// fmt.Printf("检查交易状态（尝试 %d/%d）\n", retryCount+1, maxRetries)
		status, err := getTransactionStatus(hash)
		if err != nil {
			fmt.Printf("检查交易状态失败: %v\n", err)
			retryCount++
			time.Sleep(retryInterval)
			continue
		}

		// fmt.Printf("交易 %s 状态: %s\n", txHash, status)

		if status == "confirmed" {
			fmt.Printf("存证 %s 确认成功\n", hash)
			return true
		}

		retryCount++
		time.Sleep(retryInterval)
	}

	return false
}

func TestMain(t *testing.T) {
	startTime := time.Now()
	// 启动工作线程
	var wg sync.WaitGroup
	for i := range maxWorkers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for range taskPerWorker {
				hash, _ := sendTransactionWorker()
				wg.Add(1)
				go func() {
					defer wg.Done()
					waitForConfirmation(hash)
				}()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("总耗时: %s\n", time.Since(startTime).String())
	fmt.Printf("平均TPS: %.2f\n", maxWorkers*taskPerWorker/time.Since(startTime).Seconds())
}
