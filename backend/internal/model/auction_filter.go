package model

// AuctionFilter holds optional query parameters for listing auctions.
type AuctionFilter struct {
	Q          string   // seller prefix, tokenId, or auction id
	Seller     string
	TokenID    string
	Bidder     string   // auctions this address has bid on (resolved to AuctionIDs)
	AuctionIDs []uint64 // set by handler when filtering by bidder
	Ended      *bool
}
