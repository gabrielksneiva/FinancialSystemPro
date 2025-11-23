package config

import (
	"os"
	"strconv"
	"time"
)

// Config contém todas as configurações da aplicação
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Tron     TronConfig
	App      AppConfig
}

// ServerConfig configuração do servidor HTTP
type ServerConfig struct {
	Host           string
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	Environment    string // development, staging, production
}

// DatabaseConfig configuração do PostgreSQL
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
	LogLevel        string
	SSLMode         string
	AutoMigrate     bool
}

// RedisConfig configuração do Redis
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxRetries   int
}

// JWTConfig configuração de autenticação JWT
type JWTConfig struct {
	Secret                string
	ExpirationTime        time.Duration
	RefreshSecret         string
	RefreshExpirationTime time.Duration
}

// TronConfig configuração de integração TRON
type TronConfig struct {
	MainnetRPC     string
	TestnetRPC     string
	MainnetGRPC    string
	TestnetGRPC    string
	Timeout        time.Duration
	Network        string // mainnet, testnet
	DefaultGasUsed int64
}

// AppConfig configurações gerais da aplicação
type AppConfig struct {
	Name             string
	Version          string
	LogLevel         string // debug, info, warn, error
	EnableCORS       bool
	CORSAllowOrigins string
	RateLimitPerMin  int
	EnableMetrics    bool
}

// Load carrega configurações das variáveis de ambiente
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Port:           getEnvInt("SERVER_PORT", 3000),
			ReadTimeout:    getEnvDuration("SERVER_READ_TIMEOUT", "10s"),
			WriteTimeout:   getEnvDuration("SERVER_WRITE_TIMEOUT", "10s"),
			IdleTimeout:    getEnvDuration("SERVER_IDLE_TIMEOUT", "60s"),
			MaxHeaderBytes: getEnvInt("SERVER_MAX_HEADER_BYTES", 1<<20),
			Environment:    getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/financialsystempro?sslmode=disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", "1h"),
			LogLevel:        getEnv("DB_LOG_LEVEL", "error"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			AutoMigrate:     getEnvBool("DB_AUTO_MIGRATE", false),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", "3s"),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", "3s"),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpirationTime:        getEnvDuration("JWT_EXPIRATION", "24h"),
			RefreshSecret:         getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
			RefreshExpirationTime: getEnvDuration("JWT_REFRESH_EXPIRATION", "168h"),
		},
		Tron: TronConfig{
			MainnetRPC:     getEnv("TRON_MAINNET_RPC", "https://api.tronstack.io"),
			TestnetRPC:     getEnv("TRON_TESTNET_RPC", "https://api.nile.trongrid.io"),
			MainnetGRPC:    getEnv("TRON_MAINNET_GRPC", "grpc.tronstack.io"),
			TestnetGRPC:    getEnv("TRON_TESTNET_GRPC", "grpc.nile.trongrid.io"),
			Timeout:        getEnvDuration("TRON_TIMEOUT", "30s"),
			Network:        getEnv("TRON_NETWORK", "testnet"),
			DefaultGasUsed: getEnvInt64("TRON_DEFAULT_GAS_USED", 21000),
		},
		App: AppConfig{
			Name:             getEnv("APP_NAME", "FinancialSystemPro"),
			Version:          getEnv("APP_VERSION", "1.0.0"),
			LogLevel:         getEnv("LOG_LEVEL", "info"),
			EnableCORS:       getEnvBool("ENABLE_CORS", true),
			CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
			RateLimitPerMin:  getEnvInt("RATE_LIMIT_PER_MIN", 100),
			EnableMetrics:    getEnvBool("ENABLE_METRICS", true),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultStr string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultStr); err == nil {
		return duration
	}
	return 0
}
