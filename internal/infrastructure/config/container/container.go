package container

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	"fmt"
	"os"
	"time"

	"financial-system-pro/internal/application/services"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/infrastructure/logger"
	workers "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/database"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/tracing"
	"financial-system-pro/internal/shared/validator"

	// DDD Bounded Contexts - User
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	userPers "financial-system-pro/internal/contexts/user/infrastructure/persistence"

	// DDD Bounded Contexts - Transaction
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	txnPers "financial-system-pro/internal/contexts/transaction/infrastructure/persistence"

	// DDD Bounded Contexts - Blockchain
	bcPers "financial-system-pro/internal/contexts/blockchain/infrastructure/persistence"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterRoutesFunc é uma função que registra rotas na aplicação
// Ela será fornecida pelo package api para evitar ciclo de import
type RegisterRoutesFunc func(
	app *fiber.App,
	userService *services.UserService,
	authService *services.AuthService,
	transactionService *services.TransactionService,
	tronService *services.TronService,
	logger *zap.Logger,
	qm *workers.QueueManager,
	breakerManager *breaker.BreakerManager,
	dddUserService *userSvc.UserService,
	dddTransactionService *txnSvc.TransactionService,
)

// Tipos para DDD Repositories e Services (evita conflitos no fx)
type (
	DDDUserRepositoryType                  struct{}
	DDDWalletRepositoryType                struct{}
	DDDUserServiceType                     struct{}
	DDDTransactionRepositoryType           struct{}
	DDDTransactionServiceType              struct{}
	DDDBlockchainTransactionRepositoryType struct{}
)

// defaultRegisterRoutes é a função padrão (pode ser sobrescrita)
var defaultRegisterRoutes RegisterRoutesFunc

// SetRegisterRoutes permite que o package api registre a função real
func SetRegisterRoutes(fn RegisterRoutesFunc) {
	defaultRegisterRoutes = fn
}

// Config agrupa todas as configurações
type Config struct {
	DatabaseURL         string
	JWTSecret           string
	RedisURL            string
	EncryptionKey       string // Chave para criptografar private keys
	TronVaultAddress    string // Endereço da carteira do cofre (origem dos withdraws)
	TronVaultPrivateKey string // Private key do cofre (para assinar transações)
}

// LoadConfig carrega configurações das variáveis de ambiente
func LoadConfig() Config {
	return Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		RedisURL:            os.Getenv("REDIS_URL"),
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
		TronVaultAddress:    os.Getenv("TRON_VAULT_ADDRESS"),
		TronVaultPrivateKey: os.Getenv("TRON_VAULT_PRIVATE_KEY"),
	}
}

// ProvideSharedDatabaseConnection cria a conexão com banco usando interface compartilhada
func ProvideSharedDatabaseConnection(cfg Config) (database.Connection, error) {
	if cfg.DatabaseURL == "" {
		fmt.Println("Warning: DATABASE_URL not set")
		return nil, nil
	}

	return database.NewPostgresConnection(cfg.DatabaseURL)
}

// ProvideDatabaseConnection cria a conexão com banco de dados (legacy)
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
func ProvideUserService(database *repositories.NewDatabase, lg *zap.Logger, walletManager entities.WalletManager) *services.UserService {
	if database == nil {
		return nil
	}
	return services.NewUserService(database, lg, walletManager)
}

// ProvideAuthService cria o serviço de autenticação
func ProvideAuthService(database *repositories.NewDatabase, lg *zap.Logger) *services.AuthService {
	if database == nil {
		return nil
	}
	return services.NewAuthService(database, lg)
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
	eventBus events.Bus,
	lg *zap.Logger,
) *services.TransactionService {
	if database == nil {
		return nil
	}
	return services.NewTransactionService(database, nil, tronSvc, tronPool, eventBus, lg)
}

// ProvideTronService cria o serviço de Tron
func ProvideTronService(cfg Config) *services.TronService {
	return services.NewTronService(cfg.TronVaultAddress, cfg.TronVaultPrivateKey)
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

// ProvideEventBus cria o event bus in-memory
func ProvideEventBus(lg *zap.Logger) events.Bus {
	return events.NewInMemoryBus(lg)
}

// ProvideBreakerManager cria o gerenciador de circuit breakers
func ProvideBreakerManager(lg *zap.Logger) *breaker.BreakerManager {
	return breaker.NewBreakerManager(lg)
}

// ProvideValidator cria o serviço de validação
func ProvideValidator() *validator.ValidatorService {
	return validator.New()
}

// ProvideWalletManager cria o gerenciador de carteiras TRON
func ProvideWalletManager() entities.WalletManager {
	return services.NewTronWalletManager()
}

// StartServer inicia o servidor Fiber e workers
func StartServer(
	lc fx.Lifecycle,
	app *fiber.App,
	lg *zap.Logger,
	eventBus events.Bus,
	userService *services.UserService,
	authService *services.AuthService,
	transactionService *services.TransactionService,
	tronService *services.TronService,
	registerRoutes RegisterRoutesFunc,
	qm *workers.QueueManager,
	breakerManager *breaker.BreakerManager,
	// DDD Services (podem ser nil se não estiverem disponíveis)
	dddUserService *userSvc.UserService,
	dddTransactionService *txnSvc.TransactionService,
) {
	// Inicializar distributed tracing
	shutdownTracer, err := tracing.InitTracer("financial-system-pro", lg)
	if err != nil {
		lg.Warn("failed to initialize tracer, continuing without tracing", zap.Error(err))
	} else {
		lg.Info("distributed tracing initialized successfully")
	}

	// Configurar event subscribers
	services.SetupEventSubscribers(eventBus, lg)
	lg.Info("event subscribers configured")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lg.Info("Starting Fiber server on port 3000")

			// Adicionar middleware de tracing
			app.Use(tracing.FiberTracingMiddleware("financial-system-pro"))

			// Se temos DDD services, usar as novas rotas DDD
			if dddUserService != nil && dddTransactionService != nil {
				lg.Info("registering DDD routes")
				// Importar a função do package http
				// Isso será feito dinamicamente já que estamos no package container
				// e precisamos evitar ciclo de import

				// Para isso, vamos usar a função RegisterDDDRoutes se disponível
				// Por enquanto, registrar apenas legacy routes
				if registerRoutes != nil {
					registerRoutes(app, userService, authService, transactionService, tronService, lg, qm, breakerManager, dddUserService, dddTransactionService)
				} else {
					registerFiberHealthChecks(app)
				}
			} else {
				// Registrar legacy rotas
				if registerRoutes != nil {
					registerRoutes(app, userService, authService, transactionService, tronService, lg, qm, breakerManager, dddUserService, dddTransactionService)
				} else {
					// Fallback: só health checks
					registerFiberHealthChecks(app)
				}
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

			go func() { _ = app.Listen(":3000") }()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			lg.Info("Shutting down Fiber server and workers")

			// Shutdown tracer
			if shutdownTracer != nil {
				if err := shutdownTracer(ctx); err != nil {
					lg.Warn("failed to shutdown tracer", zap.Error(err))
				}
			}

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

// ProvideUserRepository cria o repositório de usuários para o DDD User Context
func ProvideUserRepository(conn database.Connection) userRepo.UserRepository {
	if conn == nil {
		return nil
	}
	return userPers.NewPostgresUserRepository(conn)
}

// ProvideWalletRepository cria o repositório de wallets para o DDD User Context
func ProvideWalletRepository(conn database.Connection) userRepo.WalletRepository {
	if conn == nil {
		return nil
	}
	return userPers.NewPostgresWalletRepository(conn)
}

// ProvideDDDUserService cria o UserService do DDD User Context
func ProvideDDDUserService(
	userRepoImpl userRepo.UserRepository,
	walletRepoImpl userRepo.WalletRepository,
	eventBus events.Bus,
	lg *zap.Logger,
) *userSvc.UserService {
	if userRepoImpl == nil || walletRepoImpl == nil {
		return nil
	}
	return userSvc.NewUserService(userRepoImpl, walletRepoImpl, eventBus, lg)
}

// ProvideTransactionRepository cria o repositório de transações para o DDD Transaction Context
func ProvideTransactionRepository(conn database.Connection) txnRepo.TransactionRepository {
	if conn == nil {
		return nil
	}
	return txnPers.NewPostgresTransactionRepository(conn)
}

// ProvideDDDTransactionService cria o TransactionService do DDD Transaction Context
func ProvideDDDTransactionService(
	txnRepoImpl txnRepo.TransactionRepository,
	userRepoImpl userRepo.UserRepository,
	walletRepoImpl userRepo.WalletRepository,
	eventBus events.Bus,
	breakerManager *breaker.BreakerManager,
	lg *zap.Logger,
) *txnSvc.TransactionService {
	if txnRepoImpl == nil || userRepoImpl == nil || walletRepoImpl == nil {
		return nil
	}
	return txnSvc.NewTransactionService(
		txnRepoImpl,
		userRepoImpl,
		walletRepoImpl,
		eventBus,
		breakerManager,
		lg,
	)
}

// ProvideBlockchainTransactionRepository cria o repositório de transações blockchain
func ProvideBlockchainTransactionRepository(conn database.Connection) *bcPers.PostgresBlockchainTransactionRepository {
	if conn == nil {
		return nil
	}
	return bcPers.NewPostgresBlockchainTransactionRepository(conn)
}

// New cria a aplicação com todas as dependências gerenciadas por fx
func New() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),
		fx.Provide(ProvideLogger),
		fx.Provide(ProvideEventBus),
		fx.Provide(ProvideBreakerManager),
		fx.Provide(ProvideValidator),
		fx.Provide(ProvideQueueManager),
		fx.Provide(ProvideRegisterRoutes),
		fx.Provide(ProvideApp),
		fx.Provide(ProvideDatabaseConnection),
		fx.Provide(ProvideSharedDatabaseConnection),
		fx.Provide(ProvideWalletManager),
		fx.Provide(ProvideUserService),
		fx.Provide(ProvideAuthService),
		fx.Provide(ProvideTransactionWorkerPool),
		fx.Provide(ProvideTronService),
		fx.Provide(ProvideTronWorkerPool),
		fx.Provide(ProvideTransactionService),
		// DDD Repositories
		fx.Provide(ProvideUserRepository),
		fx.Provide(ProvideWalletRepository),
		fx.Provide(ProvideTransactionRepository),
		fx.Provide(ProvideBlockchainTransactionRepository),
		// DDD Services
		fx.Provide(ProvideDDDUserService),
		fx.Provide(ProvideDDDTransactionService),
		fx.Invoke(StartServer),
	)
}
