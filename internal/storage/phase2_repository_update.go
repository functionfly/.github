package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateChatSessionTitle updates a chat session title.
func (r *Phase2Repository) UpdateChatSessionTitle(ctx context.Context, id uuid.UUID, title string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_chat_sessions SET title = $1, updated_at = NOW() WHERE id = $2`, title, id)
	if err != nil {
		return fmt.Errorf("failed to update chat session title: %w", err)
	}
	return nil
}

// UpdatePerformanceGoal updates a performance goal dynamically.
func (r *Phase2Repository) UpdatePerformanceGoal(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if title, ok := updates["title"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, title)
		argIdx++
	}
	if description, ok := updates["description"]; ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}
	if category, ok := updates["category"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if priority, ok := updates["priority"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, priority)
		argIdx++
	}
	if targetDate, ok := updates["target_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("target_date = $%d", argIdx))
		args = append(args, targetDate)
		argIdx++
	}
	if progressPct, ok := updates["progress_pct"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("progress_pct = $%d", argIdx))
		args = append(args, progressPct)
		argIdx++
	}
	if completedAt, ok := updates["completed_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIdx))
		args = append(args, completedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE performance_goals SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update performance goal: %w", err)
	}
	return nil
}

// UpdatePerformanceReview updates a performance review dynamically.
func (r *Phase2Repository) UpdatePerformanceReview(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if strengths, ok := updates["strengths"]; ok {
		setParts = append(setParts, fmt.Sprintf("strengths = $%d", argIdx))
		args = append(args, strengths)
		argIdx++
	}
	if areasForImprovement, ok := updates["areas_for_improvement"]; ok {
		setParts = append(setParts, fmt.Sprintf("areas_for_improvement = $%d", argIdx))
		args = append(args, areasForImprovement)
		argIdx++
	}
	if overallRating, ok := updates["overall_rating"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("overall_rating = $%d", argIdx))
		args = append(args, overallRating)
		argIdx++
	}
	if comments, ok := updates["comments"]; ok {
		setParts = append(setParts, fmt.Sprintf("comments = $%d", argIdx))
		args = append(args, comments)
		argIdx++
	}
	if submittedAt, ok := updates["submitted_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("submitted_at = $%d", argIdx))
		args = append(args, submittedAt)
		argIdx++
	}
	if acknowledgedAt, ok := updates["acknowledged_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("acknowledged_at = $%d", argIdx))
		args = append(args, acknowledgedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE performance_reviews SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update performance review: %w", err)
	}
	return nil
}

// UpdateTimeEntry updates a time entry dynamically.
func (r *Phase2Repository) UpdateTimeEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if projectID, ok := updates["project_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, projectID)
		argIdx++
	}
	if taskID, ok := updates["task_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("task_id = $%d", argIdx))
		args = append(args, taskID)
		argIdx++
	}
	if date, ok := updates["date"]; ok {
		setParts = append(setParts, fmt.Sprintf("date = $%d", argIdx))
		args = append(args, date)
		argIdx++
	}
	if hours, ok := updates["hours"].(float64); ok {
		setParts = append(setParts, fmt.Sprintf("hours = $%d", argIdx))
		args = append(args, hours)
		argIdx++
	}
	if description, ok := updates["description"]; ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}
	if entryType, ok := updates["entry_type"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("entry_type = $%d", argIdx))
		args = append(args, entryType)
		argIdx++
	}
	if isBillable, ok := updates["is_billable"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_billable = $%d", argIdx))
		args = append(args, isBillable)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE time_entries SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update time entry: %w", err)
	}
	return nil
}

// UpdatePTORequestStatus updates a PTO request status (approve/reject).
func (r *Phase2Repository) UpdatePTORequestStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID, notes *string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE pto_requests SET status = $1, approved_by = $2, approved_at = $3, notes = $4, updated_at = NOW() WHERE id = $5`,
		status, approvedBy, now, notes, id)
	if err != nil {
		return fmt.Errorf("failed to update PTO request status: %w", err)
	}
	return nil
}
