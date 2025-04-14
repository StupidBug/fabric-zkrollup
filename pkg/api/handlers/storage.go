package handlers

import (
	"log"
	"net/http"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/evidencepool"
	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	pp *evidencepool.EvidencePool
}

func NewStorageHandler(evidencePool *evidencepool.EvidencePool) *StorageHandler {
	return &StorageHandler{
		pp: evidencePool,
	}
}

func (h *StorageHandler) Storage(c *gin.Context) {
	var req types.ProofStorageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if hashs, err := h.pp.Add(req.Evidence); err == nil {
		log.Printf("proof append success, hash: %v, proof %v", hashs, req.Evidence)
		c.JSON(http.StatusOK, types.ProofStorageResp{
			Hashs: hashs,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
	}
}

func (h *StorageHandler) Status(c *gin.Context) {
	hash := c.Query("hash")
	status := h.pp.Query(hash)
	c.JSON(http.StatusOK, types.ProofStatusResp{
		Status: status,
	})
}
