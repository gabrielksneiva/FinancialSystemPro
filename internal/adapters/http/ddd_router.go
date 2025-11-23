package http

import (
	_ "financial-system-pro/docs" // Swagger docs
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/shared/breaker"
	"time"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// DDDRouter gerencia as rotas usando DDD services
type DDDRouter struct {
	userService        *userSvc.UserService
	transactionService *txnSvc.TransactionService
	logger             *zap.Logger
	breakerManager     *breaker.BreakerManager
}

// NewDDDRouter cria uma nova instância do DDDRouter
func NewDDDRouter(
	userService *userSvc.UserService,
	transactionService *txnSvc.TransactionService,
	logger *zap.Logger,
	breakerManager *breaker.BreakerManager,
) *DDDRouter {
	return &DDDRouter{
		userService:        userService,
		transactionService: transactionService,
		logger:             logger,
		breakerManager:     breakerManager,
	}
}

// RegisterDDDRoutes registra todas as rotas DDD na aplicação
func (r *DDDRouter) RegisterDDDRoutes(app *fiber.App) {
	// Health check endpoints (sem autenticação)
	app.Get("/health", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"status": "ok", "timestamp": time.Now().Unix()})
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"ready": true})
	})
	app.Get("/alive", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"alive": true})
	})

	// Prometheus metrics endpoint
	SetupMetricsEndpoint(app)

	// Circuit breaker endpoints
	circuitBreakerHandler := NewCircuitBreakerHandler(r.breakerManager)
	app.Get("/api/circuit-breakers", circuitBreakerHandler.GetCircuitBreakerStatus)
	app.Get("/api/circuit-breakers/health", circuitBreakerHandler.GetCircuitBreakerHealth)

	// Swagger docs
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// DDD Routes
	r.registerUserRoutes(app)
	r.registerTransactionRoutes(app)
}

// registerUserRoutes registra rotas do User Context
func (r *DDDRouter) registerUserRoutes(app *fiber.App) {
	handler := NewDDDUserHandler(r.userService, r.logger)

	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)
}

// registerTransactionRoutes registra rotas do Transaction Context
func (r *DDDRouter) registerTransactionRoutes(app *fiber.App) {
	handler := NewDDDTransactionHandler(
		r.transactionService,
		r.userService,
		r.logger,
		r.breakerManager,
	)

	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/deposit", handler.Deposit)
	protectedPaths.Post("/withdraw", handler.Withdraw)
	protectedPaths.Post("/transfer", handler.Transfer)
	protectedPaths.Get("/balance", handler.Balance)
	protectedPaths.Get("/wallet", handler.GetUserWallet)
}
