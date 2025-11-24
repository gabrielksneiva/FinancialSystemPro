package inmemory

import (
	"context"
	"sync"

	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
)

type WalletRepository struct {
	mu        sync.RWMutex
	wallets   map[uuid.UUID]*userEntity.Wallet
	byAddress map[string]*userEntity.Wallet
}

func NewWalletRepository() *WalletRepository {
	return &WalletRepository{
		wallets:   make(map[uuid.UUID]*userEntity.Wallet),
		byAddress: make(map[string]*userEntity.Wallet),
	}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *userEntity.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.wallets[wallet.UserID]; exists {
		return errors.NewValidationError("wallet", "wallet already exists for user")
	}
	r.wallets[wallet.UserID] = wallet
	r.byAddress[wallet.Address] = wallet
	return nil
}

func (r *WalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wallet, exists := r.wallets[userID]
	if !exists {
		return nil, errors.NewNotFoundError("wallet")
	}
	return wallet, nil
}

func (r *WalletRepository) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wallet, exists := r.byAddress[address]
	if !exists {
		return nil, errors.NewNotFoundError("wallet")
	}
	return wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	wallet, exists := r.wallets[userID]
	if !exists {
		return errors.NewNotFoundError("wallet")
	}
	wallet.Balance = balance
	return nil
}
