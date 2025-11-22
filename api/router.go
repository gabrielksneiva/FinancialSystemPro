package api

import (
	_ "financial-system-pro/docs"
	"financial-system-pro/services"
	"time"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, trasactionService *services.NewTransactionService, tronService *services.TronService) {
	handler := &NewHandler{userService: userService, authService: authService, transactionService: trasactionService, tronService: tronService}

	// Health check endpoints (sem autenticação, sem dependência de DB)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ready": true, "timestamp": time.Now()})
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/*", fiberSwagger.WrapHandler)

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
	protectedPaths.Post("/deposit", handler.Deposit)
	protectedPaths.Post("/withdraw", handler.Withdraw)
	protectedPaths.Post("/transfer", handler.Transfer)
	protectedPaths.Get("/balance", handler.Balance)
}
