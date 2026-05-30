package model

import "time"

// Auction is the MongoDB document for an on-chain NFT auction (indexed from events).
type Auction struct {
	ChainID       int64     `bson:"chainId" json:"chainId"`
	Contract      string    `bson:"contract" json:"contract"`
	AuctionID     uint64    `bson:"auctionId" json:"auctionId"`
	Seller        string    `bson:"seller" json:"seller"`
	NFTContract   string    `bson:"nftContract" json:"nftContract"`
	TokenID       string    `bson:"tokenId" json:"tokenId"`
	StartPrice    string    `bson:"startPrice" json:"startPrice"`
	HighestBid    string    `bson:"highestBid" json:"highestBid"`
	HighestBidder string    `bson:"highestBidder" json:"highestBidder"`
	StartTime     int64     `bson:"startTime" json:"startTime"`
	Duration      int64     `bson:"duration" json:"duration"`
	Ended         bool      `bson:"ended" json:"ended"`
	UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
}
