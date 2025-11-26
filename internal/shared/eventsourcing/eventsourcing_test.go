package eventsourcing

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock domain event for testing
type mockDomainEvent struct {
	id            string
	eventType     string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
	version       int
	data          map[string]interface{}
}

func (m *mockDomainEvent) EventID() string       { return m.id }
func (m *mockDomainEvent) EventType() string     { return m.eventType }
func (m *mockDomainEvent) AggregateID() string   { return m.aggregateID }
func (m *mockDomainEvent) AggregateType() string { return m.aggregateType }
func (m *mockDomainEvent) OccurredAt() time.Time { return m.occurredAt }
func (m *mockDomainEvent) Version() int          { return m.version }
func (m *mockDomainEvent) Data() interface{}     { return m.data }

func newMockEvent(aggID, eventType string, data map[string]interface{}) *mockDomainEvent {
	return &mockDomainEvent{
		id:            uuid.New().String(),
		eventType:     eventType,
		aggregateID:   aggID,
		aggregateType: "TestAggregate",
		occurredAt:    time.Now(),
		version:       1,
		data:          data,
	}
}

func TestEventEnvelope_Structure(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	envelope := EventEnvelope{
		ID:            id,
		AggregateID:   "agg-123",
		AggregateType: "Order",
		EventType:     "OrderCreated",
		EventData:     []byte(`{"order_id":"123"}`),
		Metadata:      map[string]interface{}{"user": "admin"},
		Version:       1,
		OccurredAt:    now,
		CreatedAt:     now,
	}

	assert.Equal(t, id, envelope.ID)
	assert.Equal(t, "agg-123", envelope.AggregateID)
	assert.Equal(t, "Order", envelope.AggregateType)
	assert.Equal(t, "OrderCreated", envelope.EventType)
	assert.NotEmpty(t, envelope.EventData)
	assert.Equal(t, 1, envelope.Version)
}

func TestSnapshot_Structure(t *testing.T) {
	now := time.Now()

	snapshot := Snapshot{
		AggregateID:   "agg-456",
		AggregateType: "Account",
		Version:       10,
		State:         []byte(`{"balance":1000}`),
		CreatedAt:     now,
	}

	assert.Equal(t, "agg-456", snapshot.AggregateID)
	assert.Equal(t, "Account", snapshot.AggregateType)
	assert.Equal(t, 10, snapshot.Version)
	assert.NotEmpty(t, snapshot.State)
	assert.Equal(t, now, snapshot.CreatedAt)
}

func TestConcurrencyError_Message(t *testing.T) {
	err := NewConcurrencyError("agg-789", 5, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "concurrency conflict")
}

func TestConcurrencyError_Fields(t *testing.T) {
	err := &ConcurrencyError{
		AggregateID:     "agg-999",
		ExpectedVersion: 3,
		ActualVersion:   5,
	}

	assert.Equal(t, "agg-999", err.AggregateID)
	assert.Equal(t, 3, err.ExpectedVersion)
	assert.Equal(t, 5, err.ActualVersion)
}

func TestEventNotFoundError(t *testing.T) {
	assert.Equal(t, "no events found for aggregate", EventNotFoundError.Error())
	assert.True(t, errors.Is(EventNotFoundError, EventNotFoundError))
}

func TestSnapshotNotFoundError(t *testing.T) {
	assert.Equal(t, "no snapshot found for aggregate", SnapshotNotFoundError.Error())
	assert.True(t, errors.Is(SnapshotNotFoundError, SnapshotNotFoundError))
}

func TestMockDomainEvent_Implementation(t *testing.T) {
	data := map[string]interface{}{"amount": 100}
	event := newMockEvent("agg-001", "PaymentProcessed", data)

	assert.NotEmpty(t, event.EventID())
	assert.Equal(t, "PaymentProcessed", event.EventType())
	assert.Equal(t, "agg-001", event.AggregateID())
	assert.Equal(t, "TestAggregate", event.AggregateType())
	assert.Equal(t, 1, event.Version())
	assert.NotNil(t, event.OccurredAt())
	assert.Equal(t, data, event.Data())
}

func TestEventEnvelope_MetadataHandling(t *testing.T) {
	envelope := EventEnvelope{
		Metadata: map[string]interface{}{
			"correlation_id": "corr-123",
			"causation_id":   "cause-456",
			"user_id":        "user-789",
		},
	}

	assert.Len(t, envelope.Metadata, 3)
	assert.Equal(t, "corr-123", envelope.Metadata["correlation_id"])
	assert.Equal(t, "cause-456", envelope.Metadata["causation_id"])
	assert.Equal(t, "user-789", envelope.Metadata["user_id"])
}

func TestSnapshot_StateEncoding(t *testing.T) {
	state := []byte(`{"balance":5000,"status":"active"}`)
	snapshot := Snapshot{
		AggregateID: "acc-123",
		State:       state,
		Version:     15,
	}

	assert.NotNil(t, snapshot.State)
	assert.Contains(t, string(snapshot.State), "balance")
	assert.Contains(t, string(snapshot.State), "active")
}

func TestConcurrencyError_TypeAssertion(t *testing.T) {
	err := NewConcurrencyError("agg-111", 2, 4)

	var concErr *ConcurrencyError
	assert.True(t, errors.As(err, &concErr))
	assert.Equal(t, "agg-111", concErr.AggregateID)
	assert.Equal(t, 2, concErr.ExpectedVersion)
	assert.Equal(t, 4, concErr.ActualVersion)
}

func TestNewConcurrencyError_Creation(t *testing.T) {
	err := NewConcurrencyError("test-agg", 1, 3)

	assert.NotNil(t, err)
	assert.Error(t, err)
}
