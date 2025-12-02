package finance

import (
	"context"
	"testing"
)

func TestLedgerServiceCreditDebit(t *testing.T) {
	ctx := context.Background()
	acc, _ := NewAccount("acc-ledger-svc", "holder-ledger")
	zero, _ := NewAmount(0, "USD")
	w, _ := NewWallet("wal-ledger-svc", acc.ID(), zero)
	wRepo := newInMemoryWalletRepo(); _ = wRepo.Save(ctx, w)
	lRepo := newInMemoryLedgerRepo()
	svc := NewLedgerService(wRepo, lRepo)
	credAmt, _ := NewAmount(200, "USD")
	if err := svc.Credit(ctx, w.ID(), credAmt, "initial funding"); err != nil { t.Fatalf("credit error: %v", err) }
	wUpdated, _ := wRepo.Get(ctx, w.ID())
	if wUpdated.Balance().Value() != 200 { t.Fatalf("expected balance 200") }
	debAmt, _ := NewAmount(50, "USD")
	if err := svc.Debit(ctx, w.ID(), debAmt, "usage"); err != nil { t.Fatalf("debit error: %v", err) }
	wUpdated, _ = wRepo.Get(ctx, w.ID())
	if wUpdated.Balance().Value() != 150 { t.Fatalf("expected balance 150") }
}
