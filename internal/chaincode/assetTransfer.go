package chaincode

import (
	"bytes"
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

func a(output string) {
	log.Println("============ application-golang starts ============")

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

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("basic")

	log.Println("--> Submit Transaction: InitProofOutput, function creates the initial set of assets on the ledger")
	result, err := contract.SubmitTransaction("InitProofOutput")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))
	log.Println("--> Evaluate Transaction: GetAllProof, function returns all the current assets on the ledger")
	result, err = contract.EvaluateTransaction("GetAllProof")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))
	log.Println("--> Verify Transaction: VerifyProof")
	result, err = contract.EvaluateTransaction("VerifyProof", output)
	if err != nil {
		fmt.Printf("Failed to verify proof: %v\n", err)
		return
	}
	log.Printf("Verification result: %s\n", string(result))

	fmt.Println("Proof verification succeeded!")
	log.Println("--> Evaluate Transaction: ProofExists, function returns 'true' if an asset with given assetID exist")
	result, err = contract.EvaluateTransaction("ProofExists", "output1")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v\n", err)
	}
	log.Println(string(result))
	log.Println("--> Submit Transaction: VerifySaveProof, creates new asset with ID, color, owner, size, and appraisedValue arguments")
	result, err = contract.SubmitTransaction("VerifySaveProof", "output4", output)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))
	log.Println("--> Evaluate Transaction: GetAllProof, function returns all the current assets on the ledger")
	result, err = contract.EvaluateTransaction("GetAllProof")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	// log.Println(string(result))
	// 格式化JSON输出
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, result, "", "    "); err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}
	log.Println(prettyJSON.String())

}
func main() {
	outputBytes := []byte(`{"old_state_root":"12486946051716700098682063972940734609165340983839085472908181019624970850750","batch_root":"12464520858580237731381121317043389447999537636646894863801358742638555620238","new_state_root":"12486946051716700098682063972940734609165340983839085472908181019624970850750","proof":"0jkcGOrzZQC+igrzdSD/QoLgBS/icb+rR4MQFL4MUQ7U4NQSglWKIqSIB119ieQDZ54cbtvnWqt1rmeuDsxGjxk/b9NYx9TjBt9EC+25uLxZQSaISLva3zJzt3Gco06o1zYJD9/X5F0PyFTOvPgK1A0P+xzte8hmpqgECjwexCE=","vk":"gt5g9l+687t1piE75FMEQ0yNZBbjDLSsLdT6w/cS8zOpHFfLCi7H7WPeJkp/1HdMZ/T5L9hxL26jzFQZOpoZntdaXrNRwQx1PR88I/8ji63yLi5v99Gh83L68lN/RkDQCgltJVSHsPNP82M14xFd4zW8YHJh+g3gxWHSqSkvnk7bIjyjWdKQc9jYrnNnXWvk7L5P3TpLQOYeFjNqs2ItTRJ4MPHoigOrP1JADzO9avZOC9n5ybvUYcjZUUk0allTqWVGTmZUmjEEXxXD3n1HTubhd5bdXjIph4GxPEyheBmRkhgbHZSckNCHKaeVkktP0Hco6f7cen+vXDzROrLfMgi8DbYwG56LuYZ94taxoK4Mlo4yFXCsfxwdlnyL01vFAAAABIiWhCKbq6IpRSvMVvJTj5cAxL5OLxSiN5IvLV6XbGlpwnxLzXXtNgCBjE4aYfBktA/ekFRiENzQEyAeOkQES/2ZgXTa6fYnAcBmiW35Lnym529dRT11Mwwx9UX9XlwZ9Ir8BXxaWIVybTkYhacwJz91/Nfxe0CYfpO000jq/f6l"}`)
	id := "output2"

	VerifyMerkleRPC(id, string(outputBytes))
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
