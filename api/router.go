package api

import (
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, trasactionService *services.NewTransactionService) {
	handler := &NewHandler{userService: userService, authService: authService, transactionService: trasactionService}

	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)

	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/deposit", handler.Deposit)

}
