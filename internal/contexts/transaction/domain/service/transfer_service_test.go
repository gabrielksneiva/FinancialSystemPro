package service

import (
	"context"
	"testing"

	"financial-system-pro/internal/contexts/transaction/domain/entity"
	userentity "financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/valueobject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUserAggregate(email string, initialBalance float64) *userentity.UserAggregate {
	emailVO, _ := valueobject.NewEmail(email)
	password, _ := valueobject.HashFromRaw("password123")
	agg, _ := userentity.NewUserAggregate(emailVO, password, "0xabc", "privkey")
	if initialBalance > 0 {
		_ = agg.CreditWallet(initialBalance)
	}
	return agg
}

func TestValidateTransfer(t *testing.T) {
	service := NewTransferService()
	ctx := context.Background()

	t.Run("validates correct transfer request", func(t *testing.T) {
		request := TransferRequest{
			FromUserID: uuid.New(),
			ToUserID:   uuid.New(),
			Amount:     decimal.NewFromFloat(100.0),
			Reference:  "payment",
		}

		err := service.ValidateTransfer(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("rejects empty source user", func(t *testing.T) {
		request := TransferRequest{
			FromUserID: uuid.Nil,
			ToUserID:   uuid.New(),
			Amount:     decimal.NewFromFloat(100.0),
		}

		err := service.ValidateTransfer(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source user ID")
	})

	t.Run("rejects empty destination user", func(t *testing.T) {
		request := TransferRequest{
			FromUserID: uuid.New(),
			ToUserID:   uuid.Nil,
			Amount:     decimal.NewFromFloat(100.0),
		}

		err := service.ValidateTransfer(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination user ID")
	})

	t.Run("rejects same user transfer", func(t *testing.T) {
		userID := uuid.New()
		request := TransferRequest{
			FromUserID: userID,
			ToUserID:   userID,
			Amount:     decimal.NewFromFloat(100.0),
		}

		err := service.ValidateTransfer(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "same user")
	})

	t.Run("rejects zero amount", func(t *testing.T) {
		request := TransferRequest{
			FromUserID: uuid.New(),
			ToUserID:   uuid.New(),
			Amount:     decimal.Zero,
		}

		err := service.ValidateTransfer(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("rejects negative amount", func(t *testing.T) {
		request := TransferRequest{
			FromUserID: uuid.New(),
			ToUserID:   uuid.New(),
			Amount:     decimal.NewFromFloat(-100.0),
		}

		err := service.ValidateTransfer(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})
}

func TestExecuteTransfer(t *testing.T) {
	service := NewTransferService()
	ctx := context.Background()

	t.Run("executes successful transfer", func(t *testing.T) {
		fromUser := createTestUserAggregate("from@example.com", 500.0)
		toUser := createTestUserAggregate("to@example.com", 100.0)
		amount := decimal.NewFromFloat(200.0)

		txAgg, err := service.ExecuteTransfer(ctx, fromUser, toUser, amount, "payment")

		require.NoError(t, err)
		assert.NotNil(t, txAgg)
		assert.Equal(t, 300.0, fromUser.Wallet().GetBalance())
		assert.Equal(t, 300.0, toUser.Wallet().GetBalance())
		assert.True(t, txAgg.IsCompleted())
		assert.Equal(t, entity.TransactionTypeTransfer, txAgg.GetType())
	})

	t.Run("rejects transfer with insufficient funds", func(t *testing.T) {
		fromUser := createTestUserAggregate("from@example.com", 50.0)
		toUser := createTestUserAggregate("to@example.com", 100.0)
		amount := decimal.NewFromFloat(100.0)

		txAgg, err := service.ExecuteTransfer(ctx, fromUser, toUser, amount, "payment")

		assert.Error(t, err)
		assert.Nil(t, txAgg)
		assert.Contains(t, err.Error(), "insufficient")
		// Saldo não deve mudar
		assert.Equal(t, 50.0, fromUser.Wallet().GetBalance())
		assert.Equal(t, 100.0, toUser.Wallet().GetBalance())
	})

	t.Run("rejects transfer from inactive user", func(t *testing.T) {
		fromUser := createTestUserAggregate("from@example.com", 500.0)
		toUser := createTestUserAggregate("to@example.com", 100.0)
		fromUser.User().Deactivate()
		amount := decimal.NewFromFloat(100.0)

		txAgg, err := service.ExecuteTransfer(ctx, fromUser, toUser, amount, "payment")

		assert.Error(t, err)
		assert.Nil(t, txAgg)
		// Saldo não deve mudar
		assert.Equal(t, 500.0, fromUser.Wallet().GetBalance())
		assert.Equal(t, 100.0, toUser.Wallet().GetBalance())
	})

	t.Run("rejects transfer to inactive user", func(t *testing.T) {
		fromUser := createTestUserAggregate("from@example.com", 500.0)
		toUser := createTestUserAggregate("to@example.com", 100.0)
		toUser.User().Deactivate()
		amount := decimal.NewFromFloat(100.0)

		txAgg, err := service.ExecuteTransfer(ctx, fromUser, toUser, amount, "payment")

		assert.Error(t, err)
		assert.Nil(t, txAgg)
		assert.Contains(t, err.Error(), "inactive")
		// Saldo não deve mudar
		assert.Equal(t, 500.0, fromUser.Wallet().GetBalance())
		assert.Equal(t, 100.0, toUser.Wallet().GetBalance())
	})

	t.Run("transaction has correct events", func(t *testing.T) {
		fromUser := createTestUserAggregate("from@example.com", 500.0)
		toUser := createTestUserAggregate("to@example.com", 100.0)
		amount := decimal.NewFromFloat(200.0)

		txAgg, err := service.ExecuteTransfer(ctx, fromUser, toUser, amount, "payment")

		require.NoError(t, err)
		events := txAgg.DomainEvents()
		// Created, Confirmed, Completed
		assert.Len(t, events, 3)
	})
}

func TestCalculateFee(t *testing.T) {
	service := NewTransferService()

	t.Run("internal transfers have zero fee", func(t *testing.T) {
		amount := decimal.NewFromFloat(100.0)
		fee := service.CalculateFee(amount, "internal")

		assert.True(t, fee.IsZero())
	})

	t.Run("external transfers have 1% fee", func(t *testing.T) {
		amount := decimal.NewFromFloat(100.0)
		fee := service.CalculateFee(amount, "external")

		expected := decimal.NewFromFloat(1.0)
		assert.True(t, fee.Equal(expected))
	})

	t.Run("blockchain transfers have 1% fee", func(t *testing.T) {
		amount := decimal.NewFromFloat(500.0)
		fee := service.CalculateFee(amount, "blockchain")

		expected := decimal.NewFromFloat(5.0)
		assert.True(t, fee.Equal(expected))
	})
}

func TestValidateBalance(t *testing.T) {
	service := NewTransferService()

	t.Run("validates sufficient balance for internal transfer", func(t *testing.T) {
		user := createTestUserAggregate("user@example.com", 100.0)
		amount := decimal.NewFromFloat(50.0)

		err := service.ValidateBalance(user, amount, "internal")
		assert.NoError(t, err)
	})

	t.Run("validates sufficient balance including fee", func(t *testing.T) {
		user := createTestUserAggregate("user@example.com", 100.0)
		amount := decimal.NewFromFloat(99.0)

		err := service.ValidateBalance(user, amount, "external")
		// Precisa 99 + 0.99 (1%) = 99.99, tem 100 ✓
		assert.NoError(t, err)
	})

	t.Run("rejects insufficient balance with fee", func(t *testing.T) {
		user := createTestUserAggregate("user@example.com", 100.0)
		amount := decimal.NewFromFloat(100.0)

		err := service.ValidateBalance(user, amount, "external")
		// Precisa 100 + 1 (1%) = 101, tem apenas 100 ✗
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance including fees")
	})

	t.Run("validates exact balance for internal transfer", func(t *testing.T) {
		user := createTestUserAggregate("user@example.com", 100.0)
		amount := decimal.NewFromFloat(100.0)

		err := service.ValidateBalance(user, amount, "internal")
		// Sem taxa, 100 = 100 ✓
		assert.NoError(t, err)
	})
}

func TestTransferService_Integration(t *testing.T) {
	service := NewTransferService()
	ctx := context.Background()

	t.Run("complete transfer workflow", func(t *testing.T) {
		// Setup: 3 usuários
		alice := createTestUserAggregate("alice@example.com", 1000.0)
		bob := createTestUserAggregate("bob@example.com", 500.0)
		charlie := createTestUserAggregate("charlie@example.com", 100.0)

		// 1. Alice envia 300 para Bob
		tx1, err := service.ExecuteTransfer(ctx, alice, bob, decimal.NewFromFloat(300.0), "payment-1")
		require.NoError(t, err)
		assert.Equal(t, 700.0, alice.Wallet().GetBalance())
		assert.Equal(t, 800.0, bob.Wallet().GetBalance())
		assert.True(t, tx1.IsCompleted())

		// 2. Bob envia 200 para Charlie
		tx2, err := service.ExecuteTransfer(ctx, bob, charlie, decimal.NewFromFloat(200.0), "payment-2")
		require.NoError(t, err)
		assert.True(t, tx2.IsCompleted())
		assert.Equal(t, 600.0, bob.Wallet().GetBalance())
		assert.Equal(t, 300.0, charlie.Wallet().GetBalance())

		// 3. Charlie tenta enviar 500 para Alice (deve falhar)
		tx3, err := service.ExecuteTransfer(ctx, charlie, alice, decimal.NewFromFloat(500.0), "payment-3")
		assert.Error(t, err)
		assert.Nil(t, tx3)
		// Saldos não mudam
		assert.Equal(t, 700.0, alice.Wallet().GetBalance())
		assert.Equal(t, 300.0, charlie.Wallet().GetBalance())
	})
}
