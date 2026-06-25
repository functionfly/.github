package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateInnovationGrant updates an innovation grant dynamically.
func (r *Phase3Repository) UpdateInnovationGrant(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	if amount, ok := updates["requested_amount_cents"]; ok {
		setParts = append(setParts, fmt.Sprintf("requested_amount_cents = $%d", argIdx))
		args = append(args, amount)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if score, ok := updates["feasibility_score"]; ok {
		setParts = append(setParts, fmt.Sprintf("feasibility_score = $%d", argIdx))
		args = append(args, score)
		argIdx++
	}
	if votesFor, ok := updates["votes_for"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("votes_for = $%d", argIdx))
		args = append(args, votesFor)
		argIdx++
	}
	if votesAgainst, ok := updates["votes_against"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("votes_against = $%d", argIdx))
		args = append(args, votesAgainst)
		argIdx++
	}
	if reviewedBy, ok := updates["reviewed_by"]; ok {
		setParts = append(setParts, fmt.Sprintf("reviewed_by = $%d", argIdx))
		args = append(args, reviewedBy)
		argIdx++
	}
	if reviewedAt, ok := updates["reviewed_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("reviewed_at = $%d", argIdx))
		args = append(args, reviewedAt)
		argIdx++
	}
	if reason, ok := updates["rejection_reason"]; ok {
		setParts = append(setParts, fmt.Sprintf("rejection_reason = $%d", argIdx))
		args = append(args, reason)
		argIdx++
	}
	if fundedAt, ok := updates["funded_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("funded_at = $%d", argIdx))
		args = append(args, fundedAt)
		argIdx++
	}
	if completedAt, ok := updates["completed_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIdx))
		args = append(args, completedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE innovation_grants SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update innovation grant: %w", err)
	}
	return nil
}

// UpdateMarketplaceOpportunity updates a marketplace opportunity dynamically.
func (r *Phase3Repository) UpdateMarketplaceOpportunity(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	if skills, ok := updates["skills_required"]; ok {
		setParts = append(setParts, fmt.Sprintf("skills_required = $%d", argIdx))
		args = append(args, skills)
		argIdx++
	}
	if hours, ok := updates["hours_per_week"]; ok {
		setParts = append(setParts, fmt.Sprintf("hours_per_week = $%d", argIdx))
		args = append(args, hours)
		argIdx++
	}
	if weeks, ok := updates["duration_weeks"]; ok {
		setParts = append(setParts, fmt.Sprintf("duration_weeks = $%d", argIdx))
		args = append(args, weeks)
		argIdx++
	}
	if max, ok := updates["max_applicants"]; ok {
		setParts = append(setParts, fmt.Sprintf("max_applicants = $%d", argIdx))
		args = append(args, max)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE marketplace_opportunities SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update marketplace opportunity: %w", err)
	}
	return nil
}

// UpdateMarketplaceApplicationStatus updates a marketplace application status.
func (r *Phase3Repository) UpdateMarketplaceApplicationStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE marketplace_applications SET status = $1, reviewed_at = $2 WHERE id = $3`,
		status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update marketplace application status: %w", err)
	}
	return nil
}

// UpdateMentorshipMatch updates a mentorship match dynamically.
func (r *Phase3Repository) UpdateMentorshipMatch(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if focusArea, ok := updates["focus_area"]; ok {
		setParts = append(setParts, fmt.Sprintf("focus_area = $%d", argIdx))
		args = append(args, focusArea)
		argIdx++
	}
	if meetingFreq, ok := updates["meeting_frequency"]; ok {
		setParts = append(setParts, fmt.Sprintf("meeting_frequency = $%d", argIdx))
		args = append(args, meetingFreq)
		argIdx++
	}
	if notes, ok := updates["notes"]; ok {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, notes)
		argIdx++
	}
	if endedAt, ok := updates["ended_at"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("ended_at = $%d", argIdx))
		args = append(args, endedAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE mentorship_matches SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update mentorship match: %w", err)
	}
	return nil
}

// UpdateDocument updates a document dynamically.
func (r *Phase3Repository) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if title, ok := updates["title"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, title)
		argIdx++
	}
	if body, ok := updates["body"]; ok {
		setParts = append(setParts, fmt.Sprintf("body = $%d", argIdx))
		args = append(args, body)
		argIdx++
	}
	if docType, ok := updates["doc_type"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("doc_type = $%d", argIdx))
		args = append(args, docType)
		argIdx++
	}
	if category, ok := updates["category"]; ok {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if tags, ok := updates["tags"]; ok {
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, tags)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if viewCount, ok := updates["view_count"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("view_count = $%d", argIdx))
		args = append(args, viewCount)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE documents SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

// IncrementDocumentViewCount increments the view count for a document.
func (r *Phase3Repository) IncrementDocumentViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE documents SET view_count = view_count + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to increment document view count: %w", err)
	}
	return nil
}
