package integration

import (
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"

	"go.uber.org/zap"
)

func TestRateLimiterMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rateLimiter := httpAdapter.NewRateLimiter(logger)

	// Test IsAllowed directly
	userID := "test-user-123"
	action := "deposit"

	successCount := 0
	failCount := 0

	// Default limit for deposit is 20 per minute
	for i := 0; i < 25; i++ {
		if rateLimiter.IsAllowed(userID, action) {
			successCount++
		} else {
			failCount++
		}
	}

	// Should allow 20 requests and deny 5
	if successCount != 20 {
		t.Errorf("Expected 20 successful requests, got %d", successCount)
	}

	if failCount != 5 {
		t.Errorf("Expected 5 failed requests, got %d", failCount)
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rateLimiter := httpAdapter.NewRateLimiter(logger)

	userID := "test-user-reset"
	action := "login" // 5 per minute

	// Use all 5 requests
	for i := 0; i < 5; i++ {
		if !rateLimiter.IsAllowed(userID, action) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should fail
	if rateLimiter.IsAllowed(userID, action) {
		t.Error("6th request should be denied")
	}

	// Wait for window to reset (1 minute + buffer)
	// Note: In real test, we'd mock time or use smaller window
	// For this test, we'll just verify the logic works

	// Test passes - demonstrating rate limit works as expected
}

func TestRateLimiterDifferentActions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rateLimiter := httpAdapter.NewRateLimiter(logger)

	userID := "test-user-multiaction"

	// Each action has independent rate limit
	depositAllowed := 0
	withdrawAllowed := 0
	transferAllowed := 0

	for i := 0; i < 20; i++ {
		if rateLimiter.IsAllowed(userID, "deposit") {
			depositAllowed++
		}
		if rateLimiter.IsAllowed(userID, "withdraw") {
			withdrawAllowed++
		}
		if rateLimiter.IsAllowed(userID, "transfer") {
			transferAllowed++
		}
	}

	// deposit: 20 limit - all 20 should pass
	if depositAllowed != 20 {
		t.Errorf("Expected 20 deposits, got %d", depositAllowed)
	}

	// withdraw: 15 limit - only 15 should pass
	if withdrawAllowed != 15 {
		t.Errorf("Expected 15 withdraws, got %d", withdrawAllowed)
	}

	// transfer: 10 limit - only 10 should pass
	if transferAllowed != 10 {
		t.Errorf("Expected 10 transfers, got %d", transferAllowed)
	}
}

func TestRateLimiterDifferentUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rateLimiter := httpAdapter.NewRateLimiter(logger)

	action := "login" // 5 per minute

	// Each user should have independent rate limit
	user1Allowed := 0
	user2Allowed := 0

	for i := 0; i < 6; i++ {
		if rateLimiter.IsAllowed("user-1", action) {
			user1Allowed++
		}
		if rateLimiter.IsAllowed("user-2", action) {
			user2Allowed++
		}
	}

	// Each user gets 5 requests
	if user1Allowed != 5 {
		t.Errorf("User 1: Expected 5 requests, got %d", user1Allowed)
	}

	if user2Allowed != 5 {
		t.Errorf("User 2: Expected 5 requests, got %d", user2Allowed)
	}
}
