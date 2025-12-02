package solana

import (
	"testing"

	"financial-system-pro/internal/shared/blockchain"
	testing2 "financial-system-pro/internal/shared/blockchain/testing"
)

// TestSolanaProviderConformance runs the full conformance test suite for Solana provider.
func TestSolanaProviderConformance(t *testing.T) {
	factory := func() blockchain.Provider {
		return NewProvider("https://api.devnet.solana.com", blockchain.NetworkTestnet)
	}

	testing2.RunConformanceTests(t, factory)
}

// TestSolanaProviderCapabilities tests Solana-specific capabilities.
func TestSolanaProviderCapabilities(t *testing.T) {
	provider := NewProvider("https://api.devnet.solana.com", blockchain.NetworkTestnet)
	testing2.RunCapabilityTests(t, provider)
}
