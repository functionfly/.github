package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RealtimeUsageTrackerInterface defines the interface for real-time usage tracking
// This allows for easy mocking in tests


type RealtimeUsageTrackerInterface interface {
	RecordExecution(ctx context.Context, tenantID uuid.UUID, executionID string) (*QuotaCheckResult, error)
	RecordComputeUsage(ctx context.Context, tenantID uuid.UUID, cpuTimeMs int) error
	GetQuotaStatus(ctx context.Context, tenantID uuid.UUID) (*RealtimeQuotaStatus, error)
	CheckQuota(ctx context.Context, tenantID uuid.UUID) (*QuotaCheckResult, error)
	IsEnabled() bool
}

// RealtimeUsageTracker provides real-time usage tracking and quota enforcement
// using Redis for immediate counters with periodic sync to Postgres
//
// This addresses the gap: "Execution data recorded asynchronously after completion
// No real-time usage feedback for users approaching limits"
type RealtimeUsageTracker struct {
	redisClient *redis.Client
	repo        storage.Repository
	notifySvc   *notification.Service
	logger      *logrus.Logger

	// Configuration
	enabled                    bool
	counterTTL                 time.Duration
	syncInterval               time.Duration
	warningThresholdPercent    float64
	criticalThresholdPercent   float64

	// Background sync
	stopChan chan struct{}
	stopOnce sync.Once
}

// RealtimeUsageConfig holds configuration for the realtime usage tracker

type RealtimeUsageConfig struct {
	// Enabled controls whether realtime tracking is active
	Enabled bool

	// CounterTTL is the TTL for Redis counters (default: 24 hours)
	CounterTTL time.Duration

	// SyncInterval is how often to sync counters to Postgres (default: 1 minute)
	SyncInterval time.Duration

	// WarningThresholdPercent triggers warning at this % of quota (default: 70)
	WarningThresholdPercent float64

	// CriticalThresholdPercent triggers critical alert at this % of quota (default: 90)
	CriticalThresholdPercent float64
}

// DefaultRealtimeUsageConfig returns default configuration
func DefaultRealtimeUsageConfig() *RealtimeUsageConfig {
	return &RealtimeUsageConfig{
		Enabled:                    true,
		CounterTTL:                 24 * time.Hour,
		SyncInterval:               1 * time.Minute,
		WarningThresholdPercent:    70,
		CriticalThresholdPercent: 90,
	}
}

// RealtimeQuotaStatus represents current quota usage for a tenant
type RealtimeQuotaStatus struct {
	TenantID uuid.UUID `json:"tenant_id"`

	// Function executions (per billing period)
	ExecutionsUsed    int `json:"executions_used"`
	ExecutionsLimit   int `json:"executions_limit"`
	ExecutionsPercent float64 `json:"executions_percent"`

	// Compute time in milliseconds (per billing period)
	ComputeMsUsed    int `json:"compute_ms_used"`
	ComputeMsLimit   int `json:"compute_ms_limit"`
	ComputeMsPercent float64 `json:"compute_ms_percent"`

	// Function count (total allowed)
	FunctionsUsed  int `json:"functions_used"`
	FunctionsLimit int `json:"functions_limit"`

	// Storage in MB (if applicable)
	StorageUsedMB  int `json:"storage_used_mb"`
	StorageLimitMB int `json:"storage_limit_mb"`

	// Status
	Status           string    `json:"status"` // "ok", "warning", "critical", "exceeded"
	LastUpdated      time.Time `json:"last_updated"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	WarningSentAt    *time.Time `json:"warning_sent_at,omitempty"`
	CriticalSentAt   *time.Time `json:"critical_sent_at,omitempty"`
}

// QuotaCheckResult represents the result of a quota check
type QuotaCheckResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	Status  *RealtimeQuotaStatus `json:"status,omitempty"`
}

// UsageCounterKeys defines Redis key patterns for usage counters
type UsageCounterKeys struct{}

// NewUsageCounterKeys creates a new key generator
func NewUsageCounterKeys() *UsageCounterKeys {
	return &UsageCounterKeys{}
}

// ExecutionsKey returns the Redis key for execution counter
func (k *UsageCounterKeys) ExecutionsKey(tenantID uuid.UUID, period string) string {
	return fmt.Sprintf("usage:executions:%s:%s", tenantID.String(), period)
}

// ComputeMsKey returns the Redis key for compute time counter
func (k *UsageCounterKeys) ComputeMsKey(tenantID uuid.UUID, period string) string {
	return fmt.Sprintf("usage:compute_ms:%s:%s", tenantID.String(), period)
}

// StatusKey returns the Redis key for quota status cache
func (k *UsageCounterKeys) StatusKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("usage:status:%s", tenantID.String())
}

// LastWarningKey returns the Redis key for last warning timestamp
func (k *UsageCounterKeys) LastWarningKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("usage:last_warning:%s", tenantID.String())
}

// LastCriticalKey returns the Redis key for last critical alert timestamp
func (k *UsageCounterKeys) LastCriticalKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("usage:last_critical:%s", tenantID.String())
}

// NewRealtimeUsageTracker creates a new realtime usage tracker
func NewRealtimeUsageTracker(
	redisClient *redis.Client,
	repo storage.Repository,
	notifySvc *notification.Service,
	config *RealtimeUsageConfig,
) *RealtimeUsageTracker {
	if config == nil {
		config = DefaultRealtimeUsageConfig()
	}

	return &RealtimeUsageTracker{
		redisClient:              redisClient,
		repo:                     repo,
		notifySvc:                notifySvc,
		logger:                   logrus.New(),
		enabled:                  config.Enabled,
		counterTTL:               config.CounterTTL,
		syncInterval:             config.SyncInterval,
		warningThresholdPercent:  config.WarningThresholdPercent,
		criticalThresholdPercent: config.CriticalThresholdPercent,
		stopChan:                 make(chan struct{}),
	}
}

// IsEnabled returns whether realtime tracking is enabled
func (t *RealtimeUsageTracker) IsEnabled() bool {
	return t.enabled && t.redisClient != nil
}

// Start begins the background sync routine
func (t *RealtimeUsageTracker) Start(ctx context.Context) {
	if !t.IsEnabled() {
		t.logger.Info("Realtime usage tracking is disabled")
		return
	}

	t.logger.WithFields(logrus.Fields{
		"sync_interval":              t.syncInterval,
		"counter_ttl":                  t.counterTTL,
		"warning_threshold_percent":    t.warningThresholdPercent,
		"critical_threshold_percent": t.criticalThresholdPercent,
	}).Info("Starting realtime usage tracker")

	// Run initial sync
	go t.syncLoop(ctx)
}

// Stop stops the background sync routine
func (t *RealtimeUsageTracker) Stop() {
	t.stopOnce.Do(func() {
		close(t.stopChan)
		t.logger.Info("Realtime usage tracker stopped")
	})
}

// syncLoop periodically syncs Redis counters to Postgres
func (t *RealtimeUsageTracker) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(t.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.logger.Info("Realtime usage tracker stopping due to context cancellation")
			return
		case <-t.stopChan:
			return
		case <-ticker.C:
			if err := t.SyncCountersToDatabase(ctx); err != nil {
				t.logger.WithError(err).Error("Failed to sync counters to database")
			}
		}
	}
}

// RecordExecution increments the execution counter for a tenant synchronously
// This is called BEFORE execution to provide immediate quota enforcement
func (t *RealtimeUsageTracker) RecordExecution(ctx context.Context, tenantID uuid.UUID, executionID string) (*QuotaCheckResult, error) {
	if !t.IsEnabled() {
		// If disabled, always allow (fallback to async tracking)
		return &QuotaCheckResult{Allowed: true}, nil
	}

	keys := NewUsageCounterKeys()
	period := getCurrentBillingPeriod()

	// Check quota before incrementing
	status, err := t.GetQuotaStatus(ctx, tenantID)
	if err != nil {
		t.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get quota status")
		// Allow execution on error (fail open to avoid blocking legitimate traffic)
		return &QuotaCheckResult{Allowed: true}, nil
	}

	// Check if quota is exceeded
	if status.ExecutionsUsed >= status.ExecutionsLimit {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Execution quota exceeded: %d of %d used", status.ExecutionsUsed, status.ExecutionsLimit),
			Status:  status,
		}, nil
	}

	// Increment execution counter atomically
	execKey := keys.ExecutionsKey(tenantID, period)
	pipe := t.redisClient.Pipeline()
	incrCmd := pipe.Incr(ctx, execKey)
	pipe.Expire(ctx, execKey, t.counterTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		t.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to increment execution counter")
		// Allow on error but log
		return &QuotaCheckResult{Allowed: true}, nil
	}

	newCount := int(incrCmd.Val())
	status.ExecutionsUsed = newCount
	status.ExecutionsPercent = float64(newCount) / float64(status.ExecutionsLimit) * 100

	// Update status in Redis
	if err := t.cacheQuotaStatus(ctx, status); err != nil {
		t.logger.WithError(err).Warn("Failed to cache quota status")
	}

	// Check thresholds and send notifications
	t.checkAndNotify(ctx, tenantID, status)

	return &QuotaCheckResult{
		Allowed: true,
		Status:  status,
	}, nil
}

// RecordComputeUsage records compute time usage for a tenant
func (t *RealtimeUsageTracker) RecordComputeUsage(ctx context.Context, tenantID uuid.UUID, cpuTimeMs int) error {
	if !t.IsEnabled() || cpuTimeMs <= 0 {
		return nil
	}

	keys := NewUsageCounterKeys()
	period := getCurrentBillingPeriod()

	computeKey := keys.ComputeMsKey(tenantID, period)
	pipe := t.redisClient.Pipeline()
	pipe.IncrBy(ctx, computeKey, int64(cpuTimeMs))
	pipe.Expire(ctx, computeKey, t.counterTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record compute usage: %w", err)
	}

	return nil
}

// GetQuotaStatus retrieves the current quota status for a tenant
// First checks Redis cache, then calculates from counters
func (t *RealtimeUsageTracker) GetQuotaStatus(ctx context.Context, tenantID uuid.UUID) (*RealtimeQuotaStatus, error) {
	if !t.IsEnabled() {
		// Fall back to database query if Redis is disabled
		return t.getQuotaStatusFromDB(ctx, tenantID)
	}

	keys := NewUsageCounterKeys()
	period := getCurrentBillingPeriod()

	// Try to get cached status first
	statusKey := keys.StatusKey(tenantID)
	cachedData, err := t.redisClient.Get(ctx, statusKey).Bytes()
	if err == nil {
		var status RealtimeQuotaStatus
		if err := json.Unmarshal(cachedData, &status); err == nil {
			// Check if still in current period
			if status.PeriodStart.Equal(getPeriodStart()) {
				return &status, nil
			}
		}
	}

	// Get subscription for limits
	sub, err := t.repo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		// No subscription - use free tier defaults
		return t.getDefaultQuotaStatus(tenantID), nil
	}

	// Get counters from Redis
	execKey := keys.ExecutionsKey(tenantID, period)
	computeKey := keys.ComputeMsKey(tenantID, period)

	pipe := t.redisClient.Pipeline()
	execCmd := pipe.Get(ctx, execKey)
	computeCmd := pipe.Get(ctx, computeKey)

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get counters: %w", err)
	}

	executionsUsed, _ := strconv.Atoi(execCmd.Val())
	computeMsUsed, _ := strconv.Atoi(computeCmd.Val())

	// Calculate limits from pricing tier
	limits := t.extractLimits(sub.PricingTier)

	// Calculate percentages
	execPercent := float64(executionsUsed) / float64(limits.ExecutionsLimit) * 100
	if limits.ExecutionsLimit == 0 {
		execPercent = 0
	}

	computePercent := float64(computeMsUsed) / float64(limits.ComputeMsLimit) * 100
	if limits.ComputeMsLimit == 0 {
		computePercent = 0
	}

	// Determine status
	status := "ok"
	if execPercent >= 100 || computePercent >= 100 {
		status = "exceeded"
	} else if execPercent >= t.criticalThresholdPercent || computePercent >= t.criticalThresholdPercent {
		status = "critical"
	} else if execPercent >= t.warningThresholdPercent || computePercent >= t.warningThresholdPercent {
		status = "warning"
	}

	result := &RealtimeQuotaStatus{
		TenantID:          tenantID,
		ExecutionsUsed:    executionsUsed,
		ExecutionsLimit:   limits.ExecutionsLimit,
		ExecutionsPercent: execPercent,
		ComputeMsUsed:     computeMsUsed,
		ComputeMsLimit:    limits.ComputeMsLimit,
		ComputeMsPercent:  computePercent,
		FunctionsLimit:    limits.FunctionsLimit,
		Status:            status,
		LastUpdated:       time.Now().UTC(),
		PeriodStart:       getPeriodStart(),
		PeriodEnd:         getPeriodEnd(),
	}

	// Cache the status
	t.cacheQuotaStatus(ctx, result)

	return result, nil
}

// CheckQuota checks if a tenant has available quota for execution
// This is a lightweight check that can be called frequently
func (t *RealtimeUsageTracker) CheckQuota(ctx context.Context, tenantID uuid.UUID) (*QuotaCheckResult, error) {
	return t.RecordExecution(ctx, tenantID, "")
}

// SyncCountersToDatabase syncs Redis counters to the database
// This is called periodically and ensures data persistence
func (t *RealtimeUsageTracker) SyncCountersToDatabase(ctx context.Context) error {
	if !t.IsEnabled() {
		return nil
	}

	start := time.Now()

	// Scan for all usage keys
	iter := t.redisClient.Scan(ctx, 0, "usage:executions:*", 0).Iterator()

	synced := 0
	errors := 0

	for iter.Next(ctx) {
		key := iter.Val()

		// Parse tenant ID from key
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}

		tenantID, err := uuid.Parse(parts[2])
		if err != nil {
			t.logger.WithError(err).WithField("key", key).Warn("Failed to parse tenant ID from key")
			continue
		}

		// Get counter value
		val, err := t.redisClient.Get(ctx, key).Int()
		if err != nil {
			if err != redis.Nil {
				t.logger.WithError(err).WithField("key", key).Warn("Failed to get counter value")
				errors++
			}
			continue
		}

		// Record to database as usage event
		event := &storage.UsageEvent{
			TenantID:  tenantID,
			EventType: "function_execution_realtime",
			Quantity:  val,
			Metadata: map[string]interface{}{
				"source":    "realtime_tracker",
				"synced_at": time.Now().UTC().Format(time.RFC3339),
			},
			Timestamp: time.Now().UTC(),
		}

		if err := t.repo.RecordUsageEvent(ctx, event); err != nil {
			t.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to record usage event")
			errors++
			continue
		}

		synced++
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to iterate keys: %w", err)
	}

	duration := time.Since(start)
	t.logger.WithFields(logrus.Fields{
		"synced":   synced,
		"errors":   errors,
		"duration": duration,
	}).Info("Synced counters to database")

	return nil
}

// InvalidateQuotaStatus clears the cached quota status for a tenant
// Call this when subscription changes or at period boundaries
func (t *RealtimeUsageTracker) InvalidateQuotaStatus(ctx context.Context, tenantID uuid.UUID) error {
	if !t.IsEnabled() {
		return nil
	}

	keys := NewUsageCounterKeys()

	// Delete status cache
	statusKey := keys.StatusKey(tenantID)
	if err := t.redisClient.Del(ctx, statusKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate status: %w", err)
	}

	return nil
}

// cacheQuotaStatus caches the quota status in Redis
func (t *RealtimeUsageTracker) cacheQuotaStatus(ctx context.Context, status *RealtimeQuotaStatus) error {
	if !t.IsEnabled() {
		return nil
	}

	keys := NewUsageCounterKeys()
	statusKey := keys.StatusKey(status.TenantID)

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Cache for 5 minutes
	if err := t.redisClient.Set(ctx, statusKey, data, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to cache status: %w", err)
	}

	return nil
}

// checkAndNotify checks thresholds and sends notifications
func (t *RealtimeUsageTracker) checkAndNotify(ctx context.Context, tenantID uuid.UUID, status *RealtimeQuotaStatus) {
	if t.notifySvc == nil {
		return
	}

	keys := NewUsageCounterKeys()

	// Check for critical threshold
	if status.ExecutionsPercent >= t.criticalThresholdPercent || status.ComputeMsPercent >= t.criticalThresholdPercent {
		// Check if we already sent a critical alert recently (within 1 hour)
		lastCriticalKey := keys.LastCriticalKey(tenantID)
		lastSent, err := t.redisClient.Get(ctx, lastCriticalKey).Int64()
		if err == nil {
			lastSentTime := time.Unix(lastSent, 0)
			if time.Since(lastSentTime) < 1*time.Hour {
				return // Already sent recently
			}
		}

		// Send critical alert
		t.sendQuotaAlert(ctx, tenantID, status, "critical")

		// Record that we sent the alert
		t.redisClient.Set(ctx, lastCriticalKey, time.Now().Unix(), 1*time.Hour)
		return
	}

	// Check for warning threshold
	if status.ExecutionsPercent >= t.warningThresholdPercent || status.ComputeMsPercent >= t.warningThresholdPercent {
		// Check if we already sent a warning recently (within 24 hours)
		lastWarningKey := keys.LastWarningKey(tenantID)
		lastSent, err := t.redisClient.Get(ctx, lastWarningKey).Int64()
		if err == nil {
			lastSentTime := time.Unix(lastSent, 0)
			if time.Since(lastSentTime) < 24*time.Hour {
				return // Already sent recently
			}
		}

		// Send warning
		t.sendQuotaAlert(ctx, tenantID, status, "warning")

		// Record that we sent the warning
		t.redisClient.Set(ctx, lastWarningKey, time.Now().Unix(), 24*time.Hour)
	}
}

// sendQuotaAlert sends a quota alert notification
func (t *RealtimeUsageTracker) sendQuotaAlert(ctx context.Context, tenantID uuid.UUID, status *RealtimeQuotaStatus, alertType string) {
	if t.notifySvc == nil {
		return
	}

	var title, body string
	priority := notification.PriorityNormal

	if alertType == "critical" {
		title = "Quota Limit Critical"
		body = fmt.Sprintf("You have used %.0f%% of your execution quota (%.0f%% of compute quota). Upgrade your plan to avoid service interruption.",
			status.ExecutionsPercent, status.ComputeMsPercent)
		priority = notification.PriorityHigh
	} else {
		title = "Quota Limit Warning"
		body = fmt.Sprintf("You have used %.0f%% of your execution quota (%.0f%% of compute quota). Consider upgrading your plan.",
			status.ExecutionsPercent, status.ComputeMsPercent)
	}

	_, err := t.notifySvc.Send(ctx, notification.SendRequest{
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    title,
		Body:     body,
		Data: map[string]interface{}{
			"tenant_id":          tenantID.String(),
			"executions_used":    status.ExecutionsUsed,
			"executions_limit":   status.ExecutionsLimit,
			"executions_percent": status.ExecutionsPercent,
			"compute_ms_used":    status.ComputeMsUsed,
			"compute_ms_limit":   status.ComputeMsLimit,
			"compute_percent":    status.ComputeMsPercent,
			"alert_type":         alertType,
			"threshold":          t.warningThresholdPercent,
		},
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: priority,
	})

	if err != nil {
		t.logger.WithError(err).Error("Failed to send quota alert")
	} else {
		t.logger.WithFields(logrus.Fields{
			"tenant_id":  tenantID,
			"alert_type": alertType,
		}).Info("Sent quota alert")
	}
}

// getQuotaStatusFromDB retrieves quota status from database (fallback)
func (t *RealtimeUsageTracker) getQuotaStatusFromDB(ctx context.Context, tenantID uuid.UUID) (*RealtimeQuotaStatus, error) {
	sub, err := t.repo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		return t.getDefaultQuotaStatus(tenantID), nil
	}

	// Get usage from database
	periodStart := getPeriodStart()
	periodEnd := getPeriodEnd()

	rollups, err := t.repo.GetUsageByTenant(tenantID, "function_execution", periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	executionsUsed := 0
	for _, rollup := range rollups {
		executionsUsed += rollup.TotalQuantity
	}

	limits := t.extractLimits(sub.PricingTier)

	return &RealtimeQuotaStatus{
		TenantID:        tenantID,
		ExecutionsUsed:    executionsUsed,
		ExecutionsLimit:   limits.ExecutionsLimit,
		ComputeMsUsed:     0, // Would need to query compute usage separately
		ComputeMsLimit:    limits.ComputeMsLimit,
		FunctionsLimit:    limits.FunctionsLimit,
		Status:            "ok",
		LastUpdated:       time.Now().UTC(),
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
	}, nil
}

// getDefaultQuotaStatus returns default free tier status
func (t *RealtimeUsageTracker) getDefaultQuotaStatus(tenantID uuid.UUID) *RealtimeQuotaStatus {
	return &RealtimeQuotaStatus{
		TenantID:        tenantID,
		ExecutionsUsed:    0,
		ExecutionsLimit:   1000, // Free tier default
		ComputeMsLimit:    3600000, // 1 hour
		FunctionsLimit:    5,
		Status:            "ok",
		LastUpdated:       time.Now().UTC(),
		PeriodStart:       getPeriodStart(),
		PeriodEnd:         getPeriodEnd(),
	}
}

// TenantLimits holds quota limits for a tenant
type TenantLimits struct {
	ExecutionsLimit  int
	ComputeMsLimit   int
	FunctionsLimit   int
	StorageLimitMB   int
}

// extractLimits extracts quota limits from a pricing tier
func (t *RealtimeUsageTracker) extractLimits(tier *storage.PricingTier) TenantLimits {
	limits := TenantLimits{
		ExecutionsLimit: 1000, // Default free tier
		ComputeMsLimit:  3600000, // 1 hour default
		FunctionsLimit:  5,
		StorageLimitMB:  100,
	}

	if tier == nil || tier.Features == nil {
		return limits
	}

	features, ok := tier.Features.(map[string]interface{})
	if !ok {
		return limits
	}

	// Extract requests limit
	if v, ok := features["requests"].(float64); ok {
		limits.ExecutionsLimit = int(v)
	}

	// Extract compute limit
	if v, ok := features["included_compute_ms"].(float64); ok {
		limits.ComputeMsLimit = int(v)
	}
	if v, ok := features["included_compute_hours"].(float64); ok {
		limits.ComputeMsLimit = int(v * 3600000)
	}

	// Extract functions limit
	if v, ok := features["functions"].(float64); ok {
		limits.FunctionsLimit = int(v)
	}

	return limits
}

// getCurrentBillingPeriod returns the current billing period string (YYYY-MM)
func getCurrentBillingPeriod() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

// getPeriodStart returns the start of the current billing period
func getPeriodStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// getPeriodEnd returns the end of the current billing period
func getPeriodEnd() time.Time {
	now := time.Now().UTC()
	// First day of next month
	nextMonth := now.AddDate(0, 1, 0)
	return time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
}

// RecordUsageFromExecution records usage from an execution record
// This is called after execution completes to sync the data
func (t *RealtimeUsageTracker) RecordUsageFromExecution(ctx context.Context, exec *storage.RegistryFunctionExecution, fn *storage.RegistryFunction) error {
	if !t.IsEnabled() || fn.TenantID == nil {
		return nil
	}

	tenantID := *fn.TenantID

	// Record execution
	if _, err := t.RecordExecution(ctx, tenantID, exec.ID.String()); err != nil {
		return fmt.Errorf("failed to record execution: %w", err)
	}

	// Record compute usage if available
	if exec.DurationMs > 0 {
		if err := t.RecordComputeUsage(ctx, tenantID, exec.DurationMs); err != nil {
			return fmt.Errorf("failed to record compute usage: %w", err)
		}
	}

	return nil
}

// GetRealtimeUsageMetrics returns metrics about the realtime tracking system
func (t *RealtimeUsageTracker) GetRealtimeUsageMetrics(ctx context.Context) (map[string]interface{}, error) {
	if !t.IsEnabled() {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	// Count active counters
	var executionCounters int64
	iter := t.redisClient.Scan(ctx, 0, "usage:executions:*", 0).Iterator()
	for iter.Next(ctx) {
		executionCounters++
	}

	var computeCounters int64
	iter = t.redisClient.Scan(ctx, 0, "usage:compute_ms:*", 0).Iterator()
	for iter.Next(ctx) {
		computeCounters++
	}

	return map[string]interface{}{
		"enabled":             true,
		"execution_counters":  executionCounters,
		"compute_counters":    computeCounters,
		"counter_ttl_seconds": t.counterTTL.Seconds(),
		"sync_interval":       t.syncInterval.String(),
	}, nil
}
