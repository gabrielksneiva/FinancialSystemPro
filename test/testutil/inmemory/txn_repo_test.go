package inmemory

import (
	"context"
	"testing"
	"time"

	txnEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	domainErrors "financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepositoryFlow(t *testing.T) {
	repo := NewTransactionRepository()
	ctx := context.Background()
	uid := uuid.New()

	tx := txnEntity.NewTransaction(uid, txnEntity.TransactionTypeDeposit, decimal.NewFromInt(100))
	tx.TransactionHash = "hash123" // ensure indexed at creation
	require.NoError(t, repo.Create(ctx, tx))

	// find by id
	fetched, err := repo.FindByID(ctx, tx.ID)
	require.NoError(t, err)
	require.Equal(t, tx.ID, fetched.ID)
	require.Equal(t, txnEntity.TransactionStatusPending, fetched.Status)

	// find by user id
	list, err := repo.FindByUserID(ctx, uid)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// find by hash
	byHash, err := repo.FindByHash(ctx, "hash123")
	require.NoError(t, err)
	require.Equal(t, tx.ID, byHash.ID)

	// update status
	require.NoError(t, repo.UpdateStatus(ctx, tx.ID, txnEntity.TransactionStatusCompleted))
	fetched2, err := repo.FindByID(ctx, tx.ID)
	require.NoError(t, err)
	require.Equal(t, txnEntity.TransactionStatusCompleted, fetched2.Status)

	// update full tx (simulate error message)
	fetched2.ErrorMessage = "none"
	fetched2.UpdatedAt = time.Now()
	require.NoError(t, repo.Update(ctx, fetched2))
	fetched3, err := repo.FindByID(ctx, tx.ID)
	require.NoError(t, err)
	require.Equal(t, "none", fetched3.ErrorMessage)
}

func TestTransactionRepositoryNotFound(t *testing.T) {
	repo := NewTransactionRepository()
	ctx := context.Background()
	_, err := repo.FindByID(ctx, uuid.New())
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
	err = repo.UpdateStatus(ctx, uuid.New(), txnEntity.TransactionStatusFailed)
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
}
