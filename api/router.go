package api

import (
	_ "financial-system-pro/docs"
	"financial-system-pro/services"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, trasactionService *services.NewTransactionService) {
	handler := &NewHandler{userService: userService, authService: authService, transactionService: trasactionService}

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", fiber.StatusFound)
	})
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	usersRoutes(app, handler)
	transactionsRoutes(app, handler)
}

func usersRoutes(app *fiber.App, handler *NewHandler) {
	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)
}

func transactionsRoutes(app *fiber.App, handler *NewHandler) {
	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/api/deposits", handler.Deposit)
	protectedPaths.Post("/api/withdraws", handler.Withdraw)
	protectedPaths.Post("/api/transfers", handler.Transfer)
	protectedPaths.Get("/api/balance", handler.Balance)
}
