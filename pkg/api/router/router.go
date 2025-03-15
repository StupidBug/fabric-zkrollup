package router

import (
	"github.com/StupidBug/fabric-zkrollup/pkg/api/handlers"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/blockchain"

	"github.com/gin-gonic/gin"
)

// Router represents the HTTP router
type Router struct {
	engine  *gin.Engine
	handler *handlers.Handler
}

// NewRouter creates a new router instance
func NewRouter(bc *blockchain.Blockchain) *Router {
	engine := gin.Default()
	return &Router{
		engine:  engine,
		handler: handlers.NewHandler(bc),
	}
}

// Setup sets up the HTTP routes
func (r *Router) Setup() {
	// API routes
	v1 := r.engine.Group("/api/v1")
	{
		// Transaction endpoints
		v1.POST("/transaction/send", r.handler.SendTransaction)
		v1.GET("/transaction/get", r.handler.GetTransaction)

		// Balance endpoints
		v1.GET("/balance/get", r.handler.GetBalance)

		// Account endpoints
		v1.GET("/account/nonce", r.handler.GetNonce)

		// State endpoints
		v1.GET("/state/root", r.handler.GetStateRoot)

		// Block endpoints
		v1.GET("/blocks", r.handler.GetAllBlocks)
	}
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
