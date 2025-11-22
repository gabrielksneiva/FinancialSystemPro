package api

import (
	"context"
	"financial-system-pro/internal/container"
	"fmt"

	"github.com/joho/godotenv"
)

func Start() {
	// Carrega variáveis de ambiente do arquivo .env (opcional em produção)
	_ = godotenv.Load()

	fmt.Println("Initializing application with dependency injection...")

	// Cria a aplicação com fx
	app := container.New()

	// Inicia o fx com um context válido
	err := app.Start(context.Background())
	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		return
	}

	// Aguarda o sinal de parada
	<-make(chan struct{})

	// Para a aplicação
	app.Stop(context.Background())
}
