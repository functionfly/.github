package fom

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FailureClass string

const (
	FailureClassPrerequisite FailureClass = "prerequisite"
	FailureClassExecution    FailureClass = "execution"
	FailureClassResource     FailureClass = "resource"
	FailureClassAuth         FailureClass = "auth"
)

type FailureType struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	FailureCode    string         `json:"failure_code" db:"failure_code"`
	FailureClass   FailureClass   `json:"failure_class" db:"failure_class"`
	RecoveryAction string         `json:"recovery_action" db:"recovery_action"`
	ParentFailureID *uuid.UUID    `json:"parent_failure_id,omitempty" db:"parent_failure_id"`
	Description    string         `json:"description" db:"description"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

type GoalSource string

const (
	GoalSourceUser      GoalSource = "user"
	GoalSourceAgent     GoalSource = "agent"
	GoalSourceScheduled GoalSource = "scheduled"
	GoalSourceWebhook   GoalSource = "webhook"
)

type Goal struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	TenantID             uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	UserID               uuid.UUID       `json:"user_id" db:"user_id"`
	GoalText             string          `json:"goal_text" db:"goal_text"`
	GoalType             string          `json:"goal_type" db:"goal_type"`
	GoalCategory         string          `json:"goal_category" db:"goal_category"`
	Context              json.RawMessage `json:"context,omitempty" db:"context"`
	Constraints          json.RawMessage `json:"constraints,omitempty" db:"constraints"`
	UserTier             string          `json:"user_tier" db:"user_tier"`
	UserExperienceLevel  string          `json:"user_experience_level" db:"user_experience_level"`
	UserDomain           string          `json:"user_domain,omitempty" db:"user_domain"`
	UserGoalsHistoryCount int            `json:"user_goals_history_count" db:"user_goals_history_count"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	Source               GoalSource      `json:"source" db:"source"`
}

type Plan struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	GoalID          uuid.UUID       `json:"goal_id" db:"goal_id"`
	PlanText        string          `json:"plan_text" db:"plan_text"`
	WorkflowJSON    json.RawMessage `json:"workflow_json" db:"workflow_json"`
	ModelUsed       string          `json:"model_used,omitempty" db:"model_used"`
	GenerationTimeMs int            `json:"generation_time_ms,omitempty" db:"generation_time_ms"`
	Confidence      float64         `json:"confidence,omitempty" db:"confidence"`
	EstimatedCost   float64         `json:"estimated_cost,omitempty" db:"estimated_cost"`
	EstimatedTime   int             `json:"estimated_time,omitempty" db:"estimated_time"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

func (p *Plan) GetWorkflow() []string {
	var workflow []string
	if err := json.Unmarshal(p.WorkflowJSON, &workflow); err != nil {
		return nil
	}
	return workflow
}

type Action struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	PlanID        uuid.UUID       `json:"plan_id" db:"plan_id"`
	FunctionName  string          `json:"function_name" db:"function_name"`
	FunctionID    *uuid.UUID      `json:"function_id,omitempty" db:"function_id"`
	InputSchema   json.RawMessage `json:"input_schema,omitempty" db:"input_schema"`
	OutputSchema  json.RawMessage `json:"output_schema,omitempty" db:"output_schema"`
	ExecutionID   *uuid.UUID      `json:"execution_id,omitempty" db:"execution_id"`
	ActualCost    *float64        `json:"actual_cost,omitempty" db:"actual_cost"`
	ActualTimeMs  *int            `json:"actual_time_ms,omitempty" db:"actual_time_ms"`
	Success       *bool           `json:"success,omitempty" db:"success"`
	ErrorMessage  string          `json:"error_message,omitempty" db:"error_message"`
	SequenceOrder int             `json:"sequence_order" db:"sequence_order"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

type Result struct {
	ID                 uuid.UUID   `json:"id" db:"id"`
	PlanID             uuid.UUID   `json:"plan_id" db:"plan_id"`
	Success            bool        `json:"success" db:"success"`
	OutcomeText        string      `json:"outcome_text,omitempty" db:"outcome_text"`
	TotalCost          float64     `json:"total_cost" db:"total_cost"`
	TotalTimeMs        int         `json:"total_time_ms" db:"total_time_ms"`
	ReliabilityScore   int         `json:"reliability_score" db:"reliability_score"`
	EfficiencyScore    int         `json:"efficiency_score" db:"efficiency_score"`
	SpeedScore         int         `json:"speed_score" db:"speed_score"`
	CompletenessScore  int         `json:"completeness_score" db:"completeness_score"`
	FailureReason      string      `json:"failure_reason,omitempty" db:"failure_reason"`
	FailureCode        string      `json:"failure_code,omitempty" db:"failure_code"`
	FailedActionID     *uuid.UUID  `json:"failed_action_id,omitempty" db:"failed_action_id"`
	UserRating         *int        `json:"user_rating,omitempty" db:"user_rating"`
	UserFeedback       string      `json:"user_feedback,omitempty" db:"user_feedback"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
}

type WorkflowPattern struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PatternName     string          `json:"pattern_name" db:"pattern_name"`
	GoalType        string          `json:"goal_type" db:"goal_type"`
	WorkflowJSON    json.RawMessage `json:"workflow_json" db:"workflow_json"`
	UsageCount      int             `json:"usage_count" db:"usage_count"`
	SuccessCount    int             `json:"success_count" db:"success_count"`
	FailureCount    int             `json:"failure_count" db:"failure_count"`
	AvgCost         *float64        `json:"avg_cost,omitempty" db:"avg_cost"`
	AvgTimeMs       *int            `json:"avg_time_ms,omitempty" db:"avg_time_ms"`
	AvgSuccessRate  *float64        `json:"avg_success_rate,omitempty" db:"avg_success_rate"`
	FirstUsedAt     *time.Time      `json:"first_used_at,omitempty" db:"first_used_at"`
	LastUsedAt      *time.Time      `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type FunctionStats struct {
	FunctionName string    `json:"function_name" db:"function_name"`
	AvgCost      float64   `json:"avg_cost" db:"avg_cost"`
	AvgTimeMs    int       `json:"avg_time_ms" db:"avg_time_ms"`
	SuccessRate  float64   `json:"success_rate" db:"success_rate"`
	P50TimeMs    *int      `json:"p50_time_ms,omitempty" db:"p50_time_ms"`
	P95TimeMs    *int      `json:"p95_time_ms,omitempty" db:"p95_time_ms"`
	P99TimeMs    *int      `json:"p99_time_ms,omitempty" db:"p99_time_ms"`
	Dependencies []string  `json:"dependencies" db:"dependencies"`
	SampleCount  int       `json:"sample_count" db:"sample_count"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
}

type TrainingRecord struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	GoalText        string          `json:"goal_text" db:"goal_text"`
	GoalType        string          `json:"goal_type" db:"goal_type"`
	WorkflowJSON    json.RawMessage `json:"workflow_json" db:"workflow_json"`
	OutcomeSuccess  bool            `json:"outcome_success" db:"outcome_success"`
	OutcomeScore    int             `json:"outcome_score" db:"outcome_score"`
	TotalCost       *float64        `json:"total_cost,omitempty" db:"total_cost"`
	TotalTimeMs     *int            `json:"total_time_ms,omitempty" db:"total_time_ms"`
	IsSynthetic     bool            `json:"is_synthetic" db:"is_synthetic"`
	GenerationMethod string         `json:"generation_method,omitempty" db:"generation_method"`
	DataSource      string          `json:"data_source" db:"data_source"`
	ConfidenceLevel string          `json:"confidence_level" db:"confidence_level"`
	LabeledBy       *uuid.UUID      `json:"labeled_by,omitempty" db:"labeled_by"`
	LabelingMethod  string          `json:"labeling_method,omitempty" db:"labeling_method"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	Split           string          `json:"split" db:"split"`
}

type EventType string

const (
	EventGoalCreated        EventType = "goal_created"
	EventPlanGenerated      EventType = "plan_generated"
	EventActionStarted      EventType = "action_started"
	EventActionCompleted    EventType = "action_completed"
	EventWorkflowPartial    EventType = "workflow_partial"
	EventWorkflowDone       EventType = "workflow_completed"
	EventWorkflowFailed     EventType = "workflow_failed"
)

type FOMEvent struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	ExecutionID   uuid.UUID       `json:"execution_id" db:"execution_id"`
	PlanID        *uuid.UUID      `json:"plan_id,omitempty" db:"plan_id"`
	EventType     EventType       `json:"event_type" db:"event_type"`
	Timestamp     time.Time       `json:"timestamp" db:"timestamp"`
	Payload       json.RawMessage `json:"payload" db:"payload"`
	SequenceOrder int             `json:"sequence_order" db:"sequence_order"`
}

type WorkflowHint struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	GoalPattern         string          `json:"goal_pattern" db:"goal_pattern"`
	GoalType            string          `json:"goal_type" db:"goal_type"`
	ContextConditions   json.RawMessage `json:"context_conditions,omitempty" db:"context_conditions"`
	RecommendedWorkflow json.RawMessage `json:"recommended_workflow" db:"recommended_workflow"`
	SuccessRate         *float64        `json:"success_rate,omitempty" db:"success_rate"`
	AvgTimeSavingsMs    *int            `json:"avg_time_savings_ms,omitempty" db:"avg_time_savings_ms"`
	ConstraintTags      []string        `json:"constraint_tags" db:"constraint_tags"`
	HintUsageCount      int             `json:"hint_usage_count" db:"hint_usage_count"`
	HintSuccessCount    int             `json:"hint_success_count" db:"hint_success_count"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

func (h *WorkflowHint) GetRecommendedWorkflow() []string {
	var workflow []string
	if err := json.Unmarshal(h.RecommendedWorkflow, &workflow); err != nil {
		return nil
	}
	return workflow
}

type PrivacyBudget struct {
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	TotalBudget int       `json:"total_budget" db:"total_budget"`
	UsedBudget  int       `json:"used_budget" db:"used_budget"`
	PeriodStart time.Time `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time `json:"period_end" db:"period_end"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

func (pb *PrivacyBudget) Remaining() int {
	return pb.TotalBudget - pb.UsedBudget
}

func (pb *PrivacyBudget) IsExhausted() bool {
	return pb.UsedBudget >= pb.TotalBudget
}

type UserTier string

const (
	UserTierFree       UserTier = "free"
	UserTierPro        UserTier = "pro"
	UserTierEnterprise UserTier = "enterprise"
)

func GetDefaultBudget(tier UserTier) int {
	switch tier {
	case UserTierEnterprise:
		return 100000
	case UserTierPro:
		return 10000
	default:
		return 1000
	}
}