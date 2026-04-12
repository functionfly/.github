package execution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExecutionPricingConfig holds pricing configuration from the database
type ExecutionPricingConfig struct {
	ExecutionRatePerCallCents  int64 `json:"execution_rate_per_call_cents"`    // Cost per function execution
	ComputeRatePerMSPerGBCents int64 `json:"compute_rate_per_ms_per_gb_cents"` // Cost per ms-GB of compute
	PlatformFeePercent         int64 `json:"platform_fee_percent"`             // Platform fee percentage
	DataTransferRatePerGBCents int64 `json:"data_transfer_rate_per_gb_cents"`  // Cost per GB of data transfer
	MinimumChargeCents         int64 `json:"minimum_charge_cents"`             // Minimum charge per execution
}

// DefaultExecutionPricing returns default pricing configuration
func DefaultExecutionPricing() *ExecutionPricingConfig {
	return &ExecutionPricingConfig{
		ExecutionRatePerCallCents:  1,  // 1 cent per execution
		ComputeRatePerMSPerGBCents: 1,  // 1 cent per ms-GB
		PlatformFeePercent:         10, // 10% platform fee
		DataTransferRatePerGBCents: 0,  // Free data transfer by default
		MinimumChargeCents:         1,  // 1 cent minimum
	}
}

// loadPricingConfig loads pricing configuration from tenant's subscription tier
func (h *Handler) loadPricingConfig(ctx context.Context, tenantID uuid.UUID) (*ExecutionPricingConfig, error) {
	// Default fallback
	defaultConfig := DefaultExecutionPricing()

	// Get tenant's subscription
	subscription, err := h.BackendRepo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return defaultConfig, nil // Use defaults if no subscription found
	}
	if subscription == nil {
		return defaultConfig, nil
	}

	// Get pricing tier
	tier, err := h.BackendRepo.GetPricingTierByID(subscription.PricingTierID)
	if err != nil {
		return defaultConfig, nil // Use defaults if tier not found
	}
	if tier == nil {
		return defaultConfig, nil
	}

	// Parse pricing config from tier features
	var config ExecutionPricingConfig
	if tier.Features != nil {
		featuresJSON, _ := json.Marshal(tier.Features)
		if err := json.Unmarshal(featuresJSON, &config); err != nil {
			logrus.WithError(err).Warn("Failed to parse pricing config from tier, using defaults")
			return defaultConfig, nil
		}
	}

	// Validate and apply defaults for zero values
	if config.ExecutionRatePerCallCents == 0 {
		config.ExecutionRatePerCallCents = defaultConfig.ExecutionRatePerCallCents
	}
	if config.ComputeRatePerMSPerGBCents == 0 {
		config.ComputeRatePerMSPerGBCents = defaultConfig.ComputeRatePerMSPerGBCents
	}
	if config.PlatformFeePercent == 0 {
		config.PlatformFeePercent = defaultConfig.PlatformFeePercent
	}

	return &config, nil
}

// recordBillingUsageEvent records a usage event for billing purposes
func (h *Handler) recordBillingUsageEvent(fn *storage.RegistryFunction, exec *storage.RegistryFunctionExecution, resourceUsage *ResourceUsage) error {
	// Only record billing events for tenant-owned functions (public functions don't bill the publisher)
	if fn.TenantID == nil {
		return nil // Public/unowned functions don't generate billing events
	}

	// Create usage event for this execution
	usageEvent := &storage.UsageEvent{
		TenantID:  *fn.TenantID,
		EventType: "function_execution",
		Quantity:  1, // One execution
		Metadata: map[string]interface{}{
			"function_id":   fn.ID.String(),
			"function_name": fn.Name,
			"author":        fn.Author,
			"version":       exec.Version,
			"duration_ms":   exec.DurationMs,
			"outcome":       exec.Outcome,
			"cached":        exec.Cached,
			"timestamp":     exec.Timestamp.Format(time.RFC3339),
		},
		Timestamp: time.Now().UTC(),
	}

	if err := h.BackendRepo.RecordUsageEvent(context.Background(), usageEvent); err != nil {
		return err
	}

	// Record compute usage if resource usage is available
	if resourceUsage != nil && resourceUsage.CPUTimeUsedMs > 0 {
		computeEvent := &storage.UsageEvent{
			TenantID:  *fn.TenantID,
			EventType: "compute_time_ms",
			Quantity:  resourceUsage.CPUTimeUsedMs,
			Metadata: map[string]interface{}{
				"function_id":    fn.ID.String(),
				"function_name":  fn.Name,
				"memory_used_mb": resourceUsage.MemoryUsedMB,
				"wall_time_ms":   resourceUsage.WallTimeUsedMs,
				"cpu_time_ms":    resourceUsage.CPUTimeUsedMs,
				"timestamp":      exec.Timestamp.Format(time.RFC3339),
			},
			Timestamp: time.Now().UTC(),
		}

		if err := h.BackendRepo.RecordUsageEvent(context.Background(), computeEvent); err != nil {
			return err
		}
	}

	// Record detailed cost allocation entry for transparency and chargebacks
	if err := h.recordCostAllocationEntry(fn, exec, resourceUsage); err != nil {
		// Log but don't fail - cost allocation is not critical for execution
		// This ensures execution continues even if cost tracking fails
	}

	return nil
}

// recordCostAllocationEntry records a detailed cost allocation entry for transparency
// and internal chargebacks. This enables fine-grained cost breakdown by function, tenant, and period.
func (h *Handler) recordCostAllocationEntry(fn *storage.RegistryFunction, exec *storage.RegistryFunctionExecution, resourceUsage *ResourceUsage) error {
	// Only record for tenant-owned functions
	if fn.TenantID == nil {
		return nil
	}

	// Load pricing configuration from tenant's subscription tier
	ctx := context.Background()
	pricingConfig, err := h.loadPricingConfig(ctx, *fn.TenantID)
	if err != nil {
		// Log but continue with defaults - don't block execution
		logrus.WithError(err).Warn("Failed to load pricing config, using defaults")
		pricingConfig = DefaultExecutionPricing()
	}

	// Calculate costs (in cents) using pricing configuration
	executionCost := pricingConfig.ExecutionRatePerCallCents

	var computeCost int64 = 0
	var memoryGB int64 = 1 // default 1GB if not specified
	if resourceUsage != nil && resourceUsage.MemoryUsedMB > 0 {
		memoryGB = int64(resourceUsage.MemoryUsedMB) / 1024
		if memoryGB == 0 {
			memoryGB = 1
		}
	}

	if resourceUsage != nil && resourceUsage.CPUTimeUsedMs > 0 {
		// Cost = cpu_time_ms * memory_gb * rate
		computeCost = (int64(resourceUsage.CPUTimeUsedMs) * memoryGB * pricingConfig.ComputeRatePerMSPerGBCents) / 1000
	}

	// Platform fee is a percentage of execution + compute costs
	baseCost := executionCost + computeCost
	platformFee := (baseCost * pricingConfig.PlatformFeePercent) / 100

	// Data transfer cost (example: based on input/output size)
	dataTransferCost := int64(0) // Would calculate based on actual data transfer

	totalCost := baseCost + platformFee + dataTransferCost

	// Apply minimum charge if needed
	if totalCost < pricingConfig.MinimumChargeCents {
		totalCost = pricingConfig.MinimumChargeCents
	}

	// Determine period (monthly billing period)
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := now.AddDate(0, 1, 0)
	periodEnd := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	// Get region from metadata if available
	region := "unknown"
	if resourceUsage != nil && resourceUsage.Region != "" {
		region = resourceUsage.Region
	}

	// Build resource usage values
	var durationMs, cpuTimeMs, memoryUsedMB, wallTimeMs int64
	if resourceUsage != nil {
		durationMs = int64(exec.DurationMs)
		cpuTimeMs = int64(resourceUsage.CPUTimeUsedMs)
		memoryUsedMB = int64(resourceUsage.MemoryUsedMB)
		wallTimeMs = int64(resourceUsage.WallTimeUsedMs)
	} else {
		durationMs = int64(exec.DurationMs)
	}

	// Create cost allocation entry
	entry := &storage.CostAllocationEntry{
		TenantID:           *fn.TenantID,
		FunctionID:         fn.ID,
		FunctionName:       fn.Name,
		FunctionAuthor:     fn.Author,
		ExecutionID:        exec.ID,
		ExecutionOutcome:   exec.Outcome,
		Cached:             exec.Cached,
		DurationMs:         durationMs,
		CPUTimeMs:          cpuTimeMs,
		MemoryUsedMB:       memoryUsedMB,
		WallTimeMs:         wallTimeMs,
		ExecutionCostCents: executionCost,
		ComputeCostCents:   computeCost,
		PlatformFeeCents:   platformFee,
		DataTransferCents:  dataTransferCost,
		TotalCostCents:     totalCost,
		Region:             region,
		Timestamp:          exec.Timestamp,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		Tags: map[string]string{
			"function_version": exec.Version,
			"outcome":          exec.Outcome,
		},
		Metadata: map[string]interface{}{
			"execution_source": "registry",
			"pricing_tier":     "loaded", // Indicates pricing was loaded from config
		},
	}

	// Record the cost allocation entry
	if err := h.BackendRepo.RecordCostAllocationEntry(context.Background(), entry); err != nil {
		return err
	}

	return nil
}

// updateFunctionPopularity increments the popularity score for a function
func (h *Handler) updateFunctionPopularity(functionID uuid.UUID) error {
	return h.Repo.IncrementPopularity(functionID)
}
