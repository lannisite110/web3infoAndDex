package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

// AuctionRepository reads and writes auctions in MySQL.
type AuctionRepository struct {
	db *sql.DB
}

// NewAuctionRepository creates a MySQL auction repository.
func NewAuctionRepository(m *db.MySQL) *AuctionRepository {
	return &AuctionRepository{db: m.DB}
}

const auctionSelectCols = `chain_id, contract, auction_id, seller, nft_contract, token_id,
	start_price, highest_bid, highest_bidder, start_time, duration, ended, updated_at`

func scanAuction(row interface {
	Scan(dest ...any) error
}) (model.Auction, error) {
	var a model.Auction
	var ended int
	err := row.Scan(
		&a.ChainID, &a.Contract, &a.AuctionID, &a.Seller, &a.NFTContract, &a.TokenID,
		&a.StartPrice, &a.HighestBid, &a.HighestBidder, &a.StartTime, &a.Duration,
		&ended, &a.UpdatedAt,
	)
	if err != nil {
		return model.Auction{}, err
	}
	a.Ended = ended != 0
	return a, nil
}

// ListFiltered lists auctions with optional filters.
func (r *AuctionRepository) ListFiltered(
	ctx context.Context,
	chainID int64,
	contract string,
	f model.AuctionFilter,
) ([]model.Auction, error) {
	query := `SELECT ` + auctionSelectCols + ` FROM auctions WHERE chain_id = ?`
	args := []any{chainID}

	if contract != "" {
		query += ` AND contract = ?`
		args = append(args, contract)
	}
	if f.Seller != "" {
		query += ` AND seller = ?`
		args = append(args, strings.ToLower(strings.TrimSpace(f.Seller)))
	}
	if f.TokenID != "" {
		query += ` AND token_id = ?`
		args = append(args, strings.TrimSpace(f.TokenID))
	}
	if f.Ended != nil {
		ended := 0
		if *f.Ended {
			ended = 1
		}
		query += ` AND ended = ?`
		args = append(args, ended)
	}
	if len(f.AuctionIDs) > 0 {
		ph := make([]string, len(f.AuctionIDs))
		for i := range ph {
			ph[i] = "?"
		}
		query += ` AND auction_id IN (` + strings.Join(ph, ",") + `)`
		for _, id := range f.AuctionIDs {
			args = append(args, id)
		}
	}
	if f.Q != "" {
		q := strings.TrimSpace(f.Q)
		if id, err := strconv.ParseUint(q, 10, 64); err == nil {
			query += ` AND auction_id = ?`
			args = append(args, id)
		} else {
			query += ` AND (seller LIKE ? OR token_id = ?)`
			args = append(args, strings.ToLower(q)+"%", q)
		}
	}

	query += ` ORDER BY auction_id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Auction
	for rows.Next() {
		a, err := scanAuction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Auction{}
	}
	return out, nil
}

// Get returns one auction.
func (r *AuctionRepository) Get(
	ctx context.Context,
	chainID int64,
	contract string,
	auctionID uint64,
) (model.Auction, error) {
	query := `SELECT ` + auctionSelectCols + ` FROM auctions WHERE chain_id = ? AND contract = ? AND auction_id = ?`
	row := r.db.QueryRowContext(ctx, query, chainID, contract, auctionID)
	a, err := scanAuction(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Auction{}, sql.ErrNoRows
	}
	return a, err
}

// Upsert inserts or updates an auction row.
func (r *AuctionRepository) Upsert(ctx context.Context, a model.Auction) error {
	a.UpdatedAt = time.Now().UTC()
	ended := 0
	if a.Ended {
		ended = 1
	}
	query := `INSERT INTO auctions (
		chain_id, contract, auction_id, seller, nft_contract, token_id,
		start_price, highest_bid, highest_bidder, start_time, duration, ended, updated_at
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
	ON DUPLICATE KEY UPDATE
		seller=VALUES(seller), nft_contract=VALUES(nft_contract), token_id=VALUES(token_id),
		start_price=VALUES(start_price), highest_bid=VALUES(highest_bid),
		highest_bidder=VALUES(highest_bidder), start_time=VALUES(start_time),
		duration=VALUES(duration), ended=VALUES(ended), updated_at=VALUES(updated_at)`

	_, err := r.db.ExecContext(ctx, query,
		a.ChainID, a.Contract, a.AuctionID, a.Seller, a.NFTContract, a.TokenID,
		a.StartPrice, a.HighestBid, a.HighestBidder, a.StartTime, a.Duration, ended, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("mysql auction upsert: %w", err)
	}
	return nil
}
