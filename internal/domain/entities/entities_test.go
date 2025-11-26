package entities

import "testing"

func TestWalletInfoStruct(t *testing.T) {
	w := WalletInfo{Address: "ADDR", PublicKey: "PUB", Blockchain: BlockchainTRON}
	if w.Blockchain != BlockchainTRON {
		t.Fatalf("blockchain inesperado")
	}
	if w.Address == "" || w.PublicKey == "" {
		t.Fatalf("campos vazios")
	}
}

func TestTronTransactionRequestTags(t *testing.T) {
	req := TronTransactionRequest{Amount: 10}
	if req.Amount <= 0 {
		t.Fatalf("amount inválido")
	}
}
