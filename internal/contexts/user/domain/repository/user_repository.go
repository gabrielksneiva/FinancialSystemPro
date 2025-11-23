package repository

import (
	"context"
	"financial-system-pro/internal/contexts/user/domain/entity"

	"github.com/google/uuid"
)

// UserRepository define as operações de persistência para User
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WalletRepository define as operações de persistência para Wallet
type WalletRepository interface {
	Create(ctx context.Context, wallet *entity.Wallet) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	FindByAddress(ctx context.Context, address string) (*entity.Wallet, error)
	UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error
}
