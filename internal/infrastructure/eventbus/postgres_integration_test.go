package eventbus

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupEventStoreDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/financial_system_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for integration tests")
	}

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not available for integration tests")
	}

	// Clean tables
	_, _ = db.Exec("TRUNCATE event_store, event_processing_log, event_subscriptions CASCADE")

	return db
}

func TestPostgresStoreSaveAndRetrieve(t *testing.T) {
	db := setupEventStoreDB(t)
	defer db.Close()

	store := NewPostgresStore(db)
	ctx := context.Background()

	event, err := NewBaseEvent("test.created", "agg-pg-1", "test", map[string]string{"key": "value"})
	require.NoError(t, err)

	// Save
	err = store.SaveEvent(ctx, event)
	require.NoError(t, err)

	// Duplicate should fail
	err = store.SaveEvent(ctx, event)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate event")

	// Get by ID
	retrieved, err := store.GetEventByID(ctx, event.EventID())
	require.NoError(t, err)
	require.Equal(t, event.EventID(), retrieved.EventID())
	require.Equal(t, event.EventType(), retrieved.EventType())

	// Get by aggregate
	events, err := store.GetEvents(ctx, "agg-pg-1", 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
}

func TestPostgresStoreGetEventsSince(t *testing.T) {
	db := setupEventStoreDB(t)
	defer db.Close()

	store := NewPostgresStore(db)
	ctx := context.Background()

	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)

	// Create event in the past
	oldEvent, _ := NewBaseEvent("old.event", "agg-pg-2", "test", map[string]string{})
	oldEvent.Occurred = past
	_ = store.SaveEvent(ctx, oldEvent)

	// Create recent event
	newEvent, _ := NewBaseEvent("new.event", "agg-pg-3", "test", map[string]string{})
	_ = store.SaveEvent(ctx, newEvent)

	// Get events since 30 minutes ago
	since := now.Add(-30 * time.Minute)
	events, err := store.GetEventsSince(ctx, since)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "new.event", events[0].EventType())
}

func TestPostgresProcessingLogFlow(t *testing.T) {
	db := setupEventStoreDB(t)
	defer db.Close()

	procLog := NewPostgresProcessingLog(db)
	ctx := context.Background()

	// Not processed initially
	processed, err := procLog.IsProcessed(ctx, "evt-pg-1", "handler-pg-1")
	require.NoError(t, err)
	require.False(t, processed)

	// Mark processing
	err = procLog.MarkProcessing(ctx, "evt-pg-1", "handler-pg-1")
	require.NoError(t, err)

	// Still not completed
	processed, err = procLog.IsProcessed(ctx, "evt-pg-1", "handler-pg-1")
	require.NoError(t, err)
	require.False(t, processed)

	// Mark completed
	err = procLog.MarkCompleted(ctx, "evt-pg-1", "handler-pg-1")
	require.NoError(t, err)

	// Now processed
	processed, err = procLog.IsProcessed(ctx, "evt-pg-1", "handler-pg-1")
	require.NoError(t, err)
	require.True(t, processed)
}

func TestResilientBusWithPostgresStore(t *testing.T) {
	db := setupEventStoreDB(t)
	defer db.Close()

	store := NewPostgresStore(db)
	procLog := NewPostgresProcessingLog(db)
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()
	var handledCount int

	handler := func(ctx context.Context, event Event) error {
		handledCount++
		return nil
	}

	err := bus.Subscribe("postgres.test", "pg-handler", handler)
	require.NoError(t, err)

	event, _ := NewBaseEvent("postgres.test", "agg-pg-bus", "test", map[string]string{"foo": "bar"})
	err = bus.Publish(ctx, event)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 1, handledCount)

	// Verify persistence
	retrieved, err := store.GetEventByID(ctx, event.EventID())
	require.NoError(t, err)
	require.Equal(t, event.EventID(), retrieved.EventID())

	// Verify idempotency tracking
	processed, err := procLog.IsProcessed(ctx, event.EventID(), "pg-handler")
	require.NoError(t, err)
	require.True(t, processed)
}

func TestResilientBusReplayWithPostgres(t *testing.T) {
	db := setupEventStoreDB(t)
	defer db.Close()

	store := NewPostgresStore(db)
	procLog := NewPostgresProcessingLog(db)
	logger := zap.NewNop()

	ctx := context.Background()

	// Store events directly
	for i := 1; i <= 3; i++ {
		event, _ := NewBaseEvent("replay.test", "agg-replay-pg", "test", map[string]int{"seq": i})
		event.Ver = i
		_ = store.SaveEvent(ctx, event)
	}

	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	var replayedCount int

	handler := func(ctx context.Context, event Event) error {
		replayedCount++
		return nil
	}

	err := bus.Subscribe("replay.test", "replay-pg-handler", handler)
	require.NoError(t, err)

	// Replay from version 2
	err = bus.Replay(ctx, "agg-replay-pg", 2)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	require.Equal(t, 2, replayedCount, "should replay 2 events (version 2 and 3)")
}
