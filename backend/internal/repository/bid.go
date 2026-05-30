package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const bidsCollection = "bids"

// BidRepository persists indexed bid events.
type BidRepository struct {
	col *mongo.Collection
}

// NewBidRepository creates a repository and ensures indexes exist.
func NewBidRepository(m *db.Mongo) (*BidRepository, error) {
	repo := &BidRepository{col: m.Collection(bidsCollection)}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "chainId", Value: 1},
				{Key: "contract", Value: 1},
				{Key: "txHash", Value: 1},
				{Key: "logIndex", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "chainId", Value: 1},
				{Key: "contract", Value: 1},
				{Key: "auctionId", Value: 1},
				{Key: "blockNumber", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "chainId", Value: 1},
				{Key: "contract", Value: 1},
				{Key: "bidder", Value: 1},
			},
		},
	}

	if _, err := repo.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return nil, fmt.Errorf("bid indexes: %w", err)
	}

	return repo, nil
}

// Upsert inserts or updates a bid by (chainId, contract, txHash, logIndex).
func (r *BidRepository) Upsert(ctx context.Context, b model.Bid) error {
	b.IndexedAt = time.Now().UTC()
	filter := bson.M{
		"chainId":  b.ChainID,
		"contract": b.Contract,
		"txHash":   b.TxHash,
		"logIndex": b.LogIndex,
	}
	update := bson.M{"$set": b}
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	return err
}

// ListByAuction returns bids for one auction, oldest first.
func (r *BidRepository) ListByAuction(
	ctx context.Context,
	chainID int64,
	contract string,
	auctionID uint64,
) ([]model.Bid, error) {
	filter := bson.M{
		"chainId":   chainID,
		"contract":  contract,
		"auctionId": auctionID,
	}
	opts := options.Find().SetSort(bson.D{{Key: "blockNumber", Value: 1}, {Key: "logIndex", Value: 1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Bid
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Bid{}
	}
	return out, nil
}

// List returns bids with optional bidder filter.
func (r *BidRepository) List(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
	limit int,
) ([]model.Bid, error) {
	filter := bson.M{"chainId": chainID}
	if contract != "" {
		filter["contract"] = contract
	}
	if bidder != "" {
		filter["bidder"] = bidder
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "blockNumber", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Bid
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Bid{}
	}
	return out, nil
}

// AuctionIDsForBidder returns distinct auction IDs a bidder participated in.
func (r *BidRepository) AuctionIDsForBidder(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
) ([]uint64, error) {
	filter := bson.M{
		"chainId":  chainID,
		"contract": contract,
		"bidder":   bidder,
	}
	vals, err := r.col.Distinct(ctx, "auctionId", filter)
	if err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(vals))
	for _, v := range vals {
		switch n := v.(type) {
		case int32:
			out = append(out, uint64(n))
		case int64:
			out = append(out, uint64(n))
		case uint64:
			out = append(out, n)
		}
	}
	return out, nil
}
