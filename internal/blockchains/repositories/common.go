package repositories

import (
	"context"
	"fmt"
	"time"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/events"
	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type BaseRepo struct {
	Chain  string
	client rpc.Client
	bus    *eventbus.InMemoryBus
}

func (b *BaseRepo) Connect(ctx context.Context) error {
	return b.client.Connect(ctx)
}

func (b *BaseRepo) GetBalance(ctx context.Context, addr domain.Address) (domain.Amount, error) {
	return b.client.GetBalance(ctx, string(addr))
}

func (b *BaseRepo) CreateTransaction(ctx context.Context, from domain.Address, to domain.Address, amount domain.Amount) (domain.Transaction, error) {
	tx := domain.Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Status:    domain.TxPending,
		CreatedAt: time.Now(),
	}
	hash, _ := b.client.SubmitTransaction(ctx, tx)
	tx.Hash = hash
	// publish new tx
	b.bus.Publish(events.EventNewTransaction, b.Chain+":"+hash, events.NewTransactionEvent{Chain: b.Chain, TxHash: hash})
	return tx, nil
}

func (b *BaseRepo) BroadcastTransaction(ctx context.Context, tx domain.Transaction) (domain.TxHash, error) {
	h, err := b.client.SubmitTransaction(ctx, tx)
	if err != nil {
		return domain.TxHash(""), err
	}
	return domain.TxHash(h), nil
}

func (b *BaseRepo) GetTransactionStatus(ctx context.Context, hash domain.TxHash) (domain.TxStatus, error) {
	tx, err := b.client.GetTransaction(ctx, string(hash))
	if err != nil {
		return "", err
	}
	return tx.Status, nil
}

func (b *BaseRepo) GetTransaction(ctx context.Context, hash domain.TxHash) (domain.Transaction, error) {
	return b.client.GetTransaction(ctx, string(hash))
}

func (b *BaseRepo) GetLatestBlock(ctx context.Context) (domain.Block, error) {
	return b.client.GetLatestBlock(ctx)
}

func (b *BaseRepo) WatchBlocks(ctx context.Context) (<-chan interface{}, error) {
	ch, _ := b.bus.Subscribe(ctx, events.EventNewBlock)
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				out <- v
			}
		}
	}()
	return out, nil
}

// ConfirmTransaction confirms a transaction on the underlying client and
// emits both NewBlockEvent and TransactionConfirmedEvent via the event bus.
// Idempotency is handled by the bus using the provided event IDs.
func (b *BaseRepo) ConfirmTransaction(hash domain.TxHash, blockNum uint64) error {
	// confirm at client
	b.client.ConfirmTransaction(string(hash), blockNum)

	// publish new block event
	blkID := b.Chain + ":block:" + fmt.Sprint(blockNum)
	b.bus.Publish(events.EventNewBlock, blkID, events.NewBlockEvent{Chain: b.Chain, BlockNumber: blockNum, BlockHash: "blk_" + fmt.Sprint(blockNum), Timestamp: time.Now()})

	// publish tx confirmed event
	txID := b.Chain + ":tx:" + string(hash)
	b.bus.Publish(events.EventTransactionConfirm, txID, events.TransactionConfirmedEvent{Chain: b.Chain, TxHash: string(hash), BlockNumber: blockNum})

	return nil
}
