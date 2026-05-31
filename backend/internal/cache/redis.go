package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyAuctionList = "auctions:list:"
	keyAuction     = "auction:"
	keyBidsAuction = "bids:auction:"
	keyBidsGlobal  = "bids:global:"
)

// Client wraps Redis for API response caching.
type Client struct {
	rdb *redis.Client
	ttl time.Duration
}

// Connect parses REDIS_URL and pings Redis (Upstash rediss:// supported).
func Connect(ctx context.Context, redisURL string, ttl time.Duration) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb, ttl: ttl}, nil
}

// Ping checks connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis client.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// GetJSON loads and unmarshals a cached value. ok is false on miss.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, err
	}
	return true, nil
}

// SetJSON marshals and stores with TTL.
func (c *Client) SetJSON(ctx context.Context, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

// AuctionListKey builds a cache key from query parameters.
func AuctionListKey(chainID int64, contract string, query string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", chainID, contract, query)))
	return keyAuctionList + hex.EncodeToString(h[:8])
}

func AuctionKey(chainID int64, contract string, auctionID uint64) string {
	return fmt.Sprintf("%s%d:%s:%d", keyAuction, chainID, contract, auctionID)
}

func BidsAuctionKey(chainID int64, contract string, auctionID uint64) string {
	return fmt.Sprintf("%s%d:%s:%d", keyBidsAuction, chainID, contract, auctionID)
}

func BidsGlobalKey(chainID int64, contract, bidder string, limit int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s:%d", chainID, contract, bidder, limit)))
	return keyBidsGlobal + hex.EncodeToString(h[:8])
}

// InvalidateAuction drops detail and bid caches for one auction and list caches.
func (c *Client) InvalidateAuction(ctx context.Context, chainID int64, contract string, auctionID uint64) {
	keys := []string{
		AuctionKey(chainID, contract, auctionID),
		BidsAuctionKey(chainID, contract, auctionID),
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		slog.Warn("redis invalidate auction", "err", err)
	}
	c.invalidatePrefix(ctx, keyAuctionList)
	c.invalidatePrefix(ctx, keyBidsGlobal)
}

// InvalidateBid drops bid-related caches after a new bid.
func (c *Client) InvalidateBid(ctx context.Context, chainID int64, contract string, auctionID uint64) {
	c.InvalidateAuction(ctx, chainID, contract, auctionID)
}

func (c *Client) invalidatePrefix(ctx context.Context, prefix string) {
	iter := c.rdb.Scan(ctx, 0, prefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 50 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				slog.Warn("redis scan del", "err", err)
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		slog.Warn("redis scan", "prefix", prefix, "err", err)
		return
	}
	if len(keys) > 0 {
		if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
			slog.Warn("redis scan del", "err", err)
		}
	}
}

// BuildListQueryKey normalizes auction list query string for cache keys.
func BuildListQueryKey(params map[string]string) string {
	parts := make([]string, 0, len(params))
	for k, v := range params {
		if strings.TrimSpace(v) != "" {
			parts = append(parts, k+"="+v)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	// stable order for same logical query
	for i := 0; i < len(parts); i++ {
		for j := i + 1; j < len(parts); j++ {
			if parts[j] < parts[i] {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}
	return strings.Join(parts, "&")
}
