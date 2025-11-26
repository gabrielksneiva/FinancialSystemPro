package mappers

import (
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/valueobject"
	"financial-system-pro/internal/infrastructure/database/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserMapper_ToModel(t *testing.T) {
	mapper := UserMapper{}

	email, err := valueobject.NewEmail("test@example.com")
	require.NoError(t, err)

	password, err := valueobject.HashFromRaw("password123")
	require.NoError(t, err)

	user := &entity.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	model := mapper.ToModel(user)

	assert.NotNil(t, model)
	assert.Equal(t, user.ID, model.ID)
	assert.Equal(t, user.Email.String(), model.Email)
	assert.Equal(t, user.Password.String(), model.Password)
	assert.Equal(t, user.CreatedAt, model.CreatedAt)
	assert.Equal(t, user.UpdatedAt, model.UpdatedAt)
}

func TestUserMapper_ToModel_Nil(t *testing.T) {
	mapper := UserMapper{}
	model := mapper.ToModel(nil)
	assert.Nil(t, model)
}

func TestUserMapper_ToDomain(t *testing.T) {
	mapper := UserMapper{}

	email, err := valueobject.NewEmail("test@example.com")
	require.NoError(t, err)

	password, err := valueobject.HashFromRaw("password123")
	require.NoError(t, err)

	model := &models.UserModel{
		ID:        uuid.New(),
		Email:     email.String(),
		Password:  password.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := mapper.ToDomain(model)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, model.ID, user.ID)
	assert.Equal(t, model.Email, user.Email.String())
	assert.Equal(t, model.Password, user.Password.String())
	assert.Equal(t, model.CreatedAt, user.CreatedAt)
	assert.Equal(t, model.UpdatedAt, user.UpdatedAt)
}

func TestUserMapper_ToDomain_Nil(t *testing.T) {
	mapper := UserMapper{}
	user, err := mapper.ToDomain(nil)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserMapper_ToDomain_InvalidEmail(t *testing.T) {
	mapper := UserMapper{}

	model := &models.UserModel{
		ID:        uuid.New(),
		Email:     "invalid-email", // email inválido
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := mapper.ToDomain(model)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestWalletMapper_ToModel(t *testing.T) {
	mapper := WalletMapper{}

	wallet := &entity.Wallet{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		Address:          "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC",
		EncryptedPrivKey: "encrypted_priv_key",
		Balance:          100.50,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	model := mapper.ToModel(wallet)

	assert.NotNil(t, model)
	assert.Equal(t, wallet.ID, model.ID)
	assert.Equal(t, wallet.UserID, model.UserID)
	assert.Equal(t, wallet.Address, model.Address)
	assert.Equal(t, wallet.EncryptedPrivKey, model.EncryptedPrivKey)
	assert.Equal(t, wallet.Balance, model.Balance)
	assert.Equal(t, wallet.CreatedAt, model.CreatedAt)
	assert.Equal(t, wallet.UpdatedAt, model.UpdatedAt)
}

func TestWalletMapper_ToModel_Nil(t *testing.T) {
	mapper := WalletMapper{}
	model := mapper.ToModel(nil)
	assert.Nil(t, model)
}

func TestWalletMapper_ToDomain(t *testing.T) {
	mapper := WalletMapper{}

	model := &models.WalletModel{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		Address:          "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC",
		EncryptedPrivKey: "encrypted_priv_key",
		Balance:          100.50,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	wallet := mapper.ToDomain(model)

	assert.NotNil(t, wallet)
	assert.Equal(t, model.ID, wallet.ID)
	assert.Equal(t, model.UserID, wallet.UserID)
	assert.Equal(t, model.Address, wallet.Address)
	assert.Equal(t, model.EncryptedPrivKey, wallet.EncryptedPrivKey)
	assert.Equal(t, model.Balance, wallet.Balance)
	assert.Equal(t, model.CreatedAt, wallet.CreatedAt)
	assert.Equal(t, model.UpdatedAt, wallet.UpdatedAt)
}

func TestWalletMapper_ToDomain_Nil(t *testing.T) {
	mapper := WalletMapper{}
	wallet := mapper.ToDomain(nil)
	assert.Nil(t, wallet)
}

func TestUserWalletMapper_RoundTrip(t *testing.T) {
	userMapper := UserMapper{}
	walletMapper := WalletMapper{}

	// Create original user
	email, err := valueobject.NewEmail("roundtrip@example.com")
	require.NoError(t, err)

	password, err := valueobject.HashFromRaw("secure123")
	require.NoError(t, err)

	originalUser := &entity.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}

	// User: Domain -> Model -> Domain
	userModel := userMapper.ToModel(originalUser)
	reconstructedUser, err := userMapper.ToDomain(userModel)
	require.NoError(t, err)

	assert.Equal(t, originalUser.ID, reconstructedUser.ID)
	assert.Equal(t, originalUser.Email.String(), reconstructedUser.Email.String())
	assert.Equal(t, originalUser.Password.String(), reconstructedUser.Password.String())
	assert.Equal(t, originalUser.CreatedAt, reconstructedUser.CreatedAt)
	assert.Equal(t, originalUser.UpdatedAt, reconstructedUser.UpdatedAt)

	// Create original wallet
	originalWallet := &entity.Wallet{
		ID:               uuid.New(),
		UserID:           originalUser.ID,
		Address:          "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC",
		EncryptedPrivKey: "encrypted_key",
		Balance:          250.75,
		CreatedAt:        time.Now().Truncate(time.Microsecond),
		UpdatedAt:        time.Now().Truncate(time.Microsecond),
	}

	// Wallet: Domain -> Model -> Domain
	walletModel := walletMapper.ToModel(originalWallet)
	reconstructedWallet := walletMapper.ToDomain(walletModel)

	assert.Equal(t, originalWallet.ID, reconstructedWallet.ID)
	assert.Equal(t, originalWallet.UserID, reconstructedWallet.UserID)
	assert.Equal(t, originalWallet.Address, reconstructedWallet.Address)
	assert.Equal(t, originalWallet.EncryptedPrivKey, reconstructedWallet.EncryptedPrivKey)
	assert.Equal(t, originalWallet.Balance, reconstructedWallet.Balance)
	assert.Equal(t, originalWallet.CreatedAt, reconstructedWallet.CreatedAt)
	assert.Equal(t, originalWallet.UpdatedAt, reconstructedWallet.UpdatedAt)
}
