package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UsageForecasterConfig holds configuration for the forecasting service
type UsageForecasterConfig struct {
	Enabled           bool
	ForecastInterval  time.Duration
	MinHistoryDays    int
	MaxHistoryDays    int
	ConfidenceLevel   float64
	AlertCheckInterval time.Duration
}

// DefaultUsageForecasterConfig returns default configuration
func DefaultUsageForecasterConfig() *UsageForecasterConfig {
	return &UsageForecasterConfig{
		Enabled:           true,
		ForecastInterval:  6 * time.Hour, // Generate forecasts every 6 hours
		MinHistoryDays:    7,
		MaxHistoryDays:    90,
		ConfidenceLevel:   0.80,
		AlertCheckInterval: 15 * time.Minute, // Check alerts every 15 minutes
	}
}

// LoadUsageForecasterConfig loads configuration from environment
func LoadUsageForecasterConfig() *UsageForecasterConfig {
	config := DefaultUsageForecasterConfig()

	if v := os.Getenv("USAGE_FORECASTING_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("USAGE_FORECAST_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			config.ForecastInterval = d
		}
	}

	if v := os.Getenv("USAGE_ALERT_CHECK_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			config.AlertCheckInterval = d
		}
	}

	if v := os.Getenv("USAGE_FORECAST_MIN_HISTORY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MinHistoryDays = n
		}
	}

	if v := os.Getenv("USAGE_FORECAST_CONFIDENCE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f <= 1 {
			config.ConfidenceLevel = f
		}
	}

	return config
}

// UsageForecaster provides predictive analytics for usage
type UsageForecaster struct {
	alertRepo  *storage.UsageAlertRepository
	billingRepo storage.Repository
	config     *UsageForecasterConfig
	logger     *logrus.Logger
	stopChan   chan struct{}
	stopOnce   sync.Once
}

// NewUsageForecaster creates a new usage forecaster
func NewUsageForecaster(alertRepo *storage.UsageAlertRepository, billingRepo storage.Repository, config *UsageForecasterConfig) *UsageForecaster {
	return &UsageForecaster{
		alertRepo:   alertRepo,
		billingRepo: billingRepo,
		config:      config,
		logger:      logrus.New(),
		stopChan:    make(chan struct{}),
	}
}

// Start begins the forecasting and alerting service
func (f *UsageForecaster) Start(ctx context.Context) {
	if !f.config.Enabled {
		f.logger.Info("Usage forecasting service is disabled")
		return
	}

	f.logger.WithFields(logrus.Fields{
		"forecast_interval":    f.config.ForecastInterval,
		"alert_check_interval": f.config.AlertCheckInterval,
	}).Info("Starting usage forecasting service")

	// Run initial forecast
	go f.runInitialForecast(ctx)

	// Start forecasting loop
	go f.runForecastLoop(ctx)
}

// Stop stops the forecasting service
func (f *UsageForecaster) Stop() {
	f.stopOnce.Do(func() {
		close(f.stopChan)
	})
}

// runInitialForecast runs the initial forecast on startup
func (f *UsageForecaster) runInitialForecast(ctx context.Context) {
	if err := f.GenerateAllForecasts(ctx); err != nil {
		f.logger.WithError(err).Error("Initial forecast generation failed")
	}
}

// runForecastLoop runs the periodic forecasting loop
func (f *UsageForecaster) runForecastLoop(ctx context.Context) {
	ticker := time.NewTicker(f.config.ForecastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Info("Forecasting loop stopping due to context cancellation")
			return
		case <-f.stopChan:
			f.logger.Info("Forecasting loop stopped")
			return
		case <-ticker.C:
			if err := f.GenerateAllForecasts(ctx); err != nil {
				f.logger.WithError(err).Error("Forecast generation failed")
			}
		}
	}
}

// GenerateAllForecasts generates forecasts for all tenants
func (f *UsageForecaster) GenerateAllForecasts(ctx context.Context) error {
	start := time.Now()

	// Get all active subscriptions
	subs, err := f.billingRepo.ListAllSubscriptions(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	forecastCount := 0
	for _, sub := range subs {
		if sub.Status != "active" && sub.Status != "trialing" {
			continue
		}

		// Generate spend forecast
		if err := f.GenerateSpendForecast(ctx, sub.TenantID, sub); err != nil {
			f.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("Failed to generate spend forecast")
		} else {
			forecastCount++
		}

		// Generate usage forecasts
		if err := f.GenerateUsageForecast(ctx, sub.TenantID, "function_execution", sub); err != nil {
			f.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("Failed to generate execution forecast")
		} else {
			forecastCount++
		}
	}

	f.logger.WithFields(logrus.Fields{
		"duration":         time.Since(start),
		"forecasts_generated": forecastCount,
		"tenants_processed": len(subs),
	}).Info("Forecast generation completed")

	return nil
}

// GenerateSpendForecast generates a spend forecast for a tenant
func (f *UsageForecaster) GenerateSpendForecast(ctx context.Context, tenantID uuid.UUID, sub *storage.Subscription) error {
	// Get historical spend data
	history, err := f.alertRepo.GetDailySpendHistory(ctx, tenantID, f.config.MaxHistoryDays)
	if err != nil {
		return fmt.Errorf("failed to get spend history: %w", err)
	}

	if len(history) < f.config.MinHistoryDays {
		return fmt.Errorf("insufficient history: %d days, need %d", len(history), f.config.MinHistoryDays)
	}

	// Calculate current spend for the period
	periodStart := sub.CurrentPeriodStart
	periodEnd := sub.CurrentPeriodEnd

	summary, err := f.alertRepo.GetCurrentPeriodUsage(ctx, tenantID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to get current period usage: %w", err)
	}

	// Calculate spend from usage (this is a simplified calculation)
	currentSpend := float64(summary.EstimatedCost)

	// Generate forecast using time series analysis
	forecast := f.predictSpendAtPeriodEnd(history, currentSpend, periodStart, periodEnd)

	// Save forecast
	forecast.TenantID = tenantID
	forecast.ForecastType = "spend"
	forecast.PeriodStart = periodStart
	forecast.PeriodEnd = periodEnd
	forecast.CurrentValue = currentSpend
	forecast.DaysOfHistory = len(history)

	if err := f.alertRepo.SaveUsageForecast(ctx, forecast); err != nil {
		return fmt.Errorf("failed to save forecast: %w", err)
	}

	return nil
}

// GenerateUsageForecast generates a usage forecast for a specific metric
func (f *UsageForecaster) GenerateUsageForecast(ctx context.Context, tenantID uuid.UUID, eventType string, sub *storage.Subscription) error {
	// Get historical usage data
	history, err := f.alertRepo.GetDailyUsageHistory(ctx, tenantID, eventType, f.config.MaxHistoryDays)
	if err != nil {
		return fmt.Errorf("failed to get usage history: %w", err)
	}

	if len(history) < f.config.MinHistoryDays {
		return fmt.Errorf("insufficient history: %d days, need %d", len(history), f.config.MinHistoryDays)
	}

	// Calculate current usage for the period
	periodStart := sub.CurrentPeriodStart
	periodEnd := sub.CurrentPeriodEnd

	summary, err := f.alertRepo.GetCurrentPeriodUsage(ctx, tenantID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to get current period usage: %w", err)
	}

	var currentValue float64
	switch eventType {
	case "function_execution":
		currentValue = float64(summary.TotalExecutions)
	case "compute_time_ms":
		currentValue = float64(summary.TotalComputeMs)
	default:
		currentValue = 0
	}

	// Generate forecast
	forecast := f.predictUsageAtPeriodEnd(history, currentValue, periodStart, periodEnd)

	// Save forecast
	forecast.TenantID = tenantID
	forecast.ForecastType = eventType
	forecast.PeriodStart = periodStart
	forecast.PeriodEnd = periodEnd
	forecast.CurrentValue = currentValue
	forecast.DaysOfHistory = len(history)

	if err := f.alertRepo.SaveUsageForecast(ctx, forecast); err != nil {
		return fmt.Errorf("failed to save forecast: %w", err)
	}

	return nil
}

// predictSpendAtPeriodEnd predicts spend at period end using multiple methods
func (f *UsageForecaster) predictSpendAtPeriodEnd(history []*storage.DailyUsagePoint, currentSpend float64, periodStart, periodEnd time.Time) *storage.UsageForecast {
	values := make([]float64, len(history))
	for i, h := range history {
		values[i] = h.Value
	}

	// Calculate time remaining
	now := time.Now().UTC()
	daysElapsed := now.Sub(periodStart).Hours() / 24
	daysRemaining := periodEnd.Sub(now).Hours() / 24
	daysInPeriod := periodEnd.Sub(periodStart).Hours() / 24

	if daysElapsed <= 0 || daysRemaining <= 0 {
		// Period hasn't started or already ended
		return &storage.UsageForecast{
			PredictedValue: currentSpend,
			LowerBound:     currentSpend * 0.95,
			UpperBound:     currentSpend * 1.05,
			Confidence:     0.95,
			MethodUsed:     "current_only",
			GrowthRate:     0,
		}
	}

	// Method 1: Linear projection based on daily average
	dailyAvg := currentSpend / daysElapsed
	linearProjection := currentSpend + (dailyAvg * daysRemaining)

	// Method 2: Trend-adjusted projection using recent growth rate
	growthRate := f.calculateGrowthRate(values, 7) // 7-day growth rate
	trendProjection := linearProjection * (1 + growthRate*(daysRemaining/daysInPeriod))

	// Method 3: Exponential smoothing forecast
	smoothedProjection := f.exponentialSmoothingForecast(values, currentSpend, daysRemaining)

	// Weighted ensemble: prefer trend if recent data is available, otherwise linear
	var prediction float64
	methodUsed := "linear"

	if len(values) >= 14 && math.Abs(growthRate) > 0.05 {
		// Significant trend detected, use trend-adjusted projection
		prediction = trendProjection*0.5 + smoothedProjection*0.5
		methodUsed = "exponential_smoothing"
	} else if f.detectSeasonality(values) > 0.3 {
		// Seasonal pattern detected
		prediction = f.seasonalForecast(values, currentSpend, daysRemaining)
		methodUsed = "seasonal"
	} else {
		// Default to linear with exponential smoothing blend
		prediction = linearProjection*0.4 + smoothedProjection*0.6
		methodUsed = "linear_blend"
	}

	// Calculate confidence interval using historical volatility
	volatility := f.calculateVolatility(values)
	margin := prediction * volatility * (1 + daysRemaining/daysInPeriod)

	confidence := f.config.ConfidenceLevel
	if len(values) < 14 {
		confidence = 0.60 // Lower confidence with less data
	} else if len(values) < 30 {
		confidence = 0.75
	}

	return &storage.UsageForecast{
		PredictedValue: prediction,
		LowerBound:     math.Max(currentSpend, prediction-margin),
		UpperBound:     prediction + margin,
		Confidence:     confidence,
		MethodUsed:     methodUsed,
		GrowthRate:     growthRate,
	}
}

// predictUsageAtPeriodEnd predicts usage at period end
func (f *UsageForecaster) predictUsageAtPeriodEnd(history []*storage.DailyUsagePoint, currentValue float64, periodStart, periodEnd time.Time) *storage.UsageForecast {
	// Same logic as spend prediction but for usage metrics
	values := make([]float64, len(history))
	for i, h := range history {
		values[i] = h.Value
	}

	now := time.Now().UTC()
	daysElapsed := now.Sub(periodStart).Hours() / 24
	daysRemaining := periodEnd.Sub(now).Hours() / 24
	daysInPeriod := periodEnd.Sub(periodStart).Hours() / 24

	if daysElapsed <= 0 || daysRemaining <= 0 {
		return &storage.UsageForecast{
			PredictedValue: currentValue,
			LowerBound:     currentValue * 0.95,
			UpperBound:     currentValue * 1.05,
			Confidence:     0.95,
			MethodUsed:     "current_only",
			GrowthRate:     0,
		}
	}

	dailyAvg := currentValue / daysElapsed
	linearProjection := currentValue + (dailyAvg * daysRemaining)

	growthRate := f.calculateGrowthRate(values, 7)
	volatility := f.calculateVolatility(values)

	prediction := linearProjection * (1 + growthRate*(daysRemaining/daysInPeriod)*0.5)
	margin := prediction * volatility * (1 + daysRemaining/daysInPeriod)

	confidence := f.config.ConfidenceLevel
	if len(values) < 14 {
		confidence = 0.65
	}

	return &storage.UsageForecast{
		PredictedValue: prediction,
		LowerBound:     math.Max(currentValue, prediction-margin),
		UpperBound:     prediction + margin,
		Confidence:     confidence,
		MethodUsed:     "linear",
		GrowthRate:     growthRate,
	}
}

// calculateGrowthRate calculates the growth rate over recent days
func (f *UsageForecaster) calculateGrowthRate(values []float64, window int) float64 {
	if len(values) < window*2 {
		return 0
	}

	// Calculate average of recent window vs previous window
	recentSum := 0.0
	previousSum := 0.0

	for i := 0; i < window; i++ {
		recentSum += values[len(values)-1-i]
		previousSum += values[len(values)-1-window-i]
	}

	recentAvg := recentSum / float64(window)
	previousAvg := previousSum / float64(window)

	if previousAvg == 0 {
		return 0
	}

	return (recentAvg - previousAvg) / previousAvg
}

// calculateVolatility calculates coefficient of variation
func (f *UsageForecaster) calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0.1 // Default 10% volatility
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	if mean == 0 {
		return 0.1
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(values)))

	return stdDev / mean
}

// detectSeasonality detects if there's a weekly pattern in the data
func (f *UsageForecaster) detectSeasonality(values []float64) float64 {
	if len(values) < 14 {
		return 0 // Not enough data
	}

	// Calculate correlation between values 7 days apart
	correlations := []float64{}
	for offset := 0; offset < 7; offset++ {
		if len(values) < 14+offset {
			continue
		}
		x := []float64{}
		y := []float64{}
		for i := offset; i+7 < len(values); i += 7 {
			x = append(x, values[i])
			y = append(y, values[i+7])
		}
		if len(x) > 1 {
			correlations = append(correlations, f.correlation(x, y))
		}
	}

	if len(correlations) == 0 {
		return 0
	}

	// Average correlation
	avgCorr := 0.0
	for _, c := range correlations {
		avgCorr += c
	}
	avgCorr /= float64(len(correlations))

	// Return seasonality score (0-1)
	if avgCorr < 0 {
		return 0
	}
	return avgCorr
}

// correlation calculates Pearson correlation coefficient
func (f *UsageForecaster) correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	meanX, meanY := 0.0, 0.0
	for i := 0; i < len(x); i++ {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= n
	meanY /= n

	num, denX, denY := 0.0, 0.0, 0.0
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		num += dx * dy
		denX += dx * dx
		denY += dy * dy
	}

	if denX == 0 || denY == 0 {
		return 0
	}

	return num / math.Sqrt(denX*denY)
}

// exponentialSmoothingForecast applies exponential smoothing
func (f *UsageForecaster) exponentialSmoothingForecast(values []float64, currentValue float64, daysRemaining float64) float64 {
	if len(values) < 2 {
		return currentValue
	}

	alpha := 0.3 // Smoothing factor

	// Calculate smoothed values
	smoothed := values[0]
	for i := 1; i < len(values); i++ {
		smoothed = alpha*values[i] + (1-alpha)*smoothed
	}

	// Project forward
	dailyTrend := smoothed - values[max(0, len(values)-2)]
	if dailyTrend < 0 {
		dailyTrend = 0 // Don't project negative growth
	}

	return currentValue + (dailyTrend * daysRemaining)
}

// seasonalForecast generates a forecast using weekly seasonality
func (f *UsageForecaster) seasonalForecast(values []float64, currentValue float64, daysRemaining float64) float64 {
	if len(values) < 14 {
		return currentValue + (currentValue/float64(len(values)))*daysRemaining
	}

	// Calculate day-of-week factors
	dayOfWeekAvg := make([]float64, 7)
	dayOfWeekCount := make([]int, 7)

	for i, v := range values {
		dayOfWeek := i % 7
		dayOfWeekAvg[dayOfWeek] += v
		dayOfWeekCount[dayOfWeek]++
	}

	for i := 0; i < 7; i++ {
		if dayOfWeekCount[i] > 0 {
			dayOfWeekAvg[i] /= float64(dayOfWeekCount[i])
		}
	}

	// Calculate overall average
	overallAvg := 0.0
	for _, v := range dayOfWeekAvg {
		overallAvg += v
	}
	overallAvg /= 7

	// Calculate daily average
	daysElapsed := float64(len(values))
	dailyAvg := currentValue / daysElapsed

	// Apply seasonality to projection
	seasonalFactor := 1.0
	if overallAvg > 0 {
		// Use average of upcoming day-of-week factors
		futureFactorSum := 0.0
		for i := 0; i < int(daysRemaining) && i < 30; i++ {
			futureDay := (len(values) + i) % 7
			if dayOfWeekAvg[futureDay] > 0 {
				futureFactorSum += dayOfWeekAvg[futureDay] / overallAvg
			}
		}
		if daysRemaining > 0 {
			seasonalFactor = futureFactorSum / math.Min(daysRemaining, 30)
		}
	}

	return currentValue + (dailyAvg * seasonalFactor * daysRemaining)
}

// GetForecastForTenant retrieves the latest forecast for a tenant
func (f *UsageForecaster) GetForecastForTenant(ctx context.Context, tenantID uuid.UUID, forecastType string) (*storage.UsageForecast, error) {
	return f.alertRepo.GetLatestForecast(ctx, tenantID, forecastType)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
