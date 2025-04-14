package zk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"

	"encoding/base64"

	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
	"github.com/consensys/gnark/std/hash/mimc"
)

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

type storageCircuit struct {
	// 公开输入
	OldStateRoot   frontend.Variable `gnark:",public"` // 旧状态根
	RootHash       frontend.Variable `gnark:",public"` // 批次根
	FinalStateRoot frontend.Variable `gnark:",public"` // 最终状态根

	// 私有输入
	Evidence []frontend.Variable // 存证数据
	Path     []frontend.Variable // 默克尔路径
	Helper   []frontend.Variable // 默克尔路径辅助数据
}

// 定义电路
func (circuit *storageCircuit) Define(curveID ecc.ID, api frontend.API) error {
	hFunc, err := mimc.NewMiMC("seed", curveID, api)
	if err != nil {
		return err
	}

	// 验证批次根的默克尔路径，证明存证在批次中
	merkle.VerifyProof(api, hFunc, circuit.RootHash, circuit.Path, circuit.Helper)

	// 计算新的状态根
	stateHasher, _ := mimc.NewMiMC("seed", curveID, api)

	// 根据旧状态根和批次根的哈希值计算新的状态根
	stateHasher.Reset()
	stateHasher.Write(circuit.OldStateRoot)
	stateHasher.Write(circuit.RootHash)
	computedNewRoot := stateHasher.Sum()

	// 验证计算的新状态根是否与预期一致，证明批次存证的正确性
	api.AssertIsEqual(circuit.FinalStateRoot, computedNewRoot)

	for i := 0; i < len(circuit.Evidence); i++ {
		// 验证每一个存证数据不能为空
		api.AssertIsDifferent(circuit.Evidence[i], api.Constant(0))
	}

	return nil
}

// 计算存证状态根
func ComputeStorageMerkleRoot(evidence []string) string {
	// 创建一个切片来存储所有存证的哈希值
	hashes := make([]*big.Int, len(evidence))

	// 计算每个存证的哈希值
	for i, ev := range evidence {
		f := bn254.NewMiMC("seed")
		f.Write([]byte(ev))
		hashes[i] = new(big.Int).SetBytes(f.Sum(nil))
	}

	// 两两哈希直到只剩下一个值
	for len(hashes) > 1 {
		newHashes := make([]*big.Int, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				// 合并两个哈希值
				f := bn254.NewMiMC("seed")
				f.Write(hashes[i].Bytes())
				f.Write(hashes[i+1].Bytes())
				newHashes = append(newHashes, new(big.Int).SetBytes(f.Sum(nil)))
			} else {
				// 如果是奇数个哈希值，最后一个直接保留
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}

	return hashes[0].String()
}

// 序列化的输出结构体
type SerializedStorageProofOutput struct {
	OldStateRoot string   `json:"old_state_root"`
	BatchRoot    string   `json:"batch_root"`
	NewStateRoot string   `json:"new_state_root"`
	Evidence     []string `json:"evidence"`
	ProofData    string   `json:"proof"`
	VkData       string   `json:"vk"`
}

// 序列化StorageProofOutput
func (p *StorageProofOutput) MarshalJSON() ([]byte, error) {
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
	serialized := SerializedStorageProofOutput{
		OldStateRoot: p.OldStateRoot,
		BatchRoot:    p.BatchRoot,
		NewStateRoot: p.NewStateRoot,
		Evidence:     p.Evidence,
		ProofData:    base64.StdEncoding.EncodeToString(proofBuf.Bytes()),
		VkData:       base64.StdEncoding.EncodeToString(vkBuf.Bytes()),
	}

	// 序列化为JSON
	return json.Marshal(serialized)
}

// 反序列化为StorageProofOutput
func (p *StorageProofOutput) UnmarshalJSON(data []byte) error {
	// 解析序列化的数据
	var serialized SerializedStorageProofOutput
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
	p.Evidence = serialized.Evidence
	p.Proof = proof
	p.Vk = vk

	return nil
}

// 计算新的状态根
func computeNewStateRoot(oldRoot, batchRoot string) string {
	hasher := bn254.NewMiMC("seed")
	oldRootInt := new(big.Int)
	oldRootInt.SetString(oldRoot, 10)
	batchRootInt := new(big.Int)
	batchRootInt.SetString(batchRoot, 10)
	hasher.Write(oldRootInt.Bytes())
	hasher.Write(batchRootInt.Bytes())
	return new(big.Int).SetBytes(hasher.Sum(nil)).String()
}

// 打包存证
func Package(input StorageProofInput) (*StorageProofOutput, error) {
	// 构建Merkle树
	tree := merkletree.New(bn254.NewMiMC("seed"))

	// 添加所有存证到Merkle树
	for _, evidence := range input.Evidence {
		tree.Push([]byte(evidence))
	}

	// 构建Merkle证明
	merkleRoot, merkleProof, numLeaves, err := merkletree.BuildReaderProof(bytes.NewReader([]byte(strings.Join(input.Evidence, "\n"))), bn254.NewMiMC("seed"), 1, uint64(len(input.Evidence)-1))
	if err != nil {
		return nil, fmt.Errorf("failed to build merkle proof: %v", err)
	}

	proofHelper := merkle.GenerateProofHelper(merkleProof, uint64(len(input.Evidence)-1), numLeaves)

	// 创建电路
	circuit := storageCircuit{
		Evidence: make([]frontend.Variable, len(input.Evidence)),
		Path:     make([]frontend.Variable, len(merkleProof)),
		Helper:   make([]frontend.Variable, len(merkleProof)-1),
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

	// 将merkleRoot转换为字符串
	merkleRootStr := new(big.Int).SetBytes(merkleRoot).String()

	// 计算新的状态根
	newStateRoot := computeNewStateRoot(input.OldStateRoot, merkleRootStr)

	// 创建witness
	witness := &storageCircuit{
		OldStateRoot:   frontend.Value(input.OldStateRoot),
		RootHash:       frontend.Value(merkleRoot),
		FinalStateRoot: frontend.Value(newStateRoot),
		Evidence:       make([]frontend.Variable, len(input.Evidence)),
		Path:           make([]frontend.Variable, len(merkleProof)),
		Helper:         make([]frontend.Variable, len(merkleProof)-1),
	}

	// 设置witness的值
	for i, evidence := range input.Evidence {
		// 将字符串转换为big.Int
		evidenceBytes := []byte(evidence)
		evidenceInt := new(big.Int).SetBytes(evidenceBytes)
		witness.Evidence[i] = frontend.Value(evidenceInt)
	}
	for i, node := range merkleProof {
		witness.Path[i] = frontend.Value(node)
	}
	for i, helper := range proofHelper {
		witness.Helper[i] = frontend.Value(helper)
	}

	// 生成证明
	proof, err := groth16.Prove(r1cs, pk, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	// 创建输出
	output := &StorageProofOutput{
		OldStateRoot: input.OldStateRoot,
		BatchRoot:    merkleRootStr,
		NewStateRoot: newStateRoot,
		Evidence:     input.Evidence,
		Proof:        proof,
		Vk:           vk,
	}

	return output, nil
}

// 验证证明
func VerifyStorageProof(proofStr string) error {
	// 反序列化输入
	var output StorageProofOutput
	err := json.Unmarshal([]byte(proofStr), &output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal proof output: %v", err)
	}

	publicWitness := &merkleCircuit{
		OldStateRoot:   frontend.Value(output.OldStateRoot),
		RootHash:       frontend.Value(output.BatchRoot),
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
