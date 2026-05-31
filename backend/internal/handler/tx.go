package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/etherscan"
)

var txHashPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)

// TxHandler serves transaction lookups via Etherscan.
type TxHandler struct {
	cfg    config.Config
	ethers *etherscan.Client
}

// NewTxHandler constructs a TxHandler.
func NewTxHandler(cfg config.Config) *TxHandler {
	var client *etherscan.Client
	if cfg.EtherscanEnabled() {
		client = etherscan.NewClient(cfg.EtherscanAPIKey)
	}
	return &TxHandler{cfg: cfg, ethers: client}
}

// Get handles GET /api/v1/tx/:hash
func (h *TxHandler) Get(c *gin.Context) {
	hash := strings.TrimSpace(c.Param("hash"))
	if !txHashPattern.MatchString(hash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction hash"})
		return
	}

	if h.ethers == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "ETHERSCAN_API_KEY not configured",
		})
		return
	}

	receipt, err := h.ethers.GetTransactionReceipt(c.Request.Context(), hash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "etherscan request failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": receipt})
}
