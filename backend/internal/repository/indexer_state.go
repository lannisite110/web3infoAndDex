package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const indexerStateCollection = "indexer_state"

// IndexerState tracks the last processed block per contract.
type IndexerState struct {
	ChainID   int64  `bson:"chainId"`
	Contract  string `bson:"contract"`
	LastBlock uint64 `bson:"lastBlock"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

// IndexerStateRepository stores sync cursors.
type IndexerStateRepository struct {
	col *mongo.Collection
}

// NewIndexerStateRepository creates the repository.
func NewIndexerStateRepository(m *db.Mongo) *IndexerStateRepository {
	return &IndexerStateRepository{col: m.Collection(indexerStateCollection)}
}

// GetLastBlock returns the stored block or 0 if missing.
func (r *IndexerStateRepository) GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error) {
	filter := bson.M{"chainId": chainID, "contract": contract}
	var st IndexerState
	err := r.col.FindOne(ctx, filter).Decode(&st)
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return st.LastBlock, nil
}

// SetLastBlock upserts the sync cursor.
func (r *IndexerStateRepository) SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error {
	filter := bson.M{"chainId": chainID, "contract": contract}
	update := bson.M{
		"$set": IndexerState{
			ChainID:   chainID,
			Contract:  contract,
			LastBlock: block,
			UpdatedAt: time.Now().UTC(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("set last block: %w", err)
	}
	return nil
}
