package eventsourcing

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a generic domain event for event sourcing
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
	Version() int
	Data() interface{}
}

// EventEnvelope wraps a domain event with metadata for storage
type EventEnvelope struct {
	ID            uuid.UUID              `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	EventData     json.RawMessage        `json:"event_data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Version       int                    `json:"version"`
	OccurredAt    time.Time              `json:"occurred_at"`
	CreatedAt     time.Time              `json:"created_at"`
}

// EventStore interface defines event sourcing operations
type EventStore interface {
	// SaveEvents appends events to an aggregate's event stream
	SaveEvents(ctx context.Context, aggregateID string, events []DomainEvent, expectedVersion int) error

	// LoadEvents retrieves all events for an aggregate
	LoadEvents(ctx context.Context, aggregateID string) ([]EventEnvelope, error)

	// LoadEventsFrom retrieves events from a specific version
	LoadEventsFrom(ctx context.Context, aggregateID string, fromVersion int) ([]EventEnvelope, error)

	// LoadEventsByType retrieves all events of a specific type
	LoadEventsByType(ctx context.Context, eventType string, limit int) ([]EventEnvelope, error)

	// GetVersion returns the current version of an aggregate
	GetVersion(ctx context.Context, aggregateID string) (int, error)
}

// EventBus interface for publishing events to subscribers
type EventBus interface {
	// Publish publishes an event to all subscribers
	Publish(ctx context.Context, event DomainEvent) error

	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []DomainEvent) error

	// Subscribe registers a handler for specific event types
	Subscribe(eventType string, handler EventHandler) error
}

// EventHandler processes domain events
type EventHandler func(ctx context.Context, event DomainEvent) error

// Snapshot represents an aggregate state at a specific point in time
type Snapshot struct {
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	Version       int             `json:"version"`
	State         json.RawMessage `json:"state"`
	CreatedAt     time.Time       `json:"created_at"`
}

// SnapshotStore handles aggregate snapshots
type SnapshotStore interface {
	// SaveSnapshot stores an aggregate snapshot
	SaveSnapshot(ctx context.Context, snapshot *Snapshot) error

	// LoadSnapshot retrieves the latest snapshot for an aggregate
	LoadSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)

	// DeleteSnapshots removes snapshots older than a specific version
	DeleteSnapshots(ctx context.Context, aggregateID string, beforeVersion int) error
}

// EventSourcedAggregate interface for aggregates that support event sourcing
type EventSourcedAggregate interface {
	// AggregateID returns the unique identifier
	AggregateID() string

	// AggregateType returns the type name
	AggregateType() string

	// Version returns the current version
	Version() int

	// UncommittedEvents returns events not yet persisted
	UncommittedEvents() []DomainEvent

	// MarkEventsAsCommitted clears uncommitted events after persistence
	MarkEventsAsCommitted()

	// LoadFromHistory reconstructs aggregate from events
	LoadFromHistory(events []DomainEvent) error
}

// ConcurrencyError indicates a version conflict
type ConcurrencyError struct {
	AggregateID     string
	ExpectedVersion int
	ActualVersion   int
}

func (e *ConcurrencyError) Error() string {
	return "concurrency conflict: expected version " +
		string(rune(e.ExpectedVersion)) + " but got " +
		string(rune(e.ActualVersion))
}

// NewConcurrencyError creates a new concurrency error
func NewConcurrencyError(aggregateID string, expected, actual int) error {
	return &ConcurrencyError{
		AggregateID:     aggregateID,
		ExpectedVersion: expected,
		ActualVersion:   actual,
	}
}

// ErrEventNotFound indicates no events found for an aggregate
var ErrEventNotFound = errors.New("no events found for aggregate")

// ErrSnapshotNotFound indicates no snapshot found for an aggregate
var ErrSnapshotNotFound = errors.New("no snapshot found for aggregate")
