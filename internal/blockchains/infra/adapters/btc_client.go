package adapters

import (
	"context"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type BtcClient struct {
	backend rpc.Client
}

func NewBtcClient(backend rpc.Client) *BtcClient { return &BtcClient{backend: backend} }

func (b *BtcClient) Connect(ctx context.Context) error { return b.backend.Connect(ctx) }
func (b *BtcClient) GetBalance(ctx context.Context, addr string) (domain.Amount, error) {
	return b.backend.GetBalance(ctx, addr)
}
func (b *BtcClient) SubmitTransaction(ctx context.Context, tx domain.Transaction) (string, error) {
	return b.backend.SubmitTransaction(ctx, tx)
}
func (b *BtcClient) GetTransaction(ctx context.Context, hash string) (domain.Transaction, error) {
	return b.backend.GetTransaction(ctx, hash)
}
func (b *BtcClient) ConfirmTransaction(hash string, blockNum uint64) {
	b.backend.ConfirmTransaction(hash, blockNum)
}
func (b *BtcClient) AdvanceBlock(number uint64) { b.backend.AdvanceBlock(number) }
func (b *BtcClient) GetLatestBlock(ctx context.Context) (domain.Block, error) {
	return b.backend.GetLatestBlock(ctx)
}
