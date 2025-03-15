package main

import (
	"github.com/StupidBug/fabric-zkrollup/pkg/api/router"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/blockchain"
	"log"
)

func main() {
	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	// Start automatic block creation
	bc.StartAutoBlock()

	// Create and setup router
	r := router.NewRouter(bc)
	r.Setup()

	// Start HTTP server
	log.Println("Server is running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
