package main

import (
	"fmt"
	"log"

	"github.com/StupidBug/fabric-zkrollup/conf"
	"github.com/StupidBug/fabric-zkrollup/pkg/api/router"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/blockchain"
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
	if err := r.Run(fmt.Sprintf(":%s", conf.Port)); err != nil {
		log.Fatal(err)
	}
}
