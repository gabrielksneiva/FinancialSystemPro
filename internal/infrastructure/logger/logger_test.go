package logger

import (
	"os"
	"testing"
)

// TestProvideAndGetLogger garante criação e recuperação do logger global.
func TestProvideAndGetLogger(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	lg, err := ProvideLogger()
	if err != nil {
		t.Fatalf("erro ao criar logger: %v", err)
	}
	if lg == nil {
		t.Fatalf("logger retornou nil")
	}
	// GetLogger deve retornar mesma instância (ou não nil).
	got := GetLogger()
	if got == nil {
		t.Fatalf("GetLogger retornou nil")
	}
}
