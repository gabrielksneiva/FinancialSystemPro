package http

import (
	"context"
	"financial-system-pro/internal/infrastructure/config/container"
	"fmt"

	"github.com/joho/godotenv"
)

func Start() {
	// Carrega variáveis de ambiente do arquivo .env (opcional em produção)
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file:", err)
	} else {
		fmt.Println("✅ .env file loaded successfully")
	}

	fmt.Println("Initializing application with dependency injection...")

	// Cria a aplicação com fx
	app := container.New()

	// Inicia o fx com um context válido
	errStart := app.Start(context.Background())
	if errStart != nil {
		fmt.Printf("Error starting application: %v\n", errStart)
		return
	}

	// Aguarda o sinal de parada
	<-make(chan struct{})

	// Para a aplicação
	app.Stop(context.Background())
}
