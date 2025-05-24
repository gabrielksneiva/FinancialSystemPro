package api

import (
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App, serviceX *services.ServiceX) {
	handler := &NewHandler{serviceX: serviceX}

	app.Get("/", handler.RotaTeste2)
}
