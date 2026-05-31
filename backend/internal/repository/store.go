package repository

import (
	"context"

	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

// AuctionStore supports API reads and indexer writes.
type AuctionStore interface {
	ListFiltered(ctx context.Context, chainID int64, contract string, f model.AuctionFilter) ([]model.Auction, error)
	Get(ctx context.Context, chainID int64, contract string, auctionID uint64) (model.Auction, error)
	Upsert(ctx context.Context, a model.Auction) error
}

// BidStore supports API reads and indexer writes.
type BidStore interface {
	Upsert(ctx context.Context, b model.Bid) error
	ListByAuction(ctx context.Context, chainID int64, contract string, auctionID uint64) ([]model.Bid, error)
	List(ctx context.Context, chainID int64, contract string, bidder string, limit int) ([]model.Bid, error)
	AuctionIDsForBidder(ctx context.Context, chainID int64, contract string, bidder string) ([]uint64, error)
}

// IndexerStateStore tracks the last processed block per contract.
type IndexerStateStore interface {
	GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error)
	SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error
}
