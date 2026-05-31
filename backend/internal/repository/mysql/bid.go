package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

// BidRepository reads and writes bids in MySQL.
type BidRepository struct {
	db *sql.DB
}

// NewBidRepository creates a MySQL bid repository.
func NewBidRepository(m *db.MySQL) *BidRepository {
	return &BidRepository{db: m.DB}
}

// Upsert inserts or updates a bid row.
func (r *BidRepository) Upsert(ctx context.Context, b model.Bid) error {
	b.IndexedAt = time.Now().UTC()
	query := `INSERT INTO bids (
		chain_id, contract, auction_id, bidder, amount, tx_hash, log_index, block_number, indexed_at
	) VALUES (?,?,?,?,?,?,?,?,?)
	ON DUPLICATE KEY UPDATE
		bidder=VALUES(bidder), amount=VALUES(amount), block_number=VALUES(block_number), indexed_at=VALUES(indexed_at)`

	_, err := r.db.ExecContext(ctx, query,
		b.ChainID, b.Contract, b.AuctionID, b.Bidder, b.Amount, b.TxHash, b.LogIndex, b.BlockNumber, b.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("mysql bid upsert: %w", err)
	}
	return nil
}

// ListByAuction returns bids for one auction, oldest first.
func (r *BidRepository) ListByAuction(
	ctx context.Context,
	chainID int64,
	contract string,
	auctionID uint64,
) ([]model.Bid, error) {
	query := `SELECT chain_id, contract, auction_id, bidder, amount, tx_hash, log_index, block_number, indexed_at
		FROM bids WHERE chain_id = ? AND contract = ? AND auction_id = ?
		ORDER BY block_number ASC, log_index ASC`

	rows, err := r.db.QueryContext(ctx, query, chainID, contract, auctionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBids(rows)
}

// List returns bids with optional bidder filter.
func (r *BidRepository) List(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
	limit int,
) ([]model.Bid, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `SELECT chain_id, contract, auction_id, bidder, amount, tx_hash, log_index, block_number, indexed_at
		FROM bids WHERE chain_id = ?`
	args := []any{chainID}
	if contract != "" {
		query += ` AND contract = ?`
		args = append(args, contract)
	}
	if bidder != "" {
		query += ` AND bidder = ?`
		args = append(args, bidder)
	}
	query += ` ORDER BY block_number DESC, log_index DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBids(rows)
}

// AuctionIDsForBidder returns distinct auction IDs for a bidder.
func (r *BidRepository) AuctionIDsForBidder(
	ctx context.Context,
	chainID int64,
	contract string,
	bidder string,
) ([]uint64, error) {
	query := `SELECT DISTINCT auction_id FROM bids WHERE chain_id = ? AND contract = ? AND bidder = ?`
	rows, err := r.db.QueryContext(ctx, query, chainID, contract, strings.ToLower(bidder))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanBids(rows rowScanner) ([]model.Bid, error) {
	var out []model.Bid
	for rows.Next() {
		var b model.Bid
		if err := rows.Scan(
			&b.ChainID, &b.Contract, &b.AuctionID, &b.Bidder, &b.Amount,
			&b.TxHash, &b.LogIndex, &b.BlockNumber, &b.IndexedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Bid{}
	}
	return out, nil
}
