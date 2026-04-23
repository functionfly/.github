package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StateUsageAggregatorConfig holds configuration for the state usage aggregation service
type StateUsageAggregatorConfig struct {
	// Enabled indicates whether the service is enabled
	Enabled bool

	// AggregationInterval is the time between aggregation runs
	AggregationInterval time.Duration

	// SyncToBillingEnabled enables syncing aggregated metrics to billing usage events
	SyncToBillingEnabled bool

	// DailyAggregationEnabled enables daily rollup creation
	DailyAggregationEnabled bool

	// MonthlyAggregationEnabled enables monthly rollup creation
	MonthlyAggregationEnabled bool
}

// DefaultStateUsageAggregatorConfig returns the default configuration
func DefaultStateUsageAggregatorConfig() *StateUsageAggregatorConfig {
	return &StateUsageAggregatorConfig{
		Enabled:                   true,
		AggregationInterval:       15 * time.Minute,
		SyncToBillingEnabled:      true,
		DailyAggregationEnabled:   true,
		MonthlyAggregationEnabled: true,
	}
}

// LoadStateUsageAggregatorConfig loads configuration from environment variables
func LoadStateUsageAggregatorConfig() *StateUsageAggregatorConfig {
	config := DefaultStateUsageAggregatorConfig()

	if enabledStr := os.Getenv("STATE_USAGE_AGGREGATOR_ENABLED"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			config.Enabled = enabled
		}
	}

	if intervalStr := os.Getenv("STATE_USAGE_AGGREGATION_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil && interval > 0 {
			config.AggregationInterval = interval
		}
	}

	if syncStr := os.Getenv("STATE_USAGE_SYNC_TO_BILLING"); syncStr != "" {
		if sync, err := strconv.ParseBool(syncStr); err == nil {
			config.SyncToBillingEnabled = sync
		}
	}

	return config
}

// StateTenantUsage holds aggregated state usage for a tenant
type StateTenantUsage struct {
	TenantID          uuid.UUID
	PeriodStart       time.Time
	PeriodEnd         time.Time
	TotalStorageBytes int64
	TotalReadOps      int64
	TotalWriteOps     int64
	ActiveStates      int64
}

// StateUsageAggregator aggregates state_usage_metrics into billing usage events
// and quota tracking data
type StateUsageAggregator struct {
	db       *gorm.DB
	config   *StateUsageAggregatorConfig
	logger   *logrus.Logger
	stopChan chan struct{}
}

// NewStateUsageAggregator creates a new state usage aggregator
func NewStateUsageAggregator(db *gorm.DB, config *StateUsageAggregatorConfig) *StateUsageAggregator {
	if config == nil {
		config = DefaultStateUsageAggregatorConfig()
	}

	return &StateUsageAggregator{
		db:       db,
		config:   config,
		logger:   logrus.New(),
		stopChan: make(chan struct{}),
	}
}

// IsEnabled returns whether the aggregator is enabled
func (a *StateUsageAggregator) IsEnabled() bool {
	return a.config.Enabled
}

// Start begins the background aggregation routine
func (a *StateUsageAggregator) Start(ctx context.Context) {
	if !a.IsEnabled() {
		a.logger.Info("State usage aggregator is disabled")
		return
	}

	a.logger.WithField("interval", a.config.AggregationInterval).Info("Starting state usage aggregator")

	// Run initial aggregation
	if err := a.AggregateAndSync(ctx); err != nil {
		a.logger.WithError(err).Error("Initial state usage aggregation failed")
	}

	// Start background ticker
	go func() {
		ticker := time.NewTicker(a.config.AggregationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				a.logger.Info("State usage aggregator stopping due to context cancellation")
				return
			case <-a.stopChan:
				a.logger.Info("State usage aggregator stopped")
				return
			case <-ticker.C:
				if err := a.AggregateAndSync(ctx); err != nil {
					a.logger.WithError(err).Error("Periodic state usage aggregation failed")
				}
			}
		}
	}()
}

// Stop stops the aggregator
func (a *StateUsageAggregator) Stop() {
	close(a.stopChan)
}

// AggregateAndSync performs aggregation and syncs to billing if enabled
func (a *StateUsageAggregator) AggregateAndSync(ctx context.Context) error {
	start := time.Now()
	a.logger.Info("Starting state usage aggregation")

	// Get current period (today)
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	// Aggregate usage by tenant
	usages, err := a.AggregateTenantUsage(ctx, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to aggregate tenant usage: %w", err)
	}

	a.logger.WithField("tenant_count", len(usages)).Debug("Aggregated state usage for tenants")

	// Sync to billing if enabled
	if a.config.SyncToBillingEnabled {
		if err := a.SyncToBilling(ctx, usages, periodStart, periodEnd); err != nil {
			a.logger.WithError(err).Error("Failed to sync state usage to billing")
			// Continue even if billing sync fails
		}
	}

	duration := time.Since(start)
	a.logger.WithFields(logrus.Fields{
		"duration_ms":  duration.Milliseconds(),
		"tenant_count": len(usages),
	}).Info("State usage aggregation completed")

	return nil
}

// AggregateTenantUsage aggregates state usage metrics by tenant for a period
func (a *StateUsageAggregator) AggregateTenantUsage(ctx context.Context, periodStart, periodEnd time.Time) ([]*StateTenantUsage, error) {
	// Query to get per-tenant aggregates from state_usage_metrics
	// This aggregates raw metrics (read_ops, write_ops, storage) by tenant

	type AggResult struct {
		TenantID     uuid.UUID
		MetricType   string
		TotalValue   int64
		ActiveStates int64
	}

	var results []AggResult
	err := a.db.WithContext(ctx).Raw(`
		SELECT 
			tenant_id,
			metric_type,
			SUM(value) as total_value,
			COUNT(DISTINCT state_id) as active_states
		FROM state_usage_metrics
		WHERE period_start >= ? AND period_end <= ?
		AND state_id IS NOT NULL  -- Per-state metrics, not pre-aggregated
		GROUP BY tenant_id, metric_type
	`, periodStart, periodEnd).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query state usage metrics: %w", err)
	}

	// Group by tenant
	usageMap := make(map[uuid.UUID]*StateTenantUsage)
	for _, r := range results {
		usage, exists := usageMap[r.TenantID]
		if !exists {
			usage = &StateTenantUsage{
				TenantID:    r.TenantID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}
			usageMap[r.TenantID] = usage
		}

		switch r.MetricType {
		case "storage", "daily_storage", "monthly_storage":
			usage.TotalStorageBytes += r.TotalValue
		case "read_ops", "daily_read_ops":
			usage.TotalReadOps += r.TotalValue
		case "write_ops", "daily_write_ops":
			usage.TotalWriteOps += r.TotalValue
		}

		// Track max active states across metric types
		if r.ActiveStates > usage.ActiveStates {
			usage.ActiveStates = r.ActiveStates
		}
	}

	// Convert map to slice
	usages := make([]*StateTenantUsage, 0, len(usageMap))
	for _, usage := range usageMap {
		usages = append(usages, usage)
	}

	return usages, nil
}

// SyncToBilling creates billing usage events from aggregated state usage
func (a *StateUsageAggregator) SyncToBilling(ctx context.Context, usages []*StateTenantUsage, periodStart, periodEnd time.Time) error {
	for _, usage := range usages {
		// Create usage events for each metric type

		// Storage usage event (in MB for billing)
		if usage.TotalStorageBytes > 0 {
			storageEvent := &storage.UsageEvent{
				ID:        uuid.New(),
				TenantID:  usage.TenantID,
				EventType: "state_storage",
				Quantity:  int(usage.TotalStorageBytes / (1024 * 1024)), // Convert to MB
				Metadata: map[string]interface{}{
					"source":        "state_usage_aggregator",
					"period_start":  periodStart.Format(time.RFC3339),
					"period_end":    periodEnd.Format(time.RFC3339),
					"active_states": usage.ActiveStates,
					"unit":          "mb",
				},
				Timestamp: time.Now().UTC(),
			}

			if err := a.recordUsageEvent(ctx, storageEvent); err != nil {
				a.logger.WithError(err).WithField("tenant_id", usage.TenantID).Error("Failed to record storage usage event")
			}
		}

		// Read operations event
		if usage.TotalReadOps > 0 {
			readEvent := &storage.UsageEvent{
				ID:        uuid.New(),
				TenantID:  usage.TenantID,
				EventType: "state_read_ops",
				Quantity:  int(usage.TotalReadOps),
				Metadata: map[string]interface{}{
					"source":       "state_usage_aggregator",
					"period_start": periodStart.Format(time.RFC3339),
					"period_end":   periodEnd.Format(time.RFC3339),
					"unit":         "ops",
				},
				Timestamp: time.Now().UTC(),
			}

			if err := a.recordUsageEvent(ctx, readEvent); err != nil {
				a.logger.WithError(err).WithField("tenant_id", usage.TenantID).Error("Failed to record read ops usage event")
			}
		}

		// Write operations event
		if usage.TotalWriteOps > 0 {
			writeEvent := &storage.UsageEvent{
				ID:        uuid.New(),
				TenantID:  usage.TenantID,
				EventType: "state_write_ops",
				Quantity:  int(usage.TotalWriteOps),
				Metadata: map[string]interface{}{
					"source":       "state_usage_aggregator",
					"period_start": periodStart.Format(time.RFC3339),
					"period_end":   periodEnd.Format(time.RFC3339),
					"unit":         "ops",
				},
				Timestamp: time.Now().UTC(),
			}

			if err := a.recordUsageEvent(ctx, writeEvent); err != nil {
				a.logger.WithError(err).WithField("tenant_id", usage.TenantID).Error("Failed to record write ops usage event")
			}
		}
	}

	return nil
}

// recordUsageEvent records a usage event to the database
func (a *StateUsageAggregator) recordUsageEvent(ctx context.Context, event *storage.UsageEvent) error {
	// Use GORM to create the usage event
	return a.db.WithContext(ctx).Table("usage_events").Create(event).Error
}

// GetTenantStateUsage retrieves current state usage for a specific tenant
func (a *StateUsageAggregator) GetTenantStateUsage(ctx context.Context, tenantID uuid.UUID) (*StateTenantUsage, error) {
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	usage := &StateTenantUsage{
		TenantID:    tenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	// Get total storage
	var storageResult struct {
		Total int64
	}
	err := a.db.WithContext(ctx).Model(&staterepo.StateUsageMetric{}).
		Select("COALESCE(SUM(value), 0) as total").
		Where("tenant_id = ? AND metric_type = ? AND period_start >= ? AND period_end <= ?",
			tenantID, "storage", periodStart, periodEnd).
		Scan(&storageResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage usage: %w", err)
	}
	usage.TotalStorageBytes = storageResult.Total

	// Get read ops
	var readResult struct {
		Total int64
	}
	err = a.db.WithContext(ctx).Model(&staterepo.StateUsageMetric{}).
		Select("COALESCE(SUM(value), 0) as total").
		Where("tenant_id = ? AND metric_type IN (?, ?) AND period_start >= ? AND period_end <= ?",
			tenantID, "read_ops", "daily_read_ops", periodStart, periodEnd).
		Scan(&readResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get read ops: %w", err)
	}
	usage.TotalReadOps = readResult.Total

	// Get write ops
	var writeResult struct {
		Total int64
	}
	err = a.db.WithContext(ctx).Model(&staterepo.StateUsageMetric{}).
		Select("COALESCE(SUM(value), 0) as total").
		Where("tenant_id = ? AND metric_type IN (?, ?) AND period_start >= ? AND period_end <= ?",
			tenantID, "write_ops", "daily_write_ops", periodStart, periodEnd).
		Scan(&writeResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get write ops: %w", err)
	}
	usage.TotalWriteOps = writeResult.Total

	// Get active states count
	var activeStates int64
	err = a.db.WithContext(ctx).Model(&staterepo.StateUsageMetric{}).
		Where("tenant_id = ? AND state_id IS NOT NULL AND period_start >= ? AND period_end <= ?",
			tenantID, periodStart, periodEnd).
		Distinct("state_id").
		Count(&activeStates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active states count: %w", err)
	}
	usage.ActiveStates = activeStates

	return usage, nil
}

// ListTenantStateUsage retrieves state usage for all tenants in a period
func (a *StateUsageAggregator) ListTenantStateUsage(ctx context.Context, start, end time.Time) ([]*StateTenantUsage, error) {
	return a.AggregateTenantUsage(ctx, start, end)
}
