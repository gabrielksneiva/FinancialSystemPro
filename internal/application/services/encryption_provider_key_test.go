package services

import (
	"os"
	"testing"
)

func TestNewAESEncryptionProviderFromEnv_InvalidKeyFormat(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "short") // too short to be valid base64 32 bytes or hex
	p, err := NewAESEncryptionProviderFromEnv()
	if err == nil && p != nil {
		t.Fatalf("expected error for invalid key format")
	}
}

func TestDecodeKey_InvalidHexChar(t *testing.T) {
	// 64 chars but invalid letters 'z'
	bad := "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	if _, err := decodeKey(bad); err == nil {
		t.Fatalf("expected error for invalid hex")
	}
}
