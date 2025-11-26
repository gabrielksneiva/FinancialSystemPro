package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionCreated(t *testing.T) {
	txID := uuid.New()
	userID := uuid.New()
	amount := 1000.50
	currency := "USD"
	txType := "deposit"
	blockchain := "ethereum"
	walletAddr := "0x1234567890abcdef"

	event := NewTransactionCreated(txID, userID, amount, currency, txType, blockchain, walletAddr)

	assert.NotNil(t, event)
	assert.Equal(t, "TransactionCreated", event.EventType())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, amount, event.Amount)
	assert.Equal(t, currency, event.Currency)
	assert.Equal(t, txType, event.Type)
	assert.Equal(t, blockchain, event.BlockchainType)
	assert.Equal(t, walletAddr, event.WalletAddress)
	assert.Equal(t, txID, event.AggregateID())
	assert.NotEmpty(t, event.EventID())
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
}

func TestTransactionConfirmed(t *testing.T) {
	txID := uuid.New()
	hash := "0xabcdef123456"
	confirmations := 6
	blockNumber := int64(12345678)

	event := NewTransactionConfirmed(txID, hash, confirmations, blockNumber)

	assert.NotNil(t, event)
	assert.Equal(t, "TransactionConfirmed", event.EventType())
	assert.Equal(t, hash, event.Hash)
	assert.Equal(t, confirmations, event.Confirmations)
	assert.Equal(t, blockNumber, event.BlockNumber)
	assert.WithinDuration(t, time.Now(), event.ConfirmedAt, time.Second)
	assert.Equal(t, txID, event.AggregateID())
}

func TestTransactionCompleted(t *testing.T) {
	txID := uuid.New()
	hash := "0xabcdef123456"
	finalAmount := 1000.00

	event := NewTransactionCompleted(txID, hash, finalAmount)

	assert.NotNil(t, event)
	assert.Equal(t, "TransactionCompleted", event.EventType())
	assert.Equal(t, hash, event.Hash)
	assert.Equal(t, finalAmount, event.FinalAmount)
	assert.WithinDuration(t, time.Now(), event.CompletedAt, time.Second)
	assert.Equal(t, txID, event.AggregateID())
}

func TestTransactionFailed(t *testing.T) {
	txID := uuid.New()
	reason := "Insufficient funds"
	errorCode := "ERR_INSUFFICIENT_FUNDS"

	event := NewTransactionFailed(txID, reason, errorCode)

	assert.NotNil(t, event)
	assert.Equal(t, "TransactionFailed", event.EventType())
	assert.Equal(t, reason, event.Reason)
	assert.Equal(t, errorCode, event.ErrorCode)
	assert.WithinDuration(t, time.Now(), event.FailedAt, time.Second)
	assert.Equal(t, txID, event.AggregateID())
}

func TestTransactionEventImplementsDomainEvent(t *testing.T) {
	txID := uuid.New()
	userID := uuid.New()

	t.Run("TransactionCreated implements DomainEvent", func(t *testing.T) {
		event := NewTransactionCreated(txID, userID, 100, "USD", "deposit", "ethereum", "0xabc")
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, txID, event.AggregateID())
	})

	t.Run("TransactionConfirmed implements DomainEvent", func(t *testing.T) {
		event := NewTransactionConfirmed(txID, "0xhash", 6, 12345)
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, txID, event.AggregateID())
	})

	t.Run("TransactionCompleted implements DomainEvent", func(t *testing.T) {
		event := NewTransactionCompleted(txID, "0xhash", 1000.0)
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, txID, event.AggregateID())
	})

	t.Run("TransactionFailed implements DomainEvent", func(t *testing.T) {
		event := NewTransactionFailed(txID, "Error", "ERR_CODE")
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, txID, event.AggregateID())
	})
}
