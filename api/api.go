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
			fmt.Printf("Warning: Missing DB environment variables. DB_HOST=%s, DB_USER=%s, DB_NAME=%s, DB_PORT=%s\n", dbHost, dbUser, dbName, dbPort)
			if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
				fmt.Println("Error: Required database environment variables are missing.")
				os.Exit(1)
			}
			fmt.Println("Info: Running in Railway environment - will retry connection with available vars.")
		}
		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	}

	db := repositories.ConnectDatabase(connStr)

	database := &repositories.NewDatabase{DB: db}
	userService := &services.NewUserService{Database: database}
	authService := &services.NewAuthService{Database: database}
	tronService := services.NewTronService()

	workerPool := workers.NewTransactionWorkerPool(database, 5, 100)
	trasactionService := &services.NewTransactionService{DB: database, W: workerPool}

	app := fiber.New()

	router(app, userService, authService, trasactionService, tronService)

	app.Listen(":3000")
}
