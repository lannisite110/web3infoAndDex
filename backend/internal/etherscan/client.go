package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const sepoliaAPIBase = "https://api-sepolia.etherscan.io/api"

// Client calls Etherscan Sepolia APIs.
type Client struct {
	apiKey string
	http   *http.Client
}

// NewClient creates an Etherscan client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

// TxReceipt is a simplified transaction receipt for the REST API.
type TxReceipt struct {
	TxHash        string `json:"txHash"`
	Status        string `json:"status"`
	BlockNumber   string `json:"blockNumber"`
	From          string `json:"from"`
	To            string `json:"to"`
	GasUsed       string `json:"gasUsed"`
	EtherscanURL  string `json:"etherscanUrl"`
}

type apiResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type receiptResult struct {
	Status      string `json:"status"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	GasUsed     string `json:"gasUsed"`
}

// GetTransactionReceipt fetches a receipt via Etherscan proxy API.
func (c *Client) GetTransactionReceipt(ctx context.Context, txHash string) (*TxReceipt, error) {
	hash := strings.ToLower(strings.TrimSpace(txHash))
	params := url.Values{}
	params.Set("module", "proxy")
	params.Set("action", "eth_getTransactionReceipt")
	params.Set("txhash", hash)
	params.Set("apikey", c.apiKey)

	reqURL := sepoliaAPIBase + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("etherscan decode: %w", err)
	}

	if envelope.Status != "1" && envelope.Message != "" {
		return nil, fmt.Errorf("etherscan: %s", envelope.Message)
	}

	if string(envelope.Result) == "null" || len(envelope.Result) == 0 {
		return nil, fmt.Errorf("transaction not found")
	}

	var rc receiptResult
	if err := json.Unmarshal(envelope.Result, &rc); err != nil {
		return nil, fmt.Errorf("etherscan receipt: %w", err)
	}

	status := rc.Status
	if status == "0x1" {
		status = "success"
	} else if status == "0x0" {
		status = "failed"
	}

	return &TxReceipt{
		TxHash:       hash,
		Status:       status,
		BlockNumber:  rc.BlockNumber,
		From:         strings.ToLower(rc.From),
		To:           strings.ToLower(rc.To),
		GasUsed:      rc.GasUsed,
		EtherscanURL: "https://sepolia.etherscan.io/tx/" + hash,
	}, nil
}
