package repositories

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dsn string) *gorm.DB {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return nil
	}

	err = DB.AutoMigrate(&User{}, &Account{}, &Balance{}, &Transaction{}, &AuditLog{})
	if err != nil {
		log.Printf("failed to migrate database: %v", err)
		return nil
	}

	return DB
}
