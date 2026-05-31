package repository

import (
	"context"
	"log/slog"

	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

// DualAuctionStore writes to MongoDB and MySQL; reads use MySQL only via handler.
type DualAuctionStore struct {
	mongo AuctionStore
	mysql AuctionStore
}

// NewDualAuctionStore creates a dual-write auction store.
func NewDualAuctionStore(mongo, mysql AuctionStore) *DualAuctionStore {
	return &DualAuctionStore{mongo: mongo, mysql: mysql}
}

func (d *DualAuctionStore) ListFiltered(ctx context.Context, chainID int64, contract string, f model.AuctionFilter) ([]model.Auction, error) {
	return d.mysql.ListFiltered(ctx, chainID, contract, f)
}

func (d *DualAuctionStore) Get(ctx context.Context, chainID int64, contract string, auctionID uint64) (model.Auction, error) {
	return d.mysql.Get(ctx, chainID, contract, auctionID)
}

func (d *DualAuctionStore) Upsert(ctx context.Context, a model.Auction) error {
	if err := d.mongo.Upsert(ctx, a); err != nil {
		return err
	}
	if err := d.mysql.Upsert(ctx, a); err != nil {
		slog.Warn("mysql auction upsert failed", "auctionId", a.AuctionID, "err", err)
	}
	return nil
}

// DualBidStore writes to MongoDB and MySQL.
type DualBidStore struct {
	mongo BidStore
	mysql BidStore
}

// NewDualBidStore creates a dual-write bid store.
func NewDualBidStore(mongo, mysql BidStore) *DualBidStore {
	return &DualBidStore{mongo: mongo, mysql: mysql}
}

func (d *DualBidStore) Upsert(ctx context.Context, b model.Bid) error {
	if err := d.mongo.Upsert(ctx, b); err != nil {
		return err
	}
	if err := d.mysql.Upsert(ctx, b); err != nil {
		slog.Warn("mysql bid upsert failed", "tx", b.TxHash, "err", err)
	}
	return nil
}

func (d *DualBidStore) ListByAuction(ctx context.Context, chainID int64, contract string, auctionID uint64) ([]model.Bid, error) {
	return d.mysql.ListByAuction(ctx, chainID, contract, auctionID)
}

func (d *DualBidStore) List(ctx context.Context, chainID int64, contract string, bidder string, limit int) ([]model.Bid, error) {
	return d.mysql.List(ctx, chainID, contract, bidder, limit)
}

func (d *DualBidStore) AuctionIDsForBidder(ctx context.Context, chainID int64, contract string, bidder string) ([]uint64, error) {
	return d.mysql.AuctionIDsForBidder(ctx, chainID, contract, bidder)
}

// DualIndexerStateStore writes cursor to MongoDB and MySQL.
type DualIndexerStateStore struct {
	mongo IndexerStateStore
	mysql IndexerStateStore
}

// NewDualIndexerStateStore creates dual-write indexer state.
func NewDualIndexerStateStore(mongo, mysql IndexerStateStore) *DualIndexerStateStore {
	return &DualIndexerStateStore{mongo: mongo, mysql: mysql}
}

func (d *DualIndexerStateStore) GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error) {
	last, err := d.mongo.GetLastBlock(ctx, chainID, contract)
	if err != nil {
		return 0, err
	}
	if last > 0 {
		return last, nil
	}
	return d.mysql.GetLastBlock(ctx, chainID, contract)
}

func (d *DualIndexerStateStore) SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error {
	if err := d.mongo.SetLastBlock(ctx, chainID, contract, block); err != nil {
		return err
	}
	if err := d.mysql.SetLastBlock(ctx, chainID, contract, block); err != nil {
		slog.Warn("mysql indexer state failed", "block", block, "err", err)
	}
	return nil
}
