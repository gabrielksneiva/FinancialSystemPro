package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TransactionLog registra todas as transações para auditoria
func (db *NewDatabase) LogTransaction(userID uuid.UUID, action string, payload datatypes.JSON, result string) error {
	log := AuditLog{
		UserID:     userID,
		Action:     action,
		NewPayload: payload,
		CreatedAt:  time.Now(),
	}

	// Adicionar result ao payload se existir
	if result != "" {
		log.OldPayload = datatypes.JSON([]byte(result))
	}

	return db.DB.Create(&log).Error
}

// GetUserTransactionHistory retorna histórico de transações do usuário
func (db *NewDatabase) GetUserTransactionHistory(userID uuid.UUID, limit int) ([]Transaction, error) {
	var transactions []Transaction
	err := db.DB.Where("account_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionStats retorna estatísticas de transações
type TransactionStats struct {
	TotalDeposits  int64
	TotalWithdraws int64
	TotalTransfers int64
	DepositVolume  decimal.Decimal
	WithdrawVolume decimal.Decimal
	TransferVolume decimal.Decimal
}

func (db *NewDatabase) GetTransactionStats(userID uuid.UUID) (*TransactionStats, error) {
	var stats TransactionStats

	// Total de cada tipo
	db.DB.Model(&Transaction{}).Where("account_id = ? AND type = ?", userID, "deposit").Count(&stats.TotalDeposits)
	db.DB.Model(&Transaction{}).Where("account_id = ? AND type = ?", userID, "withdraw").Count(&stats.TotalWithdraws)
	db.DB.Model(&Transaction{}).Where("account_id = ? AND type = ?", userID, "transfer").Count(&stats.TotalTransfers)

	// Volume de cada tipo
	var result struct {
		DepositVolume  decimal.Decimal
		WithdrawVolume decimal.Decimal
		TransferVolume decimal.Decimal
	}

	db.DB.Model(&Transaction{}).
		Select(
			"COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE 0 END), 0) as deposit_volume",
			"COALESCE(SUM(CASE WHEN type = 'withdraw' THEN amount ELSE 0 END), 0) as withdraw_volume",
			"COALESCE(SUM(CASE WHEN type = 'transfer' THEN amount ELSE 0 END), 0) as transfer_volume",
		).
		Where("account_id = ?", userID).
		Scan(&result)

	stats.DepositVolume = result.DepositVolume
	stats.WithdrawVolume = result.WithdrawVolume
	stats.TransferVolume = result.TransferVolume

	return &stats, nil
}

// IsLargeTransaction verifica se é uma transação grande (para alert)
func (db *NewDatabase) IsLargeTransaction(amount decimal.Decimal) bool {
	// Defina threshold baseado em sua moeda
	threshold := decimal.NewFromInt(10000) // 10000 units
	return amount.GreaterThan(threshold)
}

// HasSuspiciousActivity verifica atividade suspeita (múltiplas falhas, etc)
func (db *NewDatabase) HasSuspiciousActivity(userID uuid.UUID, timeWindow time.Duration) (bool, error) {
	var count int64

	err := db.DB.Model(&AuditLog{}).
		Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-timeWindow)).
		Where("action LIKE ?", "%fail%").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	// Mais de 5 falhas em 5 minutos é suspeito
	return count > 5, nil
}
