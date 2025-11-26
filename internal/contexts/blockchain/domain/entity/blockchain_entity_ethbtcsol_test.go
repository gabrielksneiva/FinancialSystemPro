package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestBlockchainTypeConstants(t *testing.T) {
	if BlockchainEthereum != "ethereum" || BlockchainBitcoin != "bitcoin" || BlockchainTron != "tron" || BlockchainSolana != "solana" {
		t.Fatalf("BlockchainType constants incorretos")
	}
}

func TestGeneratedWalletFields(t *testing.T) {
	uid := uuid.New()
	w := GeneratedWallet{
		Address:    "ADDR",
		PublicKey:  "PUB",
		PrivateKey: "PRIV",
		Blockchain: BlockchainSolana,
		CreatedAt:  1234567890,
		UserID:     uid,
	}
	if w.Address != "ADDR" || w.PublicKey != "PUB" || w.PrivateKey != "PRIV" || w.Blockchain != BlockchainSolana || w.CreatedAt != 1234567890 || w.UserID != uid {
		t.Fatalf("GeneratedWallet fields incorretos: %+v", w)
	}
}
