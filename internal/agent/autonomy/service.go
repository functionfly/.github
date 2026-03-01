package autonomy

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Service handles agent autonomy and scheduled execution
type Service struct {
	db *gorm.DB
}

// NewService creates a new autonomy service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateSchedule creates a new autonomy schedule
func (s *Service) CreateSchedule(ctx context.Context, schedule *identity.AutonomySchedule) error {
	schedule.ID = uuid.New()
	schedule.IsActive = true
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	// Calculate next run time
	if schedule.ScheduleType == identity.AutonomyScheduleRecurring && schedule.CronExpression != nil {
		nextRun, err := s.calculateNextRun(*schedule.CronExpression)
		if err != nil {
			return err
		}
		schedule.NextRunAt = &nextRun
	} else if schedule.ScheduleType == identity.AutonomyScheduleOneTime && schedule.NextRunAt == nil {
		return fmt.Errorf("one_time schedule requires next_run_at")
	}

	return s.db.WithContext(ctx).Create(schedule).Error
}

// GetSchedules gets all schedules for an agent
func (s *Service) GetSchedules(ctx context.Context, agentID string) ([]identity.AutonomySchedule, error) {
	var schedules []identity.AutonomySchedule
	err := s.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&schedules).Error
	return schedules, err
}

// GetActiveSchedules gets all active schedules that are due
func (s *Service) GetActiveSchedules(ctx context.Context) ([]identity.AutonomySchedule, error) {
	var schedules []identity.AutonomySchedule
	err := s.db.WithContext(ctx).
		Where("is_active = ? AND next_run_at <= ?", true, time.Now()).
		Order("next_run_at ASC").
		Find(&schedules).Error
	return schedules, err
}

// ExecuteSchedule executes a schedule immediately
func (s *Service) ExecuteSchedule(ctx context.Context, scheduleID uuid.UUID) (map[string]any, error) {
	var schedule identity.AutonomySchedule
	if err := s.db.WithContext(ctx).Where("id = ?", scheduleID).First(&schedule).Error; err != nil {
		return nil, err
	}

	if !schedule.IsActive {
		return nil, fmt.Errorf("schedule is not active")
	}

	// Execute based on action type
	var result map[string]any
	switch schedule.ActionType {
	case identity.AutonomyActionExecuteFunction:
		result = s.executeFunctionAction(ctx, schedule)
	case identity.AutonomyActionSpawnAgent:
		result = s.spawnAgentAction(ctx, schedule)
	case identity.AutonomyActionSendMessage:
		result = s.sendMessageAction(ctx, schedule)
	case identity.AutonomyActionUpdateState:
		result = s.updateStateAction(ctx, schedule)
	case identity.AutonomyActionEvolve:
		result = s.evolveAction(ctx, schedule)
	default:
		return nil, fmt.Errorf("unknown action type: %s", schedule.ActionType)
	}

	// Update last run time
	now := time.Now()
	schedule.LastRunAt = &now

	// Calculate next run
	if schedule.ScheduleType == identity.AutonomyScheduleRecurring && schedule.CronExpression != nil {
		nextRun, _ := s.calculateNextRun(*schedule.CronExpression)
		schedule.NextRunAt = &nextRun
	} else if schedule.ScheduleType == identity.AutonomyScheduleOneTime {
		schedule.IsActive = false
	}

	schedule.UpdatedAt = now
	if err := s.db.WithContext(ctx).Save(&schedule).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// DeactivateSchedule deactivates a schedule
func (s *Service) DeactivateSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&identity.AutonomySchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]any{
			"is_active":  false,
			"updated_at": time.Now(),
		})
	return result.Error
}

// ActivateSchedule activates a schedule
func (s *Service) ActivateSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&identity.AutonomySchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]any{
			"is_active":  true,
			"updated_at": time.Now(),
		})
	return result.Error
}

// DeleteSchedule deletes a schedule
func (s *Service) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", scheduleID).Delete(&identity.AutonomySchedule{}).Error
}

// ProcessTriggeredSchedules processes schedules triggered by events
func (s *Service) ProcessTriggeredSchedules(ctx context.Context, eventType string, eventData map[string]any) ([]uuid.UUID, error) {
	var schedules []identity.AutonomySchedule
	err := s.db.WithContext(ctx).
		Where("is_active = ? AND schedule_type = ? AND trigger_event = ?", true, "trigger_based", eventType).
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	var executed []uuid.UUID
	for _, schedule := range schedules {
		if s.evaluateTriggerCondition(schedule.TriggerCondition, eventData) {
			_, err := s.ExecuteSchedule(ctx, schedule.ID)
			if err == nil {
				executed = append(executed, schedule.ID)
			}
		}
	}

	return executed, nil
}

// Action execution functions
func (s *Service) executeFunctionAction(ctx context.Context, schedule identity.AutonomySchedule) map[string]any {
	return map[string]any{
		"status":       "simulated",
		"action":       "execute_function",
		"payload":      schedule.ActionPayload,
		"agent_id":     schedule.AgentID,
		"executed_at":  time.Now().Unix(),
	}
}

func (s *Service) spawnAgentAction(ctx context.Context, schedule identity.AutonomySchedule) map[string]any {
	return map[string]any{
		"status":       "simulated",
		"action":       "spawn_agent",
		"payload":      schedule.ActionPayload,
		"agent_id":     schedule.AgentID,
		"executed_at":  time.Now().Unix(),
	}
}

func (s *Service) sendMessageAction(ctx context.Context, schedule identity.AutonomySchedule) map[string]any {
	return map[string]any{
		"status":       "simulated",
		"action":       "send_message",
		"payload":      schedule.ActionPayload,
		"agent_id":     schedule.AgentID,
		"executed_at":  time.Now().Unix(),
	}
}

func (s *Service) updateStateAction(ctx context.Context, schedule identity.AutonomySchedule) map[string]any {
	return map[string]any{
		"status":       "simulated",
		"action":       "update_state",
		"payload":      schedule.ActionPayload,
		"agent_id":     schedule.AgentID,
		"executed_at":  time.Now().Unix(),
	}
}

func (s *Service) evolveAction(ctx context.Context, schedule identity.AutonomySchedule) map[string]any {
	return map[string]any{
		"status":       "simulated",
		"action":       "evolve",
		"payload":      schedule.ActionPayload,
		"agent_id":     schedule.AgentID,
		"executed_at":  time.Now().Unix(),
	}
}

// calculateNextRun returns the next run time for a cron expression using robfig/cron.
func (s *Service) calculateNextRun(cronExpr string) (time.Time, error) {
	if cronExpr == "" {
		return time.Time{}, fmt.Errorf("cron expression is required")
	}
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return schedule.Next(time.Now().UTC()), nil
}

// evaluateTriggerCondition evaluates if trigger conditions are met
func (s *Service) evaluateTriggerCondition(condition map[string]any, eventData map[string]any) bool {
	if condition == nil || len(condition) == 0 {
		return true
	}

	// Simple equality check for now
	for key, expectedValue := range condition {
		eventValue, exists := eventData[key]
		if !exists || eventValue != expectedValue {
			return false
		}
	}
	return true
}
