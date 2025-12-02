package ethereum

import (
	"testing"

	"financial-system-pro/internal/shared/blockchain"
	testing2 "financial-system-pro/internal/shared/blockchain/testing"
)

// TestEthereumProviderConformance runs the full conformance test suite for Ethereum provider.
func TestEthereumProviderConformance(t *testing.T) {
	factory := func() blockchain.Provider {
		return NewProvider("https://eth-sepolia.g.alchemy.com/v2/demo", blockchain.NetworkTestnet)
	}

	testing2.RunConformanceTests(t, factory)
}

// TestEthereumProviderCapabilities tests Ethereum-specific capabilities.
func TestEthereumProviderCapabilities(t *testing.T) {
	provider := NewProvider("https://eth-sepolia.g.alchemy.com/v2/demo", blockchain.NetworkTestnet)
	testing2.RunCapabilityTests(t, provider)
}
