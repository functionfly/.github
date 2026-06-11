package consciousness

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	*sql.DB
}

func TestEngine_AnalyzeTenant_Timeout(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()
	engine := NewEngine(db, logger)

	tenantID := uuid.New()
	params := AnalysisParams{
		LookbackDays: 7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := engine.AnalyzeTenant(ctx, tenantID, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tenantID, result.TenantID)
	assert.True(t, result.DurationMs >= 0)
}

func TestEngine_AnalyzeTenant_Cancellation(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()
	engine := NewEngine(db, logger)

	tenantID := uuid.New()
	params := AnalysisParams{
		LookbackDays: 7,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := engine.AnalyzeTenant(ctx, tenantID, params)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestEngine_Constants(t *testing.T) {
	assert.Equal(t, 5*time.Minute, DefaultAnalysisTimeout)
	assert.Equal(t, 6*time.Hour, DefaultDedupWindow)
	assert.Equal(t, 7*24*time.Hour, DefaultInsightExpiry)
}

func TestNewEngineWithConfig(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()

	engine := NewEngineWithConfig(db, logger, 5)
	assert.Equal(t, 5, engine.maxConcurrent)

	engine2 := NewEngineWithConfig(db, logger, 20)
	assert.Equal(t, 20, engine2.maxConcurrent)
}

func TestComputeTrend(t *testing.T) {
	tests := []struct {
		previous float64
		current  float64
		expected string
	}{
		{70, 80, "improving"},
		{80, 70, "declining"},
		{75, 76, "stable"},
		{75, 74, "stable"},
		{70, 73.5, "stable"},
		{70, 74, "stable"},
		{70, 73, "stable"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := computeTrend(tt.previous, tt.current)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalysisParams(t *testing.T) {
	params := AnalysisParams{
		LookbackDays:  30,
		FunctionIDs:   []uuid.UUID{uuid.New(), uuid.New()},
		IncludeGraphs: true,
		IncludeAgents: false,
		PlanTier:      "enterprise",
		Since:         time.Now().Add(-24 * time.Hour),
	}

	assert.Equal(t, 30, params.LookbackDays)
	assert.Len(t, params.FunctionIDs, 2)
	assert.True(t, params.IncludeGraphs)
	assert.False(t, params.IncludeAgents)
	assert.Equal(t, "enterprise", params.PlanTier)
}

func TestAnalysisResult(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now()

	result := &AnalysisResult{
		TenantID:   tenantID,
		AnalyzedAt: now,
		Insights:   []*Insight{},
		AnalyzerErrors: map[string]error{
			"cost": assert.AnError,
		},
		DurationMs: 1500,
	}

	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, now, result.AnalyzedAt)
	assert.Empty(t, result.Insights)
	assert.Len(t, result.AnalyzerErrors, 1)
	assert.Equal(t, int64(1500), result.DurationMs)
}

func TestAnalyzerResult(t *testing.T) {
	result := analyzerResult{
		insights: []*Insight{
			{Title: "Test Insight"},
		},
		name: "test_analyzer",
		err:  nil,
	}

	assert.Len(t, result.insights, 1)
	assert.Equal(t, "test_analyzer", result.name)
	assert.NoError(t, result.err)
}