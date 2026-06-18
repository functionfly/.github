package statefabric

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	defaultCleanupInterval  = 1 * time.Hour
	defaultCleanupBatchSize = 1000
)

func init() {
	if val := os.Getenv("STATEFABRIC_CLEANUP_INTERVAL"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil && parsed > 0 {
			defaultCleanupInterval = parsed
		}
	}
	if val := os.Getenv("STATEFABRIC_CLEANUP_BATCH_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			defaultCleanupBatchSize = parsed
		}
	}
}

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
		Interval:  defaultCleanupInterval,
		BatchSize: defaultCleanupBatchSize,
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
	drainCh   chan struct{}
	isRunning bool
	mu        sync.Mutex
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
		db:     db,
		config: config,
		logger: logrus.WithField("service", "state_fabric_cleanup").Logger,
		stopCh: make(chan struct{}),
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

	s.mu.Lock()
	s.isRunning = true
	s.drainCh = make(chan struct{})
	s.mu.Unlock()

	// Run initial cleanup
	s.runCleanup()

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()
	defer func() {
		s.mu.Lock()
		s.isRunning = false
		close(s.drainCh)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("State fabric cleanup routine stopping (context cancelled)")
			return
		case <-s.stopCh:
			s.logger.Info("State fabric cleanup routine stopping (stop signal)")
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.isRunning {
				s.mu.Unlock()
				continue
			}
			s.mu.Unlock()
			s.runCleanup()
		}
	}
}

// Stop signals the cleanup routine to stop
func (s *CleanupService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = false
	close(s.stopCh)
}

// Drain waits for any in-progress cleanup to complete
func (s *CleanupService) Drain(timeout time.Duration) error {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	select {
	case <-s.drainCh:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("cleanup drain timed out after %s", timeout)
	}
}

// IsRunning returns whether the cleanup service is currently running
func (s *CleanupService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// runCleanup performs all cleanup operations
func (s *CleanupService) runCleanup() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		s.logger.Debug("Skipping cleanup run - previous run still in progress")
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer func() {
		cancel()
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	s.logger.Debug("Running state fabric TTL cleanup")

	fabricsDeleted := s.cleanupExpiredFabrics(ctx)
	s.logger.WithFields(logrus.Fields{
		"expiredFabricsDeleted": fabricsDeleted,
	}).Info("Cleaned up expired state fabrics")

	snapshotsDeleted := s.cleanupExpiredSnapshots(ctx)
	s.logger.WithFields(logrus.Fields{
		"expiredSnapshotsDeleted": snapshotsDeleted,
	}).Info("Cleaned up expired state fabric snapshots")
}

// cleanupExpiredFabrics deletes state fabrics that have exceeded their TTL
func (s *CleanupService) cleanupExpiredFabrics(ctx context.Context) int64 {
	var totalDeleted int64 = 0

	for {
		// Find fabrics where TTL has expired
		// TTL of 0 means no expiration
		// A fabric expires if: NOW() > updated_at + (ttl_days * INTERVAL '1 day')
		var expiredIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&StateFabric{}).
			Where("ttl_days > 0 AND updated_at < NOW() - (ttl_days * INTERVAL '1 day')").
			Limit(s.config.BatchSize).
			Pluck("id", &expiredIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query expired state fabrics")
			break
		}

		if len(expiredIDs) == 0 {
			break
		}

		// Delete related data first (cascade)
		// Delete fabric stores
		s.db.WithContext(ctx).Where("fabric_id IN ?", expiredIDs).Delete(&StateFabricStore{})
		// Delete fabric pipelines
		s.db.WithContext(ctx).Where("fabric_id IN ?", expiredIDs).Delete(&StateFabricPipeline{})
		// Delete fabric replays
		s.db.WithContext(ctx).Where("fabric_id IN ?", expiredIDs).Delete(&StateFabricReplay{})
		// Delete fabric snapshots
		s.db.WithContext(ctx).Where("fabric_id IN ?", expiredIDs).Delete(&StateFabricSnapshot{})

		// Delete fabrics
		result := s.db.WithContext(ctx).
			Where("id IN ?", expiredIDs).
			Delete(&StateFabric{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete expired state fabrics")
			break
		}

		deleted := result.RowsAffected
		totalDeleted += deleted

		if deleted < int64(s.config.BatchSize) {
			break
		}
	}

	return totalDeleted
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
			var r2Objects []struct {
				ID          uuid.UUID
				R2ObjectKey *string
			}
			if err := tx.Model(&StateFabricSnapshot{}).
				Where("id IN ?", expiredIDs).
				Find(&r2Objects).Error; err != nil {
				return err
			}

			// Track snapshots with failed R2 deletion - these should not be deleted from DB
			var failedR2Deletion []uuid.UUID
			for _, obj := range r2Objects {
				if obj.R2ObjectKey != nil && *obj.R2ObjectKey != "" {
					if s.r2Backend != nil {
						if err := s.r2Backend.DeleteSnapshotData(ctx, *obj.R2ObjectKey); err != nil {
							s.logger.WithError(err).WithFields(logrus.Fields{
								"snapshot_id":   obj.ID,
								"r2_object_key": *obj.R2ObjectKey,
							}).Warn("Failed to delete R2 snapshot data - keeping DB record for retry")
							failedR2Deletion = append(failedR2Deletion, obj.ID)
						} else if s.config.Verbose {
							s.logger.WithFields(logrus.Fields{
								"snapshot_id":   obj.ID,
								"r2_object_key": *obj.R2ObjectKey,
							}).Debug("Deleted R2 snapshot data")
						}
					}
				}
			}

			// Build delete query excluding snapshots with failed R2 deletion
			var idsToDelete []uuid.UUID
			if len(failedR2Deletion) > 0 {
				for _, id := range expiredIDs {
					skip := false
					for _, failedID := range failedR2Deletion {
						if id == failedID {
							skip = true
							break
						}
					}
					if !skip {
						idsToDelete = append(idsToDelete, id)
					}
				}
			} else {
				idsToDelete = expiredIDs
			}

			if len(idsToDelete) == 0 {
				s.logger.Info("All snapshots in batch had failed R2 deletion - skipping DB deletion")
				return nil
			}

			result := tx.Where("id IN ?", idsToDelete).Delete(&StateFabricSnapshot{})
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
