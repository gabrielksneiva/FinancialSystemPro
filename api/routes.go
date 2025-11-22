package api

import (
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registra todas as rotas da aplicação
// Esta função é chamada pelo container fx
func RegisterRoutes(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, transactionService *services.NewTransactionService, tronService *services.TronService) {
	router(app, userService, authService, transactionService, tronService)
}
