package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/crypto"
)

const (
	baseURL         = "http://localhost:8080/api/v1"
	senderAddress   = "0000000000000000000000000000000000000001"
	receiverAddress = "0000000000000000000000000000000000000002"
	transferAmount  = 100
	numTransactions = 3
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
func getStateRoot() string {
	fmt.Println(green("获取状态根..."))
	resp, err := http.Get(baseURL + "/state/root")
	if err != nil {
		log.Fatalf("获取状态根失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.StateRootResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("解析状态根响应失败: %v", err)
	}

	return result.StateRoot
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

// 工具函数 - 获取 nonce
func getNonce(address string) int {
	resp, err := http.Get(fmt.Sprintf("%s/account/nonce?address=%s", baseURL, address))
	if err != nil {
		log.Fatalf("获取nonce失败: %v", err)
	}
	defer resp.Body.Close()

	var result types.NonceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("解析nonce响应失败: %v", err)
	}

	return result.Nonce
}

// 工具函数 - 发送交易
func sendTransaction(from, to string, value, nonce int, sigR, sigS, pubX, pubY string) string {
	txReq := types.TransactionRequest{
		From:  from,
		To:    to,
		Value: strconv.Itoa(value),
		Nonce: strconv.Itoa(nonce),
		Signature: types.Signature{
			R: sigR,
			S: sigS,
		},
		PublicKey: types.PublicKey{
			X: pubX,
			Y: pubY,
		},
	}

	jsonData, err := json.Marshal(txReq)
	if err != nil {
		log.Fatalf("序列化交易请求失败: %v", err)
	}

	fmt.Println(green("发送交易..."))
	resp, err := http.Post(baseURL+"/transaction/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取交易响应失败: %v", err)
	}

	fmt.Printf("交易响应: %s\n", string(body))

	// 我们不再需要在不同格式之间切换解析。服务器回复的就是TransactionWithNumStatus
	var tx types.TransactionWithNumStatus
	if err := json.Unmarshal(body, &tx); err != nil {
		log.Fatalf("解析交易响应失败: %v", err)
	}

	if tx.Hash != "" {
		return tx.Hash
	}

	log.Fatalf("无法从响应中获取交易哈希")
	return ""
}

// 工具函数 - 检查交易状态
func getTransactionStatus(hash string) string {
	resp, err := http.Get(fmt.Sprintf("%s/transaction/get?hash=%s", baseURL, hash))
	if err != nil {
		log.Printf("获取交易状态失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取交易状态响应失败: %v", err)
		return ""
	}

	fmt.Printf("交易 %s 完整响应: %s\n", hash, string(body))

	// 直接解析为TransactionResponse对象
	var tx types.TransactionResponse
	if err := json.Unmarshal(body, &tx); err == nil && tx.Status != "" {
		return tx.Status
	}

	// 数字状态码的情况
	var txWithNumStatus types.TransactionWithNumStatus
	if err := json.Unmarshal(body, &txWithNumStatus); err == nil {
		switch txWithNumStatus.Status {
		case 0:
			return "pending"
		case 1:
			return "confirmed"
		case 2:
			return "failed"
		}
	}

	return ""
}

// 工具函数 - 等待交易确认
func waitForConfirmation(hash string, maxRetries int) bool {
	var retryCount int
	for retryCount < maxRetries {
		fmt.Printf("检查交易状态（尝试 %d/%d）\n", retryCount+1, maxRetries)
		status := getTransactionStatus(hash)
		fmt.Printf("交易 %s 状态: %s\n", hash, status)

		if status == "confirmed" {
			fmt.Println("交易确认成功")
			return true
		} else if status == "pending" {
			fmt.Println("交易处理中...")
		} else if status == "" {
			fmt.Println("未获取到交易状态")
		} else {
			fmt.Printf("未知状态: %s\n", status)
		}

		retryCount++
		time.Sleep(3 * time.Second)
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

	addresses := []string{senderAddress, receiverAddress}

	// 输出表头
	fmt.Printf("%-42s | %-10s | %-5s\n", "地址", "余额", "Nonce")
	fmt.Println("----------------------------------------------------------------------")

	// 获取并显示每个地址的余额和nonce
	for _, addr := range addresses {
		bal := getBalance(addr)
		nonce := getNonce(addr)
		fmt.Printf("%-42s | %-10d | %-5d\n", addr, bal, nonce)
	}

	fmt.Println("-----------------")
}

// 重建密钥生成工具
func rebuildKeygenTool() {
	fmt.Println(green("重新编译 keygen 工具..."))
	cmd := exec.Command("go", "build", "-o", "keygen", "./cmd/keygen")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(red("构建 keygen 工具失败:"), string(output))
		os.Exit(1)
	}
	fmt.Println(green("keygen 工具构建成功"))
}

// 生成密钥对
func generateKeyPair() (string, string, string) {
	priv, pub := crypto.GenerateKeyPair()
	privHex := fmt.Sprintf("%x", priv)
	pubXHex := fmt.Sprintf("%x", pub.X)
	pubYHex := fmt.Sprintf("%x", pub.Y)
	return privHex, pubXHex, pubYHex
}

// 签名交易
func signTransaction(from, to string, value, nonce int, privKey string) (string, string, string, string, string) {
	// 将私钥转为 big.Int
	privKeyBig := new(big.Int)
	privKeyBig.SetString(privKey, 16)

	// 创建交易数据
	txData := fmt.Sprintf("%s%s%d%d", from, to, value, nonce)

	// 签名交易
	sig := crypto.Sign(txData, privKeyBig)
	pub := crypto.PrivateKeyToPublic(privKeyBig)

	return txData, fmt.Sprintf("%x", sig.R), fmt.Sprintf("%x", sig.S), fmt.Sprintf("%x", pub.X), fmt.Sprintf("%x", pub.Y)
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

func main() {
	// 检查服务器是否运行
	if !checkServerRunning() {
		fmt.Println(red("服务器未运行。请先启动服务器。"))
		os.Exit(1)
	}
	fmt.Println(green("服务器已启动。"))

	// 重新编译 keygen 工具
	rebuildKeygenTool()

	// 获取初始状态根
	initialRoot := getStateRoot()
	fmt.Printf("初始状态根: %s\n", initialRoot)

	// 生成发送方密钥对
	fmt.Println(green("生成发送方密钥对..."))
	senderPrivKey, senderPubX, senderPubY := generateKeyPair()
	fmt.Printf("私钥: %s\n", senderPrivKey)
	fmt.Printf("公钥 X: %s\n", senderPubX)
	fmt.Printf("公钥 Y: %s\n", senderPubY)

	// 获取初始余额
	fmt.Println(green("初始余额:"))
	senderInitialBalance := getBalance(senderAddress)
	receiverInitialBalance := getBalance(receiverAddress)
	fmt.Printf("发送方 (%s): %d\n", senderAddress, senderInitialBalance)
	fmt.Printf("接收方 (%s): %d\n", receiverAddress, receiverInitialBalance)

	// 发送多笔交易
	totalTransfer := 0
	for i := 0; i < numTransactions; i++ {
		fmt.Printf(green("处理交易 %d/%d...\n"), i+1, numTransactions)

		// 获取当前 nonce
		currentNonce := getNonce(senderAddress)
		fmt.Printf("当前 nonce: %d\n", currentNonce)

		// 签名交易
		fmt.Println(green("签名交易..."))
		txHash, sigR, sigS, pubX, pubY := signTransaction(senderAddress, receiverAddress, transferAmount, currentNonce, senderPrivKey)
		fmt.Printf("交易哈希: %s\n", txHash)
		fmt.Printf("签名 R: %s\n", sigR)
		fmt.Printf("签名 S: %s\n", sigS)
		fmt.Printf("公钥 X: %s\n", pubX)
		fmt.Printf("公钥 Y: %s\n", pubY)

		// 发送交易
		hash := sendTransaction(senderAddress, receiverAddress, transferAmount, currentNonce, sigR, sigS, pubX, pubY)
		totalTransfer += transferAmount

		// 等待交易确认
		fmt.Printf(green("等待交易 %d 确认...\n"), i+1)
		if !waitForConfirmation(hash, 10) {
			fmt.Printf(red("测试失败: 交易 %d 未及时确认\n"), i+1)
			os.Exit(1)
		}
		fmt.Printf(green("交易 %d 成功确认\n"), i+1)

		// 获取交易后的余额
		fmt.Printf(green("检查交易 %d 后的余额...\n"), i+1)
		currentSenderBalance := getBalance(senderAddress)
		currentReceiverBalance := getBalance(receiverAddress)
		fmt.Printf("当前发送方余额: %d\n", currentSenderBalance)
		fmt.Printf("当前接收方余额: %d\n", currentReceiverBalance)

		// 小延迟
		time.Sleep(2 * time.Second)
	}

	// 获取最终余额
	fmt.Println(green("最终余额:"))
	senderFinalBalance := getBalance(senderAddress)
	receiverFinalBalance := getBalance(receiverAddress)
	fmt.Printf("发送方 (%s): %d\n", senderAddress, senderFinalBalance)
	fmt.Printf("接收方 (%s): %d\n", receiverAddress, receiverFinalBalance)

	// 获取最终状态根
	finalRoot := getStateRoot()
	fmt.Printf("最终状态根: %s\n", finalRoot)

	// 验证状态根是否改变
	if initialRoot != finalRoot {
		fmt.Println(green("测试通过: 状态根在转账后发生改变"))
	} else {
		fmt.Println(red("测试失败: 状态根未改变"))
		os.Exit(1)
	}

	// 验证余额是否正确变化
	expectedSenderBalance := senderInitialBalance - totalTransfer
	expectedReceiverBalance := receiverInitialBalance + totalTransfer

	if senderFinalBalance == expectedSenderBalance && receiverFinalBalance == expectedReceiverBalance {
		fmt.Println(green("测试通过: 余额正确更新"))
	} else {
		fmt.Println(red("测试失败: 余额未正确更新"))
		fmt.Printf("预期发送方余额: %d, 实际: %d\n", expectedSenderBalance, senderFinalBalance)
		fmt.Printf("预期接收方余额: %d, 实际: %d\n", expectedReceiverBalance, receiverFinalBalance)
		os.Exit(1)
	}

	fmt.Println(green("所有测试均通过!"))

	// 显示所有区块信息
	displayAllBlocks()

	// 显示所有账户状态
	displayAccountStates()

	fmt.Println(green("测试套件执行完成。"))
}
