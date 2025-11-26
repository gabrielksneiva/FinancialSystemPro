package user

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository define operações de persistência para agregados User.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hashed HashedPassword) error
}

// WalletRepository define operações relacionadas a wallets de usuário.
type WalletRepository interface {
	AttachWallet(ctx context.Context, userID uuid.UUID, walletID uuid.UUID) error
	ListWalletIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
