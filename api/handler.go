package api

import (
	"financial-system-pro/domain"
	"financial-system-pro/services"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type NewHandler struct {
	userService        *services.NewUserService
	authService        *services.NewAuthService
	transactionService *services.NewTransactionService
}

// CreateUser godoc
// @Summary      Cria um novo usuário
// @Description  Endpoint para criar usuário
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userRequest  body  domain.UserRequest  true  "Dados do usuário"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users [post]
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

// CreateUser godoc
// @Summary      Autentica usuário
// @Description  Endpoint para autenticar usuário
// @Tags         login
// @Accept       json
// @Produce      json
// @Param        loginRequest  body  domain.LoginRequest  true  "Dados de login"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /login [post]
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

// CreateUser godoc
// @Summary      Deposita valor na conta do usuário
// @Description  Endpoint para depositar valor na conta do usuário
// @Tags         deposit
// @Accept       json
// @Produce      json
// @Param        userRequest  body  domain.DepositRequest  true  "Dados do depósito"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /deposit [post]
func (h *NewHandler) Deposit(ctx *fiber.Ctx) error {
	var depositRequest domain.DepositRequest
	err := isValid(ctx, &depositRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	amount, err := decimal.NewFromString(depositRequest.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	err = h.transactionService.Deposit(ctx, amount)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Deposit succesfully"})
}
