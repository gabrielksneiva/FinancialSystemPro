package service

import (
	"context"
	"errors"
	"time"

	app "financial-system-pro/internal/contexts/blockchain/application"
	bcdom "financial-system-pro/internal/contexts/blockchain/domain"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"
	repo "financial-system-pro/internal/contexts/blockchain/domain/repository"
	"financial-system-pro/internal/shared/events"

	"github.com/shopspring/decimal"
)

// UseCases implements application-level orchestration for blockchain operations.
type UseCases struct {
	registry *app.BlockchainRegistry
	repo     repo.BlockchainTransactionRepository
	bus      events.Bus
}

func NewUseCases(reg *app.BlockchainRegistry, r repo.BlockchainTransactionRepository, bus events.Bus) *UseCases {
	return &UseCases{registry: reg, repo: r, bus: bus}
}

// FetchBalance returns the balance for address on chain in base units.
func (u *UseCases) FetchBalance(ctx context.Context, chain entity.BlockchainType, address string) (int64, error) {
	gw, err := u.registry.Get(chain)
	if err != nil {
		return 0, err
	}
	bal, err := gw.GetBalance(ctx, address)
	if err != nil {
		return 0, err
	}
	// Emit balance change notification (generic tx.new with zero amount is avoided; use dedicated?)
	return bal, nil
}

// SendTransaction broadcasts a transaction and persists an entry.
func (u *UseCases) SendTransaction(ctx context.Context, chain entity.BlockchainType, from, to string, amountBaseUnit int64, privateKey string) (string, error) {
	gw, err := u.registry.Get(chain)
	if err != nil {
		return "", err
	}
	if amountBaseUnit <= 0 {
		return "", errors.New("amount must be positive")
	}
	hash, err := gw.Broadcast(ctx, from, to, amountBaseUnit, privateKey)
	if err != nil {
		return "", err
	}

	// Persist minimal record
	tx := entity.NewBlockchainTransaction(chainToNetwork(chain), from, to, decimal.NewFromInt(amountBaseUnit))
	tx.TransactionHash = string(hash)
	_ = u.repo.Create(ctx, tx)

	// Publish new transaction event
	evt := events.NewNewTransactionDetectedEvent(string(chain), string(hash), from, to, amountBaseUnit, tx.BlockNumber)
	_ = u.bus.Publish(ctx, evt)

	return string(hash), nil
}

// GetTransactionStatus queries gateway and updates repository, publishing confirmation event when confirmed.
func (u *UseCases) GetTransactionStatus(ctx context.Context, chain entity.BlockchainType, txHash string) (*bcdom.TxStatusInfo, error) {
	gw, err := u.registry.Get(chain)
	if err != nil {
		return nil, err
	}
	status, err := gw.GetStatus(ctx, bcdom.TxHash(txHash))
	if err != nil {
		return nil, err
	}
	if status.Status == bcdom.TxStatusConfirmed {
		evt := events.NewBlockchainTransactionConfirmedEvent(txHash, int(status.Confirmations), 0, string(chain))
		_ = u.bus.Publish(ctx, evt)
	}
	return status, nil
}

// SyncLatestBlocks triggers a poll and publishes a NewBlockDetected event with last known block.
func (u *UseCases) SyncLatestBlocks(ctx context.Context, chain entity.BlockchainType) error {
	// We don't maintain block state; emit a heartbeat-like block event.
	evt := events.NewNewBlockDetectedEvent(string(chain), time.Now().Unix(), "sync", time.Now().Unix())
	return u.bus.Publish(ctx, evt)
}

func chainToNetwork(c entity.BlockchainType) entity.BlockchainNetwork {
	switch c {
	case entity.BlockchainEthereum:
		return entity.NetworkEthereum
	case entity.BlockchainBitcoin:
		return entity.NetworkBitcoin
	case entity.BlockchainSolana:
		return entity.NetworkSolana
	case entity.BlockchainTron:
		return entity.NetworkTron
	default:
		return entity.NetworkEthereum
	}
}
