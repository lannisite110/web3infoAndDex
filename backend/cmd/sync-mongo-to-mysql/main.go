package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/lannisite110/web3infoanddex/backend/internal/config"
	"github.com/lannisite110/web3infoanddex/backend/internal/db"
	mysqlrepo "github.com/lannisite110/web3infoanddex/backend/internal/repository"
	"github.com/lannisite110/web3infoanddex/backend/internal/repository/mysql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	mongo, err := db.Connect(ctx, cfg.MongoDBURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongo.Close(shutdown)
	}()

	mysqlDSN, err := config.NormalizeMySQLDSN(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("mysql dsn: %v", err)
	}
	mysqlDB, err := db.ConnectMySQL(ctx, mysqlDSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer mysqlDB.Close()

	mongoAuctions, err := mysqlrepo.NewAuctionRepository(mongo)
	if err != nil {
		log.Fatalf("mongo auctions: %v", err)
	}
	mongoBids, err := mysqlrepo.NewBidRepository(mongo)
	if err != nil {
		log.Fatalf("mongo bids: %v", err)
	}

	mysqlAuctions := mysql.NewAuctionRepository(mysqlDB)
	mysqlBids := mysql.NewBidRepository(mysqlDB)

	contract := cfg.AuctionContract
	auctions, err := mongoAuctions.List(ctx, cfg.ChainID, contract)
	if err != nil {
		log.Fatalf("list auctions: %v", err)
	}
	for _, a := range auctions {
		if err := mysqlAuctions.Upsert(ctx, a); err != nil {
			log.Fatalf("upsert auction %d: %v", a.AuctionID, err)
		}
	}
	slog.Info("synced auctions", "count", len(auctions))

	bids, err := mongoBids.List(ctx, cfg.ChainID, contract, "", 10_000)
	if err != nil {
		log.Fatalf("list bids: %v", err)
	}
	for _, b := range bids {
		if err := mysqlBids.Upsert(ctx, b); err != nil {
			log.Fatalf("upsert bid %s: %v", b.TxHash, err)
		}
	}
	slog.Info("synced bids", "count", len(bids))
	slog.Info("mongo to mysql sync complete")
}
