package eventbus

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event with versioning and metadata.
type Event interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	Version() int
	OccurredAt() time.Time
	Payload() interface{}
	Metadata() map[string]string
}

// BaseEvent provides a standard implementation of Event.
type BaseEvent struct {
	ID          string            `json:"event_id"`
	Type        string            `json:"event_type"`
	AggID       string            `json:"aggregate_id"`
	AggType     string            `json:"aggregate_type"`
	Ver         int               `json:"version"`
	Occurred    time.Time         `json:"occurred_at"`
	PayloadData json.RawMessage   `json:"payload"`
	MetadataMap map[string]string `json:"metadata,omitempty"`
}

func NewBaseEvent(eventType, aggregateID, aggregateType string, payload interface{}) (*BaseEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &BaseEvent{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggID:       aggregateID,
		AggType:     aggregateType,
		Ver:         1,
		Occurred:    time.Now().UTC(),
		PayloadData: payloadBytes,
		MetadataMap: make(map[string]string),
	}, nil
}

func (e *BaseEvent) EventID() string             { return e.ID }
func (e *BaseEvent) EventType() string           { return e.Type }
func (e *BaseEvent) AggregateID() string         { return e.AggID }
func (e *BaseEvent) AggregateType() string       { return e.AggType }
func (e *BaseEvent) Version() int                { return e.Ver }
func (e *BaseEvent) OccurredAt() time.Time       { return e.Occurred }
func (e *BaseEvent) Payload() interface{}        { return e.PayloadData }
func (e *BaseEvent) Metadata() map[string]string { return e.MetadataMap }

// Handler processes events.
type Handler func(ctx context.Context, event Event) error

// Bus defines the event bus interface with idempotency and persistence.
type Bus interface {
	// Publish stores and dispatches an event to all subscribers.
	Publish(ctx context.Context, event Event) error

	// PublishAsync stores and dispatches an event asynchronously.
	PublishAsync(ctx context.Context, event Event)

	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType string, handlerName string, handler Handler) error

	// Replay replays events for a specific aggregate or all events from a point in time.
	Replay(ctx context.Context, aggregateID string, fromVersion int) error

	// Close gracefully shuts down the event bus.
	Close() error
}

// Store persists events for replay and audit.
type Store interface {
	// SaveEvent persists an event to the store.
	SaveEvent(ctx context.Context, event Event) error

	// GetEvents retrieves events by aggregate ID.
	GetEvents(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)

	// GetEventsSince retrieves all events since a specific time.
	GetEventsSince(ctx context.Context, since time.Time) ([]Event, error)

	// GetEventByID retrieves a single event by ID.
	GetEventByID(ctx context.Context, eventID string) (Event, error)
}

// ProcessingLog tracks event processing for idempotency.
type ProcessingLog interface {
	// MarkProcessing marks an event as being processed by a handler.
	MarkProcessing(ctx context.Context, eventID, handlerName string) error

	// MarkCompleted marks an event as successfully processed.
	MarkCompleted(ctx context.Context, eventID, handlerName string) error

	// MarkFailed marks an event as failed with error details.
	MarkFailed(ctx context.Context, eventID, handlerName string, err error) error

	// IsProcessed checks if an event was already processed by a handler.
	IsProcessed(ctx context.Context, eventID, handlerName string) (bool, error)
}
