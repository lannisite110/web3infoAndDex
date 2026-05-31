package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

// ReadFallbackAuctionStore reads MySQL first; on error or missing row, reads MongoDB.
type ReadFallbackAuctionStore struct {
	primary AuctionStore
	fallback AuctionStore
}

// NewReadFallbackAuctionStore creates an auto-failover read store.
func NewReadFallbackAuctionStore(primary, fallback AuctionStore) *ReadFallbackAuctionStore {
	return &ReadFallbackAuctionStore{primary: primary, fallback: fallback}
}

func (s *ReadFallbackAuctionStore) ListFiltered(
	ctx context.Context,
	chainID int64,
	contract string,
	f model.AuctionFilter,
) ([]model.Auction, error) {
	out, err := s.primary.ListFiltered(ctx, chainID, contract, f)
	if err == nil {
		return out, nil
	}
	slog.Warn("auction list mysql failed, using mongodb", "err", err)
	return s.fallback.ListFiltered(ctx, chainID, contract, f)
}

func (s *ReadFallbackAuctionStore) Get(
	ctx context.Context,
	chainID int64,
	contract string,
	auctionID uint64,
) (model.Auction, error) {
	a, err := s.primary.Get(ctx, chainID, contract, auctionID)
	if err == nil {
		return a, nil
	}
	if shouldFallbackAuctionGet(err) {
		slog.Warn("auction get mysql unavailable, trying mongodb", "auctionId", auctionID, "err", err)
		return s.fallback.Get(ctx, chainID, contract, auctionID)
	}
	return a, err
}

func (s *ReadFallbackAuctionStore) Upsert(ctx context.Context, a model.Auction) error {
	return s.primary.Upsert(ctx, a)
}

func shouldFallbackAuctionGet(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, sql.ErrNoRows) || isReadInfrastructureError(err)
}

// ReadFallbackBidStore reads MySQL first; on error, reads MongoDB.
type ReadFallbackBidStore struct {
	primary  BidStore
	fallback BidStore
}

// NewReadFallbackBidStore creates an auto-failover bid read store.
func NewReadFallbackBidStore(primary, fallback BidStore) *ReadFallbackBidStore {
	return &ReadFallbackBidStore{primary: primary, fallback: fallback}
}

func (s *ReadFallbackBidStore) Upsert(ctx context.Context, b model.Bid) error {
	return s.primary.Upsert(ctx, b)
}

func (s *ReadFallbackBidStore) ListByAuction(
	ctx context.Context,
	chainID int64,
	contract string,
	auctionID uint64,
) ([]model.Bid, error) {
	out, err := s.primary.ListByAuction(ctx, chainID, contract, auctionID)
	if err == nil {
		return out, nil
	}
	slog.Warn("bids by auction mysql failed, using mongodb", "auctionId", auctionID, "err", err)
	return s.fallback.ListByAuction(ctx, chainID, contract, auctionID)
}

func (s *ReadFallbackBidStore) List(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
	limit int,
) ([]model.Bid, error) {
	out, err := s.primary.List(ctx, chainID, contract, bidder, limit)
	if err == nil {
		return out, nil
	}
	slog.Warn("bids list mysql failed, using mongodb", "err", err)
	return s.fallback.List(ctx, chainID, contract, bidder, limit)
}

func (s *ReadFallbackBidStore) AuctionIDsForBidder(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
) ([]uint64, error) {
	out, err := s.primary.AuctionIDsForBidder(ctx, chainID, contract, bidder)
	if err == nil {
		return out, nil
	}
	slog.Warn("bidder auction ids mysql failed, using mongodb", "err", err)
	return s.fallback.AuctionIDsForBidder(ctx, chainID, contract, bidder)
}

func isReadInfrastructureError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return stringsContainsAny(msg,
		"connection", "timeout", "refused", "denied", "TLS", "tls",
		"invalid connection", "driver", "bad connection", "no such host",
	)
}

func stringsContainsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
