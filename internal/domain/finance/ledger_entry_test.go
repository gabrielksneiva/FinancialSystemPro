package finance

import "testing"

func TestLedgerEntryCreation(t *testing.T) {
	amt, _ := NewAmount(50, "USD")
	le, err := NewLedgerEntry("le-1", "wal-x", EntryTypeCredit, amt, 50, "deposit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if le.ID() != "le-1" || le.WalletID() != "wal-x" {
		t.Fatalf("ledger entry fields mismatch")
	}
	if le.Type() != EntryTypeCredit {
		t.Fatalf("expected credit type")
	}
	if le.BalanceAfter() != 50 {
		t.Fatalf("expected balance after 50")
	}
}

func TestLedgerEntryValidation(t *testing.T) {
	amt, _ := NewAmount(10, "USD")
	_, err := NewLedgerEntry("", "wal", EntryTypeCredit, amt, 10, "x")
	if err == nil {
		t.Fatalf("expected error for empty id")
	}
	_, err = NewLedgerEntry("id", "", EntryTypeCredit, amt, 10, "x")
	if err == nil {
		t.Fatalf("expected error for empty wallet id")
	}
	_, err = NewLedgerEntry("id2", "wal", "invalid", amt, 10, "x")
	if err == nil {
		t.Fatalf("expected error for invalid type")
	}
}
