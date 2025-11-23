package http_test

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/internal/application/dto"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateRequest_ValidJSON testa validação com JSON válido
func TestValidateRequest_ValidJSON(t *testing.T) {
	type TestRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	reqBody := TestRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestValidateRequest_InvalidJSON testa validação com JSON inválido
func TestValidateRequest_InvalidJSON(t *testing.T) {
	type TestRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_MissingRequiredField testa campo obrigatório faltando
func TestValidateRequest_MissingRequiredField(t *testing.T) {
	type TestRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	// Missing password field
	reqBody := map[string]string{
		"email": "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_InvalidEmail testa email inválido
func TestValidateRequest_InvalidEmail(t *testing.T) {
	type TestRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	reqBody := map[string]string{
		"email": "invalid-email",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_MinLength testa validação de tamanho mínimo
func TestValidateRequest_MinLength(t *testing.T) {
	type TestRequest struct {
		Password string `json:"password" validate:"required,min=6"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	// Password too short
	reqBody := map[string]string{
		"password": "123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_MultipleValidationErrors testa múltiplos erros
func TestValidateRequest_MultipleValidationErrors(t *testing.T) {
	type TestRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Age      int    `json:"age" validate:"required,min=18"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	// Multiple invalid fields
	reqBody := map[string]interface{}{
		"email":    "not-email",
		"password": "123",
		"age":      15,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_EmptyBody testa body vazio
func TestValidateRequest_EmptyBody(t *testing.T) {
	type TestRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidateRequest_NoValidationTags testa struct sem tags de validação
func TestValidateRequest_NoValidationTags(t *testing.T) {
	type TestRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	reqBody := TestRequest{
		Name: "Test",
		Age:  25,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestValidateRequest_NumericValidation testa validação numérica
func TestValidateRequest_NumericValidation(t *testing.T) {
	type TestRequest struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
		Count  int     `json:"count" validate:"required,min=1,max=100"`
	}

	tests := []struct {
		name           string
		amount         float64
		count          int
		expectedStatus int
	}{
		{"Valid values", 100.50, 50, fiber.StatusOK},
		{"Zero amount", 0, 50, fiber.StatusBadRequest},
		{"Negative amount", -10.0, 50, fiber.StatusBadRequest},
		{"Count too low", 100.0, 0, fiber.StatusBadRequest},
		{"Count too high", 100.0, 101, fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/test", func(c *fiber.Ctx) error {
				var req TestRequest
				if err := dto.ValidateRequest(c, &req); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(req)
			})

			reqBody := TestRequest{
				Amount: tt.amount,
				Count:  tt.count,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestValidateRequest_CustomValidation testa validação customizada
func TestValidateRequest_CustomValidation(t *testing.T) {
	type TestRequest struct {
		Currency string `json:"currency" validate:"required,oneof=USD EUR BRL"`
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	tests := []struct {
		name           string
		currency       string
		expectedStatus int
	}{
		{"Valid USD", "USD", fiber.StatusOK},
		{"Valid EUR", "EUR", fiber.StatusOK},
		{"Valid BRL", "BRL", fiber.StatusOK},
		{"Invalid currency", "JPY", fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{"currency": tt.currency}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
