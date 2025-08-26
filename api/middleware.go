package api

import (
	"financial-system-pro/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func VerifyJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		token, _ = strings.CutPrefix(token, "Bearer ")

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
		}

		decodedToken, err := utils.DecodeJWTToken(token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error while decode token"})
		}

		claims, ok := decodedToken.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		c.Locals("ID", claims["ID"])

		return c.Next()
	}
}
