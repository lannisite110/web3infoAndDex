package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/db"
)

// IndexerStateRepository stores sync cursors in MySQL.
type IndexerStateRepository struct {
	db *sql.DB
}

// NewIndexerStateRepository creates the repository.
func NewIndexerStateRepository(m *db.MySQL) *IndexerStateRepository {
	return &IndexerStateRepository{db: m.DB}
}

// GetLastBlock returns the stored block or 0 if missing.
func (r *IndexerStateRepository) GetLastBlock(ctx context.Context, chainID int64, contract string) (uint64, error) {
	query := `SELECT last_block FROM indexer_state WHERE chain_id = ? AND contract = ?`
	var block uint64
	err := r.db.QueryRowContext(ctx, query, chainID, contract).Scan(&block)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return block, err
}

// SetLastBlock upserts the sync cursor.
func (r *IndexerStateRepository) SetLastBlock(ctx context.Context, chainID int64, contract string, block uint64) error {
	now := time.Now().UTC()
	query := `INSERT INTO indexer_state (chain_id, contract, last_block, updated_at)
		VALUES (?,?,?,?)
		ON DUPLICATE KEY UPDATE last_block=VALUES(last_block), updated_at=VALUES(updated_at)`
	_, err := r.db.ExecContext(ctx, query, chainID, contract, block, now)
	if err != nil {
		return fmt.Errorf("mysql set last block: %w", err)
	}
	return nil
}
