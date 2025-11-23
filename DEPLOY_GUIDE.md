# 🚀 Guia de Deploy - Sistema de Withdraw TRON

## Passo 1: Obter Credenciais da Wallet Cofre

### 1.1 - Exportar Private Key da sua wallet TronLink

1. Abra o TronLink (extensão do navegador)
2. Clique no menu (3 pontinhos)
3. Vá em **"Exportar Private Key"** ou **"Export Private Key"**
4. Digite sua senha
5. **COPIE** a private key (será algo como: `abc123def456...` - 64 caracteres em hexadecimal)
6. **COPIE** também o endereço da wallet (será algo como: `TEu5YqTxxxxxxxxxxxxxxx`)

### 1.2 - Gerar chave de criptografia (ENCRYPTION_KEY)

Execute no terminal:
```bash
openssl rand -base64 32
```

Copie o resultado (será algo como: `Xy9P2qR3sT4uV5wX6yZ7aB8cD9eF0gH1iJ2kL3mN4o=`)

---

## Passo 2: Configurar Variáveis no Railway

1. Acesse [Railway Dashboard](https://railway.app)
2. Selecione seu projeto
3. Vá em **Variables** (aba de variáveis de ambiente)
4. Adicione as seguintes variáveis:

### Variáveis Obrigatórias:

```bash
# Banco de Dados (já deve estar configurado)
DATABASE_URL=postgresql://...

# JWT (já deve estar configurado)
JWT_SECRET=seu_jwt_secret

# Redis (já deve estar configurado)
REDIS_URL=redis://...

# Criptografia (NOVO - gerado no passo 1.2)
ENCRYPTION_KEY=Xy9P2qR3sT4uV5wX6yZ7aB8cD9eF0gH1iJ2kL3mN4o=

# TRON Testnet
TRON_TESTNET_RPC=https://api.shasta.trongrid.io
TRON_TESTNET_GRPC=grpc.shasta.trongrid.io:50051

# TRON Vault (NOVO - sua wallet do passo 1.1)
TRON_VAULT_ADDRESS=TEu5YqTxxxxxxxxxxxxxxx
TRON_VAULT_PRIVATE_KEY=abc123def456...
```

5. Clique em **"Add Variable"** para cada uma
6. Depois clique em **"Deploy"** ou aguarde o redeploy automático

---

## Passo 3: Testar Localmente (Opcional)

### 3.1 - Criar arquivo `.env` local

```bash
cp .env.example .env
```

Edite o `.env` e preencha com suas credenciais reais (as mesmas do Railway).

### 3.2 - Rodar localmente

```bash
# Compilar
go build

# Rodar
./financial-system-pro
```

### 3.3 - Testar endpoints

```bash
# 1. Criar usuário
curl -X POST http://localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "teste@example.com",
    "password": "senha123"
  }'

# 2. Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "teste@example.com",
    "password": "senha123"
  }'

# Copie o token JWT retornado

# 3. Ver wallet gerada automaticamente
curl -X GET http://localhost:3000/api/wallet \
  -H "Authorization: Bearer SEU_TOKEN_JWT"

# Resposta esperada:
# {
#   "wallet_address": "TXyz...",
#   "blockchain": "tron",
#   "user_id": "uuid"
# }

# 4. Fazer deposit (adicionar saldo interno)
curl -X POST http://localhost:3000/api/deposit \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "100"
  }'

# 5. Verificar saldo
curl -X GET http://localhost:3000/api/balance \
  -H "Authorization: Bearer SEU_TOKEN_JWT"

# 6. Fazer withdraw TRON (vai enviar da VAULT para a wallet do usuário)
curl -X POST http://localhost:3000/api/withdraw \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "10",
    "withdraw_type": "tron"
  }'

# Resposta esperada:
# {
#   "message": "Withdrawal registered...",
#   "tx_id": "uuid",
#   "amount": "10",
#   "to_address": "TXyz...",
#   "status": "confirmed"
# }
```

---

## Passo 4: Verificar no Railway (Produção)

### 4.1 - Ver logs de deploy

1. No Railway, vá em **Deployments**
2. Clique no último deployment
3. Veja os logs para confirmar:
   - ✅ `TRON workers iniciados`
   - ✅ `wallet generated and saved`
   - ✅ Nenhum erro de conexão

### 4.2 - Testar em produção

Use os mesmos comandos curl do passo 3.3, mas substitua:
- `http://localhost:3000` por `https://seu-app.railway.app`

---

## Passo 5: Monitorar Transações TRON

### 5.1 - Ver no TronScan (Testnet)

1. Acesse [Shasta TronScan](https://shasta.tronscan.org/)
2. Cole o endereço da sua **VAULT** (cofre)
3. Você verá as transações saindo para as wallets dos usuários

### 5.2 - Ver no TronScan (wallet do usuário)

1. No mesmo site, cole o `wallet_address` retornado pelo endpoint `/api/wallet`
2. Você verá os TRX chegando do cofre

---

## Troubleshooting

### Erro: "TRON vault not configured"
- Verifique se `TRON_VAULT_ADDRESS` e `TRON_VAULT_PRIVATE_KEY` estão configurados no Railway
- Faça redeploy após adicionar as variáveis

### Erro: "wallet not found for user"
- O usuário foi criado antes de configurar o WalletManager
- Solução: crie um novo usuário

### Erro: "error sending TRON transaction from vault"
- Verifique se a VAULT tem saldo de TRX_TEST
- Verifique se a private key está correta (64 caracteres hex)
- Veja os logs completos no Railway

### Como obter TRX de teste (faucet):
1. Acesse https://www.trongrid.io/faucet
2. Cole o endereço da sua VAULT
3. Clique em "Submit"
4. Aguarde alguns segundos

---

## Checklist Final ✅

- [ ] ENCRYPTION_KEY gerada e configurada
- [ ] TRON_VAULT_ADDRESS configurada
- [ ] TRON_VAULT_PRIVATE_KEY configurada
- [ ] Deploy feito no Railway
- [ ] Testado criar usuário
- [ ] Testado GET /api/wallet
- [ ] Testado POST /api/withdraw com withdraw_type=tron
- [ ] Verificado transação no TronScan
- [ ] Logs sem erros

---

## Estrutura do Fluxo Completo

```
┌─────────────────────────────────────────────────────────┐
│  1. User cria conta                                     │
│     POST /api/users                                     │
│     → Sistema gera wallet TRON automaticamente          │
│     → Salva no BD com private key criptografada        │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  2. User faz login                                      │
│     POST /api/login                                     │
│     → Recebe JWT token                                  │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  3. User consulta sua wallet                            │
│     GET /api/wallet                                     │
│     → Retorna: { wallet_address: "TXyz..." }           │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  4. User faz deposit (crédito interno)                  │
│     POST /api/deposit { amount: "100" }                 │
│     → Saldo interno aumenta                             │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  5. User faz withdraw TRON                              │
│     POST /api/withdraw {                                │
│       amount: "10",                                     │
│       withdraw_type: "tron"                             │
│     }                                                   │
│     → Sistema debita saldo interno (-10)                │
│     → Sistema busca wallet do user no BD                │
│     → Sistema envia TRX:                                │
│         FROM: TRON_VAULT_ADDRESS (sua wallet cheia)     │
│         TO:   User's wallet (TXyz...)                   │
│         AMOUNT: 10 TRX                                  │
│     → TronWorkerPool monitora confirmação               │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  6. Confirmação automática                              │
│     → Worker verifica TX na blockchain a cada 10s       │
│     → Atualiza status no BD quando confirmado           │
│     → Envia callback (se configurado)                   │
└─────────────────────────────────────────────────────────┘
```

---

## Próximos Passos Avançados (Opcional)

1. **Webhook de Confirmação**: Implementar endpoint para receber callbacks quando TX confirmar
2. **Multi-chain**: Adicionar suporte para Ethereum/Bitcoin usando o mesmo padrão
3. **Admin Dashboard**: Criar painel para monitorar todas as transações
4. **Rate Limiting por IP**: Além de por usuário
5. **2FA**: Adicionar autenticação de dois fatores
6. **KYC**: Sistema de verificação de identidade

---

Qualquer dúvida durante o processo, é só perguntar! 🚀
