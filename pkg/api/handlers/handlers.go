package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/api/types"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/blockchain"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/transaction"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	blockchain *blockchain.Blockchain
}

func NewHandler(bc *blockchain.Blockchain) *Handler {
	return &Handler{blockchain: bc}
}

// SendTransaction handles transaction submission
func (h *Handler) SendTransaction(c *gin.Context) {
	var req types.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse nonce
	nonce, err := strconv.ParseUint(req.Nonce, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nonce"})
		return
	}

	// Parse signature
	r := new(big.Int)
	s := new(big.Int)
	if _, ok := r.SetString(req.Signature.R, 16); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature R value"})
		return
	}
	if _, ok := s.SetString(req.Signature.S, 16); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature S value"})
		return
	}

	// Parse public key
	x := new(big.Int)
	y := new(big.Int)
	if _, ok := x.SetString(req.PublicKey.X, 16); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public key X value"})
		return
	}
	if _, ok := y.SetString(req.PublicKey.Y, 16); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public key Y value"})
		return
	}

	// Create public key
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// Store public key in state
	h.blockchain.SetPublicKey(req.From, pubKey)

	// Create transaction
	tx := transaction.Transaction{
		From:      req.From,
		To:        req.To,
		Value:     req.Value,
		Nonce:     nonce,
		Status:    transaction.StatusPending,
		Timestamp: time.Now().Unix(),
		Signature: transaction.Signature{
			R: r,
			S: s,
		},
	}

	// Compute hash
	tx.Hash = tx.ComputeHash()

	// Add to blockchain
	if err := h.blockchain.AddTransaction(tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create response
	resp := types.TransactionResponse{
		Hash:      hex.EncodeToString(tx.Hash[:]),
		From:      tx.From,
		To:        tx.To,
		Value:     strconv.Itoa(tx.Value),
		Nonce:     tx.Nonce,
		Status:    tx.Status.String(),
		Timestamp: tx.Timestamp,
		Signature: types.Signature{
			R: req.Signature.R,
			S: req.Signature.S,
		},
		PublicKey: types.PublicKey{
			X: req.PublicKey.X,
			Y: req.PublicKey.Y,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransaction handles transaction retrieval
func (h *Handler) GetTransaction(c *gin.Context) {
	hashHex := c.Query("hash")
	if hashHex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing hash parameter"})
		return
	}

	hashBytes, err := hex.DecodeString(hashHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash format"})
		return
	}

	var hash [32]byte
	copy(hash[:], hashBytes)

	tx := h.blockchain.GetTransactionByHash(hash)
	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	resp := types.TransactionResponse{
		Hash:      hex.EncodeToString(tx.Hash[:]),
		From:      tx.From,
		To:        tx.To,
		Value:     strconv.Itoa(tx.Value),
		Nonce:     tx.Nonce,
		Status:    tx.Status.String(),
		Timestamp: tx.Timestamp,
	}

	c.JSON(http.StatusOK, resp)
}

// GetBalance handles balance retrieval
func (h *Handler) GetBalance(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address parameter"})
		return
	}

	balance := h.blockchain.GetBalance(address)

	resp := types.BalanceResponse{
		Address: address,
		Balance: strconv.Itoa(balance),
	}

	c.JSON(http.StatusOK, resp)
}

// GetNonce handles nonce retrieval
func (h *Handler) GetNonce(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address parameter"})
		return
	}

	nonce := h.blockchain.GetNonce(address)
	resp := types.NonceResponse{
		Address: address,
		Nonce:   int(nonce),
	}
	c.JSON(http.StatusOK, resp)
}

// GetStateRoot handles state root retrieval
func (h *Handler) GetStateRoot(c *gin.Context) {
	stateRoot := h.blockchain.GetStateRoot()
	resp := types.StateRootResponse{
		StateRoot: stateRoot,
	}
	c.JSON(http.StatusOK, resp)
}

// CreateBlock handles block creation requests
func (h *Handler) CreateBlock(c *gin.Context) {
	if err := h.blockchain.CreateBlock(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{
		Status: "success",
	})
}

// GetBlockchainInfo returns information about the blockchain
func (h *Handler) GetBlockchainInfo(c *gin.Context) {
	latestBlock := h.blockchain.GetLatestBlock()
	if latestBlock == nil {
		c.JSON(http.StatusOK, gin.H{
			"height":          0,
			"latestBlockHash": "",
			"timestamp":       time.Now().Unix(),
		})
		return
	}

	hash := latestBlock.ComputeHash()
	c.JSON(http.StatusOK, gin.H{
		"height":          latestBlock.Header.Height,
		"latestBlockHash": hex.EncodeToString(hash[:]),
		"timestamp":       latestBlock.Header.Timestamp.Unix(),
	})
}

// GetTransactionPool returns all transactions in the pool
func (h *Handler) GetTransactionPool(c *gin.Context) {
	txs := h.blockchain.GetTransactionPool()
	c.JSON(http.StatusOK, gin.H{
		"count":        len(txs),
		"transactions": txs,
	})
}

// GetAllBlocks handles retrieving all blocks with their transactions
func (h *Handler) GetAllBlocks(c *gin.Context) {
	blocks := h.blockchain.GetAllBlocks()

	var response []types.BlockResponse
	for _, block := range blocks {
		blockHash := block.ComputeHash()
		prevHash := block.Header.PrevHash

		var transactions []types.TransactionResponse
		for _, tx := range block.Transactions {
			transactions = append(transactions, types.TransactionResponse{
				Hash:      hex.EncodeToString(tx.Hash[:]),
				From:      tx.From,
				To:        tx.To,
				Value:     strconv.Itoa(tx.Value),
				Nonce:     tx.Nonce,
				Status:    tx.Status.String(),
				Timestamp: tx.Timestamp,
			})
		}

		response = append(response, types.BlockResponse{
			Height:           block.Header.Height,
			Hash:             hex.EncodeToString(blockHash[:]),
			PrevHash:         hex.EncodeToString(prevHash[:]),
			MerkleRoot:       hex.EncodeToString(block.Header.MerkleRoot[:]),
			StateRoot:        block.Header.StateRoot,
			Timestamp:        block.Header.Timestamp.Unix(),
			TransactionCount: block.Header.TransactionCount,
			Transactions:     transactions,
		})
	}

	apiResponse := types.BlocksResponse{
		Status: "success",
		Data: types.BlocksData{
			Blocks: response,
		},
	}

	c.JSON(http.StatusOK, apiResponse)
}
