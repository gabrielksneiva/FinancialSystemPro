-- Multi-chain wallets table
CREATE TABLE IF NOT EXISTS onchain_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    blockchain TEXT NOT NULL,
    address TEXT NOT NULL,
    public_key TEXT NOT NULL,
    encrypted_priv_key TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_onchain_wallets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_onchain_user_blockchain UNIQUE (user_id, blockchain),
    CONSTRAINT uq_onchain_address UNIQUE (address)
);
