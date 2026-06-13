package fom

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGoal(ctx context.Context, goal *Goal) error {
	contextJSON, _ := json.Marshal(goal.Context)
	constraintsJSON, _ := json.Marshal(goal.Constraints)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_goals (
			id, tenant_id, user_id, goal_text, goal_type, goal_category,
			context, constraints, user_tier, user_experience_level, user_domain,
			user_goals_history_count, source, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, goal.ID, goal.TenantID, goal.UserID, goal.GoalText, goal.GoalType, goal.GoalCategory,
		contextJSON, constraintsJSON, goal.UserTier, goal.UserExperienceLevel, goal.UserDomain,
		goal.UserGoalsHistoryCount, goal.Source, goal.CreatedAt)

	return err
}

func (r *Repository) GetGoal(ctx context.Context, id uuid.UUID) (*Goal, error) {
	goal := &Goal{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, goal_text, goal_type, goal_category,
			context, constraints, user_tier, user_experience_level, user_domain,
			user_goals_history_count, source, created_at
		FROM fom_goals WHERE id = $1
	`, id).Scan(
		&goal.ID, &goal.TenantID, &goal.UserID, &goal.GoalText, &goal.GoalType, &goal.GoalCategory,
		&goal.Context, &goal.Constraints, &goal.UserTier, &goal.UserExperienceLevel, &goal.UserDomain,
		&goal.UserGoalsHistoryCount, &goal.Source, &goal.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return goal, err
}

func (r *Repository) CreatePlan(ctx context.Context, plan *Plan) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_plans (
			id, goal_id, plan_text, workflow_json, model_used, generation_time_ms,
			confidence, estimated_cost, estimated_time, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, plan.ID, plan.GoalID, plan.PlanText, plan.WorkflowJSON, plan.ModelUsed,
		plan.GenerationTimeMs, plan.Confidence, plan.EstimatedCost, plan.EstimatedTime, plan.CreatedAt)

	return err
}

func (r *Repository) GetPlan(ctx context.Context, id uuid.UUID) (*Plan, error) {
	plan := &Plan{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, goal_id, plan_text, workflow_json, model_used, generation_time_ms,
			confidence, estimated_cost, estimated_time, created_at
		FROM fom_plans WHERE id = $1
	`, id).Scan(
		&plan.ID, &plan.GoalID, &plan.PlanText, &plan.WorkflowJSON, &plan.ModelUsed,
		&plan.GenerationTimeMs, &plan.Confidence, &plan.EstimatedCost, &plan.EstimatedTime, &plan.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return plan, err
}

func (r *Repository) ListPlansByGoal(ctx context.Context, goalID uuid.UUID) ([]*Plan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, goal_id, plan_text, workflow_json, model_used, generation_time_ms,
			confidence, estimated_cost, estimated_time, created_at
		FROM fom_plans WHERE goal_id = $1 ORDER BY created_at DESC
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		plan := &Plan{}
		err := rows.Scan(
			&plan.ID, &plan.GoalID, &plan.PlanText, &plan.WorkflowJSON, &plan.ModelUsed,
			&plan.GenerationTimeMs, &plan.Confidence, &plan.EstimatedCost, &plan.EstimatedTime, &plan.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r *Repository) CreateAction(ctx context.Context, action *Action) error {
	inputJSON, _ := json.Marshal(action.InputSchema)
	outputJSON, _ := json.Marshal(action.OutputSchema)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_actions (
			id, plan_id, function_name, function_id, input_schema, output_schema,
			execution_id, actual_cost, actual_time_ms, success, error_message,
			sequence_order, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, action.ID, action.PlanID, action.FunctionName, action.FunctionID, inputJSON, outputJSON,
		action.ExecutionID, action.ActualCost, action.ActualTimeMs, action.Success,
		action.ErrorMessage, action.SequenceOrder, action.CreatedAt)

	return err
}

func (r *Repository) CreateResult(ctx context.Context, result *Result) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_results (
			id, plan_id, success, outcome_text, total_cost, total_time_ms,
			reliability_score, efficiency_score, speed_score, completeness_score,
			failure_reason, failure_code, failed_action_id, user_rating, user_feedback, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, result.ID, result.PlanID, result.Success, result.OutcomeText, result.TotalCost, result.TotalTimeMs,
		result.ReliabilityScore, result.EfficiencyScore, result.SpeedScore, result.CompletenessScore,
		result.FailureReason, result.FailureCode, result.FailedActionID, result.UserRating, result.UserFeedback, result.CreatedAt)

	return err
}

func (r *Repository) GetResult(ctx context.Context, id uuid.UUID) (*Result, error) {
	result := &Result{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, plan_id, success, outcome_text, total_cost, total_time_ms,
			reliability_score, efficiency_score, speed_score, completeness_score,
			failure_reason, failure_code, failed_action_id, user_rating, user_feedback, created_at
		FROM fom_results WHERE id = $1
	`, id).Scan(
		&result.ID, &result.PlanID, &result.Success, &result.OutcomeText, &result.TotalCost, &result.TotalTimeMs,
		&result.ReliabilityScore, &result.EfficiencyScore, &result.SpeedScore, &result.CompletenessScore,
		&result.FailureReason, &result.FailureCode, &result.FailedActionID, &result.UserRating, &result.UserFeedback, &result.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return result, err
}

func (r *Repository) GetWorkflowPattern(ctx context.Context, patternName, goalType string) (*WorkflowPattern, error) {
	pattern := &WorkflowPattern{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, pattern_name, goal_type, workflow_json, usage_count, success_count, failure_count,
			avg_cost, avg_time_ms, avg_success_rate, first_used_at, last_used_at, created_at
		FROM fom_workflow_patterns WHERE pattern_name = $1 AND goal_type = $2
	`, patternName, goalType).Scan(
		&pattern.ID, &pattern.PatternName, &pattern.GoalType, &pattern.WorkflowJSON, &pattern.UsageCount,
		&pattern.SuccessCount, &pattern.FailureCount, &pattern.AvgCost, &pattern.AvgTimeMs,
		&pattern.AvgSuccessRate, &pattern.FirstUsedAt, &pattern.LastUsedAt, &pattern.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return pattern, err
}

func (r *Repository) ListWorkflowPatterns(ctx context.Context, goalType string, limit int) ([]*WorkflowPattern, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pattern_name, goal_type, workflow_json, usage_count, success_count, failure_count,
			avg_cost, avg_time_ms, avg_success_rate, first_used_at, last_used_at, created_at
		FROM fom_workflow_patterns WHERE goal_type = $1
		ORDER BY usage_count DESC LIMIT $2
	`, goalType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []*WorkflowPattern
	for rows.Next() {
		pattern := &WorkflowPattern{}
		err := rows.Scan(
			&pattern.ID, &pattern.PatternName, &pattern.GoalType, &pattern.WorkflowJSON, &pattern.UsageCount,
			&pattern.SuccessCount, &pattern.FailureCount, &pattern.AvgCost, &pattern.AvgTimeMs,
			&pattern.AvgSuccessRate, &pattern.FirstUsedAt, &pattern.LastUsedAt, &pattern.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, rows.Err()
}

func (r *Repository) GetFunctionStats(ctx context.Context, functionName string) (*FunctionStats, error) {
	stats := &FunctionStats{}
	err := r.db.QueryRowContext(ctx, `
		SELECT function_name, avg_cost, avg_time_ms, success_rate, p50_time_ms, p95_time_ms, p99_time_ms,
			dependencies, sample_count, last_updated
		FROM fom_function_stats WHERE function_name = $1
	`, functionName).Scan(
		&stats.FunctionName, &stats.AvgCost, &stats.AvgTimeMs, &stats.SuccessRate,
		&stats.P50TimeMs, &stats.P95TimeMs, &stats.P99TimeMs, &stats.Dependencies, &stats.SampleCount, &stats.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stats, err
}

func (r *Repository) UpsertFunctionStats(ctx context.Context, stats *FunctionStats) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_function_stats (
			function_name, avg_cost, avg_time_ms, success_rate, p50_time_ms, p95_time_ms, p99_time_ms,
			dependencies, sample_count, last_updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (function_name) DO UPDATE SET
			avg_cost = $2, avg_time_ms = $3, success_rate = $4, p50_time_ms = $5, p95_time_ms = $6,
			p99_time_ms = $7, dependencies = $8, sample_count = $9, last_updated = $10
	`, stats.FunctionName, stats.AvgCost, stats.AvgTimeMs, stats.SuccessRate,
		stats.P50TimeMs, stats.P95TimeMs, stats.P99TimeMs, stats.Dependencies, stats.SampleCount, stats.LastUpdated)

	return err
}

func (r *Repository) ListFunctionStats(ctx context.Context, limit int) ([]*FunctionStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT function_name, avg_cost, avg_time_ms, success_rate, p50_time_ms, p95_time_ms, p99_time_ms,
			dependencies, sample_count, last_updated
		FROM fom_function_stats ORDER BY sample_count DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statsList []*FunctionStats
	for rows.Next() {
		stats := &FunctionStats{}
		err := rows.Scan(
			&stats.FunctionName, &stats.AvgCost, &stats.AvgTimeMs, &stats.SuccessRate,
			&stats.P50TimeMs, &stats.P95TimeMs, &stats.P99TimeMs, &stats.Dependencies, &stats.SampleCount, &stats.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, stats)
	}
	return statsList, rows.Err()
}

func (r *Repository) GetPrivacyBudget(ctx context.Context, tenantID uuid.UUID) (*PrivacyBudget, error) {
	budget := &PrivacyBudget{}
	err := r.db.QueryRowContext(ctx, `
		SELECT tenant_id, total_budget, used_budget, period_start, period_end, created_at, updated_at
		FROM fom_privacy_budget WHERE tenant_id = $1
	`, tenantID).Scan(
		&budget.TenantID, &budget.TotalBudget, &budget.UsedBudget,
		&budget.PeriodStart, &budget.PeriodEnd, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return budget, err
}

func (r *Repository) CreateOrUpdatePrivacyBudget(ctx context.Context, budget *PrivacyBudget) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_privacy_budget (tenant_id, total_budget, used_budget, period_start, period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			total_budget = $2, used_budget = $3, period_start = $4, period_end = $5, updated_at = $7
	`, budget.TenantID, budget.TotalBudget, budget.UsedBudget, budget.PeriodStart, budget.PeriodEnd, budget.CreatedAt, budget.UpdatedAt)

	return err
}

func (r *Repository) ConsumePrivacyBudget(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE fom_privacy_budget
		SET used_budget = used_budget + 1, updated_at = $2
		WHERE tenant_id = $1 AND used_budget < total_budget
	`, tenantID, time.Now().UTC())
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *Repository) ListTrainingRecords(ctx context.Context, goalType, split string, limit int) ([]*TrainingRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, goal_text, goal_type, workflow_json, outcome_success, outcome_score,
			total_cost, total_time_ms, is_synthetic, generation_method, data_source,
			confidence_level, labeled_by, labeling_method, created_at, split
		FROM fom_training_records
		WHERE goal_type = $1 AND split = $2
		ORDER BY created_at DESC LIMIT $3
	`, goalType, split, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*TrainingRecord
	for rows.Next() {
		record := &TrainingRecord{}
		err := rows.Scan(
			&record.ID, &record.GoalText, &record.GoalType, &record.WorkflowJSON,
			&record.OutcomeSuccess, &record.OutcomeScore, &record.TotalCost, &record.TotalTimeMs,
			&record.IsSynthetic, &record.GenerationMethod, &record.DataSource, &record.ConfidenceLevel,
			&record.LabeledBy, &record.LabelingMethod, &record.CreatedAt, &record.Split,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *Repository) CreateTrainingRecord(ctx context.Context, record *TrainingRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fom_training_records (
			id, goal_text, goal_type, workflow_json, outcome_success, outcome_score,
			total_cost, total_time_ms, is_synthetic, generation_method, data_source,
			confidence_level, labeled_by, labeling_method, created_at, split
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, record.ID, record.GoalText, record.GoalType, record.WorkflowJSON, record.OutcomeSuccess, record.OutcomeScore,
		record.TotalCost, record.TotalTimeMs, record.IsSynthetic, record.GenerationMethod, record.DataSource,
		record.ConfidenceLevel, record.LabeledBy, record.LabelingMethod, record.CreatedAt, record.Split)

	return err
}

func (r *Repository) ListWorkflowHints(ctx context.Context, goalType string) ([]*WorkflowHint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, goal_pattern, goal_type, context_conditions, recommended_workflow,
			success_rate, avg_time_savings_ms, constraint_tags, hint_usage_count,
			hint_success_count, created_at, updated_at
		FROM fom_workflow_hints WHERE goal_type = $1 ORDER BY hint_usage_count DESC
	`, goalType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hints []*WorkflowHint
	for rows.Next() {
		hint := &WorkflowHint{}
		err := rows.Scan(
			&hint.ID, &hint.GoalPattern, &hint.GoalType, &hint.ContextConditions, &hint.RecommendedWorkflow,
			&hint.SuccessRate, &hint.AvgTimeSavingsMs, &hint.ConstraintTags, &hint.HintUsageCount,
			&hint.HintSuccessCount, &hint.CreatedAt, &hint.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		hints = append(hints, hint)
	}
	return hints, rows.Err()
}

func (r *Repository) UpdateWorkflowHintStats(ctx context.Context, hintID uuid.UUID, success bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE fom_workflow_hints
		SET hint_usage_count = hint_usage_count + 1,
			hint_success_count = hint_success_count + $2,
			updated_at = $3
		WHERE id = $1
	`, hintID, boolToInt(success), time.Now().UTC())

	return err
}

func (r *Repository) ListFailureTypes(ctx context.Context) ([]*FailureType, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, failure_code, failure_class, recovery_action, parent_failure_id, description, created_at
		FROM fom_failure_types ORDER BY failure_class, failure_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*FailureType
	for rows.Next() {
		ft := &FailureType{}
		err := rows.Scan(
			&ft.ID, &ft.FailureCode, &ft.FailureClass, &ft.RecoveryAction, &ft.ParentFailureID, &ft.Description, &ft.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		types = append(types, ft)
	}
	return types, rows.Err()
}