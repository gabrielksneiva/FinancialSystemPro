package domain

import (
	"testing"
)

func TestTransactionAndBlockBasics(t *testing.T) {
	blk := Block{Number: 10}
	if blk.Number != 10 {
		t.Fatalf("bad block")
	}

	tx := Transaction{Status: TxPending}
	if tx.Status != TxPending {
		t.Fatalf("expected pending")
	}
	// change status
	tx.Status = TxConfirmed
	if tx.Status != TxConfirmed {
		t.Fatalf("expected confirmed")
	}
}
