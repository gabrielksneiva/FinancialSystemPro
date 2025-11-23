package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"financial-system-pro/internal/application/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== TESTES PARA WITHDRAW =====

func TestWithdraw_Success(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
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

func TestWithdraw_InvalidAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount": "not-a-number",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWithdraw_ZeroAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount": "0.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWithdraw_ServiceError(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		return nil, errors.New("insufficient balance")
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/withdraw", handler.Withdraw)

	body := map[string]interface{}{
		"amount": "100.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWithdraw_TronType(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.WithdrawTronFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		return &services.ServiceResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "TRON withdrawal processed"},
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

// ===== TESTES PARA CallRPCMethod =====

func TestCallRPCMethod_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/rpc", handler.CallRPCMethod)

	req := httptest.NewRequest("POST", "/rpc", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCallRPCMethod_MissingMethod(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/rpc", handler.CallRPCMethod)

	body := map[string]interface{}{
		"params": []interface{}{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/rpc", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

func TestCallRPCMethod_NilRPCClient(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.GetRPCClientFunc = func() *services.RPCClient {
		return nil
	}

	app := fiber.New()
	app.Post("/rpc", handler.CallRPCMethod)

	body := map[string]interface{}{
		"method": "eth_blockNumber",
		"params": []interface{}{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/rpc", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

// ===== TESTES ADICIONAIS PARA checkDatabaseAvailable =====

func TestBalance_MissingUserID(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	// Sem middleware de user_id
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalance_InvalidUserIDType(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 123) // Tipo inválido (int em vez de string)
		return c.Next()
	})
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalance_EmptyUserID(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "") // String vazia
		return c.Next()
	})
	app.Get("/balance", handler.Balance)

	req := httptest.NewRequest("GET", "/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ===== TESTES PARA GetUserWallet edge cases =====

func TestGetUserWallet_MissingUserID(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	// Sem middleware
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetUserWallet_InvalidUserIDType(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 999) // Tipo inválido
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetUserWallet_InvalidUUIDFormat(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "not-a-valid-uuid")
		return c.Next()
	})
	app.Get("/wallet", handler.GetUserWallet)

	req := httptest.NewRequest("GET", "/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
