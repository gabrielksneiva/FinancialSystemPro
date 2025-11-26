package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent interface para todos os eventos de domínio DDD
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
}

// BaseDomainEvent contém campos comuns a todos os eventos DDD
type BaseDomainEvent struct {
	eventID     string
	eventType   string
	occurredAt  time.Time
	aggregateID uuid.UUID
}

// NewBaseDomainEvent cria um novo evento base DDD
func NewBaseDomainEvent(eventType string, aggregateID uuid.UUID) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:     uuid.New().String(),
		eventType:   eventType,
		occurredAt:  time.Now(),
		aggregateID: aggregateID,
	}
}

func (e BaseDomainEvent) EventID() string        { return e.eventID }
func (e BaseDomainEvent) EventType() string      { return e.eventType }
func (e BaseDomainEvent) OccurredAt() time.Time  { return e.occurredAt }
func (e BaseDomainEvent) AggregateID() uuid.UUID { return e.aggregateID }
