package http

import (
	"financial-system-pro/internal/application/services"
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/infrastructure/config/container"
	workers "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RegisterRoutes é a função exportada que registra todas as rotas da aplicação
// Mantém suporte ao legacy services
func RegisterRoutes(
	app *fiber.App,
	userService *services.NewUserService,
	authService *services.NewAuthService,
	transactionService *services.NewTransactionService,
	tronService *services.TronService,
	logger *zap.Logger,
	qm *workers.QueueManager,
	breakerManager *breaker.BreakerManager,
) {
	router(app, userService, authService, transactionService, tronService, logger, qm, breakerManager)
}

// RegisterDDDRoutes é a função para registrar apenas rotas DDD
// Usada quando temos os DDD services disponíveis
func RegisterDDDRoutes(
	app *fiber.App,
	userService *userSvc.UserService,
	transactionService *txnSvc.TransactionService,
	logger *zap.Logger,
	breakerManager *breaker.BreakerManager,
) {
	if userService != nil && transactionService != nil {
		dddRouter := NewDDDRouter(userService, transactionService, logger, breakerManager)
		dddRouter.RegisterDDDRoutes(app)
	}
}

// init registra a função de rotas no container durante a inicialização do package
func init() {
	container.SetRegisterRoutes(RegisterRoutes)
}
