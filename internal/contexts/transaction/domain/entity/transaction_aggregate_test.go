package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionAggregate(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.50)
	currency := "USD"
	blockchain := "ethereum"
	walletAddr := "0x1234567890abcdef"

	t.Run("creates valid deposit aggregate", func(t *testing.T) {
		agg, err := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, currency, blockchain, walletAddr)

		require.NoError(t, err)
		assert.NotNil(t, agg)
		assert.NotNil(t, agg.Transaction())
		assert.Equal(t, userID, agg.GetUserID())
		assert.Equal(t, amount, agg.GetAmount())
		assert.Equal(t, TransactionTypeDeposit, agg.GetType())
		assert.Equal(t, TransactionStatusPending, agg.GetStatus())
		assert.Equal(t, walletAddr, agg.Transaction().ToAddress)
		assert.True(t, agg.IsPending())

		// Verifica evento de criação
		events := agg.DomainEvents()
		assert.Len(t, events, 1)
	})

	t.Run("creates valid withdrawal aggregate", func(t *testing.T) {
		agg, err := NewTransactionAggregate(userID, TransactionTypeWithdraw, amount, currency, blockchain, walletAddr)

		require.NoError(t, err)
		assert.Equal(t, TransactionTypeWithdraw, agg.GetType())
	})

	t.Run("rejects zero amount", func(t *testing.T) {
		agg, err := NewTransactionAggregate(userID, TransactionTypeDeposit, decimal.Zero, currency, blockchain, walletAddr)

		assert.Error(t, err)
		assert.Nil(t, agg)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("rejects negative amount", func(t *testing.T) {
		agg, err := NewTransactionAggregate(userID, TransactionTypeDeposit, decimal.NewFromFloat(-100), currency, blockchain, walletAddr)

		assert.Error(t, err)
		assert.Nil(t, agg)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("rejects nil userID", func(t *testing.T) {
		agg, err := NewTransactionAggregate(uuid.Nil, TransactionTypeDeposit, amount, currency, blockchain, walletAddr)

		assert.Error(t, err)
		assert.Nil(t, agg)
		assert.Contains(t, err.Error(), "userID cannot be empty")
	})
}

func TestTransactionAggregate_Confirm(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.0)
	agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
	agg.ClearDomainEvents() // Limpa evento de criação

	t.Run("confirms pending transaction", func(t *testing.T) {
		txHash := "0x123456789abcdef"
		confirmations := 6
		blockNumber := int64(12345678)

		err := agg.Confirm(txHash, confirmations, blockNumber)

		require.NoError(t, err)
		assert.Equal(t, txHash, agg.Transaction().TransactionHash)
		assert.True(t, agg.IsPending()) // Ainda pending até Complete

		// Verifica evento de confirmação
		events := agg.DomainEvents()
		assert.Len(t, events, 1)
	})

	t.Run("rejects empty hash", func(t *testing.T) {
		agg2, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		err := agg2.Confirm("", 6, 12345)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hash cannot be empty")
	})

	t.Run("rejects negative confirmations", func(t *testing.T) {
		agg2, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		err := agg2.Confirm("0xhash", -1, 12345)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "confirmations cannot be negative")
	})
}

func TestTransactionAggregate_Complete(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.0)

	t.Run("completes confirmed transaction", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		_ = agg.Confirm("0xhash", 6, 12345)
		agg.ClearDomainEvents()

		finalAmount := decimal.NewFromFloat(99.50) // Com taxa
		err := agg.Complete(finalAmount)

		require.NoError(t, err)
		assert.True(t, agg.IsCompleted())
		assert.False(t, agg.IsPending())
		assert.NotNil(t, agg.Transaction().CompletedAt)
		assert.Equal(t, finalAmount, agg.GetAmount())

		// Verifica evento de conclusão
		events := agg.DomainEvents()
		assert.Len(t, events, 1)
	})

	t.Run("completes with same amount when finalAmount is zero", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		_ = agg.Confirm("0xhash", 6, 12345)

		err := agg.Complete(decimal.Zero)

		require.NoError(t, err)
		assert.Equal(t, amount, agg.GetAmount()) // Mantém amount original
	})

	t.Run("rejects completion without confirmation", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")

		err := agg.Complete(amount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be confirmed")
	})

	t.Run("rejects double completion", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		_ = agg.Confirm("0xhash", 6, 12345)
		_ = agg.Complete(amount)

		err := agg.Complete(amount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already completed")
	})
}

func TestTransactionAggregate_Fail(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.0)

	t.Run("fails pending transaction", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		agg.ClearDomainEvents()

		reason := "Insufficient funds"
		errorCode := "ERR_INSUFFICIENT_FUNDS"

		err := agg.Fail(reason, errorCode)

		require.NoError(t, err)
		assert.True(t, agg.IsFailed())
		assert.False(t, agg.IsPending())
		assert.Equal(t, reason, agg.Transaction().ErrorMessage)

		// Verifica evento de falha
		events := agg.DomainEvents()
		assert.Len(t, events, 1)
	})

	t.Run("fails confirmed but not completed transaction", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		_ = agg.Confirm("0xhash", 6, 12345)

		err := agg.Fail("Network error", "ERR_NETWORK")

		require.NoError(t, err)
		assert.True(t, agg.IsFailed())
	})

	t.Run("cannot fail completed transaction", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		_ = agg.Confirm("0xhash", 6, 12345)
		_ = agg.Complete(amount)

		err := agg.Fail("Too late", "ERR_TOO_LATE")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot fail completed")
	})
}

func TestTransactionAggregate_StateValidations(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.0)

	t.Run("CanBeConfirmed", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")

		assert.True(t, agg.CanBeConfirmed())

		_ = agg.Confirm("0xhash", 6, 12345)
		assert.True(t, agg.CanBeConfirmed()) // Ainda pode receber mais confirmações

		_ = agg.Complete(amount)
		assert.False(t, agg.CanBeConfirmed()) // Completa, não pode mais confirmar
	})

	t.Run("CanBeCompleted", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")

		assert.False(t, agg.CanBeCompleted()) // Sem hash, não pode completar

		_ = agg.Confirm("0xhash", 6, 12345)
		assert.True(t, agg.CanBeCompleted()) // Com hash, pode completar

		_ = agg.Complete(amount)
		assert.False(t, agg.CanBeCompleted()) // Já completa
	})
}

func TestTransactionAggregate_DomainEvents(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.0)

	t.Run("accumulates events correctly", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		assert.Len(t, agg.DomainEvents(), 1) // Created

		_ = agg.Confirm("0xhash", 6, 12345)
		assert.Len(t, agg.DomainEvents(), 2) // Created + Confirmed

		_ = agg.Complete(amount)
		assert.Len(t, agg.DomainEvents(), 3) // Created + Confirmed + Completed
	})

	t.Run("clears events", func(t *testing.T) {
		agg, _ := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		assert.Len(t, agg.DomainEvents(), 1)

		agg.ClearDomainEvents()
		assert.Len(t, agg.DomainEvents(), 0)
	})
}

func TestTransactionAggregate_FullLifecycle(t *testing.T) {
	userID := uuid.New()
	amount := decimal.NewFromFloat(1000.50)

	t.Run("successful deposit flow", func(t *testing.T) {
		// 1. Criar transação
		agg, err := NewTransactionAggregate(userID, TransactionTypeDeposit, amount, "USD", "ethereum", "0xabc")
		require.NoError(t, err)
		assert.True(t, agg.IsPending())

		// 2. Confirmar na blockchain
		err = agg.Confirm("0x123456", 6, 12345678)
		require.NoError(t, err)
		assert.True(t, agg.IsPending())
		assert.NotEmpty(t, agg.Transaction().TransactionHash)

		// 3. Completar
		finalAmount := decimal.NewFromFloat(1000.00) // -0.50 de taxa
		err = agg.Complete(finalAmount)
		require.NoError(t, err)
		assert.True(t, agg.IsCompleted())
		assert.NotNil(t, agg.Transaction().CompletedAt)

		// 4. Verificar eventos
		events := agg.DomainEvents()
		assert.Len(t, events, 3) // Created, Confirmed, Completed
	})

	t.Run("failed withdrawal flow", func(t *testing.T) {
		// 1. Criar transação
		agg, err := NewTransactionAggregate(userID, TransactionTypeWithdraw, amount, "USD", "ethereum", "0xabc")
		require.NoError(t, err)

		// 2. Confirmar
		err = agg.Confirm("0x789abc", 3, 12345679)
		require.NoError(t, err)

		// 3. Falhar (ex: saldo insuficiente detectado)
		err = agg.Fail("Insufficient balance", "ERR_INSUFFICIENT_BALANCE")
		require.NoError(t, err)
		assert.True(t, agg.IsFailed())
		assert.NotEmpty(t, agg.Transaction().ErrorMessage)

		// 4. Não pode mais completar
		err = agg.Complete(amount)
		assert.Error(t, err)
	})
}
