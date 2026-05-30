package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
)

// Health responds with service liveness for load balancers and Render health checks.
func Health(mongo *db.Mongo) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload := gin.H{
			"status":    "ok",
			"service":   "web3infoanddex-api",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"mongodb":   "ok",
		}

		if err := mongo.Ping(c.Request.Context()); err != nil {
			payload["status"] = "degraded"
			payload["mongodb"] = "error"
			c.JSON(http.StatusServiceUnavailable, payload)
			return
		}

		c.JSON(http.StatusOK, payload)
	}
}
