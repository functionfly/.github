package browser

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// IsolatedManager manages per-agent isolated browser instances.
type IsolatedManager struct {
	config     Config
	sessionMgr *SessionManager
	instances  map[string]*IsolatedInstance
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// IsolatedInstance represents an isolated browser instance for an agent.
type IsolatedInstance struct {
	ID        uuid.UUID `json:"id"`
	AgentID   string   `json:"agent_id"`
	Port      int      `json:"port"`
	PID       int      `json:"pid,omitempty"`
	Status    string   `json:"status"` // starting, running, stopped
	StartedAt time.Time `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

// NewIsolatedManager creates a new isolated browser manager.
func NewIsolatedManager(config Config, sessionMgr *SessionManager) *IsolatedManager {
	return &IsolatedManager{
		config:     config,
		sessionMgr: sessionMgr,
		instances:  make(map[string]*IsolatedInstance),
		stopCh:     make(chan struct{}),
	}
}

// Acquire acquires an isolated browser for an agent.
func (im *IsolatedManager) Acquire(ctx context.Context, agentID string) (*IsolatedInstance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Check if agent already has an isolated instance
	if instance, ok := im.instances[agentID]; ok {
		if instance.Status == "running" {
			logrus.Debugf("Isolated browser: reusing existing instance for agent %s", agentID)
			return instance, nil
		}
	}

	// Create new isolated instance
	port := im.allocatePort()
	instance := &IsolatedInstance{
		ID:        uuid.New(),
		AgentID:   agentID,
		Port:      port,
		Status:    "starting",
		StartedAt: time.Now().UTC(),
	}

	im.instances[agentID] = instance

	// Create session in Redis
	_, err := im.sessionMgr.CreateSession(ctx, agentID, SessionTypeIsolated, port)
	if err != nil {
		logrus.Errorf("Isolated browser: failed to create session for agent %s: %v", agentID, err)
		// Don't fail the acquisition, just log
	}

	logrus.Debugf("Isolated browser: created new instance for agent %s on port %d", agentID, port)
	return instance, nil
}

// Release releases an isolated browser instance for an agent.
func (im *IsolatedManager) Release(agentID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	instance, ok := im.instances[agentID]
	if !ok {
		return nil
	}

	instance.Status = "stopped"
	now := time.Now().UTC()
	instance.StoppedAt = &now

	// Close the session in Redis
	// Note: We don't delete the instance immediately to allow for state inspection

	logrus.Debugf("Isolated browser: released instance for agent %s", agentID)
	return nil
}

// Get returns an isolated instance for an agent.
func (im *IsolatedManager) Get(agentID string) (*IsolatedInstance, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	instance, ok := im.instances[agentID]
	return instance, ok
}

// List returns all isolated instances.
func (im *IsolatedManager) List() []*IsolatedInstance {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instances := make([]*IsolatedInstance, 0, len(im.instances))
	for _, instance := range im.instances {
		instances = append(instances, instance)
	}
	return instances
}

// allocatePort finds an available port for an isolated instance.
func (im *IsolatedManager) allocatePort() int {
	// For isolated browsers, use ports starting from 19222
	basePort := 19222
	usedPorts := make(map[int]bool)

	im.mu.RLock()
	for _, instance := range im.instances {
		usedPorts[instance.Port] = true
	}
	im.mu.RUnlock()

	for port := basePort; port < basePort+1000; port++ {
		if !usedPorts[port] {
			return port
		}
	}

	// Fallback: return error port (should never happen)
	return 0
}

// CleanupOldInstances removes stopped instances older than the TTL.
func (im *IsolatedManager) CleanupOldInstances(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	cutoff := time.Now().UTC().Add(-im.config.SessionTTL)
	for agentID, instance := range im.instances {
		if instance.Status == "stopped" && instance.StoppedAt != nil && instance.StoppedAt.Before(cutoff) {
			delete(im.instances, agentID)
			logrus.Debugf("Isolated browser: cleaned up old instance for agent %s", agentID)
		}
	}
	return nil
}

// StartCleanupRoutine starts a background cleanup routine.
func (im *IsolatedManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-im.stopCh:
				return
			case <-ticker.C:
				if err := im.CleanupOldInstances(ctx); err != nil {
					logrus.Errorf("Isolated browser: cleanup failed: %v", err)
				}
			}
		}
	}()
}

// Stop stops the cleanup routine.
func (im *IsolatedManager) Stop() {
	close(im.stopCh)
}

// HealthCheck checks the health of all isolated instances.
func (im *IsolatedManager) HealthCheck(ctx context.Context) map[string]string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	health := make(map[string]string)
	for agentID, instance := range im.instances {
		health[agentID] = instance.Status
	}
	return health
}
