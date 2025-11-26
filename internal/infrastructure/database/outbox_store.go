package repositories

import (
	"context"
	"financial-system-pro/internal/infrastructure/messaging"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormOutboxStore implements messaging.OutboxStore using existing GORM database connection.
type GormOutboxStore struct {
	db *NewDatabase
}

type outboxRow struct {
	ID          string `gorm:"primaryKey"`
	AggregateID string
	EventType   string
	Payload     string
	Attempts    int
	LastError   string
	PublishedAt *time.Time
	CreatedAt   time.Time
}

func NewGormOutboxStore(db *NewDatabase) *GormOutboxStore { return &GormOutboxStore{db: db} }

func (s *GormOutboxStore) FetchPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	var rows []outboxRow
	q := s.db.DB.WithContext(ctx).Table("outbox").Where("published_at IS NULL").Order("created_at ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]messaging.OutboxRecord, 0, len(rows))
	for _, r := range rows {
		res = append(res, messaging.OutboxRecord{ID: uuid.MustParse(r.ID), AggregateID: r.AggregateID, EventType: r.EventType, Payload: r.Payload, Attempts: r.Attempts, LastError: r.LastError, PublishedAt: r.PublishedAt})
	}
	return res, nil
}

func (s *GormOutboxStore) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	return s.db.DB.WithContext(ctx).Table("outbox").Where("id = ?", id.String()).Updates(map[string]any{"published_at": &publishedAt}).Error
}

func (s *GormOutboxStore) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	return s.db.DB.WithContext(ctx).Table("outbox").Where("id = ?", id.String()).Updates(map[string]any{"last_error": errMsg, "attempts": gorm.Expr("attempts + 1")}).Error
}
