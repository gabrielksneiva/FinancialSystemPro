package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	domainErrors "financial-system-pro/internal/domain/errors"
	repo "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== TESTES PARA CREATE USER =====

func TestCreateUser_SuccessWithMock(t *testing.T) {
	handler, userMock, _, _, _ := NewMockedHandler()

	userMock.CreateNewUserFunc = func(req *dto.UserRequest) *domainErrors.AppError {
		assert.Equal(t, "john@example.com", req.Email)
		assert.Equal(t, "password123", req.Password)
		return nil
	}

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := map[string]string{
		"email":    "john@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	handler, userMock, _, _, _ := NewMockedHandler()

	userMock.CreateNewUserFunc = func(req *dto.UserRequest) *domainErrors.AppError {
		return domainErrors.NewValidationError("email", "Email already registered")
	}

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := map[string]string{
		"email":    "existing@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ===== TESTES PARA LOGIN =====

func TestLogin_SuccessWithMock(t *testing.T) {
	handler, _, authMock, _, _ := NewMockedHandler()

	authMock.LoginFunc = func(loginData *dto.LoginRequest) (string, *domainErrors.AppError) {
		assert.Equal(t, "user@example.com", loginData.Email)
		assert.Equal(t, "correct-password", loginData.Password)
		return "jwt-token-12345", nil
	}

	app := fiber.New()
	app.Post("/login", handler.Login)

	body := map[string]string{
		"email":    "user@example.com",
		"password": "correct-password",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	handler, _, authMock, _, _ := NewMockedHandler()

	authMock.LoginFunc = func(loginData *dto.LoginRequest) (string, *domainErrors.AppError) {
		return "", domainErrors.NewValidationError("password", "Invalid password")
	}

	app := fiber.New()
	app.Post("/login", handler.Login)

	body := map[string]string{
		"email":    "user@example.com",
		"password": "wrong-password",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ===== TESTES PARA BALANCE =====

func TestBalance_SuccessWithMock(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.GetBalanceFunc = func(c *fiber.Ctx, userID string) (decimal.Decimal, error) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", userID)
		return decimal.NewFromInt(5000), nil
	}

	app := fiber.New()
	// Middleware para injetar user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "123e4567-e89b-12d3-a456-426614174000")
		return c.Next()
	})
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestBalance_ServiceError(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.GetBalanceFunc = func(c *fiber.Ctx, userID string) (decimal.Decimal, error) {
		return decimal.Zero, errors.New("database connection failed")
	}

	app := fiber.New()
	// Middleware para injetar user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "123e4567-e89b-12d3-a456-426614174000")
		return c.Next()
	})
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// ===== TESTES PARA GET USER WALLET =====

func TestGetUserWallet_SuccessWithMock(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	userID := uuid.New()
	txMock.GetWalletInfoFunc = func(uid uuid.UUID) (*repo.WalletInfo, error) {
		assert.Equal(t, userID, uid)
		return &repo.WalletInfo{
			UserID:      uid,
			TronAddress: "TWalletAddress123456789012345678",
		}, nil
	}

	app := fiber.New()
	// Middleware para injetar user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.String())
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetUserWallet_NotFound(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.GetWalletInfoFunc = func(uid uuid.UUID) (*repo.WalletInfo, error) {
		return nil, errors.New("wallet not found")
	}

	app := fiber.New()
	// Middleware para injetar user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
} // ===== TESTES PARA DEPOSIT =====

func TestDeposit_SuccessWithMock(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.DepositFunc = func(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		assert.True(t, amount.Equal(decimal.NewFromFloat(100.50)))
		return &services.ServiceResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "Deposit queued"},
		}, nil
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/deposit", handler.Deposit)

	body := map[string]interface{}{
		"amount": "100.50",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ===== TESTES PARA WITHDRAW =====

func TestWithdraw_SuccessWithMock(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawFunc = func(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		assert.True(t, amount.Equal(decimal.NewFromInt(50)))
		return &services.ServiceResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "Withdrawal processed"},
		}, nil
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount": "50.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ===== TESTES PARA TRANSFER =====

func TestTransfer_SuccessWithMock(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.TransferFunc = func(c *fiber.Ctx, amount decimal.Decimal, to string, callbackURL string) (*services.ServiceResponse, error) {
		assert.True(t, amount.Equal(decimal.NewFromInt(200)))
		assert.Equal(t, "recipient@example.com", to)
		return &services.ServiceResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "Transfer successful"},
		}, nil
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/transfer", handler.Transfer)

	body := map[string]interface{}{
		"amount": "200.00",
		"to":     "recipient@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ===== TESTES PARA TRON OPERATIONS =====

func TestGetTronBalance_SuccessWithMock(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.GetBalanceFunc = func(address string) (int64, error) {
		assert.Equal(t, "TAddress123456789012345678901234", address)
		return 5000000, nil
	}

	app := fiber.New()
	app.Get("/tron/balance", handler.GetTronBalance)

	req := httptest.NewRequest("GET", "/tron/balance?address=TAddress123456789012345678901234", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTronBalance_InvalidAddress(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return false
	}

	app := fiber.New()
	app.Get("/tron/balance", handler.GetTronBalance)

	req := httptest.NewRequest("GET", "/tron/balance?address=InvalidAddress", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreateTronWallet_SuccessWithMock(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.CreateWalletFunc = func() (*entities.TronWallet, error) {
		return &entities.TronWallet{
			Address:    "TNewWallet123456789012345678901",
			PrivateKey: "private-key-xyz",
			PublicKey:  "public-key-xyz",
		}, nil
	}

	app := fiber.New()
	app.Post("/tron/wallet", handler.CreateTronWallet)

	req := httptest.NewRequest("POST", "/tron/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCheckTronNetwork_Connected(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.IsTestnetConnectedFunc = func() bool {
		return true
	}
	tronMock.GetNetworkInfoFunc = func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"network": "Testnet",
			"block":   12345,
		}, nil
	}

	app := fiber.New()
	app.Get("/tron/network", handler.CheckTronNetwork)

	req := httptest.NewRequest("GET", "/tron/network", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCheckTronNetwork_Disconnected(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.IsTestnetConnectedFunc = func() bool {
		return false
	}

	app := fiber.New()
	app.Get("/tron/network", handler.CheckTronNetwork)

	req := httptest.NewRequest("GET", "/tron/network", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

func TestSendTronTransaction_SuccessWithMock(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return len(address) > 0 && address[0] == 'T'
	}
	tronMock.SendTransactionFunc = func(from, to string, amount int64, privateKey string) (string, error) {
		assert.Equal(t, "TFrom123456789012345678901234", from)
		assert.Equal(t, "TTo123456789012345678901234", to)
		assert.Equal(t, int64(1000000), amount)
		return "tx-hash-success", nil
	}

	app := fiber.New()
	app.Post("/tron/send", handler.SendTronTransaction)

	body := map[string]interface{}{
		"from_address": "TFrom123456789012345678901234",
		"to_address":   "TTo123456789012345678901234",
		"amount":       1000000,
		"private_key":  "mock-private-key",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/send", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)
}

func TestEstimateTronGas_SuccessWithMock(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return len(address) > 0 && address[0] == 'T'
	}
	tronMock.EstimateGasForTransactionFunc = func(from, to string, amount int64) (int64, error) {
		return 25000, nil
	}

	app := fiber.New()
	app.Post("/tron/estimate-gas", handler.EstimateTronGas)

	body := map[string]interface{}{
		"from_address": "TFrom123456789012345678901234",
		"to_address":   "TTo123456789012345678901234",
		"amount":       1000000,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/estimate-gas", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
