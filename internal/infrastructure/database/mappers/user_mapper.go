package mappers

import (
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/valueobject"
	"financial-system-pro/internal/infrastructure/database/models"
)

// UserMapper converte entre entidade de domínio User e UserModel (GORM)
type UserMapper struct{}

// ToModel converte entidade de domínio para modelo GORM
func (m UserMapper) ToModel(user *entity.User) *models.UserModel {
	if user == nil {
		return nil
	}

	return &models.UserModel{
		ID:        user.ID,
		Email:     user.Email.String(),
		Password:  user.Password.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToDomain converte modelo GORM para entidade de domínio
func (m UserMapper) ToDomain(model *models.UserModel) (*entity.User, error) {
	if model == nil {
		return nil, nil
	}

	email, err := valueobject.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}

	// Password já está hasheado no banco, então apenas convertemos
	password := valueobject.HashedPassword(model.Password)

	return &entity.User{
		ID:        model.ID,
		Email:     email,
		Password:  password,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

// WalletMapper converte entre entidade de domínio Wallet e WalletModel (GORM)
type WalletMapper struct{}

// ToModel converte entidade de domínio para modelo GORM
func (m WalletMapper) ToModel(wallet *entity.Wallet) *models.WalletModel {
	if wallet == nil {
		return nil
	}

	return &models.WalletModel{
		ID:               wallet.ID,
		UserID:           wallet.UserID,
		Address:          wallet.Address,
		EncryptedPrivKey: wallet.EncryptedPrivKey,
		Balance:          wallet.Balance,
		CreatedAt:        wallet.CreatedAt,
		UpdatedAt:        wallet.UpdatedAt,
	}
}

// ToDomain converte modelo GORM para entidade de domínio
func (m WalletMapper) ToDomain(model *models.WalletModel) *entity.Wallet {
	if model == nil {
		return nil
	}

	return &entity.Wallet{
		ID:               model.ID,
		UserID:           model.UserID,
		Address:          model.Address,
		EncryptedPrivKey: model.EncryptedPrivKey,
		Balance:          model.Balance,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}
