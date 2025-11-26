package services

import (
	"financial-system-pro/internal/domain/entities"
	"testing"

	"github.com/shopspring/decimal"
)

func TestDetermineCategory(t *testing.T) {
	cases := []struct {
		chain    entities.BlockchainType
		expected BlockchainCategory
	}{
		{entities.BlockchainTRON, CategoryAccount},
		{entities.BlockchainEthereum, CategoryEVM},
		{entities.BlockchainBitcoin, CategoryUTXO},
		{entities.BlockchainType("x"), BlockchainCategory("unknown")},
	}
	for _, c := range cases {
		got := DetermineCategory(c.chain)
		if got != c.expected {
			t.Fatalf("expected %s got %s for chain %s", c.expected, got, c.chain)
		}
	}
}

func TestConvertAmountToBaseUnit_Success(t *testing.T) {
	// TRON 1.23 TRX -> 1.23 * 1e6 SUN = 1230000
	sun, err := ConvertAmountToBaseUnit(entities.BlockchainTRON, decimal.NewFromFloat(1.23))
	if err != nil || sun != 1_230_000 {
		t.Fatalf("tron conversion failed: %v sun=%d", err, sun)
	}
	// BTC 0.5 BTC -> 0.5 * 1e8 = 50_000_000 sat
	sat, err := ConvertAmountToBaseUnit(entities.BlockchainBitcoin, decimal.NewFromFloat(0.5))
	if err != nil || sat != 50_000_000 {
		t.Fatalf("btc conversion failed: %v sat=%d", err, sat)
	}
}

func TestConvertAmountToBaseUnit_EthereumOverflow(t *testing.T) {
	// 100 ETH should overflow int64 when converted to wei (1e20)
	_, err := ConvertAmountToBaseUnit(entities.BlockchainEthereum, decimal.NewFromInt(100))
	if err == nil {
		t.Fatalf("expected overflow error, got nil")
	}
	if !contains(err.Error(), "valor muito grande") {
		t.Fatalf("expected overflow error message, got %v", err)
	}
}

func TestConvertAmountToBaseUnit_Unsupported(t *testing.T) {
	_, err := ConvertAmountToBaseUnit(entities.BlockchainType("unsupported"), decimal.NewFromInt(1))
	if err == nil || !contains(err.Error(), "unsupported blockchain") {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestVaultEnvKeys(t *testing.T) {
	a, p := VaultEnvKeys(entities.BlockchainEthereum)
	if a != "ETH_VAULT_ADDRESS" || p != "ETH_VAULT_PRIVATE_KEY" {
		t.Fatalf("ethereum keys mismatch")
	}
	a, p = VaultEnvKeys(entities.BlockchainBitcoin)
	if a != "BTC_VAULT_ADDRESS" || p != "BTC_VAULT_PRIVATE_KEY" {
		t.Fatalf("bitcoin keys mismatch")
	}
	a, p = VaultEnvKeys(entities.BlockchainTRON)
	if a != "" || p != "" {
		t.Fatalf("tron should not use env vault keys")
	}
}

// contains small helper to avoid importing strings for simple check
func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && (index(haystack, needle) >= 0)
}

// naive substring search
func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
