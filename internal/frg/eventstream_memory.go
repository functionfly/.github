// Package frg provides an in-memory event stream implementation
// This is the default implementation for the Fly.io + Cloudflare stack
// NATS support is available via build tag: go build -tags nats
package frg

import (
	"sync"

	"github.com/google/uuid"
)

// InMemoryEventStream is a simple in-memory implementation for testing and single-node deployment
type InMemoryEventStream struct {
	subscribers map[uuid.UUID][]func(*GraphEvent)
	mu          sync.RWMutex
}

// NewInMemoryEventStream creates an in-memory event stream
func NewInMemoryEventStream() *InMemoryEventStream {
	return &InMemoryEventStream{
		subscribers: make(map[uuid.UUID][]func(*GraphEvent)),
	}
}

// Publish implements EventStream
func (s *InMemoryEventStream) Publish(event *GraphEvent) error {
	s.mu.RLock()
	handlers := s.subscribers[event.InstanceID]
	s.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}

	return nil
}

// Subscribe implements EventStream
func (s *InMemoryEventStream) Subscribe(instanceID uuid.UUID, handler func(*GraphEvent)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscribers[instanceID] = append(s.subscribers[instanceID], handler)
	return nil
}

// Close implements EventStream
func (s *InMemoryEventStream) Close() error {
	return nil
}
