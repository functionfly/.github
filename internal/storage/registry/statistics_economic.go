package registry

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EconomicAnalysisResult holds economic monitoring data.
type EconomicAnalysisResult struct {
	TopRevenueGenerators []RevenueGeneratorRow
	SuspiciousGrowth     []SuspiciousGrowthRow
	ArtificialBoosting   []ArtificialBoostingRow
}

type RevenueGeneratorRow struct {
	FunctionID     uuid.UUID
	FunctionName   string
	Author         string
	Revenue30d     float64
	ExecutionCount int
	GrowthRate     float64
}

type SuspiciousGrowthRow struct {
	FunctionID   uuid.UUID
	FunctionName string
	Author       string
	Pattern      string
	Details      string
	DetectedAt   time.Time
}

type ArtificialBoostingRow struct {
	FunctionID      uuid.UUID
	FunctionName    string
	DetectedPattern string
	Confidence      int
	RelatedAccounts []string
}

// AnalyzeEconomicData performs economic monitoring and analysis.
func (r *RegistryRepository) AnalyzeEconomicData() (*EconomicAnalysisResult, error) {
	result := &EconomicAnalysisResult{}

	// Get top revenue generators
	revenueGenerators, err := r.getTopRevenueGenerators(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue generators: %w", err)
	}
	result.TopRevenueGenerators = revenueGenerators

	// Detect suspicious growth patterns
	suspiciousGrowth, err := r.detectSuspiciousGrowth()
	if err != nil {
		return nil, fmt.Errorf("failed to detect suspicious growth: %w", err)
	}
	result.SuspiciousGrowth = suspiciousGrowth

	// Detect artificial boosting
	artificialBoosting, err := r.detectArtificialBoosting()
	if err != nil {
		return nil, fmt.Errorf("failed to detect artificial boosting: %w", err)
	}
	result.ArtificialBoosting = artificialBoosting

	return result, nil
}

// getTopRevenueGenerators returns functions ranked by revenue in the last 30 days.
func (r *RegistryRepository) getTopRevenueGenerators(limit int) ([]RevenueGeneratorRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		WITH function_revenue AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				COALESCE(f.price_per_call, 0) as price_per_call,
				COUNT(e.id) as execution_count,
				COUNT(e.id) * COALESCE(f.price_per_call, 0) as revenue_30d
			FROM registry_functions f
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY f.id, f.name, f.author, f.price_per_call
		),
		function_growth AS (
			SELECT
				fr.function_id,
				fr.function_name,
				fr.author,
				fr.revenue_30d,
				fr.execution_count,
				-- Calculate growth rate: (current - previous) / previous * 100
				CASE
					WHEN prev.revenue_30d > 0 THEN
						((fr.revenue_30d - prev.revenue_30d) / prev.revenue_30d) * 100
					ELSE 0
				END as growth_rate
			FROM function_revenue fr
			LEFT JOIN (
				-- Previous 30-day period revenue
				SELECT
					f.id as function_id,
					COUNT(e.id) * COALESCE(f.price_per_call, 0) as revenue_30d
				FROM registry_functions f
				LEFT JOIN registry_function_executions e ON e.function_id = f.id
					AND e.timestamp BETWEEN NOW() - INTERVAL '60 days' AND NOW() - INTERVAL '30 days'
				GROUP BY f.id, f.price_per_call
			) prev ON fr.function_id = prev.function_id
			WHERE fr.revenue_30d > 0  -- Only include functions with revenue
			ORDER BY fr.revenue_30d DESC
			LIMIT ?
		)
		SELECT * FROM function_growth
	`

	var rows []RevenueGeneratorRow
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

// detectSuspiciousGrowth identifies functions with anomalous growth patterns.
func (r *RegistryRepository) detectSuspiciousGrowth() ([]SuspiciousGrowthRow, error) {
	var suspicious []SuspiciousGrowthRow

	// Look for functions with sudden spikes in the last 7 days
	query := `
		WITH recent_growth AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				-- Recent 7-day revenue
				COALESCE(SUM(CASE WHEN e.timestamp > NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) * f.price_per_call, 0) as recent_revenue,
				-- Previous 7-day revenue (week before)
				COALESCE(SUM(CASE WHEN e.timestamp BETWEEN NOW() - INTERVAL '14 days' AND NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) * f.price_per_call, 0) as prev_revenue,
				-- Recent execution count
				SUM(CASE WHEN e.timestamp > NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) as recent_executions
			FROM registry_functions f
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '14 days'
			WHERE f.price_per_call > 0
			GROUP BY f.id, f.name, f.author, f.price_per_call
		)
		SELECT
			function_id,
			function_name,
			author,
			recent_revenue,
			prev_revenue,
			recent_executions
		FROM recent_growth
		WHERE recent_revenue > 0 AND prev_revenue > 0
			AND (recent_revenue / NULLIF(prev_revenue, 0)) > 5.0  -- 500% growth
			AND recent_executions > 100  -- Significant volume
		ORDER BY (recent_revenue / NULLIF(prev_revenue, 0)) DESC
		LIMIT 10
	`

	var rows []struct {
		FunctionID       uuid.UUID `json:"function_id"`
		FunctionName     string    `json:"function_name"`
		Author           string    `json:"author"`
		RecentRevenue    float64   `json:"recent_revenue"`
		PrevRevenue      float64   `json:"prev_revenue"`
		RecentExecutions int       `json:"recent_executions"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		growthMultiplier := row.RecentRevenue / row.PrevRevenue
		suspicious = append(suspicious, SuspiciousGrowthRow{
			FunctionID:   row.FunctionID,
			FunctionName: row.FunctionName,
			Author:       row.Author,
			Pattern:      "Sudden revenue spike",
			Details: fmt.Sprintf("%.1fx revenue increase (%.2f -> %.2f) in 7 days with %d executions",
				growthMultiplier, row.PrevRevenue, row.RecentRevenue, row.RecentExecutions),
			DetectedAt: time.Now(),
		})
	}

	return suspicious, nil
}

// detectArtificialBoosting identifies functions that appear to be artificially boosted.
func (r *RegistryRepository) detectArtificialBoosting() ([]ArtificialBoostingRow, error) {
	var artificial []ArtificialBoostingRow

	// Look for functions with coordinated rating patterns that don't match execution volume
	query := `
		WITH function_metrics AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				-- Rating metrics
				COALESCE(rat.total_ratings, 0) as total_ratings,
				COALESCE(rat.overall_score, 0) as overall_score,
				-- Execution metrics
				COUNT(e.id) as execution_count,
				COUNT(DISTINCT e.tenant_id) as unique_tenants,
				COUNT(DISTINCT e.caller_ip) as unique_ips
			FROM registry_functions f
			LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY f.id, f.name, f.author, rat.total_ratings, rat.overall_score
		)
		SELECT
			function_id,
			function_name,
			author,
			total_ratings,
			overall_score,
			execution_count,
			unique_tenants,
			unique_ips
		FROM function_metrics
		WHERE total_ratings > 10  -- Has some ratings
			AND execution_count < 100  -- Low execution volume
			AND unique_tenants < 5  -- Low tenant diversity
			AND overall_score > 4.0  -- High rating score
		ORDER BY overall_score DESC
		LIMIT 5
	`

	var rows []struct {
		FunctionID     uuid.UUID `json:"function_id"`
		FunctionName   string    `json:"function_name"`
		Author         string    `json:"author"`
		TotalRatings   int       `json:"total_ratings"`
		OverallScore   float64   `json:"overall_score"`
		ExecutionCount int       `json:"execution_count"`
		UniqueTenants  int       `json:"unique_tenants"`
		UniqueIPs      int       `json:"unique_ips"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 60
		if row.UniqueTenants == 1 {
			confidence = 95
		} else if row.UniqueTenants <= 3 {
			confidence = 80
		}

		relatedAccounts := []string{row.Author}
		pattern := fmt.Sprintf("High rating (%.1f/5) with low usage (%d executions, %d tenants) - possible artificial boosting",
			row.OverallScore, row.ExecutionCount, row.UniqueTenants)

		artificial = append(artificial, ArtificialBoostingRow{
			FunctionID:      row.FunctionID,
			FunctionName:    row.FunctionName,
			DetectedPattern: pattern,
			Confidence:      confidence,
			RelatedAccounts: relatedAccounts,
		})
	}

	return artificial, nil
}
