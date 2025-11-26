-- Event Sourcing Tables for FASE 4

-- Event Store: stores all domain events for aggregates
CREATE TABLE IF NOT EXISTS event_store (
    id UUID PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    metadata JSONB,
    version INTEGER NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_aggregate_version UNIQUE (aggregate_id, version)
);

-- Indexes for efficient event retrieval
CREATE INDEX IF NOT EXISTS idx_event_store_aggregate_id ON event_store(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_event_store_aggregate_type ON event_store(aggregate_type);
CREATE INDEX IF NOT EXISTS idx_event_store_event_type ON event_store(event_type);
CREATE INDEX IF NOT EXISTS idx_event_store_occurred_at ON event_store(occurred_at);
CREATE INDEX IF NOT EXISTS idx_event_store_created_at ON event_store(created_at);

-- Snapshot Store: stores aggregate snapshots for performance
CREATE TABLE IF NOT EXISTS aggregate_snapshots (
    aggregate_id VARCHAR(255) PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    version INTEGER NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for snapshot retrieval
CREATE INDEX IF NOT EXISTS idx_snapshots_aggregate_type ON aggregate_snapshots(aggregate_type);
CREATE INDEX IF NOT EXISTS idx_snapshots_version ON aggregate_snapshots(version);

-- Read Model: User summary for CQRS queries
CREATE TABLE IF NOT EXISTS user_read_model (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    wallet_address VARCHAR(255),
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_deposits DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_withdrawals DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_transfers_sent DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_transfers_received DECIMAL(20, 8) NOT NULL DEFAULT 0,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_transaction_at TIMESTAMP
);

-- Indexes for user read model
CREATE INDEX IF NOT EXISTS idx_user_read_email ON user_read_model(email);
CREATE INDEX IF NOT EXISTS idx_user_read_active ON user_read_model(is_active);
CREATE INDEX IF NOT EXISTS idx_user_read_balance ON user_read_model(balance);

-- Read Model: Transaction summary for CQRS queries
CREATE TABLE IF NOT EXISTS transaction_read_model (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    status VARCHAR(50) NOT NULL,
    blockchain VARCHAR(50),
    tx_hash VARCHAR(255),
    from_address VARCHAR(255),
    to_address VARCHAR(255),
    confirmations INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES user_read_model(id) ON DELETE CASCADE
);

-- Indexes for transaction read model
CREATE INDEX IF NOT EXISTS idx_tx_read_user_id ON transaction_read_model(user_id);
CREATE INDEX IF NOT EXISTS idx_tx_read_type ON transaction_read_model(type);
CREATE INDEX IF NOT EXISTS idx_tx_read_status ON transaction_read_model(status);
CREATE INDEX IF NOT EXISTS idx_tx_read_created ON transaction_read_model(created_at);
CREATE INDEX IF NOT EXISTS idx_tx_read_hash ON transaction_read_model(tx_hash);

-- Event Processing Checkpoint: tracks last processed event for projections
CREATE TABLE IF NOT EXISTS event_processing_checkpoint (
    projection_name VARCHAR(100) PRIMARY KEY,
    last_event_id UUID NOT NULL,
    last_event_timestamp TIMESTAMP NOT NULL,
    last_event_version INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE event_store IS 'Event sourcing event store - append-only log of all domain events';
COMMENT ON TABLE aggregate_snapshots IS 'Aggregate snapshots for performance optimization';
COMMENT ON TABLE user_read_model IS 'CQRS read model for user queries';
COMMENT ON TABLE transaction_read_model IS 'CQRS read model for transaction queries';
COMMENT ON TABLE event_processing_checkpoint IS 'Tracks projection progress for event replay';
