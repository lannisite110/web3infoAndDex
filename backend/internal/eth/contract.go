package eth

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lannisite110/web3infoanddex/backend/internal/model"
)

//go:embed NFTAuction.abi.json
var nftAuctionABIJSON []byte

// Client wraps an Ethereum JSON-RPC connection and parsed NFTAuction ABI.
type Client struct {
	ethclient *ethclient.Client
	abi       abi.ABI
}

// NewClient dials Sepolia (or any EVM) RPC.
func NewClient(rpcURL string) (*Client, error) {
	parsed, err := abi.JSON(bytes.NewReader(nftAuctionABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse abi: %w", err)
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("eth dial: %w", err)
	}

	return &Client{ethclient: client, abi: parsed}, nil
}

// Close closes the RPC connection.
func (c *Client) Close() {
	if c != nil && c.ethclient != nil {
		c.ethclient.Close()
	}
}

// Raw returns the underlying ethclient for block queries and filters.
func (c *Client) Raw() *ethclient.Client {
	return c.ethclient
}

// ParsedABI returns the contract ABI.
func (c *Client) ParsedABI() abi.ABI {
	return c.abi
}

type auctionTuple struct {
	Seller        common.Address
	NftContract   common.Address
	TokenId       *big.Int
	StartPrice    *big.Int
	StartTime     *big.Int
	Duration      *big.Int
	HighestBidder common.Address
	HighestBid    *big.Int
	Ended         bool
}

// AuctionCount returns the on-chain auction counter.
func (c *Client) AuctionCount(ctx context.Context, contract common.Address) (uint64, error) {
	data, err := c.abi.Pack("auctionCount")
	if err != nil {
		return 0, err
	}
	out, err := c.call(ctx, contract, data)
	if err != nil {
		return 0, err
	}
	var count *big.Int
	if err := c.abi.UnpackIntoInterface(&count, "auctionCount", out); err != nil {
		return 0, err
	}
	return count.Uint64(), nil
}

// FetchAuction reads getAuction(id) and maps to the MongoDB model.
func (c *Client) FetchAuction(
	ctx context.Context,
	contract common.Address,
	chainID int64,
	auctionID uint64,
) (model.Auction, error) {
	data, err := c.abi.Pack("getAuction", new(big.Int).SetUint64(auctionID))
	if err != nil {
		return model.Auction{}, err
	}
	out, err := c.call(ctx, contract, data)
	if err != nil {
		return model.Auction{}, err
	}

	values, err := c.abi.Unpack("getAuction", out)
	if err != nil || len(values) != 1 {
		return model.Auction{}, fmt.Errorf("unpack getAuction: %w", err)
	}

	converted := abi.ConvertType(values[0], new(auctionTuple))
	tuple, ok := converted.(*auctionTuple)
	if !ok {
		return model.Auction{}, fmt.Errorf("unexpected auction tuple type %T", converted)
	}

	return model.Auction{
		ChainID:       chainID,
		Contract:      strings.ToLower(contract.Hex()),
		AuctionID:     auctionID,
		Seller:        strings.ToLower(tuple.Seller.Hex()),
		NFTContract:   strings.ToLower(tuple.NftContract.Hex()),
		TokenID:       tuple.TokenId.String(),
		StartPrice:    tuple.StartPrice.String(),
		HighestBid:    tuple.HighestBid.String(),
		HighestBidder: strings.ToLower(tuple.HighestBidder.Hex()),
		StartTime:     tuple.StartTime.Int64(),
		Duration:      tuple.Duration.Int64(),
		Ended:         tuple.Ended,
	}, nil
}

func (c *Client) call(ctx context.Context, contract common.Address, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{To: &contract, Data: data}
	return c.ethclient.CallContract(ctx, msg, nil)
}
