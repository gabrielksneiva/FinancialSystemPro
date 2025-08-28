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

	app.Post("/api/users", handler.CreateUser)
	app.Post("/api/login", handler.Login)

	protectedPaths := app.Group("/api", VerifyJWTMiddleware())
	protectedPaths.Post("/deposit", handler.Deposit)

	protectedPaths.Get("/balance", handler.Balance)

}
