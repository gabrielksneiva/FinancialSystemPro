package eventbus

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// ResilientBus implements Bus with persistence, idempotency, and retry capabilities.
type ResilientBus struct {
	store         Store
	processingLog ProcessingLog
	logger        *zap.Logger

	mu            sync.RWMutex
	subscriptions map[string][]subscription
	asyncQueue    chan asyncTask
	wg            sync.WaitGroup
	closeChan     chan struct{}
}

type subscription struct {
	handlerName string
	handler     Handler
}

type asyncTask struct {
	ctx   context.Context
	event Event
}

// NewResilientBus creates a new resilient event bus with persistent storage.
func NewResilientBus(store Store, processingLog ProcessingLog, logger *zap.Logger) *ResilientBus {
	bus := &ResilientBus{
		store:         store,
		processingLog: processingLog,
		logger:        logger,
		subscriptions: make(map[string][]subscription),
		asyncQueue:    make(chan asyncTask, 1000),
		closeChan:     make(chan struct{}),
	}

	// Start async workers
	for i := 0; i < 5; i++ {
		bus.wg.Add(1)
		go bus.asyncWorker()
	}

	return bus
}

func (b *ResilientBus) Publish(ctx context.Context, event Event) error {
	// First, persist the event
	if err := b.store.SaveEvent(ctx, event); err != nil {
		b.logger.Error("failed to save event",
			zap.String("event_id", event.EventID()),
			zap.String("event_type", event.EventType()),
			zap.Error(err))
		return fmt.Errorf("failed to save event: %w", err)
	}

	b.logger.Info("event saved",
		zap.String("event_id", event.EventID()),
		zap.String("event_type", event.EventType()),
		zap.String("aggregate_id", event.AggregateID()))

	// Dispatch to handlers
	return b.dispatch(ctx, event)
}

func (b *ResilientBus) PublishAsync(ctx context.Context, event Event) {
	select {
	case b.asyncQueue <- asyncTask{ctx: ctx, event: event}:
		b.logger.Debug("event queued for async processing",
			zap.String("event_id", event.EventID()),
			zap.String("event_type", event.EventType()))
	case <-b.closeChan:
		b.logger.Warn("bus closed, dropping async event",
			zap.String("event_id", event.EventID()))
	default:
		b.logger.Warn("async queue full, dropping event",
			zap.String("event_id", event.EventID()))
	}
}

func (b *ResilientBus) Subscribe(eventType string, handlerName string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscriptions[eventType] = append(b.subscriptions[eventType], subscription{
		handlerName: handlerName,
		handler:     handler,
	})

	b.logger.Info("handler subscribed",
		zap.String("event_type", eventType),
		zap.String("handler_name", handlerName),
		zap.Int("total_handlers", len(b.subscriptions[eventType])))

	return nil
}

func (b *ResilientBus) Replay(ctx context.Context, aggregateID string, fromVersion int) error {
	events, err := b.store.GetEvents(ctx, aggregateID, fromVersion)
	if err != nil {
		return fmt.Errorf("failed to get events for replay: %w", err)
	}

	b.logger.Info("replaying events",
		zap.String("aggregate_id", aggregateID),
		zap.Int("from_version", fromVersion),
		zap.Int("event_count", len(events)))

	for _, event := range events {
		if err := b.dispatch(ctx, event); err != nil {
			b.logger.Error("replay dispatch failed",
				zap.String("event_id", event.EventID()),
				zap.Error(err))
			// Continue with other events
		}
	}

	return nil
}

func (b *ResilientBus) Close() error {
	close(b.closeChan)
	close(b.asyncQueue)
	b.wg.Wait()
	b.logger.Info("event bus closed")
	return nil
}

func (b *ResilientBus) dispatch(ctx context.Context, event Event) error {
	b.mu.RLock()
	subs := b.subscriptions[event.EventType()]
	b.mu.RUnlock()

	if len(subs) == 0 {
		b.logger.Warn("no handlers for event type",
			zap.String("event_type", event.EventType()),
			zap.String("event_id", event.EventID()))
		return nil
	}

	var errors []error
	for _, sub := range subs {
		if err := b.handleWithIdempotency(ctx, event, sub); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d handlers failed", len(errors))
	}

	return nil
}

func (b *ResilientBus) handleWithIdempotency(ctx context.Context, event Event, sub subscription) error {
	// Check if already processed
	processed, err := b.processingLog.IsProcessed(ctx, event.EventID(), sub.handlerName)
	if err != nil {
		b.logger.Error("failed to check processing status",
			zap.String("event_id", event.EventID()),
			zap.String("handler", sub.handlerName),
			zap.Error(err))
		return err
	}

	if processed {
		b.logger.Debug("event already processed, skipping",
			zap.String("event_id", event.EventID()),
			zap.String("handler", sub.handlerName))
		return nil
	}

	// Mark as processing
	if err := b.processingLog.MarkProcessing(ctx, event.EventID(), sub.handlerName); err != nil {
		b.logger.Error("failed to mark processing",
			zap.String("event_id", event.EventID()),
			zap.String("handler", sub.handlerName),
			zap.Error(err))
		return err
	}

	// Execute handler
	handlerErr := sub.handler(ctx, event)

	// Mark result
	if handlerErr != nil {
		b.logger.Error("handler execution failed",
			zap.String("event_id", event.EventID()),
			zap.String("event_type", event.EventType()),
			zap.String("handler", sub.handlerName),
			zap.Error(handlerErr))
		_ = b.processingLog.MarkFailed(ctx, event.EventID(), sub.handlerName, handlerErr)
		return handlerErr
	}

	if err := b.processingLog.MarkCompleted(ctx, event.EventID(), sub.handlerName); err != nil {
		b.logger.Error("failed to mark completed",
			zap.String("event_id", event.EventID()),
			zap.String("handler", sub.handlerName),
			zap.Error(err))
		return err
	}

	b.logger.Info("event processed successfully",
		zap.String("event_id", event.EventID()),
		zap.String("handler", sub.handlerName))

	return nil
}

func (b *ResilientBus) asyncWorker() {
	defer b.wg.Done()

	for {
		select {
		case task, ok := <-b.asyncQueue:
			if !ok {
				return
			}
			if err := b.Publish(task.ctx, task.event); err != nil {
				b.logger.Error("async publish failed",
					zap.String("event_id", task.event.EventID()),
					zap.Error(err))
			}
		case <-b.closeChan:
			return
		}
	}
}
