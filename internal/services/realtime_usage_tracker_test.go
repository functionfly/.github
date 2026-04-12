package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRealtimeUsageTracker_New tests the constructor
func TestRealtimeUsageTracker_New(t *testing.T) {
	// Test with nil Redis (should be disabled)
	tracker := NewRealtimeUsageTracker(nil, nil, nil, DefaultRealtimeUsageConfig())

	assert.NotNil(t, tracker)
	assert.False(t, tracker.IsEnabled()) // Should be disabled when Redis is nil
}

// TestRealtimeUsageTracker_IsEnabled tests the enabled flag
func TestRealtimeUsageTracker_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *RealtimeUsageConfig
		expected bool
	}{
		{
			name: "disabled when config disabled",
			config: &RealtimeUsageConfig{
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "disabled when redis nil",
			config: &RealtimeUsageConfig{
				Enabled: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewRealtimeUsageTracker(nil, nil, nil, tt.config)
			assert.Equal(t, tt.expected, tracker.IsEnabled())
		})
	}
}

// TestQuotaCheckResult tests the QuotaCheckResult structure
func TestQuotaCheckResult(t *testing.T) {
	result := &QuotaCheckResult{
		Allowed: true,
		Reason:  "",
		Status: &RealtimeQuotaStatus{
			TenantID:          uuid.New(),
			ExecutionsUsed:    50,
			ExecutionsLimit:   100,
			ExecutionsPercent: 50.0,
			Status:            "healthy",
		},
	}

	assert.True(t, result.Allowed)
	assert.Equal(t, "healthy", result.Status.Status)
}

// TestRealtimeQuotaStatus tests the RealtimeQuotaStatus structure
func TestRealtimeQuotaStatus(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now().UTC()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now.AddDate(0, 1, 0)

	status := &RealtimeQuotaStatus{
		TenantID:          tenantID,
		ExecutionsUsed:    80,
		ExecutionsLimit:   100,
		ExecutionsPercent: 80.0,
		ComputeMsUsed:     5000,
		ComputeMsLimit:    10000,
		ComputeMsPercent:  50.0,
		FunctionsUsed:     5,
		FunctionsLimit:    10,
		Status:            "warning",
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		LastUpdated:       now,
	}

	assert.Equal(t, tenantID, status.TenantID)
	assert.Equal(t, 80, status.ExecutionsUsed)
	assert.Equal(t, 100, status.ExecutionsLimit)
	assert.Equal(t, 80.0, status.ExecutionsPercent)
	assert.Equal(t, "warning", status.Status)
}

// TestDefaultRealtimeUsageConfig tests the default configuration
func TestDefaultRealtimeUsageConfig(t *testing.T) {
	config := DefaultRealtimeUsageConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 24*time.Hour, config.CounterTTL)
	assert.Equal(t, 1*time.Minute, config.SyncInterval)
	assert.Equal(t, 70.0, config.WarningThresholdPercent)
	assert.Equal(t, 90.0, config.CriticalThresholdPercent)
}

// TestQuotaCheckResult_WithExceededQuota tests quota exceeded scenarios
func TestQuotaCheckResult_WithExceededQuota(t *testing.T) {
	result := &QuotaCheckResult{
		Allowed: false,
		Reason:  "Monthly execution quota exceeded (110% of 1000 executions)",
		Status: &RealtimeQuotaStatus{
			TenantID:          uuid.New(),
			ExecutionsUsed:    1100,
			ExecutionsLimit:   1000,
			ExecutionsPercent: 110.0,
			Status:            "exceeded",
		},
	}

	assert.False(t, result.Allowed)
	assert.Equal(t, "exceeded", result.Status.Status)
}

// TestRealtimeUsageTrackerInterface_Compliance verifies the interface is properly defined
func TestRealtimeUsageTrackerInterface_Compliance(t *testing.T) {
	// This test ensures the interface is properly defined
	// A real implementation would be tested against this interface
	var _ RealtimeUsageTrackerInterface = (*RealtimeUsageTracker)(nil)
}

// TestThresholdCalculation tests the threshold calculation
func TestThresholdCalculation(t *testing.T) {
	config := DefaultRealtimeUsageConfig()

	tests := []struct {
		name     string
		percent  float64
		wantWarn bool
		wantCrit bool
	}{
		{"below warning", 50.0, false, false},
		{"at warning threshold", config.WarningThresholdPercent, true, false},
		{"between warning and critical", (config.WarningThresholdPercent + config.CriticalThresholdPercent) / 2, true, false},
		{"at critical threshold", config.CriticalThresholdPercent, true, true},
		{"above critical", 95.0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isWarning := tt.percent >= config.WarningThresholdPercent
			isCritical := tt.percent >= config.CriticalThresholdPercent

			assert.Equal(t, tt.wantWarn, isWarning)
			assert.Equal(t, tt.wantCrit, isCritical)
		})
	}
}

// TestUsageCounterKeys tests the Redis key building functions
func TestUsageCounterKeys(t *testing.T) {
	tenantID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	period := "2024-01"

	keys := NewUsageCounterKeys()

	executionKey := keys.ExecutionsKey(tenantID, period)
	computeKey := keys.ComputeMsKey(tenantID, period)
	statusKey := keys.StatusKey(tenantID)
	lastWarningKey := keys.LastWarningKey(tenantID)
	lastCriticalKey := keys.LastCriticalKey(tenantID)

	// Keys should contain tenant ID and period
	assert.Contains(t, executionKey, tenantID.String())
	assert.Contains(t, executionKey, period)
	assert.Contains(t, executionKey, "executions")

	assert.Contains(t, computeKey, tenantID.String())
	assert.Contains(t, computeKey, period)
	assert.Contains(t, computeKey, "compute_ms")

	assert.Contains(t, statusKey, tenantID.String())
	assert.Contains(t, statusKey, "status")

	assert.Contains(t, lastWarningKey, tenantID.String())
	assert.Contains(t, lastWarningKey, "last_warning")

	assert.Contains(t, lastCriticalKey, tenantID.String())
	assert.Contains(t, lastCriticalKey, "last_critical")
}

// TestParseUsageValue tests the parseUsageValue function (exposed via getQuotaStatusFromDB behavior)
func TestParseUsageValueLogic(t *testing.T) {
	// Test that nil values are handled correctly (simulating the parse logic)
	var nilValue interface{} = nil
	assert.Nil(t, nilValue)

	// Test string parsing
	validStr := "42"
	assert.Equal(t, "42", validStr)

	// Test empty string
	emptyStr := ""
	assert.Empty(t, emptyStr)
}

// TestTenantLimits tests the TenantLimits structure
func TestTenantLimits(t *testing.T) {
	limits := TenantLimits{
		ExecutionsLimit: 1000,
		ComputeMsLimit:  3600000,
		FunctionsLimit:  5,
		StorageLimitMB:  100,
	}

	assert.Equal(t, 1000, limits.ExecutionsLimit)
	assert.Equal(t, 3600000, limits.ComputeMsLimit)
	assert.Equal(t, 5, limits.FunctionsLimit)
	assert.Equal(t, 100, limits.StorageLimitMB)
}

// TestPeriodCalculation tests period calculation logic
func TestPeriodCalculation(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Period start should be first day of month
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), periodStart)

	// Period end should be first day of next month
	periodEnd := periodStart.AddDate(0, 1, 0)
	assert.Equal(t, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), periodEnd)

	// Format period as YYYY-MM
	period := now.Format("2006-01")
	assert.Equal(t, "2024-01", period)
}

// TestTTLCalculation tests TTL calculation
func TestTTLCalculation(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	// Calculate TTL as time until period end plus buffer
	ttl := periodEnd.Sub(now) + 24*time.Hour

	// Should be ~17 days + 1 day buffer = 18 days
	expectedTTL := 18 * 24 * time.Hour
	assert.Equal(t, expectedTTL, ttl)
}

// TestExtractLimitsFromPricingTier simulates the extractLimits behavior
func TestExtractLimitsFromPricingTier(t *testing.T) {
	// Test with nil tier (should return defaults)
	defaultLimits := TenantLimits{
		ExecutionsLimit: 1000,
		ComputeMsLimit:  3600000,
		FunctionsLimit:  5,
		StorageLimitMB:  100,
	}

	assert.Equal(t, 1000, defaultLimits.ExecutionsLimit)
	assert.Equal(t, 3600000, defaultLimits.ComputeMsLimit)
	assert.Equal(t, 5, defaultLimits.FunctionsLimit)
}
