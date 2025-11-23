package services

import (
	"encoding/hex"
	"os"
	"testing"
)

// TestMain configures environment before tests
func TestMain(m *testing.M) {
	// Set encryption key for testing
	os.Setenv("ENCRYPTION_KEY", "HPico2FSovTxbVU1c/927+9W+Cef/mdyJDVIL1uZ17U=")
	code := m.Run()
	os.Exit(code)
}

func TestGenerateTronAddress(t *testing.T) {
	twm := NewTronWalletManager()

	// Public key de teste (65 bytes com prefixo 0x04)
	// Este é um exemplo público conhecido
	pubKeyHex := "0465d68d0a8b7e1f98b9a7d8e9f1c2a3b4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9"
	pubKeyBytes, _ := hex.DecodeString(pubKeyHex)

	address := twm.generateTronAddress(pubKeyBytes)

	t.Logf("Tamanho do endereço gerado: %d", len(address))
	t.Logf("Endereço: %s", address)

	if len(address) != 34 {
		t.Errorf("Esperado endereço com 34 caracteres, obteve %d: %s", len(address), address)
	}

	if address[0] != 'T' {
		t.Errorf("Endereço deveria começar com 'T', obteve: %c", address[0])
	}
}

func TestGenerateWallet(t *testing.T) {
	twm := NewTronWalletManager()

	wallet, err := twm.GenerateWallet()
	if err != nil {
		t.Fatalf("Erro ao gerar carteira: %v", err)
	}

	t.Logf("Endereço gerado: %s", wallet.Address)
	t.Logf("Tamanho: %d caracteres", len(wallet.Address))
	t.Logf("Public key length: %d", len(wallet.PublicKey))
	t.Logf("Private key length: %d", len(wallet.PrivateKey))

	if len(wallet.Address) != 34 {
		t.Errorf("Esperado endereço com 34 caracteres, obteve %d: %s", len(wallet.Address), wallet.Address)
	}

	if wallet.Address[0] != 'T' {
		t.Errorf("Endereço deveria começar com 'T', obteve: %s", wallet.Address)
	}
}
