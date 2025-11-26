package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OutboxRecord representa uma mensagem de evento persistida para entrega confiável.
type OutboxRecord struct {
	ID          uuid.UUID
	Aggregate   string
	Type        string
	Payload     []byte
	CreatedAt   time.Time
	Published   bool
	PublishedAt *time.Time
	Attempts    int
	LastError   string
}

// EventsOutboxPort define operações para persistência e leitura de eventos pendentes (Outbox pattern).
type EventsOutboxPort interface {
	Save(ctx context.Context, rec *OutboxRecord) error
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
	ListPending(ctx context.Context, limit int) ([]*OutboxRecord, error)
}
