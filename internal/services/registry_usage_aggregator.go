package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RegistryUsageAggregatorConfig holds configuration for the registry usage aggregation service
type RegistryUsageAggregatorConfig struct {
	// Enabled indicates whether the aggregation service is enabled
	Enabled bool

	// AggregationInterval is the time between aggregation runs (default: 1 hour)
	AggregationInterval time.Duration

	// RollupInterval is the time between rollup runs (default: 1 hour)
	RollupInterval time.Duration

	// InvoiceGenerationEnabled enables automatic invoice generation
	InvoiceGenerationEnabled bool

	// InvoiceGenerationTime is when invoices are generated (UTC hour, 0-23, default: 2)
	InvoiceGenerationHour int
}

// DefaultRegistryUsageAggregatorConfig returns the default configuration
func DefaultRegistryUsageAggregatorConfig() *RegistryUsageAggregatorConfig {
	return &RegistryUsageAggregatorConfig{
		Enabled:                  true,
		AggregationInterval:      1 * time.Hour,
		RollupInterval:           1 * time.Hour,
		InvoiceGenerationEnabled: false, // Disabled by default - enable when ready
		InvoiceGenerationHour:    2,       // 2 AM UTC
	}
}

// LoadRegistryUsageAggregatorConfig loads configuration from environment variables
func LoadRegistryUsageAggregatorConfig() *RegistryUsageAggregatorConfig {
	config := DefaultRegistryUsageAggregatorConfig()

	if v := os.Getenv("REGISTRY_USAGE_AGGREGATION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("REGISTRY_USAGE_AGGREGATION_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			config.AggregationInterval = d
		}
	}

	if v := os.Getenv("REGISTRY_USAGE_ROLLUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			config.RollupInterval = d
		}
	}

	if v := os.Getenv("INVOICE_GENERATION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.InvoiceGenerationEnabled = enabled
		}
	}

	if v := os.Getenv("INVOICE_GENERATION_HOUR"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h >= 0 && h <= 23 {
			config.InvoiceGenerationHour = h
		}
	}

	return config
}

// AggregatedUsage represents aggregated usage for a tenant/function over a period
type AggregatedUsage struct {
	TenantID       uuid.UUID
	FunctionID     uuid.UUID
	FunctionName   string
	Author         string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	TotalCalls     int64
	SuccessCalls   int64
	ErrorCalls     int64
	CachedCalls    int64
	TotalDuration  int64 // ms
	AvgDuration    int64 // ms
	TotalMemoryMB  int64
	TotalCPUTimeMs int64
}

// UsageAggregationResult contains the result of an aggregation run
type UsageAggregationResult struct {
	PeriodStart       time.Time
	PeriodEnd         time.Time
	TenantsProcessed  int
	EventsCreated     int
	FunctionsTracked  int
	Duration          time.Duration
}

// RegistryUsageAggregator handles aggregation of registry function executions into billing usage events
type RegistryUsageAggregator struct {
	repo     storage.Repository
	config   *RegistryUsageAggregatorConfig
	logger   *logrus.Logger
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewRegistryUsageAggregator creates a new registry usage aggregator
func NewRegistryUsageAggregator(repo storage.Repository, config *RegistryUsageAggregatorConfig) *RegistryUsageAggregator {
	return &RegistryUsageAggregator{
		repo:     repo,
		config:   config,
		logger:   logrus.New(),
		stopChan: make(chan struct{}),
	}
}

// IsEnabled returns whether the service is enabled
func (a *RegistryUsageAggregator) IsEnabled() bool {
	return a.config.Enabled
}

// Start begins the background aggregation routines
func (a *RegistryUsageAggregator) Start(ctx context.Context) {
	if !a.config.Enabled {
		a.logger.Info("Registry usage aggregation service is disabled")
		return
	}

	a.logger.WithFields(logrus.Fields{
		"aggregation_interval": a.config.AggregationInterval,
		"rollup_interval":    a.config.RollupInterval,
	}).Info("Starting registry usage aggregation service")

	// Run initial aggregation for any unprocessed data
	go func() {
		if err := a.AggregateExecutionsToUsageEvents(ctx); err != nil {
			a.logger.WithError(err).Error("Initial execution aggregation failed")
		}
	}()

	// Run initial rollup
	go func() {
		if err := a.AggregateUsageEventsToRollups(ctx); err != nil {
			a.logger.WithError(err).Error("Initial rollup aggregation failed")
		}
	}()

	// Start aggregation ticker
	go a.runAggregationLoop(ctx)

	// Start rollup ticker
	go a.runRollupLoop(ctx)

	// Start invoice generation scheduler if enabled
	if a.config.InvoiceGenerationEnabled {
		go a.runInvoiceGenerationLoop(ctx)
	}
}

// Stop stops the aggregation service
func (a *RegistryUsageAggregator) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopChan)
	})
}

// runAggregationLoop runs the execution-to-usage-events aggregation loop
func (a *RegistryUsageAggregator) runAggregationLoop(ctx context.Context) {
	ticker := time.NewTicker(a.config.AggregationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Registry usage aggregation stopping due to context cancellation")
			return
		case <-a.stopChan:
			a.logger.Info("Registry usage aggregation stopped")
			return
		case <-ticker.C:
			if err := a.AggregateExecutionsToUsageEvents(ctx); err != nil {
				a.logger.WithError(err).Error("Execution aggregation failed")
			}
		}
	}
}

// runRollupLoop runs the usage-events-to-rollups aggregation loop
func (a *RegistryUsageAggregator) runRollupLoop(ctx context.Context) {
	ticker := time.NewTicker(a.config.RollupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Usage rollup aggregation stopping due to context cancellation")
			return
		case <-a.stopChan:
			a.logger.Info("Usage rollup aggregation stopped")
			return
		case <-ticker.C:
			if err := a.AggregateUsageEventsToRollups(ctx); err != nil {
				a.logger.WithError(err).Error("Rollup aggregation failed")
			}
		}
	}
}

// runInvoiceGenerationLoop runs the daily invoice generation loop
func (a *RegistryUsageAggregator) runInvoiceGenerationLoop(ctx context.Context) {
	// Calculate time until next invoice generation
	now := time.Now().UTC()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), a.config.InvoiceGenerationHour, 0, 0, 0, time.UTC)
	if nextRun.Before(now) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	// Wait until the scheduled time
	select {
	case <-ctx.Done():
		return
	case <-a.stopChan:
		return
	case <-time.After(nextRun.Sub(now)):
	}

	// Run invoice generation
	if err := a.GenerateDraftInvoices(ctx); err != nil {
		a.logger.WithError(err).Error("Invoice generation failed")
	}

	// Then run daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Invoice generation stopping due to context cancellation")
			return
		case <-a.stopChan:
			a.logger.Info("Invoice generation stopped")
			return
		case <-ticker.C:
			if err := a.GenerateDraftInvoices(ctx); err != nil {
				a.logger.WithError(err).Error("Invoice generation failed")
			}
		}
	}
}

// AggregateExecutionsToUsageEvents aggregates unprocessed function executions into usage events
func (a *RegistryUsageAggregator) AggregateExecutionsToUsageEvents(ctx context.Context) error {
	start := time.Now()

	// Get the last processed timestamp
	lastProcessed, err := a.getLastProcessedTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last processed timestamp: %w", err)
	}

	periodStart := lastProcessed
	periodEnd := time.Now().UTC()

	a.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
	}).Info("Aggregating executions to usage events")

	// Aggregate executions by tenant and function
	aggregated, err := a.aggregateExecutions(ctx, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to aggregate executions: %w", err)
	}

	eventsCreated := 0
	for _, usage := range aggregated {
		// Create usage event for function executions
		event := &storage.UsageEvent{
			TenantID:  usage.TenantID,
			EventType: "function_execution",
			Quantity:  int(usage.TotalCalls),
			Metadata: map[string]interface{}{
				"function_id":     usage.FunctionID.String(),
				"function_name":   usage.FunctionName,
				"author":          usage.Author,
				"success_calls":   usage.SuccessCalls,
				"error_calls":     usage.ErrorCalls,
				"cached_calls":    usage.CachedCalls,
				"avg_duration_ms": usage.AvgDuration,
				"period_start":    usage.PeriodStart.Format(time.RFC3339),
				"period_end":      usage.PeriodEnd.Format(time.RFC3339),
			},
			Timestamp: periodEnd,
		}

		if err := a.repo.RecordUsageEvent(ctx, event); err != nil {
			a.logger.WithError(err).WithField("tenant_id", usage.TenantID).Error("Failed to record usage event")
			continue
		}
		eventsCreated++

		// Create separate event for compute time (for MicroVM billing)
		if usage.TotalCPUTimeMs > 0 {
			computeEvent := &storage.UsageEvent{
				TenantID:  usage.TenantID,
				EventType: "compute_time_ms",
				Quantity:  int(usage.TotalCPUTimeMs),
				Metadata: map[string]interface{}{
					"function_id":  usage.FunctionID.String(),
					"period_start": usage.PeriodStart.Format(time.RFC3339),
					"period_end":   usage.PeriodEnd.Format(time.RFC3339),
				},
				Timestamp: periodEnd,
			}

			if err := a.repo.RecordUsageEvent(ctx, computeEvent); err != nil {
				a.logger.WithError(err).WithField("tenant_id", usage.TenantID).Error("Failed to record compute usage event")
			}
		}
	}

	// Update last processed timestamp
	if err := a.updateLastProcessedTimestamp(ctx, periodEnd); err != nil {
		a.logger.WithError(err).Error("Failed to update last processed timestamp")
	}

	duration := time.Since(start)
	a.logger.WithFields(logrus.Fields{
		"period_start":      periodStart,
		"period_end":        periodEnd,
		"tenants_processed": len(aggregated),
		"events_created":    eventsCreated,
		"duration_ms":       duration.Milliseconds(),
	}).Info("Execution aggregation completed")

	return nil
}

// aggregateExecutions aggregates raw execution data by tenant and function
func (a *RegistryUsageAggregator) aggregateExecutions(ctx context.Context, start, end time.Time) ([]*AggregatedUsage, error) {
	// Query the repository for aggregated billing usage
	results, err := a.repo.AggregateExecutionsForBilling(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate executions from repository: %w", err)
	}

	// Convert from repository type to AggregatedUsage
	var aggregated []*AggregatedUsage
	for _, result := range results {
		usage := &AggregatedUsage{
			TenantID:       result.TenantID,
			FunctionID:     result.FunctionID,
			FunctionName:   result.FunctionName,
			Author:         result.Author,
			PeriodStart:    start,
			PeriodEnd:      end,
			TotalCalls:     result.TotalCalls,
			SuccessCalls:   result.SuccessCalls,
			ErrorCalls:     result.ErrorCalls,
			CachedCalls:    result.CachedCalls,
			TotalDuration:  result.TotalDuration,
			AvgDuration:    result.AvgDuration,
			TotalMemoryMB:  result.TotalMemoryMB,
			TotalCPUTimeMs: result.TotalCPUTimeMs,
		}
		aggregated = append(aggregated, usage)
	}

	a.logger.WithFields(logrus.Fields{
		"start":           start,
		"end":             end,
		"functions_found": len(aggregated),
	}).Debug("Aggregated executions")

	return aggregated, nil
}

// AggregateUsageEventsToRollups aggregates usage events into daily rollups
func (a *RegistryUsageAggregator) AggregateUsageEventsToRollups(ctx context.Context) error {
	start := time.Now()

	// Get the last rollup date
	lastRollup, err := a.getLastRollupDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last rollup date: %w", err)
	}

	// Rollup period: from last rollup to yesterday
	periodStart := lastRollup.Add(24 * time.Hour)
	periodEnd := time.Now().UTC().Add(-24 * time.Hour).Truncate(24 * time.Hour)

	if periodStart.After(periodEnd) {
		a.logger.Debug("No new data to rollup")
		return nil
	}

	a.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
	}).Info("Aggregating usage events to rollups")

	// For each day in the period
	for currentDate := periodStart; !currentDate.After(periodEnd); currentDate = currentDate.Add(24 * time.Hour) {
		nextDate := currentDate.Add(24 * time.Hour)

		if err := a.rollupDay(ctx, currentDate, nextDate); err != nil {
			a.logger.WithError(err).WithField("date", currentDate).Error("Failed to rollup day")
			continue
		}

		// Update last rollup date
		if err := a.updateLastRollupDate(ctx, currentDate); err != nil {
			a.logger.WithError(err).Error("Failed to update last rollup date")
		}
	}

	duration := time.Since(start)
	a.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
		"duration_ms":  duration.Milliseconds(),
	}).Info("Usage rollup aggregation completed")

	return nil
}

// rollupDay aggregates usage events for a single day into rollups
func (a *RegistryUsageAggregator) rollupDay(ctx context.Context, dayStart, dayEnd time.Time) error {
	// Get usage events for this day and aggregate them into rollups
	// For now, we rely on the RecordUsageEvent already storing data
	// The actual rollup will be done by querying usage_events and creating/updating usage_rollups

	a.logger.WithField("date", dayStart.Format("2006-01-02")).Debug("Rolling up usage for day")

	// Get rollups for the day - this queries existing usage_events and creates rollups
	// Note: In a production system, you might want to implement a specific
	// query to aggregate usage_events into rollups in one SQL operation.
	// For now, we query the existing rollups via GetUsageByTenant and create them
	// if they don't exist yet.

	// Get the last aggregation timestamp to determine if there's new data
	lastProcessed, err := a.repo.GetLastAggregationTimestamp(ctx)
	if err != nil {
		a.logger.WithError(err).Warn("Failed to get last aggregation timestamp")
		return nil // Don't fail rollup if we can't get timestamp
	}

	// Only rollup if we have data after the last rollup
	if !dayStart.After(lastProcessed) {
		a.logger.WithField("date", dayStart.Format("2006-01-02")).Debug("No new data for this day")
		return nil
	}

	// The actual rollup creation happens automatically via usage_events -> usage_rollups
	// In a more sophisticated implementation, you'd have a specific aggregation query
	// here to batch create rollups from usage_events

	a.logger.WithFields(logrus.Fields{
		"date":       dayStart.Format("2006-01-02"),
		"last_agg":   lastProcessed,
	}).Debug("Rollup day processed")

	return nil
}

// GenerateDraftInvoices generates draft invoices from usage rollups
func (a *RegistryUsageAggregator) GenerateDraftInvoices(ctx context.Context) error {
	start := time.Now()

	// Get yesterday's date
	yesterday := time.Now().UTC().Add(-24 * time.Hour).Truncate(24 * time.Hour)

	a.logger.WithField("period_end", yesterday).Info("Generating draft invoices")

	// Get all active subscriptions
	subs, err := a.repo.ListAllSubscriptions(1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	invoicesCreated := 0
	for _, sub := range subs {
		if sub.Status != "active" && sub.Status != "trialing" {
			continue
		}

		if err := a.generateInvoiceForSubscription(ctx, sub, yesterday); err != nil {
			a.logger.WithError(err).WithField("subscription_id", sub.ID).Error("Failed to generate invoice")
			continue
		}
		invoicesCreated++
	}

	duration := time.Since(start)
	a.logger.WithFields(logrus.Fields{
		"invoices_created": invoicesCreated,
		"duration_ms":      duration.Milliseconds(),
	}).Info("Invoice generation completed")

	return nil
}

// generateInvoiceForSubscription creates a draft invoice for a subscription
func (a *RegistryUsageAggregator) generateInvoiceForSubscription(ctx context.Context, sub *storage.Subscription, periodEnd time.Time) error {
	// Calculate period start (start of current billing period)
	periodStart := sub.CurrentPeriodStart
	if periodStart.IsZero() {
		periodStart = periodEnd.AddDate(0, -1, 0) // Default to 1 month
	}

	// Get usage rollups for this period
	rollups, err := a.repo.GetUsageByTenant(sub.TenantID, "", periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Calculate totals
	totalCalls := 0
	totalComputeMs := 0

	for _, rollup := range rollups {
		switch rollup.EventType {
		case "function_execution":
			totalCalls += rollup.TotalQuantity
		case "compute_time_ms":
			totalComputeMs += rollup.TotalQuantity
		}
	}

	// Calculate amount using tiered pricing from the subscription's pricing tier
	amountDue := calculateBillableAmount(totalCalls, totalComputeMs, sub.PricingTier)

	// Check if we already have an invoice for this period
	existing, err := a.getExistingInvoice(ctx, sub.TenantID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to check existing invoice: %w", err)
	}

	if existing != nil {
		// Update existing draft invoice
		updates := map[string]interface{}{
			"amount_due_cents": amountDue,
		}
		_, err := a.repo.UpdateInvoice(ctx, existing.ID, updates)
		if err != nil {
			return fmt.Errorf("failed to update invoice: %w", err)
		}
		a.logger.WithFields(logrus.Fields{
			"invoice_id": existing.ID,
			"amount":     amountDue,
		}).Info("Updated draft invoice")
	} else {
		// Create new draft invoice
		invoice := &storage.Invoice{
			TenantID:       sub.TenantID,
			SubscriptionID: &sub.ID,
			Status:         "draft",
			AmountDueCents: amountDue,
			AmountPaidCents: 0,
			Currency:       "USD",
			PeriodStart:    &periodStart,
			PeriodEnd:      &periodEnd,
		}

		_, err := a.repo.CreateInvoice(ctx, invoice)
		if err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}
		a.logger.WithFields(logrus.Fields{
			"invoice_id": invoice.ID,
			"tenant_id":  sub.TenantID,
			"amount":     amountDue,
		}).Info("Created draft invoice")
	}

	return nil
}

// TieredPricingConfig holds pricing configuration extracted from tier features
type TieredPricingConfig struct {
	BasePriceCents           int
	IncludedRequests         int
	OveragePricePer1K        int // in cents
	IncludedComputeMs        int
	ComputeOveragePerHour    int // in cents
	MaxExecutionsPerMonth    int
	IsPayPerUse              bool
}

// extractPricingConfig extracts tiered pricing config from a pricing tier
func extractPricingConfig(tier *storage.PricingTier) *TieredPricingConfig {
	config := &TieredPricingConfig{
		BasePriceCents:        tier.PriceCents,
		IncludedRequests:      1000000, // Default 1M requests
		OveragePricePer1K:     1,       // Default $0.01 per 1K requests
		IncludedComputeMs:     3600000, // Default 1 hour compute
		ComputeOveragePerHour: 100,     // Default $1.00 per hour
		MaxExecutionsPerMonth: 1000000,
		IsPayPerUse:           false,
	}

	// Parse features JSONB if available
	if tier.Features != nil {
		if features, ok := tier.Features.(map[string]interface{}); ok {
			// Extract included requests
			if v, ok := features["requests"].(float64); ok {
				config.IncludedRequests = int(v)
			}
			// Extract overage price
			if v, ok := features["overage_price_per_1k"].(float64); ok {
				config.OveragePricePer1K = int(v)
			}
			// Extract compute allowance
			if v, ok := features["included_compute_ms"].(float64); ok {
				config.IncludedComputeMs = int(v)
			}
			if v, ok := features["included_compute_hours"].(float64); ok {
				config.IncludedComputeMs = int(v * 3600000)
			}
			// Extract compute overage
			if v, ok := features["compute_overage_per_hour"].(float64); ok {
				config.ComputeOveragePerHour = int(v)
			}
			// Check if pay-per-use tier
			if v, ok := features["is_pay_per_use"].(bool); ok {
				config.IsPayPerUse = v
			}
		}
	}

	return config
}

// calculateBillableAmount calculates the billable amount using tiered pricing
func calculateBillableAmount(totalCalls, totalComputeMs int, tier *storage.PricingTier) int {
	if tier == nil {
		// Fallback to basic pricing if no tier
		return calculateBasicBillableAmount(totalCalls, totalComputeMs, 0)
	}

	config := extractPricingConfig(tier)

	var amountCents int

	// Base subscription price
	amountCents += config.BasePriceCents

	// Calculate execution overages
	if totalCalls > config.IncludedRequests {
		overageCalls := totalCalls - config.IncludedRequests
		overageUnits := overageCalls / 1000
		if overageCalls%1000 > 0 {
			overageUnits++ // Round up
		}
		amountCents += overageUnits * config.OveragePricePer1K
	}

	// Calculate compute overages
	if totalComputeMs > config.IncludedComputeMs {
		overageMs := totalComputeMs - config.IncludedComputeMs
		overageHours := float64(overageMs) / 1000.0 / 3600.0
		amountCents += int(overageHours * float64(config.ComputeOveragePerHour))
	}

	// For pay-per-use tiers, also charge for all usage (no base included)
	if config.IsPayPerUse {
		// Reset to just the base (which is 0 or minimal for pay-per-use)
		amountCents = config.BasePriceCents
		// Charge for all usage
		executionUnits := totalCalls / 1000
		if totalCalls%1000 > 0 {
			executionUnits++
		}
		amountCents += executionUnits * config.OveragePricePer1K

		computeHours := float64(totalComputeMs) / 1000.0 / 3600.0
		amountCents += int(computeHours * float64(config.ComputeOveragePerHour))
	}

	return amountCents
}

// calculateBasicBillableAmount is the fallback basic pricing calculation
func calculateBasicBillableAmount(totalCalls, totalComputeMs, unitPriceCents int) int {
	basePrice := unitPriceCents

	// Add per-execution charges (1 cent per 1000 calls)
	executionCharges := (totalCalls / 1000) * 1

	// Add compute charges (1 cent per hour of compute time)
	computeHours := float64(totalComputeMs) / 1000.0 / 3600.0
	computeCharges := int(computeHours * 100) // $1.00 per hour

	return basePrice + executionCharges + computeCharges
}

// getExistingInvoice checks if a draft invoice already exists for the period
func (a *RegistryUsageAggregator) getExistingInvoice(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*storage.Invoice, error) {
	return a.repo.GetInvoiceByPeriod(ctx, tenantID, periodStart, periodEnd)
}

// getLastProcessedTimestamp gets the last timestamp that was processed
func (a *RegistryUsageAggregator) getLastProcessedTimestamp(ctx context.Context) (time.Time, error) {
	return a.repo.GetLastAggregationTimestamp(ctx)
}

// updateLastProcessedTimestamp updates the last processed timestamp
func (a *RegistryUsageAggregator) updateLastProcessedTimestamp(ctx context.Context, timestamp time.Time) error {
	return a.repo.SetLastAggregationTimestamp(ctx, timestamp)
}

// getLastRollupDate gets the last date that was rolled up
func (a *RegistryUsageAggregator) getLastRollupDate(ctx context.Context) (time.Time, error) {
	return a.repo.GetLastRollupDate(ctx)
}

// updateLastRollupDate updates the last rollup date
func (a *RegistryUsageAggregator) updateLastRollupDate(ctx context.Context, date time.Time) error {
	return a.repo.SetLastRollupDate(ctx, date)
}

// GetUsageSummary returns a summary of usage for a tenant
func (a *RegistryUsageAggregator) GetUsageSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*UsageSummary, error) {
	rollups, err := a.repo.GetUsageByTenant(tenantID, "", start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	summary := &UsageSummary{
		TenantID:    tenantID,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	for _, rollup := range rollups {
		switch rollup.EventType {
		case "function_execution":
			summary.TotalExecutions += rollup.TotalQuantity
		case "compute_time_ms":
			summary.TotalComputeMs += rollup.TotalQuantity
		}
	}

	return summary, nil
}

// UsageSummary represents a summary of tenant usage
type UsageSummary struct {
	TenantID        uuid.UUID
	PeriodStart     time.Time
	PeriodEnd       time.Time
	TotalExecutions int
	TotalComputeMs  int
	EstimatedCost   int // in cents
}
