#!/bin/bash

# Deploy script para Railway
# Este script faz deploy da aplicação no Railway

set -e

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  Deploy FinancialSystemPro no Railway                         ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Verificar se está logado no Railway
echo -e "${BLUE}📋 Verificando autenticação Railway...${NC}"
if ! railway whoami > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Não está logado. Fazendo login...${NC}"
    railway login
fi

echo -e "${GREEN}✅ Autenticado${NC}"
echo ""

# 2. Listar serviços
echo -e "${BLUE}📦 Serviços disponíveis:${NC}"
railway service

echo ""

# 3. Verificar se há serviço selecionado
echo -e "${BLUE}🔍 Linkando serviço Go...${NC}"
railway service --select

echo ""

# 4. Compilar e fazer build
echo -e "${BLUE}🔨 Buildando aplicação...${NC}"
go mod download
go mod tidy
go build -o main .

echo -e "${GREEN}✅ Build concluído${NC}"
echo ""

# 5. Fazer deploy
echo -e "${BLUE}🚀 Fazendo deploy no Railway...${NC}"
railway up

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Deploy concluído com sucesso!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# 6. Exibir informações da aplicação
echo -e "${BLUE}📡 Acessando aplicação:${NC}"
echo ""
echo "Dashboard: https://railway.app/project/eea37a8b-1398-43fc-9e40-332701f4bc17"
echo ""

# 7. Abrir logs
echo -e "${BLUE}📋 Deseja ver os logs? (s/n)${NC}"
read -r response
if [ "$response" = "s" ] || [ "$response" = "S" ]; then
    railway logs --follow
fi
