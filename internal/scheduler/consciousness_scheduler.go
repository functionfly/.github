package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/functionfly/functionfly/internal/consciousness"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type ConsciousnessSchedulerConfig struct {
	Cron            string
	Enabled          bool
	MaxConcurrent    int
}

func DefaultConsciousnessSchedulerConfig() *ConsciousnessSchedulerConfig {
	return &ConsciousnessSchedulerConfig{
		Cron:         "*/30 * * * *",
		Enabled:       true,
		MaxConcurrent: 10,
	}
}

func LoadConsciousnessSchedulerConfig() *ConsciousnessSchedulerConfig {
	config := DefaultConsciousnessSchedulerConfig()

	if v := os.Getenv("CONSCIOUSNESS_CRON"); v != "" {
		if _, err := cron.ParseStandard(v); err != nil {
			logrus.WithError(err).Warn("Invalid CONSCIOUSNESS_CRON, using default")
			consciousness.SchedulerConfigInvalid.Set(1)
		} else {
			config.Cron = v
			consciousness.SchedulerConfigInvalid.Set(0)
		}
	} else {
		consciousness.SchedulerConfigInvalid.Set(0)
	}

	if v := os.Getenv("CONSCIOUSNESS_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("CONSCIOUSNESS_MAX_CONCURRENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxConcurrent = n
		}
	}

	return config
}

type ConsciousnessScheduler struct {
	cron            *cron.Cron
	engine          *consciousness.Engine
	logger          *logrus.Logger
	config          *ConsciousnessSchedulerConfig
	stopOnce        sync.Once
	cancel          context.CancelFunc
	running         atomic.Bool
	lastRunAt       atomic.Value
	lastRunDuration time.Duration
	lastError       error
	consecutiveFail int
	mu              sync.RWMutex
}

func NewConsciousnessScheduler(db *sql.DB) *ConsciousnessScheduler {
	logger := logrus.StandardLogger()
	engine := consciousness.NewEngine(db, logger)
	return &ConsciousnessScheduler{
		cron:   cron.New(),
		engine: engine,
		logger: logger,
		config: LoadConsciousnessSchedulerConfig(),
	}
}

func (s *ConsciousnessScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Consciousness scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid consciousness cron expression: %w", err)
	}

	if err := s.engine.HealthCheck(ctx); err != nil {
		s.logger.WithError(err).Error("Consciousness engine health check failed")
		return fmt.Errorf("consciousness dependency check failed: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		if s.running.Load() {
			s.logger.Info("Previous consciousness analysis still running, skipping this cycle")
			return
		}
		s.runAnalysis(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add consciousness cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":           s.config.Cron,
		"max_concurrent": s.config.MaxConcurrent,
	}).Info("Consciousness scheduler started")

	return nil
}

func (s *ConsciousnessScheduler) Stop() error {
	s.stopOnce.Do(func() {
		s.running.Store(false)
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Consciousness scheduler stopped")
	})
	return nil
}

func (s *ConsciousnessScheduler) runAnalysis(ctx context.Context) {
	s.running.Store(true)
	defer s.running.Store(false)

	start := time.Now()
	s.logger.Info("Starting consciousness analysis run")

	if err := s.engine.AnalyzeAllTenants(ctx); err != nil {
		s.logger.WithError(err).Error("Consciousness analysis run failed")
		s.lastError = err
		s.consecutiveFail++
		consciousness.SchedulerRunTotal.WithLabelValues("error").Inc()
	} else {
		s.lastError = nil
		s.consecutiveFail = 0
		consciousness.SchedulerRunTotal.WithLabelValues("success").Inc()
	}

	s.lastRunAt.Store(time.Now())
	s.lastRunDuration = time.Since(start)

	s.logger.WithFields(logrus.Fields{
		"duration_ms":        s.lastRunDuration.Milliseconds(),
		"consecutive_failures": s.consecutiveFail,
	}).Info("Consciousness analysis run completed")
}

func (s *ConsciousnessScheduler) RunNow(ctx context.Context) error {
	s.logger.Info("Manually triggering consciousness analysis")
	if s.running.Load() {
		return fmt.Errorf("analysis already in progress")
	}
	s.runAnalysis(ctx)
	return nil
}

func (s *ConsciousnessScheduler) GetStatus() map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	lastRunAt := "never"
	if v := s.lastRunAt.Load(); v != nil {
		if t, ok := v.(time.Time); ok {
			lastRunAt = t.Format(time.RFC3339)
		}
	}

	var lastErr string
	s.mu.RLock()
	if s.lastError != nil {
		lastErr = s.lastError.Error()
	}
	s.mu.RUnlock()

	return map[string]interface{}{
		"enabled":             s.config.Enabled,
		"cron":                s.config.Cron,
		"next_run":            nextRun,
		"last_run_at":         lastRunAt,
		"last_run_duration_ms": s.lastRunDuration.Milliseconds(),
		"last_error":          lastErr,
		"consecutive_failures": s.consecutiveFail,
		"is_running":          s.running.Load(),
	}
}