package persistence

import (
	"context"
	"errors"
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/infrastructure/database/mappers"
	"financial-system-pro/internal/infrastructure/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormWalletRepository implementa WalletRepository usando GORM com mappers
type GormWalletRepository struct {
	db     *gorm.DB
	mapper mappers.WalletMapper
}

// NewGormWalletRepository cria um novo repositório de wallets usando GORM
func NewGormWalletRepository(db *gorm.DB) repository.WalletRepository {
	return &GormWalletRepository{
		db:     db,
		mapper: mappers.WalletMapper{},
	}
}

// Create insere uma nova wallet no banco
func (r *GormWalletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	model := r.mapper.ToModel(wallet)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByUserID busca a wallet de um usuário
func (r *GormWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	var model models.WalletModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model), nil
}

// FindByAddress busca uma wallet por endereço
func (r *GormWalletRepository) FindByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	var model models.WalletModel
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model), nil
}

// UpdateBalance atualiza o saldo de uma wallet
func (r *GormWalletRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	return r.db.WithContext(ctx).
		Model(&models.WalletModel{}).
		Where("user_id = ?", userID).
		Update("balance", balance).
		Error
}
