package physics

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type HintEngine struct {
	repo   HintRepository
	cache  *redis.Client
}

type HintRepository interface {
	ListWorkflowHints(ctx context.Context, goalType string) ([]*WorkflowHint, error)
	UpdateWorkflowHintStats(ctx context.Context, hintID uuid.UUID, success bool) error
}

func NewHintEngine(repo HintRepository, cache *redis.Client) *HintEngine {
	return &HintEngine{
		repo:  repo,
		cache: cache,
	}
}

type WorkflowHint struct {
	ID                  uuid.UUID       `json:"id"`
	GoalPattern         string          `json:"goal_pattern"`
	GoalType            string          `json:"goal_type"`
	ContextConditions   json.RawMessage `json:"context_conditions,omitempty"`
	RecommendedWorkflow json.RawMessage `json:"recommended_workflow"`
	SuccessRate         *float64        `json:"success_rate,omitempty"`
	AvgTimeSavingsMs    *int            `json:"avg_time_savings_ms,omitempty"`
	ConstraintTags      []string        `json:"constraint_tags"`
	HintUsageCount      int             `json:"hint_usage_count"`
	HintSuccessCount    int             `json:"hint_success_count"`
	CreatedAt           interface{}     `json:"created_at"`
	UpdatedAt           interface{}     `json:"updated_at"`
}

type ExecutionContext struct {
	UserTier          string
	Constraint        string
	IsSpeedCritical   bool
	IsReliabilityCritical bool
	IsBudgetSensitive bool
}

type Goal struct {
	GoalType string `json:"goal_type"`
}

func (h *HintEngine) ResolveHint(goal *Goal, execCtx *ExecutionContext) ([]string, error) {
	hints, err := h.repo.ListWorkflowHints(context.Background(), goal.GoalType)
	if err != nil {
		return nil, err
	}

	matching := h.filterByContext(hints, execCtx)

	sort.Slice(matching, func(i, j int) bool {
		return h.scoreHint(matching[i]) > h.scoreHint(matching[j])
	})

	if len(matching) > 0 {
		workflow := h.getRecommendedWorkflow(matching[0])
		return workflow, nil
	}

	return nil, nil
}

func (h *HintEngine) filterByContext(hints []*WorkflowHint, ctx *ExecutionContext) []*WorkflowHint {
	matching := make([]*WorkflowHint, 0, len(hints))

	for _, hint := range hints {
		if h.matchesContext(hint, ctx) {
			matching = append(matching, hint)
		}
	}

	return matching
}

func (h *HintEngine) matchesContext(hint *WorkflowHint, ctx *ExecutionContext) bool {
	if hint.ContextConditions == nil {
		return true
	}

	var conditions map[string]interface{}
	if err := json.Unmarshal(hint.ContextConditions, &conditions); err != nil {
		return true
	}

	if userTier, ok := conditions["user_tier"]; ok {
		if ctx.UserTier != "" && ctx.UserTier != userTier.(string) {
			return false
		}
	}

	if constraint, ok := conditions["constraint"]; ok {
		if ctx.Constraint != "" && ctx.Constraint != constraint.(string) {
			return false
		}
	}

	if speedCritical, ok := conditions["speed_critical"].(bool); ok {
		if ctx.IsSpeedCritical != speedCritical {
			return false
		}
	}

	if reliabilityCritical, ok := conditions["reliability_critical"].(bool); ok {
		if ctx.IsReliabilityCritical != reliabilityCritical {
			return false
		}
	}

	if budgetSensitive, ok := conditions["budget_sensitive"].(bool); ok {
		if ctx.IsBudgetSensitive != budgetSensitive {
			return false
		}
	}

	return true
}

func (h *HintEngine) scoreHint(hint *WorkflowHint) float64 {
	successRate := 0.95
	if hint.SuccessRate != nil {
		successRate = *hint.SuccessRate
	}

	timeSavings := 0.0
	if hint.AvgTimeSavingsMs != nil {
		timeSavings = float64(*hint.AvgTimeSavingsMs) / 1000.0
	}

	return successRate*0.6 + timeSavings*0.4
}

func (h *HintEngine) getRecommendedWorkflow(hint *WorkflowHint) []string {
	var workflow []string
	if err := json.Unmarshal(hint.RecommendedWorkflow, &workflow); err != nil {
		return nil
	}
	return workflow
}

func (h *HintEngine) RecordHintUsage(hintID uuid.UUID, success bool) error {
	return h.repo.UpdateWorkflowHintStats(context.Background(), hintID, success)
}