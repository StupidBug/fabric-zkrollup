package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/zk"
)

var id = 0

func JsonVerify(output *zk.ProofOutput) {
	outputBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Printf("Failed to marshal proof output: %v\n", err)
		return
	}
	id++
	VerifyMerkleRPC(strconv.Itoa(id), string(outputBytes))
}

func VerifyMerkleRPC(id string, output string) {
	contract := connectToNetwork()
	log.Println("--> Submit Transaction: VerifySaveProof")
	result, err := contract.SubmitTransaction("VerifySaveProof", id, output)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Printf("Verification result: %s\n", string(result))
	log.Println("Proof verification succeeded!")
	log.Println("--> Evaluate Transaction: GetAllProof, function returns all the current proofs")

	result, err = contract.EvaluateTransaction("GetAllProof")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}

	// 格式化JSON输出
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, result, "", "    "); err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}
	log.Println(prettyJSON.String())
}

type TokenBalance struct {
	ClientID string `json:"clientID"`
	TokenID  string `json:"tokenID"`
	Balance  int    `json:"balance"`
}

func GetAllTokenBalances() (accounts []types.Account) {
	contract := connectToNetwork()
	log.Println("--> Submit Transaction: GetAllTokenBalances")
	result, err := contract.EvaluateTransaction("GetAllTokenBalances")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	var TokenBalances []TokenBalance
	_ = json.Unmarshal(result, &TokenBalances)
	for _, tb := range TokenBalances {
		accounts = append(accounts, types.Account{
			Address: tb.ClientID,
			Balance: tb.Balance,
			Nonce:   0,
		})
	}
	log.Print(accounts)
	return
}
