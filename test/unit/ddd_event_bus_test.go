package http_test

import (
	"context"
	"testing"

	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEventBusDepositCompleted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	bus := events.NewInMemoryBus(logger)
	handlers := services.NewEventHandlers(logger)

	// Subscribe ao evento de depósito
	depositHandled := false
	bus.Subscribe("deposit.completed", func(ctx context.Context, e events.Event) error {
		depositHandled = true
		return handlers.OnDepositCompleted(ctx, e)
	})

	// Publicar evento
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.50)
	event := events.NewDepositCompletedEvent(userID, amount, "tx-123")

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.True(t, depositHandled, "Deposit handler should have been called")
}

func TestEventBusWithdrawCompleted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	bus := events.NewInMemoryBus(logger)
	handlers := services.NewEventHandlers(logger)

	withdrawHandled := false
	bus.Subscribe("withdraw.completed", func(ctx context.Context, e events.Event) error {
		withdrawHandled = true
		return handlers.OnWithdrawCompleted(ctx, e)
	})

	userID := uuid.New()
	amount := decimal.NewFromFloat(50.00)
	event := events.NewWithdrawCompletedEvent(userID, amount, "tx-456")

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.True(t, withdrawHandled, "Withdraw handler should have been called")
}

func TestSetupEventSubscribersIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	bus := events.NewInMemoryBus(logger)

	// Configura todos os subscribers
	services.SetupEventSubscribers(bus, logger)

	// Verificar que subscribers foram registrados
	userID := uuid.New()
	amount := decimal.NewFromFloat(100.00)

	// Publicar vários eventos e garantir que não há erro
	eventsToPublish := []events.Event{
		events.NewDepositCompletedEvent(userID, amount, "tx-1"),
		events.NewWithdrawCompletedEvent(userID, amount, "tx-2"),
		events.NewTransferCompletedEvent(userID, uuid.New(), amount, "tx-3"),
		events.NewTransactionFailedEvent(userID, "deposit", amount, "error", "ERROR_CODE"),
	}

	for _, evt := range eventsToPublish {
		err := bus.Publish(context.Background(), evt)
		assert.NoError(t, err, "Should publish event without error")
	}
}
