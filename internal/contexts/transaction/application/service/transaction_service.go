package service

import (
	"context"
	"encoding/json"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/contexts/transaction/domain/repository"
	"financial-system-pro/internal/contexts/transaction/domain/valueobject"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// TransactionService gerencia a lógica de negócio de transações
type TransactionService struct {
	txRepo         repository.TransactionRepository
	userRepo       userRepo.UserRepository
	walletRepo     userRepo.WalletRepository
	eventBus       events.Bus
	breakerManager *breaker.BreakerManager
	logger         *zap.Logger
}

// NewTransactionService cria uma nova instância do serviço
func NewTransactionService(
	txRepo repository.TransactionRepository,
	userRepo userRepo.UserRepository,
	walletRepo userRepo.WalletRepository,
	eventBus events.Bus,
	breakerManager *breaker.BreakerManager,
	logger *zap.Logger,
) *TransactionService {
	return &TransactionService{
		txRepo:         txRepo,
		userRepo:       userRepo,
		walletRepo:     walletRepo,
		eventBus:       eventBus,
		breakerManager: breakerManager,
		logger:         logger,
	}
}

// WithOutbox mantido para compatibilidade com testes antigos; não realiza ação e retorna o próprio serviço.
func (s *TransactionService) WithOutbox(_ interface{}) *TransactionService { return s }

// writeOutbox registra um evento no outbox se configurado.
func (s *TransactionService) writeOutbox(ctx context.Context, typ string, payload interface{}) {
	// Outbox persistence to be handled by new adapter (not legacy services)
	_, _ = json.Marshal(payload)
}

// ProcessDeposit processa um depósito
func (s *TransactionService) ProcessDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, callbackURL string) error {
	// Money VO
	money, err := valueobject.NewMoney(amount, valueobject.Currency("BRL"))
	if err != nil {
		return err
	}
	// Criar transação
	tx := entity.NewTransaction(userID, entity.TransactionTypeDeposit, amount)
	tx.CallbackURL = callbackURL

	if err := s.txRepo.Create(ctx, tx); err != nil {
		s.logger.Error("failed to create deposit transaction", zap.Error(err))
		s.writeOutbox(ctx, "deposit.failed", map[string]interface{}{"error": "create_tx", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	// Buscar wallet do usuário usando circuit breaker
	breaker := s.breakerManager.GetBreaker(breaker.BreakerTransactionToUser)

	walletInterface, err := breaker.Execute(func() (interface{}, error) {
		return s.walletRepo.FindByUserID(ctx, userID)
	})

	if err != nil {
		s.logger.Error("failed to get user wallet (circuit breaker)",
			zap.Error(err),
			zap.String("breaker_state", breaker.State().String()),
		)
		tx.Fail("failed to get user wallet")
		_ = s.txRepo.Update(ctx, tx)
		s.writeOutbox(ctx, "deposit.failed", map[string]interface{}{"error": "wallet_lookup", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	wallet := walletInterface.(*userEntity.Wallet)

	// Atualizar saldo
	newBalance := wallet.Balance + money.Amount().InexactFloat64()
	if err := s.walletRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("failed to update balance", zap.Error(err))
		tx.Fail("failed to update balance")
		_ = s.txRepo.Update(ctx, tx)
		s.writeOutbox(ctx, "deposit.failed", map[string]interface{}{"error": "update_balance", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	// Marcar como concluída
	tx.Complete("deposit-" + tx.ID.String())
	if err := s.txRepo.Update(ctx, tx); err != nil {
		s.logger.Error("failed to update transaction", zap.Error(err))
		s.writeOutbox(ctx, "deposit.failed", map[string]interface{}{"error": "update_tx", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.NewDepositCompletedEvent(
		userID,
		money.Amount(),
		tx.TransactionHash,
	))
	s.writeOutbox(ctx, "deposit.completed", map[string]interface{}{"user_id": userID.String(), "amount": money.Amount().String(), "tx_hash": tx.TransactionHash})

	s.logger.Info("deposit processed successfully",
		zap.String("tx_id", tx.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
	)

	return nil
}

// ProcessWithdraw processa um saque
func (s *TransactionService) ProcessWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	money, err := valueobject.NewMoney(amount, valueobject.Currency("BRL"))
	if err != nil {
		return err
	}
	// Validar saldo usando circuit breaker
	breaker := s.breakerManager.GetBreaker(breaker.BreakerTransactionToUser)

	walletInterface, err := breaker.Execute(func() (interface{}, error) {
		return s.walletRepo.FindByUserID(ctx, userID)
	})

	if err != nil {
		s.logger.Error("failed to get user wallet", zap.Error(err))
		s.writeOutbox(ctx, "withdraw.failed", map[string]interface{}{"error": "wallet_lookup", "user_id": userID.String(), "amount": amount.String()})
		return ErrInsufficientBalance
	}

	wallet := walletInterface.(*userEntity.Wallet)

	if wallet.Balance < money.Amount().InexactFloat64() {
		return ErrInsufficientBalance
	}

	// Criar transação
	tx := entity.NewTransaction(userID, entity.TransactionTypeWithdraw, amount)
	tx.FromAddress = wallet.Address

	if err := s.txRepo.Create(ctx, tx); err != nil {
		s.logger.Error("failed to create withdraw transaction", zap.Error(err))
		s.writeOutbox(ctx, "withdraw.failed", map[string]interface{}{"error": "create_tx", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	// Atualizar saldo
	newBalance := wallet.Balance - money.Amount().InexactFloat64()
	if err := s.walletRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("failed to update balance", zap.Error(err))
		tx.Fail("failed to update balance")
		_ = s.txRepo.Update(ctx, tx)
		s.writeOutbox(ctx, "withdraw.failed", map[string]interface{}{"error": "update_balance", "user_id": userID.String(), "amount": amount.String()})
		return err
	}

	// Marcar como concluída
	tx.Complete("withdraw-" + tx.ID.String())
	_ = s.txRepo.Update(ctx, tx)
	s.writeOutbox(ctx, "withdraw.completed", map[string]interface{}{"user_id": userID.String(), "amount": money.Amount().String(), "tx_hash": tx.TransactionHash})

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.NewWithdrawCompletedEvent(
		userID,
		money.Amount(),
		tx.TransactionHash,
	))

	return nil
}

// GetTransactionHistory retorna o histórico de transações de um usuário
func (s *TransactionService) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	return s.txRepo.FindByUserID(ctx, userID)
}
