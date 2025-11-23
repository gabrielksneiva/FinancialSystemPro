package http

import (
	"context"
	"financial-system-pro/internal/contexts/transaction/application/service"
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// registerV2DDDRoutes registra rotas v2 usando serviços DDD diretamente.
func registerV2DDDRoutes(app *fiber.App, userService *userSvc.UserService, txnService *txnSvc.TransactionService, logger *zap.Logger, breakerManager *breaker.BreakerManager) {
	api := app.Group("/v2")

	// Users
	api.Post("/users", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil || body.Email == "" || body.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		user, err := userService.CreateUser(context.Background(), body.Email, body.Password)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": user.ID, "email": user.Email})
	})

	api.Post("/auth/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		user, err := userService.Authenticate(context.Background(), body.Email, body.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		token, tErr := utils.CreateJWTToken(map[string]interface{}{"ID": user.ID})
		if tErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token generation failed"})
		}
		return c.JSON(fiber.Map{"token": token})
	})

	// Transactions
	txGroup := api.Group("/transactions", VerifyJWTMiddleware())

	txGroup.Post("/deposit", func(c *fiber.Ctx) error {
		var body struct {
			Amount string `json:"amount"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		amt, err := decimal.NewFromString(body.Amount)
		if err != nil || amt.LessThanOrEqual(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid amount"})
		}
		userID, err := extractUserIDFromJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		if err := txnService.ProcessDeposit(context.Background(), userID, amt, ""); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "deposit_queued"})
	})

	txGroup.Post("/withdraw", func(c *fiber.Ctx) error {
		var body struct {
			Amount string `json:"amount"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		amt, err := decimal.NewFromString(body.Amount)
		if err != nil || amt.LessThanOrEqual(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid amount"})
		}
		userID, err := extractUserIDFromJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		if err := txnService.ProcessWithdraw(context.Background(), userID, amt); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "withdraw_processed"})
	})

	txGroup.Get("/history", func(c *fiber.Ctx) error {
		userID, err := extractUserIDFromJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		list, err := txnService.GetTransactionHistory(context.Background(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"transactions": list})
	})

	api.Get("/users/:id/wallet", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		wallet, err := userService.GetUserWallet(context.Background(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wallet not found"})
		}
		return c.JSON(wallet)
	})
}

// extractUserIDFromJWT retorna userID do token presente em contexto
func extractUserIDFromJWT(c *fiber.Ctx) (uuid.UUID, error) {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	uid, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.UUID{}, err
	}
	return uid, nil
}

// Garantir uso do pacote service (import alias txnSvc já usado) para evitar import pruning
var _ = service.NewTransactionService
