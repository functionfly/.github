package browser

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Pool manages a shared pool of browser instances.
type Pool struct {
	config     Config
	redis      *redis.Client
	sessionMgr *SessionManager
	mu         sync.Mutex
	available  []int
	allocated  map[string]int // agentID -> port
	stopCh     chan struct{}
}

// NewPool creates a new browser pool manager.
func NewPool(config Config, redisClient *redis.Client, sessionMgr *SessionManager) *Pool {
	pool := &Pool{
		config:     config,
		redis:      redisClient,
		sessionMgr: sessionMgr,
		available:  make([]int, 0, config.PoolSize),
		allocated:  make(map[string]int),
		stopCh:     make(chan struct{}),
	}

	// Initialize available ports
	for i := 0; i < config.PoolSize; i++ {
		pool.available = append(pool.available, 9222+i) // Default CDP ports start at 9222
	}

	return pool
}

// Acquire attempts to acquire a browser for an agent.
// Returns the port number for the acquired browser.
func (p *Pool) Acquire(ctx context.Context, agentID string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if agent already has a browser allocated (sticky session)
	if port, ok := p.allocated[agentID]; ok {
		logrus.Debugf("Browser pool: reusing allocated browser for agent %s on port %d", agentID, port)
		return port, nil
	}

	// Try to get from available pool
	if len(p.available) == 0 {
		return 0, fmt.Errorf("browser pool exhausted")
	}

	// Random selection from available browsers for load balancing
	idx := rand.Intn(len(p.available))
	port := p.available[idx]
	p.available = append(p.available[:idx], p.available[idx+1:]...)
	p.allocated[agentID] = port

	logrus.Debugf("Browser pool: allocated browser for agent %s on port %d", agentID, port)
	return port, nil
}

// Release returns a browser to the pool.
func (p *Pool) Release(agentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	port, ok := p.allocated[agentID]
	if !ok {
		logrus.Debugf("Browser pool: no browser to release for agent %s", agentID)
		return
	}

	delete(p.allocated, agentID)
	p.available = append(p.available, port)

	logrus.Debugf("Browser pool: released browser on port %d back to pool", port)
}

// ReleasePort releases a specific port back to the pool.
func (p *Pool) ReleasePort(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.available = append(p.available, port)
}

// IsAvailable checks if there are any available browsers in the pool.
func (p *Pool) IsAvailable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available) > 0
}

// AvailableCount returns the number of available browsers.
func (p *Pool) AvailableCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}

// AllocatedCount returns the number of allocated browsers.
func (p *Pool) AllocatedCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.allocated)
}

// GetAllocatedPort returns the port allocated to an agent, or 0 if none.
func (p *Pool) GetAllocatedPort(agentID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.allocated[agentID]
}

// StartPoolHealthCheck starts a background goroutine that monitors pool health.
func (p *Pool) StartPoolHealthCheck(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopCh:
				return
			case <-ticker.C:
				p.checkPoolHealth(ctx)
			}
		}
	}()
}

// checkPoolHealth monitors the pool and logs warnings.
func (p *Pool) checkPoolHealth(ctx context.Context) {
	available := p.AvailableCount()
	allocated := p.AllocatedCount()
	total := available + allocated

	// Log warning if pool is getting low
	if available <= 2 {
		logrus.Warnf("Browser pool: only %d/%d browsers available", available, total)
	}

	// Update Redis metrics
	if p.redis != nil {
		p.redis.Set(ctx, "browser:pool:available", available, 0)
		p.redis.Set(ctx, "browser:pool:allocated", allocated, 0)
	}
}

// Stop stops the pool health check.
func (p *Pool) Stop() {
	close(p.stopCh)
}

// BrowserInfo contains information about a browser instance.
type BrowserInfo struct {
	Port       int    `json:"port"`
	AgentID    string `json:"agent_id,omitempty"`
	URL        string `json:"url,omitempty"`
	Isolate    bool   `json:"is_isolated"`
	PID        int    `json:"pid,omitempty"`
	StartedAt  time.Time `json:"started_at,omitempty"`
}
