package repositories

import (
	"strings"

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

func (d *NewDatabase) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return d.DB.Model(&Transaction{}).Where("id = ?", txID).Updates(updates).Error
}

func (db *NewDatabase) FindUserByField(field string, value any) (*User, error) {
	var user User
	err := db.DB.Model(&User{}).Where(field+" = ?", value).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *NewDatabase) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	if strings.EqualFold(transactionType, "withdraw") {
		err := db.DB.Model(&Balance{}).Where("user_id = ? AND amount >= ?", userID, amount).Update("amount", gorm.Expr("amount - ?", amount)).Error
		if err != nil {
			return err
		}
		return nil
	}

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

// SaveWalletInfo salva as informações de carteira TRON do usuário (private key criptografada)
func (db *NewDatabase) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	wallet := WalletInfo{
		UserID:           userID,
		TronAddress:      tronAddress,
		EncryptedPrivKey: encryptedPrivKey,
	}
	return db.DB.Create(&wallet).Error
}

// GetWalletInfo retorna as informações de carteira TRON do usuário
func (db *NewDatabase) GetWalletInfo(userID uuid.UUID) (*WalletInfo, error) {
	var wallet WalletInfo
	err := db.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetWalletInfoByAddress retorna as informações de carteira por endereço TRON
func (db *NewDatabase) GetWalletInfoByAddress(tronAddress string) (*WalletInfo, error) {
	var wallet WalletInfo
	err := db.DB.Where("tron_address = ?", tronAddress).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// UpdateWalletInfo atualiza as informações de carteira TRON
func (db *NewDatabase) UpdateWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return db.DB.Model(&WalletInfo{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"tron_address":       tronAddress,
			"encrypted_priv_key": encryptedPrivKey,
		}).Error
}
