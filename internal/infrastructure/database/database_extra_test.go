package repositories

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMinFunction cobre helper min.
func TestMinFunction(t *testing.T) {
	if min(2, 5) != 2 {
		t.Fatalf("min incorreto")
	}
	if min(10, 3) != 3 {
		t.Fatalf("min incorreto")
	}
}

// TestSaveAndGetWalletInfo usa sqlite in-memory para cobrir caminhos básicos.
func TestSaveAndGetWalletInfo(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao abrir sqlite: %v", err)
	}
	// Criar tabela simplificada evitando defaults específicos de Postgres.
	if err := gdb.Exec("CREATE TABLE wallet_infos (id TEXT, created_at DATETIME, updated_at DATETIME, tron_address TEXT, encrypted_priv_key TEXT, user_id TEXT);").Error; err != nil {
		t.Fatalf("falha ao criar tabela manual: %v", err)
	}
	nd := &NewDatabase{DB: gdb, Logger: zap.NewNop()}
	uid := uuid.New()
	if err := nd.SaveWalletInfo(uid, "TADDR1234567890", "ENCRYPTEDKEY"); err != nil {
		t.Fatalf("SaveWalletInfo falhou: %v", err)
	}
	w, err := nd.GetWalletInfo(uid)
	if err != nil {
		t.Fatalf("GetWalletInfo falhou: %v", err)
	}
	if w == nil || w.EncryptedPrivKey != "ENCRYPTEDKEY" {
		t.Fatalf("dados inconsistentes: %+v", w)
	}
}
