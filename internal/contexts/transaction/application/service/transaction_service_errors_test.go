package service

import (
	"context"
	"testing"

	txEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	txRepoIface "financial-system-pro/internal/contexts/transaction/domain/repository"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepoIface "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// txRepoMock: implementação mínima da interface TransactionRepository
type txRepoMock struct{}

func (r *txRepoMock) Create(ctx context.Context, tx *txEntity.Transaction) error { return nil }
func (r *txRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*txEntity.Transaction, error) {
	return nil, nil
}
func (r *txRepoMock) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*txEntity.Transaction, error) {
	return []*txEntity.Transaction{}, nil
}
func (r *txRepoMock) FindByHash(ctx context.Context, hash string) (*txEntity.Transaction, error) {
	return nil, nil
}
func (r *txRepoMock) Update(ctx context.Context, tx *txEntity.Transaction) error { return nil }
func (r *txRepoMock) UpdateStatus(ctx context.Context, id uuid.UUID, status txEntity.TransactionStatus) error {
	return nil
}

var _ txRepoIface.TransactionRepository = (*txRepoMock)(nil)

// walletRepoMock: implementação mínima da interface WalletRepository
type walletRepoMock struct {
	balance float64
	fail    bool
}

func (w *walletRepoMock) Create(ctx context.Context, wallet *userEntity.Wallet) error { return nil }
func (w *walletRepoMock) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	if w.fail {
		return nil, ErrInsufficientBalance
	}
	return &userEntity.Wallet{UserID: userID, Address: "ADDR", Balance: w.balance}, nil
}
func (w *walletRepoMock) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	return nil, nil
}
func (w *walletRepoMock) UpdateBalance(ctx context.Context, userID uuid.UUID, newBalance float64) error {
	w.balance = newBalance
	return nil
}

var _ userRepoIface.WalletRepository = (*walletRepoMock)(nil)

// userRepoMock: implementação mínima da interface UserRepository
type userRepoMock struct{}

func (u *userRepoMock) Create(ctx context.Context, user *userEntity.User) error { return nil }
func (u *userRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return nil, nil
}
func (u *userRepoMock) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	return nil, nil
}
func (u *userRepoMock) Update(ctx context.Context, user *userEntity.User) error { return nil }
func (u *userRepoMock) Delete(ctx context.Context, id uuid.UUID) error          { return nil }

var _ userRepoIface.UserRepository = (*userRepoMock)(nil)

func TestProcessWithdraw_Insufficient(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	svc := NewTransactionService(&txRepoMock{}, &userRepoMock{}, &walletRepoMock{balance: 5}, eventBus, br, lg)
	err := svc.ProcessWithdraw(context.Background(), uuid.New(), decimal.NewFromFloat(10))
	if err == nil {
		t.Fatalf("esperava erro saldo insuficiente")
	}
}

func TestProcessWithdraw_Success(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	svc := NewTransactionService(&txRepoMock{}, &userRepoMock{}, &walletRepoMock{balance: 15}, eventBus, br, lg)
	err := svc.ProcessWithdraw(context.Background(), uuid.New(), decimal.NewFromFloat(10))
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
}
