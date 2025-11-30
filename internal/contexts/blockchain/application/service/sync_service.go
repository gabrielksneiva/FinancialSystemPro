package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	appsvc "financial-system-pro/internal/application/services"
	bc "financial-system-pro/internal/blockchain"
	"financial-system-pro/internal/contexts/blockchain/infrastructure/persistence"
)

// SyncService orchestrates connector syncs and persistence + outbox event creation.
type SyncService struct {
	Connector  bc.Connector
	BlocksRepo *persistence.BlockRepository
	Outbox     appsvc.EventsOutboxPort
}

func NewSyncService(conn bc.Connector, br *persistence.BlockRepository, out appsvc.EventsOutboxPort) *SyncService {
	return &SyncService{Connector: conn, BlocksRepo: br, Outbox: out}
}

// SyncBlocks fetches latest blocks since `since`, persists new ones and enqueues outbox events.
func (s *SyncService) SyncBlocks(ctx context.Context, since uint64) (int, error) {
	blocks, err := s.Connector.SyncLatestBlocks(ctx, since)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, b := range blocks {
		// check exists
		if _, err := s.BlocksRepo.GetByNumber(ctx, b.Number); err == nil {
			// already exists
			continue
		}
		bm := &persistence.BlockModel{Number: b.Number, Hash: b.Hash}
		if err := s.BlocksRepo.Save(ctx, bm); err != nil {
			continue
		}
		// create outbox event
		payload, _ := json.Marshal(map[string]interface{}{"number": b.Number, "hash": b.Hash})
		_ = s.Outbox.Save(ctx, &appsvc.OutboxRecord{ID: uuid.New(), Aggregate: "blockchain", Type: "block.detected", Payload: payload})
		created++
	}
	return created, nil
}
