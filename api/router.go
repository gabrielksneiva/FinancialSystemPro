package api

import (
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService) {
	handler := &NewHandler{userService: userService, authService: authService}

	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)

	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/deposit", func(c *fiber.Ctx) error { return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OK"}) })

}
