package main

import (
	"log"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/router"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/evidencepool"
)

func main() {
	evidencePool := evidencepool.NewEvidencePool()
	evidencePool.StartAutoPackage()

	r := router.NewStorageRouter(evidencePool)
	r.Setup()

	// Start HTTP server
	log.Println("Server is running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
