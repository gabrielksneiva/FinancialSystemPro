package persistence

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositories_CRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:blk_repo?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open err: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate err: %v", err)
	}

	ctx := context.Background()

	// blocks
	br := NewBlockRepository(db)
	bm := &BlockModel{Number: 123, Hash: "0xabc"}
	if err := br.Save(ctx, bm); err != nil {
		t.Fatalf("save block err: %v", err)
	}
	got, err := br.GetByNumber(ctx, 123)
	if err != nil || got.Hash != "0xabc" {
		t.Fatalf("get block err: %v got=%v", err, got)
	}

	// transactions
	tr := NewTransactionRepository(db)
	tx := &TransactionModel{ID: uuid.New().String(), Hash: "txh1", BlockNumber: 123}
	if err := tr.Save(ctx, tx); err != nil {
		t.Fatalf("save tx err: %v", err)
	}
	gtx, err := tr.GetByHash(ctx, "txh1")
	if err != nil || gtx.ID == "" {
		t.Fatalf("get tx err: %v got=%v", err, gtx)
	}

	// balances
	brc := NewBalanceRepository(db)
	if err := brc.Upsert(ctx, "addr1", "100"); err != nil {
		t.Fatalf("upsert bal err: %v", err)
	}
	bgot, err := brc.Get(ctx, "addr1")
	if err != nil || bgot.Balance != "100" {
		t.Fatalf("get bal err: %v got=%v", err, bgot)
	}
}
