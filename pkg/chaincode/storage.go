package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/StupidBug/fabric-zkrollup/pkg/zk"
)

type StorageClient interface {
	func(*zk.StorageProofOutput) error
}

func Validate(output *zk.StorageProofOutput) {
	outputBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Printf("Failed to marshal proof output: %v\n", err)
		return
	}
	id++
	OrginValidate(strconv.Itoa(id), string(outputBytes))
}

func OrginValidate(id string, storageOutput string) {
	contract := connectToNetwork()

	log.Println("--> Submit Transaction: VerifySaveStorageProof")
	result, err := contract.SubmitTransaction("VerifySaveStorageProof", id, storageOutput)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Printf("Verification result: %s\n", string(result))
	log.Println("Proof verification succeeded!")
	log.Println("--> Evaluate Transaction: GetAllStorageProof, function returns all the current proofs")

	result, err = contract.EvaluateTransaction("GetAllStorageProof")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, result, "", "    "); err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}
	log.Println(prettyJSON.String())
}
