package entity

import (
	"testing"
)

func TestBlockchainNetworkConstants(t *testing.T) {
	if NetworkTron != "TRON" {
		t.Fatalf("BlockchainNetwork constants incorretos")
	}
}
