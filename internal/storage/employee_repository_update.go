package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateDepartment updates a department dynamically.
func (r *EmployeeRepository) UpdateDepartment(ctx context.Context, id int64, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}
	if slug, ok := updates["slug"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("slug = $%d", argIdx))
		args = append(args, slug)
		argIdx++
	}
	if description, ok := updates["description"]; ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}
	if parentID, ok := updates["parent_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("parent_id = $%d", argIdx))
		args = append(args, parentID)
		argIdx++
	}
	if headID, ok := updates["head_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("head_id = $%d", argIdx))
		args = append(args, headID)
		argIdx++
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, isActive)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE departments SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}
	return nil
}

// UpdateEmployee updates an employee dynamically.
func (r *EmployeeRepository) UpdateEmployee(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if employmentType, ok := updates["employment_type"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("employment_type = $%d", argIdx))
		args = append(args, employmentType)
		argIdx++
	}
	if clearanceLevel, ok := updates["clearance_level"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("clearance_level = $%d", argIdx))
		args = append(args, clearanceLevel)
		argIdx++
	}
	if departmentID, ok := updates["department_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("department_id = $%d", argIdx))
		args = append(args, departmentID)
		argIdx++
	}
	if managerID, ok := updates["manager_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("manager_id = $%d", argIdx))
		args = append(args, managerID)
		argIdx++
	}
	if workLocation, ok := updates["work_location"]; ok {
		setParts = append(setParts, fmt.Sprintf("work_location = $%d", argIdx))
		args = append(args, workLocation)
		argIdx++
	}
	if officeLocation, ok := updates["office_location"]; ok {
		setParts = append(setParts, fmt.Sprintf("office_location = $%d", argIdx))
		args = append(args, officeLocation)
		argIdx++
	}
	if timezone, ok := updates["timezone"]; ok {
		setParts = append(setParts, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, timezone)
		argIdx++
	}
	if bio, ok := updates["bio"]; ok {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, bio)
		argIdx++
	}
	if pronouns, ok := updates["pronouns"]; ok {
		setParts = append(setParts, fmt.Sprintf("pronouns = $%d", argIdx))
		args = append(args, pronouns)
		argIdx++
	}
	if emergencyContact, ok := updates["emergency_contact"].(map[string]interface{}); ok {
		b, _ := json.Marshal(emergencyContact)
		setParts = append(setParts, fmt.Sprintf("emergency_contact = $%d", argIdx))
		args = append(args, b)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE employees SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}
	return nil
}

// UpdateEmployeeSkill updates an employee skill dynamically.
func (r *EmployeeRepository) UpdateEmployeeSkill(ctx context.Context, id int64, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if proficiency, ok := updates["proficiency"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("proficiency = $%d", argIdx))
		args = append(args, proficiency)
		argIdx++
	}
	if yearsExp, ok := updates["years_exp"]; ok {
		setParts = append(setParts, fmt.Sprintf("years_exp = $%d", argIdx))
		args = append(args, yearsExp)
		argIdx++
	}
	if endorsements, ok := updates["endorsements"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("endorsements = $%d", argIdx))
		args = append(args, endorsements)
		argIdx++
	}
	if verified, ok := updates["verified"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("verified = $%d", argIdx))
		args = append(args, verified)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE employee_skills SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update employee skill: %w", err)
	}
	return nil
}

// UpdateProject updates a project dynamically.
func (r *EmployeeRepository) UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}
	if description, ok := updates["description"]; ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
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
	if ownerID, ok := updates["owner_id"].(uuid.UUID); ok {
		setParts = append(setParts, fmt.Sprintf("owner_id = $%d", argIdx))
		args = append(args, ownerID)
		argIdx++
	}
	if startDate, ok := updates["start_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argIdx))
		args = append(args, startDate)
		argIdx++
	}
	if targetDate, ok := updates["target_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("target_date = $%d", argIdx))
		args = append(args, targetDate)
		argIdx++
	}
	if tags, ok := updates["tags"].(map[string]interface{}); ok {
		b, _ := json.Marshal(tags)
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, b)
		argIdx++
	}
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		b, _ := json.Marshal(metadata)
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, b)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE projects SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// UpdateTask updates a task dynamically.
func (r *EmployeeRepository) UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	if assigneeID, ok := updates["assignee_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, assigneeID)
		argIdx++
	}
	if dueDate, ok := updates["due_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, dueDate)
		argIdx++
	}
	if estimatedHours, ok := updates["estimated_hours"]; ok {
		setParts = append(setParts, fmt.Sprintf("estimated_hours = $%d", argIdx))
		args = append(args, estimatedHours)
		argIdx++
	}
	if actualHours, ok := updates["actual_hours"]; ok {
		setParts = append(setParts, fmt.Sprintf("actual_hours = $%d", argIdx))
		args = append(args, actualHours)
		argIdx++
	}
	if tags, ok := updates["tags"].(map[string]interface{}); ok {
		b, _ := json.Marshal(tags)
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, b)
		argIdx++
	}
	if position, ok := updates["position"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("position = $%d", argIdx))
		args = append(args, position)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// UpdateTaskStatus updates just the status of a task.
func (r *EmployeeRepository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// UpdateLearningProgress updates an employee's learning progress dynamically.
func (r *EmployeeRepository) UpdateLearningProgress(ctx context.Context, id int64, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if progressPct, ok := updates["progress_pct"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("progress_pct = $%d", argIdx))
		args = append(args, progressPct)
		argIdx++
	}
	if startedAt, ok := updates["started_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("started_at = $%d", argIdx))
		args = append(args, startedAt)
		argIdx++
	}
	if completedAt, ok := updates["completed_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIdx))
		args = append(args, completedAt)
		argIdx++
	}
	if score, ok := updates["score"]; ok {
		setParts = append(setParts, fmt.Sprintf("score = $%d", argIdx))
		args = append(args, score)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE employee_learning SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update learning progress: %w", err)
	}
	return nil
}

// UpdateKnowledgeArticle updates a knowledge article dynamically.
func (r *EmployeeRepository) UpdateKnowledgeArticle(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if title, ok := updates["title"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, title)
		argIdx++
	}
	if body, ok := updates["body"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("body = $%d", argIdx))
		args = append(args, body)
		argIdx++
	}
	if category, ok := updates["category"]; ok {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if tags, ok := updates["tags"].(map[string]interface{}); ok {
		b, _ := json.Marshal(tags)
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, b)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if publishedAt, ok := updates["published_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("published_at = $%d", argIdx))
		args = append(args, publishedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE knowledge_articles SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update knowledge article: %w", err)
	}
	return nil
}

// UpdateCompensationRecord updates a compensation record dynamically.
func (r *EmployeeRepository) UpdateCompensationRecord(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if baseSalaryCents, ok := updates["base_salary_cents"].(int64); ok {
		setParts = append(setParts, fmt.Sprintf("base_salary_cents = $%d", argIdx))
		args = append(args, baseSalaryCents)
		argIdx++
	}
	if currency, ok := updates["currency"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("currency = $%d", argIdx))
		args = append(args, currency)
		argIdx++
	}
	if payFrequency, ok := updates["pay_frequency"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("pay_frequency = $%d", argIdx))
		args = append(args, payFrequency)
		argIdx++
	}
	if endDate, ok := updates["end_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIdx))
		args = append(args, endDate)
		argIdx++
	}
	if reviewDate, ok := updates["review_date"]; ok {
		setParts = append(setParts, fmt.Sprintf("review_date = $%d", argIdx))
		args = append(args, reviewDate)
		argIdx++
	}
	if notes, ok := updates["notes"]; ok {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, notes)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE compensation_records SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update compensation record: %w", err)
	}
	return nil
}

// MarkNotificationRead marks a notification as read.
func (r *EmployeeRepository) MarkNotificationRead(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE fwos_notifications SET is_read = TRUE, read_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}
	return nil
}

// MarkAllNotificationsRead marks all notifications as read for a user.
func (r *EmployeeRepository) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE fwos_notifications SET is_read = TRUE, read_at = NOW() WHERE user_id = $1 AND is_read = FALSE`, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications read: %w", err)
	}
	return nil
}
