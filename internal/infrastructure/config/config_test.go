package config

import (
	"os" // for Clearenv only
	"strings"
	"testing"
	"time"
)

func containsString(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

func TestLoadConfigDefaults(t *testing.T) {
	// Clear any existing environment variables
	os.Clearenv()

	cfg := Load()

	// Test Server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected Server.Host to be '0.0.0.0', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Expected Server.Port to be 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Environment != "development" {
		t.Errorf("Expected Server.Environment to be 'development', got '%s'", cfg.Server.Environment)
	}

	// Test Database defaults
	if cfg.Database.MaxOpenConns != 100 {
		t.Errorf("Expected Database.MaxOpenConns to be 100, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("Expected Database.MaxIdleConns to be 10, got %d", cfg.Database.MaxIdleConns)
	}

	// Test Redis defaults
	if cfg.Redis.Host != "localhost" {
		t.Errorf("Expected Redis.Host to be 'localhost', got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Expected Redis.Port to be 6379, got %d", cfg.Redis.Port)
	}

	// Test JWT defaults
	if cfg.JWT.ExpirationTime == 0 {
		t.Error("Expected JWT.ExpirationTime to be set")
	}

	// Test Tron defaults
	if cfg.Tron.Network != "testnet" {
		t.Errorf("Expected Tron.Network to be 'testnet', got '%s'", cfg.Tron.Network)
	}

	// Test App defaults
	if cfg.App.LogLevel != "info" {
		t.Errorf("Expected App.LogLevel to be 'info', got '%s'", cfg.App.LogLevel)
	}
}

func TestLoadConfigFromEnvironment(t *testing.T) {
	// Set environment variables with automatic cleanup
	t.Setenv("SERVER_HOST", "0.0.0.0")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://db.example.com:5433/testdb")
	t.Setenv("REDIS_HOST", "cache.example.com")
	t.Setenv("JWT_SECRET", "test-secret-key-1234567890123456")

	cfg := Load()

	// Test overridden values
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected Server.Host to be '0.0.0.0', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Expected Server.Port to be 9000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Environment != "production" {
		t.Errorf("Expected Server.Environment to be 'production', got '%s'", cfg.Server.Environment)
	}
	if !containsString(cfg.Database.DSN, "db.example.com") {
		t.Errorf("Expected Database.DSN to contain 'db.example.com', got '%s'", cfg.Database.DSN)
	}
	if cfg.Redis.Host != "cache.example.com" {
		t.Errorf("Expected Redis.Host to be 'cache.example.com', got '%s'", cfg.Redis.Host)
	}
}

func TestLoadConfigDurations(t *testing.T) {
	t.Setenv("SERVER_READ_TIMEOUT", "30s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "45s")

	cfg := Load()

	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected Server.ReadTimeout to be 30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 45*time.Second {
		t.Errorf("Expected Server.WriteTimeout to be 45s, got %v", cfg.Server.WriteTimeout)
	}
}

func TestLoadConfigBooleans(t *testing.T) {
	t.Setenv("DB_AUTO_MIGRATE", "false")
	t.Setenv("ENABLE_METRICS", "false")

	cfg := Load()

	if cfg.Database.AutoMigrate != false {
		t.Error("Expected Database.AutoMigrate to be false")
	}
	if cfg.App.EnableMetrics != false {
		t.Error("Expected App.EnableMetrics to be false")
	}
}

func TestLoadConfigDatabaseDSN(t *testing.T) {
	dsnValue := "postgres://myuser:mypass@myhost:5432/mydb?sslmode=disable"
	t.Setenv("DATABASE_URL", dsnValue)

	cfg := Load()

	// DSN should be set from environment variable
	if cfg.Database.DSN != dsnValue {
		t.Errorf("Expected Database.DSN to be '%s', got '%s'", dsnValue, cfg.Database.DSN)
	}
}
