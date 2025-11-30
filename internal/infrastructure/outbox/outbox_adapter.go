package outbox

import (
	"context"
	appsvc "financial-system-pro/internal/application/services"
	repos "financial-system-pro/internal/infrastructure/database"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormOutboxAdapter implementa EventsOutboxPort usando a conexão existente.
type GormOutboxAdapter struct{ db *repos.NewDatabase }

// NewGormOutboxAdapter constrói adapter.
func NewGormOutboxAdapter(db *repos.NewDatabase) *GormOutboxAdapter {
	return &GormOutboxAdapter{db: db}
}

// outboxModel mapeia para tabela `outbox`.
type outboxModel struct {
	ID          string `gorm:"primaryKey"`
	Aggregate   string
	Type        string
	Payload     []byte
	CreatedAt   time.Time
	Published   bool
	PublishedAt *time.Time
	Attempts    int
	LastError   string
}

func (a *GormOutboxAdapter) Save(ctx context.Context, rec *appsvc.OutboxRecord) error {
	m := &outboxModel{ID: rec.ID.String(), Aggregate: rec.Aggregate, Type: rec.Type, Payload: rec.Payload, CreatedAt: time.Now(), Published: false, Attempts: 0}
	return a.db.DB.WithContext(ctx).Table("outbox").Create(m).Error
}

func (a *GormOutboxAdapter) MarkPublished(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return a.db.DB.WithContext(ctx).Table("outbox").Where("id = ?", id.String()).Updates(map[string]interface{}{"published": true, "published_at": &now}).Error
}

func (a *GormOutboxAdapter) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	return a.db.DB.WithContext(ctx).Table("outbox").Where("id = ?", id.String()).Updates(map[string]interface{}{"last_error": errMsg, "attempts": gorm.Expr("attempts + 1")}).Error
}

func (a *GormOutboxAdapter) ListPending(ctx context.Context, limit int) ([]*appsvc.OutboxRecord, error) {
	var rows []outboxModel
	q := a.db.DB.WithContext(ctx).Table("outbox").Where("published = ?", false).Order("created_at asc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*appsvc.OutboxRecord, 0, len(rows))
	for _, r := range rows {
		out = append(out, &appsvc.OutboxRecord{ID: uuid.MustParse(r.ID), Aggregate: r.Aggregate, Type: r.Type, Payload: r.Payload, CreatedAt: r.CreatedAt, Published: r.Published, PublishedAt: r.PublishedAt, Attempts: r.Attempts, LastError: r.LastError})
	}
	return out, nil
}
