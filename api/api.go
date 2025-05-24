package api

import (
	"financial-system-pro/repositories"
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

func Start() {
	new_service := &services.ServiceX{Teste: "Injeção de dependências for the win"}

	repositories.ConnectDatabase("host=localhost user=admin password=g123 dbname=database port=5432 sslmode=disable")
	
	app := fiber.New()

	router(app, new_service)

	app.Listen(":3000")
}
