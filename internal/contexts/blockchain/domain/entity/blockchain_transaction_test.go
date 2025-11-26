package entity

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBlockchainTransactionFields(t *testing.T) {
	tx := BlockchainTransaction{
		FromAddress: "FROM",
		ToAddress:   "TO",
		Amount:      decimal.NewFromInt(1),
		Status:      "pending",
	}
	if tx.FromAddress != "FROM" || tx.ToAddress != "TO" || tx.Amount.Cmp(decimal.NewFromInt(1)) != 0 || tx.Status != "pending" {
		t.Fatalf("BlockchainTransaction fields incorretos: %+v", tx)
	}
}
