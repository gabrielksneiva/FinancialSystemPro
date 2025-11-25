package services

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

// helper to generate a 32-byte random key and set env
func setMasterKey(t *testing.T) []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("rand error: %v", err)
	}
	os.Setenv("ENCRYPTION_MASTER_KEY", base64.StdEncoding.EncodeToString(key))
	return key
}

func TestAESEncryptionProvider_EncryptDecrypt_RoundTrip(t *testing.T) {
	setMasterKey(t)
	p, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("provider init failed: %v", err)
	}
	plain := "super-secret-data"
	enc, err := p.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if enc == plain {
		t.Fatalf("expected ciphertext different from plain text")
	}
	dec, err := p.Decrypt(enc)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if dec != plain {
		t.Fatalf("round trip mismatch: got %s", dec)
	}
}

func TestAESEncryptionProvider_Decrypt_InvalidBase64(t *testing.T) {
	setMasterKey(t)
	p, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("provider init failed: %v", err)
	}
	if _, err := p.Decrypt("!!!invalid!!!"); err == nil {
		t.Fatalf("expected error for invalid base64")
	}
}

func TestAESEncryptionProvider_Decrypt_ShortCiphertext(t *testing.T) {
	setMasterKey(t)
	p, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("provider init failed: %v", err)
	}
	// Nonce size for AES-GCM is typically 12; use shorter to trigger error
	if _, err := p.Decrypt(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Fatalf("expected error for short ciphertext")
	}
}

func TestAESEncryptionProvider_NoMasterKeyReturnsNoop(t *testing.T) {
	os.Unsetenv("ENCRYPTION_MASTER_KEY")
	p, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plain := "data"
	enc, err := p.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if enc != plain {
		t.Fatalf("expected noop encryption")
	}
	dec, err := p.Decrypt(enc)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if dec != plain {
		t.Fatalf("expected noop decryption")
	}
}
