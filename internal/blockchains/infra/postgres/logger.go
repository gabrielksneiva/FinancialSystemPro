package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PGLogger struct {
	db *sqlx.DB
}

func NewPGLogger(db *sqlx.DB) *PGLogger {
	return &PGLogger{db: db}
}

func (p *PGLogger) LogEvent(ctx context.Context, eventType string, payload string) error {
	if p.db == nil {
		return fmt.Errorf("db not configured")
	}
	_, err := p.db.ExecContext(ctx, `INSERT INTO events_log (event_type, payload, created_at) VALUES ($1, $2, $3)`, eventType, payload, time.Now())
	return err
}

func (p *PGLogger) EnsureSchema(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("db not configured")
	}
	_, err := p.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS events_log (id SERIAL PRIMARY KEY, event_type TEXT, payload TEXT, created_at TIMESTAMP)`)
	return err
}
