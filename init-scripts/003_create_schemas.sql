-- Criação dos schemas separados por bounded context
-- Isso facilita separação futura em microserviços e melhora organização do DDD

-- Schema para o contexto de User
CREATE SCHEMA IF NOT EXISTS user_context;

-- Schema para o contexto de Transaction
CREATE SCHEMA IF NOT EXISTS transaction_context;

-- Schema para o contexto de Blockchain
CREATE SCHEMA IF NOT EXISTS blockchain_context;

-- Conceder permissões (ajustar conforme seu usuário de banco)
GRANT ALL PRIVILEGES ON SCHEMA user_context TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA transaction_context TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA blockchain_context TO postgres;

-- Comentários explicativos
COMMENT ON SCHEMA user_context IS 'Bounded Context: Gerenciamento de usuários, autenticação e wallets dos usuários';
COMMENT ON SCHEMA transaction_context IS 'Bounded Context: Transações financeiras (depósito, saque, transferência)';
COMMENT ON SCHEMA blockchain_context IS 'Bounded Context: Integração com blockchain TRON, transações on-chain';
