package integration

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type queueManagerFake struct{ taskID string }

func (q queueManagerFake) EnqueueDeposit(_ context.Context, userID, amount, cb string) (string, error) {
	return q.taskID, nil
}
func (q queueManagerFake) EnqueueWithdraw(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}
func (q queueManagerFake) EnqueueTransfer(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}
func (q queueManagerFake) IsConnected() bool { return true }

// Reuse testLogger from existing integration file (multichain)

func TestQueueDepositIntegration_Success(t *testing.T) {
	rl := httpAdapter.NewRateLimiter(zap.NewNop())
	h := httpAdapter.NewHandlerForTesting(nil, nil, nil, queueManagerFake{taskID: "TASK_INTEGRATION"}, newTestLogger(), rl)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)

	req := httptest.NewRequest(fiber.MethodPost, "/api/queue/test-deposit", strings.NewReader(`{"user_id":"uid123","amount":"25"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 got %d body=%s", resp.StatusCode, body)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "OK") {
		t.Fatalf("expected OK body got %s", b)
	}
}
