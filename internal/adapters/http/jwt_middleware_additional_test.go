package http

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// setupAppWithJWTMiddleware creates a minimal app to exercise middleware branches.
func setupAppWithJWTMiddleware(t *testing.T) *fiber.App {
	t.Helper()
	t.Setenv("SECRET_KEY", "branch-secret")
	t.Setenv("EXPIRATION_TIME", "3600")
	app := fiber.New()
	grp := app.Group("/p", VerifyJWTMiddleware())
	grp.Get("/protected", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	return app
}

func TestJWTMiddleware_InvalidTokenBranch(t *testing.T) {
	app := setupAppWithJWTMiddleware(t)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("branch-secret"))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	req := httptest.NewRequest("GET", "/p/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized { // falls into !decodedToken.Valid path
		t.Fatalf("expected 401 invalid token got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_InvalidUserID_Missing(t *testing.T) {
	app := setupAppWithJWTMiddleware(t)
	// MapClaims without ID
	claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("branch-secret"))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	req := httptest.NewRequest("GET", "/p/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 missing user ID got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_InvalidUserID_EmptyString(t *testing.T) {
	app := setupAppWithJWTMiddleware(t)
	claims := jwt.MapClaims{"ID": "", "exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("branch-secret"))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	req := httptest.NewRequest("GET", "/p/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 empty user ID got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_InvalidUserID_WrongType(t *testing.T) {
	app := setupAppWithJWTMiddleware(t)
	claims := jwt.MapClaims{"ID": 12345, "exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("branch-secret"))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	req := httptest.NewRequest("GET", "/p/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 wrong type user ID got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	app := setupAppWithJWTMiddleware(t)
	claims := jwt.MapClaims{"ID": "user-1", "exp": time.Now().Add(-time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("branch-secret"))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	req := httptest.NewRequest("GET", "/p/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusInternalServerError { // decode error or expired
		t.Fatalf("expected 500 expired token got %d", resp.StatusCode)
	}
}
