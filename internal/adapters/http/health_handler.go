package http

import (
	"financial-system-pro/internal/application/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthCheck verifica saúde de todos os components
// @Summary      Health check completo
// @Description  Verifica status de DB, Redis e Tron RPC
// @Tags         System
// @Produce      json
// @Success      200  {object}  dto.HealthStatus
// @Success      503  {object}  dto.HealthStatus
// @Router       /health/full [get]
func (h *Handler) HealthCheckFull(ctx *fiber.Ctx) error {
	status := dto.HealthStatus{
		Status:   "healthy",
		Uptime:   int64(time.Since(startTime).Seconds()),
		Services: make(map[string]string),
	}

	// Check Database
	dbStatus := "up"
	if h.dddUserService == nil || h.dddTransactionService == nil {
		dbStatus = "down"
		status.Status = "degraded"
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
	if h.tronGateway == nil {
		tronStatus = "disabled"
	} else {
		// TronGateway doesn't expose HealthCheck; treat connection boolean as health
		if !h.tronGateway.IsTestnetConnected() {
			tronStatus = "down"
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
func (h *Handler) ReadinessProbe(ctx *fiber.Ctx) error {
	if h.dddUserService == nil || h.dddTransactionService == nil {
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
func (h *Handler) LivenessProbe(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(dto.ProbeResponse{
		Alive: true,
	})
}
