package entity

import (
	"testing"
)

func TestBlockchainType_ValidValues(t *testing.T) {
	chains := []BlockchainType{BlockchainEthereum, BlockchainBitcoin, BlockchainTron, BlockchainSolana}
	for _, c := range chains {
		if c == "" {
			t.Fatalf("BlockchainType vazio: %+v", c)
		}
	}
}
