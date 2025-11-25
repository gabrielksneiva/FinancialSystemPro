package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repos "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func makeWalletDB(t *testing.T) *repos.NewDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.Exec("CREATE TABLE on_chain_wallets (id TEXT, created_at DATETIME, updated_at DATETIME, user_id TEXT, blockchain TEXT, address TEXT, public_key TEXT, encrypted_priv_key TEXT);").Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	return &repos.NewDatabase{DB: db}
}

func TestOnChainWalletRepository_SaveAndFind(t *testing.T) {
	db := makeWalletDB(t)
	repo := NewOnChainWalletRepositoryAdapter(db)
	uid := uuid.New()
	gw := &entities.GeneratedWallet{Address: "ADDR", PublicKey: "PUB", Blockchain: entities.BlockchainEthereum}
	if err := repo.Save(context.Background(), uid, gw, "ENC"); err != nil {
		t.Fatalf("save erro: %v", err)
	}
	found, err := repo.FindByUserAndChain(context.Background(), uid, entities.BlockchainEthereum)
	if err != nil || found == nil {
		t.Fatalf("find falhou: %v", err)
	}
	if found.Address != "ADDR" {
		t.Fatalf("address inesperado")
	}
	list, err := repo.ListByUser(context.Background(), uid)
	if err != nil || len(list) != 1 {
		t.Fatalf("list falhou: %v len=%d", err, len(list))
	}
	exists, err := repo.Exists(context.Background(), uid, entities.BlockchainEthereum)
	if err != nil || !exists {
		t.Fatalf("exists falhou: %v exists=%v", err, exists)
	}
}
