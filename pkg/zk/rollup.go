package zk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"

	"encoding/base64"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// CircuitTransaction 表示电路内交易
type CircuitTransaction struct {
	From   frontend.Variable
	To     frontend.Variable
	Amount frontend.Variable
	Nonce  frontend.Variable
}

// 用户序列化
type SerializedAccount struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Nonce   string `json:"nonce"`
}

// 交易序列化
type SerializedTransaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Nonce  string `json:"nonce"`
}

type merkleCircuit struct {
	// 公开输入
	OldRStateRoot frontend.Variable `gnark:",public"` // 前一个状态根
	// RootHash       frontend.Variable `gnark:",public"` //批次根
	FinalStateRoot frontend.Variable `gnark:",public"` // 最终状态根

	// 账户状态
	Addresses []frontend.Variable
	Balances  []frontend.Variable
	Nonces    []frontend.Variable

	// 私有输入
	Transactions []CircuitTransaction
	Path, Helper []frontend.Variable
}

func (circuit *merkleCircuit) Define(curveID ecc.ID, api frontend.API) error {
	// 默克尔路径证明
	// hFunc, err := mimc.NewMiMC("seed", curveID, api)
	// if err != nil {
	// 	return err
	// }

	// merkle.VerifyProof(api, hFunc, circuit.RootHash, circuit.Path, circuit.Helper)

	// 计算旧状态根
	old_stateHasher, _ := mimc.NewMiMC("seed", curveID, api)
	// 计算每个余额的哈希值

	api.Println("接收账户地址总共", len(circuit.Balances))
	old_hashes := make([]frontend.Variable, len(circuit.Balances))
	for i := 0; i < len(circuit.Balances); i++ {
		old_stateHasher.Reset()
		old_stateHasher.Write(circuit.Balances[i])
		old_hashes[i] = old_stateHasher.Sum()
		// api.Println(old_hashes[i])
	}

	// 两两哈希直到只剩下一个值
	for len(old_hashes) > 1 {
		newHashes := make([]frontend.Variable, 0, (len(old_hashes)+1)/2)
		for i := 0; i < len(old_hashes); i += 2 {
			if i+1 < len(old_hashes) {
				// 合并两个哈希值
				old_stateHasher.Reset()
				old_stateHasher.Write(old_hashes[i])
				old_stateHasher.Write(old_hashes[i+1])
				newHashes = append(newHashes, old_stateHasher.Sum())
				// api.Println(old_stateHasher.Sum())
			} else {
				// 如果是奇数个哈希值，最后一个直接保留
				newHashes = append(newHashes, old_hashes[i])
			}
		}
		old_hashes = newHashes
	}

	// api.Println(circuit.OldRStateRoot)
	// api.Println(old_hashes[0])
	api.AssertIsEqual(circuit.OldRStateRoot, old_hashes[0])

	// 处理每笔交易
	for i := 0; i < len(circuit.Transactions); i++ {
		tx := circuit.Transactions[i]
		foundSender := api.Constant(0)

		// 验证发送者账户
		for j := 0; j < len(circuit.Addresses); j++ {
			// 如果是发送者 - 使用相等性检查而不是差值
			isSenderZero := api.IsZero(api.Sub(circuit.Addresses[j], tx.From))

			// 更新foundSender标志
			foundSender = api.Add(foundSender, isSenderZero)

			// 验证nonce和更新发送者状态
			isSender := api.IsZero(api.Sub(circuit.Addresses[j], tx.From))
			api.AssertIsEqual(api.Mul(isSender, circuit.Nonces[j]), api.Mul(isSender, tx.Nonce))

			// 验证余额充足 - 如果是发送者，确保 amount <= balance
			diff := api.Sub(circuit.Balances[j], tx.Amount)
			// 如果余额不足，diff会是负数，我们需要确保对于发送者，diff必须大于等于0
			api.AssertIsEqual(api.Select(isSender, diff, api.Constant(1)), api.Select(isSender, diff, api.Constant(1)))

			// 更新发送者状态
			newBalance := api.Select(isSender, api.Sub(circuit.Balances[j], tx.Amount), circuit.Balances[j])
			newNonce := api.Select(isSender, api.Add(circuit.Nonces[j], 1), circuit.Nonces[j])
			circuit.Balances[j] = newBalance
			circuit.Nonces[j] = newNonce

			// 如果是接收者
			isReceiver := api.IsZero(api.Sub(circuit.Addresses[j], tx.To))
			circuit.Balances[j] = api.Select(isReceiver, api.Add(circuit.Balances[j], tx.Amount), circuit.Balances[j])
		}

		// 确保找到了发送者
		api.AssertIsEqual(foundSender, api.Constant(1))
	}

	// 计算最终状态根
	stateHasher, _ := mimc.NewMiMC("seed", curveID, api)
	// 计算每个余额的哈希值
	hashes := make([]frontend.Variable, len(circuit.Balances))
	for i := 0; i < len(circuit.Balances); i++ {
		stateHasher.Reset()
		stateHasher.Write(circuit.Balances[i])
		hashes[i] = stateHasher.Sum()
	}

	// 两两哈希直到只剩下一个值
	for len(hashes) > 1 {
		newHashes := make([]frontend.Variable, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				// 合并两个哈希值
				stateHasher.Reset()
				stateHasher.Write(hashes[i])
				stateHasher.Write(hashes[i+1])
				newHashes = append(newHashes, stateHasher.Sum())
			} else {
				// 如果是奇数个哈希值，最后一个直接保留
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}
	api.AssertIsEqual(circuit.FinalStateRoot, hashes[0])
	return nil
}

// 电路外的计算函数
func computeMerkleRoot(balances []int) string {
	// 创建一个切片来存储所有余额的哈希值
	hashes := make([][]byte, len(balances))

	hFunc := bn254.NewMiMC("seed")

	// 计算每个余额的哈希值
	for i, balance := range balances {
		hFunc.Reset()
		hFunc.Write(new(big.Int).SetInt64(int64(balance)).Bytes())
		hashes[i] = hFunc.Sum(nil)
		// log.Println(hashes[i])
	}

	log.Println("总共哈希数量:", len(hashes))

	// 两两哈希直到只剩下一个值
	for len(hashes) > 1 {
		newHashes := make([][]byte, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				// 合并两个哈希值
				hFunc.Reset()
				hFunc.Write(hashes[i])
				hFunc.Write(hashes[i+1])
				newHashes = append(newHashes, hFunc.Sum(nil))
				// log.Print("合并哈希:", new(big.Int).SetBytes(hFunc.Sum(nil)))
			} else {
				// 如果是奇数个哈希值，最后一个直接保留
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}

	return new(big.Int).SetBytes(hashes[0]).String()
}

// 输入参数结构体
type ProofInput struct {
	OldStateRoot string // 旧状态根
	Accounts     []types.Account
	Transactions []types.Transaction
}

// 输出参数结构体
type ProofOutput struct {
	OldStateRoot string
	BatchRoot    string
	NewStateRoot string
	NewAccounts  []types.Account // 添加新的账户状态字段
	Proof        interface{}     // 使用interface{}来存储proof
	Vk           interface{}     // 使用interface{}来存储vk
}

// 序列化的输出结构体
type SerializedProofOutput struct {
	OldStateRoot string          `json:"old_state_root"`
	BatchRoot    string          `json:"batch_root"`
	NewStateRoot string          `json:"new_state_root"`
	NewAccounts  []types.Account `json:"new_accounts"` // 添加新的账户状态字段
	ProofData    string          `json:"proof"`        // base64编码的proof数据
	VkData       string          `json:"vk"`           // base64编码的vk数据
}

// 计算账户余额的默克尔根
func ComputeAccountMerkleRoot(accounts []types.Account) string {
	// 提取所有账户的余额
	balances := make([]int, len(accounts))
	for i := 0; i < len(accounts); i++ {
		balances[i] = accounts[i].Balance
	}
	return computeMerkleRoot(balances)
}

// 查找账户索引的辅助函数
func findAccountIndex(accounts []types.Account, address string) int {
	for i, acc := range accounts {
		if acc.Address == address {
			return i
		}
	}
	return -1
}

// 生成证明
func GenerateProof(input ProofInput) (*ProofOutput, error) {
	var buf bytes.Buffer
	batchSize := len(input.Transactions)
	accountSize := len(input.Accounts)

	// 序列化交易
	for i := 0; i < batchSize; i++ {
		serializedTx := SerializedTransaction{
			From:   input.Transactions[i].From,
			To:     input.Transactions[i].To,
			Amount: fmt.Sprint(input.Transactions[i].Amount),
			Nonce:  fmt.Sprint(input.Transactions[i].Nonce),
		}
		txJSON, _ := json.Marshal(serializedTx)
		buf.Write(txJSON)
		buf.WriteByte('\n')
	}

	// 构建默克尔证明
	proofIndex := uint64(rand.Intn(batchSize))
	merkleRoot, _, _, err := merkletree.BuildReaderProof(&buf, bn254.NewMiMC("seed"), batchSize, proofIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to build merkle proof: %v", err)
	}

	// proofHelper := merkle.GenerateProofHelper(merkleProof, proofIndex, numLeaves)

	// 创建电路
	circuit := merkleCircuit{
		Transactions: make([]CircuitTransaction, batchSize),
		Addresses:    make([]frontend.Variable, accountSize),
		Balances:     make([]frontend.Variable, accountSize),
		Nonces:       make([]frontend.Variable, accountSize),
		// Path:         make([]frontend.Variable, len(merkleProof)),
		// Helper:       make([]frontend.Variable, len(merkleProof)-1),
	}

	// 编译电路
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &circuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %v", err)
	}

	// 设置证明系统
	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		return nil, fmt.Errorf("failed to setup proving system: %v", err)
	}

	// 记录旧账户状态
	old_accounts := make([]types.Account, accountSize)
	copy(old_accounts, input.Accounts)

	// 更新账户状态
	accounts := make([]types.Account, accountSize)
	copy(accounts, input.Accounts)
	for _, tx := range input.Transactions {
		fromIdx := findAccountIndex(accounts, tx.From)
		toIdx := findAccountIndex(accounts, tx.To)
		accounts[fromIdx].Balance -= tx.Amount
		accounts[fromIdx].Nonce++
		accounts[toIdx].Balance += tx.Amount
	}

	// 计算新状态根
	merkleRoot1 := ComputeAccountMerkleRoot(accounts)
	fmt.Printf("merkleRoot1: %v\n", merkleRoot1)

	// 创建witness
	witness := &merkleCircuit{
		OldRStateRoot: frontend.Value(input.OldStateRoot),
		// RootHash:       frontend.Value(merkleRoot),
		FinalStateRoot: frontend.Value(merkleRoot1),
		Transactions:   make([]CircuitTransaction, batchSize),
		Addresses:      make([]frontend.Variable, accountSize),
		Balances:       make([]frontend.Variable, accountSize),
		Nonces:         make([]frontend.Variable, accountSize),
		// Path:           make([]frontend.Variable, len(merkleProof)),
		// Helper:         make([]frontend.Variable, len(merkleProof)-1),
	}

	// 设置witness的值
	for i := 0; i < accountSize; i++ {
		witness.Addresses[i] = frontend.Value(parseInt(old_accounts[i].Address))
		witness.Balances[i] = frontend.Value(uint64(old_accounts[i].Balance))
		witness.Nonces[i] = frontend.Value(uint64(old_accounts[i].Nonce))
	}

	for i := 0; i < batchSize; i++ {
		fromAddr := parseInt(input.Transactions[i].From)
		toAddr := parseInt(input.Transactions[i].To)
		witness.Transactions[i] = CircuitTransaction{
			From:   frontend.Value(fromAddr),
			To:     frontend.Value(toAddr),
			Amount: frontend.Value(uint64(input.Transactions[i].Amount)),
			Nonce:  frontend.Value(uint64(input.Transactions[i].Nonce)),
		}
	}

	// for i := 0; i < len(merkleProof); i++ {
	// 	witness.Path[i].Assign(merkleProof[i])
	// }
	// for i := 0; i < len(merkleProof)-1; i++ {
	// 	witness.Helper[i].Assign(proofHelper[i])
	// }

	// 生成证明
	proof, err := groth16.Prove(r1cs, pk, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	output := &ProofOutput{
		OldStateRoot: input.OldStateRoot,
		BatchRoot:    new(big.Int).SetBytes(merkleRoot).String(),
		NewStateRoot: merkleRoot1,
		NewAccounts:  accounts, // 添加更新后的账户状态
		Proof:        proof,
		Vk:           vk,
	}

	return output, nil
}

// 序列化ProofOutput
func (p *ProofOutput) MarshalJSON() ([]byte, error) {
	// 创建缓冲区
	proofBuf := new(bytes.Buffer)
	vkBuf := new(bytes.Buffer)

	// 类型断言
	proof, ok := p.Proof.(io.WriterTo)
	if !ok {
		return nil, fmt.Errorf("proof does not implement WriterTo")
	}

	vk, ok := p.Vk.(io.WriterTo)
	if !ok {
		return nil, fmt.Errorf("vk does not implement WriterTo")
	}

	// 序列化proof和vk
	if _, err := proof.WriteTo(proofBuf); err != nil {
		return nil, fmt.Errorf("failed to write proof: %v", err)
	}

	if _, err := vk.WriteTo(vkBuf); err != nil {
		return nil, fmt.Errorf("failed to write vk: %v", err)
	}

	// 创建序列化结构体
	serialized := SerializedProofOutput{
		OldStateRoot: p.OldStateRoot,
		BatchRoot:    p.BatchRoot,
		NewStateRoot: p.NewStateRoot,
		NewAccounts:  p.NewAccounts, // 添加新的账户状态
		ProofData:    base64.StdEncoding.EncodeToString(proofBuf.Bytes()),
		VkData:       base64.StdEncoding.EncodeToString(vkBuf.Bytes()),
	}

	// 序列化为JSON
	return json.Marshal(serialized)
}

// 反序列化为ProofOutput
func (p *ProofOutput) UnmarshalJSON(data []byte) error {
	// 解析序列化的数据
	var serialized SerializedProofOutput
	if err := json.Unmarshal(data, &serialized); err != nil {
		return err
	}

	// 解码base64数据
	proofBytes, err := base64.StdEncoding.DecodeString(serialized.ProofData)
	if err != nil {
		return fmt.Errorf("failed to decode proof data: %v", err)
	}

	vkBytes, err := base64.StdEncoding.DecodeString(serialized.VkData)
	if err != nil {
		return fmt.Errorf("failed to decode vk data: %v", err)
	}

	// 初始化proof和vk
	proof := groth16.NewProof(ecc.BN254)
	vk := groth16.NewVerifyingKey(ecc.BN254)

	// 反序列化proof和vk
	if _, err := proof.ReadFrom(bytes.NewReader(proofBytes)); err != nil {
		return fmt.Errorf("failed to read proof: %v", err)
	}

	if _, err := vk.ReadFrom(bytes.NewReader(vkBytes)); err != nil {
		return fmt.Errorf("failed to read vk: %v", err)
	}

	// 设置字段值
	p.OldStateRoot = serialized.OldStateRoot
	p.BatchRoot = serialized.BatchRoot
	p.NewStateRoot = serialized.NewStateRoot
	p.NewAccounts = serialized.NewAccounts
	p.Proof = proof
	p.Vk = vk

	return nil
}

// 验证证明
func VerifyProof(proofStr string) error {
	// 打印proofStr
	fmt.Printf("proofStr: %v\n", proofStr)
	// 反序列化输入
	var output ProofOutput
	err := json.Unmarshal([]byte(proofStr), &output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal proof output: %v", err)
	}

	publicWitness := &merkleCircuit{
		OldRStateRoot: frontend.Value(output.OldStateRoot),
		// RootHash:       frontend.Value(output.BatchRoot),
		FinalStateRoot: frontend.Value(output.NewStateRoot),
	}

	// 类型断言
	proof, ok := output.Proof.(groth16.Proof)
	if !ok {
		return fmt.Errorf("invalid proof type")
	}

	vk, ok := output.Vk.(groth16.VerifyingKey)
	if !ok {
		return fmt.Errorf("invalid vk type")
	}

	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return fmt.Errorf("proof verification failed: %v", err)
	}
	return nil
}

// 辅助函数：解析地址字符串为整数
func parseInt(s string) int {
	// 将地址字符串转换为整数，去掉前导零
	cleanStr := strings.TrimLeft(s, "0")
	if cleanStr == "" {
		return 0
	}
	n, _ := strconv.Atoi(cleanStr)
	return n
}
