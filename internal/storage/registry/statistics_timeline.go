package registry

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ExecutionTimelineBucket is one time bucket for the execution timeline (conversation overlay).
type ExecutionTimelineBucket struct {
	Bucket      string  `json:"bucket"`       // date (YYYY-MM-DD)
	Value       float64 `json:"value"`        // metric value (e.g. avg latency ms, error %, trust %)
	SampleCount int     `json:"sample_count"` // number of executions in the bucket
}

// GetExecutionTimelineBuckets returns time-bucketed execution metrics for a function.
// metric must be "latency" (avg duration_ms), "errors" (error %), or "trust" (verification %).
// from/to are inclusive; buckets are daily in UTC.
func (r *RegistryRepository) GetExecutionTimelineBuckets(functionID uuid.UUID, from, to time.Time, metric string) ([]ExecutionTimelineBucket, error) {
	dayExpr := "to_char(timestamp AT TIME ZONE 'UTC', 'YYYY-MM-DD')"
	var valueExpr string
	switch metric {
	case "latency":
		valueExpr = "COALESCE(AVG(duration_ms), 0)"
	case "errors":
		valueExpr = "COALESCE(SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0)"
	case "trust":
		valueExpr = "COALESCE(SUM(CASE WHEN verification_status = 'verified' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0)"
	default:
		valueExpr = "COALESCE(AVG(duration_ms), 0)"
	}
	query := fmt.Sprintf(`
		SELECT %s AS bucket, %s AS value, COUNT(*)::INT AS sample_count
		FROM registry_function_executions
		WHERE function_id = $1 AND timestamp >= $2 AND timestamp < $3
		GROUP BY %s
		ORDER BY bucket
	`, dayExpr, valueExpr, dayExpr)

	var rows []struct {
		Bucket      string  `gorm:"column:bucket"`
		Value       float64 `gorm:"column:value"`
		SampleCount int     `gorm:"column:sample_count"`
	}
	// to is exclusive for the range so we add one day
	toExclusive := to.AddDate(0, 0, 1)
	if err := r.db.Raw(query, functionID, from, toExclusive).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get execution timeline: %w", err)
	}
	out := make([]ExecutionTimelineBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, ExecutionTimelineBucket{
			Bucket:      row.Bucket,
			Value:       row.Value,
			SampleCount: row.SampleCount,
		})
	}
	return out, nil
}
