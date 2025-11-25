package container

import (
	"context"
	"financial-system-pro/internal/application/services"
	repos "financial-system-pro/internal/infrastructure/database"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// cria DB mínimo para on-chain wallet repository
func makeDB(t *testing.T) *repos.NewDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.Exec("CREATE TABLE on_chain_wallets (id TEXT, created_at DATETIME, updated_at DATETIME, user_id TEXT, blockchain TEXT, address TEXT, public_key TEXT, encrypted_priv_key TEXT);").Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	return &repos.NewDatabase{DB: db}
}

func TestProvideMultiChainWalletService_NoEncryption(t *testing.T) {
	os.Unsetenv("ENCRYPTION_MASTER_KEY")
	reg := services.NewBlockchainRegistry(services.NewEthereumService())
	repo := services.NewOnChainWalletRepositoryAdapter(makeDB(t))
	svc := ProvideMultiChainWalletService(reg, repo)
	if svc == nil {
		t.Fatalf("esperava serviço não nil")
	}
	// Gera e salva sem erro
	_, err := svc.GenerateAndPersist(context.Background(), uuid.New(), "ethereum")
	if err != nil {
		t.Fatalf("erro generate persist: %v", err)
	}
}

func TestProvideMultiChainWalletService_WithValidEncryption(t *testing.T) {
	// 32 bytes em hex (64 chars)
	os.Setenv("ENCRYPTION_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	reg := services.NewBlockchainRegistry(services.NewEthereumService())
	repo := services.NewOnChainWalletRepositoryAdapter(makeDB(t))
	svc := ProvideMultiChainWalletService(reg, repo)
	if svc == nil {
		t.Fatalf("esperava serviço não nil")
	}
	uid := uuid.New()
	w, err := svc.GenerateAndPersist(context.Background(), uid, "ethereum")
	if err != nil {
		t.Fatalf("erro generate persist: %v", err)
	}
	if w.Blockchain != "ethereum" {
		t.Fatalf("blockchain inesperada: %s", w.Blockchain)
	}
}
