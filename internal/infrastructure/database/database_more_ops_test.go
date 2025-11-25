package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// cria DB simples para testes de Balance e WalletInfo updates
func makeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	// balance table
	if err := db.Exec("CREATE TABLE balances (id TEXT, created_at DATETIME, updated_at DATETIME, user_id TEXT, amount TEXT);").Error; err != nil {
		t.Fatalf("create balances: %v", err)
	}
	if err := db.Exec("CREATE TABLE wallet_infos (id TEXT, created_at DATETIME, updated_at DATETIME, tron_address TEXT, encrypted_priv_key TEXT, user_id TEXT);").Error; err != nil {
		t.Fatalf("create wallet_infos: %v", err)
	}
	return db
}

func TestTransaction_DepositCreatesBalance(t *testing.T) {
	gdb := makeTestDB(t)
	nd := &NewDatabase{DB: gdb, Logger: zap.NewNop()}
	uid := uuid.New()
	amt := decimal.NewFromInt(50)
	if err := nd.Transaction(uid, amt, "deposit"); err != nil {
		t.Fatalf("deposit erro: %v", err)
	}
	bal, err := nd.Balance(uid)
	if err != nil {
		t.Fatalf("balance erro: %v", err)
	}
	if !bal.Equal(amt) {
		t.Fatalf("esperado %s obtido %s", amt.String(), bal.String())
	}
}

func TestTransaction_DepositIncrementExisting(t *testing.T) {
	gdb := makeTestDB(t)
	nd := &NewDatabase{DB: gdb, Logger: zap.NewNop()}
	uid := uuid.New()
	first := decimal.NewFromInt(10)
	second := decimal.NewFromInt(5)
	if err := nd.Transaction(uid, first, "deposit"); err != nil {
		t.Fatalf("deposit1: %v", err)
	}
	if err := nd.Transaction(uid, second, "deposit"); err != nil {
		t.Fatalf("deposit2: %v", err)
	}
	bal, _ := nd.Balance(uid)
	if !bal.Equal(decimal.NewFromInt(15)) {
		t.Fatalf("esperado 15 obtido %s", bal.String())
	}
}

func TestTransaction_WithdrawReducesBalance(t *testing.T) {
	gdb := makeTestDB(t)
	nd := &NewDatabase{DB: gdb, Logger: zap.NewNop()}
	uid := uuid.New()
	if err := nd.Transaction(uid, decimal.NewFromInt(20), "deposit"); err != nil {
		t.Fatalf("deposit erro: %v", err)
	}
	if err := nd.Transaction(uid, decimal.NewFromInt(5), "withdraw"); err != nil {
		t.Fatalf("withdraw erro: %v", err)
	}
	bal, _ := nd.Balance(uid)
	if !bal.GreaterThanOrEqual(decimal.NewFromInt(15)) {
		t.Fatalf("esperado saldo >=15 obtido %s", bal.String())
	}
}

func TestUpdateWalletInfoAndByAddress(t *testing.T) {
	gdb := makeTestDB(t)
	nd := &NewDatabase{DB: gdb, Logger: zap.NewNop()}
	uid := uuid.New()
	if err := nd.SaveWalletInfo(uid, "TRONADDR1", "KEY1"); err != nil {
		t.Fatalf("save1: %v", err)
	}
	if err := nd.UpdateWalletInfo(uid, "TRONADDR2", "KEY2"); err != nil {
		t.Fatalf("update: %v", err)
	}
	w2, err := nd.GetWalletInfo(uid)
	if err != nil || w2.TronAddress != "TRONADDR2" {
		t.Fatalf("wallet não atualizada: %+v err=%v", w2, err)
	}
	wByAddr, err := nd.GetWalletInfoByAddress("TRONADDR2")
	if err != nil || wByAddr == nil || wByAddr.EncryptedPrivKey != "KEY2" {
		t.Fatalf("GetWalletInfoByAddress falhou")
	}
}
