package tron

import (
	"testing"

	"financial-system-pro/internal/shared/blockchain"
	testing2 "financial-system-pro/internal/shared/blockchain/testing"
)

// TestTronProviderConformance runs the full conformance test suite for Tron provider.
func TestTronProviderConformance(t *testing.T) {
	factory := func() blockchain.Provider {
		return NewProvider("https://api.shasta.trongrid.io", blockchain.NetworkTestnet)
	}

	testing2.RunConformanceTests(t, factory)
}

// TestTronProviderCapabilities tests Tron-specific capabilities.
func TestTronProviderCapabilities(t *testing.T) {
	provider := NewProvider("https://api.shasta.trongrid.io", blockchain.NetworkTestnet)
	testing2.RunCapabilityTests(t, provider)
}
