package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"financial-system-pro/internal/shared/eventsourcing"
	"log"
	"time"

	"github.com/google/uuid"
)

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db *sql.DB
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db *sql.DB) *PostgresEventStore {
	return &PostgresEventStore{db: db}
}

// SaveEvents appends events to an aggregate's event stream
func (s *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, events []eventsourcing.DomainEvent, expectedVersion int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// opcional: logar em vez de deixar vazio
			log.Printf("rollback falhou: %v", err)
		}
	}()

	// Check current version for optimistic concurrency control
	var currentVersion int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1`,
		aggregateID,
	).Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if currentVersion != expectedVersion {
		return eventsourcing.NewConcurrencyError(aggregateID, expectedVersion, currentVersion)
	}

	// Insert events
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO event_store (
			id, aggregate_id, aggregate_type, event_type, 
			event_data, metadata, version, occurred_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, event := range events {
		eventData, err := json.Marshal(event.Data())
		if err != nil {
			return err
		}

		metadata, err := json.Marshal(map[string]interface{}{
			"event_id":   event.EventID(),
			"event_type": event.EventType(),
		})
		if err != nil {
			return err
		}

		version := expectedVersion + i + 1

		_, err = stmt.ExecContext(ctx,
			uuid.New(),
			aggregateID,
			event.AggregateType(),
			event.EventType(),
			eventData,
			metadata,
			version,
			event.OccurredAt(),
			time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadEvents retrieves all events for an aggregate
func (s *PostgresEventStore) LoadEvents(ctx context.Context, aggregateID string) ([]eventsourcing.EventEnvelope, error) {
	return s.LoadEventsFrom(ctx, aggregateID, 0)
}

// LoadEventsFrom retrieves events from a specific version
func (s *PostgresEventStore) LoadEventsFrom(ctx context.Context, aggregateID string, fromVersion int) ([]eventsourcing.EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, aggregate_id, aggregate_type, event_type, 
			   event_data, metadata, version, occurred_at, created_at
		FROM event_store
		WHERE aggregate_id = $1 AND version > $2
		ORDER BY version ASC
	`, aggregateID, fromVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []eventsourcing.EventEnvelope
	for rows.Next() {
		var envelope eventsourcing.EventEnvelope
		var metadataJSON []byte

		err := rows.Scan(
			&envelope.ID,
			&envelope.AggregateID,
			&envelope.AggregateType,
			&envelope.EventType,
			&envelope.EventData,
			&metadataJSON,
			&envelope.Version,
			&envelope.OccurredAt,
			&envelope.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &envelope.Metadata); err != nil {
			return nil, err
		}

		events = append(events, envelope)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(events) == 0 && fromVersion == 0 {
		return nil, eventsourcing.ErrEventNotFound
	}

	return events, nil
}

// LoadEventsByType retrieves all events of a specific type
func (s *PostgresEventStore) LoadEventsByType(ctx context.Context, eventType string, limit int) ([]eventsourcing.EventEnvelope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, aggregate_id, aggregate_type, event_type, 
			   event_data, metadata, version, occurred_at, created_at
		FROM event_store
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []eventsourcing.EventEnvelope
	for rows.Next() {
		var envelope eventsourcing.EventEnvelope
		var metadataJSON []byte

		err := rows.Scan(
			&envelope.ID,
			&envelope.AggregateID,
			&envelope.AggregateType,
			&envelope.EventType,
			&envelope.EventData,
			&metadataJSON,
			&envelope.Version,
			&envelope.OccurredAt,
			&envelope.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &envelope.Metadata); err != nil {
			return nil, err
		}

		events = append(events, envelope)
	}

	return events, rows.Err()
}

// GetVersion returns the current version of an aggregate
func (s *PostgresEventStore) GetVersion(ctx context.Context, aggregateID string) (int, error) {
	var version int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1`,
		aggregateID,
	).Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return version, err
}
