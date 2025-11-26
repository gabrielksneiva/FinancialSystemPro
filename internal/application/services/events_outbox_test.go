package services

import (
	"context"
	repos "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/events"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockBus struct{ published int }

func (m *mockBus) PublishAsync(ctx context.Context, event events.Event) { m.published++ }

func TestGormOutboxAdapter_CRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:outbox_unit?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db err: %v", err)
	}
	nd := &repos.NewDatabase{DB: db, Logger: zap.NewNop()}
	// create table
	ddl := `CREATE TABLE outbox (
        id TEXT PRIMARY KEY,
        aggregate TEXT,
        type TEXT,
        payload BLOB,
        created_at DATETIME,
        published BOOLEAN,
        published_at DATETIME,
        attempts INTEGER,
        last_error TEXT
    )`
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatalf("ddl err: %v", err)
	}

	adapter := NewGormOutboxAdapter(nd)
	rec := &OutboxRecord{ID: uuid.New(), Aggregate: "transaction", Type: "withdraw.completed", Payload: []byte("{}")}
	if err := adapter.Save(context.Background(), rec); err != nil {
		t.Fatalf("save err: %v", err)
	}
	list, err := adapter.ListPending(context.Background(), 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("list err: %v len=%d", err, len(list))
	}
	if err := adapter.MarkFailed(context.Background(), rec.ID, "network"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	if err := adapter.MarkPublished(context.Background(), rec.ID); err != nil {
		t.Fatalf("mark published: %v", err)
	}
}

func TestEventsDispatcher_DispatchOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:outbox_dispatch?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db err: %v", err)
	}
	nd := &repos.NewDatabase{DB: db, Logger: zap.NewNop()}
	if err := db.Exec(`CREATE TABLE outbox (id TEXT PRIMARY KEY, aggregate TEXT, type TEXT, payload BLOB, created_at DATETIME, published BOOLEAN, published_at DATETIME, attempts INTEGER, last_error TEXT)`).Error; err != nil {
		t.Fatalf("ddl err: %v", err)
	}
	adapter := NewGormOutboxAdapter(nd)
	bus := &mockBus{}
	dispatcher := NewEventsDispatcher(adapter, bus)
	// seed two records
	_ = adapter.Save(context.Background(), &OutboxRecord{ID: uuid.New(), Aggregate: "transaction", Type: "deposit.completed", Payload: []byte("{}")})
	_ = adapter.Save(context.Background(), &OutboxRecord{ID: uuid.New(), Aggregate: "transaction", Type: "withdraw.completed", Payload: []byte("{}")})
	sent, err := dispatcher.DispatchOnce(context.Background(), 10)
	if err != nil {
		t.Fatalf("dispatch err: %v", err)
	}
	if sent != 2 || bus.published != 2 {
		t.Fatalf("expected 2 sent, got %d published %d", sent, bus.published)
	}
}
