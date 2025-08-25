package api

import (
	"financial-system-pro/domain"
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
)

type NewHandler struct {
	userService *services.NewUserService
	authService *services.NewAuthService
}

func (h *NewHandler) CreateUser(ctx *fiber.Ctx) error {
	var userRequest domain.UserRequest
	err := isValid(ctx, &userRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err = h.userService.CreateNewUser(&userRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created succesfully"})
}

func (h *NewHandler) Login(ctx *fiber.Ctx) error {
	var loginRequest domain.LoginRequest
	err := isValid(ctx, &loginRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenJWT, err := h.authService.Login(&loginRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Login succesfully", "token": tokenJWT})
}

func (h* NewHandler) Deposit(ctx *fiber.Ctx) error{
	var depositRequest domain.DepositRequest
	err := isValid(ctx, &depositRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Deposit succesfully"})
}