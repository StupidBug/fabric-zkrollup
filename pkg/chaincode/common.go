package chaincode

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/StupidBug/fabric-zkrollup/conf"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func connectToNetwork() *gateway.Contract {
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
		gateway.WithConfig(config.FromFile(filepath.Clean(conf.CcpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(conf.Channel)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}
	contract := network.GetContract(conf.Contract)
	return contract
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")

	certPath := filepath.Join(conf.CredPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(conf.CredPath, "keystore")
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
