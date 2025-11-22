-- Criar tabela WalletInfo para armazenar informações de carteira TRON
CREATE TABLE IF NOT EXISTS wallet_infos (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    tron_address TEXT NOT NULL UNIQUE,
    encrypted_priv_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_wallet_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Índice para buscar carteira por endereço TRON
CREATE INDEX IF NOT EXISTS idx_wallet_tron_address ON wallet_infos(tron_address);

-- Índice para buscar carteira por user_id (já único, mas deixa para query optimization)
CREATE INDEX IF NOT EXISTS idx_wallet_user_id ON wallet_infos(user_id);

-- Adicionar coluna para hash de transação TRON na tabela transactions se não existir
ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS tron_tx_hash TEXT,
ADD COLUMN IF NOT EXISTS tron_tx_status TEXT DEFAULT 'pending'; -- pending, confirmed, failed
