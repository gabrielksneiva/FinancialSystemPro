package postgres

import (
	"context"
	"database/sql"
	"financial-system-pro/internal/shared/eventsourcing"
	"time"
)

// PostgresSnapshotStore implements SnapshotStore using PostgreSQL
type PostgresSnapshotStore struct {
	db *sql.DB
}

// NewPostgresSnapshotStore creates a new PostgreSQL snapshot store
func NewPostgresSnapshotStore(db *sql.DB) *PostgresSnapshotStore {
	return &PostgresSnapshotStore{db: db}
}

// SaveSnapshot stores an aggregate snapshot
func (s *PostgresSnapshotStore) SaveSnapshot(ctx context.Context, snapshot *eventsourcing.Snapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO aggregate_snapshots (
			aggregate_id, aggregate_type, version, state, created_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (aggregate_id) 
		DO UPDATE SET
			version = EXCLUDED.version,
			state = EXCLUDED.state,
			created_at = EXCLUDED.created_at
	`,
		snapshot.AggregateID,
		snapshot.AggregateType,
		snapshot.Version,
		snapshot.State,
		time.Now(),
	)

	return err
}

// LoadSnapshot retrieves the latest snapshot for an aggregate
func (s *PostgresSnapshotStore) LoadSnapshot(ctx context.Context, aggregateID string) (*eventsourcing.Snapshot, error) {
	snapshot := &eventsourcing.Snapshot{}

	err := s.db.QueryRowContext(ctx, `
		SELECT aggregate_id, aggregate_type, version, state, created_at
		FROM aggregate_snapshots
		WHERE aggregate_id = $1
		ORDER BY version DESC
		LIMIT 1
	`, aggregateID).Scan(
		&snapshot.AggregateID,
		&snapshot.AggregateType,
		&snapshot.Version,
		&snapshot.State,
		&snapshot.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, eventsourcing.SnapshotNotFoundError
	}

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// DeleteSnapshots removes snapshots older than a specific version
func (s *PostgresSnapshotStore) DeleteSnapshots(ctx context.Context, aggregateID string, beforeVersion int) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM aggregate_snapshots
		WHERE aggregate_id = $1 AND version < $2
	`, aggregateID, beforeVersion)

	return err
}
