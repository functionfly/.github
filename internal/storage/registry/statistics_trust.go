package registry

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrustDistribution holds counts of functions per trust bucket (DB trust_score 0-1: excellent >= 0.8, good >= 0.6, fair >= 0.4, poor < 0.4).
type TrustDistribution struct {
	Excellent int
	Good      int
	Fair      int
	Poor      int
}

// GetTrustDistribution returns the count of registry functions in each trust bucket.
// Uses reliability_score only (base schema: 0-100 scale, normalized to 0-1 for buckets). Run migration 000036 for trust_score.
func (r *RegistryRepository) GetTrustDistribution() (TrustDistribution, error) {
	var out TrustDistribution
	query := `
		SELECT
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.8) AS excellent,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.6 AND COALESCE(rat.reliability_score, 0) / 100.0 < 0.8) AS good,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.4 AND COALESCE(rat.reliability_score, 0) / 100.0 < 0.6) AS fair,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 < 0.4) AS poor
		FROM registry_functions f
		LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
	`
	if err := r.db.Raw(query).Scan(&out).Error; err != nil {
		return TrustDistribution{}, fmt.Errorf("failed to get trust distribution: %w", err)
	}
	return out, nil
}

// HighRiskFunctionRow is one row for admin trust dashboard high-risk list (low trust + risk metrics).
type HighRiskFunctionRow struct {
	FunctionID      uuid.UUID
	Author          string
	Name            string
	TrustScore      float64
	ErrorRate       float64
	TimeoutRate     float64
	TenantDiversity int
	TrustUpdatedAt  *time.Time
}

// GetHighRiskFunctions returns functions with low trust score (e.g. < 0.35), ordered by trust ascending, limit n.
// Uses reliability_score only (0-100 → 0-1) so base schema works; run migration 000036 for trust_score/error_rate/timeout_rate/tenant_diversity.
func (r *RegistryRepository) GetHighRiskFunctions(limit int) ([]HighRiskFunctionRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []HighRiskFunctionRow
	query := `
		SELECT
			f.id AS function_id,
			f.author,
			f.name,
			COALESCE(rat.reliability_score, 0) / 100.0 AS trust_score,
			0::float AS error_rate,
			0::float AS timeout_rate,
			0::integer AS tenant_diversity,
			rat.updated_at AS trust_updated_at
		FROM registry_functions f
		LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
		WHERE COALESCE(rat.reliability_score, 0) / 100.0 < 0.35
		ORDER BY COALESCE(rat.reliability_score, 0) ASC NULLS LAST, f.created_at ASC
		LIMIT ?
	`
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get high-risk functions: %w", err)
	}
	return rows, nil
}
