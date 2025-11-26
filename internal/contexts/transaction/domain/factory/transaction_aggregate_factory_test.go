package factory

import (
	"testing"

	"financial-system-pro/internal/contexts/transaction/domain/entity"
	domainEntities "financial-system-pro/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionAggregateFactory(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	assert.NotNil(t, factory)
}

func TestTransactionAggregateFactory_CreateDeposit(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.50)

	t.Run("create valid deposit", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TAbcDef123456",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.Equal(t, userID, aggregate.Transaction().UserID)
		assert.Equal(t, entity.TransactionTypeDeposit, aggregate.Transaction().Type)
		assert.True(t, amount.Equal(aggregate.Transaction().Amount))
		assert.Len(t, aggregate.DomainEvents(), 1)
	})

	t.Run("reject nil userID", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			uuid.Nil,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TAbcDef123456",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
		assert.Contains(t, err.Error(), "userID")
	})

	t.Run("reject zero amount", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			decimal.Zero,
			"USD",
			domainEntities.BlockchainTRON,
			"TAbcDef123456",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject negative amount", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			decimal.NewFromFloat(-10),
			"USD",
			domainEntities.BlockchainTRON,
			"TAbcDef123456",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject empty blockchain", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			amount,
			"USD",
			"",
			"TAbcDef123456",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject empty wallet address", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("use default currency if empty", func(t *testing.T) {
		aggregate, err := factory.CreateDeposit(
			userID,
			amount,
			"",
			domainEntities.BlockchainTRON,
			"TAbcDef123456",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
	})
}

func TestTransactionAggregateFactory_CreateWithdrawal(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	userID := uuid.New()
	amount := decimal.NewFromFloat(50.75)

	t.Run("create valid withdrawal", func(t *testing.T) {
		aggregate, err := factory.CreateWithdrawal(
			userID,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TExternalAddress123",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.Equal(t, userID, aggregate.Transaction().UserID)
		assert.Equal(t, entity.TransactionTypeWithdraw, aggregate.Transaction().Type)
		assert.True(t, amount.Equal(aggregate.Transaction().Amount))
	})

	t.Run("reject empty destination address", func(t *testing.T) {
		aggregate, err := factory.CreateWithdrawal(
			userID,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
		assert.Contains(t, err.Error(), "destination address")
	})

	t.Run("reject empty blockchain", func(t *testing.T) {
		aggregate, err := factory.CreateWithdrawal(
			userID,
			amount,
			"USD",
			"",
			"TExternalAddress123",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})
}

func TestTransactionAggregateFactory_CreateTransfer(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	fromUserID := uuid.New()
	toUserID := uuid.New()
	amount := decimal.NewFromFloat(25.00)

	t.Run("create valid transfer", func(t *testing.T) {
		aggregate, err := factory.CreateTransfer(
			fromUserID,
			amount,
			"USD",
			toUserID,
			"payment for services",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.Equal(t, fromUserID, aggregate.Transaction().UserID)
		assert.Equal(t, entity.TransactionTypeTransfer, aggregate.Transaction().Type)
		assert.True(t, amount.Equal(aggregate.Transaction().Amount))
	})

	t.Run("reject same user transfer", func(t *testing.T) {
		sameID := uuid.New()
		aggregate, err := factory.CreateTransfer(
			sameID,
			amount,
			"USD",
			sameID,
			"self-transfer",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
		assert.Contains(t, err.Error(), "same user")
	})

	t.Run("reject nil destination userID", func(t *testing.T) {
		aggregate, err := factory.CreateTransfer(
			fromUserID,
			amount,
			"USD",
			uuid.Nil,
			"payment",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})
}

func TestTransactionAggregateFactory_CreateBlockchainTransaction(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	userID := uuid.New()
	amount := decimal.NewFromFloat(75.50)

	t.Run("create valid blockchain deposit", func(t *testing.T) {
		aggregate, err := factory.CreateBlockchainTransaction(
			userID,
			entity.TransactionTypeDeposit,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TAddress123",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
	})

	t.Run("create valid blockchain withdrawal", func(t *testing.T) {
		aggregate, err := factory.CreateBlockchainTransaction(
			userID,
			entity.TransactionTypeWithdraw,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TAddress123",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
	})

	t.Run("reject invalid transaction type", func(t *testing.T) {
		aggregate, err := factory.CreateBlockchainTransaction(
			userID,
			entity.TransactionTypeTransfer,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"TAddress123",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
		assert.Contains(t, err.Error(), "invalid transaction type")
	})

	t.Run("reject empty address", func(t *testing.T) {
		aggregate, err := factory.CreateBlockchainTransaction(
			userID,
			entity.TransactionTypeDeposit,
			amount,
			"USD",
			domainEntities.BlockchainTRON,
			"",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})
}

func TestTransactionAggregateFactory_Integration(t *testing.T) {
	factory := NewTransactionAggregateFactory()
	userID := uuid.New()
	otherUserID := uuid.New()

	// Create deposit
	deposit, err := factory.CreateDeposit(
		userID,
		decimal.NewFromFloat(100),
		"USD",
		domainEntities.BlockchainTRON,
		"TAddress1",
	)
	require.NoError(t, err)

	// Create withdrawal
	withdrawal, err := factory.CreateWithdrawal(
		userID,
		decimal.NewFromFloat(50),
		"USD",
		domainEntities.BlockchainTRON,
		"TAddress2",
	)
	require.NoError(t, err)

	// Create transfer
	transfer, err := factory.CreateTransfer(
		userID,
		decimal.NewFromFloat(25),
		"USD",
		otherUserID,
		"payment",
	)
	require.NoError(t, err)

	// Verify all created successfully
	assert.NotNil(t, deposit)
	assert.NotNil(t, withdrawal)
	assert.NotNil(t, transfer)

	// Verify types
	assert.Equal(t, entity.TransactionTypeDeposit, deposit.Transaction().Type)
	assert.Equal(t, entity.TransactionTypeWithdraw, withdrawal.Transaction().Type)
	assert.Equal(t, entity.TransactionTypeTransfer, transfer.Transaction().Type)

	// Verify all have events
	assert.NotEmpty(t, deposit.DomainEvents())
	assert.NotEmpty(t, withdrawal.DomainEvents())
	assert.NotEmpty(t, transfer.DomainEvents())
}
