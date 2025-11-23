-- Migração de tabelas existentes para os schemas apropriados
-- Este script move as tabelas do schema público para os bounded contexts

-- ============================================
-- USER CONTEXT: Tabelas relacionadas a usuários
-- ============================================

-- Mover tabela users
ALTER TABLE IF EXISTS public.users SET SCHEMA user_context;

-- Mover tabela wallet_info
ALTER TABLE IF EXISTS public.wallet_info SET SCHEMA user_context;

-- ============================================
-- TRANSACTION CONTEXT: Tabelas de transações
-- ============================================

-- Mover tabela transactions (caso exista)
ALTER TABLE IF EXISTS public.transactions SET SCHEMA transaction_context;

-- Mover tabela transaction_queue (caso exista)
ALTER TABLE IF EXISTS public.transaction_queue SET SCHEMA transaction_context;

-- ============================================
-- BLOCKCHAIN CONTEXT: Tabelas blockchain
-- ============================================

-- Mover tabela tron_transactions (caso exista)
ALTER TABLE IF EXISTS public.tron_transactions SET SCHEMA blockchain_context;

-- Mover tabela blockchain_wallets (caso exista)
ALTER TABLE IF EXISTS public.blockchain_wallets SET SCHEMA blockchain_context;

-- ============================================
-- Criar views no schema público para retrocompatibilidade
-- ============================================

-- View para users (mantém compatibilidade com código existente)
CREATE OR REPLACE VIEW public.users AS
SELECT * FROM user_context.users;

-- View para wallet_info
CREATE OR REPLACE VIEW public.wallet_info AS
SELECT * FROM user_context.wallet_info;

-- Comentários
COMMENT ON VIEW public.users IS 'Compatibilidade: View apontando para user_context.users';
COMMENT ON VIEW public.wallet_info IS 'Compatibilidade: View apontando para user_context.wallet_info';
