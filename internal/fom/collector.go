package fom

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Collector struct {
	db              *sql.DB
	redis           *redis.Client
	buffer          []*FOMRecord
	bufferSize      int
	flushInterval   time.Duration
	mu              sync.Mutex
	stopCh          chan struct{}
	streamingExporter StreamingExporter
}

type StreamingExporter interface {
	OnFOMRecord(record map[string]interface{})
}

type FOMRecord struct {
	GoalID      uuid.UUID
	PlanID      uuid.UUID
	ActionID    uuid.UUID
	ResultID    uuid.UUID

	GoalText    string
	GoalType    string
	Workflow    []string

	Success     bool
	TotalCost   float64
	TotalTimeMs int
	OutcomeScore int

	IsSynthetic bool
	CreatedAt   time.Time
}

type ExecutionContext struct {
	TenantID            uuid.UUID
	UserID              uuid.UUID
	GoalText            string
	GoalType            string
	GoalCategory        string
	Context             map[string]interface{}
	Constraints         map[string]interface{}
	UserTier            string
	UserExperienceLevel string
	UserDomain          string
	Source              GoalSource
}

func NewCollector(db *sql.DB, redis *redis.Client, bufferSize int, flushInterval time.Duration) *Collector {
	c := &Collector{
		db:            db,
		redis:         redis,
		buffer:        make([]*FOMRecord, 0, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
	}
	go c.backgroundFlush()
	return c
}

func (c *Collector) SetStreamingExporter(exporter StreamingExporter) {
	c.streamingExporter = exporter
}

func (c *Collector) backgroundFlush() {
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Flush(context.Background()); err != nil {
				log.Error("FOM flush error", "error", err)
			}
		case <-c.stopCh:
			return
		}
	}
}

func (c *Collector) Stop() {
	close(c.stopCh)
	if err := c.Flush(context.Background()); err != nil {
		log.Error("FOM final flush error", "error", err)
	}
}

func (c *Collector) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) == 0 {
		return nil
	}

	for _, record := range c.buffer {
		if err := c.writeRecord(ctx, record); err != nil {
			log.Error("FOM write error", "error", err, "goal_id", record.GoalID)
		}
	}
	c.buffer = c.buffer[:0]
	return nil
}

func (c *Collector) writeRecord(ctx context.Context, record *FOMRecord) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	workflowJSON, err := json.Marshal(record.Workflow)
	if err != nil {
		return err
	}

	resultID := uuid.New()
	if record.ResultID != uuid.Nil {
		resultID = record.ResultID
	}

	var reliabilityScore, efficiencyScore, speedScore, completenessScore int
	if record.Success {
		reliabilityScore = 100
		completenessScore = 100
		efficiencyScore = calculateEfficiencyScore(record.TotalCost)
		speedScore = calculateSpeedScore(record.TotalTimeMs)
	} else {
		reliabilityScore = 0
		completenessScore = 0
		efficiencyScore = 0
		speedScore = 0
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO fom_results (
			id, plan_id, success, total_cost, total_time_ms,
			reliability_score, efficiency_score, speed_score, completeness_score,
			outcome_text, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, resultID, record.PlanID, record.Success, record.TotalCost, record.TotalTimeMs,
		reliabilityScore, efficiencyScore, speedScore, completenessScore,
		formatOutcomeText(record), record.CreatedAt)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO fom_training_records (
			id, goal_text, goal_type, workflow_json, outcome_success, outcome_score,
			total_cost, total_time_ms, is_synthetic, data_source, created_at, split
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, uuid.New(), record.GoalText, record.GoalType, workflowJSON, record.Success,
		record.OutcomeScore, record.TotalCost, record.TotalTimeMs, record.IsSynthetic,
		"production", record.CreatedAt, "train")
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	if c.streamingExporter != nil {
		c.streamingExporter.OnFOMRecord(map[string]interface{}{
			"goal_text":       record.GoalText,
			"goal_type":       record.GoalType,
			"workflow":        record.Workflow,
			"outcome_success": record.Success,
			"outcome_score":   record.OutcomeScore,
			"total_cost":      record.TotalCost,
			"total_time_ms":   record.TotalTimeMs,
			"timestamp":       time.Now().UTC(),
		})
	}

	return nil
}

func calculateEfficiencyScore(cost float64) int {
	if cost <= 0 {
		return 100
	}
	if cost < 0.01 {
		return 100
	}
	if cost < 0.10 {
		return 90
	}
	if cost < 0.50 {
		return 70
	}
	if cost < 1.00 {
		return 50
	}
	return 20
}

func calculateSpeedScore(timeMs int) int {
	if timeMs < 1000 {
		return 100
	}
	if timeMs < 5000 {
		return 90
	}
	if timeMs < 10000 {
		return 70
	}
	if timeMs < 30000 {
		return 50
	}
	return 20
}

func formatOutcomeText(record *FOMRecord) string {
	if record.Success {
		return "Workflow completed successfully"
	}
	return "Workflow failed"
}

func (c *Collector) CollectExecution(ctx context.Context, record *FOMRecord) error {
	record.CreatedAt = time.Now().UTC()

	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case c.buffer <- record:
		if len(c.buffer) >= c.bufferSize {
			go func() {
				flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				c.Flush(flushCtx)
			}()
		}
		return nil
	default:
		return c.writeRecord(ctx, record)
	}
}

func (c *Collector) CollectGoal(ctx context.Context, goal *Goal) error {
	contextJSON, _ := json.Marshal(goal.Context)
	constraintsJSON, _ := json.Marshal(goal.Constraints)

	_, err := c.db.ExecContext(ctx, `
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

func (c *Collector) CollectPlan(ctx context.Context, plan *Plan) error {
	_, err := c.db.ExecContext(ctx, `
		INSERT INTO fom_plans (
			id, goal_id, plan_text, workflow_json, model_used, generation_time_ms,
			confidence, estimated_cost, estimated_time, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, plan.ID, plan.GoalID, plan.PlanText, plan.WorkflowJSON, plan.ModelUsed,
		plan.GenerationTimeMs, plan.Confidence, plan.EstimatedCost, plan.EstimatedTime, plan.CreatedAt)

	return err
}

func (c *Collector) CollectAction(ctx context.Context, action *Action) error {
	inputJSON, _ := json.Marshal(action.InputSchema)
	outputJSON, _ := json.Marshal(action.OutputSchema)

	_, err := c.db.ExecContext(ctx, `
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

func (c *Collector) RecordEvent(ctx context.Context, event *FOMEvent) error {
	_, err := c.db.ExecContext(ctx, `
		INSERT INTO fom_events (id, execution_id, plan_id, event_type, timestamp, payload, sequence_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.ExecutionID, event.PlanID, event.EventType, event.Timestamp, event.Payload, event.SequenceOrder)

	return err
}

func (c *Collector) UpdateWorkflowPattern(ctx context.Context, goalType string, workflow []string, success bool) error {
	workflowJSON, err := json.Marshal(workflow)
	if err != nil {
		return err
	}

	patternName := generatePatternName(workflow)
	now := time.Now().UTC()

	_, err = c.db.ExecContext(ctx, `
		INSERT INTO fom_workflow_patterns (
			pattern_name, goal_type, workflow_json, usage_count, success_count, failure_count,
			first_used_at, last_used_at, created_at
		) VALUES ($1, $2, $3, 1, $4, $5, $6, $6, $6)
		ON CONFLICT (pattern_name, goal_type) DO UPDATE SET
			usage_count = fom_workflow_patterns.usage_count + 1,
			success_count = fom_workflow_patterns.success_count + $4,
			failure_count = fom_workflow_patterns.failure_count + $5,
			last_used_at = $6
	`, patternName, goalType, workflowJSON, boolToInt(success), boolToInt(!success), now)

	return err
}

func generatePatternName(workflow []string) string {
	if len(workflow) == 0 {
		return "empty"
	}
	return workflow[0] + "_" + workflow[len(workflow)-1]
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}