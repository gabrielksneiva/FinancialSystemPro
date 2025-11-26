package mappers

import (
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/infrastructure/database/models"
)

// TransactionMapper converte entre entidade de domínio Transaction e TransactionModel (GORM)
type TransactionMapper struct{}

// ToModel converte entidade de domínio para modelo GORM
func (m TransactionMapper) ToModel(tx *entity.Transaction) *models.TransactionModel {
	if tx == nil {
		return nil
	}

	model := &models.TransactionModel{
		ID:          tx.ID,
		UserID:      tx.UserID,
		Amount:      tx.Amount,
		Type:        string(tx.Type),
		Status:      string(tx.Status),
		Hash:        tx.TransactionHash,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}

	if tx.CompletedAt != nil {
		model.CompletedAt = tx.CompletedAt
	}

	return model
}

// ToDomain converte modelo GORM para entidade de domínio
func (m TransactionMapper) ToDomain(model *models.TransactionModel) *entity.Transaction {
	if model == nil {
		return nil
	}

	tx := &entity.Transaction{
		ID:              model.ID,
		UserID:          model.UserID,
		Amount:          model.Amount,
		Type:            entity.TransactionType(model.Type),
		Status:          entity.TransactionStatus(model.Status),
		TransactionHash: model.Hash,
		FromAddress:     model.FromAddress,
		ToAddress:       model.ToAddress,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		CompletedAt:     model.CompletedAt,
	}

	return tx
}

// ToDomainList converte lista de modelos GORM para lista de entidades de domínio
func (m TransactionMapper) ToDomainList(models []*models.TransactionModel) []*entity.Transaction {
	if models == nil {
		return nil
	}

	transactions := make([]*entity.Transaction, 0, len(models))
	for _, model := range models {
		transactions = append(transactions, m.ToDomain(model))
	}

	return transactions
}
