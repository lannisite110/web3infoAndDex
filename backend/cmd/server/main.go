package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lannisite110/web3infoanddex/backend/internal/cache"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/eth"
	"github.com/lannisite110/web3infoanddex/backend/internal/handler"
	"github.com/lannisite110/web3infoanddex/backend/internal/indexer"
	"github.com/lannisite110/web3infoanddex/backend/internal/migrate"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
	mysqlrepo "github.com/lannisite110/web3infoanddex/backend/internal/repository/mysql"
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
	defer closeMongo(mongo)

	mysqlDSN, err := config.NormalizeMySQLDSN(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("mysql dsn: %v", err)
	}
	slog.Info("connecting to mysql")
	mysql, err := connectMySQLWithRetry(mysqlDSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer closeMySQL(mysql)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := migrate.Apply(ctx, mysql); err != nil {
		log.Fatalf("mysql migrate: %v", err)
	}
	slog.Info("mysql migrations ok")

	slog.Info("connecting to redis")
	redisClient, err := connectRedisWithRetry(cfg.RedisURL, cfg.CacheTTL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("redis close", "err", err)
		}
	}()

	mongoAuctionRepo, err := repository.NewAuctionRepository(mongo)
	if err != nil {
		log.Fatalf("mongo auction repository: %v", err)
	}
	mongoBidRepo, err := repository.NewBidRepository(mongo)
	if err != nil {
		log.Fatalf("mongo bid repository: %v", err)
	}

	mysqlAuctionRepo := mysqlrepo.NewAuctionRepository(mysql)
	mysqlBidRepo := mysqlrepo.NewBidRepository(mysql)
	mysqlStateRepo := mysqlrepo.NewIndexerStateRepository(mysql)

	mongoStateRepo := repository.NewIndexerStateRepository(mongo)
	dualAuctions := repository.NewDualAuctionStore(mongoAuctionRepo, mysqlAuctionRepo)
	dualBids := repository.NewDualBidStore(mongoBidRepo, mysqlBidRepo)
	dualState := repository.NewDualIndexerStateStore(
		repository.NewMongoIndexerStateStore(mongoStateRepo),
		repository.NewMySQLIndexerStateStore(mysqlStateRepo),
	)

	auctionHandler := handler.NewAuctionHandler(
		mysqlAuctionRepo,
		mysqlBidRepo,
		redisClient,
		cfg.ChainID,
		cfg.AuctionContract,
	)
	txHandler := handler.NewTxHandler(cfg)
	openSeaHandler := handler.NewOpenSeaHandler(cfg)

	var ethClient *eth.Client
	var cancelIndexer context.CancelFunc

	if cfg.IndexerEnabled() {
		ethClient, err = eth.NewClient(cfg.SepoliaRPCURL)
		if err != nil {
			log.Fatalf("ethereum client: %v", err)
		}

		idx, err := indexer.New(cfg, ethClient, dualAuctions, dualBids, dualState, redisClient)
		if err != nil {
			log.Fatalf("indexer: %v", err)
		}

		var indexerCtx context.Context
		indexerCtx, cancelIndexer = context.WithCancel(context.Background())
		go idx.Run(indexerCtx)
		slog.Info("indexer enabled", "rpc", maskRPC(cfg.SepoliaRPCURL))
	} else {
		slog.Warn("indexer disabled: set SEPOLIA_RPC_URL and NFT_AUCTION_ADDRESS")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg.CORSOrigins))

	r.GET("/health", handler.Health(handler.HealthDeps{
		Mongo:  mongo,
		MySQL:  mysql,
		Redis:  redisClient,
		Config: cfg,
	}))
	api := r.Group("/api/v1")
	{
		api.GET("/auctions", auctionHandler.List)
		api.GET("/auctions/:id", auctionHandler.Get)
		api.GET("/auctions/:id/bids", auctionHandler.ListBids)
		api.GET("/bids", auctionHandler.ListAllBids)
		api.GET("/tx/:hash", txHandler.Get)
		api.GET("/nft/metadata", openSeaHandler.NFTMetadata)
	}

	addr := ":" + cfg.Port
	slog.Info("server listening",
		"addr", addr,
		"mongodb", cfg.MongoDBName,
		"storageRead", cfg.StorageRead,
	)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	waitForShutdown(cancelIndexer, ethClient)
}

func connectMySQLWithRetry(dsn string) (*db.MySQL, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		m, err := db.ConnectMySQL(ctx, dsn)
		cancel()
		if err == nil {
			return m, nil
		}
		lastErr = err
		slog.Warn("mysql connect failed, retrying", "attempt", attempt, "err", err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
		}
	}
	return nil, lastErr
}

func connectRedisWithRetry(redisURL string, ttl time.Duration) (*cache.Client, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		c, err := cache.Connect(ctx, redisURL, ttl)
		cancel()
		if err == nil {
			return c, nil
		}
		lastErr = err
		slog.Warn("redis connect failed, retrying", "attempt", attempt, "err", err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}
	return nil, lastErr
}

func closeMongo(mongo *db.Mongo) {
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := mongo.Close(shutdown); err != nil {
		slog.Error("mongodb close", "err", err)
	}
}

func closeMySQL(mysql *db.MySQL) {
	if err := mysql.Close(); err != nil {
		slog.Error("mysql close", "err", err)
	}
}

func maskRPC(url string) string {
	if i := strings.LastIndex(url, "/"); i >= 0 && len(url) > i+8 {
		return url[:i+1] + "***"
	}
	return "***"
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
		allowed[normalizeOrigin(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader("Origin"))
		if origin != "" && corsOriginAllowed(allowed, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func normalizeOrigin(o string) string {
	o = strings.TrimSpace(o)
	return strings.TrimSuffix(o, "/")
}

func corsOriginAllowed(allowed map[string]struct{}, origin string) bool {
	if _, ok := allowed[origin]; ok {
		return true
	}
	if !strings.HasSuffix(origin, ".vercel.app") {
		return false
	}
	for o := range allowed {
		if strings.HasSuffix(o, ".vercel.app") {
			return true
		}
	}
	return false
}
