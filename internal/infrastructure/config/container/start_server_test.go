package container

import (
	"context"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// minimalLifecycle implementa Append para capturar hooks.
type minimalLifecycle struct{ hooks []fx.Hook }

func (m *minimalLifecycle) Append(h fx.Hook) { m.hooks = append(m.hooks, h) }

// TestStartServer_Basic inicia e para hooks sem dependências externas (db/redis nulas).
func TestStartServer_Basic(t *testing.T) {
	lg := zap.NewNop()
	app := fiber.New()
	bus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	ml := &minimalLifecycle{}
	// Chamada: serviços DDD nil forçam ramo legacy fallback
	StartServer(ml, app, lg, bus, nil, nil, nil, nil, nil, nil, nil, br, nil, nil)
	if len(ml.hooks) == 0 {
		t.Fatalf("esperava hooks registrados")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	// executa OnStart / OnStop encadeados
	for _, h := range ml.hooks {
		if h.OnStart != nil {
			_ = h.OnStart(ctx)
		}
		if h.OnStop != nil {
			_ = h.OnStop(ctx)
		}
	}
}
