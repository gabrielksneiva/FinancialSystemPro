package adapters

import (
	"context"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type SolClient struct {
	backend rpc.Client
}

func NewSolClient(backend rpc.Client) *SolClient { return &SolClient{backend: backend} }

func (s *SolClient) Connect(ctx context.Context) error { return s.backend.Connect(ctx) }
func (s *SolClient) GetBalance(ctx context.Context, addr string) (domain.Amount, error) {
	return s.backend.GetBalance(ctx, addr)
}
func (s *SolClient) SubmitTransaction(ctx context.Context, tx domain.Transaction) (string, error) {
	return s.backend.SubmitTransaction(ctx, tx)
}
func (s *SolClient) GetTransaction(ctx context.Context, hash string) (domain.Transaction, error) {
	return s.backend.GetTransaction(ctx, hash)
}
func (s *SolClient) ConfirmTransaction(hash string, blockNum uint64) {
	s.backend.ConfirmTransaction(hash, blockNum)
}
func (s *SolClient) AdvanceBlock(number uint64) { s.backend.AdvanceBlock(number) }
func (s *SolClient) GetLatestBlock(ctx context.Context) (domain.Block, error) {
	return s.backend.GetLatestBlock(ctx)
}
