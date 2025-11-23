package service

import (
	"context"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/contexts/transaction/domain/repository"
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

// ProcessDeposit processa um depósito
func (s *TransactionService) ProcessDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, callbackURL string) error {
	// Criar transação
	tx := entity.NewTransaction(userID, entity.TransactionTypeDeposit, amount)
	tx.CallbackURL = callbackURL

	if err := s.txRepo.Create(ctx, tx); err != nil {
		s.logger.Error("failed to create deposit transaction", zap.Error(err))
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
		s.txRepo.Update(ctx, tx)
		return err
	}

	wallet := walletInterface.(*userEntity.Wallet)

	// Atualizar saldo
	newBalance := wallet.Balance + amount.InexactFloat64()
	if err := s.walletRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("failed to update balance", zap.Error(err))
		tx.Fail("failed to update balance")
		s.txRepo.Update(ctx, tx)
		return err
	}

	// Marcar como concluída
	tx.Complete("deposit-" + tx.ID.String())
	if err := s.txRepo.Update(ctx, tx); err != nil {
		s.logger.Error("failed to update transaction", zap.Error(err))
		return err
	}

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.NewDepositCompletedEvent(
		userID,
		amount,
		tx.TransactionHash,
	))

	s.logger.Info("deposit processed successfully",
		zap.String("tx_id", tx.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
	)

	return nil
}

// ProcessWithdraw processa um saque
func (s *TransactionService) ProcessWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	// Validar saldo usando circuit breaker
	breaker := s.breakerManager.GetBreaker(breaker.BreakerTransactionToUser)

	walletInterface, err := breaker.Execute(func() (interface{}, error) {
		return s.walletRepo.FindByUserID(ctx, userID)
	})

	if err != nil {
		s.logger.Error("failed to get user wallet", zap.Error(err))
		return ErrInsufficientBalance
	}

	wallet := walletInterface.(*userEntity.Wallet)

	if wallet.Balance < amount.InexactFloat64() {
		return ErrInsufficientBalance
	}

	// Criar transação
	tx := entity.NewTransaction(userID, entity.TransactionTypeWithdraw, amount)
	tx.FromAddress = wallet.Address

	if err := s.txRepo.Create(ctx, tx); err != nil {
		s.logger.Error("failed to create withdraw transaction", zap.Error(err))
		return err
	}

	// Atualizar saldo
	newBalance := wallet.Balance - amount.InexactFloat64()
	if err := s.walletRepo.UpdateBalance(ctx, userID, newBalance); err != nil {
		s.logger.Error("failed to update balance", zap.Error(err))
		tx.Fail("failed to update balance")
		s.txRepo.Update(ctx, tx)
		return err
	}

	// Marcar como concluída
	tx.Complete("withdraw-" + tx.ID.String())
	s.txRepo.Update(ctx, tx)

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.NewWithdrawCompletedEvent(
		userID,
		amount,
		tx.TransactionHash,
	))

	return nil
}

// GetTransactionHistory retorna o histórico de transações de um usuário
func (s *TransactionService) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	return s.txRepo.FindByUserID(ctx, userID)
}
