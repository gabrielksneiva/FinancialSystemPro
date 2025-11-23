.PHONY: help up down restart logs build run migrate clean test fmt

help: ## Mostra este help
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Inicia PostgreSQL e Redis
	docker-compose up -d
	@echo "✅ PostgreSQL e Redis iniciados!"
	@echo "   PostgreSQL: localhost:5432"
	@echo "   Redis: localhost:6379"

down: ## Para todos os containers
	docker-compose down

restart: ## Reinicia os containers
	docker-compose restart

logs: ## Mostra logs dos containers
	docker-compose logs -f

logs-postgres: ## Logs do PostgreSQL
	docker-compose logs -f postgres

logs-redis: ## Logs do Redis
	docker-compose logs -f redis

build: ## Compila a aplicação Go
	go build -o app .

fmt: ## Formata código com gofmt e goimports
	gofmt -w .
	@command -v goimports >/dev/null 2>&1 || { echo "Instalando goimports..."; go install golang.org/x/tools/cmd/goimports@latest; }
	goimports -w .

vet: ## Executa go vet
	go vet ./...

run: ## Roda a aplicação (certifique-se que PostgreSQL e Redis estão up)
	./app

dev: up build run ## Inicia tudo e roda a aplicação

migrate: ## Executa migrations no PostgreSQL local
	@echo "Executando migrations..."
	docker-compose exec postgres psql -U admin -d financialsystempro -f /docker-entrypoint-initdb.d/001_create_tables.sql
	docker-compose exec postgres psql -U admin -d financialsystempro -f /docker-entrypoint-initdb.d/002_create_wallet_info.sql
	@echo "✅ Migrations executadas!"

clean: ## Remove containers e volumes
	docker-compose down -v
	rm -f app

test: ## Roda os testes
	go test -v ./...

test-coverage: ## Roda testes com cobertura
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -n 1

test-wallet: ## Testa geração de carteira TRON
	go test -v ./services -run TestGenerateTronAddress

lint: ## Roda golangci-lint localmente
	@command -v golangci-lint >/dev/null 2>&1 || { echo "❌ golangci-lint não instalado. Instalando..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --timeout=5m

lint-fix: ## Roda golangci-lint e tenta corrigir automaticamente
	@command -v golangci-lint >/dev/null 2>&1 || { echo "❌ golangci-lint não instalado. Instalando..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --fix --timeout=5m

ci-local: lint test-coverage ## Simula CI localmente (lint + testes com cobertura)
	@echo "✅ CI local passou!"

psql: ## Conecta no PostgreSQL via psql
	docker-compose exec postgres psql -U admin -d financialsystempro

redis-cli: ## Conecta no Redis via CLI
	docker-compose exec redis redis-cli

status: ## Mostra status dos containers
	docker-compose ps
