package container

import (
	"financial-system-pro/internal/application/services"
	"os"
	"testing"
)

func TestWalletManager_NoEncryption(t *testing.T) {
	os.Unsetenv("ENCRYPTION_MASTER_KEY")
	wm := ProvideWalletManager()
	if wm == nil {
		t.Fatalf("esperava wallet manager não nil")
	}
}

func TestEncryptionProvider_FromEnv(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	prov, err := services.NewAESEncryptionProviderFromEnv()
	if err != nil || prov == nil {
		t.Fatalf("esperava provider válido, err=%v", err)
	}
}
