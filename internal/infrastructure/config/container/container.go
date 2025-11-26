package container

import (
	"context"
	"fmt"
	"os"
	"time"

	"financial-system-pro/internal/application/services"
	bcApp "financial-system-pro/internal/contexts/blockchain/application"
	bcGw "financial-system-pro/internal/contexts/blockchain/infrastructure/gateway"
	txnSvc "financial-system-pro/internal/contexts/transaction/application/service"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	txnPers "financial-system-pro/internal/contexts/transaction/infrastructure/persistence"
	userSvc "financial-system-pro/internal/contexts/user/application/service"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	userPers "financial-system-pro/internal/contexts/user/infrastructure/persistence"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/infrastructure/logger"
	messaging "financial-system-pro/internal/infrastructure/messaging"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/database"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/tracing"
	"financial-system-pro/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterRoutesFunc é uma função que registra rotas na aplicação
// Ela será fornecida pelo package api para evitar ciclo de import
type RegisterRoutesFunc func(
	app *fiber.App,
	dddUserService *userSvc.UserService,
	dddTransactionService *txnSvc.TransactionService,
	logger *zap.Logger,
	breakerManager *breaker.BreakerManager,
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

	conn, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ProvideDatabaseConnection cria a conexão com banco de dados (legacy)
func ProvideDatabaseConnection(cfg Config) (*repositories.NewDatabase, error) {
	if cfg.DatabaseURL == "" {
		fmt.Println("Warning: DATABASE_URL not set, running without database")
		return nil, nil
	}

	db, err := repositories.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &repositories.NewDatabase{DB: db}, nil
}

// ProvideUserService removed - replaced by DDD UserService in contexts/user/application/service
// ProvideAuthService removed - DDD contexts handle authentication via UserService.Authenticate

// ProvideTransactionWorkerPool cria o pool de workers para transações
// Legacy worker pool & multi-chain providers removed (TronService replaced by TronGateway)
// Ethereum/Bitcoin simple services removed during DDD consolidation; can be re-added via context-specific gateways later.

// ProvideApp cria a aplicação Fiber
func ProvideApp() *fiber.App {
	return fiber.New()
}

// ProvideLogger cria o logger Zap centralizado
func ProvideLogger() (*zap.Logger, error) {
	return logger.ProvideLogger()
}

// ProvideTronGateway constrói TronGateway usando variáveis de ambiente
func ProvideTronGateway() *bcGw.TronGateway {
	return bcGw.NewTronGatewayFromEnv()
}

// ProvideETHGateway constrói Ethereum gateway
func ProvideETHGateway() *bcGw.ETHGateway { return bcGw.NewETHGatewayFromEnv() }

// ProvideBTCGateway constrói Bitcoin gateway
func ProvideBTCGateway() *bcGw.BTCGateway { return bcGw.NewBTCGatewayFromEnv() }

// ProvideSOLGateway constrói Solana gateway
func ProvideSOLGateway() *bcGw.SOLGateway { return bcGw.NewSOLGatewayFromEnv() }

// ProvideDDDBlockchainRegistry monta registro DDD de blockchains
func ProvideDDDBlockchainRegistry(tron *bcGw.TronGateway, eth *bcGw.ETHGateway, btc *bcGw.BTCGateway, sol *bcGw.SOLGateway) *bcApp.BlockchainRegistry {
	return bcApp.NewBlockchainRegistry(tron, eth, btc, sol)
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
func ProvideWalletManager() entities.WalletManager { return services.NewTronWalletManager() }

// StartServer inicia o servidor Fiber e workers
func StartServer(
	lc fx.Lifecycle,
	app *fiber.App,
	lg *zap.Logger,
	eventBus events.Bus,
	registerRoutes RegisterRoutesFunc,
	breakerManager *breaker.BreakerManager,
	db *repositories.NewDatabase,
	// DDD Services
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

			// Registrar apenas rotas DDD se disponíveis, senão health checks
			if registerRoutes != nil && dddUserService != nil && dddTransactionService != nil {
				lg.Info("registering DDD v2 routes")
				registerRoutes(app, dddUserService, dddTransactionService, lg, breakerManager)
			} else {
				lg.Warn("DDD services missing; registering health checks only")
				registerFiberHealthChecks(app)
			}

			// Iniciar OutboxProcessor periódico se banco disponível
			if db != nil {
				store := repositories.NewGormOutboxStore(db)
				processor := messaging.NewOutboxProcessor(store, eventBus, lg)
				lg.Info("starting outbox processor goroutine")
				go func() {
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							lg.Info("outbox processor stopping")
							return
						case <-ticker.C:
							if err := processor.ProcessBatch(context.Background(), 50); err != nil {
								lg.Warn("outbox batch failed", zap.Error(err))
							}
						}
					}
				}()
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

			// queue manager removido na reconstrução
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
// Queue manager removed in reconstruction (async Redis queue deprecated in DDD phase)

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

// ProvideBlockchainTransactionRepository removed - no longer needed in DDD refactor
// (blockchain transactions handled via blockchain context gateway now)

// New cria a aplicação com todas as dependências gerenciadas por fx
func New() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),
		fx.Provide(ProvideLogger),
		fx.Provide(ProvideEventBus),
		fx.Provide(ProvideBreakerManager),
		fx.Provide(ProvideValidator),
		fx.Provide(ProvideRegisterRoutes),
		fx.Provide(ProvideApp),
		fx.Provide(ProvideDatabaseConnection),
		fx.Provide(ProvideSharedDatabaseConnection),
		fx.Provide(ProvideWalletManager),
		// Legacy ProvideUserService and ProvideAuthService removed - using DDD services
		// TronGateway + DDD blockchain registry
		fx.Provide(ProvideTronGateway),
		fx.Provide(ProvideETHGateway),
		fx.Provide(ProvideBTCGateway),
		fx.Provide(ProvideSOLGateway),
		fx.Provide(ProvideDDDBlockchainRegistry),
		// DDD repositories & services
		fx.Provide(ProvideUserRepository),
		fx.Provide(ProvideWalletRepository),
		fx.Provide(ProvideTransactionRepository),
		fx.Provide(ProvideDDDUserService),
		fx.Provide(ProvideDDDTransactionService),
		fx.Invoke(StartServer),
	)
}

// LinkUserMultiChain associa MultiChainWalletService ao UserService legacy
// Legacy link functions removidos (serviços legacy eliminados)
