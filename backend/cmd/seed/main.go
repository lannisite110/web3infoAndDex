// Seed inserts one sample auction document for local API testing (phase 2.2).
//
// Usage:
//
//	MONGODB_URI=... NFT_AUCTION_ADDRESS=0x... go run ./cmd/seed
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository"
	mysqlrepo "github.com/lannisite110/web3infoanddex/backend/internal/repository/mysql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.AuctionContract == "" {
		log.Fatal("NFT_AUCTION_ADDRESS is required for seed data")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mongo, err := db.Connect(ctx, cfg.MongoDBURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer mongo.Close(context.Background())

	repo, err := repository.NewAuctionRepository(mongo)
	if err != nil {
		log.Fatalf("repository: %v", err)
	}

	sample := model.Auction{
		ChainID:       cfg.ChainID,
		Contract:      cfg.AuctionContract, // already lowercased in config.Load
		AuctionID:     1,
		Seller:        "0x927bdb433b7380175d177401ead64d86069188a0",
		NFTContract:   os.Getenv("TEST_NFT_ADDRESS"),
		TokenID:       "10",
		StartPrice:    "10000000000000000",
		HighestBid:    "0",
		HighestBidder: "0x0000000000000000000000000000000000000000",
		StartTime:     time.Now().Unix(),
		Duration:      1800,
		Ended:         false,
	}

	if sample.NFTContract == "" {
		sample.NFTContract = "0x8D8AD875810933D40dba91378c680d39223114c9"
	}
	sample.NFTContract = strings.ToLower(sample.NFTContract)

	if err := repo.Upsert(ctx, sample); err != nil {
		log.Fatalf("mongo upsert: %v", err)
	}

	mysqlDSN, err := config.NormalizeMySQLDSN(cfg.MySQLDSN)
	if err == nil {
		mysqlDB, err := db.ConnectMySQL(ctx, mysqlDSN)
		if err == nil {
			defer mysqlDB.Close()
			mysqlAuctions := mysqlrepo.NewAuctionRepository(mysqlDB)
			if err := mysqlAuctions.Upsert(ctx, sample); err != nil {
				log.Printf("mysql seed upsert (optional): %v", err)
			}
		}
	}

	log.Printf("seeded auction #%d on contract %s", sample.AuctionID, sample.Contract)
}
