package http

import (
	"context"
	"financial-system-pro/internal/application/dto"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/shared/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// DDDUserHandler gerencia as rotas de usuário do DDD User Context
type DDDUserHandler struct {
	userService *userSvc.UserService
	logger      *zap.Logger
}

// NewDDDUserHandler cria uma nova instância do DDDUserHandler
func NewDDDUserHandler(userService *userSvc.UserService, logger *zap.Logger) *DDDUserHandler {
	return &DDDUserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser cria um novo usuário
func (h *DDDUserHandler) CreateUser(ctx *fiber.Ctx) error {
	var createUserReq dto.UserRequest
	if validErr := dto.ValidateRequest(ctx, &createUserReq); validErr != nil {
		h.logger.Warn("invalid create user request", zap.Error(validErr))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": validErr.Message,
			"code":  validErr.Code,
		})
	}

	// Chamar o serviço DDD UserService
	user, err := h.userService.CreateUser(context.Background(), createUserReq.Email, createUserReq.Password)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err), zap.String("email", createUserReq.Email))

		// Audit: Log failed user creation
		auditHelper := NewAuditLogHelper(h.logger)
		auditHelper.LogUserCreation(uuid.Nil, createUserReq.Email, false, ctx)

		// Usar error_mapper para converter domain error em HTTP response
		statusCode := MapDomainErrorToHTTP(err)
		response := MapDomainErrorToResponse(err)
		return ctx.Status(statusCode).JSON(response)
	}

	// Audit: Log successful user creation
	auditHelper := NewAuditLogHelper(h.logger)
	auditHelper.LogUserCreation(user.ID, createUserReq.Email, true, ctx)

	// Evento PublishUserCreated já foi disparado dentro de CreateUser
	h.logger.Info("user created successfully", zap.String("user_id", user.ID.String()), zap.String("email", string(user.Email)))

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    user.ID,
		"email": string(user.Email),
	})
}

// Login autentica um usuário
func (h *DDDUserHandler) Login(ctx *fiber.Ctx) error {
	var loginReq dto.LoginRequest
	if validErr := dto.ValidateRequest(ctx, &loginReq); validErr != nil {
		h.logger.Warn("invalid login request", zap.Error(validErr))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": validErr.Message,
			"code":  validErr.Code,
		})
	}

	// Chamar o serviço DDD UserService
	user, err := h.userService.Authenticate(context.Background(), loginReq.Email, loginReq.Password)
	if err != nil {
		h.logger.Warn("authentication failed", zap.Error(err), zap.String("email", loginReq.Email))

		// Audit: Log failed login attempt (use empty UUID for failed authentication)
		auditHelper := NewAuditLogHelper(h.logger)
		auditHelper.LogLoginAttempt(uuid.Nil, loginReq.Email, false, ctx)

		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Gerar JWT token
	token, err := utils.CreateJWTToken(jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   string(user.Email),
	})
	if err != nil {
		h.logger.Error("failed to generate JWT", zap.Error(err), zap.String("user_id", user.ID.String()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Audit: Log successful login
	auditHelper := NewAuditLogHelper(h.logger)
	auditHelper.LogLoginAttempt(user.ID, loginReq.Email, true, ctx) // Evento PublishUserAuthenticated já foi disparado dentro de Authenticate
	h.logger.Info("user authenticated successfully", zap.String("user_id", user.ID.String()), zap.String("email", string(user.Email)))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"email": string(user.Email),
		},
	})
}
