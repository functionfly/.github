package vault

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// DynamicLeaseSweeperConfig configures the dynamic-lease background
// worker. It runs alongside the secret expiration sweeper.
type DynamicLeaseSweeperConfig struct {
	Interval  time.Duration
	BatchSize int
	Logger    *logrus.Logger
}

// DynamicLeaseSweeper drops expired DB users in the background.
// One DB connection per expired lease is opened, the user is
// dropped, and the lease is marked revoked.
type DynamicLeaseSweeper struct {
	repo    *Repository
	service *DynamicSecretService
	cfg     DynamicLeaseSweeperConfig
	logger  *logrus.Logger
}

// NewDynamicLeaseSweeper constructs a sweeper.
func NewDynamicLeaseSweeper(repo *Repository, cfg DynamicLeaseSweeperConfig) *DynamicLeaseSweeper {
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Minute
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return &DynamicLeaseSweeper{
		repo:    repo,
		service: NewDynamicSecretService(repo),
		cfg:     cfg,
		logger:  logger,
	}
}

// Start runs the sweeper until ctx is cancelled.
func (s *DynamicLeaseSweeper) Start(ctx context.Context) {
	if s.cfg.Interval <= 0 {
		s.logger.Info("Dynamic lease sweeper disabled (interval = 0)")
		return
	}
	s.logger.WithField("interval", s.cfg.Interval.String()).Info("Starting dynamic lease sweeper")

	s.runOnce(ctx)
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping dynamic lease sweeper")
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

// RunOnce runs a single sweep cycle.
func (s *DynamicLeaseSweeper) RunOnce(ctx context.Context) DynamicLeaseSweepCounts {
	return s.runOnce(ctx)
}

// DynamicLeaseSweepCounts reports the work done in a single sweep.
type DynamicLeaseSweepCounts struct {
	LeasesScanned int
	UsersDropped  int
	Errors        int
}

func (s *DynamicLeaseSweeper) runOnce(ctx context.Context) DynamicLeaseSweepCounts {
	var counts DynamicLeaseSweepCounts
	leases, err := s.repo.ListExpiredLeases(ctx, s.cfg.BatchSize)
	if err != nil {
		s.logger.WithError(err).Error("Dynamic lease sweep failed")
		return counts
	}
	counts.LeasesScanned = len(leases)
	for i := range leases {
		lease := leases[i]
		target, err := s.repo.GetTarget(ctx, lease.TargetID, lease.TenantID)
		if err != nil || target == nil {
			s.logger.WithError(err).WithField("target_id", lease.TargetID).Warn("Target lookup failed during sweep")
			counts.Errors++
			continue
		}
		if err := s.service.Revoke(ctx, &lease, target, "lease_expired"); err != nil {
			s.logger.WithError(err).WithField("lease_id", lease.LeaseID).Warn("Failed to drop user during sweep")
			counts.Errors++
			continue
		}
		counts.UsersDropped++
	}
	if counts.LeasesScanned > 0 {
		s.logger.WithFields(logrus.Fields{
			"scanned":       counts.LeasesScanned,
			"users_dropped": counts.UsersDropped,
			"errors":        counts.Errors,
		}).Info("Dynamic lease sweep cycle complete")
	}
	return counts
}
