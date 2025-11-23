package repositories

import (
	"time"

	"github.com/google/uuid"
)

// LogAudit cria um novo log de auditoria
func LogAudit(log *AuditLog) error {
	return DB.Create(log).Error
}

// GetAuditLogs busca logs de auditoria com filtros
func GetAuditLogs(userID uuid.UUID, limit int, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	err := DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetAuditLogsByAction busca logs de auditoria por ação
func GetAuditLogsByAction(action string, limit int, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	err := DB.
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// PurgeOldAuditLogs remove logs de auditoria antigos (retention policy)
func PurgeOldAuditLogs(daysOld int) error {
	cutoffTime := time.Now().AddDate(0, 0, -daysOld)
	return DB.Where("created_at < ?", cutoffTime).Delete(&AuditLog{}).Error
}
