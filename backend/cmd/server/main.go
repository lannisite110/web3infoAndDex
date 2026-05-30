package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/eth"
	"github.com/lannisite110/web3infoanddex/backend/internal/handler"
	"github.com/lannisite110/web3infoanddex/backend/internal/indexer"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("connecting to mongodb", "uri", config.MongoURIRedacted(cfg.MongoDBURI))

	mongo, err := connectMongoWithRetry(cfg.MongoDBURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongo.Close(shutdown); err != nil {
			slog.Error("mongodb close", "err", err)
		}
	}()

	auctionRepo, err := repository.NewAuctionRepository(mongo)
	if err != nil {
		log.Fatalf("auction repository: %v", err)
	}

	auctionHandler := handler.NewAuctionHandler(auctionRepo, cfg.ChainID, cfg.AuctionContract)

	stateRepo := repository.NewIndexerStateRepository(mongo)

	var ethClient *eth.Client
	var cancelIndexer context.CancelFunc

	if cfg.IndexerEnabled() {
		var err error
		ethClient, err = eth.NewClient(cfg.SepoliaRPCURL)
		if err != nil {
			log.Fatalf("ethereum client: %v", err)
		}

		idx, err := indexer.New(cfg, ethClient, auctionRepo, stateRepo)
		if err != nil {
			log.Fatalf("indexer: %v", err)
		}

		var indexerCtx context.Context
		indexerCtx, cancelIndexer = context.WithCancel(context.Background())
		go idx.Run(indexerCtx)
		slog.Info("indexer enabled", "rpc", cfg.SepoliaRPCURL)
	} else {
		slog.Warn("indexer disabled: set SEPOLIA_RPC_URL and NFT_AUCTION_ADDRESS")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg.CORSOrigins))

	r.GET("/health", handler.Health(mongo))
	r.GET("/api/v1/auctions", auctionHandler.List)

	addr := ":" + cfg.Port
	slog.Info("server listening", "addr", addr, "mongodb", cfg.MongoDBName)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	waitForShutdown(cancelIndexer, ethClient)
}

func connectMongoWithRetry(uri, dbName string) (*db.Mongo, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		m, err := db.Connect(ctx, uri, dbName)
		cancel()
		if err == nil {
			return m, nil
		}
		lastErr = err
		slog.Warn("mongodb connect failed, retrying", "attempt", attempt, "err", err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
		}
	}
	return nil, lastErr
}

func waitForShutdown(cancelIndexer context.CancelFunc, ethClient *eth.Client) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutdown signal received")
	if cancelIndexer != nil {
		cancelIndexer()
	}
	if ethClient != nil {
		ethClient.Close()
	}
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
