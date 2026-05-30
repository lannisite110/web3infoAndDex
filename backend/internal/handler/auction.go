package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
)

// AuctionHandler serves auction list API.
type AuctionHandler struct {
	repo      *repository.AuctionRepository
	chainID   int64
	contract  string
}

// NewAuctionHandler constructs an AuctionHandler.
func NewAuctionHandler(repo *repository.AuctionRepository, chainID int64, contract string) *AuctionHandler {
	return &AuctionHandler{repo: repo, chainID: chainID, contract: contract}
}

// List handles GET /api/v1/auctions
func (h *AuctionHandler) List(c *gin.Context) {
	auctions, err := h.repo.List(c.Request.Context(), h.chainID, h.contract)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list auctions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auctions": auctions})
}
