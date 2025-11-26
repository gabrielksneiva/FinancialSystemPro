package http

import (
	_ "financial-system-pro/docs" // Swagger docs
	txnDDD "financial-system-pro/internal/contexts/transaction/application/service"
	userDDD "financial-system-pro/internal/contexts/user/application/service"
	workers "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/breaker"
	"time"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// router (legacy signature removido). Agora utiliza serviços DDD diretamente.
func router(app *fiber.App, userService *userDDD.UserService, transactionService *txnDDD.TransactionService, logger *zap.Logger, qm *workers.QueueManager, breakerManager *breaker.BreakerManager) {
	rateLimiter := NewRateLimiter(logger)

	// Criar adapters para as interfaces
	loggerAdapter := NewZapLoggerAdapter(logger)
	queueManagerAdapter := NewQueueManagerAdapter(qm)
	rateLimiterAdapter := NewRateLimiterAdapter(rateLimiter)

	handler := &Handler{
		dddUserService:        userService,
		dddTransactionService: transactionService,
		queueManager:          queueManagerAdapter,
		logger:                loggerAdapter,
		rateLimiter:           rateLimiterAdapter,
	}

	// Health check endpoints (sem autenticação, sem dependência de DB)
	app.Get("/health", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"status": "ok", "timestamp": time.Now().Unix()})
	})
	app.Get("/ready", handler.ReadinessProbe)
	app.Get("/alive", handler.LivenessProbe)
	app.Get("/health/full", handler.HealthCheckFull)
	app.Get("/metrics-old", GetMetrics) // Métricas antigas customizadas

	// Prometheus metrics endpoint
	SetupMetricsEndpoint(app)

	// Circuit breaker endpoints
	circuitBreakerHandler := NewCircuitBreakerHandler(breakerManager)
	app.Get("/api/circuit-breakers", circuitBreakerHandler.GetCircuitBreakerStatus)
	app.Get("/api/circuit-breakers/health", circuitBreakerHandler.GetCircuitBreakerHealth)

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// Test endpoint for queue
	app.Post("/api/queue/test-deposit", handler.TestQueueDeposit)

	// Setup versioned API routes
	setupV1Routes(app, handler)
	// Future: setupV2Routes(app, handler)
}

// setupV1Routes configura as rotas da API v1
func setupV1Routes(app *fiber.App, handler *Handler) {
	v1 := app.Group("/api/v1")

	// Users routes
	v1.Post("/users", handler.CreateUser)
	v1.Post("/login", handler.Login)

	// Protected routes
	protected := v1.Group("", VerifyJWTMiddleware())

	// Audit routes (admin/monitoring)
	protected.Get("/audit", handler.GetAuditLogs)
	protected.Get("/audit/stats", handler.GetAuditStats)

	// Transactions routes
	protected.Post("/deposit", handler.rateLimiter.Middleware("deposit"), handler.Deposit)
	protected.Post("/withdraw", handler.rateLimiter.Middleware("withdraw"), handler.Withdraw)
	protected.Post("/transfer", handler.rateLimiter.Middleware("transfer"), handler.Transfer)
	protected.Get("/balance", handler.Balance)
	protected.Get("/wallet", handler.GetUserWallet)
	protected.Post("/wallets/generate", handler.GenerateWallet)

	// Tron routes
	protected.Get("/tron/balance", handler.GetTronBalance)
	protected.Post("/tron/send", handler.SendTronTransaction)
	protected.Get("/tron/tx-status", handler.GetTronTransactionStatus)
	protected.Post("/tron/wallet", handler.CreateTronWallet)
	protected.Get("/tron/network", handler.CheckTronNetwork)
	protected.Post("/tron/estimate-energy", handler.EstimateTronGas)
	protected.Get("/tron/rpc-status", handler.GetRPCStatus)
	protected.Get("/tron/rpc-methods", handler.GetAvailableMethods)
	protected.Post("/tron/rpc-call", handler.CallRPCMethod)
}

// setupV2Routes configura as rotas da API v2 (para futuro com breaking changes)
// func setupV2Routes(app *fiber.App, handler *Handler) {
// 	v2 := app.Group("/api/v2")

// 	// Adicionar rotas v2 com alterações quando necessário
// 	// Por exemplo: v2.Post("/users", handler.CreateUserV2)
// 	_ = v2 // placeholder para evitar erro de "imported but not used"
// }
