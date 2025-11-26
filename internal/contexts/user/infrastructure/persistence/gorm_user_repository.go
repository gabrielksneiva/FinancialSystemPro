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

// GormUserRepository implementa UserRepository usando GORM com mappers
type GormUserRepository struct {
	db     *gorm.DB
	mapper mappers.UserMapper
}

// NewGormUserRepository cria um novo repositório de usuários usando GORM
func NewGormUserRepository(db *gorm.DB) repository.UserRepository {
	return &GormUserRepository{
		db:     db,
		mapper: mappers.UserMapper{},
	}
}

// Create insere um novo usuário no banco
func (r *GormUserRepository) Create(ctx context.Context, user *entity.User) error {
	model := r.mapper.ToModel(user)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID busca um usuário por ID
func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var model models.UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model)
}

// FindByEmail busca um usuário por email
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model models.UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model)
}

// Update atualiza um usuário existente
func (r *GormUserRepository) Update(ctx context.Context, user *entity.User) error {
	model := r.mapper.ToModel(user)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete remove um usuário
func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.UserModel{}, "id = ?", id).Error
}
