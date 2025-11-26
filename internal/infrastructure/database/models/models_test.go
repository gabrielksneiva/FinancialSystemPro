package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserModel_TableName(t *testing.T) {
	model := UserModel{}
	assert.Equal(t, "user_context.users", model.TableName())
}

func TestUserModel_Fields(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	model := UserModel{
		ID:        id,
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, id, model.ID)
	assert.Equal(t, "test@example.com", model.Email)
	assert.Equal(t, "hashedpassword", model.Password)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now, model.UpdatedAt)
}

func TestUserModel_DefaultValues(t *testing.T) {
	model := UserModel{}

	assert.Equal(t, uuid.Nil, model.ID)
	assert.Empty(t, model.Email)
	assert.Empty(t, model.Password)
	assert.True(t, model.CreatedAt.IsZero())
	assert.True(t, model.UpdatedAt.IsZero())
}

func TestWalletModel_TableName(t *testing.T) {
	model := WalletModel{}
	assert.Equal(t, "user_context.wallets", model.TableName())
}

func TestWalletModel_Fields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	model := WalletModel{
		ID:               id,
		UserID:           userID,
		Address:          "0xABCDEF1234567890",
		EncryptedPrivKey: "encrypted_key_data",
		Balance:          100.50,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, id, model.ID)
	assert.Equal(t, userID, model.UserID)
	assert.Equal(t, "0xABCDEF1234567890", model.Address)
	assert.Equal(t, "encrypted_key_data", model.EncryptedPrivKey)
	assert.Equal(t, 100.50, model.Balance)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now, model.UpdatedAt)
}

func TestWalletModel_DefaultValues(t *testing.T) {
	model := WalletModel{}

	assert.Equal(t, uuid.Nil, model.ID)
	assert.Equal(t, uuid.Nil, model.UserID)
	assert.Empty(t, model.Address)
	assert.Empty(t, model.EncryptedPrivKey)
	assert.Equal(t, 0.0, model.Balance)
	assert.True(t, model.CreatedAt.IsZero())
	assert.True(t, model.UpdatedAt.IsZero())
}

func TestUserModel_Instantiation(t *testing.T) {
	model := &UserModel{
		Email:    "user@domain.com",
		Password: "secure_hash",
	}

	assert.NotNil(t, model)
	assert.Equal(t, "user@domain.com", model.Email)
	assert.Equal(t, "secure_hash", model.Password)
}

func TestWalletModel_Instantiation(t *testing.T) {
	userID := uuid.New()

	model := &WalletModel{
		UserID:           userID,
		Address:          "wallet_address_123",
		EncryptedPrivKey: "encrypted_private_key",
		Balance:          0.0,
	}

	assert.NotNil(t, model)
	assert.Equal(t, userID, model.UserID)
	assert.Equal(t, "wallet_address_123", model.Address)
	assert.Equal(t, 0.0, model.Balance)
}

func TestWalletModel_BalanceTypes(t *testing.T) {
	model := WalletModel{Balance: 123.456}
	assert.Equal(t, 123.456, model.Balance)

	model.Balance = 0
	assert.Equal(t, 0.0, model.Balance)

	model.Balance = -50.00
	assert.Equal(t, -50.0, model.Balance)
}
