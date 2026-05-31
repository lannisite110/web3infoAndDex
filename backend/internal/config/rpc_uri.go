package config

import (
	"fmt"
	"strings"
)

// ValidateRPCURL ensures SEPOLIA_RPC_URL is an HTTP(S) endpoint, not a contract address.
func ValidateRPCURL(url string) error {
	if url == "" {
		return nil
	}
	lower := strings.ToLower(url)
	if strings.HasPrefix(lower, "0x") {
		return fmt.Errorf("SEPOLIA_RPC_URL looks like a contract address (%s); use your Alchemy HTTPS URL instead", url)
	}
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "ws://") && !strings.HasPrefix(lower, "wss://") {
		return fmt.Errorf("SEPOLIA_RPC_URL must start with https:// (got %q)", url)
	}
	return nil
}
