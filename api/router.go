package api

import (
	_ "financial-system-pro/docs"
	"financial-system-pro/services"
	"financial-system-pro/workers"
	"time"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func router(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, trasactionService *services.NewTransactionService, tronService *services.TronService, logger *zap.Logger, qm *workers.QueueManager) {
	rateLimiter := NewRateLimiter(logger)

	handler := &NewHandler{
		userService:        userService,
		authService:        authService,
		transactionService: trasactionService,
		tronService:        tronService,
		queueManager:       qm,
		logger:             logger,
		rateLimiter:        rateLimiter,
	}

	// Health check endpoints (sem autenticação, sem dependência de DB)
	app.Get("/health", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"status": "ok", "timestamp": time.Now().Unix()})
	})
	app.Get("/ready", handler.ReadinessProbe)
	app.Get("/alive", handler.LivenessProbe)
	app.Get("/health/full", handler.HealthCheckFull)
	app.Get("/metrics", GetMetrics)

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// Test endpoint for queue
	app.Post("/api/queue/test-deposit", handler.TestQueueDeposit)

	usersRoutes(app, handler)
	transactionsRoutes(app, handler)
	tronRoutes(app, handler)
}

func tronRoutes(app *fiber.App, handler *NewHandler) {
	protectedPaths := app.Group("/api/tron", VerifyJWTMiddleware())
	protectedPaths.Get("/balance", handler.GetTronBalance)
	protectedPaths.Post("/send", handler.SendTronTransaction)
	protectedPaths.Get("/tx-status", handler.GetTronTransactionStatus)
	protectedPaths.Post("/wallet", handler.CreateTronWallet)
	protectedPaths.Get("/network", handler.CheckTronNetwork)
	protectedPaths.Post("/estimate-energy", handler.EstimateTronGas)

	// Novos endpoints RPC
	protectedPaths.Get("/rpc-status", handler.GetRPCStatus)
	protectedPaths.Get("/rpc-methods", handler.GetAvailableMethods)
	protectedPaths.Post("/rpc-call", handler.CallRPCMethod)
}

func usersRoutes(app *fiber.App, handler *NewHandler) {
	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)
}

func transactionsRoutes(app *fiber.App, handler *NewHandler) {
	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/deposit", handler.rateLimiter.Middleware("deposit"), handler.Deposit)
	protectedPaths.Post("/withdraw", handler.rateLimiter.Middleware("withdraw"), handler.Withdraw)
	protectedPaths.Post("/transfer", handler.rateLimiter.Middleware("transfer"), handler.Transfer)
	protectedPaths.Get("/balance", handler.Balance)
}
