package http_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== TESTES PARA FUNÇÕES COM 0% COBERTURA =====

// Testes para GetRPCStatus
func TestGetRPCStatus_Success(t *testing.T) {
	handler, _, _, _, tronMock := NewMockedHandler()

	tronMock.GetConnectionStatusFunc = func() map[string]interface{} {
		return map[string]interface{}{
			"rpc_connected":  true,
			"grpc_connected": true,
			"network":        "testnet",
		}
	}

	app := fiber.New()
	app.Get("/rpc-status", handler.GetRPCStatus)

	req := httptest.NewRequest("GET", "/rpc-status", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// Testes para GetAvailableMethods
func TestGetAvailableMethods_Success(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Get("/methods", handler.GetAvailableMethods)

	req := httptest.NewRequest("GET", "/methods", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// Testes para CreateUser - cobertura completa
func TestCreateUser_ValidationError(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	// JSON inválido para forçar erro de validação
	req := httptest.NewRequest("POST", "/users", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Testes para Login - cobertura completa
func TestLogin_ValidationError(t *testing.T) {
	handler, _, _, _, _ := NewMockedHandler()

	app := fiber.New()
	app.Post("/login", handler.Login)

	// JSON inválido para forçar erro de validação
	req := httptest.NewRequest("POST", "/login", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
