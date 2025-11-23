package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== TESTES ADICIONAIS PARA AUMENTAR COBERTURA =====

// Testes para GetTronTransactionStatus
func TestGetTronTransactionStatus_Success(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.GetTransactionStatusFunc = func(txHash string) (string, error) {
		assert.Equal(t, "valid-tx-hash-123", txHash)
		return "confirmed", nil
	}

	app := fiber.New()
	app.Get("/tron/tx-status", handler.GetTronTransactionStatus)

	req := httptest.NewRequest("GET", "/tron/tx-status?tx_hash=valid-tx-hash-123", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTronTransactionStatus_MissingHash(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Get("/tron/tx-status", handler.GetTronTransactionStatus)

	req := httptest.NewRequest("GET", "/tron/tx-status", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTronTransactionStatus_ServiceError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.GetTransactionStatusFunc = func(txHash string) (string, error) {
		return "", errors.New("transaction not found")
	}

	app := fiber.New()
	app.Get("/tron/tx-status", handler.GetTronTransactionStatus)

	req := httptest.NewRequest("GET", "/tron/tx-status?tx_hash=invalid-hash", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para CreateTronWallet
func TestCreateTronWallet_ServiceError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.CreateWalletFunc = func() (*entities.TronWallet, error) {
		return nil, errors.New("failed to create wallet")
	}

	app := fiber.New()
	app.Post("/tron/wallet", handler.CreateTronWallet)

	req := httptest.NewRequest("POST", "/tron/wallet", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para CheckTronNetwork
func TestCheckTronNetwork_GetNetworkInfoError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.IsTestnetConnectedFunc = func() bool {
		return true
	}
	tronMock.GetNetworkInfoFunc = func() (map[string]interface{}, error) {
		return nil, errors.New("network info unavailable")
	}

	app := fiber.New()
	app.Get("/tron/network", handler.CheckTronNetwork)

	req := httptest.NewRequest("GET", "/tron/network", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para EstimateTronGas - casos de erro
func TestEstimateTronGas_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/tron/estimate-gas", handler.EstimateTronGas)

	req := httptest.NewRequest("POST", "/tron/estimate-gas", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestEstimateTronGas_InvalidFromAddress(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return address != "InvalidAddress"
	}

	app := fiber.New()
	app.Post("/tron/estimate-gas", handler.EstimateTronGas)

	body := map[string]interface{}{
		"from_address": "InvalidAddress",
		"to_address":   "TTo123456789012345678901234",
		"amount":       1000000,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/estimate-gas", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestEstimateTronGas_ServiceError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return len(address) > 0 && address[0] == 'T'
	}
	tronMock.EstimateGasForTransactionFunc = func(from, to string, amount int64) (int64, error) {
		return 0, errors.New("estimation failed")
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
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para SendTronTransaction - casos de erro
func TestSendTronTransaction_InvalidFromAddress(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return address != "InvalidFrom"
	}

	app := fiber.New()
	app.Post("/tron/send", handler.SendTronTransaction)

	body := map[string]interface{}{
		"from_address": "InvalidFrom",
		"to_address":   "TTo123456789012345678901234",
		"amount":       1000000,
		"private_key":  "mock-key",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/send", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendTronTransaction_InvalidToAddress(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return address != "InvalidTo"
	}

	app := fiber.New()
	app.Post("/tron/send", handler.SendTronTransaction)

	body := map[string]interface{}{
		"from_address": "TFrom123456789012345678901234",
		"to_address":   "InvalidTo",
		"amount":       1000000,
		"private_key":  "mock-key",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/send", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendTronTransaction_ServiceError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return len(address) > 0 && address[0] == 'T'
	}
	tronMock.SendTransactionFunc = func(from, to string, amount int64, privateKey string) (string, error) {
		return "", errors.New("insufficient balance")
	}

	app := fiber.New()
	app.Post("/tron/send", handler.SendTronTransaction)

	body := map[string]interface{}{
		"from_address": "TFrom123456789012345678901234",
		"to_address":   "TTo123456789012345678901234",
		"amount":       1000000,
		"private_key":  "mock-key",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tron/send", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para Deposit - casos de erro
func TestDeposit_InvalidAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/deposit", handler.Deposit)

	body := map[string]interface{}{
		"amount": "invalid-number",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeposit_NegativeAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/deposit", handler.Deposit)

	body := map[string]interface{}{
		"amount": "-100.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeposit_ServiceError(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.DepositFunc = func(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
		return nil, errors.New("queue is full")
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/deposit", handler.Deposit)

	body := map[string]interface{}{
		"amount": "100.00",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/deposit", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para Transfer - casos de erro
func TestTransfer_InvalidAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/transfer", handler.Transfer)

	body := map[string]interface{}{
		"amount": "not-a-number",
		"to":     "recipient@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTransfer_ZeroAmount(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/transfer", handler.Transfer)

	body := map[string]interface{}{
		"amount": "0.00",
		"to":     "recipient@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTransfer_ServiceError(t *testing.T) {
	handler, _, _, txMock, _ := NewMockedHandler()

	txMock.TransferFunc = func(userID string, amount decimal.Decimal, to string, callbackURL string) (*services.ServiceResponse, error) {
		return nil, errors.New("insufficient funds")
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/transfer", handler.Transfer)

	body := map[string]interface{}{
		"amount": "100.00",
		"to":     "recipient@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Testes para GetTronBalance - casos de erro
func TestGetTronBalance_MissingAddress(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Get("/tron/balance", handler.GetTronBalance)

	req := httptest.NewRequest("GET", "/tron/balance", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTronBalance_ServiceError(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.ValidateAddressFunc = func(address string) bool {
		return len(address) > 0 && address[0] == 'T'
	}
	tronMock.GetBalanceFunc = func(address string) (int64, error) {
		return 0, errors.New("connection timeout")
	}

	app := fiber.New()
	app.Get("/tron/balance", handler.GetTronBalance)

	req := httptest.NewRequest("GET", "/tron/balance?address=TAddress123456789012345678901234", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
