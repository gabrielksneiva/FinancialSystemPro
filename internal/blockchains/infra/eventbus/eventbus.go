package eventbus

import (
	"context"
	"sync"

	"financial-system-pro/internal/blockchains/events"
)

// Simple in-memory event bus with idempotency using a map
type InMemoryBus struct {
	mu        sync.RWMutex
	subs      map[events.EventType][]chan interface{}
	processed map[string]struct{}
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		subs:      make(map[events.EventType][]chan interface{}),
		processed: make(map[string]struct{}),
	}
}

func (b *InMemoryBus) Publish(evtType events.EventType, id string, payload interface{}) {
	b.mu.Lock()
	if _, ok := b.processed[id]; ok {
		b.mu.Unlock()
		return
	}
	b.processed[id] = struct{}{}
	subs := append([]chan interface{}{}, b.subs[evtType]...)
	b.mu.Unlock()

	for _, ch := range subs {
		// deliver async
		go func(c chan interface{}) { c <- payload }(ch)
	}
}

func (b *InMemoryBus) Subscribe(ctx context.Context, evtType events.EventType) (<-chan interface{}, func()) {
	ch := make(chan interface{}, 16)
	b.mu.Lock()
	b.subs[evtType] = append(b.subs[evtType], ch)
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subs[evtType]
		for i, c := range subs {
			if c == ch {
				b.subs[evtType] = append(subs[:i], subs[i+1:]...)
				close(c)
				break
			}
		}
	}
	return ch, cancel
}
