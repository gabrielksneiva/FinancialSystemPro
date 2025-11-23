package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransaction_CreatesWithCorrectDefaults(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.50)
	txType := TransactionTypeDeposit

	tx := NewTransaction(userID, txType, amount)

	assert.NotEqual(t, uuid.Nil, tx.ID, "ID deve ser gerado")
	assert.Equal(t, userID, tx.UserID, "UserID deve corresponder")
	assert.Equal(t, txType, tx.Type, "Type deve corresponder")
	assert.True(t, amount.Equal(tx.Amount), "Amount deve corresponder")
	assert.Equal(t, TransactionStatusPending, tx.Status, "Status inicial deve ser pending")
	assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt deve ser definido")
	assert.False(t, tx.UpdatedAt.IsZero(), "UpdatedAt deve ser definido")
	assert.Nil(t, tx.CompletedAt, "CompletedAt deve ser nil inicialmente")
	assert.Empty(t, tx.TransactionHash, "TransactionHash deve estar vazio")
	assert.Empty(t, tx.ErrorMessage, "ErrorMessage deve estar vazio")
}

func TestTransaction_Complete_SetsCorrectState(t *testing.T) {
	tx := NewTransaction(uuid.New(), TransactionTypeWithdraw, decimal.NewFromInt(50))
	txHash := "0xabc123def456"
	beforeComplete := time.Now()

	tx.Complete(txHash)

	assert.Equal(t, TransactionStatusCompleted, tx.Status, "Status deve ser completed")
	assert.Equal(t, txHash, tx.TransactionHash, "TransactionHash deve ser definido")
	require.NotNil(t, tx.CompletedAt, "CompletedAt não deve ser nil")
	assert.False(t, tx.CompletedAt.Before(beforeComplete), "CompletedAt deve ser após chamada")
	assert.False(t, tx.UpdatedAt.Before(beforeComplete), "UpdatedAt deve ser atualizado")
}

func TestTransaction_Complete_UpdatesTimestamps(t *testing.T) {
	tx := NewTransaction(uuid.New(), TransactionTypeDeposit, decimal.NewFromInt(100))
	originalUpdatedAt := tx.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	tx.Complete("hash-xyz")

	assert.True(t, tx.UpdatedAt.After(originalUpdatedAt), "UpdatedAt deve ser atualizado")
	require.NotNil(t, tx.CompletedAt)
	assert.True(t, tx.CompletedAt.After(originalUpdatedAt), "CompletedAt deve ser após criação")
}

func TestTransaction_Fail_SetsErrorState(t *testing.T) {
	tx := NewTransaction(uuid.New(), TransactionTypeTransfer, decimal.NewFromInt(200))
	errorMsg := "insufficient balance"
	beforeFail := time.Now()

	tx.Fail(errorMsg)

	assert.Equal(t, TransactionStatusFailed, tx.Status, "Status deve ser failed")
	assert.Equal(t, errorMsg, tx.ErrorMessage, "ErrorMessage deve ser definido")
	assert.False(t, tx.UpdatedAt.Before(beforeFail), "UpdatedAt deve ser atualizado")
}

func TestTransaction_Fail_AfterComplete_ChangesStatus(t *testing.T) {
	tx := NewTransaction(uuid.New(), TransactionTypeDeposit, decimal.NewFromInt(75))
	tx.Complete("hash-original")

	completedAt := tx.CompletedAt
	require.NotNil(t, completedAt)

	tx.Fail("rollback due to error")

	assert.Equal(t, TransactionStatusFailed, tx.Status, "Status deve mudar para failed")
	assert.Equal(t, "rollback due to error", tx.ErrorMessage)
	assert.Equal(t, completedAt, tx.CompletedAt, "CompletedAt deve permanecer (histórico)")
}

func TestTransaction_MultipleOperations_MaintainConsistency(t *testing.T) {
	tx := NewTransaction(uuid.New(), TransactionTypeWithdraw, decimal.NewFromFloat(123.45))

	// Primeira tentativa: sucesso
	tx.Complete("hash-1")
	assert.Equal(t, TransactionStatusCompleted, tx.Status)
	firstUpdatedAt := tx.UpdatedAt

	time.Sleep(5 * time.Millisecond)

	// Reprocessamento: falha
	tx.Fail("network timeout")
	assert.Equal(t, TransactionStatusFailed, tx.Status)
	assert.True(t, tx.UpdatedAt.After(firstUpdatedAt), "UpdatedAt deve refletir última mudança")

	// Segunda tentativa: sucesso novamente
	time.Sleep(5 * time.Millisecond)
	tx.Complete("hash-2")
	assert.Equal(t, TransactionStatusCompleted, tx.Status)
	assert.Equal(t, "hash-2", tx.TransactionHash, "Hash mais recente deve prevalecer")
}

func TestTransactionTypes_AreCorrectlyDefined(t *testing.T) {
	assert.Equal(t, TransactionType("deposit"), TransactionTypeDeposit)
	assert.Equal(t, TransactionType("withdraw"), TransactionTypeWithdraw)
	assert.Equal(t, TransactionType("transfer"), TransactionTypeTransfer)
}

func TestTransactionStatus_AreCorrectlyDefined(t *testing.T) {
	assert.Equal(t, TransactionStatus("pending"), TransactionStatusPending)
	assert.Equal(t, TransactionStatus("completed"), TransactionStatusCompleted)
	assert.Equal(t, TransactionStatus("failed"), TransactionStatusFailed)
}

func TestTransaction_AmountPrecision_IsMaintained(t *testing.T) {
	userID := uuid.New()
	amount := decimal.RequireFromString("0.123456789012345")

	tx := NewTransaction(userID, TransactionTypeDeposit, amount)

	assert.True(t, amount.Equal(tx.Amount), "Precisão decimal deve ser mantida")
}
