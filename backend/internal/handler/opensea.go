package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
)

// OpenSeaHandler is a placeholder for future OpenSea metadata integration.
type OpenSeaHandler struct {
	cfg config.Config
}

// NewOpenSeaHandler constructs an OpenSeaHandler.
func NewOpenSeaHandler(cfg config.Config) *OpenSeaHandler {
	return &OpenSeaHandler{cfg: cfg}
}

// NFTMetadata handles GET /api/v1/nft/metadata?contract=&tokenId=
func (h *OpenSeaHandler) NFTMetadata(c *gin.Context) {
	contract := strings.TrimSpace(c.Query("contract"))
	tokenID := strings.TrimSpace(c.Query("tokenId"))
	if contract == "" || tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract and tokenId are required"})
		return
	}

	if h.cfg.OpenSeaAPIKey == "" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "OpenSea integration not configured",
			"message": "Set OPENSEA_API_KEY to enable NFT metadata (mainnet collections recommended; Sepolia has limited OpenSea support).",
			"contract": contract,
			"tokenId":  tokenID,
		})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "OpenSea fetch not yet implemented",
		"message": "API key is set; wire OpenSea v2 client in a follow-up deploy.",
		"baseUrl": h.cfg.OpenSeaBaseURL,
	})
}
