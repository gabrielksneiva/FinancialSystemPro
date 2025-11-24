package services

import (
	"context"
	w "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
)

// EventsPort abstrai o EventBus para camada application
// Permite trocar implementação (in-memory, Kafka, etc.) sem afetar serviços
// Apenas métodos usados pelo TransactionService atualmente.
// (Pode ser expandido conforme novas publicações síncronas forem necessárias.)

type EventsPort interface {
	PublishAsync(ctx context.Context, event events.Event)
}

// TronPort abstrai operações da TronService necessárias aos serviços de transação
// Mantém somente métodos realmente usados no TransactionService.

type TronPort interface {
	SendTransaction(fromAddress, toAddress string, amount int64, privateKey string) (string, error)
	HasVaultConfigured() bool
	GetVaultAddress() string
	GetVaultPrivateKey() string
}

// QueuePort abstrai submissão de jobs de transação.
// Em vez de acessar diretamente o canal de jobs, expomos um método.

type QueuePort interface {
	QueueTransaction(job w.TransactionJob) error
}

// TronConfirmationPort abstrai submissão de jobs de confirmação de TX Tron.

type TronConfirmationPort interface {
	SubmitConfirmationJob(userID uuid.UUID, txID uuid.UUID, txHash string, callbackURL string)
}

// TokenProvider abstrai criação de tokens (ex: JWT) para não acoplar serviços a lib específica.
// Claims representadas por map[string]interface{} para manter flexibilidade.
type TokenProvider interface {
	CreateToken(claims map[string]interface{}) (string, error)
}

// PasswordHasher abstrai hashing e comparação de senhas.
// Permite trocar algoritmo (bcrypt, argon2, scrypt) sem tocar regras de negócio.
type PasswordHasher interface {
	Compare(raw string, hashed string) (bool, error)
	// Optionally used in user creation flows (não implementado ainda aqui)
	Hash(raw string) (string, error)
}
