package http

import (
	"context"
	"financial-system-pro/internal/application/dto"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HealthCheck verifica saúde de todos os components
// @Summary      Health check completo
// @Description  Verifica status de DB, Redis e Tron RPC
// @Tags         System
// @Produce      json
// @Success      200  {object}  dto.HealthStatus
// @Success      503  {object}  dto.HealthStatus
// @Router       /health/full [get]
func (h *NewHandler) HealthCheckFull(ctx *fiber.Ctx) error {
	status := dto.HealthStatus{
		Status:   "healthy",
		Uptime:   int64(time.Since(startTime).Seconds()),
		Services: make(map[string]string),
	}

	// Check Database
	dbStatus := "up"
	if h.userService == nil || h.transactionService == nil {
		dbStatus = "down"
		status.Status = "degraded"
	} else {
		// Tentar um query leve no banco
		db := h.userService.GetDatabase()
		if db != nil {
			_, err := db.FindUserByField("id", "00000000-0000-0000-0000-000000000000")
			if err != nil && err.Error() != "record not found" {
				dbStatus = "down"
				status.Status = "degraded"
			}
		}
	}
	status.Services["database"] = dbStatus

	// Check Redis / Queue
	redisStatus := "up"
	if h.queueManager == nil {
		redisStatus = "disabled"
	} else if !h.queueManager.IsConnected() {
		redisStatus = "down"
		status.Status = "degraded"
	}
	status.Services["redis_queue"] = redisStatus

	// Check Tron RPC
	tronStatus := "up"
	if h.tronService == nil {
		tronStatus = "disabled"
	} else {
		// Tentar um call leve na Tron RPC
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.tronService.HealthCheck(ctx); err != nil {
			tronStatus = "down"
			h.logger.Warn("tron rpc health check failed", zap.Error(err))
			// Não degrada status geral pois Tron é opcional
		}
	}
	status.Services["tron_rpc"] = tronStatus

	statusCode := fiber.StatusOK
	if status.Status == "degraded" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return ctx.Status(statusCode).JSON(status)
}

// ReadinessProbe verifica se app está pronto para receber traffic
// @Summary      Readiness probe para Kubernetes/Railway
// @Description  Retorna 200 se app está pronto, 503 se não
// @Tags         System
// @Produce      json
// @Success      200  {object}  dto.ProbeResponse
// @Failure      503  {object}  dto.ProbeResponse
// @Router       /ready [get]
func (h *NewHandler) ReadinessProbe(ctx *fiber.Ctx) error {
	if h.userService == nil || h.transactionService == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(dto.ProbeResponse{
			Ready:  false,
			Reason: "database not ready",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.ProbeResponse{
		Ready: true,
	})
}

// LivenessProbe verifica se app está vivo
// @Summary      Liveness probe para Kubernetes/Railway
// @Description  Retorna 200 se app está vivo (simples)
// @Tags         System
// @Produce      json
// @Success      200  {object}  dto.ProbeResponse
// @Router       /alive [get]
func (h *NewHandler) LivenessProbe(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(dto.ProbeResponse{
		Alive: true,
	})
}
