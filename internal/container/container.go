package container

import (
	"context"
	"fmt"
	"os"

	"financial-system-pro/internal/container/logger"
	"financial-system-pro/repositories"
	"financial-system-pro/services"
	"financial-system-pro/workers"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config agrupa todas as configurações
type Config struct {
	DatabaseURL string
	JWTSecret   string
}

// LoadConfig carrega configurações das variáveis de ambiente
func LoadConfig() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}

// ProvideDatabaseConnection cria a conexão com banco de dados
func ProvideDatabaseConnection(cfg Config) (*repositories.NewDatabase, error) {
	if cfg.DatabaseURL == "" {
		fmt.Println("Warning: DATABASE_URL not set, running without database")
		return nil, nil
	}

	db := repositories.ConnectDatabase(cfg.DatabaseURL)
	if db == nil {
		return nil, fmt.Errorf("failed to connect to database")
	}

	return &repositories.NewDatabase{DB: db}, nil
}

// ProvideUserService cria o serviço de usuários
func ProvideUserService(database *repositories.NewDatabase) *services.NewUserService {
	if database == nil {
		return nil
	}
	return &services.NewUserService{Database: database}
}

// ProvideAuthService cria o serviço de autenticação
func ProvideAuthService(database *repositories.NewDatabase) *services.NewAuthService {
	if database == nil {
		return nil
	}
	return &services.NewAuthService{Database: database}
}

// ProvideTransactionWorkerPool cria o pool de workers para transações
func ProvideTransactionWorkerPool(database *repositories.NewDatabase) *workers.TransactionWorkerPool {
	if database == nil {
		return nil
	}
	return workers.NewTransactionWorkerPool(database, 5, 100)
}

// ProvideTransactionService cria o serviço de transações
func ProvideTransactionService(
	database *repositories.NewDatabase,
	pool *workers.TransactionWorkerPool,
) *services.NewTransactionService {
	if database == nil || pool == nil {
		return nil
	}
	return &services.NewTransactionService{DB: database, W: pool}
}

// ProvideTronService cria o serviço de Tron
func ProvideTronService() *services.TronService {
	return services.NewTronService()
}

// ProvideApp cria a aplicação Fiber
func ProvideApp() *fiber.App {
	return fiber.New()
}

// ProvideLogger cria o logger Zap centralizado
func ProvideLogger() (*zap.Logger, error) {
	return logger.ProvideLogger()
}

// StartServer inicia o servidor Fiber
func StartServer(lc fx.Lifecycle, app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, transactionService *services.NewTransactionService, tronService *services.TronService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Starting Fiber server on :3000")

			// Registrar rotas no Fiber
			registerFiberRoutes(app, userService, authService, transactionService, tronService)

			go app.Listen(":3000")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Shutting down Fiber server")
			return app.Shutdown()
		},
	})
}

// registerFiberRoutes registra as rotas (cópia do router original)
func registerFiberRoutes(app *fiber.App, userService *services.NewUserService, authService *services.NewAuthService, transactionService *services.NewTransactionService, tronService *services.TronService) {
	// Health check endpoints (sem autenticação)
	app.Get("/health", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"ready": true})
	})
}

// New cria a aplicação com todas as dependências gerenciadas por fx
func New() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),
		fx.Provide(ProvideLogger),
		fx.Provide(ProvideApp),
		fx.Provide(ProvideDatabaseConnection),
		fx.Provide(ProvideUserService),
		fx.Provide(ProvideAuthService),
		fx.Provide(ProvideTransactionWorkerPool),
		fx.Provide(ProvideTransactionService),
		fx.Provide(ProvideTronService),
		fx.Invoke(StartServer),
	)
}
