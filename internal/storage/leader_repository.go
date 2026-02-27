package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Leader represents a leader election entry
type Leader struct {
	ID         uuid.UUID
	InstanceID string
	Region     string
	AcquiredAt time.Time
	ExpiresAt  time.Duration
	IsPrimary  bool
}

// LeaderRepository defines leader election operations
type LeaderRepository interface {
	// GetLeader retrieves the current leader
	GetLeader(ctx context.Context) (*Leader, error)

	// ClaimLeadership attempts to claim leadership for this instance
	ClaimLeadership(ctx context.Context, instanceID, region string, ttl time.Duration) error

	// ReleaseLeadership releases leadership for this instance
	ReleaseLeadership(ctx context.Context, instanceID string) error

	// GetAllRegions retrieves all registered regions
	GetAllRegions(ctx context.Context) ([]string, error)

	// RegisterRegion registers a new region
	RegisterRegion(ctx context.Context, region string) error

	// DeregisterRegion removes a region from the pool
	DeregisterRegion(ctx context.Context, region string) error

	// GetRegionHealth returns health status for a region
	GetRegionHealth(ctx context.Context, region string) (*RegionHealth, error)

	// UpdateRegionHealth updates health status for a region
	UpdateRegionHealth(ctx context.Context, region string, healthy bool, latency time.Duration) error
}

// RegionHealth represents health information for a region
type RegionHealth struct {
	Region     string
	Healthy    bool
	LastCheck  time.Time
	Latency    time.Duration
	FailoverAt *time.Time
}
