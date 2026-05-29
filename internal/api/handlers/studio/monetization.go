package studio

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

const monetizationMetricsWindow = 30 * 24 * time.Hour

// MonetizationMetrics holds usage signals used to score pricing models.
type MonetizationMetrics struct {
	ExecutionCount   int     `json:"execution_count"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	ErrorRate        float64 `json:"error_rate"`
	UserCount        int     `json:"user_count"`
}

// PricingRecommendation is an AI-scored monetization option.
type PricingRecommendation struct {
	Model           string  `json:"model"`
	Price           float64 `json:"price"`
	Confidence      float64 `json:"confidence"`
	Reasoning       string  `json:"reasoning"`
	ExpectedRevenue float64 `json:"expected_revenue"`
}

// CurrentPricing reflects the active monetization configuration.
type CurrentPricing struct {
	Model string  `json:"model"`
	Price float64 `json:"price"`
}

// MonetizationAnalysis is returned by the optimizer endpoint.
type MonetizationAnalysis struct {
	FunctionID         string                  `json:"function_id"`
	Metrics            MonetizationMetrics     `json:"metrics"`
	CurrentPricing     CurrentPricing          `json:"current_pricing"`
	Recommendations    []PricingRecommendation `json:"recommendations"`
	BestRecommendation PricingRecommendation   `json:"best_recommendation"`
}

// GetMonetizationAnalysis loads metrics and generates pricing recommendations.
func (r *MarketplaceRepository) GetMonetizationAnalysis(ctx context.Context, tenantID, functionID string) (*MonetizationAnalysis, error) {
	if err := r.assertFunctionAccess(ctx, tenantID, functionID); err != nil {
		return nil, err
	}

	metrics, err := r.getFunctionMetrics(ctx, functionID)
	if err != nil {
		return nil, err
	}

	current, err := r.getCurrentPricing(ctx, functionID)
	if err != nil {
		return nil, err
	}

	recommendations := generatePricingRecommendations(metrics)
	best := recommendations[0]
	for _, rec := range recommendations[1:] {
		if rec.Confidence > best.Confidence {
			best = rec
		}
	}

	return &MonetizationAnalysis{
		FunctionID:         functionID,
		Metrics:            metrics,
		CurrentPricing:     current,
		Recommendations:    recommendations,
		BestRecommendation: best,
	}, nil
}

// ApplyMonetizationRecommendation persists the selected pricing model.
func (r *MarketplaceRepository) ApplyMonetizationRecommendation(ctx context.Context, tenantID, functionID, model string, price float64) error {
	model = strings.TrimSpace(strings.ToLower(model))
	if !validPricingModel(model) {
		return fmt.Errorf("unsupported pricing model: %s", model)
	}
	if price < 0 {
		return fmt.Errorf("price must be non-negative")
	}

	if err := r.assertFunctionOwned(ctx, tenantID, functionID); err != nil {
		return err
	}

	return r.upsertFunctionListing(ctx, functionID, model, price)
}

func (r *MarketplaceRepository) assertFunctionAccess(ctx context.Context, tenantID, functionID string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM registry_functions
			WHERE id = $1::uuid
			  AND (
				visibility IN ('public', 'unlisted')
				OR tenant_id = NULLIF($2, '')::uuid
			  )
		)`, functionID, tenantID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("verify function access: %w", err)
	}
	if !exists {
		return fmt.Errorf("function not found")
	}
	return nil
}

func (r *MarketplaceRepository) assertFunctionOwned(ctx context.Context, tenantID, functionID string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM registry_functions
			WHERE id = $1::uuid AND tenant_id = $2::uuid
		)`, functionID, tenantID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("verify function ownership: %w", err)
	}
	if !exists {
		return fmt.Errorf("function not found or not owned by tenant")
	}
	return nil
}

func (r *MarketplaceRepository) getFunctionMetrics(ctx context.Context, functionID string) (MonetizationMetrics, error) {
	windowStart := time.Now().UTC().Add(-monetizationMetricsWindow)

	var metrics MonetizationMetrics
	var errorRatePct sql.NullFloat64

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(COUNT(*), 0)::int,
			COALESCE(AVG(duration_ms), 0),
			COALESCE(
				SUM(CASE WHEN outcome IN ('error', 'timeout') THEN 1 ELSE 0 END) * 100.0
				/ NULLIF(COUNT(*), 0),
				0
			),
			COALESCE(COUNT(DISTINCT COALESCE(user_id::text, caller_ip::text)), 0)::int
		FROM registry_function_executions
		WHERE function_id = $1::uuid
		  AND timestamp >= $2`,
		functionID, windowStart,
	).Scan(
		&metrics.ExecutionCount,
		&metrics.AverageLatencyMs,
		&errorRatePct,
		&metrics.UserCount,
	)
	if err != nil {
		return MonetizationMetrics{}, fmt.Errorf("load function metrics: %w", err)
	}

	if errorRatePct.Valid {
		metrics.ErrorRate = errorRatePct.Float64 / 100.0
	}

	// Fall back to popularity when execution history is sparse.
	if metrics.ExecutionCount == 0 {
		var popularity int
		if err := r.db.QueryRowContext(ctx, `
			SELECT COALESCE(popularity_score, 0)
			FROM registry_functions
			WHERE id = $1::uuid`, functionID).Scan(&popularity); err == nil && popularity > 0 {
			metrics.ExecutionCount = popularity
			metrics.UserCount = max(1, popularity/8)
		}
	}

	return metrics, nil
}

func (r *MarketplaceRepository) getCurrentPricing(ctx context.Context, functionID string) (CurrentPricing, error) {
	var (
		pricePerCall sql.NullFloat64
		model        sql.NullString
		subMonthly   sql.NullFloat64
		revShare     sql.NullFloat64
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(rf.price_per_call, 0),
			fl.pricing_model,
			fl.subscription_monthly_usd,
			fl.revenue_share_percent
		FROM registry_functions rf
		LEFT JOIN function_listings fl ON fl.function_id = rf.id AND fl.is_active = TRUE
		WHERE rf.id = $1::uuid`,
		functionID,
	).Scan(&pricePerCall, &model, &subMonthly, &revShare)
	if err != nil {
		return CurrentPricing{}, fmt.Errorf("load current pricing: %w", err)
	}

	current := CurrentPricing{Model: "free", Price: 0}
	if model.Valid && model.String != "" {
		current.Model = model.String
	}

	switch current.Model {
	case "per_call":
		if pricePerCall.Valid {
			current.Price = pricePerCall.Float64
		}
	case "subscription":
		if subMonthly.Valid {
			current.Price = subMonthly.Float64
		}
	case "revenue_share":
		if revShare.Valid {
			current.Price = revShare.Float64
		}
	default:
		if pricePerCall.Valid && pricePerCall.Float64 > 0 {
			current.Model = "per_call"
			current.Price = pricePerCall.Float64
		}
	}

	return current, nil
}

func (r *MarketplaceRepository) upsertFunctionListing(ctx context.Context, functionID, model string, price float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin pricing transaction: %w", err)
	}
	defer tx.Rollback()

	var pricePerCall, subMonthly, revShare sql.NullFloat64
	switch model {
	case "free":
		price = 0
	case "per_call":
		pricePerCall = sql.NullFloat64{Float64: price, Valid: true}
	case "subscription":
		subMonthly = sql.NullFloat64{Float64: price, Valid: true}
	case "revenue_share":
		revShare = sql.NullFloat64{Float64: price, Valid: true}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE function_listings
		SET pricing_model = $2,
		    price_per_call = $3,
		    subscription_monthly_usd = $4,
		    revenue_share_percent = $5,
		    is_active = TRUE,
		    updated_at = NOW()
		WHERE function_id = $1::uuid`,
		functionID, model, pricePerCall, subMonthly, revShare,
	)
	if err != nil {
		return fmt.Errorf("update function listing: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update function listing rows affected: %w", err)
	}
	if rows == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO function_listings (
				function_id, pricing_model, price_per_call,
				subscription_monthly_usd, revenue_share_percent, is_active, updated_at
			) VALUES ($1::uuid, $2, $3, $4, $5, TRUE, NOW())`,
			functionID, model, pricePerCall, subMonthly, revShare,
		)
		if err != nil {
			return fmt.Errorf("insert function listing: %w", err)
		}
	}

	perCallUpdate := 0.0
	if model == "per_call" {
		perCallUpdate = price
	} else if model == "free" {
		perCallUpdate = 0
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE registry_functions
		SET price_per_call = $1, updated_at = NOW()
		WHERE id = $2::uuid`,
		perCallUpdate, functionID,
	)
	if err != nil {
		return fmt.Errorf("update registry pricing: %w", err)
	}

	return tx.Commit()
}

func validPricingModel(model string) bool {
	switch model {
	case "free", "per_call", "subscription", "revenue_share":
		return true
	default:
		return false
	}
}

func generatePricingRecommendations(metrics MonetizationMetrics) []PricingRecommendation {
	executions := float64(max(metrics.ExecutionCount, 1))
	users := float64(max(metrics.UserCount, 1))
	latency := metrics.AverageLatencyMs
	if latency <= 0 {
		latency = 42
	}
	errorRate := metrics.ErrorRate

	perCallPrice := roundPrice(math.Max(0.001, math.Min(0.05, latency/5000.0+0.005)))
	perCallConfidence := clamp01(
		0.55+
			min(executions/5000.0, 0.25)+
			boolScore(latency < 100, 0.08)+
			boolScore(errorRate < 0.05, 0.08)+
			boolScore(executions/users > 5, 0.04),
	)

	subPrice := roundPrice(math.Max(4.99, math.Min(49.99, perCallPrice*executions/users*20)))
	subConfidence := clamp01(
		0.45+
			min(users/2000.0, 0.25)+
			boolScore(users >= 50, 0.08)+
			boolScore(errorRate < 0.08, 0.05),
	)

	revSharePercent := roundPrice(math.Max(10, math.Min(30, 15+executions/users)))
	revConfidence := clamp01(
		0.35+
			boolScore(executions/users < 20, 0.15)+
			boolScore(errorRate < 0.03, 0.08)+
			min(executions/10000.0, 0.12),
	)

	perCallRevenue := executions * perCallPrice * 0.7
	subRevenue := users * 0.1 * subPrice
	revRevenue := executions * 0.001 * revSharePercent

	recommendations := []PricingRecommendation{
		{
			Model:           "per_call",
			Price:           perCallPrice,
			Confidence:      perCallConfidence,
			Reasoning:       perCallReasoning(metrics),
			ExpectedRevenue: roundPrice(perCallRevenue),
		},
		{
			Model:           "subscription",
			Price:           subPrice,
			Confidence:      subConfidence,
			Reasoning:       "Professional users prefer predictable pricing for production use",
			ExpectedRevenue: roundPrice(subRevenue),
		},
		{
			Model:           "revenue_share",
			Price:           revSharePercent,
			Confidence:      revConfidence,
			Reasoning:       "High-value workflows benefit from usage-based revenue share",
			ExpectedRevenue: roundPrice(revRevenue),
		},
	}

	if metrics.ExecutionCount < 25 && metrics.UserCount < 10 {
		freeConfidence := clamp01(0.72 - min(float64(metrics.ExecutionCount)/100.0, 0.2))
		recommendations = append(recommendations, PricingRecommendation{
			Model:           "free",
			Price:           0,
			Confidence:      freeConfidence,
			Reasoning:       "Early-stage functions grow faster with free access while building adoption",
			ExpectedRevenue: 0,
		})
	}

	return recommendations
}

func perCallReasoning(metrics MonetizationMetrics) string {
	switch {
	case metrics.AverageLatencyMs > 0 && metrics.AverageLatencyMs < 100 && metrics.ExecutionCount > 100:
		return "High execution volume with low latency suggests pay-per-call model"
	case metrics.ErrorRate > 0.08:
		return "Per-call pricing limits downside while reliability improves"
	default:
		return "Usage-based pricing aligns cost with value for API-style workloads"
	}
}

func roundPrice(v float64) float64 {
	return math.Round(v*100) / 100
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func boolScore(ok bool, score float64) float64 {
	if ok {
		return score
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
