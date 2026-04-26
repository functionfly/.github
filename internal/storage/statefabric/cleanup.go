package statefabric

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CleanupConfig holds configuration for the state fabric cleanup worker
type CleanupConfig struct {
	// Interval between cleanup runs
	Interval time.Duration
	// Batch size for deletion operations
	BatchSize int
	// Enable verbose logging
	Verbose bool
	// Max age for expired snapshots (default: immediate deletion when expires_at < now)
	MaxAge time.Duration
}

// DefaultCleanupConfig returns the default configuration
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		Interval:  1 * time.Hour, // Run every hour by default
		BatchSize: 1000,
		Verbose:   false,
		MaxAge:    0, // Delete immediately when expires_at < now
	}
}

// CleanupService handles periodic cleanup of expired state fabric snapshots
type CleanupService struct {
	db        *gorm.DB
	r2Backend *R2StorageBackend
	config    CleanupConfig
	logger    *logrus.Logger
	stopCh    chan struct{}
}

// NewCleanupService creates a new state fabric cleanup service
func NewCleanupService(db *gorm.DB, config CleanupConfig) *CleanupService {
	if config.Interval == 0 {
		config = DefaultCleanupConfig()
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	return &CleanupService{
		db:        db,
		config:    config,
		logger:    logrus.WithField("service", "state_fabric_cleanup").Logger,
		stopCh:    make(chan struct{}),
	}
}

// NewCleanupServiceWithR2 creates a new state fabric cleanup service with R2 support
func NewCleanupServiceWithR2(db *gorm.DB, r2Backend *R2StorageBackend, config CleanupConfig) *CleanupService {
	if config.Interval == 0 {
		config = DefaultCleanupConfig()
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	return &CleanupService{
		db:        db,
		r2Backend: r2Backend,
		config:    config,
		logger:    logrus.WithField("service", "state_fabric_cleanup").Logger,
		stopCh:    make(chan struct{}),
	}
}

// StartCleanupRoutine starts the background cleanup process
func (s *CleanupService) StartCleanupRoutine(ctx context.Context) {
	s.logger.WithFields(logrus.Fields{
		"interval":  s.config.Interval,
		"batchSize": s.config.BatchSize,
	}).Info("Starting state fabric TTL cleanup routine")

	// Run initial cleanup
	s.runCleanup()

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("State fabric cleanup routine stopping (context cancelled)")
			return
		case <-s.stopCh:
			s.logger.Info("State fabric cleanup routine stopping (stop signal)")
			return
		case <-ticker.C:
			s.runCleanup()
		}
	}
}

// Stop signals the cleanup routine to stop
func (s *CleanupService) Stop() {
	close(s.stopCh)
}

// runCleanup performs all cleanup operations
func (s *CleanupService) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s.logger.Debug("Running state fabric TTL cleanup")

	// Clean up expired snapshots
	snapshotsDeleted := s.cleanupExpiredSnapshots(ctx)
	s.logger.WithFields(logrus.Fields{
		"expiredSnapshotsDeleted": snapshotsDeleted,
	}).Info("Cleaned up expired state fabric snapshots")
}

// cleanupExpiredSnapshots deletes state fabric snapshots that have expired
func (s *CleanupService) cleanupExpiredSnapshots(ctx context.Context) int64 {
	var totalDeleted int64 = 0
	cutoffTime := time.Now().Add(-s.config.MaxAge)

	for {
		var expiredIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&StateFabricSnapshot{}).
			Where("expires_at IS NOT NULL AND expires_at < ?", cutoffTime).
			Limit(s.config.BatchSize).
			Pluck("id", &expiredIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query expired state fabric snapshots")
			break
		}

		if len(expiredIDs) == 0 {
			break
		}

		// Delete expired snapshots in a transaction
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// First, get R2 object keys for potential cleanup
			var r2Objects []struct {
				ID          uuid.UUID
				R2ObjectKey *string
			}
			if err := tx.Model(&StateFabricSnapshot{}).
				Where("id IN ?", expiredIDs).
				Find(&r2Objects).Error; err != nil {
				return err
			}

			// Delete R2 data for expired snapshots
			for _, obj := range r2Objects {
				if obj.R2ObjectKey != nil && *obj.R2ObjectKey != "" {
					if s.r2Backend != nil {
						if err := s.r2Backend.DeleteSnapshotData(ctx, *obj.R2ObjectKey); err != nil {
							s.logger.WithError(err).WithFields(logrus.Fields{
								"snapshot_id":   obj.ID,
								"r2_object_key": *obj.R2ObjectKey,
							}).Warn("Failed to delete R2 snapshot data")
						} else if s.config.Verbose {
							s.logger.WithFields(logrus.Fields{
								"snapshot_id":   obj.ID,
								"r2_object_key": *obj.R2ObjectKey,
							}).Debug("Deleted R2 snapshot data")
						}
					}
				}
			}

			// Delete the snapshots
			result := tx.Where("id IN ?", expiredIDs).Delete(&StateFabricSnapshot{})
			if result.Error != nil {
				return result.Error
			}

			return nil
		})

		if err != nil {
			s.logger.WithError(err).Error("Failed to delete expired state fabric snapshots")
			break
		}

		deleted := int64(len(expiredIDs))
		totalDeleted += deleted

		if s.config.Verbose {
			s.logger.WithFields(logrus.Fields{
				"batchDeleted": deleted,
				"totalDeleted": totalDeleted,
			}).Debug("Deleted batch of expired state fabric snapshots")
		}

		// Check if we need to continue
		if deleted < int64(s.config.BatchSize) {
			break
		}
	}

	return totalDeleted
}

// CleanupExpiredSnapshotsOnce runs a single cleanup of expired snapshots
// This can be called manually via an admin endpoint
func (s *CleanupService) CleanupExpiredSnapshotsOnce(ctx context.Context) (int64, error) {
	deleted := s.cleanupExpiredSnapshots(ctx)
	return deleted, nil
}

// GetStats returns cleanup statistics for monitoring
func (s *CleanupService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	cutoffTime := time.Now().Add(-s.config.MaxAge)

	// Count expired snapshots waiting for cleanup
	var expiredSnapshotsCount int64
	s.db.WithContext(ctx).Model(&StateFabricSnapshot{}).
		Where("expires_at IS NOT NULL AND expires_at < ?", cutoffTime).
		Count(&expiredSnapshotsCount)
	stats["expiredSnapshotsPending"] = expiredSnapshotsCount

	// Count total snapshots
	var totalSnapshots int64
	s.db.WithContext(ctx).Model(&StateFabricSnapshot{}).Count(&totalSnapshots)
	stats["totalSnapshots"] = totalSnapshots

	// Count snapshots with expiration set
	var snapshotsWithExpiration int64
	s.db.WithContext(ctx).Model(&StateFabricSnapshot{}).
		Where("expires_at IS NOT NULL").
		Count(&snapshotsWithExpiration)
	stats["snapshotsWithExpiration"] = snapshotsWithExpiration

	return stats, nil
}

// RunManualCleanup triggers a manual cleanup run with detailed results
func (s *CleanupService) RunManualCleanup(ctx context.Context) (*CleanupResult, error) {
	result := &CleanupResult{
		StartedAt: time.Now(),
	}

	// Get stats before cleanup
	beforeStats, err := s.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	result.SnapshotsBefore = beforeStats["expiredSnapshotsPending"].(int64)

	// Run cleanup
	deleted := s.cleanupExpiredSnapshots(ctx)
	result.ExpiredSnapshotsDeleted = deleted

	// Get stats after cleanup
	afterStats, err := s.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	result.SnapshotsAfter = afterStats["expiredSnapshotsPending"].(int64)
	result.CompletedAt = time.Now()
	result.DurationMs = result.CompletedAt.Sub(result.StartedAt).Milliseconds()

	return result, nil
}

// CleanupResult contains the results of a manual cleanup run
type CleanupResult struct {
	StartedAt               time.Time `json:"started_at"`
	CompletedAt             time.Time `json:"completed_at"`
	DurationMs              int64     `json:"duration_ms"`
	ExpiredSnapshotsDeleted int64     `json:"expired_snapshots_deleted"`
	SnapshotsBefore         int64     `json:"snapshots_before"`
	SnapshotsAfter          int64     `json:"snapshots_after"`
}
