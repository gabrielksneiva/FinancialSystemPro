package tracing

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// FiberTracingMiddleware cria middleware de tracing para Fiber
func FiberTracingMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		// Extrair contexto de propagação dos headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Context(),
			propagation.HeaderCarrier(c.GetReqHeaders()),
		)

		// Criar span
		spanName := c.Method() + " " + c.Route().Path
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Adicionar atributos
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.url", c.OriginalURL()),
			attribute.String("http.route", c.Route().Path),
			attribute.String("http.user_agent", c.Get("User-Agent")),
			attribute.String("http.client_ip", c.IP()),
		)

		// Salvar span no contexto do Fiber
		c.Locals("otel-span", span)
		c.SetUserContext(ctx)

		// Processar request
		err := c.Next()

		// Registrar status code
		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.Bool("error", true))
		}

		return err
	}
}

// GetSpanFromFiber recupera o span do contexto do Fiber
func GetSpanFromFiber(c *fiber.Ctx) trace.Span {
	if span, ok := c.Locals("otel-span").(trace.Span); ok {
		return span
	}
	return nil
}
