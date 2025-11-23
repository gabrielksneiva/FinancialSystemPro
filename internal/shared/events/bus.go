package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Event representa um evento de domínio
type Event interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// Handler é uma função que processa eventos
type Handler func(ctx context.Context, event Event) error

// Bus é a interface do event bus
type Bus interface {
	Subscribe(eventType string, handler Handler)
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event)
}

// InMemoryBus implementa um event bus em memória
// Use este para começar, depois migre para RabbitMQ/Kafka se necessário
type InMemoryBus struct {
	handlers map[string][]Handler
	logger   *zap.Logger
	mu       sync.RWMutex
}

// NewInMemoryBus cria uma nova instância do event bus
func NewInMemoryBus(logger *zap.Logger) *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

// Subscribe registra um handler para um tipo de evento
func (b *InMemoryBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.logger.Info("handler subscribed",
		zap.String("event_type", eventType),
		zap.Int("total_handlers", len(b.handlers[eventType])),
	)
}

// Publish publica um evento de forma síncrona
// Todos os handlers são executados e erros são coletados
func (b *InMemoryBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventType()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		b.logger.Warn("no handlers for event",
			zap.String("event_type", event.EventType()),
		)
		return nil
	}

	b.logger.Info("publishing event",
		zap.String("event_type", event.EventType()),
		zap.String("aggregate_id", event.AggregateID()),
		zap.Int("handlers_count", len(handlers)),
	)

	var errs []error
	for i, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Error("handler failed",
				zap.String("event_type", event.EventType()),
				zap.Int("handler_index", i),
				zap.Error(err),
			)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to process event: %d/%d handlers failed", len(errs), len(handlers))
	}

	return nil
}

// PublishAsync publica um evento de forma assíncrona
// Não espera pelos handlers, apenas enfileira
func (b *InMemoryBus) PublishAsync(ctx context.Context, event Event) {
	go func() {
		if err := b.Publish(ctx, event); err != nil {
			b.logger.Error("async event publish failed",
				zap.String("event_type", event.EventType()),
				zap.Error(err),
			)
		}
	}()
}

// BaseEvent é uma implementação base para eventos
type BaseEvent struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	AggregateId string    `json:"aggregate_id"`
}

func (e BaseEvent) EventType() string {
	return e.Type
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

func (e BaseEvent) AggregateID() string {
	return e.AggregateId
}

// NewBaseEvent cria um novo evento base
func NewBaseEvent(eventType, aggregateID string) BaseEvent {
	return BaseEvent{
		Type:        eventType,
		Timestamp:   time.Now(),
		AggregateId: aggregateID,
	}
}
