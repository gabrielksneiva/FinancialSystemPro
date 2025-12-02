-- 009_create_finance_ledger.sql
-- Finance domain ledger and related tables

-- Finance Accounts table
CREATE TABLE IF NOT EXISTS finance_accounts (
    id TEXT PRIMARY KEY,
    holder_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_finance_accounts_holder ON finance_accounts(holder_id);

-- Finance Wallets table
CREATE TABLE IF NOT EXISTS finance_wallets (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    balance_value BIGINT NOT NULL DEFAULT 0 CHECK (balance_value >= 0),
    balance_currency TEXT NOT NULL CHECK (LENGTH(balance_currency) = 3),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1, -- optimistic locking
    CONSTRAINT fk_wallet_account FOREIGN KEY (account_id) REFERENCES finance_accounts(id) ON DELETE CASCADE
);

CREATE INDEX idx_finance_wallets_account ON finance_wallets(account_id);
CREATE INDEX idx_finance_wallets_currency ON finance_wallets(balance_currency);

-- Finance Ledger Entries table (append-only, immutable)
CREATE TABLE IF NOT EXISTS finance_ledger_entries (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    entry_type TEXT NOT NULL CHECK (entry_type IN ('credit', 'debit')),
    amount_value BIGINT NOT NULL CHECK (amount_value >= 0),
    amount_currency TEXT NOT NULL CHECK (LENGTH(amount_currency) = 3),
    balance_after BIGINT NOT NULL CHECK (balance_after >= 0),
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ledger_wallet FOREIGN KEY (wallet_id) REFERENCES finance_wallets(id) ON DELETE CASCADE
);

-- Indexes for efficient querying
CREATE INDEX idx_ledger_entries_wallet ON finance_ledger_entries(wallet_id);
CREATE INDEX idx_ledger_entries_created ON finance_ledger_entries(created_at DESC);
CREATE INDEX idx_ledger_entries_wallet_created ON finance_ledger_entries(wallet_id, created_at DESC);

-- Trigger to update updated_at on finance_accounts
CREATE OR REPLACE FUNCTION update_finance_accounts_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_finance_accounts_update
BEFORE UPDATE ON finance_accounts
FOR EACH ROW
EXECUTE FUNCTION update_finance_accounts_timestamp();

-- Trigger to update updated_at and version on finance_wallets
CREATE OR REPLACE FUNCTION update_finance_wallets_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_finance_wallets_update
BEFORE UPDATE ON finance_wallets
FOR EACH ROW
EXECUTE FUNCTION update_finance_wallets_timestamp();

-- Comments for documentation
COMMENT ON TABLE finance_accounts IS 'Financial account aggregates representing account holders';
COMMENT ON TABLE finance_wallets IS 'Wallets holding single-currency balances with optimistic locking';
COMMENT ON TABLE finance_ledger_entries IS 'Append-only immutable ledger of all balance mutations';
COMMENT ON COLUMN finance_wallets.version IS 'Optimistic locking version counter';
