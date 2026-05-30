package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuctionHandler serves auction and bid list APIs.
type AuctionHandler struct {
	auctions *repository.AuctionRepository
	bids     *repository.BidRepository
	chainID  int64
	contract string
}

// NewAuctionHandler constructs an AuctionHandler.
func NewAuctionHandler(
	auctions *repository.AuctionRepository,
	bids *repository.BidRepository,
	chainID int64,
	contract string,
) *AuctionHandler {
	return &AuctionHandler{
		auctions: auctions,
		bids:     bids,
		chainID:  chainID,
		contract: contract,
	}
}

// List handles GET /api/v1/auctions?q=&seller=&tokenId=&ended=&bidder=
func (h *AuctionHandler) List(c *gin.Context) {
	filter := model.AuctionFilter{
		Q:       c.Query("q"),
		Seller:  c.Query("seller"),
		TokenID: c.Query("tokenId"),
	}

	if v := strings.TrimSpace(c.Query("bidder")); v != "" {
		ids, err := h.bids.AuctionIDsForBidder(
			c.Request.Context(),
			h.chainID,
			h.contract,
			strings.ToLower(v),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to filter by bidder"})
			return
		}
		if len(ids) == 0 {
			c.JSON(http.StatusOK, gin.H{"auctions": []model.Auction{}})
			return
		}
		filter.AuctionIDs = ids
	}

	if v := strings.TrimSpace(c.Query("ended")); v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			b := true
			filter.Ended = &b
		case "false", "0", "no":
			b := false
			filter.Ended = &b
		}
	}

	auctions, err := h.auctions.ListFiltered(c.Request.Context(), h.chainID, h.contract, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list auctions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auctions": auctions})
}

// Get handles GET /api/v1/auctions/:id
func (h *AuctionHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	auction, err := h.auctions.Get(c.Request.Context(), h.chainID, h.contract, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "auction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get auction"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auction": auction})
}

// ListBids handles GET /api/v1/auctions/:id/bids
func (h *AuctionHandler) ListBids(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	bids, err := h.bids.ListByAuction(c.Request.Context(), h.chainID, h.contract, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bids"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auctionId": id, "bids": bids})
}

// ListAllBids handles GET /api/v1/bids?bidder=&limit=
func (h *AuctionHandler) ListAllBids(c *gin.Context) {
	bidder := strings.ToLower(strings.TrimSpace(c.Query("bidder")))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	bids, err := h.bids.List(c.Request.Context(), h.chainID, h.contract, bidder, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bids"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bids": bids})
}
