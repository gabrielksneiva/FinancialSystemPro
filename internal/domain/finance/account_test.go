package finance

import (
	"testing"
)

func TestAccountCreation(t *testing.T) {
	acc, err := NewAccount("acc-1", "holder-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.ID() != "acc-1" || acc.HolderID() != "holder-1" {
		t.Fatalf("account fields mismatch")
	}
	if acc.Status() != AccountStatusActive {
		t.Fatalf("expected active status")
	}
}

func TestAccountInvalid(t *testing.T) {
	_, err := NewAccount("", "holder")
	if err == nil {
		t.Fatalf("expected error for empty id")
	}
	_, err = NewAccount("id", "")
	if err == nil {
		t.Fatalf("expected error for empty holder")
	}
}

func TestAccountDeactivateActivate(t *testing.T) {
	acc, _ := NewAccount("acc-2", "holder-2")
	if err := acc.Deactivate(); err != nil {
		t.Fatalf("deactivate error: %v", err)
	}
	if acc.Status() != AccountStatusInactive {
		t.Fatalf("expected inactive")
	}
	if err := acc.Activate(); err != nil {
		t.Fatalf("activate error: %v", err)
	}
	if acc.Status() != AccountStatusActive {
		t.Fatalf("expected active again")
	}
}
