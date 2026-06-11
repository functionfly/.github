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
	// Edge state version retention - number of versions to keep per key
	EdgeStateVersionRetention int
	// Edge state version max age - delete versions older than this
	EdgeStateVersionMaxAge time.Duration
}

// DefaultCleanupConfig returns the default configuration
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		Interval:                  1 * time.Hour, // Run every hour by default
		BatchSize:                1000,
		Verbose:                  false,
		MaxAge:                   0, // Delete immediately when expires_at < now
		EdgeStateVersionRetention: 10, // Keep last 10 versions of edge state per key
		EdgeStateVersionMaxAge:   30 * 24 * time.Hour, // Delete versions older than 30 days
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

	// Clean up old edge state versions
	if s.config.EdgeStateVersionRetention > 0 || s.config.EdgeStateVersionMaxAge > 0 {
		edgeStateCleaned := s.cleanupEdgeStateVersions(ctx)
		s.logger.WithFields(logrus.Fields{
			"edgeStateVersionsRemoved": edgeStateCleaned,
		}).Info("Cleaned up old edge state versions")
	}
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

		// Delete expired snapshots - handle R2 deletion failures properly
		// Track which snapshots had R2 deletion failures
		var r2DeletionFailures []uuid.UUID

		// First, get R2 object keys and attempt deletion
		var r2Objects []struct {
			ID          uuid.UUID
			R2ObjectKey *string
		}
		if err := s.db.WithContext(ctx).
			Model(&StateFabricSnapshot{}).
			Where("id IN ?", expiredIDs).
			Find(&r2Objects).Error; err != nil {
			s.logger.WithError(err).Error("Failed to query R2 object keys for cleanup")
			break
		}

		// Attempt R2 deletion for each snapshot
		for _, obj := range r2Objects {
			if obj.R2ObjectKey != nil && *obj.R2ObjectKey != "" {
				if s.r2Backend != nil {
					if err := s.r2Backend.DeleteSnapshotData(ctx, *obj.R2ObjectKey); err != nil {
						s.logger.WithError(err).WithFields(logrus.Fields{
							"snapshot_id":   obj.ID,
							"r2_object_key": *obj.R2ObjectKey,
						}).Warn("Failed to delete R2 snapshot data - will retry in next cycle")
						r2DeletionFailures = append(r2DeletionFailures, obj.ID)
					} else if s.config.Verbose {
						s.logger.WithFields(logrus.Fields{
							"snapshot_id":   obj.ID,
							"r2_object_key": *obj.R2ObjectKey,
						}).Debug("Deleted R2 snapshot data")
					}
				}
			}
		}

		// Remove failed deletions from the list to delete
		idsToDelete := make([]uuid.UUID, 0, len(expiredIDs))
		for _, id := range expiredIDs {
			failed := false
			for _, failID := range r2DeletionFailures {
				if id == failID {
					failed = true
					break
				}
			}
			if !failed {
				idsToDelete = append(idsToDelete, id)
			}
		}

		// Delete only snapshots that had successful R2 deletion (or no R2 data)
		if len(idsToDelete) > 0 {
			if err := s.db.WithContext(ctx).
				Where("id IN ?", idsToDelete).
				Delete(&StateFabricSnapshot{}).Error; err != nil {
				s.logger.WithError(err).Error("Failed to delete expired state fabric snapshots from DB")
				break
			}
			totalDeleted += int64(len(idsToDelete))
		}

		// If all R2 deletions failed, stop processing to avoid data inconsistency
		if len(r2DeletionFailures) == len(expiredIDs) && len(expiredIDs) > 0 {
			s.logger.Warn("All R2 snapshot deletions failed - stopping cleanup to prevent orphaned data")
			break
		}

		if s.config.Verbose {
			s.logger.WithFields(logrus.Fields{
				"batchDeleted":  len(idsToDelete),
				"r2Failures":    len(r2DeletionFailures),
				"totalDeleted":  totalDeleted,
			}).Debug("Deleted batch of expired state fabric snapshots")
		}

		// Check if we need to continue
		if int64(len(idsToDelete)) < int64(s.config.BatchSize) {
			break
		}
	}

	return totalDeleted
}

// cleanupEdgeStateVersions cleans up old edge state versions from fabric settings
// This removes versions beyond the retention limit and older than max age
func (s *CleanupService) cleanupEdgeStateVersions(ctx context.Context) int64 {
	if s.r2Backend == nil {
		s.logger.Debug("R2 backend not configured, skipping edge state version cleanup")
		return 0
	}

	var totalDeleted int64 = 0

	// Get all fabrics with edge state
	var fabrics []StateFabric
	if err := s.db.WithContext(ctx).
		Where("settings LIKE ?", "%_edge_state%").
		Find(&fabrics).Error; err != nil {
		s.logger.WithError(err).Error("Failed to query fabrics for edge state cleanup")
		return 0
	}

	for _, fabric := range fabrics {
		deleted := s.cleanupFabricEdgeState(ctx, &fabric)
		totalDeleted += deleted
	}

	return totalDeleted
}

// cleanupFabricEdgeState cleans up edge state versions for a single fabric
func (s *CleanupService) cleanupFabricEdgeState(ctx context.Context, fabric *StateFabric) int64 {
	if fabric.Settings == nil {
		return 0
	}

	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		return 0
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return 0
	}

	var deleted int64
	now := time.Now()
	maxAge := s.config.EdgeStateVersionMaxAge
	retention := s.config.EdgeStateVersionRetention

	for key, value := range stateMap {
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is a versioned entry with history
		versions, hasVersions := entry["_versions"].([]interface{})
		if !hasVersions || len(versions) == 0 {
			continue
		}

		filteredVersions := make([]interface{}, 0)
		cutoffTime := now.Add(-maxAge)

		for i, v := range versions {
			versionEntry, ok := v.(map[string]interface{})
			if !ok {
				continue
			}

			// Check creation time
			if createdAtStr, ok := versionEntry["createdAt"].(string); ok {
				if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
					// Skip if older than max age
					if maxAge > 0 && createdAt.Before(cutoffTime) {
						deleted++
						continue
					}
				}
			}

			// Keep if within retention limit
			if retention <= 0 || i < retention {
				filteredVersions = append(filteredVersions, v)
			} else {
				deleted++
			}
		}

		// Update the entry with filtered versions
		if deleted > 0 {
			if len(filteredVersions) > 0 {
				entry["_versions"] = filteredVersions
			} else {
				delete(entry, "_versions")
			}
			stateMap[key] = entry
		}
	}

	// If we deleted versions, update the fabric settings
	if deleted > 0 {
		fabric.Settings["_edge_state"] = stateMap
		if err := s.db.WithContext(ctx).Model(&StateFabric{}).
			Where("id = ?", fabric.ID).
			Update("settings", fabric.Settings).Error; err != nil {
			s.logger.WithError(err).WithField("fabric_id", fabric.ID).Error("Failed to update fabric settings after edge state cleanup")
		}
	}

	return deleted
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
