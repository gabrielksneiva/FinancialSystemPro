package http

import (
	"financial-system-pro/internal/shared/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func VerifyJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		token, _ = strings.CutPrefix(token, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
		}

		decodedToken, err := utils.DecodeJWTToken(token)
		if err != nil {
			// Qualquer erro ao decodificar o token deve ser tratado como não autorizado
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		if !decodedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		claims, ok := decodedToken.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		userID, ok := claims["ID"].(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
		}

		c.Locals("user_id", userID)

		return c.Next()
	}
}
