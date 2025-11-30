package domain

import (
	"testing"
)

func TestNewTxHash(t *testing.T) {
	if _, err := NewTxHash(""); err == nil {
		t.Fatalf("expected error for empty hash")
	}
	if h, err := NewTxHash("0xabc"); err != nil || h == "" {
		t.Fatalf("expected valid hash")
	}
}
