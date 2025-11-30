package domain

import (
	"testing"
)

func TestNewAddress(t *testing.T) {
	if _, err := NewAddress(""); err == nil {
		t.Fatalf("expected error for empty address")
	}
	if a, err := NewAddress("abcde"); err != nil || a == "" {
		t.Fatalf("expected valid address")
	}
}

func TestNewAmountFromString(t *testing.T) {
	if _, err := NewAmountFromString(""); err == nil {
		t.Fatalf("expected error for empty")
	}
	if a, err := NewAmountFromString("1000"); err != nil || a.Value.String() != "1000" {
		t.Fatalf("expected amount 1000")
	}
}

func TestNewAmountFromString_InvalidAndNegative(t *testing.T) {
	if _, err := NewAmountFromString("notanumber"); err == nil {
		t.Fatalf("expected error for non-numeric string")
	}
	if _, err := NewAmountFromString("-5"); err == nil {
		t.Fatalf("expected error for negative amount")
	}
}
