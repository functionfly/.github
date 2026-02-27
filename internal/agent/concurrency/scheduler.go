package concurrency

import (
	"context"
	"fmt"
	"sync"
)

// PriorityScheduler manages concurrency pools for all agents
type PriorityScheduler struct {
	mu         sync.RWMutex
	agentPools map[string]*ConcurrencyPool
}

// NewPriorityScheduler creates a new priority scheduler
func NewPriorityScheduler() *PriorityScheduler {
	return &PriorityScheduler{
		agentPools: make(map[string]*ConcurrencyPool),
	}
}

// GetOrCreatePool returns the concurrency pool for an agent, creating it if needed
func (s *PriorityScheduler) GetOrCreatePool(agentID, planTier string) *ConcurrencyPool {
	s.mu.RLock()
	pool, exists := s.agentPools[agentID]
	s.mu.RUnlock()

	if exists {
		return pool
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists = s.agentPools[agentID]; exists {
		return pool
	}

	pool = NewConcurrencyPool(agentID, planTier)
	s.agentPools[agentID] = pool
	return pool
}

// AcquireSlot acquires an execution slot for an agent
func (s *PriorityScheduler) AcquireSlot(ctx context.Context, agentID, planTier string) (*ConcurrencyPool, error) {
	pool := s.GetOrCreatePool(agentID, planTier)
	if err := pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire execution slot: %w", err)
	}
	return pool, nil
}

// ReleaseSlot releases an execution slot for an agent
func (s *PriorityScheduler) ReleaseSlot(agentID string) {
	s.mu.RLock()
	pool, exists := s.agentPools[agentID]
	s.mu.RUnlock()

	if exists {
		pool.Release()
	}
}

// GetAllStats returns statistics for all active agent pools
func (s *PriorityScheduler) GetAllStats() []*PoolStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]*PoolStats, 0, len(s.agentPools))
	for _, pool := range s.agentPools {
		stats = append(stats, pool.Stats())
	}
	return stats
}

// RemovePool removes a pool for an agent (e.g., when agent is deleted)
func (s *PriorityScheduler) RemovePool(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agentPools, agentID)
}

// TotalActiveExecutions returns the total number of active executions across all agents
func (s *PriorityScheduler) TotalActiveExecutions() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	for _, pool := range s.agentPools {
		total += pool.ActiveExecutions()
	}
	return total
}
