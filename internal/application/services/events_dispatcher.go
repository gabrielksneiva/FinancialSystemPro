package services

import (
	"context"
	"financial-system-pro/internal/shared/events"
	"time"
)

// EventsDispatcher lê pendentes e publica via EventsPort.
type EventsDispatcher struct {
	Outbox EventsOutboxPort
	Bus    EventsPort
}

func NewEventsDispatcher(outbox EventsOutboxPort, bus EventsPort) *EventsDispatcher {
	return &EventsDispatcher{Outbox: outbox, Bus: bus}
}

// DispatchOnce publica até `limit` eventos; retorna quantidade enviada.
func (d *EventsDispatcher) DispatchOnce(ctx context.Context, limit int) (int, error) {
	pend, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, rec := range pend {
		// Convert to a domain event with base metadata for bus compatibility
		d.Bus.PublishAsync(ctx, NewOutboxDomainEvent(rec))
		if err := d.Outbox.MarkPublished(ctx, rec.ID); err != nil {
			// tenta marcar falha e continua
			_ = d.Outbox.MarkFailed(ctx, rec.ID, err.Error())
			continue
		}
		sent++
		// pequena pausa para não saturar
		time.Sleep(1 * time.Millisecond)
	}
	return sent, nil
}

// OutboxDomainEvent wraps Outbox payload into a domain-compatible event.
type OutboxDomainEvent struct {
	events.OldBaseEvent
	Payload []byte
}

func NewOutboxDomainEvent(rec *OutboxRecord) OutboxDomainEvent {
	return OutboxDomainEvent{OldBaseEvent: events.NewOldBaseEvent(rec.Type, rec.Aggregate), Payload: rec.Payload}
}
