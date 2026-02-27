package concurrency

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
)

// ConcurrencyPool manages execution slots for a single agent
type ConcurrencyPool struct {
	AgentID       string
	PlanTier      string
	ReservedSlots int   // Guaranteed slots for paid tiers
	BurstCeiling  int   // Max burst (calls/sec); -1 = unlimited
	ActiveCount   int64 // atomic counter of in-flight executions
	mu            sync.Mutex
}

// NewConcurrencyPool creates a pool for an agent based on its plan tier
func NewConcurrencyPool(agentID, planTier string) *ConcurrencyPool {
	return &ConcurrencyPool{
		AgentID:       agentID,
		PlanTier:      planTier,
		ReservedSlots: plans.AgentMaxConcurrency(planTier),
		BurstCeiling:  burstCeilingForPlan(planTier),
	}
}

// Acquire attempts to acquire an execution slot. Returns an error if the pool is full.
func (p *ConcurrencyPool) Acquire(ctx context.Context) error {
	current := atomic.AddInt64(&p.ActiveCount, 1)

	// Check burst ceiling (if set)
	if p.BurstCeiling > 0 && int(current) > p.BurstCeiling {
		atomic.AddInt64(&p.ActiveCount, -1)
		return fmt.Errorf("burst ceiling exceeded for agent %s: %d/%d active executions",
			p.AgentID, current, p.BurstCeiling)
	}

	return nil
}

// Release releases an execution slot
func (p *ConcurrencyPool) Release() {
	atomic.AddInt64(&p.ActiveCount, -1)
}

// ActiveExecutions returns the current number of active executions
func (p *ConcurrencyPool) ActiveExecutions() int64 {
	return atomic.LoadInt64(&p.ActiveCount)
}

// IsAtCapacity returns true if the pool is at or above its reserved slot limit
func (p *ConcurrencyPool) IsAtCapacity() bool {
	if p.ReservedSlots < 0 {
		return false // Unlimited
	}
	return atomic.LoadInt64(&p.ActiveCount) >= int64(p.ReservedSlots)
}

func burstCeilingForPlan(planTier string) int {
	switch planTier {
	case plans.PlanAgentScale:
		return plans.AgentScaleBurstCeiling
	case plans.PlanAgentPro:
		return plans.AgentProBurstCeiling
	case plans.PlanAgentEnterprise:
		return plans.AgentEnterpriseBurstCeiling
	default:
		return plans.AgentStarterBurstCeiling
	}
}

// PoolStats holds statistics for a concurrency pool
type PoolStats struct {
	AgentID       string    `json:"agent_id"`
	PlanTier      string    `json:"plan_tier"`
	ReservedSlots int       `json:"reserved_slots"`
	BurstCeiling  int       `json:"burst_ceiling"`
	ActiveCount   int64     `json:"active_count"`
	Utilization   float64   `json:"utilization_pct"`
	Timestamp     time.Time `json:"timestamp"`
}

// Stats returns current pool statistics
func (p *ConcurrencyPool) Stats() *PoolStats {
	active := atomic.LoadInt64(&p.ActiveCount)
	var utilization float64
	if p.ReservedSlots > 0 {
		utilization = float64(active) / float64(p.ReservedSlots) * 100
	}
	return &PoolStats{
		AgentID:       p.AgentID,
		PlanTier:      p.PlanTier,
		ReservedSlots: p.ReservedSlots,
		BurstCeiling:  p.BurstCeiling,
		ActiveCount:   active,
		Utilization:   utilization,
		Timestamp:     time.Now(),
	}
}
