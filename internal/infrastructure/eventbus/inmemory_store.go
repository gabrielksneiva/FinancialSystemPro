package eventbus

import (
	"context"
	"sync"
	"time"
)

// InMemoryStore implements Store using in-memory storage for development/testing.
type InMemoryStore struct {
	mu     sync.RWMutex
	events map[string]Event
	byAgg  map[string][]Event
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		events: make(map[string]Event),
		byAgg:  make(map[string][]Event),
	}
}

func (s *InMemoryStore) SaveEvent(ctx context.Context, event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.EventID()]; exists {
		return ErrDuplicateEvent
	}

	s.events[event.EventID()] = event
	s.byAgg[event.AggregateID()] = append(s.byAgg[event.AggregateID()], event)
	return nil
}

func (s *InMemoryStore) GetEvents(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Event
	for _, e := range s.byAgg[aggregateID] {
		if e.Version() >= fromVersion {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

func (s *InMemoryStore) GetEventsSince(ctx context.Context, since time.Time) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Event
	for _, e := range s.events {
		if e.OccurredAt().After(since) || e.OccurredAt().Equal(since) {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

func (s *InMemoryStore) GetEventByID(ctx context.Context, eventID string) (Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[eventID]
	if !ok {
		return nil, ErrEventNotFound
	}
	return event, nil
}

// InMemoryProcessingLog implements ProcessingLog using in-memory storage.
type InMemoryProcessingLog struct {
	mu        sync.RWMutex
	processed map[string]map[string]string // eventID -> handlerName -> status
}

func NewInMemoryProcessingLog() *InMemoryProcessingLog {
	return &InMemoryProcessingLog{
		processed: make(map[string]map[string]string),
	}
}

func (p *InMemoryProcessingLog) MarkProcessing(ctx context.Context, eventID, handlerName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processed[eventID] == nil {
		p.processed[eventID] = make(map[string]string)
	}
	p.processed[eventID][handlerName] = "processing"
	return nil
}

func (p *InMemoryProcessingLog) MarkCompleted(ctx context.Context, eventID, handlerName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processed[eventID] == nil {
		p.processed[eventID] = make(map[string]string)
	}
	p.processed[eventID][handlerName] = "completed"
	return nil
}

func (p *InMemoryProcessingLog) MarkFailed(ctx context.Context, eventID, handlerName string, err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processed[eventID] == nil {
		p.processed[eventID] = make(map[string]string)
	}
	p.processed[eventID][handlerName] = "failed"
	return nil
}

func (p *InMemoryProcessingLog) IsProcessed(ctx context.Context, eventID, handlerName string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.processed[eventID] == nil {
		return false, nil
	}

	status, exists := p.processed[eventID][handlerName]
	return exists && status == "completed", nil
}

var (
	ErrDuplicateEvent = &BusError{Code: "DUPLICATE_EVENT", Message: "event already exists"}
	ErrEventNotFound  = &BusError{Code: "EVENT_NOT_FOUND", Message: "event not found"}
)

type BusError struct {
	Code    string
	Message string
}

func (e *BusError) Error() string {
	return e.Code + ": " + e.Message
}
