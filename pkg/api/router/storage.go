package router

import (
	"github.com/StupidBug/fabric-zkrollup/pkg/api/handlers"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/evidencepool"

	"github.com/gin-gonic/gin"
)

// Router represents the HTTP router
type StorageRouter struct {
	engine  *gin.Engine
	handler *handlers.StorageHandler
}

// NewRouter creates a new router instance
func NewStorageRouter(evidencePool *evidencepool.EvidencePool) *StorageRouter {
	engine := gin.Default()
	return &StorageRouter{
		engine:  engine,
		handler: handlers.NewStorageHandler(evidencePool),
	}
}

// Setup sets up the HTTP routes
func (r *StorageRouter) Setup() {
	// API routes
	v1 := r.engine.Group("/api/v1")
	{
		// 存证
		v1.POST("/proof/storage", r.handler.Storage)
		v1.GET("/proof/status", r.handler.Status)
	}
}

// Run starts the HTTP server
func (r *StorageRouter) Run(addr string) error {
	return r.engine.Run(addr)
}
