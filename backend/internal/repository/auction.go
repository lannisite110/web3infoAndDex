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

const auctionsCollection = "auctions"

// AuctionRepository persists indexed auction records.
type AuctionRepository struct {
	col *mongo.Collection
}

// NewAuctionRepository creates a repository and ensures indexes exist.
func NewAuctionRepository(m *db.Mongo) (*AuctionRepository, error) {
	repo := &AuctionRepository{col: m.Collection(auctionsCollection)}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "chainId", Value: 1},
				{Key: "contract", Value: 1},
				{Key: "auctionId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "updatedAt", Value: -1}},
		},
	}

	if _, err := repo.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return nil, fmt.Errorf("auction indexes: %w", err)
	}

	return repo, nil
}

// List returns all auctions for a chain and contract, newest first.
func (r *AuctionRepository) List(ctx context.Context, chainID int64, contract string) ([]model.Auction, error) {
	filter := bson.M{"chainId": chainID}
	if contract != "" {
		filter["contract"] = contract
	}

	opts := options.Find().SetSort(bson.D{{Key: "auctionId", Value: 1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Auction
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Auction{}
	}
	return out, nil
}

// Upsert inserts or updates an auction by (chainId, contract, auctionId).
func (r *AuctionRepository) Upsert(ctx context.Context, a model.Auction) error {
	a.UpdatedAt = time.Now().UTC()
	filter := bson.M{
		"chainId":   a.ChainID,
		"contract":  a.Contract,
		"auctionId": a.AuctionID,
	}
	update := bson.M{"$set": a}
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	return err
}
