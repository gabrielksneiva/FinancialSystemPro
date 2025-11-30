package service

import (
	"context"
	"testing"

	bc "financial-system-pro/internal/blockchain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	appsvc "financial-system-pro/internal/application/services"
	"financial-system-pro/internal/contexts/blockchain/infrastructure/persistence"
)

// mock connector
type fakeConnector struct{ blocks []bc.Block }

func (f *fakeConnector) FetchBalance(ctx context.Context, address bc.Address) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (f *fakeConnector) SendTransaction(ctx context.Context, rawTx string) (bc.TxHash, error) {
	return "", nil
}
func (f *fakeConnector) GetTransactionStatus(ctx context.Context, hash bc.TxHash) (string, error) {
	return "", nil
}
func (f *fakeConnector) SyncLatestBlocks(ctx context.Context, since uint64) ([]bc.Block, error) {
	return f.blocks, nil
}

// mock outbox
type mockOutbox struct{ saved int }

func (m *mockOutbox) Save(ctx context.Context, rec *appsvc.OutboxRecord) error          { m.saved++; return nil }
func (m *mockOutbox) MarkPublished(ctx context.Context, id uuid.UUID) error             { return nil }
func (m *mockOutbox) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error { return nil }
func (m *mockOutbox) ListPending(ctx context.Context, limit int) ([]*appsvc.OutboxRecord, error) {
	return nil, nil
}

func TestSyncService_PersistBlocksAndCreateOutbox(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:syncsvc?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db err: %v", err)
	}
	if err := persistence.AutoMigrate(db); err != nil {
		t.Fatalf("migrate err: %v", err)
	}

	br := persistence.NewBlockRepository(db)
	mock := &mockOutbox{}
	conn := &fakeConnector{blocks: []bc.Block{{Number: 1, Hash: "h1"}, {Number: 2, Hash: "h2"}}}
	svc := NewSyncService(conn, br, mock)
	n, err := svc.SyncBlocks(context.Background(), 0)
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if n != 2 || mock.saved != 2 {
		t.Fatalf("expected 2 persisted+outbox got n=%d saved=%d", n, mock.saved)
	}

	// verify payload stored in outbox save was JSON by re-creating payload
	// (mock only counted saves); verify blocks exist in DB
	b1, err := br.GetByNumber(context.Background(), 1)
	if err != nil || b1.Hash != "h1" {
		t.Fatalf("block1 missing: %v %v", err, b1)
	}

	// re-run sync should not create duplicates
	n2, _ := svc.SyncBlocks(context.Background(), 0)
	if n2 != 0 {
		t.Fatalf("expected 0 new on second sync, got %d", n2)
	}
}
