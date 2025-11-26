package domain

import (
	"context"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"
	"testing"
)

type dummyGateway struct{}

func (d *dummyGateway) GenerateWallet(ctx context.Context) (*entity.GeneratedWallet, error) {
	return nil, nil
}
func (d *dummyGateway) ValidateAddress(address string) bool { return true }
func (d *dummyGateway) EstimateFee(ctx context.Context, from, to string, amount int64) (*FeeQuote, error) {
	return nil, nil
}
func (d *dummyGateway) Broadcast(ctx context.Context, from, to string, amount int64, priv string) (TxHash, error) {
	return "", nil
}
func (d *dummyGateway) GetStatus(ctx context.Context, hash TxHash) (*TxStatusInfo, error) {
	return nil, nil
}
func (d *dummyGateway) ChainType() entity.BlockchainType                              { return entity.BlockchainEthereum }
func (d *dummyGateway) GetBalance(ctx context.Context, address string) (int64, error) { return 0, nil }
func (d *dummyGateway) GetTransactionHistory(ctx context.Context, address string, limit, offset int) ([]*entity.BlockchainTransaction, error) {
	return nil, nil
}
func (d *dummyGateway) SubscribeNewBlocks(ctx context.Context, handler BlockEventHandler) error {
	return nil
}
func (d *dummyGateway) SubscribeNewTransactions(ctx context.Context, address string, handler TxEventHandler) error {
	return nil
}

func TestBlockchainGatewayPort_Contract(t *testing.T) {
	var _ BlockchainGatewayPort = &dummyGateway{}
}
