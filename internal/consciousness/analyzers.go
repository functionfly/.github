package consciousness

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Analyzer is the interface that all consciousness analyzers implement.
type Analyzer interface {
	Name() string
	Category() InsightCategory
	Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error)
}

// AnalysisParams provides context for an analysis run.
type AnalysisParams struct {
	LookbackDays  int
	FunctionIDs   []uuid.UUID
	IncludeGraphs bool
	IncludeAgents bool
	PlanTier      string
	Since         time.Time
}
