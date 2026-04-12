package storage

import (
	"time"

	"github.com/google/uuid"
)

// UsageAlert represents a usage alert configuration for a tenant
type UsageAlert struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Name              string     `json:"name"`                            // Alert name/description
	AlertType         string     `json:"alert_type"`                      // 'spend_cap', 'usage_spike', 'threshold'
	ThresholdValue    float64    `json:"threshold_value"`                 // Threshold value (cents for spend, count for usage)
	ThresholdOperator string     `json:"threshold_operator"`              // 'gte', 'lte', 'percentage_of_cap'
	PeriodType        string     `json:"period_type"`                     // 'billing_period', 'daily', 'weekly'
	NotificationChannels []string `json:"notification_channels"`          // ['email', 'in_app', 'webhook']
	IsEnabled         bool       `json:"is_enabled"`
	LastTriggeredAt   *time.Time `json:"last_triggered_at,omitempty"`
	TriggerCount      int        `json:"trigger_count"`
	CooldownMinutes   int        `json:"cooldown_minutes"`                // Minimum minutes between alerts
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UsageAlertHistory tracks when alerts were triggered
type UsageAlertHistory struct {
	ID            uuid.UUID              `json:"id"`
	AlertID       uuid.UUID              `json:"alert_id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	TriggeredAt   time.Time              `json:"triggered_at"`
	TriggeredValue float64               `json:"triggered_value"`              // The value that triggered the alert
	ThresholdValue float64               `json:"threshold_value"`              // The threshold that was exceeded
	Message       string                 `json:"message"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	AcknowledgedAt *time.Time            `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID            `json:"acknowledged_by,omitempty"`
}

// SpendCap represents a spending limit for a tenant
type SpendCap struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	CapAmountCents  int        `json:"cap_amount_cents"`              // Hard cap amount
	WarningThresholds []int    `json:"warning_thresholds"`            // Percentage thresholds for warnings (e.g., [50, 75, 90])
	CurrentSpendCents int      `json:"current_spend_cents"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	ActionOnCap     string     `json:"action_on_cap"`                 // 'throttle', 'suspend', 'notify_only'
	IsHardCap       bool       `json:"is_hard_cap"`                   // If true, enforce the cap
	IsEnabled       bool       `json:"is_enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UsageForecast represents a predicted usage forecast
type UsageForecast struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	ForecastType    string                 `json:"forecast_type"`               // 'spend', 'executions', 'compute_time'
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	CurrentValue    float64                `json:"current_value"`              // Current actual value
	PredictedValue  float64                `json:"predicted_value"`            // Predicted value at period end
	LowerBound      float64                `json:"lower_bound"`                // Lower bound of prediction (80% confidence)
	UpperBound      float64                `json:"upper_bound"`                // Upper bound of prediction (80% confidence)
	Confidence      float64                `json:"confidence"`                 // Prediction confidence (0-1)
	MethodUsed      string                 `json:"method_used"`                // 'linear', 'exponential_smoothing', 'seasonal'
	GrowthRate      float64                `json:"growth_rate"`                // Detected growth rate
	DaysOfHistory   int                    `json:"days_of_history"`            // Days of historical data used
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// UsageTrend represents usage trend analysis
type UsageTrend struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	EventType          string    `json:"event_type"`
	PeriodAnalyzed     string    `json:"period_analyzed"`            // e.g., '7d', '30d'
	AvgDailyUsage      float64   `json:"avg_daily_usage"`
	PeakDailyUsage     float64   `json:"peak_daily_usage"`
	MinDailyUsage      float64   `json:"min_daily_usage"`
	TrendDirection     string    `json:"trend_direction"`            // 'increasing', 'decreasing', 'stable'
	TrendPercentChange float64   `json:"trend_percent_change"`       // Percent change over period
	SeasonalityScore   float64   `json:"seasonality_score"`        // 0-1, how seasonal the usage is
	VolatilityScore    float64   `json:"volatility_score"`         // Coefficient of variation
	AnomalyCount       int       `json:"anomaly_count"`            // Number of anomalous days
	ForecastAccuracy   float64   `json:"forecast_accuracy"`        // Last forecast vs actual (0-1)
	CalculatedAt       time.Time `json:"calculated_at"`
}

// DailyUsagePoint represents a single day's usage for time series analysis
type DailyUsagePoint struct {
	Date     time.Time `json:"date"`
	Value    float64   `json:"value"`    // Usage value (cents for spend, count for executions)
	IsAnomaly bool     `json:"is_anomaly"`
}

// UsageForecastConfig holds configuration for the forecasting engine
type UsageForecastConfig struct {
	MinHistoryDays      int     `json:"min_history_days"`        // Minimum days of history needed (default: 7)
	MaxHistoryDays      int     `json:"max_history_days"`        // Maximum days to use for forecasting (default: 90)
	DefaultConfidence   float64 `json:"default_confidence"`      // Default confidence interval (default: 0.80)
	SeasonalityWindow   int     `json:"seasonality_window"`      // Days to check for seasonality (default: 7)
	AnomalyThreshold    float64 `json:"anomaly_threshold"`       // Z-score threshold for anomalies (default: 3.0)
	GrowthSmoothing   float64 `json:"growth_smoothing"`        // Exponential smoothing factor (default: 0.3)
}

// DefaultUsageForecastConfig returns the default forecasting configuration
func DefaultUsageForecastConfig() *UsageForecastConfig {
	return &UsageForecastConfig{
		MinHistoryDays:    7,
		MaxHistoryDays:    90,
		DefaultConfidence: 0.80,
		SeasonalityWindow: 7,
		AnomalyThreshold:  3.0,
		GrowthSmoothing:   0.3,
	}
}
