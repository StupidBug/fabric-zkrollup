package main

import (
	"log"
	"zkrollup/internal/api/router"
	"zkrollup/internal/core/blockchain"
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
