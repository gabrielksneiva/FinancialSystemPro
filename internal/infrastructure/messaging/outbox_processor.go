package messaging

import (
	"context"
	"encoding/json"
	"time"

	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OutboxRecord representa um registro pendente no outbox.
type OutboxRecord struct {
	ID          uuid.UUID
	AggregateID string
	EventType   string
	Payload     string
	Attempts    int
	LastError   string
	PublishedAt *time.Time
}

// OutboxStore define operações necessárias para processar eventos pendentes.
type OutboxStore interface {
	FetchPending(ctx context.Context, limit int) ([]OutboxRecord, error)
	MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

// OutboxProcessor processa registros do outbox e publica via event bus.
type OutboxProcessor struct {
	store OutboxStore
	bus   events.Bus
	log   *zap.Logger
}

func NewOutboxProcessor(store OutboxStore, bus events.Bus, log *zap.Logger) *OutboxProcessor {
	return &OutboxProcessor{store: store, bus: bus, log: log}
}

// ProcessBatch publica até limit eventos pendentes.
func (p *OutboxProcessor) ProcessBatch(ctx context.Context, limit int) error {
	pending, err := p.store.FetchPending(ctx, limit)
	if err != nil {
		return err
	}
	for _, rec := range pending {
		var envelope map[string]interface{}
		if err := json.Unmarshal([]byte(rec.Payload), &envelope); err != nil {
			p.log.Error("failed to unmarshal outbox payload", zap.String("event_type", rec.EventType), zap.Error(err))
			_ = p.store.MarkFailed(ctx, rec.ID, err.Error())
			continue
		}
		base := events.NewOldBaseEvent(rec.EventType, rec.AggregateID)
		publishable := &GenericPublishedEvent{BaseEvent: base, Data: envelope}
		if err := p.bus.Publish(ctx, publishable); err != nil {
			p.log.Error("failed to publish event", zap.String("event_type", rec.EventType), zap.Error(err))
			_ = p.store.MarkFailed(ctx, rec.ID, err.Error())
			continue
		}
		_ = p.store.MarkPublished(ctx, rec.ID, time.Now())
	}
	return nil
}

// GenericPublishedEvent é um wrapper para publicar carga arbitrária mantendo OldBaseEvent.
type GenericPublishedEvent struct {
	BaseEvent events.OldBaseEvent
	Data      map[string]interface{}
}

func (g *GenericPublishedEvent) EventType() string     { return g.BaseEvent.EventType() }
func (g *GenericPublishedEvent) OccurredAt() time.Time { return g.BaseEvent.OccurredAt() }
func (g *GenericPublishedEvent) AggregateID() string   { return g.BaseEvent.AggregateID() }
