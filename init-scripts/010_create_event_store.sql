-- 010_create_event_store.sql
-- Persistent event store for Event Bus with idempotency and replay support

-- Event store table (append-only)
CREATE TABLE IF NOT EXISTS event_store (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT UNIQUE NOT NULL,
    event_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB,
    version INTEGER NOT NULL DEFAULT 1,
    occurred_at TIMESTAMP NOT NULL,
    stored_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying and replay
CREATE INDEX idx_event_store_event_id ON event_store(event_id);
CREATE INDEX idx_event_store_aggregate ON event_store(aggregate_id, aggregate_type);
CREATE INDEX idx_event_store_type ON event_store(event_type);
CREATE INDEX idx_event_store_occurred ON event_store(occurred_at DESC);
CREATE INDEX idx_event_store_aggregate_version ON event_store(aggregate_id, version);

-- Event processing tracking (for idempotency and delivery guarantees)
CREATE TABLE IF NOT EXISTS event_processing_log (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL,
    handler_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('processing', 'completed', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 1,
    last_error TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    UNIQUE(event_id, handler_name)
);

CREATE INDEX idx_event_processing_event ON event_processing_log(event_id);
CREATE INDEX idx_event_processing_status ON event_processing_log(status);
CREATE INDEX idx_event_processing_handler ON event_processing_log(handler_name);

-- Event subscriptions registry (for tracking active subscribers)
CREATE TABLE IF NOT EXISTS event_subscriptions (
    id SERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    handler_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_type, handler_name)
);

CREATE INDEX idx_event_subscriptions_type ON event_subscriptions(event_type);

-- Event snapshots (optional optimization for event sourcing)
CREATE TABLE IF NOT EXISTS event_snapshots (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    version INTEGER NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(aggregate_id, aggregate_type, version)
);

CREATE INDEX idx_event_snapshots_aggregate ON event_snapshots(aggregate_id, aggregate_type);

-- Comments for documentation
COMMENT ON TABLE event_store IS 'Append-only event store for all domain events';
COMMENT ON TABLE event_processing_log IS 'Tracks event handler execution for idempotency and retry';
COMMENT ON TABLE event_subscriptions IS 'Registry of active event type subscriptions';
COMMENT ON TABLE event_snapshots IS 'Optional aggregate state snapshots for performance optimization';
COMMENT ON COLUMN event_processing_log.status IS 'processing: in-flight; completed: success; failed: needs retry';
