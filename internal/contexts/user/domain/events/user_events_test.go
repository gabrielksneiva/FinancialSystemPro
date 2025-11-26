package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreated(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	walletAddr := "0x1234567890abcdef"

	event := NewUserCreated(userID, email, walletAddr)

	assert.NotNil(t, event)
	assert.Equal(t, "UserCreated", event.EventType())
	assert.Equal(t, email, event.Email)
	assert.Equal(t, walletAddr, event.WalletAddress)
	assert.Equal(t, 0.0, event.WalletBalance)
	assert.Equal(t, userID, event.AggregateID())
	assert.NotEmpty(t, event.EventID())
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
}

func TestUserDeactivated(t *testing.T) {
	userID := uuid.New()
	reason := "Account suspended"

	event := NewUserDeactivated(userID, reason)

	assert.NotNil(t, event)
	assert.Equal(t, "UserDeactivated", event.EventType())
	assert.Equal(t, reason, event.Reason)
	assert.Equal(t, userID, event.AggregateID())
}

func TestUserActivated(t *testing.T) {
	userID := uuid.New()

	event := NewUserActivated(userID)

	assert.NotNil(t, event)
	assert.Equal(t, "UserActivated", event.EventType())
	assert.Equal(t, userID, event.AggregateID())
}

func TestPasswordChanged(t *testing.T) {
	userID := uuid.New()

	event := NewPasswordChanged(userID)

	assert.NotNil(t, event)
	assert.Equal(t, "PasswordChanged", event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.WithinDuration(t, time.Now(), event.ChangedAt, time.Second)
}

func TestWalletCredited(t *testing.T) {
	userID := uuid.New()
	amount := 100.50
	newBalance := 250.75
	reference := "deposit-123"

	event := NewWalletCredited(userID, amount, newBalance, reference)

	assert.NotNil(t, event)
	assert.Equal(t, "WalletCredited", event.EventType())
	assert.Equal(t, amount, event.Amount)
	assert.Equal(t, newBalance, event.NewBalance)
	assert.Equal(t, reference, event.Reference)
	assert.Equal(t, userID, event.AggregateID())
}

func TestWalletDebited(t *testing.T) {
	userID := uuid.New()
	amount := 50.25
	newBalance := 200.50
	reference := "withdrawal-456"

	event := NewWalletDebited(userID, amount, newBalance, reference)

	assert.NotNil(t, event)
	assert.Equal(t, "WalletDebited", event.EventType())
	assert.Equal(t, amount, event.Amount)
	assert.Equal(t, newBalance, event.NewBalance)
	assert.Equal(t, reference, event.Reference)
	assert.Equal(t, userID, event.AggregateID())
}

func TestEventImplementsDomainEvent(t *testing.T) {
	userID := uuid.New()

	t.Run("UserCreated implements DomainEvent", func(t *testing.T) {
		event := NewUserCreated(userID, "test@example.com", "0xabc")
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, userID, event.AggregateID())
	})

	t.Run("WalletCredited implements DomainEvent", func(t *testing.T) {
		event := NewWalletCredited(userID, 100, 100, "ref")
		require.NotEmpty(t, event.EventID())
		require.NotEmpty(t, event.EventType())
		require.NotZero(t, event.OccurredAt())
		require.Equal(t, userID, event.AggregateID())
	})
}
