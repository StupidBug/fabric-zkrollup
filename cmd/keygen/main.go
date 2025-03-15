package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/StupidBug/fabric-zkrollup/pkg/core/crypto"
)

func main() {
	// Define flags for key generation
	genKeyCmd := flag.Bool("genkey", false, "Generate a new key pair")

	// Define flags for signing
	signCmd := flag.Bool("sign", false, "Sign a transaction")
	fromAddr := flag.String("from", "", "From address")
	toAddr := flag.String("to", "", "To address")
	value := flag.Int("value", 0, "Transfer value")
	nonce := flag.Int("nonce", 0, "Transaction nonce")
	privKey := flag.String("privkey", "", "Private key for signing")

	flag.Parse()

	if *genKeyCmd {
		// Generate new key pair
		priv, pub := crypto.GenerateKeyPair()
		fmt.Printf("Private key: %x\n", priv)
		fmt.Printf("Public key X: %x\n", pub.X)
		fmt.Printf("Public key Y: %x\n", pub.Y)
		return
	}

	if *signCmd {
		if *fromAddr == "" || *toAddr == "" || *privKey == "" {
			log.Fatal("Missing required parameters for signing")
		}

		// Parse private key
		privKeyBig := new(big.Int)
		privKeyBig.SetString(*privKey, 16)

		// Create transaction data
		txData := fmt.Sprintf("%s%s%d%d", *fromAddr, *toAddr, *value, *nonce)

		// Sign transaction
		sig := crypto.Sign(txData, privKeyBig)
		pub := crypto.PrivateKeyToPublic(privKeyBig)

		fmt.Printf("Transaction hash: %s\n", txData)
		fmt.Printf("Signature R: %x\n", sig.R)
		fmt.Printf("Signature S: %x\n", sig.S)
		fmt.Printf("Public key X: %x\n", pub.X)
		fmt.Printf("Public key Y: %x\n", pub.Y)
		return
	}

	fmt.Println("Please specify either -genkey or -sign")
}
