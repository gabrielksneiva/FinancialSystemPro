package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	appsvc "financial-system-pro/internal/application/services"
	bc "financial-system-pro/internal/blockchain"
	"financial-system-pro/internal/contexts/blockchain/infrastructure/persistence"
)

// TransactionMonitor checks pending transactions, updates confirmations and enqueues events.
type TransactionMonitor struct {
	Connector bc.Connector
	TxRepo    *persistence.TransactionRepository
	Outbox    appsvc.EventsOutboxPort
}

func NewTransactionMonitor(conn bc.Connector, txr *persistence.TransactionRepository, out appsvc.EventsOutboxPort) *TransactionMonitor {
	return &TransactionMonitor{Connector: conn, TxRepo: txr, Outbox: out}
}

func (m *TransactionMonitor) CheckPending(ctx context.Context) (int, error) {
	pending, err := m.TxRepo.ListPending(ctx)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, t := range pending {
		st, err := m.Connector.GetTransactionStatus(ctx, bc.TxHash(t.Hash))
		if err != nil {
			continue
		}
		if st == "confirmed" {
			// mark confirmed in DB (set confirmations >0)
			_ = m.TxRepo.UpdateConfirmations(ctx, t.Hash, 1)
			// create outbox event
			payload, _ := json.Marshal(map[string]interface{}{"hash": t.Hash, "status": "confirmed"})
			_ = m.Outbox.Save(ctx, &appsvc.OutboxRecord{ID: uuid.New(), Aggregate: "transaction", Type: "transaction.confirmed", Payload: payload})
			updated++
		}
	}
	return updated, nil
}
