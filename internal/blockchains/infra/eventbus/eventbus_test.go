package eventbus

import (
    "context"
    "testing"
    "time"

    "financial-system-pro/internal/blockchains/events"
)

// This comment indicates the start of the Go code
func TestInMemoryBus_PublishSubscribe_Idempotency(t *testing.T) {
    b := NewInMemoryBus()
    ctx := context.Background()
    ch, cancel := b.Subscribe(ctx, events.EventNewBlock)
    defer cancel()

    b.Publish(events.EventNewBlock, "id1", events.NewBlockEvent{Chain: "ethereum", BlockNumber: 1})
    select {
    case v := <-ch:
        nb, ok := v.(events.NewBlockEvent)
        if !ok {
            t.Fatalf("unexpected type")
        }
        if nb.BlockNumber != 1 {
            t.Fatalf("bad block number")
        }
    case <-time.After(time.Second):
        t.Fatalf("timeout waiting event")
    }

    // publish same id again -> should be ignored
    b.Publish(events.EventNewBlock, "id1", events.NewBlockEvent{Chain: "ethereum", BlockNumber: 1})
    select {
    case <-ch:
        t.Fatalf("should not receive duplicate")
    case <-time.After(100 * time.Millisecond):
        // ok
    }
}
