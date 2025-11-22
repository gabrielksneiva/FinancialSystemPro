# Deploy no Railway - Guia Rápido

## 🚀 Passo 1: Preparar o Código

```bash
# Ir para a pasta do projeto
cd /mnt/c/Users/gabri/repos/FinancialSystemPro

# Atualizar dependências
go mod download
go mod tidy

# Verificar se compila
go build -o main .
```

## 📝 Passo 2: Configurar Ambiente

As variáveis de ambiente já estão configuradas automaticamente pelo Railway:

- `DB_HOST`: postgres.railway.internal
- `DB_PORT`: 5432
- `DB_USER`: postgres
- `DB_PASSWORD`: (visto no comando anterior)
- `DB_NAME`: railway

**⚠️ Importante**: Você deve adicionar manualmente no Railway Dashboard:

```
JWT_SECRET=seu_secret_super_seguro_aqui
TRON_TESTNET_RPC=https://api.shasta.trongrid.io
TRON_TESTNET_GRPC=grpc.shasta.trongrid.io:50051
```

## 🔗 Passo 3: Linkar Serviço

```bash
# Listar serviços
railway service

# Selecionar o serviço da aplicação (FinancialSystemPro)
railway service --select
```

## 🚀 Passo 4: Deploy

### Opção A: Automático (Recomendado)

```bash
# Usar o script de deploy
chmod +x deploy.sh
./deploy.sh
```

### Opção B: Manual

```bash
# Fazer deploy
railway up

# Ver logs
railway logs --follow
```

## ✅ Passo 5: Verificar Deployment

```bash
# Ir ao dashboard
open https://railway.app/project/eea37a8b-1398-43fc-9e40-332701f4bc17

# Ou testar via CLI
railway variables

# Ou testar via cURL quando deployed
curl https://seu-app.railway.app/health
```

## 🔐 Passo 6: Configurar Variáveis Adicionais

1. Acesse o Dashboard: https://railway.app/project/eea37a8b-1398-43fc-9e40-332701f4bc17
2. Clique no serviço (FinancialSystemPro)
3. Vá em "Variables" (aba)
4. Clique em "Add Variable" e adicione:

```
JWT_SECRET = seu_secret_jwt_aqui
TRON_TESTNET_RPC = https://api.shasta.trongrid.io
TRON_TESTNET_GRPC = grpc.shasta.trongrid.io:50051
PORT = 3000
```

5. Clique em "Save"

## 📊 Passo 7: Monitorar

```bash
# Ver status do deployment
railway status

# Ver logs em tempo real
railway logs --follow

# Ver variáveis
railway variables

# Rebuild e redeploy
railway up
```

## 🌐 URLs Úteis

- **Dashboard**: https://railway.app/project/eea37a8b-1398-43fc-9e40-332701f4bc17
- **Documentação Railway**: https://docs.railway.app
- **Status do Serviço**: Aparecerá no dashboard após deploy

## 🐛 Troubleshooting

### Build fails
```bash
# Limpar cache
go clean -cache
go mod tidy

# Rebuild
go build -o main .
railway up
```

### Erro de conexão com banco
- Verificar se PostgreSQL está rodando no Railway
- Confirmar DB_HOST, DB_PORT, etc
- Checar credenciais em `railway variables`

### Aplicação não inicia
```bash
# Ver logs detalhados
railway logs --follow

# Verificar variáveis de ambiente
railway variables
```

### Reset completo
```bash
# Se precisar resetar tudo
railway remove  # Remove o serviço
railway add     # Adiciona novamente
railway up      # Faz deploy novamente
```

## 🎯 Próximos Passos

1. ✅ Fazer deploy
2. ✅ Testar endpoints
3. ✅ Configurar domínio customizado
4. ✅ Adicionar monitoramento

## 📞 Dúvidas?

Consulte:
- Docs Railway: https://docs.railway.app
- Comunidade: https://discord.gg/railway
- Issues: https://github.com/railwayapp/cli/issues
