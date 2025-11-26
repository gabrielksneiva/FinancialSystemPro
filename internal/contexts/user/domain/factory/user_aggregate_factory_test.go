package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFactory() *UserAggregateFactory {
	return NewUserAggregateFactory()
}

func TestNewUserAggregateFactory(t *testing.T) {
	factory := setupFactory()
	assert.NotNil(t, factory)
}

func TestUserAggregateFactory_Create(t *testing.T) {
	factory := setupFactory()

	t.Run("create with all fields", func(t *testing.T) {
		aggregate, err := factory.Create(
			"test@example.com",
			"SecurePass123!",
			"0xAddress123",
			"encrypted-priv-key",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.NotNil(t, aggregate.User())
		assert.NotNil(t, aggregate.Wallet())
		assert.Equal(t, "0xAddress123", aggregate.Wallet().Address)
		assert.Equal(t, "encrypted-priv-key", aggregate.Wallet().EncryptedPrivKey)
	})

	t.Run("reject invalid email", func(t *testing.T) {
		aggregate, err := factory.Create("invalid", "SecurePass123!", "", "")
		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})
}

func TestUserAggregateFactory_CreateSimple(t *testing.T) {
	factory := setupFactory()

	t.Run("create simple aggregate successfully", func(t *testing.T) {
		aggregate, err := factory.CreateSimple("test@example.com", "SecurePass123!")

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.NotNil(t, aggregate.User())
		assert.NotNil(t, aggregate.Wallet())
		assert.True(t, aggregate.User().IsActive())
		assert.Equal(t, 0.0, aggregate.Wallet().GetBalance())
	})

	t.Run("reject invalid email", func(t *testing.T) {
		aggregate, err := factory.CreateSimple("invalid-email", "SecurePass123!")

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject short password", func(t *testing.T) {
		aggregate, err := factory.CreateSimple("test@example.com", "short")

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject empty email", func(t *testing.T) {
		aggregate, err := factory.CreateSimple("", "SecurePass123!")

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})

	t.Run("reject empty password", func(t *testing.T) {
		aggregate, err := factory.CreateSimple("test@example.com", "")

		assert.Error(t, err)
		assert.Nil(t, aggregate)
	})
}

func TestUserAggregateFactory_CreateWithWallet(t *testing.T) {
	factory := setupFactory()

	t.Run("create with wallet successfully", func(t *testing.T) {
		aggregate, err := factory.CreateWithWallet(
			"wallet@example.com",
			"SecurePass123!",
			"TAddress123ABC",
			"encrypted-key-123",
		)

		require.NoError(t, err)
		assert.NotNil(t, aggregate)
		assert.Equal(t, "TAddress123ABC", aggregate.Wallet().Address)
		assert.Equal(t, "encrypted-key-123", aggregate.Wallet().EncryptedPrivKey)
	})

	t.Run("reject empty wallet address", func(t *testing.T) {
		aggregate, err := factory.CreateWithWallet(
			"wallet@example.com",
			"SecurePass123!",
			"",
			"encrypted-key-123",
		)

		assert.Error(t, err)
		assert.Nil(t, aggregate)
		assert.Contains(t, err.Error(), "wallet address")
	})
}

func TestUserAggregateFactory_Integration(t *testing.T) {
	factory := setupFactory()

	// Create multiple users with different methods
	user1, err := factory.CreateSimple("user1@example.com", "Password123!")
	require.NoError(t, err)

	user2, err := factory.CreateWithWallet(
		"user2@example.com",
		"Password123!",
		"TAddress456",
		"encrypted-key-456",
	)
	require.NoError(t, err)

	// Verify simple user has no blockchain wallet
	assert.Empty(t, user1.Wallet().Address)
	assert.Empty(t, user1.Wallet().EncryptedPrivKey)

	// Verify wallet user has wallet
	assert.NotEmpty(t, user2.Wallet().Address)
	assert.NotEmpty(t, user2.Wallet().EncryptedPrivKey)

	// Both should be active
	assert.True(t, user1.User().IsActive())
	assert.True(t, user2.User().IsActive())

	// Both should have zero balance
	assert.Equal(t, 0.0, user1.Wallet().GetBalance())
	assert.Equal(t, 0.0, user2.Wallet().GetBalance())

	// Verify different IDs
	assert.NotEqual(t, user1.User().ID, user2.User().ID)
}
