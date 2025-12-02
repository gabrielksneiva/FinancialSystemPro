package finance

import "testing"

func TestWalletCreation(t *testing.T) {
	acc, _ := NewAccount("acc-w-1", "holder-x")
	amt, _ := NewAmount(0, "USD")
	w, err := NewWallet("wal-1", acc.ID(), amt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID() != "wal-1" || w.AccountID() != acc.ID() {
		t.Fatalf("wallet fields mismatch")
	}
	if w.Balance().Value() != 0 {
		t.Fatalf("expected zero balance")
	}
}

func TestWalletInvalid(t *testing.T) {
	amt, _ := NewAmount(0, "USD")
	_, err := NewWallet("", "acc", amt)
	if err == nil {
		t.Fatalf("expected error for empty id")
	}
	_, err = NewWallet("wal", "", amt)
	if err == nil {
		t.Fatalf("expected error for empty account id")
	}
}

func TestWalletCreditDebit(t *testing.T) {
	acc, _ := NewAccount("acc-w-2", "holder-y")
	zero, _ := NewAmount(0, "USD")
	w, _ := NewWallet("wal-2", acc.ID(), zero)
	credit, _ := NewAmount(100, "USD")
	if err := w.Credit(credit); err != nil {
		t.Fatalf("credit error: %v", err)
	}
	if w.Balance().Value() != 100 {
		t.Fatalf("expected balance 100")
	}
	debit, _ := NewAmount(40, "USD")
	if err := w.Debit(debit); err != nil {
		t.Fatalf("debit error: %v", err)
	}
	if w.Balance().Value() != 60 {
		t.Fatalf("expected balance 60")
	}
	// insufficient
	too, _ := NewAmount(100, "USD")
	if err := w.Debit(too); err == nil {
		t.Fatalf("expected insufficient funds error")
	}
}
