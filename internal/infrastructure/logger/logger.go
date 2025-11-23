package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// ProvideLogger cria e retorna uma instância do logger Zap
func ProvideLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	var logger *zap.Logger
	var err error

	if env == "production" {
		logger, err = zap.NewProduction()
	} else {
		// Development: colorful console output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}

	if err != nil {
		return nil, err
	}

	// Defer flush para garantir que logs sejam escritos
	defer func() { _ = logger.Sync() }()

	globalLogger = logger
	return logger, nil
}

// GetLogger retorna o logger global
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// Fallback se não foi inicializado via fx
		logger, _ := ProvideLogger()
		return logger
	}
	return globalLogger
}

// Info loga mensagem de informação
func Info(message string, fields ...zap.Field) {
	GetLogger().Info(message, fields...)
}

// Error loga mensagem de erro
func Error(message string, fields ...zap.Field) {
	GetLogger().Error(message, fields...)
}

// Warn loga mensagem de aviso
func Warn(message string, fields ...zap.Field) {
	GetLogger().Warn(message, fields...)
}

// Debug loga mensagem de debug
func Debug(message string, fields ...zap.Field) {
	GetLogger().Debug(message, fields...)
}

// Fatal loga mensagem fatal
func Fatal(message string, fields ...zap.Field) {
	GetLogger().Fatal(message, fields...)
}
