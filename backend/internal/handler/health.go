package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/cache"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
)

// HealthDeps groups dependencies for the health check endpoint.
type HealthDeps struct {
	Mongo      *db.Mongo
	MySQL      *db.MySQL
	Redis      *cache.Client
	Config     config.Config
}

// Health responds with service liveness for load balancers and Render health checks.
func Health(deps HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload := gin.H{
			"status":      "ok",
			"service":     "web3infoanddex-api",
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"mongodb":     "ok",
			"mysql":       "ok",
			"redis":       "ok",
			"storageRead": deps.Config.StorageRead,
			"phase":       "4b",
		}

		degraded := false

		if err := deps.Mongo.Ping(c.Request.Context()); err != nil {
			payload["mongodb"] = "error"
			degraded = true
		}
		if err := deps.MySQL.Ping(c.Request.Context()); err != nil {
			payload["mysql"] = "error"
			degraded = true
		}
		if err := deps.Redis.Ping(c.Request.Context()); err != nil {
			payload["redis"] = "error"
			degraded = true
		}

		if deps.Config.EtherscanEnabled() {
			payload["etherscan"] = "configured"
		} else {
			payload["etherscan"] = "not_configured"
		}
		if deps.Config.OpenSeaAPIKey != "" {
			payload["opensea"] = "key_set"
		} else {
			payload["opensea"] = "placeholder"
		}

		if degraded {
			payload["status"] = "degraded"
			c.JSON(http.StatusServiceUnavailable, payload)
			return
		}

		c.JSON(http.StatusOK, payload)
	}
}
