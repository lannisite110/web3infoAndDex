package model

import "time"

// Bid is one on-chain Bid event indexed from NFTAuction.
type Bid struct {
	ChainID     int64     `bson:"chainId" json:"chainId"`
	Contract    string    `bson:"contract" json:"contract"`
	AuctionID   uint64    `bson:"auctionId" json:"auctionId"`
	Bidder      string    `bson:"bidder" json:"bidder"`
	Amount      string    `bson:"amount" json:"amount"`
	TxHash      string    `bson:"txHash" json:"txHash"`
	LogIndex    uint      `bson:"logIndex" json:"logIndex"`
	BlockNumber uint64    `bson:"blockNumber" json:"blockNumber"`
	IndexedAt   time.Time `bson:"indexedAt" json:"indexedAt"`
}
