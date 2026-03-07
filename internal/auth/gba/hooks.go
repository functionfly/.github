package gba

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// HookRequest represents a request context for hooks
type HookRequest struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	Email     string
	IPAddress string
	UserAgent string
	Host      string
	Headers   map[string]string
	Metadata  map[string]interface{}
}

// HookFunc is a function that can be registered as a hook
type HookFunc func(ctx context.Context, req *HookRequest) error

// HookManager manages authentication hooks
type HookManager struct {
	hooks map[string][]HookFunc
	mu    sync.RWMutex
}

// NewHookManager creates a new hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[string][]HookFunc),
	}
}

// Register registers a hook for a specific event
// Events: before:signup, after:signup, before:signin, after:signin, before:signout, etc.
func (hm *HookManager) Register(event string, fn HookFunc) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.hooks[event] == nil {
		hm.hooks[event] = make([]HookFunc, 0)
	}
	hm.hooks[event] = append(hm.hooks[event], fn)
}

// Execute runs all hooks for a specific event
// Returns on first error or after all hooks complete successfully
func (hm *HookManager) Execute(ctx context.Context, event string, req *HookRequest) error {
	hm.mu.RLock()
	hooks := hm.hooks[event]
	hm.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

// HasHooks returns true if there are hooks registered for an event
func (hm *HookManager) HasHooks(event string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.hooks[event]) > 0
}

// Clear removes all hooks for an event
func (hm *HookManager) Clear(event string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.hooks, event)
}

// ClearAll removes all hooks
func (hm *HookManager) ClearAll() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hooks = make(map[string][]HookFunc)
}
