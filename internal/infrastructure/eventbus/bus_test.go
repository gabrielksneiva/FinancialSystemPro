package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInMemoryStoreBasicOperations(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	event, err := NewBaseEvent("test.created", "agg-1", "test", map[string]string{"key": "value"})
	require.NoError(t, err)

	// Save
	err = store.SaveEvent(ctx, event)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := store.GetEventByID(ctx, event.EventID())
	require.NoError(t, err)
	require.Equal(t, event.EventID(), retrieved.EventID())

	// Get by aggregate
	events, err := store.GetEvents(ctx, "agg-1", 1)
	require.NoError(t, err)
	require.Len(t, events, 1)

	// Duplicate should fail
	err = store.SaveEvent(ctx, event)
	require.Error(t, err)
	require.Equal(t, ErrDuplicateEvent, err)
}

func TestInMemoryProcessingLogIdempotency(t *testing.T) {
	log := NewInMemoryProcessingLog()
	ctx := context.Background()

	// Not processed initially
	processed, err := log.IsProcessed(ctx, "evt-1", "handler-1")
	require.NoError(t, err)
	require.False(t, processed)

	// Mark processing
	err = log.MarkProcessing(ctx, "evt-1", "handler-1")
	require.NoError(t, err)

	// Still not completed
	processed, err = log.IsProcessed(ctx, "evt-1", "handler-1")
	require.NoError(t, err)
	require.False(t, processed)

	// Mark completed
	err = log.MarkCompleted(ctx, "evt-1", "handler-1")
	require.NoError(t, err)

	// Now processed
	processed, err = log.IsProcessed(ctx, "evt-1", "handler-1")
	require.NoError(t, err)
	require.True(t, processed)
}

func TestResilientBusPublishAndSubscribe(t *testing.T) {
	store := NewInMemoryStore()
	procLog := NewInMemoryProcessingLog()
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()
	var handled bool
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		handled = true
		mu.Unlock()
		return nil
	}

	err := bus.Subscribe("test.event", "test-handler", handler)
	require.NoError(t, err)

	event, _ := NewBaseEvent("test.event", "agg-1", "test", map[string]string{"data": "value"})
	err = bus.Publish(ctx, event)
	require.NoError(t, err)

	// Give async processing time
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	require.True(t, handled)
	mu.Unlock()

	// Verify event stored
	retrieved, err := store.GetEventByID(ctx, event.EventID())
	require.NoError(t, err)
	require.Equal(t, event.EventID(), retrieved.EventID())

	// Verify processing logged
	processed, err := procLog.IsProcessed(ctx, event.EventID(), "test-handler")
	require.NoError(t, err)
	require.True(t, processed)
}

func TestResilientBusIdempotency(t *testing.T) {
	store := NewInMemoryStore()
	procLog := NewInMemoryProcessingLog()
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()
	var callCount int
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	err := bus.Subscribe("idempotent.event", "idempotent-handler", handler)
	require.NoError(t, err)

	event, _ := NewBaseEvent("idempotent.event", "agg-2", "test", map[string]string{})

	// First publish
	err = bus.Publish(ctx, event)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	require.Equal(t, 1, callCount)
	mu.Unlock()

	// Try to publish same event again (should fail due to duplicate)
	err = bus.Publish(ctx, event)
	require.Error(t, err)

	// Manually dispatch same event (should be skipped due to idempotency)
	_ = bus.dispatch(ctx, event)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	require.Equal(t, 1, callCount, "handler should only be called once due to idempotency")
	mu.Unlock()
}

func TestResilientBusReplay(t *testing.T) {
	store := NewInMemoryStore()
	procLog := NewInMemoryProcessingLog()
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()

	// Store some events
	for i := 1; i <= 3; i++ {
		event, _ := NewBaseEvent("replay.event", "agg-replay", "test", map[string]int{"seq": i})
		event.Ver = i
		_ = store.SaveEvent(ctx, event)
	}

	var replayed []int
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		defer mu.Unlock()
		
		_ = event.Payload()
		replayed = append(replayed, len(replayed)+1)
		return nil
	}

	err := bus.Subscribe("replay.event", "replay-handler", handler)
	require.NoError(t, err)

	// Replay from version 2
	err = bus.Replay(ctx, "agg-replay", 2)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	require.Len(t, replayed, 2, "should replay events with version >= 2")
	mu.Unlock()
}

func TestResilientBusAsyncPublish(t *testing.T) {
	store := NewInMemoryStore()
	procLog := NewInMemoryProcessingLog()
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()
	var handled int
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		handled++
		mu.Unlock()
		return nil
	}

	err := bus.Subscribe("async.event", "async-handler", handler)
	require.NoError(t, err)

	// Publish multiple async events
	for i := 0; i < 5; i++ {
		event, _ := NewBaseEvent("async.event", "agg-async", "test", map[string]int{"i": i})
		bus.PublishAsync(ctx, event)
	}

	// Wait for async processing
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	require.Equal(t, 5, handled)
	mu.Unlock()
}

func TestResilientBusConcurrentPublish(t *testing.T) {
	store := NewInMemoryStore()
	procLog := NewInMemoryProcessingLog()
	logger := zap.NewNop()
	bus := NewResilientBus(store, procLog, logger)
	defer bus.Close()

	ctx := context.Background()
	var handled int
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		handled++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	err := bus.Subscribe("concurrent.event", "concurrent-handler", handler)
	require.NoError(t, err)

	// Concurrent publishes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event, _ := NewBaseEvent("concurrent.event", "agg-concurrent", "test", map[string]int{"idx": idx})
			_ = bus.Publish(ctx, event)
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	require.Equal(t, 10, handled, "all events should be handled")
	mu.Unlock()
}
