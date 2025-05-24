package api

import (
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

type NewHandler struct {
	serviceX *services.ServiceX
}

func RotaTeste(ctx *fiber.Ctx) error {
	return ctx.SendString("Hello, World!")
}

func (h *NewHandler) RotaTeste2(ctx *fiber.Ctx) error {

	return ctx.SendString("Teste")
}
