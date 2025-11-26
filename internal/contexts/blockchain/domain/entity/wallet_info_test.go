package entity

import (
	"testing"
)

func TestWalletInfoFields(t *testing.T) {
	w := WalletInfo{Address: "ADDR", EncryptedPrivKey: "PRIV"}
	if w.Address != "ADDR" || w.EncryptedPrivKey != "PRIV" {
		t.Fatalf("WalletInfo fields incorretos: %+v", w)
	}
}
