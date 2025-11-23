package http

import (
	"context"
	"financial-system-pro/internal/infrastructure/config/container"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func Start() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	// Carrega variáveis de ambiente do arquivo .env (opcional em produção)
	err := godotenv.Load()
	if err != nil {
		logger.Info("Warning: Error loading .env file",
			zap.Error(err),
		)
	} else {
		logger.Info(".env file loaded successfully")
	}

	logger.Info("Initializing application with dependency injection")

	// Cria a aplicação com fx
	app := container.New()

	// Inicia o fx com um context válido
	errStart := app.Start(context.Background())
	if errStart != nil {
		logger.Error("Error starting application",
			zap.Error(errStart),
		)
		return
	}

	// Aguarda o sinal de parada
	<-make(chan struct{})

	// Para a aplicação
	_ = app.Stop(context.Background())
}
