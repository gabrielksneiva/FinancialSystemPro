package http

import (
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/infrastructure/config/container"
	workers "financial-system-pro/internal/infrastructure/queue"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RegisterRoutes é a função exportada que registra todas as rotas da aplicação
func RegisterRoutes(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, transactionService *services.NewTransactionService, tronService *services.TronService, logger *zap.Logger, qm *workers.QueueManager) {
	router(app, userService, authService, transactionService, tronService, logger, qm)
}

// init registra a função de rotas no container durante a inicialização do package
func init() {
	container.SetRegisterRoutes(RegisterRoutes)
}
