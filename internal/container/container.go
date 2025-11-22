package container

import (
	"context"
	"financial-system-pro/domain"
	"fmt"
	"os"
	"time"

	"financial-system-pro/internal/logger"
	"financial-system-pro/internal/validator"
	"financial-system-pro/repositories"
	"financial-system-pro/services"
	"financial-system-pro/workers"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterRoutesFunc é uma função que registra rotas na aplicação
// Ela será fornecida pelo package api para evitar ciclo de import
type RegisterRoutesFunc func(
	app *fiber.App,
	userService *services.NewUserService,
	authService *services.NewAuthService,
	transactionService *services.NewTransactionService,
	tronService *services.TronService,
	logger *zap.Logger,
	qm *workers.QueueManager,
)

// defaultRegisterRoutes é a função padrão (pode ser sobrescrita)
var defaultRegisterRoutes RegisterRoutesFunc

// SetRegisterRoutes permite que o package api registre a função real
func SetRegisterRoutes(fn RegisterRoutesFunc) {
	defaultRegisterRoutes = fn
}

// Config agrupa todas as configurações
type Config struct {
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
}

// LoadConfig carrega configurações das variáveis de ambiente
func LoadConfig() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		RedisURL:    os.Getenv("REDIS_URL"),
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
func ProvideUserService(database *repositories.NewDatabase, lg *zap.Logger, walletManager domain.WalletManager) *services.NewUserService {
	if database == nil {
		return nil
	}
	return &services.NewUserService{
		Database:      database,
		Logger:        lg,
		WalletManager: walletManager,
	}
}

// ProvideAuthService cria o serviço de autenticação
func ProvideAuthService(database *repositories.NewDatabase, lg *zap.Logger) *services.NewAuthService {
	if database == nil {
		return nil
	}
	return &services.NewAuthService{
		Database: database,
		Logger:   lg,
	}
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
	tronPool *workers.TronWorkerPool,
	tronSvc *services.TronService,
	lg *zap.Logger,
) *services.NewTransactionService {
	if database == nil || pool == nil {
		return nil
	}
	return &services.NewTransactionService{
		DB:             database,
		W:              pool,
		TronWorkerPool: tronPool,
		TronService:    tronSvc,
		Logger:         lg,
	}
}

// ProvideTronService cria o serviço de Tron
func ProvideTronService() *services.TronService {
	return services.NewTronService()
}

// ProvideTronWorkerPool cria o pool de workers para confirmação de transações TRON
func ProvideTronWorkerPool(
	database *repositories.NewDatabase,
	tronSvc *services.TronService,
	lg *zap.Logger,
) *workers.TronWorkerPool {
	pool := workers.NewTronWorkerPool(database, tronSvc, 5, lg)
	pool.Start() // Inicia automaticamente
	return pool
}

// ProvideApp cria a aplicação Fiber
func ProvideApp() *fiber.App {
	return fiber.New()
}

// ProvideLogger cria o logger Zap centralizado
func ProvideLogger() (*zap.Logger, error) {
	return logger.ProvideLogger()
}

// ProvideValidator cria o serviço de validação
func ProvideValidator() *validator.ValidatorService {
	return validator.New()
}

// ProvideWalletManager cria o gerenciador de carteiras TRON
func ProvideWalletManager() domain.WalletManager {
	return services.NewTronWalletManager()
}

// StartServer inicia o servidor Fiber e workers
func StartServer(lc fx.Lifecycle, app *fiber.App, lg *zap.Logger, userService *services.NewUserService, authService *services.NewAuthService, transactionService *services.NewTransactionService, tronService *services.TronService, registerRoutes RegisterRoutesFunc, qm *workers.QueueManager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lg.Info("Starting Fiber server on port 3000")

			// Registrar rotas via callback function
			if registerRoutes != nil {
				registerRoutes(app, userService, authService, transactionService, tronService, lg, qm)
			} else {
				// Fallback: só health checks
				registerFiberHealthChecks(app)
			}

			// Iniciar workers de fila se disponível
			// TODO: Iniciar workers quando QueueManager conectar com sucesso
			if qm != nil {
				// Aguardar um pouco para QueueManager conectar
				time.Sleep(5 * time.Second)
				if err := qm.StartWorkers(ctx); err != nil {
					lg.Warn("failed to start queue workers", zap.Error(err))
					// Não quebra a app
				}
			}

			go app.Listen(":3000")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			lg.Info("Shutting down Fiber server and workers")
			if qm != nil {
				qm.Close()
			}
			return app.Shutdown()
		},
	})
}

// registerFiberHealthChecks registra endpoints de health check
func registerFiberHealthChecks(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(200).JSON(fiber.Map{"ready": true})
	})
}

// ProvideRegisterRoutes fornece a função de registro de rotas
// Usa a versão padrão ou a registrada pelo package api
func ProvideRegisterRoutes() RegisterRoutesFunc {
	return defaultRegisterRoutes
}

// ProvideQueueManager fornece o gerenciador de fila Redis
// Redis é OPCIONAL - se falhar, app continua sem async queue
func ProvideQueueManager(cfg Config, lg *zap.Logger, database *repositories.NewDatabase) *workers.QueueManager {
	lg.Info("[REDIS DEBUG] ProvideQueueManager called",
		zap.String("redis_url_length", fmt.Sprintf("%d chars", len(cfg.RedisURL))),
		zap.Bool("redis_url_empty", cfg.RedisURL == ""))

	if cfg.RedisURL == "" {
		lg.Warn("[REDIS DEBUG] REDIS_URL environment variable not set, running without async queue")
		return nil
	}

	lg.Info("[REDIS DEBUG] attempting to initialize queue manager with redis url")
	qm := workers.NewQueueManager(cfg.RedisURL, lg, database)
	if qm == nil {
		lg.Warn("[REDIS DEBUG] queue manager is nil, running without async queue")
		return nil
	}

	lg.Info("[REDIS DEBUG] queue manager initialized successfully")
	return qm
}

// New cria a aplicação com todas as dependências gerenciadas por fx
func New() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),
		fx.Provide(ProvideLogger),
		fx.Provide(ProvideValidator),
		fx.Provide(ProvideQueueManager),
		fx.Provide(ProvideRegisterRoutes),
		fx.Provide(ProvideApp),
		fx.Provide(ProvideDatabaseConnection),
		fx.Provide(ProvideWalletManager),
		fx.Provide(ProvideUserService),
		fx.Provide(ProvideAuthService),
		fx.Provide(ProvideTransactionWorkerPool),
		fx.Provide(ProvideTronService),
		fx.Provide(ProvideTronWorkerPool),
		fx.Provide(ProvideTransactionService),
		fx.Invoke(StartServer),
	)
}
