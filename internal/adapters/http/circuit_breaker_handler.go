package http

import (
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
)

// CircuitBreakerHandler gerencia endpoints de circuit breaker
type CircuitBreakerHandler struct {
	manager *breaker.BreakerManager
}

// NewCircuitBreakerHandler cria um novo handler
func NewCircuitBreakerHandler(manager *breaker.BreakerManager) *CircuitBreakerHandler {
	return &CircuitBreakerHandler{
		manager: manager,
	}
}

// GetCircuitBreakerStatus retorna o status de todos os circuit breakers
// @Summary Get circuit breaker status
// @Description Retorna o estado atual de todos os circuit breakers
// @Tags Observability
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/circuit-breakers [get]
func (h *CircuitBreakerHandler) GetCircuitBreakerStatus(c *fiber.Ctx) error {
	states := h.manager.GetAllStates()

	response := fiber.Map{
		"circuit_breakers": states,
		"legend": fiber.Map{
			"closed":    "Normal operation - requests passing through",
			"open":      "Circuit breaker tripped - requests failing fast",
			"half-open": "Testing if service recovered - limited requests",
		},
	}

	return c.JSON(response)
}

// GetCircuitBreakerHealth retorna health check baseado nos circuit breakers
// @Summary Circuit breaker health check
// @Description Verifica se algum circuit breaker está aberto
// @Tags Observability
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 503 {object} map[string]interface{}
// @Router /api/circuit-breakers/health [get]
func (h *CircuitBreakerHandler) GetCircuitBreakerHealth(c *fiber.Ctx) error {
	states := h.manager.GetAllStates()

	openBreakers := []string{}
	for name, state := range states {
		if state == "open" {
			openBreakers = append(openBreakers, name)
		}
	}

	if len(openBreakers) > 0 {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":        "unhealthy",
			"open_breakers": openBreakers,
			"message":       "Some circuit breakers are open",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "healthy",
		"message": "All circuit breakers operational",
	})
}
