package persistence

import (
	"context"
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/contexts/transaction/domain/repository"
	"financial-system-pro/internal/infrastructure/database/mappers"
	"financial-system-pro/internal/infrastructure/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormTransactionRepository implementa TransactionRepository usando GORM com mappers
type GormTransactionRepository struct {
	db     *gorm.DB
	mapper mappers.TransactionMapper
}

// NewGormTransactionRepository cria um novo repositório de transações usando GORM
func NewGormTransactionRepository(db *gorm.DB) repository.TransactionRepository {
	return &GormTransactionRepository{
		db:     db,
		mapper: mappers.TransactionMapper{},
	}
}

// Create insere uma nova transação no banco
func (r *GormTransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	model := r.mapper.ToModel(tx)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID busca uma transação por ID
func (r *GormTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var model models.TransactionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model), nil
}

// FindByUserID busca todas as transações de um usuário
func (r *GormTransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	var models []*models.TransactionModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return r.mapper.ToDomainList(models), nil
}

// FindByHash busca uma transação por hash
func (r *GormTransactionRepository) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	var model models.TransactionModel
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model), nil
}

// Update atualiza uma transação existente
func (r *GormTransactionRepository) Update(ctx context.Context, tx *entity.Transaction) error {
	model := r.mapper.ToModel(tx)
	return r.db.WithContext(ctx).Save(model).Error
}

// UpdateStatus atualiza apenas o status de uma transação
func (r *GormTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.TransactionModel{}).
		Where("id = ?", id).
		Update("status", string(status)).
		Error
}
