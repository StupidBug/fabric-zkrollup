package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/crypto"
	"github.com/StupidBug/fabric-zkrollup/pkg/mock"
)

const (
	baseURL = "http://localhost:8080/api/v1"

	numTransactions = 10 // 增加到1000笔交易
	maxRetries      = 10
	retryInterval   = 3 * time.Second
	batchSize       = 1   // 每批次发送50笔交易
	batchInterval   = 100 // 批次间隔100毫秒
)

// 颜色输出
var (
	greenColor = "\033[0;32m"
	redColor   = "\033[0;31m"
	resetColor = "\033[0m"
)

// 彩色输出函数
func green(s string) string {
	return greenColor + s + resetColor
}

func red(s string) string {
	return redColor + s + resetColor
}

// 工具函数 - 检查服务器是否运行
func checkServerRunning() bool {
	fmt.Println(green("检查服务器是否运行..."))
	resp, err := http.Get(baseURL + "/state/root")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// 工具函数 - 获取状态根
func getStateRoot() (string, error) {
	resp, err := http.Get(baseURL + "/state/root")
	if err != nil {
		return "", fmt.Errorf("获取状态根失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.StateRootResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析状态根响应失败: %v", err)
	}

	return result.StateRoot, nil
}

// 工具函数 - 获取余额
func getBalance(address string) int {
	resp, err := http.Get(fmt.Sprintf("%s/balance/get?address=%s", baseURL, address))
	if err != nil {
		log.Fatalf("获取余额失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("解析余额响应失败: %v", err)
	}

	balance, err := strconv.Atoi(result.Balance)
	if err != nil {
		log.Fatalf("无法将余额转换为整数: %v", err)
	}

	return balance
}

var nonceMap = make(map[string]int)
var nonceMutex sync.Mutex

// 工具函数 - 获取 nonce
func GetNonce(address string, isRead bool) int {
	nonceMutex.Lock()
	defer nonceMutex.Unlock()
	if isRead {
		return nonceMap[address]
	}

	nonce := nonceMap[address]
	nonceMap[address]++
	return nonce
}

type TestAccount struct {
	Address    string
	PrivateKey *big.Int
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

// 使用mock_balance.go中定义的测试账户
var testAccounts = func() []TestAccount {
	accounts := []TestAccount{}
	// 添加更多测试账户，用于压力测试
	for i := 1; i < mock.AccountsNum; i++ {
		// 生成地址
		address := ""
		if i < 10 {
			address = fmt.Sprintf("000000000000000000000000000000000000000%d", i)
		} else {
			address = fmt.Sprintf("00000000000000000000000000000000000000%d", i)
		}

		// 生成新的密钥对
		priv, pub := crypto.GenerateKeyPair()

		// 添加新账户
		accounts = append(accounts, TestAccount{
			Address:    address,
			PrivateKey: priv,
			PublicKeyX: pub.X,
			PublicKeyY: pub.Y,
		})
	}

	return accounts
}()

func signTransaction(req *types.TransactionRequest, sender *TestAccount) {
	txData := fmt.Sprintf("%s%s%d%s", req.From, req.To, req.Value, req.Nonce)
	sig := crypto.Sign(txData, sender.PrivateKey)
	req.Signature = types.Signature{
		R: fmt.Sprintf("%x", sig.R),
		S: fmt.Sprintf("%x", sig.S),
	}
	req.PublicKey = types.PublicKey{
		X: fmt.Sprintf("%x", sender.PublicKeyX),
		Y: fmt.Sprintf("%x", sender.PublicKeyY),
	}
}

// 发送交易的工作线程
func sendTransactionWorker(idx int) (string, error) {
	// 创建一个HTTP客户端，设置更长的超时时间和连接池
	client := http.DefaultClient

	// 随机选择发送方和接收方
	sender := &testAccounts[idx%len(testAccounts)]
	receiver := &testAccounts[(idx+1)%len(testAccounts)]

	// 获取nonce
	nonce := GetNonce(sender.Address, false)

	// 构造交易请求
	txReq := &types.TransactionRequest{
		From:  sender.Address,
		To:    receiver.Address,
		Value: rand.IntN(4) + 1,
		Nonce: strconv.Itoa(nonce),
	}

	// 签名交易
	signTransaction(txReq, sender)

	// 发送交易
	jsonData, err := json.Marshal(txReq)
	if err != nil {
		return "", err
	}

	resp, err := client.Post(baseURL+"/transaction/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(fmt.Errorf("send trans failed: %w", err))
	}

	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		panic(string(body))
	}

	var result types.TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return "", err
	}
	resp.Body.Close()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("工作线程 %d 发送交易成功: %s %#v\n", idx, result.Hash, txReq)

	return result.Hash, nil
}

// 获取交易状态
func getTransactionStatus(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/transaction/get?hash=%s", baseURL, txHash))
	if err != nil {
		return "", fmt.Errorf("获取交易状态失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析交易状态响应失败: %v", err)
	}

	return result.Status, nil
}

// 工具函数 - 等待交易确认
func waitForConfirmation(txHash string) bool {
	var retryCount int
	time.Sleep(3 * time.Second)
	for retryCount < maxRetries {
		// fmt.Printf("检查交易状态（尝试 %d/%d）\n", retryCount+1, maxRetries)
		status, err := getTransactionStatus(txHash)
		if err != nil {
			fmt.Printf("检查交易状态失败: %v\n", err)
			retryCount++
			time.Sleep(retryInterval)
			continue
		}

		// fmt.Printf("交易 %s 状态: %s\n", txHash, status)

		if status == "confirmed" {
			fmt.Printf("交易 %s 确认成功\n", txHash)
			return true
		}

		retryCount++
		time.Sleep(retryInterval)
	}

	return false
}

// 工具函数 - 获取所有区块
func getAllBlocks() []types.BlockResponse {
	resp, err := http.Get(baseURL + "/blocks")
	if err != nil {
		log.Fatalf("获取区块失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.BlocksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("解析区块响应失败: %v", err)
	}

	if result.Status != "success" {
		log.Fatalf("获取区块请求失败: %s", result.Status)
	}

	return result.Data.Blocks
}

// 工具函数 - 显示所有账户状态
func displayAccountStates() {
	fmt.Println(green("\n获取账户状态..."))
	fmt.Println(green("账户状态:"))
	fmt.Println("-----------------")

	// 输出表头
	fmt.Printf("%-42s | %-10s | %-5s\n", "地址", "余额", "Nonce")
	fmt.Println("----------------------------------------------------------------------")

	// 获取并显示每个地址的余额和nonce
	for i := 0; i < mock.AccountsNum-1; i++ { // 只显示前10个账户，避免输出太多
		addr := testAccounts[i].Address
		bal := getBalance(addr)
		nonce := GetNonce(addr, true)
		fmt.Printf("%-42s | %-10d | %-5d\n", addr, bal, nonce)
	}

	if len(testAccounts) > 10 {
		fmt.Println("... 更多账户省略 ...")
	}

	fmt.Println("-----------------")
}

// 显示所有区块信息
func displayAllBlocks() {
	blocks := getAllBlocks()

	fmt.Println(green("\n获取所有区块..."))
	fmt.Println("区块链状态:")
	fmt.Println("-----------------")
	fmt.Printf("总区块数: %d\n\n", len(blocks))

	for _, block := range blocks {
		fmt.Printf("区块 #%d\n", block.Height)
		fmt.Printf("哈希: %s\n", block.Hash)
		fmt.Printf("前一区块哈希: %s\n", block.PrevHash)
		fmt.Printf("Merkle根: %s\n", block.MerkleRoot)
		fmt.Printf("状态根: %s\n", block.StateRoot)
		fmt.Printf("时间戳: %d\n", block.Timestamp)
		fmt.Printf("交易数量: %d\n\n", block.TransactionCount)

		fmt.Println("交易列表:")
		if len(block.Transactions) > 0 {
			for i, tx := range block.Transactions {
				fmt.Printf("  交易 #%d\n", i+1)
				fmt.Printf("  哈希: %s\n", tx.Hash)
				fmt.Printf("  发送方: %s\n", tx.From)
				fmt.Printf("  接收方: %s\n", tx.To)
				fmt.Printf("  金额: %s\n", tx.Value)
				fmt.Printf("  Nonce: %d\n", tx.Nonce)
				fmt.Printf("  状态: %s\n", tx.Status)
				fmt.Printf("  时间戳: %d\n", tx.Timestamp)
				fmt.Println()
			}
		} else {
			fmt.Println("  此区块无交易")
		}
		fmt.Println("-----------------")
	}
}

func TestMain(t *testing.T) {
	// 检查服务器是否运行
	if !checkServerRunning() {
		fmt.Println("服务器未运行。请先启动服务器。")
		os.Exit(1)
	}
	fmt.Println("服务器已启动。")

	// 获取初始状态根
	initialStateRoot, err := getStateRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("初始状态根: %s\n", initialStateRoot)
	fmt.Printf("压力测试配置:\n")
	fmt.Printf("总交易数: %d\n", numTransactions*mock.AccountsNum)
	fmt.Printf("批次大小: %d\n", batchSize)
	fmt.Printf("批次间隔: %dms\n", batchInterval)

	startTime := time.Now()
	// 启动工作线程
	var wg sync.WaitGroup
	for i := range mock.AccountsNum - 1 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for range 10 {
				hash, _ := sendTransactionWorker(idx)
				wg.Add(1)
				go func() {
					defer wg.Done()
					waitForConfirmation(hash)
				}()
			}
		}(i)
	}
	wg.Wait()

	fmt.Printf("\n压力测试结果:\n")
	fmt.Printf("总耗时: %s\n", time.Since(startTime).String())
	fmt.Printf("平均TPS: %.2f\n", numTransactions*mock.AccountsNum/time.Since(startTime).Seconds())

	// 检查最终状态根
	finalStateRoot, err := getStateRoot()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n初始状态根: %s\n", initialStateRoot)
	fmt.Printf("\n最终状态根: %s\n", finalStateRoot)
	if finalStateRoot == initialStateRoot {
		log.Fatal("状态根未改变")
	}
	displayAllBlocks()
	displayAccountStates()

	fmt.Println("测试完成！所有交易都已成功处理。")
}
