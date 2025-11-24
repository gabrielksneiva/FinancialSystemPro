ALTER TABLE transactions ADD COLUMN IF NOT EXISTS onchain_tx_hash text;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS onchain_tx_status text;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS onchain_chain text;