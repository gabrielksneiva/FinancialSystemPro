package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
)

// OnChainWalletRepositoryPort define operações de persistência para carteiras multi-chain
type OnChainWalletRepositoryPort interface {
	Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, encryptedPrivKey string) error
	FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error)
	Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error)
}

type onChainWalletRepositoryAdapter struct{ db *repositories.NewDatabase }

func NewOnChainWalletRepositoryAdapter(db *repositories.NewDatabase) OnChainWalletRepositoryPort {
	return &onChainWalletRepositoryAdapter{db: db}
}

func (a *onChainWalletRepositoryAdapter) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, encryptedPrivKey string) error {
	record := &repositories.OnChainWallet{
		ID:               uuid.New(),
		UserID:           userID,
		Blockchain:       string(info.Blockchain),
		Address:          info.Address,
		PublicKey:        info.PublicKey,
		EncryptedPrivKey: encryptedPrivKey,
	}
	return a.db.DB.WithContext(ctx).Create(record).Error
}

func (a *onChainWalletRepositoryAdapter) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	var w repositories.OnChainWallet
	err := a.db.DB.WithContext(ctx).Where("user_id = ? AND blockchain = ?", userID, string(chain)).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (a *onChainWalletRepositoryAdapter) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	var list []*repositories.OnChainWallet
	err := a.db.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func (a *onChainWalletRepositoryAdapter) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	var count int64
	err := a.db.DB.WithContext(ctx).Model(&repositories.OnChainWallet{}).Where("user_id = ? AND blockchain = ?", userID, string(chain)).Count(&count).Error
	return count > 0, err
}
