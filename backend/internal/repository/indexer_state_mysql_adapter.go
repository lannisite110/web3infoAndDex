package repository

import (
	"context"

	mysqlrepo "github.com/lannisite110/web3infoanddex/backend/internal/repository/mysql"
)

// mongoIndexerState adapts IndexerStateRepository to IndexerStateStore.
type mongoIndexerState struct {
	*IndexerStateRepository
}

func (r *mongoIndexerState) GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error) {
	return r.IndexerStateRepository.GetLastBlock(ctx, chainID, contract)
}

func (r *mongoIndexerState) SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error {
	return r.IndexerStateRepository.SetLastBlock(ctx, chainID, contract, block)
}

// NewMongoIndexerStateStore wraps the Mongo indexer state repo.
func NewMongoIndexerStateStore(r *IndexerStateRepository) IndexerStateStore {
	return &mongoIndexerState{r}
}

// mysqlIndexerState adapts mysql.IndexerStateRepository.
type mysqlIndexerState struct {
	*mysqlrepo.IndexerStateRepository
}

func (r *mysqlIndexerState) GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error) {
	return r.IndexerStateRepository.GetLastBlock(ctx, chainID, contract)
}

func (r *mysqlIndexerState) SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error {
	return r.IndexerStateRepository.SetLastBlock(ctx, chainID, contract, block)
}

// NewMySQLIndexerStateStore wraps the MySQL indexer state repo.
func NewMySQLIndexerStateStore(r *mysqlrepo.IndexerStateRepository) IndexerStateStore {
	return &mysqlIndexerState{r}
}
