package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateFWOSIncident updates an incident dynamically.
func (r *Phase5Repository) UpdateFWOSIncident(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	if severity, ok := updates["severity"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, severity)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if commanderID, ok := updates["commander_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("commander_id = $%d", argIdx))
		args = append(args, commanderID)
		argIdx++
	}
	if rootCause, ok := updates["root_cause"]; ok {
		setParts = append(setParts, fmt.Sprintf("root_cause = $%d", argIdx))
		args = append(args, rootCause)
		argIdx++
	}
	if impact, ok := updates["impact"]; ok {
		setParts = append(setParts, fmt.Sprintf("impact = $%d", argIdx))
		args = append(args, impact)
		argIdx++
	}
	if duration, ok := updates["duration_minutes"]; ok {
		setParts = append(setParts, fmt.Sprintf("duration_minutes = $%d", argIdx))
		args = append(args, duration)
		argIdx++
	}
	if acknowledgedAt, ok := updates["acknowledged_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("acknowledged_at = $%d", argIdx))
		args = append(args, acknowledgedAt)
		argIdx++
	}
	if resolvedAt, ok := updates["resolved_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("resolved_at = $%d", argIdx))
		args = append(args, resolvedAt)
		argIdx++
	}
	if closedAt, ok := updates["closed_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("closed_at = $%d", argIdx))
		args = append(args, closedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE incidents SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}
	return nil
}

// UpdatePostmortem updates a postmortem dynamically.
func (r *Phase5Repository) UpdatePostmortem(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if summary, ok := updates["summary"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("summary = $%d", argIdx))
		args = append(args, summary)
		argIdx++
	}
	if rootCause, ok := updates["root_cause"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("root_cause = $%d", argIdx))
		args = append(args, rootCause)
		argIdx++
	}
	if contributingFactors, ok := updates["contributing_factors"]; ok {
		setParts = append(setParts, fmt.Sprintf("contributing_factors = $%d", argIdx))
		args = append(args, contributingFactors)
		argIdx++
	}
	if whatWentWell, ok := updates["what_went_well"]; ok {
		setParts = append(setParts, fmt.Sprintf("what_went_well = $%d", argIdx))
		args = append(args, whatWentWell)
		argIdx++
	}
	if whatWentWrong, ok := updates["what_went_wrong"]; ok {
		setParts = append(setParts, fmt.Sprintf("what_went_wrong = $%d", argIdx))
		args = append(args, whatWentWrong)
		argIdx++
	}
	if actionItems, ok := updates["action_items"]; ok {
		setParts = append(setParts, fmt.Sprintf("action_items = $%d", argIdx))
		args = append(args, actionItems)
		argIdx++
	}
	if lessonsLearned, ok := updates["lessons_learned"]; ok {
		setParts = append(setParts, fmt.Sprintf("lessons_learned = $%d", argIdx))
		args = append(args, lessonsLearned)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
		if status == "published" {
			setParts = append(setParts, fmt.Sprintf("published_at = $%d", argIdx))
			args = append(args, time.Now())
			argIdx++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE postmortems SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update postmortem: %w", err)
	}
	return nil
}

// UpdateFeatureFlag updates a feature flag dynamically.
func (r *Phase5Repository) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	if isEnabled, ok := updates["is_enabled"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_enabled = $%d", argIdx))
		args = append(args, isEnabled)
		argIdx++
	}
	if rolloutPct, ok := updates["rollout_pct"]; ok {
		setParts = append(setParts, fmt.Sprintf("rollout_pct = $%d", argIdx))
		args = append(args, rolloutPct)
		argIdx++
	}
	if variants, ok := updates["variants"]; ok {
		setParts = append(setParts, fmt.Sprintf("variants = $%d", argIdx))
		args = append(args, variants)
		argIdx++
	}
	if audience, ok := updates["target_audience"]; ok {
		setParts = append(setParts, fmt.Sprintf("target_audience = $%d", argIdx))
		args = append(args, audience)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE feature_flags SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update feature flag: %w", err)
	}
	return nil
}

// UpdateDataClassification updates a data classification dynamically.
func (r *Phase5Repository) UpdateDataClassification(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if classification, ok := updates["classification"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("classification = $%d", argIdx))
		args = append(args, classification)
		argIdx++
	}
	if reason, ok := updates["reason"]; ok {
		setParts = append(setParts, fmt.Sprintf("reason = $%d", argIdx))
		args = append(args, reason)
		argIdx++
	}
	if expiresAt, ok := updates["expires_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("expires_at = $%d", argIdx))
		args = append(args, expiresAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE data_classifications SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update data classification: %w", err)
	}
	return nil
}

// UpdateLifecycleWorkflowInstance updates a workflow instance (step completion, status).
func (r *Phase5Repository) UpdateLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
		if status == "completed" {
			setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIdx))
			args = append(args, time.Now())
			argIdx++
		}
	}
	if currentStep, ok := updates["current_step"]; ok {
		setParts = append(setParts, fmt.Sprintf("current_step = $%d", argIdx))
		args = append(args, currentStep)
		argIdx++
	}
	if stepsStatus, ok := updates["steps_status"]; ok {
		setParts = append(setParts, fmt.Sprintf("steps_status = $%d", argIdx))
		args = append(args, stepsStatus)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE lifecycle_workflow_instances SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update lifecycle workflow instance: %w", err)
	}
	return nil
}

// RevokeCertificate marks a certificate as revoked.
func (r *Phase5Repository) RevokeCertificate(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE employee_certificates SET status = 'revoked', revoked_at = NOW(), revoke_reason = $1 WHERE id = $2`,
		reason, id)
	if err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}
	return nil
}
