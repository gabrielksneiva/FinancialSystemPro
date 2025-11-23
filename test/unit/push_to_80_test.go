package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== TESTES ADICIONAIS PARA ALCANÇAR 80% =====

// Testes para Withdraw - cobertura completa
func TestWithdraw_WithdrawTronSuccess(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawTronFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		return &services.ServiceResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "TRON withdrawal queued"},
		}, nil
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount":        "100.00",
		"withdraw_type": "tron",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestWithdraw_WithdrawTronError(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawTronFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		return nil, errors.New("no TRON wallet found")
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount":        "100.00",
		"withdraw_type": "tron",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWithdraw_NegativeAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount": "-50.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes adicionais para CreateUser
func TestCreateUser_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes adicionais para Login
func TestLogin_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/login", handler.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("not valid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes adicionais para Deposit
func TestDeposit_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/deposit", handler.Deposit)

	req := httptest.NewRequest("POST", "/deposit", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes adicionais para Transfer
func TestTransfer_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/transfer", handler.Transfer)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader([]byte("[invalid]")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes adicionais para EstimateTronGas
func TestEstimateTronGas_InvalidToAddress(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return address != "InvalidTo"
	}

	app := fiber.New()
	app.Post("/tron/estimate-gas", handler.EstimateTronGas)

	body := map[string]interface{}{
		"from_address": "TFrom123456789012345678901234",
		"to_address":   "InvalidTo",
		"amount":       1000000,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/estimate-gas", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes para SendTronTransaction - casos adicionais
func TestSendTronTransaction_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/tron/send", handler.SendTronTransaction)

	req := httptest.NewRequest("POST", "/tron/send", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes para CreateTronWallet - edge cases
func TestCreateTronWallet_AlreadyExists(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.CreateWalletFunc = func() (*entities.TronWallet, error) {
		return &entities.TronWallet{
			Address:    "TNewAddress123456789012345678",
			PrivateKey: "priv-key-123",
			PublicKey:  "pub-key-123",
		}, nil
	}

	app := fiber.New()
	app.Post("/tron/wallet", handler.CreateTronWallet)

	req := httptest.NewRequest("POST", "/tron/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// Testes adicionais para GetTronBalance
func TestGetTronBalance_EmptyAddress(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Get("/tron/balance", handler.GetTronBalance)

	req := httptest.NewRequest("GET", "/tron/balance?address=", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes para GetUserWallet - mais edge cases
func TestGetUserWallet_EmptyUserID(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "")
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetUserWallet_NonStringUserID(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 12345) // Não é string
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes para Balance - mais edge cases
func TestBalance_NonStringUserIDInLocals(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 999) // int em vez de string
		return c.Next()
	})
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
