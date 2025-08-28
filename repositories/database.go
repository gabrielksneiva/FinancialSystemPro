package repositories

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type NewDatabase struct {
	DB *gorm.DB
}

func (d *NewDatabase) Insert(value any) error {
	return d.DB.Create(value).Error
}

func (db *NewDatabase) FindUserByField(field string, value any) (*User, error) {
	var user User
	err := db.DB.Model(&User{}).Where(field+" = ?", value).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *NewDatabase) Deposit(userID uuid.UUID, amount decimal.Decimal) error {
	var balance Balance
	err := db.DB.Where("user_id = ?", userID).First(&balance).Error
	if err == gorm.ErrRecordNotFound {
		balance = Balance{
			UserID: userID,
			Amount: amount,
		}
		return db.DB.Create(&balance).Error
	} else if err != nil {
		return err
	}

	return db.DB.Model(&Balance{}).Where("user_id = ?", userID).Update("amount", gorm.Expr("amount + ?", amount)).Error
}

func (db *NewDatabase) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	var balance Balance
	err := db.DB.Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		return decimal.Zero, err
	}

	return balance.Amount, nil
}