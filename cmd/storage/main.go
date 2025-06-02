package main

import (
	"fmt"
	"log"

	"github.com/StupidBug/fabric-zkrollup/conf"
	"github.com/StupidBug/fabric-zkrollup/pkg/api/router"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/evidencepool"
)

func main() {
	evidencePool := evidencepool.NewEvidencePool()
	evidencePool.StartAutoPackage()

	r := router.NewStorageRouter(evidencePool)
	r.Setup()

	// Start HTTP server
	if err := r.Run(fmt.Sprintf(":%s", conf.Port)); err != nil {
		log.Fatal(err)
	}
}
