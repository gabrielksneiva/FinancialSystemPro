package api

import (
	"financial-system-pro/repositories"
	"financial-system-pro/services"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func Start() {
	// Carrega variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Erro ao carregar o arquivo .env")
		os.Exit(1)
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_ADMIN")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" {
		fmt.Println("One or more required environment variables (DB_HOST, DB_ADMIN, DB_PASSWORD, DB_NAME, DB_PORT) are missing.")
		os.Exit(1)
	}

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	db := repositories.ConnectDatabase(connStr)

	database := &repositories.NewDatabase{DB: db}
	userService := &services.NewUserService{Database: database}
	authService := &services.NewAuthService{Database: database}
	trasactionService := &services.NewTransactionService{Database: database}

	app := fiber.New()

	router(app, userService, authService, trasactionService)

	app.Listen(":3000")
}
