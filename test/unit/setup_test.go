package http_test

import (
	"financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/test/mocks"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// setupTestHandler cria um handler de teste com mocks
func setupTestHandler() (*http.NewHandler, *fiber.App, *mocks.MockDatabase, *mocks.MockTronService) {
	mockDB := mocks.NewMockDatabase()
	mockTron := mocks.NewMockTronService()
	logger := zap.NewNop()

	userService := &services.UserService{
		Logger: logger,
	}

	authService := &services.AuthService{
		Logger: logger,
	}

	transactionService := &services.TransactionService{
		Logger: logger,
	}

	handler := http.NewHandlerForTesting(
		userService,
		authService,
		transactionService,
		&services.TronService{},
		nil, // queueManager
		http.NewZapLoggerAdapter(logger),
		http.NewRateLimiter(logger),
	)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	return handler, app, mockDB, mockTron
}
