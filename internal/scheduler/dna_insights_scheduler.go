package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	dnaStorage "github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// DNAInsightsSchedulerConfig holds configuration for the DNA insights scheduler.
type DNAInsightsSchedulerConfig struct {
	// Cron is the cron expression (default: "0 1 * * *" — 1 AM daily)
	Cron string

	// Enabled controls whether the scheduler is active
	Enabled bool
}

// DefaultDNAInsightsSchedulerConfig returns default configuration.
func DefaultDNAInsightsSchedulerConfig() *DNAInsightsSchedulerConfig {
	return &DNAInsightsSchedulerConfig{
		Cron:    "0 1 * * *",
		Enabled: true,
	}
}

// LoadDNAInsightsSchedulerConfig loads configuration from environment.
func LoadDNAInsightsSchedulerConfig() *DNAInsightsSchedulerConfig {
	config := DefaultDNAInsightsSchedulerConfig()

	if v := os.Getenv("DNA_INSIGHTS_CRON"); v != "" {
		config.Cron = v
	}
	if v := os.Getenv("DNA_INSIGHTS_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	return config
}

// DNAInsightsScheduler runs a daily aggregation job that computes per-tenant
// DNA insights and inserts them into the function_dna_insights table.
type DNAInsightsScheduler struct {
	cron     *cron.Cron
	repo     *dnaStorage.Repository
	logger   *logrus.Logger
	config   *DNAInsightsSchedulerConfig
	stopOnce sync.Once
	cancel   context.CancelFunc
}

// NewDNAInsightsScheduler creates a new DNA insights scheduler.
func NewDNAInsightsScheduler(repo *dnaStorage.Repository) *DNAInsightsScheduler {
	return &DNAInsightsScheduler{
		cron:   cron.New(),
		repo:   repo,
		logger: logrus.StandardLogger(),
		config: LoadDNAInsightsSchedulerConfig(),
	}
}

// Start begins the insights scheduler.
func (s *DNAInsightsScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("DNA insights scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid DNA insights cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runInsightsAggregation(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add insights cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron": s.config.Cron,
	}).Info("DNA insights scheduler started")

	return nil
}

// Stop stops the insights scheduler.
func (s *DNAInsightsScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("DNA insights scheduler stopped")
	})
	return nil
}

// runInsightsAggregation computes and stores insights for each tenant.
// Processes tenants in parallel (up to 5 concurrent) for faster completion.
func (s *DNAInsightsScheduler) runInsightsAggregation(ctx context.Context) {
	ctx, span := tracing.StartSpan(ctx, "dna.insights_aggregation")
	defer tracing.Finish(ctx)

	start := time.Now()
	s.logger.Info("Starting DNA insights aggregation")

	tenants, err := s.repo.GetDistinctTenantIDs(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get tenant list for insights aggregation")
		tracing.RecordError(ctx, err)
		return
	}

	tracing.SetAttribute(ctx, "tenant_count", len(tenants))

	// Aggregate for the previous 24-hour period
	periodEnd := time.Now().Truncate(time.Hour)
	periodStart := periodEnd.Add(-24 * time.Hour)
	since := 24 * time.Hour

	var (
		mu           sync.Mutex
		successCount int
		failedTenants []string
	)

	// Process up to 5 tenants concurrently
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, tenantID := range tenants {
		wg.Add(1)
		go func(tid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			insights, err := s.repo.GetTenantInsights(ctx, tid, since)
			if err != nil {
				s.logger.WithError(err).WithField("tenant_id", tid).Error("Failed to compute insights for tenant")
				mu.Lock()
				failedTenants = append(failedTenants, tid)
				mu.Unlock()
				return
			}

			if err := s.repo.InsertInsight(ctx, insights, tid, periodStart, periodEnd); err != nil {
				s.logger.WithError(err).WithField("tenant_id", tid).Error("Failed to insert insights for tenant")
				mu.Lock()
				failedTenants = append(failedTenants, tid)
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(tenantID)
	}

	wg.Wait()

	logFields := logrus.Fields{
		"tenants_processed": successCount,
		"tenants_total":     len(tenants),
		"tenants_failed":    len(failedTenants),
		"period_start":      periodStart.Format(time.RFC3339),
		"period_end":        periodEnd.Format(time.RFC3339),
		"duration_ms":       time.Since(start).Milliseconds(),
	}
	if len(failedTenants) > 0 {
		logFields["failed_tenant_ids"] = failedTenants
	}
	s.logger.WithFields(logFields).Info("DNA insights aggregation completed")
}

// RunNow triggers an immediate insights aggregation (for admin/manual use).
func (s *DNAInsightsScheduler) RunNow(ctx context.Context) error {
	s.logger.Info("Manually triggering DNA insights aggregation")
	s.runInsightsAggregation(ctx)
	return nil
}

// GetStatus returns the current scheduler status.
func (s *DNAInsightsScheduler) GetStatus() map[string]interface{} {
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
