package repositories

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dsn string) *gorm.DB {
	var err error
	maxRetries := 5
	var retryDelay time.Duration = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		log.Printf("attempt %d/%d failed to connect to database: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			log.Printf("retrying in %v (serverless database may be waking up)...", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // exponential backoff
		}
	}

	if err != nil {
		log.Fatalf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	log.Println("successfully connected to database")

	err = DB.AutoMigrate(&User{}, &Account{}, &Balance{}, &Transaction{}, &AuditLog{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return DB
}
