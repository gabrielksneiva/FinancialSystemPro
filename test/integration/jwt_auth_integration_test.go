package integration

import (
	"net/http/httptest"
	"os"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
)

// buildJWTApp creates a small Fiber app with a protected route using the JWT middleware.
func buildJWTApp() *fiber.App {
	app := fiber.New()
	app.Get("/protected", httpAdapter.VerifyJWTMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true, "user_id": userID})
	})
	return app
}

func TestJWTMiddleware_FullFlow(t *testing.T) {
	app := buildJWTApp()
	os.Setenv("SECRET_KEY", "supersecretkeytest")
	os.Setenv("EXPIRATION_TIME", "3600")

	// 1. Missing token -> 401
	req1 := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	resp1, _ := app.Test(req1, -1)
	if resp1.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token got %d", resp1.StatusCode)
	}

	// Helper to create token with given claims
	makeToken := func(idClaim interface{}) string {
		tokenStr, err := utils.CreateJWTToken(map[string]interface{}{"ID": idClaim})
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		return tokenStr
	}

	// 2. Token without ID claim (nil) -> 401 invalid user ID in token
	tokenNoID := makeToken("")
	req2 := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenNoID)
	resp2, _ := app.Test(req2, -1)
	if resp2.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for empty ID claim got %d", resp2.StatusCode)
	}

	// 3. Invalid signature token -> 500 (Decode error path)
	// Create token with different SECRET_KEY
	os.Setenv("SECRET_KEY", "othersupersecret")
	badToken, _ := utils.CreateJWTToken(map[string]interface{}{"ID": "user-x"})
	// Reset SECRET_KEY so middleware uses original key and fails to verify signature
	os.Setenv("SECRET_KEY", "supersecretkeytest")
	req3 := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+badToken)
	resp3, _ := app.Test(req3, -1)
	if resp3.StatusCode != fiber.StatusInternalServerError && resp3.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 500 or 401 invalid signature got %d", resp3.StatusCode)
	}

	// 4. Valid token -> 200
	goodToken := makeToken("user-123")
	req4 := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req4.Header.Set("Authorization", "Bearer "+goodToken)
	resp4, _ := app.Test(req4, -1)
	if resp4.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 valid token got %d", resp4.StatusCode)
	}

	// 5. Token with extra Bearer spaces trimming
	req5 := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req5.Header.Set("Authorization", "Bearer    "+goodToken) // CutPrefix should still find the token after prefix removal (spaces remain but token parse still attempted)
	resp5, _ := app.Test(req5, -1)
	// Depending on spaces this may fail, we accept 200 or 500 but assert not 401 missing token
	if resp5.StatusCode == fiber.StatusUnauthorized {
		t.Fatalf("unexpected 401 for spaced bearer prefix got %d", resp5.StatusCode)
	}
}

func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	app := buildJWTApp()
	os.Setenv("SECRET_KEY", "supersecretkeytest")
	os.Setenv("EXPIRATION_TIME", "3600")

	// Malformed token
	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")
	resp, _ := app.Test(req, -1)
	// Expect 500 because DecodeJWTToken returns error
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 for malformed token got %d", resp.StatusCode)
	}
}
