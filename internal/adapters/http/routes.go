package http

import (
	txnDDD "financial-system-pro/internal/contexts/transaction/application/service"
	userDDD "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/infrastructure/config/container"
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RegisterRoutes é a função exportada que registra todas as rotas da aplicação
// Mantém suporte ao legacy services
func RegisterRoutes(
	app *fiber.App,
	dddUserService *userDDD.UserService,
	dddTransactionService *txnDDD.TransactionService,
	logger *zap.Logger,
	breakerManager *breaker.BreakerManager,
) {
	// Apenas rotas DDD v2
	registerV2DDDRoutes(app, dddUserService, dddTransactionService, logger, breakerManager)
}

// RegisterDDDRoutes é a função para registrar apenas rotas DDD
// Usada quando temos os DDD services disponíveis
// Mantida para compat legacy (não usada no fluxo v2 novo)
// RegisterDDDRoutes (legacy compat) removido: v2 usa RegisterRoutes diretamente.

// init registra a função de rotas no container durante a inicialização do package
func init() {
	container.SetRegisterRoutes(RegisterRoutes)
}
