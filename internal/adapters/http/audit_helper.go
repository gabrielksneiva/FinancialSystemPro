package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"

	repo "financial-system-pro/internal/infrastructure/database"
)

// AuditLogHelper fornece funções auxiliares para logging de auditoria
type AuditLogHelper struct {
	logger *zap.Logger
}

// NewAuditLogHelper cria uma nova instância de AuditLogHelper
func NewAuditLogHelper(logger *zap.Logger) *AuditLogHelper {
	return &AuditLogHelper{
		logger: logger,
	}
}

// LogLoginAttempt registra uma tentativa de login
func (a *AuditLogHelper) LogLoginAttempt(userID uuid.UUID, email string, success bool, ctx *fiber.Ctx) {
	action := "LOGIN_FAILED"
	if success {
		action = "LOGIN_SUCCESS"
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"email":"%s"}`, email)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}

// LogDeposit registra uma operação de depósito
func (a *AuditLogHelper) LogDeposit(userID uuid.UUID, amount string, success bool, ctx *fiber.Ctx) {
	action := "DEPOSIT_FAILED"
	if success {
		action = "DEPOSIT_SUCCESS"
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"amount":"%s"}`, amount)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}

// LogWithdraw registra uma operação de saque
func (a *AuditLogHelper) LogWithdraw(userID uuid.UUID, amount string, success bool, ctx *fiber.Ctx) {
	action := "WITHDRAW_FAILED"
	if success {
		action = "WITHDRAW_SUCCESS"
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"amount":"%s"}`, amount)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}

// LogTransfer registra uma operação de transferência
func (a *AuditLogHelper) LogTransfer(userID uuid.UUID, receiverID string, amount string, success bool, ctx *fiber.Ctx) {
	action := "TRANSFER_FAILED"
	if success {
		action = "TRANSFER_SUCCESS"
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"receiver_id":"%s","amount":"%s"}`, receiverID, amount)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}

// LogTronTransaction registra uma transação TRON
func (a *AuditLogHelper) LogTronTransaction(userID uuid.UUID, txType, toAddress, amount string, success bool, ctx *fiber.Ctx) {
	action := fmt.Sprintf("TRON_%s_FAILED", txType)
	if success {
		action = fmt.Sprintf("TRON_%s_SUCCESS", txType)
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"to_address":"%s","amount":"%s"}`, toAddress, amount)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}

// LogUserCreation registra a criação de um usuário
func (a *AuditLogHelper) LogUserCreation(userID uuid.UUID, email string, success bool, ctx *fiber.Ctx) {
	action := "USER_CREATION_FAILED"
	if success {
		action = "USER_CREATION_SUCCESS"
	}

	auditLog := &repo.AuditLog{
		UserID:     userID,
		Action:     action,
		OldPayload: datatypes.JSON(`{}`),
		NewPayload: datatypes.JSON(fmt.Sprintf(`{"email":"%s"}`, email)),
		IP:         ctx.IP(),
	}

	err := repo.LogAudit(auditLog)
	if err != nil {
		a.logger.Error("failed to log audit", zap.Error(err), zap.String("action", action))
	}
}
