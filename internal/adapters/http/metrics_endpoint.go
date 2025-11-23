package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupMetricsEndpoint configura o endpoint /metrics para Prometheus
func SetupMetricsEndpoint(app *fiber.App) {
	// Endpoint para métricas Prometheus
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
