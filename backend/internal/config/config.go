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
	DefaultCacheTTL      = 60 * time.Second
	DefaultOpenSeaBase   = "https://api.opensea.io/api/v2"
	DefaultStorageRead   = "auto"
)

// Config holds runtime settings loaded from environment variables.
type Config struct {
	Port            string
	CORSOrigins     []string
	MongoDBURI      string
	MongoDBName     string
	MySQLDSN        string
	RedisURL        string
	CacheTTL        time.Duration
	StorageRead     string
	ChainID         int64
	AuctionContract string
	SepoliaRPCURL   string
	DeployBlock     uint64
	SyncInterval    time.Duration
	EtherscanAPIKey string
	OpenSeaAPIKey   string
	OpenSeaBaseURL  string
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

	mongoURI := sanitizeEnv(os.Getenv("MONGODB_URI"))
	if mongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI is required (MongoDB Atlas connection string)")
	}
	if err := ValidateMongoURI(mongoURI); err != nil {
		return Config{}, err
	}

	dbName := sanitizeEnv(os.Getenv("MONGODB_DB"))
	if dbName == "" {
		dbName = DefaultMongoDBName
	}

	mysqlDSN := sanitizeEnv(os.Getenv("MYSQL_DSN"))

	redisURL := sanitizeEnv(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required (Upstash rediss:// URL)")
	}

	cacheTTL := DefaultCacheTTL
	if v := strings.TrimSpace(os.Getenv("CACHE_TTL_SEC")); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil || sec < 1 {
			return Config{}, fmt.Errorf("invalid CACHE_TTL_SEC %q", v)
		}
		cacheTTL = time.Duration(sec) * time.Second
	}

	storageRead := strings.ToLower(sanitizeEnv(os.Getenv("STORAGE_READ")))
	if storageRead == "" {
		storageRead = DefaultStorageRead
	}
	switch storageRead {
	case "mysql", "mongo", "auto":
	default:
		return Config{}, fmt.Errorf("STORAGE_READ must be mysql, mongo, or auto (got %q)", storageRead)
	}

	if mysqlDSN == "" && storageRead == "mysql" {
		return Config{}, fmt.Errorf("MYSQL_DSN is required when STORAGE_READ=mysql")
	}
	if mysqlDSN == "" && storageRead == "auto" {
		// MySQL optional: API/indexer use Mongo when Railway MySQL is unavailable.
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

	auctionContract := strings.ToLower(sanitizeEnv(os.Getenv("NFT_AUCTION_ADDRESS")))
	sepoliaRPC := sanitizeEnv(os.Getenv("SEPOLIA_RPC_URL"))
	if err := ValidateRPCURL(sepoliaRPC); err != nil {
		return Config{}, err
	}

	openSeaBase := sanitizeEnv(os.Getenv("OPENSEA_BASE_URL"))
	if openSeaBase == "" {
		openSeaBase = DefaultOpenSeaBase
	}

	return Config{
		Port:            port,
		CORSOrigins:     origins,
		MongoDBURI:      mongoURI,
		MongoDBName:     dbName,
		MySQLDSN:        mysqlDSN,
		RedisURL:        redisURL,
		CacheTTL:        cacheTTL,
		StorageRead:     storageRead,
		ChainID:         chainID,
		AuctionContract: auctionContract,
		SepoliaRPCURL:   sepoliaRPC,
		DeployBlock:     deployBlock,
		SyncInterval:    syncInterval,
		EtherscanAPIKey: sanitizeEnv(os.Getenv("ETHERSCAN_API_KEY")),
		OpenSeaAPIKey:   sanitizeEnv(os.Getenv("OPENSEA_API_KEY")),
		OpenSeaBaseURL:  openSeaBase,
	}, nil
}

// IndexerEnabled reports whether background chain sync should run.
func (c Config) IndexerEnabled() bool {
	return c.SepoliaRPCURL != "" && c.AuctionContract != ""
}

// EtherscanEnabled reports whether the tx lookup API can call Etherscan.
func (c Config) EtherscanEnabled() bool {
	return c.EtherscanAPIKey != ""
}

// sanitizeEnv trims whitespace and surrounding quotes from dashboard-pasted values.
func sanitizeEnv(v string) string {
	v = strings.TrimSpace(v)
	return strings.Trim(v, `"'`)
}
