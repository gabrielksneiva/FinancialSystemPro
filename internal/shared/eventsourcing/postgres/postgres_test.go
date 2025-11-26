package postgres

import (
	"context"
	"database/sql"
	"financial-system-pro/internal/shared/eventsourcing"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock domain event for testing
type testEvent struct {
	id            string
	eventType     string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
	version       int
	data          map[string]interface{}
}

func (e *testEvent) EventID() string       { return e.id }
func (e *testEvent) EventType() string     { return e.eventType }
func (e *testEvent) AggregateID() string   { return e.aggregateID }
func (e *testEvent) AggregateType() string { return e.aggregateType }
func (e *testEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *testEvent) Version() int          { return e.version }
func (e *testEvent) Data() interface{}     { return e.data }

func newTestEvent(aggID string) *testEvent {
	return &testEvent{
		id:            uuid.New().String(),
		eventType:     "TestEvent",
		aggregateID:   aggID,
		aggregateType: "TestAggregate",
		occurredAt:    time.Now(),
		version:       1,
		data:          map[string]interface{}{"test": "data"},
	}
}

func TestNewPostgresEventStore(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewPostgresEventStore(db)

	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

func TestPostgresEventStore_GetVersion_NoEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("agg-123").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(0))

	store := NewPostgresEventStore(db)
	version, err := store.GetVersion(context.Background(), "agg-123")

	assert.NoError(t, err)
	assert.Equal(t, 0, version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresEventStore_GetVersion_WithEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("agg-456").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(5))

	store := NewPostgresEventStore(db)
	version, err := store.GetVersion(context.Background(), "agg-456")

	assert.NoError(t, err)
	assert.Equal(t, 5, version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresEventStore_LoadEvents_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT (.+) FROM event_store").
		WithArgs("agg-999", 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "aggregate_id", "aggregate_type", "event_type",
			"event_data", "metadata", "version", "occurred_at", "created_at",
		}))

	store := NewPostgresEventStore(db)
	events, err := store.LoadEvents(context.Background(), "agg-999")

	assert.Error(t, err)
	assert.Equal(t, eventsourcing.ErrEventNotFound, err)
	assert.Nil(t, events)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresEventStore_LoadEventsFrom(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "aggregate_id", "aggregate_type", "event_type",
		"event_data", "metadata", "version", "occurred_at", "created_at",
	}).AddRow(
		id, "agg-123", "Order", "OrderCreated",
		[]byte(`{"order_id":"123"}`), []byte(`{"user":"admin"}`), 3, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM event_store").
		WithArgs("agg-123", 2).
		WillReturnRows(rows)

	store := NewPostgresEventStore(db)
	events, err := store.LoadEventsFrom(context.Background(), "agg-123", 2)

	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, id, events[0].ID)
	assert.Equal(t, "agg-123", events[0].AggregateID)
	assert.Equal(t, 3, events[0].Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresEventStore_LoadEventsByType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "aggregate_id", "aggregate_type", "event_type",
		"event_data", "metadata", "version", "occurred_at", "created_at",
	}).AddRow(
		id, "agg-456", "Payment", "PaymentProcessed",
		[]byte(`{"amount":100}`), []byte(`{}`), 1, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM event_store WHERE event_type").
		WithArgs("PaymentProcessed", 10).
		WillReturnRows(rows)

	store := NewPostgresEventStore(db)
	events, err := store.LoadEventsByType(context.Background(), "PaymentProcessed", 10)

	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "PaymentProcessed", events[0].EventType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresEventStore_SaveEvents_ConcurrencyError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("agg-123").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(5))
	mock.ExpectRollback()

	store := NewPostgresEventStore(db)
	event := newTestEvent("agg-123")
	err = store.SaveEvents(context.Background(), "agg-123", []eventsourcing.DomainEvent{event}, 3)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSnapshotStore_SaveSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	snapshot := &eventsourcing.Snapshot{
		AggregateID:   "agg-789",
		AggregateType: "Account",
		Version:       10,
		State:         []byte(`{"balance":1000}`),
	}

	mock.ExpectExec("INSERT INTO aggregate_snapshots").
		WithArgs(snapshot.AggregateID, snapshot.AggregateType, snapshot.Version, snapshot.State, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	store := NewPostgresSnapshotStore(db)
	err = store.SaveSnapshot(context.Background(), snapshot)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSnapshotStore_LoadSnapshot_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"aggregate_id", "aggregate_type", "version", "state", "created_at",
	}).AddRow("agg-111", "Order", 15, []byte(`{"items":[]}`), now)

	mock.ExpectQuery("SELECT (.+) FROM aggregate_snapshots").
		WithArgs("agg-111").
		WillReturnRows(rows)

	store := NewPostgresSnapshotStore(db)
	snapshot, err := store.LoadSnapshot(context.Background(), "agg-111")

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, "agg-111", snapshot.AggregateID)
	assert.Equal(t, 15, snapshot.Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSnapshotStore_LoadSnapshot_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT (.+) FROM aggregate_snapshots").
		WithArgs("agg-222").
		WillReturnError(sql.ErrNoRows)

	store := NewPostgresSnapshotStore(db)
	snapshot, err := store.LoadSnapshot(context.Background(), "agg-222")

	assert.Error(t, err)
	assert.Equal(t, eventsourcing.ErrSnapshotNotFound, err)
	assert.Nil(t, snapshot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSnapshotStore_DeleteSnapshots(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM aggregate_snapshots").
		WithArgs("agg-333", 10).
		WillReturnResult(sqlmock.NewResult(0, 3))

	store := NewPostgresSnapshotStore(db)
	err = store.DeleteSnapshots(context.Background(), "agg-333", 10)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewPostgresSnapshotStore(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewPostgresSnapshotStore(db)

	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}
