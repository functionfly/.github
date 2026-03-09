package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// FactoryPipelineScheduler manages scheduled factory pipeline executions
type FactoryPipelineScheduler struct {
	cron       *cron.Cron
	factory    factoryRunner
	mu         sync.RWMutex
	scheduleID cron.EntryID
	config     *FactoryScheduleConfig
	isRunning  bool
	lastRun    time.Time
	nextRun    time.Time
}

// factoryRunner defines the interface for running the factory pipeline
type factoryRunner interface {
	Run(ctx context.Context) (*factory.FactoryRun, error)
}

// FactoryScheduleConfig represents the factory schedule configuration
type FactoryScheduleConfig struct {
	Enabled  bool   `json:"enabled"`
	Cron     string `json:"cron"`
	Timezone string `json:"timezone"`
}

// NewFactoryPipelineScheduler creates a new factory pipeline scheduler
func NewFactoryPipelineScheduler(factorySvc factoryRunner) *FactoryPipelineScheduler {
	return &FactoryPipelineScheduler{
		cron:    cron.New(),
		factory: factorySvc,
		config: &FactoryScheduleConfig{
			Enabled:  false,
			Cron:     "0 0 * * *", // Daily at midnight UTC
			Timezone: "UTC",
		},
	}
}

// Start starts the scheduler with the given configuration
func (s *FactoryPipelineScheduler) Start(ctx context.Context, config FactoryScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		logrus.Warn("Factory pipeline scheduler already running")
		return nil
	}

	s.config = &config

	// Validate cron expression
	if _, err := cron.ParseStandard(config.Cron); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Only add the job if scheduling is enabled
	if config.Enabled {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executePipeline(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Factory pipeline scheduler started with cron: %s", config.Cron)
	} else {
		logrus.Info("Factory pipeline scheduler initialized but scheduling is disabled")
	}

	s.cron.Start()
	s.isRunning = true

	return nil
}

// Stop stops the scheduler
func (s *FactoryPipelineScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping factory pipeline scheduler")
	<-s.cron.Stop().Done()
	s.isRunning = false

	return nil
}

// UpdateConfig updates the schedule configuration
func (s *FactoryPipelineScheduler) UpdateConfig(ctx context.Context, config FactoryScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate cron expression
	if _, err := cron.ParseStandard(config.Cron); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Remove existing schedule
	if s.scheduleID > 0 {
		s.cron.Remove(s.scheduleID)
	}

	s.config = &config

	// Add new schedule if enabled
	if config.Enabled {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executePipeline(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Factory pipeline schedule updated: cron=%s, enabled=%v", config.Cron, config.Enabled)
	} else {
		s.scheduleID = 0
		logrus.Info("Factory pipeline scheduling disabled")
	}

	return nil
}

// GetStatus returns the current scheduler status
func (s *FactoryPipelineScheduler) GetStatus() FactoryScheduleStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return FactoryScheduleStatus{
		IsEnabled: s.config.Enabled,
		Cron:      s.config.Cron,
		Timezone:  s.config.Timezone,
		IsRunning: s.isRunning,
		LastRun:   s.lastRun,
		NextRun:   s.nextRun,
	}
}

// FactoryScheduleStatus represents the current scheduler status
type FactoryScheduleStatus struct {
	IsEnabled bool      `json:"is_enabled"`
	Cron      string    `json:"cron"`
	Timezone  string    `json:"timezone"`
	IsRunning bool      `json:"is_running"`
	LastRun   time.Time `json:"last_run"`
	NextRun   time.Time `json:"next_run"`
}

// executePipeline runs the factory pipeline
func (s *FactoryPipelineScheduler) executePipeline(ctx context.Context) {
	logrus.Info("Starting scheduled factory pipeline run")

	run, err := s.factory.Run(ctx)
	if err != nil {
		logrus.WithError(err).Error("Scheduled factory pipeline run failed")
		return
	}

	s.mu.Lock()
	s.lastRun = time.Now()
	// Update next run time
	if s.config.Enabled {
		schedule, err := cron.ParseStandard(s.config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}
	}
	s.mu.Unlock()

	logrus.Infof("Scheduled factory pipeline run completed: status=%s, opportunities_scanned=%d, functions_generated=%d, functions_published=%d",
		run.Status, run.OpportunitiesScanned, run.FunctionsGenerated, run.FunctionsPublished)
}
