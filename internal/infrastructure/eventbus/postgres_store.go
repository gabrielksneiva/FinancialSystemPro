package eventbus

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) SaveEvent(ctx context.Context, event Event) error {
	payloadJSON, err := json.Marshal(event.Payload())
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(event.Metadata())
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO event_store 
		(event_id, event_type, aggregate_id, aggregate_type, payload, metadata, version, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.EventID(), event.EventType(), event.AggregateID(), event.AggregateType(),
		payloadJSON, metadataJSON, event.Version(), event.OccurredAt())

	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return errors.New("duplicate event: event already exists")
	}

	return err
}

func (s *PostgresStore) GetEvents(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, payload, metadata, version, occurred_at
		FROM event_store
		WHERE aggregate_id = $1 AND version >= $2
		ORDER BY version ASC
	`, aggregateID, fromVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

func (s *PostgresStore) GetEventsSince(ctx context.Context, since time.Time) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, payload, metadata, version, occurred_at
		FROM event_store
		WHERE occurred_at >= $1
		ORDER BY occurred_at ASC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

func (s *PostgresStore) GetEventByID(ctx context.Context, eventID string) (Event, error) {
	var id, eventType, aggID, aggType string
	var payloadJSON, metadataJSON []byte
	var version int
	var occurredAt time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT event_id, event_type, aggregate_id, aggregate_type, payload, metadata, version, occurred_at
		FROM event_store
		WHERE event_id = $1
	`, eventID).Scan(&id, &eventType, &aggID, &aggType, &payloadJSON, &metadataJSON, &version, &occurredAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("event not found")
	}
	if err != nil {
		return nil, err
	}

	var metadata map[string]string
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		metadata = make(map[string]string)
	}

	return &BaseEvent{
		ID:          id,
		Type:        eventType,
		AggID:       aggID,
		AggType:     aggType,
		Ver:         version,
		Occurred:    occurredAt,
		PayloadData: payloadJSON,
		MetadataMap: metadata,
	}, nil
}

func (s *PostgresStore) scanEvents(rows *sql.Rows) ([]Event, error) {
	var events []Event
	for rows.Next() {
		var id, eventType, aggID, aggType string
		var payloadJSON, metadataJSON []byte
		var version int
		var occurredAt time.Time

		if err := rows.Scan(&id, &eventType, &aggID, &aggType, &payloadJSON, &metadataJSON, &version, &occurredAt); err != nil {
			return nil, err
		}

		var metadata map[string]string
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			metadata = make(map[string]string)
		}

		events = append(events, &BaseEvent{
			ID:          id,
			Type:        eventType,
			AggID:       aggID,
			AggType:     aggType,
			Ver:         version,
			Occurred:    occurredAt,
			PayloadData: payloadJSON,
			MetadataMap: metadata,
		})
	}

	return events, rows.Err()
}

// PostgresProcessingLog implements ProcessingLog using PostgreSQL.
type PostgresProcessingLog struct {
	db *sql.DB
}

func NewPostgresProcessingLog(db *sql.DB) *PostgresProcessingLog {
	return &PostgresProcessingLog{db: db}
}

func (p *PostgresProcessingLog) MarkProcessing(ctx context.Context, eventID, handlerName string) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO event_processing_log (event_id, handler_name, status, attempts)
		VALUES ($1, $2, 'processing', 1)
		ON CONFLICT (event_id, handler_name) DO UPDATE SET
			status = 'processing',
			attempts = event_processing_log.attempts + 1,
			started_at = CURRENT_TIMESTAMP
	`, eventID, handlerName)
	return err
}

func (p *PostgresProcessingLog) MarkCompleted(ctx context.Context, eventID, handlerName string) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE event_processing_log
		SET status = 'completed', completed_at = CURRENT_TIMESTAMP
		WHERE event_id = $1 AND handler_name = $2
	`, eventID, handlerName)
	return err
}

func (p *PostgresProcessingLog) MarkFailed(ctx context.Context, eventID, handlerName string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_, dbErr := p.db.ExecContext(ctx, `
		UPDATE event_processing_log
		SET status = 'failed', last_error = $3, completed_at = CURRENT_TIMESTAMP
		WHERE event_id = $1 AND handler_name = $2
	`, eventID, handlerName, errMsg)
	return dbErr
}

func (p *PostgresProcessingLog) IsProcessed(ctx context.Context, eventID, handlerName string) (bool, error) {
	var status string
	err := p.db.QueryRowContext(ctx, `
		SELECT status FROM event_processing_log
		WHERE event_id = $1 AND handler_name = $2
	`, eventID, handlerName).Scan(&status)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return status == "completed", nil
}
