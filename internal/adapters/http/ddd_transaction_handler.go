package http

import (
	"context"
	"financial-system-pro/internal/application/dto"
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// DDDTransactionHandler gerencia as rotas de transações do DDD Transaction Context
type DDDTransactionHandler struct {
	transactionService *txnSvc.TransactionService
	userService        *userSvc.UserService
	logger             *zap.Logger
	breakerManager     *breaker.BreakerManager
}

// NewDDDTransactionHandler cria uma nova instância do DDDTransactionHandler
func NewDDDTransactionHandler(
	transactionService *txnSvc.TransactionService,
	userService *userSvc.UserService,
	logger *zap.Logger,
	breakerManager *breaker.BreakerManager,
) *DDDTransactionHandler {
	return &DDDTransactionHandler{
		transactionService: transactionService,
		userService:        userService,
		logger:             logger,
		breakerManager:     breakerManager,
	}
}

// Deposit processa um depósito
func (h *DDDTransactionHandler) Deposit(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userIDStr, ok := userIDLocal.(string)
	if !ok || userIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	var depositReq dto.DepositRequest
	if validErr := dto.ValidateRequest(ctx, &depositReq); validErr != nil {
		h.logger.Warn("invalid deposit request", zap.Error(validErr))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": validErr.Message,
			"code":  validErr.Code,
		})
	}

	amount, err := decimal.NewFromString(depositReq.Amount)
	if err != nil {
		h.logger.Warn("invalid amount format", zap.String("amount", depositReq.Amount))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	// Usar circuit breaker para chamar o serviço DDD
	breaker := h.breakerManager.GetBreaker("BreakerTransactionToUser")
	if breaker == nil {
		h.logger.Error("circuit breaker not available")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Service unavailable"})
	}

	// Processar depósito com circuit breaker
	_, err = breaker.Execute(func() (interface{}, error) {
		return nil, h.transactionService.ProcessDeposit(context.Background(), userID, amount, depositReq.CallbackURL)
	})

	// Initialize audit helper
	auditHelper := NewAuditLogHelper(h.logger)

	if err != nil {
		h.logger.Error("failed to process deposit", zap.Error(err), zap.String("user_id", userID.String()))

		// Audit: Log failed deposit
		auditHelper.LogDeposit(userID, amount.String(), false, ctx)

		if err.Error() == "circuit breaker is open" {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service temporarily unavailable (circuit breaker open)",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process deposit"})
	}

	// Audit: Log successful deposit
	auditHelper.LogDeposit(userID, amount.String(), true, ctx)

	h.logger.Info("deposit processed successfully", zap.String("user_id", userID.String()), zap.String("amount", amount.String()))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "deposit_initiated",
		"amount":  amount.String(),
		"user_id": userID.String(),
	})
}

// Withdraw processa um saque
func (h *DDDTransactionHandler) Withdraw(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userIDStr, ok := userIDLocal.(string)
	if !ok || userIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	var withdrawReq dto.WithdrawRequest
	if validErr := dto.ValidateRequest(ctx, &withdrawReq); validErr != nil {
		h.logger.Warn("invalid withdraw request", zap.Error(validErr))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": validErr.Message,
			"code":  validErr.Code,
		})
	}

	amount, err := decimal.NewFromString(withdrawReq.Amount)
	if err != nil {
		h.logger.Warn("invalid amount format", zap.String("amount", withdrawReq.Amount))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	// Usar circuit breaker para chamar o serviço DDD
	breaker := h.breakerManager.GetBreaker("BreakerTransactionToBlockchain")
	if breaker == nil {
		h.logger.Error("circuit breaker not available")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Service unavailable"})
	}

	// Processar saque com circuit breaker
	_, err = breaker.Execute(func() (interface{}, error) {
		return nil, h.transactionService.ProcessWithdraw(context.Background(), userID, amount)
	})

	// Initialize audit helper
	auditHelper := NewAuditLogHelper(h.logger)

	if err != nil {
		h.logger.Error("failed to process withdraw", zap.Error(err), zap.String("user_id", userID.String()))

		// Audit: Log failed withdraw
		auditHelper.LogWithdraw(userID, amount.String(), false, ctx)

		if err.Error() == "circuit breaker is open" {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service temporarily unavailable (circuit breaker open)",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process withdraw"})
	}

	// Audit: Log successful withdraw
	auditHelper.LogWithdraw(userID, amount.String(), true, ctx)

	h.logger.Info("withdraw processed successfully", zap.String("user_id", userID.String()), zap.String("amount", amount.String()))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "withdraw_initiated",
		"amount":  amount.String(),
		"user_id": userID.String(),
	})
}

// Transfer faz transferência entre usuários (não implementado no TransactionService por enquanto)
func (h *DDDTransactionHandler) Transfer(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userIDStr, ok := userIDLocal.(string)
	if !ok || userIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	fromUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	var transferReq dto.TransferRequest
	if validErr := dto.ValidateRequest(ctx, &transferReq); validErr != nil {
		h.logger.Warn("invalid transfer request", zap.Error(validErr))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": validErr.Message,
			"code":  validErr.Code,
		})
	}

	amount, err := decimal.NewFromString(transferReq.Amount)
	if err != nil {
		h.logger.Warn("invalid amount format", zap.String("amount", transferReq.Amount))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	h.logger.Info("transfer feature not yet implemented for DDD", zap.String("from_user_id", fromUserID.String()))

	return ctx.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Transfer feature not yet implemented",
	})
}

// Balance retorna o saldo do usuário
func (h *DDDTransactionHandler) Balance(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userIDStr, ok := userIDLocal.(string)
	if !ok || userIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	// Chamar o serviço DDD TransactionService
	history, err := h.transactionService.GetTransactionHistory(context.Background(), userID)
	if err != nil {
		h.logger.Error("failed to get transaction history", zap.Error(err), zap.String("user_id", userID.String()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get balance"})
	}

	// Calculator saldo a partir do histórico
	var balance decimal.Decimal
	for _, tx := range history {
		if tx.Status == "completed" {
			balance = balance.Add(tx.Amount)
		}
	}

	h.logger.Info("balance queried successfully", zap.String("user_id", userID.String()), zap.String("balance", balance.String()))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"balance": balance.String(),
		"user_id": userID.String(),
	})
}

// GetUserWallet retorna a carteira do usuário
func (h *DDDTransactionHandler) GetUserWallet(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userIDStr, ok := userIDLocal.(string)
	if !ok || userIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	// Chamar o serviço DDD UserService para obter a carteira
	wallet, err := h.userService.GetUserWallet(context.Background(), userID)
	if err != nil {
		h.logger.Error("failed to get user wallet", zap.Error(err), zap.String("user_id", userID.String()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get wallet"})
	}

	if wallet == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
	}

	h.logger.Info("wallet queried successfully", zap.String("user_id", userID.String()), zap.String("wallet_address", wallet.Address))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":      wallet.ID,
		"user_id": wallet.UserID,
		"address": wallet.Address,
		"balance": wallet.Balance,
	})
}
