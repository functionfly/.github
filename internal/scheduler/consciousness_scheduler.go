package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/consciousness"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// ConsciousnessSchedulerConfig holds configuration for the consciousness scheduler.
type ConsciousnessSchedulerConfig struct {
	Cron    string
	Enabled bool
}

// DefaultConsciousnessSchedulerConfig returns default configuration.
func DefaultConsciousnessSchedulerConfig() *ConsciousnessSchedulerConfig {
	return &ConsciousnessSchedulerConfig{
		Cron:    "*/30 * * * *", // Every 30 minutes
		Enabled: true,
	}
}

// LoadConsciousnessSchedulerConfig loads configuration from environment.
func LoadConsciousnessSchedulerConfig() *ConsciousnessSchedulerConfig {
	config := DefaultConsciousnessSchedulerConfig()

	if v := os.Getenv("CONSCIOUSNESS_CRON"); v != "" {
		config.Cron = v
	}
	if v := os.Getenv("CONSCIOUSNESS_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	return config
}

// ConsciousnessScheduler runs the consciousness analysis engine periodically.
type ConsciousnessScheduler struct {
	cron     *cron.Cron
	engine   *consciousness.Engine
	logger   *logrus.Logger
	config   *ConsciousnessSchedulerConfig
	stopOnce sync.Once
	cancel   context.CancelFunc
}

// NewConsciousnessScheduler creates a new consciousness scheduler.
func NewConsciousnessScheduler(db *sql.DB) *ConsciousnessScheduler {
	logger := logrus.StandardLogger()
	return &ConsciousnessScheduler{
		cron:   cron.New(),
		engine: consciousness.NewEngine(db, logger),
		logger: logger,
		config: LoadConsciousnessSchedulerConfig(),
	}
}

// Start begins the consciousness scheduler.
func (s *ConsciousnessScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Consciousness scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid consciousness cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runAnalysis(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add consciousness cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron": s.config.Cron,
	}).Info("Consciousness scheduler started")

	return nil
}

// Stop stops the consciousness scheduler.
func (s *ConsciousnessScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Consciousness scheduler stopped")
	})
	return nil
}

// runAnalysis runs the consciousness analysis for all eligible tenants.
func (s *ConsciousnessScheduler) runAnalysis(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting consciousness analysis run")

	if err := s.engine.AnalyzeAllTenants(ctx); err != nil {
		s.logger.WithError(err).Error("Consciousness analysis run failed")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("Consciousness analysis run completed")
}

// RunNow triggers an immediate analysis run (for admin/manual use).
func (s *ConsciousnessScheduler) RunNow(ctx context.Context) error {
	s.logger.Info("Manually triggering consciousness analysis")
	s.runAnalysis(ctx)
	return nil
}

// GetStatus returns the current scheduler status.
func (s *ConsciousnessScheduler) GetStatus() map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":  s.config.Enabled,
		"cron":     s.config.Cron,
		"next_run": nextRun,
	}
}
