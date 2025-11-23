package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracerProvider é o provider global de tracing
var TracerProvider *sdktrace.TracerProvider

// InitTracer inicializa o OpenTelemetry tracer com Jaeger
func InitTracer(serviceName string, logger *zap.Logger) (func(context.Context) error, error) {
	// Endpoint do Jaeger (pode ser configurado via env var)
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:4318/v1/traces" // Jaeger OTLP HTTP endpoint
	}

	// Criar exporter OTLP HTTP
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(), // Para desenvolvimento
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Criar resource com informações do serviço
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", getEnv()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Criar TracerProvider
	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample todas em dev
	)

	// Registrar como global
	otel.SetTracerProvider(TracerProvider)

	logger.Info("distributed tracing initialized",
		zap.String("service", serviceName),
		zap.String("endpoint", jaegerEndpoint),
	)

	// Retornar função de cleanup
	return TracerProvider.Shutdown, nil
}

// GetTracer retorna um tracer para o pacote especificado
func GetTracer(instrumentationName string) trace.Tracer {
	return otel.Tracer(instrumentationName)
}

// StartSpan inicia um novo span
func StartSpan(ctx context.Context, tracerName, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := GetTracer(tracerName)
	ctx, span := tracer.Start(ctx, spanName)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}

// getEnv retorna o ambiente atual
func getEnv() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}
