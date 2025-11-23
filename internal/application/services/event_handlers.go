package services

import (
	"context"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/metrics"

	"go.uber.org/zap"
)

// EventHandlers gerencia os handlers de eventos
type EventHandlers struct {
	logger *zap.Logger
}

// NewEventHandlers cria uma nova instância de EventHandlers
func NewEventHandlers(logger *zap.Logger) *EventHandlers {
	return &EventHandlers{
		logger: logger,
	}
}

// SetupEventSubscribers registra todos os handlers de eventos
func SetupEventSubscribers(bus events.Bus, logger *zap.Logger) {
	handlers := NewEventHandlers(logger)

	// Eventos de Transaction
	bus.Subscribe("deposit.completed", handlers.OnDepositCompleted)
	bus.Subscribe("withdraw.completed", handlers.OnWithdrawCompleted)
	bus.Subscribe("transfer.completed", handlers.OnTransferCompleted)
	bus.Subscribe("transaction.failed", handlers.OnTransactionFailed)

	// Eventos de User
	bus.Subscribe("user.created", handlers.OnUserCreated)
	bus.Subscribe("user.authenticated", handlers.OnUserAuthenticated)

	// Eventos de Blockchain
	bus.Subscribe("wallet.created", handlers.OnWalletCreated)
	bus.Subscribe("blockchain.transaction.confirmed", handlers.OnBlockchainTransactionConfirmed)

	logger.Info("event subscribers registered successfully")
}

// OnDepositCompleted processa eventos de depósito completado
func (h *EventHandlers) OnDepositCompleted(ctx context.Context, e events.Event) error {
	event := e.(events.DepositCompletedEvent)

	h.logger.Info("📥 deposit completed event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("amount", event.Amount.String()),
		zap.String("tx_hash", event.TxHash),
	)

	// Registrar métrica
	metrics.RecordDeposit(event.Amount.InexactFloat64(), true)

	// Aqui você pode adicionar lógica adicional:
	// - Atualizar estatísticas do usuário
	// - Enviar notificação por email/push
	// - Atualizar analytics
	// - Acionar webhooks

	return nil
} // OnWithdrawCompleted processa eventos de saque completado
func (h *EventHandlers) OnWithdrawCompleted(ctx context.Context, e events.Event) error {
	event := e.(events.WithdrawCompletedEvent)

	h.logger.Info("📤 withdraw completed event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("amount", event.Amount.String()),
		zap.String("tx_hash", event.TxHash),
	)

	// Registrar métrica
	metrics.RecordWithdraw(event.Amount.InexactFloat64(), true)

	// Lógica adicional para saques

	return nil
} // OnTransferCompleted processa eventos de transferência completada
func (h *EventHandlers) OnTransferCompleted(ctx context.Context, e events.Event) error {
	event := e.(events.TransferCompletedEvent)

	h.logger.Info("💸 transfer completed event received",
		zap.String("from_user_id", event.FromUserID.String()),
		zap.String("to_user_id", event.ToUserID.String()),
		zap.String("amount", event.Amount.String()),
		zap.String("tx_hash", event.TxHash),
	)

	// Registrar métrica
	metrics.RecordTransfer(event.Amount.InexactFloat64(), true)

	// Lógica adicional para transferências
	// - Notificar ambos os usuários
	// - Atualizar rankings/leaderboards

	return nil
} // OnTransactionFailed processa eventos de transação falhada
func (h *EventHandlers) OnTransactionFailed(ctx context.Context, e events.Event) error {
	event := e.(events.TransactionFailedEvent)

	h.logger.Error("❌ transaction failed event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("tx_type", event.TxType),
		zap.String("amount", event.Amount.String()),
		zap.String("reason", event.Reason),
		zap.String("error_code", event.ErrorCode),
	)

	// Lógica de tratamento de falhas:
	// - Alertar time de suporte
	// - Criar ticket automático
	// - Notificar usuário
	// - Registrar em sistema de monitoramento

	return nil
}

// OnUserCreated processa eventos de criação de usuário
func (h *EventHandlers) OnUserCreated(ctx context.Context, e events.Event) error {
	event := e.(events.UserCreatedEvent)

	h.logger.Info("👤 user created event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("email", event.Email),
		zap.String("name", event.Name),
	)

	// Registrar métrica
	metrics.RecordUserCreated()

	// Lógica pós-criação:
	// - Enviar email de boas-vindas
	// - Criar wallet inicial
	// - Registrar em sistema de analytics
	// - Adicionar a campanhas de marketing

	return nil
} // OnUserAuthenticated processa eventos de autenticação
func (h *EventHandlers) OnUserAuthenticated(ctx context.Context, e events.Event) error {
	event := e.(events.UserAuthenticatedEvent)

	h.logger.Info("🔐 user authenticated event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("email", event.Email),
		zap.String("ip_address", event.IPAddress),
	)

	// Lógica de segurança:
	// - Detectar login suspeito (IP diferente, localização)
	// - Atualizar last_login_at
	// - Registrar para auditoria

	return nil
}

// OnWalletCreated processa eventos de criação de carteira
func (h *EventHandlers) OnWalletCreated(ctx context.Context, e events.Event) error {
	event := e.(events.WalletCreatedEvent)

	h.logger.Info("💼 wallet created event received",
		zap.String("user_id", event.UserID.String()),
		zap.String("wallet_address", event.WalletAddress),
		zap.String("blockchain_type", event.BlockchainType),
	)

	// Registrar métrica
	metrics.RecordWalletCreated(event.BlockchainType)

	// Lógica pós-criação de carteira:
	// - Notificar usuário
	// - Fazer backup das chaves
	// - Registrar em sistema de monitoramento

	return nil
} // OnBlockchainTransactionConfirmed processa confirmações de transações blockchain
func (h *EventHandlers) OnBlockchainTransactionConfirmed(ctx context.Context, e events.Event) error {
	event := e.(events.BlockchainTransactionConfirmedEvent)

	h.logger.Info("⛓️ blockchain transaction confirmed event received",
		zap.String("tx_hash", event.TxHash),
		zap.Int("confirmations", event.Confirmations),
		zap.Int64("block_number", event.BlockNumber),
		zap.String("blockchain_type", event.BlockchainType),
	)

	// Lógica de confirmação:
	// - Liberar fundos se confirmações suficientes
	// - Atualizar status da transação
	// - Notificar usuário

	return nil
}
