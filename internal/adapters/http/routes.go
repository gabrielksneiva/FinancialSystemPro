package http

import (
	"financial-system-pro/internal/application/services"
	txnDDD "financial-system-pro/internal/contexts/transaction/application/service"
	userDDD "financial-system-pro/internal/contexts/user/application/service"
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
	userService *services.UserService,
	authService *services.AuthService,
	transactionService *services.TransactionService,
	tronService *services.TronService,
	multiChainWalletService *services.MultiChainWalletService,
	logger *zap.Logger,
	qm *workers.QueueManager,
	breakerManager *breaker.BreakerManager,
	dddUserService *userDDD.UserService,
	dddTransactionService *txnDDD.TransactionService,
) {
	// Legacy rotas
	router(app, userService, authService, transactionService, tronService, multiChainWalletService, logger, qm, breakerManager)
	// Rotas v2 baseadas nos bounded contexts DDD
	if dddUserService != nil && dddTransactionService != nil {
		registerV2DDDRoutes(app, dddUserService, dddTransactionService, logger, breakerManager)
	}
}

// RegisterDDDRoutes é a função para registrar apenas rotas DDD
// Usada quando temos os DDD services disponíveis
// Mantida para compat legacy (não usada no fluxo v2 novo)
func RegisterDDDRoutes(
	app *fiber.App,
	userService services.UserServiceInterface,
	transactionService services.TransactionServiceInterface,
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
