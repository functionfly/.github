package state

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CleanupConfig holds configuration for the cleanup worker
type CleanupConfig struct {
	// Interval between cleanup runs
	Interval time.Duration
	// Batch size for deletion operations
	BatchSize int
	// Enable verbose logging
	Verbose bool
}

// DefaultCleanupConfig returns the default configuration
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		Interval:  1 * time.Hour,
		BatchSize: 1000,
		Verbose:   false,
	}
}

// CleanupService handles periodic cleanup of expired state entries
type CleanupService struct {
	db     *gorm.DB
	config CleanupConfig
	logger *logrus.Logger
}

// NewCleanupService creates a new state cleanup service
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
		logger: logrus.New(),
	}
}

// StartCleanupRoutine starts the background cleanup process
func (s *CleanupService) StartCleanupRoutine(ctx context.Context) {
	s.logger.WithFields(logrus.Fields{
		"interval":  s.config.Interval,
		"batchSize": s.config.BatchSize,
	}).Info("Starting state cleanup routine")

	// Run initial cleanup
	s.CleanupExpiredEntries()

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping state cleanup routine")
			return
		case <-ticker.C:
			s.CleanupExpiredEntries()
		}
	}
}

// CleanupExpiredEntries removes all expired state entries
func (s *CleanupService) CleanupExpiredEntries() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Clean up expired state values
	valuesDeleted := s.cleanupExpiredStateValues(ctx)
	s.logger.WithFields(logrus.Fields{
		"valuesDeleted": valuesDeleted,
	}).Info("Cleaned up expired state values")

	// Clean up expired states (based on TTLDays)
	statesDeleted := s.cleanupExpiredStates(ctx)
	s.logger.WithFields(logrus.Fields{
		"statesDeleted": statesDeleted,
	}).Info("Cleaned up expired states")

	// Clean up expired agent memories
	memoriesDeleted := s.cleanupExpiredAgentMemories(ctx)
	s.logger.WithFields(logrus.Fields{
		"memoriesDeleted": memoriesDeleted,
	}).Info("Cleaned up expired agent memories")

	// Clean up old events (retention policy)
	eventsDeleted := s.cleanupOldEvents(ctx)
	s.logger.WithFields(logrus.Fields{
		"eventsDeleted": eventsDeleted,
	}).Info("Cleaned up old state events")

	// Clean up old snapshots (retention policy)
	snapshotsDeleted := s.cleanupOldSnapshots(ctx)
	s.logger.WithFields(logrus.Fields{
		"snapshotsDeleted": snapshotsDeleted,
	}).Info("Cleaned up old state snapshots")
}

// cleanupExpiredStateValues deletes state values that have expired
func (s *CleanupService) cleanupExpiredStateValues(ctx context.Context) int64 {
	var totalDeleted int64 = 0

	for {
		var expiredIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&StateValue{}).
			Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
			Limit(s.config.BatchSize).
			Pluck("id", &expiredIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query expired state values")
			break
		}

		if len(expiredIDs) == 0 {
			break
		}

		result := s.db.WithContext(ctx).
			Where("id IN ?", expiredIDs).
			Delete(&StateValue{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete expired state values")
			break
		}

		deleted := result.RowsAffected
		totalDeleted += deleted

		if s.config.Verbose {
			s.logger.WithFields(logrus.Fields{
				"batchDeleted": deleted,
				"totalDeleted": totalDeleted,
			}).Debug("Deleted batch of expired state values")
		}

		// Check if we need to continue
		if deleted < int64(s.config.BatchSize) {
			break
		}
	}

	return totalDeleted
}

// cleanupExpiredStates deletes states that have exceeded their TTL
func (s *CleanupService) cleanupExpiredStates(ctx context.Context) int64 {
	var totalDeleted int64 = 0

	for {
		// Find states where TTL has expired
		// TTL of 0 means no expiration
		var expiredStateIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&State{}).
			Where("ttl_days > 0 AND updated_at < ?", time.Now().AddDate(0, 0, -365)).
			Limit(s.config.BatchSize).
			Pluck("id", &expiredStateIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query expired states")
			break
		}

		if len(expiredStateIDs) == 0 {
			break
		}

		// Delete related data first (cascade)
		// Delete state values
		s.db.WithContext(ctx).Where("state_id IN ?", expiredStateIDs).Delete(&StateValue{})
		// Delete state events
		s.db.WithContext(ctx).Where("state_id IN ?", expiredStateIDs).Delete(&StateEvent{})
		// Delete state snapshots
		s.db.WithContext(ctx).Where("state_id IN ?", expiredStateIDs).Delete(&StateSnapshot{})
		// Delete state permissions
		s.db.WithContext(ctx).Where("state_id IN ?", expiredStateIDs).Delete(&StatePermission{})
		// Delete state triggers
		s.db.WithContext(ctx).Where("source_state_id IN ?", expiredStateIDs).Delete(&StateTrigger{})

		// Delete states
		result := s.db.WithContext(ctx).
			Where("id IN ?", expiredStateIDs).
			Delete(&State{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete expired states")
			break
		}

		deleted := result.RowsAffected
		totalDeleted += deleted

		if s.config.Verbose {
			s.logger.WithFields(logrus.Fields{
				"batchDeleted": deleted,
				"totalDeleted": totalDeleted,
			}).Debug("Deleted batch of expired states")
		}

		if deleted < int64(s.config.BatchSize) {
			break
		}
	}

	return totalDeleted
}

// cleanupExpiredAgentMemories deletes agent memories that have expired
func (s *CleanupService) cleanupExpiredAgentMemories(ctx context.Context) int64 {
	var totalDeleted int64 = 0

	for {
		var expiredIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&AgentMemory{}).
			Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
			Limit(s.config.BatchSize).
			Pluck("id", &expiredIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query expired agent memories")
			break
		}

		if len(expiredIDs) == 0 {
			break
		}

		result := s.db.WithContext(ctx).
			Where("id IN ?", expiredIDs).
			Delete(&AgentMemory{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete expired agent memories")
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

// cleanupOldEvents removes old events based on retention policy
// Events are kept for 90 days by default
func (s *CleanupService) cleanupOldEvents(ctx context.Context) int64 {
	retentionDays := 90 // Default retention

	var totalDeleted int64 = 0
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	for {
		var oldEventIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&StateEvent{}).
			Where("timestamp < ?", cutoffTime).
			Limit(s.config.BatchSize).
			Pluck("id", &oldEventIDs).Error

		if err != nil {
			s.logger.WithError(err).Error("Failed to query old state events")
			break
		}

		if len(oldEventIDs) == 0 {
			break
		}

		result := s.db.WithContext(ctx).
			Where("id IN ?", oldEventIDs).
			Delete(&StateEvent{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete old state events")
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

// cleanupOldSnapshots removes old snapshots based on retention policy
// Keep only the last 30 snapshots per state
func (s *CleanupService) cleanupOldSnapshots(ctx context.Context) int64 {
	var totalDeleted int64 = 0

	// Get states that have too many snapshots
	var stateIDs []uuid.UUID
	err := s.db.WithContext(ctx).
		Model(&StateSnapshot{}).
		Select("DISTINCT state_id").
		Pluck("state_id", &stateIDs).Error

	if err != nil {
		s.logger.WithError(err).Error("Failed to query states with snapshots")
		return 0
	}

	for _, stateID := range stateIDs {
		// Get IDs of snapshots to keep (latest 30)
		var keepIDs []uuid.UUID
		err := s.db.WithContext(ctx).
			Model(&StateSnapshot{}).
			Where("state_id = ?", stateID).
			Order("created_at DESC").
			Limit(30).
			Pluck("id", &keepIDs).Error

		if err != nil {
			continue
		}

		if len(keepIDs) == 0 {
			continue
		}

		// Delete older snapshots
		result := s.db.WithContext(ctx).
			Where("state_id = ? AND id NOT IN ?", stateID, keepIDs).
			Delete(&StateSnapshot{})

		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Failed to delete old snapshots")
			continue
		}

		totalDeleted += result.RowsAffected
	}

	return totalDeleted
}

// GetStats returns cleanup statistics
func (s *CleanupService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count expired values
	var expiredValuesCount int64
	s.db.WithContext(ctx).Model(&StateValue{}).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Count(&expiredValuesCount)
	stats["expiredValues"] = expiredValuesCount

	// Count expired memories
	var expiredMemoriesCount int64
	s.db.WithContext(ctx).Model(&AgentMemory{}).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Count(&expiredMemoriesCount)
	stats["expiredMemories"] = expiredMemoriesCount

	// Count total states
	var totalStates int64
	s.db.WithContext(ctx).Model(&State{}).Count(&totalStates)
	stats["totalStates"] = totalStates

	// Count total events
	var totalEvents int64
	s.db.WithContext(ctx).Model(&StateEvent{}).Count(&totalEvents)
	stats["totalEvents"] = totalEvents

	// Count total snapshots
	var totalSnapshots int64
	s.db.WithContext(ctx).Model(&StateSnapshot{}).Count(&totalSnapshots)
	stats["totalSnapshots"] = totalSnapshots

	return stats, nil
}
