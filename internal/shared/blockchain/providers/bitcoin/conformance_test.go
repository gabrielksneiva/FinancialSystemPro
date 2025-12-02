package bitcoin

import (
	"testing"

	"financial-system-pro/internal/shared/blockchain"
	testing2 "financial-system-pro/internal/shared/blockchain/testing"
)

// TestBitcoinProviderConformance runs the full conformance test suite for Bitcoin provider.
func TestBitcoinProviderConformance(t *testing.T) {
	factory := func() blockchain.Provider {
		return NewProvider("https://blockstream.info/testnet/api", blockchain.NetworkTestnet)
	}

	testing2.RunConformanceTests(t, factory)
}

// TestBitcoinProviderCapabilities tests Bitcoin-specific capabilities.
func TestBitcoinProviderCapabilities(t *testing.T) {
	provider := NewProvider("https://blockstream.info/testnet/api", blockchain.NetworkTestnet)
	testing2.RunCapabilityTests(t, provider)
}
