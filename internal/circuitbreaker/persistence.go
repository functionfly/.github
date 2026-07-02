package circuitbreaker

import (
	"context"
	"sync"
	"time"
)

// StoredState represents the persisted state of a circuit breaker.
type StoredState struct {
	State       int
	FailCount   int
	ReopenCount int
	Since       time.Time
	LastFailure time.Time
}

// Persistence defines the interface for persisting circuit breaker state.
// Implementations should be safe for concurrent use.
type Persistence interface {
	// Load retrieves the stored state for a given key.
	// Returns nil if no state is stored (not an error).
	Load(ctx context.Context, key string) (*StoredState, error)

	// Save persists the state for a given key.
	Save(ctx context.Context, key string, state *StoredState) error
}

// AsyncPersistence wraps a Persistence implementation with write coalescing.
// Writes are buffered for the given interval and deduplicated by key.
type AsyncPersistence struct {
	inner    Persistence
	interval time.Duration
	mu       sync.Mutex
	pending  map[string]*StoredState
	timer    *time.Timer
	stopCh   chan struct{}
	stopped  bool
}

// NewAsyncPersistence creates a new async persistence wrapper.
// Writes are coalesced over the given interval (recommended: 1s).
func NewAsyncPersistence(inner Persistence, interval time.Duration) *AsyncPersistence {
	if interval <= 0 {
		interval = 1 * time.Second
	}
	return &AsyncPersistence{
		inner:    inner,
		interval: interval,
		pending:  make(map[string]*StoredState),
		stopCh:   make(chan struct{}),
	}
}

// Save buffers a state write. The actual write happens after the coalescing interval.
func (ap *AsyncPersistence) Save(ctx context.Context, key string, state *StoredState) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.stopped {
		return ap.inner.Save(ctx, key, state)
	}

	ap.pending[key] = state

	if ap.timer == nil {
		ap.timer = time.AfterFunc(ap.interval, ap.flush)
	}

	return nil
}

// Load delegates directly to the inner persistence (no caching).
func (ap *AsyncPersistence) Load(ctx context.Context, key string) (*StoredState, error) {
	return ap.inner.Load(ctx, key)
}

// flush writes all pending states to the inner persistence.
func (ap *AsyncPersistence) flush() {
	ap.mu.Lock()
	pending := ap.pending
	ap.pending = make(map[string]*StoredState)
	ap.timer = nil
	ap.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for key, state := range pending {
		if err := ap.inner.Save(ctx, key, state); err != nil {
			// Log but don't fail — the in-memory state is the source of truth
			// for the current instance. The DB is for cross-instance sync.
			_ = err
		}
	}
}

// Stop flushes any remaining writes and stops the async writer.
func (ap *AsyncPersistence) Stop() {
	ap.mu.Lock()
	ap.stopped = true
	if ap.timer != nil {
		ap.timer.Stop()
		ap.timer = nil
	}
	ap.mu.Unlock()

	ap.flush()
}
