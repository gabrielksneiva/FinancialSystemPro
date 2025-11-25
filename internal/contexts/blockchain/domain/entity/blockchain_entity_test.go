package entity

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewBlockchainTransactionAndConfirm(t *testing.T) {
	amt := decimal.NewFromInt(42)
	tx := NewBlockchainTransaction(NetworkTron, "FROM", "TO", amt)
	if tx.Network != NetworkTron || tx.Amount.Cmp(amt) != 0 || tx.Status != "pending" {
		t.Fatalf("inicialização incorreta: %+v", tx)
	}
	tx.Confirm("HASH123", 100, 3)
	if tx.Status != "confirmed" || tx.TransactionHash != "HASH123" || tx.BlockNumber != 100 || tx.Confirmations != 3 || tx.ConfirmedAt == nil {
		t.Fatalf("confirmação incorreta: %+v", tx)
	}
}
