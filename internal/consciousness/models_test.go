package consciousness

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsight_JSONMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected JSONMap
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: JSONMap{},
			wantErr:  false,
		},
		{
			name:     "empty byte slice",
			input:    []byte{},
			expected: JSONMap{},
			wantErr:  false,
		},
		{
			name:     "valid JSON bytes",
			input:    []byte(`{"key": "value", "number": 42}`),
			expected: JSONMap{"key": "value", "number": float64(42)},
			wantErr:  false,
		},
		{
			name:     "valid JSON string",
			input:    `{"key": "value"}`,
			expected: JSONMap{"key": "value"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: JSONMap{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONMap
			err := j.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, j)
			}
		})
	}
}

func TestSeverityWeight(t *testing.T) {
	tests := []struct {
		severity InsightSeverity
		expected int
	}{
		{SeverityCritical, 4},
		{SeverityWarning, 3},
		{SeverityOpportunity, 2},
		{SeverityInfo, 1},
		{InsightSeverity("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			assert.Equal(t, tt.expected, SeverityWeight(tt.severity))
		})
	}
}

func TestScoreLabel(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{95, "excellent"},
		{90, "excellent"},
		{70, "good"},
		{50, "needs_attention"},
		{30, "at_risk"},
		{29, "critical"},
		{0, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, ScoreLabel(tt.score))
		})
	}
}

func TestDefaultScoreWeights(t *testing.T) {
	w := DefaultScoreWeights()

	assert.Equal(t, 0.25, w.Health)
	assert.Equal(t, 0.20, w.Efficiency)
	assert.Equal(t, 0.20, w.Scalability)
	assert.Equal(t, 0.20, w.Reliability)
	assert.Equal(t, 0.15, w.Optimization)

	total := w.Health + w.Efficiency + w.Scalability + w.Reliability + w.Optimization
	assert.Equal(t, 1.0, total)
}

func TestPlanRequestLimit(t *testing.T) {
	tests := []struct {
		plan     string
		expected int
	}{
		{"free", 100_000},
		{"starter", 1_000_000},
		{"professional", 10_000_000},
		{"enterprise", -1},
		{"agent_enterprise", -1},
		{"unknown", 100_000},
	}

	for _, tt := range tests {
		t.Run(tt.plan, func(t *testing.T) {
			assert.Equal(t, tt.expected, planRequestLimit(tt.plan))
		})
	}
}

func TestCategoryEnabled(t *testing.T) {
	enabled := []string{"traffic", "cost", "health"}

	assert.True(t, categoryEnabled("traffic", enabled))
	assert.True(t, categoryEnabled("cost", enabled))
	assert.False(t, categoryEnabled("scaling", enabled))
	assert.False(t, categoryEnabled("redundancy", enabled))
}

func TestSeverityMeetsThreshold(t *testing.T) {
	tests := []struct {
		severity  InsightSeverity
		threshold string
		expected  bool
	}{
		{SeverityInfo, "info", true},
		{SeverityInfo, "warning", false},
		{SeverityWarning, "warning", true},
		{SeverityWarning, "info", true},
		{SeverityCritical, "warning", true},
		{SeverityCritical, "critical", true},
		{SeverityOpportunity, "warning", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity)+"_"+tt.threshold, func(t *testing.T) {
			result := severityMeetsThreshold(tt.severity, tt.threshold)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActionLabel(t *testing.T) {
	tests := []struct {
		action   ActionType
		expected string
	}{
		{ActionMergeFunctions, "Merge Functions"},
		{ActionScaleConfig, "Adjust Scaling"},
		{ActionSwapMarketplace, "Swap to Marketplace"},
		{ActionOptimize, "Optimize"},
		{ActionNone, "None"},
		{ActionType("unknown"), "None"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			assert.Equal(t, tt.expected, actionLabel(tt.action))
		})
	}
}

func TestDerefFloat64(t *testing.T) {
	var nilPtr *float64
	val := 42.0

	assert.Equal(t, 0.0, derefFloat64(nilPtr))
	assert.Equal(t, 42.0, derefFloat64(&val))
}

func TestInsight_TrajectoryConstants(t *testing.T) {
	assert.Equal(t, InsightTrajectory("improving"), TrajectoryImproving)
	assert.Equal(t, InsightTrajectory("stable"), TrajectoryStable)
	assert.Equal(t, InsightTrajectory("degrading"), TrajectoryDegrading)
	assert.Equal(t, InsightTrajectory("critical"), TrajectoryCritical)
}

func TestInsightStatus_Constants(t *testing.T) {
	assert.Equal(t, InsightStatus("active"), StatusActive)
	assert.Equal(t, InsightStatus("dismissed"), StatusDismissed)
	assert.Equal(t, InsightStatus("applied"), StatusApplied)
	assert.Equal(t, InsightStatus("expired"), StatusExpired)
	assert.Equal(t, InsightStatus("superseded"), StatusSuperseded)
}

func TestInsightCategory_Constants(t *testing.T) {
	assert.Equal(t, InsightCategory("traffic"), CategoryTraffic)
	assert.Equal(t, InsightCategory("cost"), CategoryCost)
	assert.Equal(t, InsightCategory("redundancy"), CategoryRedundancy)
	assert.Equal(t, InsightCategory("health"), CategoryHealth)
	assert.Equal(t, InsightCategory("marketplace"), CategoryMarketplace)
	assert.Equal(t, InsightCategory("scaling"), CategoryScaling)
}

func TestInsightSeverity_Constants(t *testing.T) {
	assert.Equal(t, InsightSeverity("info"), SeverityInfo)
	assert.Equal(t, InsightSeverity("warning"), SeverityWarning)
	assert.Equal(t, InsightSeverity("critical"), SeverityCritical)
	assert.Equal(t, InsightSeverity("opportunity"), SeverityOpportunity)
}

func TestInsight_JSON(t *testing.T) {
	insight := &Insight{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Category: CategoryCost,
		Severity: SeverityWarning,
		Title:    "Test Insight",
		Message:  "Test message",
		InsightData: JSONMap{
			"key": "value",
			"num": float64(42),
		},
	}

	data, err := json.Marshal(insight)
	require.NoError(t, err)

	var decoded Insight
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, insight.ID, decoded.ID)
	assert.Equal(t, insight.Category, decoded.Category)
	assert.Equal(t, insight.Severity, decoded.Severity)
	assert.Equal(t, insight.Title, decoded.Title)
}

func TestDefaultPreferences(t *testing.T) {
	tenantID := uuid.New()
	prefs := DefaultPreferences(tenantID)

	assert.Equal(t, tenantID, prefs.TenantID)
	assert.True(t, prefs.EmailEnabled)
	assert.False(t, prefs.SlackEnabled)
	assert.True(t, prefs.InAppEnabled)
	assert.False(t, prefs.WebhookEnabled)
	assert.Equal(t, "daily", prefs.DigestFrequency)
	assert.Equal(t, "UTC", prefs.Timezone)
	assert.Contains(t, prefs.EnabledCategories, "traffic")
	assert.Contains(t, prefs.EnabledCategories, "cost")
	assert.Contains(t, prefs.EnabledCategories, "redundancy")
	assert.Contains(t, prefs.EnabledCategories, "health")
	assert.Contains(t, prefs.EnabledCategories, "marketplace")
	assert.Contains(t, prefs.EnabledCategories, "scaling")
	assert.Equal(t, "warning", prefs.MinNotifySeverity)
	assert.False(t, prefs.AutoApplyEnabled)
	assert.Empty(t, prefs.AutoApplyCategories)
}

func TestListInsightsParams_Defaults(t *testing.T) {
	params := ListInsightsParams{
		TenantID: uuid.New(),
	}

	assert.Nil(t, params.Category)
	assert.Nil(t, params.Severity)
	assert.Nil(t, params.Status)
	assert.Equal(t, 0, params.Limit)
	assert.Equal(t, 0, params.Offset)
}
