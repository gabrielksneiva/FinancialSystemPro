-- 008_create_blockchain_tables.sql
-- Adds blocks, transactions and balances tables for blockchain persistence

CREATE TABLE IF NOT EXISTS blocks (
  id SERIAL PRIMARY KEY,
  number BIGINT NOT NULL UNIQUE,
  hash TEXT,
  occurred_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
  id TEXT PRIMARY KEY,
  hash TEXT NOT NULL,
  block_number BIGINT,
  from_addr TEXT,
  to_addr TEXT,
  value TEXT,
  raw BYTEA,
  confirmations BIGINT DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS balances (
  address TEXT PRIMARY KEY,
  balance TEXT,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
