package rpc

import (
	"context"
	"errors"
	"sync"
	"time"

	"financial-system-pro/internal/blockchains/domain"
)

var ErrNotFound = errors.New("not found")

// Client defines the minimal RPC client behaviour required by repositories.
type Client interface {
	Connect(ctx context.Context) error
	GetBalance(ctx context.Context, addr string) (domain.Amount, error)
	SubmitTransaction(ctx context.Context, tx domain.Transaction) (string, error)
	GetTransaction(ctx context.Context, hash string) (domain.Transaction, error)
	ConfirmTransaction(hash string, blockNum uint64)
	AdvanceBlock(number uint64)
	GetLatestBlock(ctx context.Context) (domain.Block, error)
	// Testing helpers
	SetBalance(addr string, a domain.Amount)
}

type SimulatedClient struct {
	mu          sync.RWMutex
	balances    map[string]*domain.Amount
	txs         map[string]domain.Transaction
	latestBlock domain.Block
	SubmitErr   error
}

func NewSimulatedClient() *SimulatedClient {
	return &SimulatedClient{
		balances:    make(map[string]*domain.Amount),
		txs:         make(map[string]domain.Transaction),
		latestBlock: domain.Block{Number: 0, Hash: "", Time: time.Now()},
	}
}

func (s *SimulatedClient) Connect(ctx context.Context) error {
	return nil
}

func (s *SimulatedClient) SetBalance(addr string, a domain.Amount) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[addr] = &a
}

func (s *SimulatedClient) GetBalance(ctx context.Context, addr string) (domain.Amount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.balances[addr]; ok {
		return *v, nil
	}
	return domain.Amount{Value: nil}, ErrNotFound
}

func (s *SimulatedClient) SubmitTransaction(ctx context.Context, tx domain.Transaction) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.SubmitErr != nil {
		return "", s.SubmitErr
	}
	if tx.Hash == "" {
		tx.Hash = "tx_" + time.Now().Format("20060102150405.000000")
	}
	tx.Status = domain.TxPending
	tx.CreatedAt = time.Now()
	s.txs[tx.Hash] = tx
	return tx.Hash, nil
}

func (s *SimulatedClient) GetTransaction(ctx context.Context, hash string) (domain.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if tx, ok := s.txs[hash]; ok {
		return tx, nil
	}
	return domain.Transaction{}, ErrNotFound
}

func (s *SimulatedClient) ConfirmTransaction(hash string, blockNum uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, ok := s.txs[hash]
	if !ok {
		return
	}
	tx.Status = domain.TxConfirmed
	tx.BlockNum = blockNum
	s.txs[hash] = tx
}

func (s *SimulatedClient) AdvanceBlock(number uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latestBlock = domain.Block{Number: number, Hash: "blk_" + time.Now().Format("20060102150405"), Time: time.Now()}
}

func (s *SimulatedClient) GetLatestBlock(ctx context.Context) (domain.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestBlock, nil
}
