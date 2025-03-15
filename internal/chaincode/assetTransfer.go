package chaincode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"zkrollup/internal/zk"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var ccpPath = filepath.Join(
	"/",
	"home",
	"zkr",
	"hyperledger-fabric",
	"fabric-samples",
	"test-network",
	"organizations",
	"peerOrganizations",
	"org1.example.com",
	"connection-org1.yaml",
)

var credPath = filepath.Join(
	"/",
	"home",
	"zkr",
	"hyperledger-fabric",
	"fabric-samples",
	"test-network",
	"organizations",
	"peerOrganizations",
	"org1.example.com",
	"users",
	"User1@org1.example.com",
	"msp",
)

const (
	myChannel     = "mychannel"
	smartContract = "basic"
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
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(myChannel)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract(smartContract)
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
	log.Println(string(result))
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
