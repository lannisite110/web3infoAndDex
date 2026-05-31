package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lannisite110/web3infoanddex/backend/internal/cache"
	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/eth"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
)

// Indexer syncs NFTAuction events and on-chain state into MongoDB and MySQL.
type Indexer struct {
	cfg       config.Config
	eth       *eth.Client
	auctions  repository.AuctionStore
	bids      repository.BidStore
	state     repository.IndexerStateStore
	cache     *cache.Client
	contract  common.Address
	createdID common.Hash
	bidID     common.Hash
	endedID   common.Hash
}

// New builds an Indexer for the configured auction contract.
func New(
	cfg config.Config,
	ethClient *eth.Client,
	auctions repository.AuctionStore,
	bids repository.BidStore,
	state repository.IndexerStateStore,
	cacheClient *cache.Client,
) (*Indexer, error) {
	if cfg.AuctionContract == "" {
		return nil, fmt.Errorf("NFT_AUCTION_ADDRESS is required for indexer")
	}
	if !common.IsHexAddress(cfg.AuctionContract) {
		return nil, fmt.Errorf("invalid NFT_AUCTION_ADDRESS %q", cfg.AuctionContract)
	}

	parsed := ethClient.ParsedABI()
	created, ok := parsed.Events["AuctionCreated"]
	if !ok {
		return nil, fmt.Errorf("ABI missing AuctionCreated event")
	}
	bid, ok := parsed.Events["Bid"]
	if !ok {
		return nil, fmt.Errorf("ABI missing Bid event")
	}
	ended, ok := parsed.Events["AuctionEnded"]
	if !ok {
		return nil, fmt.Errorf("ABI missing AuctionEnded event")
	}

	return &Indexer{
		cfg:       cfg,
		eth:       ethClient,
		auctions:  auctions,
		bids:      bids,
		state:     state,
		cache:     cacheClient,
		contract:  common.HexToAddress(cfg.AuctionContract),
		createdID: created.ID,
		bidID:     bid.ID,
		endedID:   ended.ID,
	}, nil
}

// Run performs initial backfill then polls for new blocks until ctx is cancelled.
func (idx *Indexer) Run(ctx context.Context) {
	slog.Info("indexer starting",
		"contract", idx.contract.Hex(),
		"interval", idx.cfg.SyncInterval.String(),
	)

	if err := idx.backfillBids(ctx); err != nil {
		slog.Error("indexer bid backfill failed", "err", err)
	}

	if err := idx.syncOnce(ctx, true); err != nil {
		slog.Error("indexer initial sync failed", "err", err)
	}

	ticker := time.NewTicker(idx.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("indexer stopped")
			return
		case <-ticker.C:
			if err := idx.syncOnce(ctx, false); err != nil {
				slog.Error("indexer sync failed", "err", err)
			}
		}
	}
}

func (idx *Indexer) backfillAuctions(ctx context.Context) error {
	count, err := idx.eth.AuctionCount(ctx, idx.contract)
	if err != nil {
		return err
	}
	slog.Info("indexer backfill", "auctionCount", count)

	contractHex := strings.ToLower(idx.contract.Hex())
	for id := uint64(1); id <= count; id++ {
		auction, err := idx.eth.FetchAuction(ctx, idx.contract, idx.cfg.ChainID, id)
		if err != nil {
			return fmt.Errorf("fetch auction %d: %w", id, err)
		}
		auction.Contract = contractHex
		if err := idx.auctions.Upsert(ctx, auction); err != nil {
			return err
		}
	}
	slog.Info("indexer backfill done", "auctionCount", count)
	return nil
}

func (idx *Indexer) backfillBids(ctx context.Context) error {
	head, err := idx.eth.Raw().BlockNumber(ctx)
	if err != nil {
		return err
	}

	from := idx.cfg.DeployBlock
	if from == 0 {
		if head > 200_000 {
			from = head - 200_000
		}
		slog.Warn("NFT_AUCTION_DEPLOY_BLOCK not set; bid backfill uses recent blocks only",
			"from", from, "to", head,
		)
	}

	slog.Info("indexer bid backfill", "from", from, "to", head)
	const maxRange = uint64(2000)
	recorded := 0
	for start := from; start <= head; {
		end := start + maxRange - 1
		if end > head {
			end = head
		}
		n, err := idx.processBidRange(ctx, start, end)
		if err != nil {
			return err
		}
		recorded += n
		start = end + 1
	}
	slog.Info("indexer bid backfill done", "bids", recorded)
	return nil
}

func (idx *Indexer) processBidRange(ctx context.Context, from, to uint64) (int, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{idx.contract},
		Topics:    [][]common.Hash{{idx.bidID}},
	}

	logs, err := idx.eth.Raw().FilterLogs(ctx, query)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, lg := range logs {
		if err := idx.recordBid(ctx, lg); err != nil {
			slog.Warn("record bid", "tx", lg.TxHash.Hex(), "err", err)
			continue
		}
		count++
	}
	return count, nil
}

func (idx *Indexer) syncOnce(ctx context.Context, forceBackfill bool) error {
	head, err := idx.eth.Raw().BlockNumber(ctx)
	if err != nil {
		return err
	}

	contractHex := strings.ToLower(idx.contract.Hex())
	last, err := idx.state.GetLastBlock(ctx, idx.cfg.ChainID, contractHex)
	if err != nil {
		return err
	}

	if last == 0 {
		if idx.cfg.DeployBlock > 0 {
			last = idx.cfg.DeployBlock - 1
		} else {
			last = head
		}
	}

	if forceBackfill {
		if err := idx.backfillAuctions(ctx); err != nil {
			return err
		}
	}

	from := last + 1
	if from > head {
		return nil
	}

	const maxRange = uint64(2000)
	for start := from; start <= head; {
		end := start + maxRange - 1
		if end > head {
			end = head
		}
		if err := idx.processRange(ctx, start, end); err != nil {
			return err
		}
		if err := idx.state.SetLastBlock(ctx, idx.cfg.ChainID, contractHex, end); err != nil {
			return err
		}
		start = end + 1
	}

	slog.Info("indexer synced", "from", from, "to", head)
	return nil
}

func (idx *Indexer) processRange(ctx context.Context, from, to uint64) error {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{idx.contract},
		Topics: [][]common.Hash{
			{idx.createdID, idx.bidID, idx.endedID},
		},
	}

	logs, err := idx.eth.Raw().FilterLogs(ctx, query)
	if err != nil {
		return err
	}

	seen := make(map[uint64]struct{})
	for _, lg := range logs {
		if lg.Topics[0] == idx.bidID {
			if err := idx.recordBid(ctx, lg); err != nil {
				slog.Warn("record bid", "tx", lg.TxHash.Hex(), "err", err)
			}
		}

		id, err := idx.auctionIDFromLog(lg)
		if err != nil {
			slog.Warn("skip log", "tx", lg.TxHash.Hex(), "err", err)
			continue
		}
		seen[id] = struct{}{}
	}

	for id := range seen {
		if err := idx.syncAuction(ctx, id); err != nil {
			slog.Warn("sync auction", "id", id, "err", err)
		}
	}
	return nil
}

func (idx *Indexer) recordBid(ctx context.Context, lg types.Log) error {
	var out struct {
		AuctionId *big.Int
		Bidder    common.Address
		Amount    *big.Int
	}
	if err := idx.eth.ParsedABI().UnpackIntoInterface(&out, "Bid", lg.Data); err != nil {
		return err
	}

	bid := model.Bid{
		ChainID:     idx.cfg.ChainID,
		Contract:    strings.ToLower(idx.contract.Hex()),
		AuctionID:   out.AuctionId.Uint64(),
		Bidder:      strings.ToLower(out.Bidder.Hex()),
		Amount:      out.Amount.String(),
		TxHash:      strings.ToLower(lg.TxHash.Hex()),
		LogIndex:    lg.Index,
		BlockNumber: lg.BlockNumber,
	}
	if err := idx.bids.Upsert(ctx, bid); err != nil {
		return err
	}
	if idx.cache != nil {
		idx.cache.InvalidateBid(ctx, idx.cfg.ChainID, bid.Contract, bid.AuctionID)
	}
	return nil
}

func (idx *Indexer) auctionIDFromLog(lg types.Log) (uint64, error) {
	switch lg.Topics[0] {
	case idx.createdID:
		var out struct {
			AuctionId *big.Int
		}
		if err := idx.eth.ParsedABI().UnpackIntoInterface(&out, "AuctionCreated", lg.Data); err != nil {
			return 0, err
		}
		return out.AuctionId.Uint64(), nil
	case idx.bidID:
		var out struct {
			AuctionId *big.Int
		}
		if err := idx.eth.ParsedABI().UnpackIntoInterface(&out, "Bid", lg.Data); err != nil {
			return 0, err
		}
		return out.AuctionId.Uint64(), nil
	case idx.endedID:
		var out struct {
			AuctionId *big.Int
		}
		if err := idx.eth.ParsedABI().UnpackIntoInterface(&out, "AuctionEnded", lg.Data); err != nil {
			return 0, err
		}
		return out.AuctionId.Uint64(), nil
	default:
		return 0, fmt.Errorf("unknown topic")
	}
}

func (idx *Indexer) syncAuction(ctx context.Context, auctionID uint64) error {
	auction, err := idx.eth.FetchAuction(ctx, idx.contract, idx.cfg.ChainID, auctionID)
	if err != nil {
		return err
	}
	if err := idx.auctions.Upsert(ctx, auction); err != nil {
		return err
	}
	if idx.cache != nil {
		idx.cache.InvalidateAuction(ctx, idx.cfg.ChainID, auction.Contract, auction.AuctionID)
	}
	return nil
}
