package api

import (
	"financial-system-pro/repositories"
	"financial-system-pro/services"
	"financial-system-pro/workers"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func Start() {
	// Carrega variáveis de ambiente do arquivo .env (opcional em produção)
	_ = godotenv.Load()

	// Tentar usar DATABASE_URL do Railway primeiro, depois fallback para variáveis individuais
	var connStr string
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		connStr = dbURL
		fmt.Println("Using DATABASE_URL from environment")
	} else {
		dbHost := os.Getenv("DB_HOST")
		dbUser := os.Getenv("DB_ADMIN")
		if dbUser == "" {
			dbUser = os.Getenv("DB_USER")
		}
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dbPort := os.Getenv("DB_PORT")

		if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" {
			fmt.Printf("Warning: Missing DB environment variables.\n")
			fmt.Println("Info: Will attempt to connect later or use in-memory mode")
		}
		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	}

	// Inicia o app e o servidor HTTP ANTES de tentar conectar ao banco
	app := fiber.New()

	// Serviços que não dependem de banco
	tronService := services.NewTronService()

	// Services e database são inicializados em background
	var userService *services.NewUserService
	var authService *services.NewAuthService
	var trasactionService *services.NewTransactionService

	// Tenta conectar ao banco em background
	go func() {
		db := repositories.ConnectDatabase(connStr)
		if db != nil {
			database := &repositories.NewDatabase{DB: db}
			userService = &services.NewUserService{Database: database}
			authService = &services.NewAuthService{Database: database}

			workerPool := workers.NewTransactionWorkerPool(database, 5, 100)
			trasactionService = &services.NewTransactionService{DB: database, W: workerPool}
			fmt.Println("Database connected successfully")
		} else {
			fmt.Println("Warning: Could not connect to database, some features may be unavailable")
		}
	}()

	// Router usará os services (que podem estar nil inicialmente)
	router(app, userService, authService, trasactionService, tronService)

	// Inicia o servidor - isto vai servir /health mesmo que o banco esteja indisponível
	fmt.Println("Starting server on :3000")
	app.Listen(":3000")
}
