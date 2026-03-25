package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// TrustScoreScheduler manages scheduled trust score recalculations
type TrustScoreScheduler struct {
	cron     *cron.Cron
	repo     *registry.RegistryRepository
	scheduleID cron.EntryID
	mu        sync.RWMutex
	isRunning bool
	nextRun  time.Time
}

// TrustScoreScheduleConfig represents the trust score schedule configuration
type TrustScoreScheduleConfig struct {
	Enabled bool
	Cron    string // e.g., "0 * * * *" for hourly
}

// NewTrustScoreScheduler creates a new trust score scheduler
func NewTrustScoreScheduler(repo *registry.RegistryRepository) *TrustScoreScheduler {
	return &TrustScoreScheduler{
		cron: cron.New(),
		repo: repo,
	}
}

// Start starts the scheduler with the given configuration
func (s *TrustScoreScheduler) Start(ctx context.Context, config TrustScoreScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		logrus.Warn("Trust score scheduler already running")
		return nil
	}

	s.isRunning = true

	if config.Enabled && config.Cron != "" {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executeTrustScoreRecalculation(context.Background())
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to add trust score schedule")
			return err
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Trust score scheduler started with cron: %s", config.Cron)
	} else {
		logrus.Info("Trust score scheduler initialized but scheduling is disabled")
	}

	s.cron.Start()
	return nil
}

// Stop stops the scheduler
func (s *TrustScoreScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping trust score scheduler")
	s.cron.Stop()
	s.isRunning = false

	return nil
}

// UpdateConfig updates the schedule configuration
func (s *TrustScoreScheduler) UpdateConfig(ctx context.Context, config TrustScoreScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing schedule
	if s.scheduleID > 0 {
		s.cron.Remove(s.scheduleID)
		s.scheduleID = 0
	}

	// Add new schedule if enabled
	if config.Enabled && config.Cron != "" {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executeTrustScoreRecalculation(ctx)
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to update trust score schedule")
			return err
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Trust score schedule updated: cron=%s, enabled=%v", config.Cron, config.Enabled)
	}

	return nil
}

// GetStatus returns the current scheduler status
func (s *TrustScoreScheduler) GetStatus() TrustScoreSchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return TrustScoreSchedulerStatus{
		IsRunning: s.isRunning,
		NextRun:   s.nextRun,
		Enabled:   s.scheduleID > 0,
	}
}

// TrustScoreSchedulerStatus represents the current scheduler status
type TrustScoreSchedulerStatus struct {
	IsRunning bool      `json:"is_running"`
	NextRun   time.Time `json:"next_run,omitempty"`
	Enabled   bool      `json:"enabled"`
}

// executeTrustScoreRecalculation runs the trust score recalculation
func (s *TrustScoreScheduler) executeTrustScoreRecalculation(ctx context.Context) {
	logrus.Info("Starting scheduled trust score recalculation")

	// First, aggregate hourly metrics for the past hour
	hour := time.Now().Add(-1 * time.Hour).Truncate(time.Hour)
	if err := s.repo.AggregateHourlyMetrics(hour); err != nil {
		logrus.WithError(err).Error("Failed to aggregate hourly metrics")
	}

	// Then recalculate trust scores for all functions
	job, err := s.repo.RefreshAllTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh trust scores")
		return
	}

	logrus.WithFields(logrus.Fields{
		"job_id":               job.ID,
		"functions_processed": job.FunctionsProcessed,
		"functions_total":     job.FunctionsTotal,
		"status":              job.Status,
	}).Info("Completed scheduled trust score recalculation")

	// Update next run time
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scheduleID > 0 {
		entries := s.cron.Entries()
		for _, entry := range entries {
			if entry.ID == s.scheduleID {
				s.nextRun = entry.Next
				break
			}
		}
	}
}

// TriggerImmediate triggers an immediate trust score recalculation
func (s *TrustScoreScheduler) TriggerImmediate(ctx context.Context) error {
	logrus.Info("Triggering immediate trust score recalculation")
	go s.executeTrustScoreRecalculation(ctx)
	return nil
}
