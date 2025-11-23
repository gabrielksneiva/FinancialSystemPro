package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_GetMetricsBasicShape(t *testing.T) {
	// Reset global metrics counters for test isolation
	GlobalMetrics.DepositCount = 0
	GlobalMetrics.WithdrawCount = 0
	GlobalMetrics.TransferCount = 0
	GlobalMetrics.FailureCount = 0
	GlobalMetrics.TotalRequestTime = 0
	GlobalMetrics.RequestCount = 0
	GlobalMetrics.LastUpdated = time.Now()

	// Simular algumas operações
	RecordDeposit()
	RecordWithdraw()
	RecordTransfer()
	RecordFailure()
	RecordRequestTime(10 * time.Millisecond)

	app := fiber.New()
	app.Get("/metrics-json", GetMetrics)
	req := httptest.NewRequest("GET", "/metrics-json", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&payload)
	assert.NoError(t, err)

	// Verificações principais de estrutura
	tx, ok := payload["transactions"].(map[string]interface{})
	assert.True(t, ok)
	assert.EqualValues(t, 1, tx["deposits"])
	assert.EqualValues(t, 1, tx["withdraws"])
	assert.EqualValues(t, 1, tx["transfers"])
	assert.EqualValues(t, 1, tx["failures"])

	api, ok := payload["api"].(map[string]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, api["total_requests"].(float64), 3.0) // deposit+withdraw+transfer counted

	system, ok := payload["system"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, system["uptime_seconds"])
	assert.NotNil(t, system["goroutines"])
}
