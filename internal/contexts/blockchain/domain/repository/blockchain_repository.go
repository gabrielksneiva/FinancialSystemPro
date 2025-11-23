package repository

import (
	"context"
	"financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/google/uuid"
)

// BlockchainTransactionRepository define as operações de persistência para transações blockchain
type BlockchainTransactionRepository interface {
	Create(ctx context.Context, tx *entity.BlockchainTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.BlockchainTransaction, error)
	FindByHash(ctx context.Context, hash string) (*entity.BlockchainTransaction, error)
	FindByAddress(ctx context.Context, address string) ([]*entity.BlockchainTransaction, error)
	Update(ctx context.Context, tx *entity.BlockchainTransaction) error
	UpdateConfirmations(ctx context.Context, hash string, confirmations int) error
}

// WalletInfoRepository define as operações de persistência para informações de wallets
type WalletInfoRepository interface {
	Create(ctx context.Context, wallet *entity.WalletInfo) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.WalletInfo, error)
	FindByAddress(ctx context.Context, address string) (*entity.WalletInfo, error)
	UpdateBalance(ctx context.Context, address string, balance entity.WalletInfo) error
	UpdateNonce(ctx context.Context, address string, nonce int64) error
}
