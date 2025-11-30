package adapters

import (
	"context"
	"time"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

// EthClient is a lightweight facade for Ethereum RPC interactions.
// For now it delegates to a wrapped rpc.Client (e.g. SimulatedClient) but
// this file is the place to implement a real go-ethereum client.
type EthClient struct {
	backend rpc.Client
}

func NewEthClient(backend rpc.Client) *EthClient {
	return &EthClient{backend: backend}
}

func (e *EthClient) Connect(ctx context.Context) error {
	return e.backend.Connect(ctx)
}

func (e *EthClient) GetBalance(ctx context.Context, addr string) (domain.Amount, error) {
	return e.backend.GetBalance(ctx, addr)
}

func (e *EthClient) SubmitTransaction(ctx context.Context, tx domain.Transaction) (string, error) {
	return e.backend.SubmitTransaction(ctx, tx)
}

func (e *EthClient) GetTransaction(ctx context.Context, hash string) (domain.Transaction, error) {
	return e.backend.GetTransaction(ctx, hash)
}

func (e *EthClient) ConfirmTransaction(hash string, blockNum uint64) {
	e.backend.ConfirmTransaction(hash, blockNum)
}

func (e *EthClient) AdvanceBlock(number uint64) {
	e.backend.AdvanceBlock(number)
}

func (e *EthClient) GetLatestBlock(ctx context.Context) (domain.Block, error) {
	return e.backend.GetLatestBlock(ctx)
}

// For completeness, provide a method to simulate delays in real adapter
func (e *EthClient) simulateNetworkDelay() {
	time.Sleep(5 * time.Millisecond)
}
