package services

import (
	"context"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
)

// WalletRepositoryPort encapsula operações de persistência de wallets Tron (incremental).
type WalletRepositoryPort interface {
	SaveInfo(ctx context.Context, userID uuid.UUID, tronAddress, encryptedPrivKey string) error
	GetInfo(ctx context.Context, userID uuid.UUID) (*repositories.WalletInfo, error)
}

type walletRepositoryAdapter struct{ db DatabasePort }

func NewWalletRepositoryAdapter(db DatabasePort) WalletRepositoryPort {
	return &walletRepositoryAdapter{db: db}
}

func (w *walletRepositoryAdapter) SaveInfo(ctx context.Context, userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return w.db.SaveWalletInfo(userID, tronAddress, encryptedPrivKey)
}
func (w *walletRepositoryAdapter) GetInfo(ctx context.Context, userID uuid.UUID) (*repositories.WalletInfo, error) {
	return w.db.GetWalletInfo(userID)
}
