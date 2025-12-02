package persistence

import (
	"context"

	"gorm.io/gorm"
)

// Repositories using GORM for blockchain models.
type BlockRepository struct{ db *gorm.DB }

func NewBlockRepository(db *gorm.DB) *BlockRepository { return &BlockRepository{db: db} }

func (r *BlockRepository) Save(ctx context.Context, b *BlockModel) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *BlockRepository) GetByNumber(ctx context.Context, number uint64) (*BlockModel, error) {
	var bm BlockModel
	if err := r.db.WithContext(ctx).Where("number = ?", number).First(&bm).Error; err != nil {
		return nil, err
	}
	return &bm, nil
}

type TransactionRepository struct{ db *gorm.DB }

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(ctx context.Context, t *TransactionModel) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TransactionRepository) GetByHash(ctx context.Context, hash string) (*TransactionModel, error) {
	var tm TransactionModel
	if err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&tm).Error; err != nil {
		return nil, err
	}
	return &tm, nil
}

func (r *TransactionRepository) ListPending(ctx context.Context) ([]*TransactionModel, error) {
	var rows []*TransactionModel
	if err := r.db.WithContext(ctx).Where("confirmations = 0").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TransactionRepository) UpdateConfirmations(ctx context.Context, hash string, confirmations int64) error {
	return r.db.WithContext(ctx).Model(&TransactionModel{}).Where("hash = ?", hash).Updates(map[string]interface{}{"confirmations": confirmations}).Error
}

type BalanceRepository struct{ db *gorm.DB }

func NewBalanceRepository(db *gorm.DB) *BalanceRepository { return &BalanceRepository{db: db} }

func (r *BalanceRepository) Upsert(ctx context.Context, address string, balance string) error {
	// try update, else create
	bm := BalanceModel{Address: address, Balance: balance}
	return r.db.WithContext(ctx).Clauses().Save(&bm).Error
}

func (r *BalanceRepository) Get(ctx context.Context, address string) (*BalanceModel, error) {
	var bm BalanceModel
	if err := r.db.WithContext(ctx).Where("address = ?", address).First(&bm).Error; err != nil {
		return nil, err
	}
	return &bm, nil
}
