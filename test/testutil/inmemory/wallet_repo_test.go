package inmemory

import (
	"context"
	"testing"
	"time"

	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	domainErrors "financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestWallet(userID uuid.UUID) *userEntity.Wallet {
	return &userEntity.Wallet{
		ID:               uuid.New(),
		UserID:           userID,
		Address:          "ADDR-" + uuid.New().String(),
		EncryptedPrivKey: "enc",
		Balance:          0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func TestWalletRepositoryFlow(t *testing.T) {
	repo := NewWalletRepository()
	ctx := context.Background()
	uid := uuid.New()

	w := newTestWallet(uid)
	require.NoError(t, repo.Create(ctx, w))

	// find by user id
	fetched, err := repo.FindByUserID(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, w.Address, fetched.Address)

	// find by address
	byAddr, err := repo.FindByAddress(ctx, w.Address)
	require.NoError(t, err)
	require.Equal(t, w.ID, byAddr.ID)

	// duplicate wallet for same user
	dup := newTestWallet(uid)
	err = repo.Create(ctx, dup)
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrValidation, err.(*domainErrors.AppError).Code)

	// update balance
	require.NoError(t, repo.UpdateBalance(ctx, uid, 42.5))
	fetched2, err := repo.FindByUserID(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, 42.5, fetched2.Balance)
}

func TestWalletRepositoryNotFound(t *testing.T) {
	repo := NewWalletRepository()
	ctx := context.Background()
	_, err := repo.FindByUserID(ctx, uuid.New())
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
	err = repo.UpdateBalance(ctx, uuid.New(), 10)
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
}
