package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/cache"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
)

type auctionListCache struct {
	Auctions []model.Auction `json:"auctions"`
}

type auctionGetCache struct {
	Auction model.Auction `json:"auction"`
}

type bidsCache struct {
	AuctionID uint64      `json:"auctionId"`
	Bids      []model.Bid `json:"bids"`
}

type bidsGlobalCache struct {
	Bids []model.Bid `json:"bids"`
}

// AuctionHandler serves auction and bid list APIs (MySQL + Redis).
type AuctionHandler struct {
	auctions repository.AuctionStore
	bids     repository.BidStore
	cache    *cache.Client
	chainID  int64
	contract string
}

// NewAuctionHandler constructs an AuctionHandler.
func NewAuctionHandler(
	auctions repository.AuctionStore,
	bids repository.BidStore,
	cacheClient *cache.Client,
	chainID int64,
	contract string,
) *AuctionHandler {
	return &AuctionHandler{
		auctions: auctions,
		bids:     bids,
		cache:    cacheClient,
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

	queryKey := cache.BuildListQueryKey(map[string]string{
		"q":       filter.Q,
		"seller":  filter.Seller,
		"tokenId": filter.TokenID,
		"ended":   c.Query("ended"),
		"bidder":  c.Query("bidder"),
	})

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

	cacheKey := cache.AuctionListKey(h.chainID, h.contract, queryKey)
	if h.cache != nil {
		var cached auctionListCache
		if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"auctions": cached.Auctions})
			return
		}
	}

	auctions, err := h.auctions.ListFiltered(c.Request.Context(), h.chainID, h.contract, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list auctions"})
		return
	}

	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, auctionListCache{Auctions: auctions})
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{"auctions": auctions})
}

// Get handles GET /api/v1/auctions/:id
func (h *AuctionHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	cacheKey := cache.AuctionKey(h.chainID, h.contract, id)
	if h.cache != nil {
		var cached auctionGetCache
		if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"auction": cached.Auction})
			return
		}
	}

	auction, err := h.auctions.Get(c.Request.Context(), h.chainID, h.contract, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "auction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get auction"})
		return
	}

	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, auctionGetCache{Auction: auction})
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{"auction": auction})
}

// ListBids handles GET /api/v1/auctions/:id/bids
func (h *AuctionHandler) ListBids(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	cacheKey := cache.BidsAuctionKey(h.chainID, h.contract, id)
	if h.cache != nil {
		var cached bidsCache
		if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"auctionId": id, "bids": cached.Bids})
			return
		}
	}

	bids, err := h.bids.ListByAuction(c.Request.Context(), h.chainID, h.contract, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bids"})
		return
	}

	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, bidsCache{AuctionID: id, Bids: bids})
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{"auctionId": id, "bids": bids})
}

// ListAllBids handles GET /api/v1/bids?bidder=&limit=
func (h *AuctionHandler) ListAllBids(c *gin.Context) {
	bidder := strings.ToLower(strings.TrimSpace(c.Query("bidder")))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	cacheKey := cache.BidsGlobalKey(h.chainID, h.contract, bidder, limit)
	if h.cache != nil {
		var cached bidsGlobalCache
		if ok, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{"bids": cached.Bids})
			return
		}
	}

	bids, err := h.bids.List(c.Request.Context(), h.chainID, h.contract, bidder, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bids"})
		return
	}

	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, bidsGlobalCache{Bids: bids})
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{"bids": bids})
}
