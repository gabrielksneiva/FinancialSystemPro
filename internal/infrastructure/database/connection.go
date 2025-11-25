package repositories

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Testability hooks
var (
	gormOpen func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) = gorm.Open
	sleepFn                                                         = time.Sleep
)

// ConnectDatabase tenta conectar com retries e retorna erro em vez de finalizar o processo.
func ConnectDatabase(dsn string) (*gorm.DB, error) {
	return connectDatabaseWithRetry(dsn, 5, 2*time.Second)
}

// connectDatabaseWithRetry parametriza retries para facilitar testes.
func connectDatabaseWithRetry(dsn string, maxRetries int, initialDelay time.Duration) (*gorm.DB, error) {
	var err error
	retryDelay := initialDelay
	for attempt := 1; attempt <= maxRetries; attempt++ {
		DB, err = gormOpen(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("successfully connected to database")
			return DB, nil
		}
		log.Printf("attempt %d/%d failed to connect to database: %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			log.Printf("retrying in %v (serverless database may be waking up)...", retryDelay)
			sleepFn(retryDelay)
			retryDelay *= 2
		}
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
