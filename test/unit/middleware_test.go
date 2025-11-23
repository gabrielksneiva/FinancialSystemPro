package http_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"financial-system-pro/internal/shared/utils"
)

// MiddlewareTestSuite tests middleware functionality
type MiddlewareTestSuite struct {
	suite.Suite
	app    *fiber.App
	logger *zap.Logger
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	os.Setenv("SECRET_KEY", "test-secret-key-for-middleware-testing")
	os.Setenv("EXPIRATION_TIME", "3600")
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.logger = zap.NewNop()
	suite.app = fiber.New()
}

// Test JWT Middleware - Valid Token
func (suite *MiddlewareTestSuite) TestJWTMiddleware_ValidToken() {
	// Create a valid token
	claims := jwt.MapClaims{
		"ID": "123e4567-e89b-12d3-a456-426614174000",
	}
	token, err := utils.CreateJWTToken(claims)
	require.NoError(suite.T(), err)

	suite.app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}

		tokenString := authHeader[7:] // Remove "Bearer "
		parsedToken, err := utils.DecodeJWTToken(tokenString)
		if err != nil || !parsedToken.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		claims := parsedToken.Claims.(jwt.MapClaims)
		c.Locals("user_id", claims["ID"])
		return c.Next()
	})

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{"user_id": userID})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := suite.app.Test(req)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp.StatusCode)
}

// Test JWT Middleware - Missing Token
func (suite *MiddlewareTestSuite) TestJWTMiddleware_MissingToken() {
	suite.app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}
		return c.Next()
	})

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)

	resp, err := suite.app.Test(req)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 401, resp.StatusCode)
}

// Test JWT Middleware - Invalid Token
func (suite *MiddlewareTestSuite) TestJWTMiddleware_InvalidToken() {
	suite.app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}

		tokenString := authHeader[7:]
		_, err := utils.DecodeJWTToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		return c.Next()
	})

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := suite.app.Test(req)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 401, resp.StatusCode)
}

// Test JWT Middleware - Expired Token
func (suite *MiddlewareTestSuite) TestJWTMiddleware_ExpiredToken() {
	// Set very short expiration
	os.Setenv("EXPIRATION_TIME", "-1") // Already expired

	claims := jwt.MapClaims{
		"ID": "123e4567-e89b-12d3-a456-426614174000",
	}
	token, _ := utils.CreateJWTToken(claims)

	// Reset expiration
	os.Setenv("EXPIRATION_TIME", "3600")

	suite.app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenString := authHeader[7:]

		parsedToken, err := utils.DecodeJWTToken(tokenString)
		if err != nil || !parsedToken.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "token expired or invalid"})
		}

		return c.Next()
	})

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := suite.app.Test(req)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 401, resp.StatusCode)
}

// Test JWT Middleware - Malformed Authorization Header
func (suite *MiddlewareTestSuite) TestJWTMiddleware_MalformedHeader() {
	suite.app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{"error": "malformed authorization header"})
		}
		return c.Next()
	})

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "just-a-token"},
		{"Wrong prefix", "Basic token"},
		{"Empty Bearer", "Bearer "},
		{"Only Bearer", "Bearer"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			resp, err := suite.app.Test(req)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), 401, resp.StatusCode)
		})
	}
}

// Test Logging Middleware
func (suite *MiddlewareTestSuite) TestLoggingMiddleware() {
	requestCount := 0

	suite.app.Use(func(c *fiber.Ctx) error {
		requestCount++
		suite.logger.Info("Request received",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)
		return c.Next()
	})

	suite.app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := suite.app.Test(req)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp.StatusCode)
	assert.Equal(suite.T(), 1, requestCount)
}

// Test CORS Middleware
func (suite *MiddlewareTestSuite) TestCORSMiddleware() {
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}
		return c.Next()
	})

	suite.app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	resp, err := suite.app.Test(req)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 204, resp.StatusCode)
	assert.Equal(suite.T(), "*", resp.Header.Get("Access-Control-Allow-Origin"))

	// Test regular request
	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, err := suite.app.Test(req2)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp2.StatusCode)
	assert.Equal(suite.T(), "*", resp2.Header.Get("Access-Control-Allow-Origin"))
}

// Test Request ID Middleware
func (suite *MiddlewareTestSuite) TestRequestIDMiddleware() {
	suite.app.Use(func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = "generated-request-id-123"
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	})

	suite.app.Get("/test", func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id")
		return c.JSON(fiber.Map{"request_id": requestID})
	})

	// Test with provided request ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-456")

	resp, err := suite.app.Test(req)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp.StatusCode)
	assert.Equal(suite.T(), "custom-id-456", resp.Header.Get("X-Request-ID"))

	// Test without request ID
	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, err := suite.app.Test(req2)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp2.StatusCode)
	assert.NotEmpty(suite.T(), resp2.Header.Get("X-Request-ID"))
}

// Test Error Handling Middleware
func (suite *MiddlewareTestSuite) TestErrorHandlingMiddleware() {
	suite.app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			suite.logger.Error("Request error", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{
				"error":   "internal server error",
				"message": err.Error(),
			})
		}
		return nil
	})

	suite.app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(500, "something went wrong")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := suite.app.Test(req)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500, resp.StatusCode)
}

// Test Multiple Middleware Chaining
func (suite *MiddlewareTestSuite) TestMiddlewareChaining() {
	var executionOrder []string

	// Middleware 1
	suite.app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "middleware1")
		return c.Next()
	})

	// Middleware 2
	suite.app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "middleware2")
		return c.Next()
	})

	// Middleware 3
	suite.app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "middleware3")
		return c.Next()
	})

	suite.app.Get("/test", func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "handler")
		return c.JSON(fiber.Map{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := suite.app.Test(req)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, resp.StatusCode)
	assert.Equal(suite.T(), []string{"middleware1", "middleware2", "middleware3", "handler"}, executionOrder)
}

// Run the test suite
func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

// Benchmark middleware performance
func BenchmarkJWTMiddleware(b *testing.B) {
	os.Setenv("SECRET_KEY", "test-key")
	os.Setenv("EXPIRATION_TIME", "3600")

	app := fiber.New()

	claims := jwt.MapClaims{"ID": "test-user-id"}
	token, _ := utils.CreateJWTToken(claims)

	app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString := authHeader[7:]
			utils.DecodeJWTToken(tokenString)
		}
		return c.Next()
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		app.Test(req)
	}
}
