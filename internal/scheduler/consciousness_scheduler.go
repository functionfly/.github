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

const (
	DefaultConsciousnessCron    = "*/30 * * * *"
	DefaultConsciousnessTimeout = 10 * time.Minute
)

// ConsciousnessSchedulerConfig holds configuration for the consciousness scheduler.
type ConsciousnessSchedulerConfig struct {
	Cron    string
	Enabled bool
	Timeout time.Duration
}

// DefaultConsciousnessSchedulerConfig returns default configuration.
func DefaultConsciousnessSchedulerConfig() *ConsciousnessSchedulerConfig {
	return &ConsciousnessSchedulerConfig{
		Cron:    DefaultConsciousnessCron,
		Enabled: true,
		Timeout: DefaultConsciousnessTimeout,
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
	if v := os.Getenv("CONSCIOUSNESS_TIMEOUT"); v != "" {
		if timeout, err := time.ParseDuration(v); err == nil {
			config.Timeout = timeout
		}
	}

	return config
}

// ConsciousnessScheduler runs the consciousness analysis engine periodically.
type ConsciousnessScheduler struct {
	cron       *cron.Cron
	engine     *consciousness.Engine
	logger     *logrus.Logger
	config     *ConsciousnessSchedulerConfig
	stopOnce   sync.Once
	stopCh     chan struct{}
	stoppedCh  chan struct{}
	runningCh  chan struct{}
	runningMu  sync.Mutex
}

// NewConsciousnessScheduler creates a new consciousness scheduler.
func NewConsciousnessScheduler(db *sql.DB) *ConsciousnessScheduler {
	logger := logrus.StandardLogger()
	return &ConsciousnessScheduler{
		cron:   cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger))),
		engine: consciousness.NewEngine(db, logger),
		logger: logger,
		config: LoadConsciousnessSchedulerConfig(),
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
		runningCh: make(chan struct{}, 1),
	}
}

// NewConsciousnessSchedulerWithConfig creates a new consciousness scheduler with custom config.
func NewConsciousnessSchedulerWithConfig(db *sql.DB, config *ConsciousnessSchedulerConfig) *ConsciousnessScheduler {
	logger := logrus.StandardLogger()
	return &ConsciousnessScheduler{
		cron:   cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger))),
		engine: consciousness.NewEngine(db, logger),
		logger: logger,
		config: config,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
		runningCh: make(chan struct{}, 1),
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

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runAnalysis(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add consciousness cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":    s.config.Cron,
		"timeout": s.config.Timeout,
	}).Info("Consciousness scheduler started")

	return nil
}

// Stop stops the consciousness scheduler gracefully.
func (s *ConsciousnessScheduler) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		close(s.stopCh)

		s.runningMu.Lock()
		if s.runningCh != nil {
			select {
			case <-s.runningCh:
			default:
			}
		}
		s.runningMu.Unlock()

		ctx := s.cron.Stop()
		<-ctx.Done()
		s.logger.Info("Consciousness scheduler stopped")
		err = nil
	})
	return err
}

// IsRunning returns true if an analysis is currently in progress.
func (s *ConsciousnessScheduler) IsRunning() bool {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	select {
	case <-s.runningCh:
		return false
	default:
		return true
	}
}

// runAnalysis runs the consciousness analysis for all eligible tenants.
func (s *ConsciousnessScheduler) runAnalysis(ctx context.Context) {
	select {
	case s.runningCh <- struct{}{}:
	default:
	}

	defer func() {
		s.runningMu.Lock()
		select {
		case <-s.runningCh:
		default:
		}
		s.runningMu.Unlock()
	}()

	select {
	case <-s.stopCh:
		s.logger.Info("Skipping consciousness analysis - scheduler is stopping")
		return
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Info("Skipping consciousness analysis - context cancelled")
		return
	default:
	}

	start := time.Now()
	s.logger.Info("Starting consciousness analysis run")

	runCtx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	if err := s.engine.AnalyzeAllTenants(runCtx); err != nil {
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

	select {
	case s.runningCh <- struct{}{}:
	default:
	}

	go func() {
		defer func() {
			s.runningMu.Lock()
			select {
			case <-s.runningCh:
			default:
			}
			s.runningMu.Unlock()
		}()

		runCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
		s.runAnalysis(runCtx)
	}()

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
		"enabled":     s.config.Enabled,
		"cron":        s.config.Cron,
		"timeout":     s.config.Timeout.String(),
		"next_run":    nextRun,
		"is_running":  s.IsRunning(),
	}
}

// ConsciousnessRetentionConfig holds configuration for consciousness data retention.
type ConsciousnessRetentionConfig struct {
	Enabled          bool
	Cron             string // Cron for retention cleanup
	InsightsDays     int    // Days to retain insights (default 90)
	DeliveryLogsDays int    // Days to retain delivery logs (default 30)
}

// DefaultConsciousnessRetentionConfig returns default retention configuration.
func DefaultConsciousnessRetentionConfig() *ConsciousnessRetentionConfig {
	return &ConsciousnessRetentionConfig{
		Enabled:          true,
		Cron:             "0 3 * * *", // Daily at 3 AM
		InsightsDays:     90,
		DeliveryLogsDays: 30,
	}
}

// LoadConsciousnessRetentionConfig loads retention configuration from environment.
func LoadConsciousnessRetentionConfig() *ConsciousnessRetentionConfig {
	config := DefaultConsciousnessRetentionConfig()

	if v := os.Getenv("CONSCIOUSNESS_RETENTION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_RETENTION_CRON"); v != "" {
		config.Cron = v
	}
	if v := os.Getenv("CONSCIOUSNESS_INSIGHTS_RETENTION_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			config.InsightsDays = days
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_DELIVERY_LOGS_RETENTION_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			config.DeliveryLogsDays = days
		}
	}

	return config
}

// CleanupScheduler handles periodic cleanup of old consciousness data.
type CleanupScheduler struct {
	cron     *cron.Cron
	repo     *consciousness.Repository
	logger   *logrus.Logger
	config   *ConsciousnessRetentionConfig
	stopOnce sync.Once
	cancel   context.CancelFunc
}

// NewCleanupScheduler creates a new consciousness cleanup scheduler.
func NewCleanupScheduler(db *sql.DB) *CleanupScheduler {
	logger := logrus.StandardLogger()
	repo := consciousness.NewRepository(db, logger)
	return &CleanupScheduler{
		cron:   cron.New(),
		repo:   repo,
		logger: logger,
		config: LoadConsciousnessRetentionConfig(),
	}
}

// Start begins the cleanup scheduler.
func (s *CleanupScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Consciousness retention cleanup is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid retention cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runCleanup(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add cleanup cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":                s.config.Cron,
		"insights_days":       s.config.InsightsDays,
		"delivery_logs_days":  s.config.DeliveryLogsDays,
	}).Info("Consciousness retention cleanup scheduler started")

	return nil
}

// Stop stops the cleanup scheduler.
func (s *CleanupScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Consciousness retention cleanup scheduler stopped")
	})
	return nil
}

// runCleanup runs the data retention cleanup.
func (s *CleanupScheduler) runCleanup(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting consciousness data retention cleanup")

	// First expire old insights
	expired, err := s.repo.ExpireOldInsights(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to expire old insights")
	} else if expired > 0 {
		s.logger.WithField("expired_count", expired).Info("Expired old insights")
	}

	// Then physically delete insights past retention
	deletedInsights, err := s.repo.DeleteInsightsOlderThan(ctx, s.config.InsightsDays)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to delete old insights")
	} else if deletedInsights > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_count": deletedInsights,
			"retention_days": s.config.InsightsDays,
		}).Info("Deleted old insights past retention period")
	}

	// Cleanup old delivery logs
	deletedLogs, err := s.repo.CleanupOldDeliveryLogs(ctx, s.config.DeliveryLogsDays)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to cleanup old delivery logs")
	} else if deletedLogs > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_count": deletedLogs,
			"retention_days": s.config.DeliveryLogsDays,
		}).Info("Deleted old delivery logs past retention period")
	}

	s.logger.WithFields(logrus.Fields{
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("Consciousness data retention cleanup completed")
}

// RunNow triggers an immediate cleanup run (for admin/manual use).
func (s *CleanupScheduler) RunNow(ctx context.Context) error {
	s.logger.Info("Manually triggering consciousness retention cleanup")
	s.runCleanup(ctx)
	return nil
}

// GetStatus returns the current cleanup scheduler status.
func (s *CleanupScheduler) GetStatus() map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":           s.config.Enabled,
		"cron":              s.config.Cron,
		"next_run":          nextRun,
		"insights_days":     s.config.InsightsDays,
		"delivery_logs_days": s.config.DeliveryLogsDays,
	}
}
