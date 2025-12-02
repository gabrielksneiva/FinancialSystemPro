package blockchain

import (
	"context"

	dom "github.com/gabrielksneiva/blockchains-utils/domain"
	rpc "github.com/gabrielksneiva/blockchains-utils/infra/rpc"
	"github.com/shopspring/decimal"

	bc "financial-system-pro/internal/blockchain"
)

// SimConnector adapts blockchains-utils' SimulatedClient to our bc.Connector interface for tests.
type SimConnector struct {
	c *rpc.SimulatedClient
}

func NewSimConnector() *SimConnector {
	return &SimConnector{c: rpc.NewSimulatedClient()}
}

func (s *SimConnector) FetchBalance(ctx context.Context, address bc.Address) (decimal.Decimal, error) {
	a, err := s.c.GetBalance(ctx, string(address))
	if err != nil {
		return decimal.Zero, err
	}
	if a.Value == nil {
		return decimal.Zero, nil
	}
	return decimal.NewFromBigInt(a.Value, 0), nil
}

func (s *SimConnector) SendTransaction(ctx context.Context, rawTx string) (bc.TxHash, error) {
	// Not a raw tx API in simulated client; create a Transaction domain and submit
	tx := dom.Transaction{Hash: rawTx}
	h, err := s.c.SubmitTransaction(ctx, tx)
	if err != nil {
		return "", err
	}
	return bc.TxHash(h), nil
}

func (s *SimConnector) GetTransactionStatus(ctx context.Context, hash bc.TxHash) (string, error) {
	tx, err := s.c.GetTransaction(ctx, string(hash))
	if err != nil {
		return "pending", nil
	}
	switch tx.Status {
	case dom.TxPending:
		return "pending", nil
	case dom.TxConfirmed:
		return "confirmed", nil
	default:
		return "unknown", nil
	}
}

func (s *SimConnector) SyncLatestBlocks(ctx context.Context, since uint64) ([]bc.Block, error) {
	b, err := s.c.GetLatestBlock(ctx)
	if err != nil {
		return nil, err
	}
	return []bc.Block{{Number: b.Number, Hash: b.Hash, Time: b.Time}}, nil
}

// Expose underlying simulated client for test setup
func (s *SimConnector) Underlying() *rpc.SimulatedClient { return s.c }
