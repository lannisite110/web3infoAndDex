package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultMongoDBName   = "web3dex"
	DefaultChainID       = 11155111
	DefaultSyncInterval  = 15 * time.Second
)

// Config holds runtime settings loaded from environment variables.
type Config struct {
	Port            string
	CORSOrigins     []string
	MongoDBURI      string
	MongoDBName     string
	ChainID         int64
	AuctionContract string
	SepoliaRPCURL   string
	DeployBlock     uint64
	SyncInterval    time.Duration
}

// Load reads configuration from the environment with sensible defaults for local dev.
func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cors := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	var origins []string
	if cors == "" {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	} else {
		for _, o := range strings.Split(cors, ",") {
			if t := strings.TrimSpace(o); t != "" {
				origins = append(origins, t)
			}
		}
	}

	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("invalid PORT %q: %w", port, err)
	}

	mongoURI := strings.TrimSpace(os.Getenv("MONGODB_URI"))
	if mongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI is required (MongoDB Atlas connection string)")
	}

	dbName := strings.TrimSpace(os.Getenv("MONGODB_DB"))
	if dbName == "" {
		dbName = DefaultMongoDBName
	}

	chainID := int64(DefaultChainID)
	if v := strings.TrimSpace(os.Getenv("CHAIN_ID")); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid CHAIN_ID %q: %w", v, err)
		}
		chainID = parsed
	}

	syncInterval := DefaultSyncInterval
	if v := strings.TrimSpace(os.Getenv("SYNC_INTERVAL_SEC")); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil || sec < 1 {
			return Config{}, fmt.Errorf("invalid SYNC_INTERVAL_SEC %q", v)
		}
		syncInterval = time.Duration(sec) * time.Second
	}

	var deployBlock uint64
	if v := strings.TrimSpace(os.Getenv("NFT_AUCTION_DEPLOY_BLOCK")); v != "" {
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid NFT_AUCTION_DEPLOY_BLOCK %q: %w", v, err)
		}
		deployBlock = parsed
	}

	auctionContract := strings.ToLower(strings.TrimSpace(os.Getenv("NFT_AUCTION_ADDRESS")))
	sepoliaRPC := strings.TrimSpace(os.Getenv("SEPOLIA_RPC_URL"))

	return Config{
		Port:            port,
		CORSOrigins:     origins,
		MongoDBURI:      mongoURI,
		MongoDBName:     dbName,
		ChainID:         chainID,
		AuctionContract: auctionContract,
		SepoliaRPCURL:   sepoliaRPC,
		DeployBlock:     deployBlock,
		SyncInterval:    syncInterval,
	}, nil
}

// IndexerEnabled reports whether background chain sync should run.
func (c Config) IndexerEnabled() bool {
	return c.SepoliaRPCURL != "" && c.AuctionContract != ""
}
