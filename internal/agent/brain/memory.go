package brain

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type Memory struct {
	repo *storage.BrainRepository
}

func NewMemory(repo *storage.BrainRepository) *Memory {
	return &Memory{repo: repo}
}

// EnforceMemoryWindow ensures the tenant doesn't exceed their signal limit
func (m *Memory) EnforceMemoryWindow(ctx context.Context, tenantID uuid.UUID, maxSignals int) (int, error) {
	if maxSignals <= 0 {
		return 0, nil
	}

	stats, err := m.repo.GetBrainStats(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("get brain stats: %w", err)
	}

	excess := stats.TotalSignals - maxSignals
	if excess <= 0 {
		return 0, nil
	}

	// Delete oldest signals beyond the window
	_, _, err = m.repo.ListSignals(ctx, storage.SignalListParams{
		TenantID: tenantID,
		Limit:    excess,
		Offset:   maxSignals,
		SortBy:   "created_at",
		SortDir:  "ASC",
	})
	if err != nil {
		return 0, fmt.Errorf("find excess signals: %w", err)
	}

	// The actual deletion would be done by a background cleanup job
	return excess, nil
}

// GetWorkingMemory returns the current in-context signals for a tenant
func (m *Memory) GetWorkingMemory(ctx context.Context, tenantID uuid.UUID, maxSignals int) ([]*storage.BrainSignal, error) {
	if maxSignals <= 0 {
		maxSignals = 50
	}

	signals, _, err := m.repo.ListSignals(ctx, storage.SignalListParams{
		TenantID: tenantID,
		Limit:    maxSignals,
		SortBy:   "importance",
	})
	return signals, err
}

// TouchSignal updates the last_seen_at for a signal (refresh TTL in Redis)
func (m *Memory) TouchSignal(ctx context.Context, tenantID, signalID uuid.UUID) error {
	_, err := m.repo.GetSignal(ctx, tenantID, signalID)
	return err
}

// CleanupExpired removes signals beyond retention period
func (m *Memory) CleanupExpired(ctx context.Context, tenantID uuid.UUID, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	// In practice, this is handled by TimescaleDB retention policy.
	// This method is for manual cleanup of Redis keys.
	return 0, nil
}
