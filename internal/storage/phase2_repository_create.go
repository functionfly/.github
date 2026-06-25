package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateChatSession creates a new AI chat session.
func (r *Phase2Repository) CreateChatSession(ctx context.Context, sess *AIChatSession) (*AIChatSession, error) {
	sess.ID = uuid.New()
	sess.IsActive = true
	sess.CreatedAt = time.Now()
	sess.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_chat_sessions (id, user_id, tenant_id, title, context_type, context_reference, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		sess.ID, sess.UserID, sess.TenantID, sess.Title, sess.ContextType, sess.ContextReference, sess.IsActive, sess.CreatedAt, sess.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}
	return sess, nil
}

// CreateChatMessage creates a new AI chat message.
func (r *Phase2Repository) CreateChatMessage(ctx context.Context, msg *AIChatMessage) (*AIChatMessage, error) {
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()

	var metadataParam interface{}
	if msg.Metadata != nil {
		b, _ := json.Marshal(msg.Metadata)
		metadataParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_chat_messages (id, session_id, role, content, tokens_used, model, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		msg.ID, msg.SessionID, msg.Role, msg.Content, msg.TokensUsed, msg.Model, metadataParam, msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat message: %w", err)
	}
	return msg, nil
}

// CreatePerformanceGoal creates a new performance goal.
func (r *Phase2Repository) CreatePerformanceGoal(ctx context.Context, goal *PerformanceGoal) (*PerformanceGoal, error) {
	goal.ID = uuid.New()
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO performance_goals (id, employee_id, tenant_id, title, description, category, status, priority, target_date, completed_at, progress_pct, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		goal.ID, goal.EmployeeID, goal.TenantID, goal.Title, goal.Description, goal.Category, goal.Status, goal.Priority, goal.TargetDate, goal.CompletedAt, goal.ProgressPct, goal.CreatedAt, goal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create performance goal: %w", err)
	}
	return goal, nil
}

// CreatePerformanceReview creates a new performance review.
func (r *Phase2Repository) CreatePerformanceReview(ctx context.Context, rev *PerformanceReview) (*PerformanceReview, error) {
	rev.ID = uuid.New()
	rev.CreatedAt = time.Now()
	rev.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO performance_reviews (id, employee_id, reviewer_id, tenant_id, review_period, review_type, status, strengths, areas_for_improvement, overall_rating, comments, submitted_at, acknowledged_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		rev.ID, rev.EmployeeID, rev.ReviewerID, rev.TenantID, rev.ReviewPeriod, rev.ReviewType, rev.Status, rev.Strengths, rev.AreasForImprovement, rev.OverallRating, rev.Comments, rev.SubmittedAt, rev.AcknowledgedAt, rev.CreatedAt, rev.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create performance review: %w", err)
	}
	return rev, nil
}

// CreatePeerFeedback creates peer feedback.
func (r *Phase2Repository) CreatePeerFeedback(ctx context.Context, fb *PeerFeedback) (*PeerFeedback, error) {
	fb.ID = uuid.New()
	fb.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO peer_feedback (id, review_id, from_employee_id, to_employee_id, tenant_id, feedback_text, rating, is_anonymous, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		fb.ID, fb.ReviewID, fb.FromEmployeeID, fb.ToEmployeeID, fb.TenantID, fb.FeedbackText, fb.Rating, fb.IsAnonymous, fb.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer feedback: %w", err)
	}
	return fb, nil
}

// CreateTimeEntry creates a new time entry.
func (r *Phase2Repository) CreateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error) {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO time_entries (id, employee_id, tenant_id, project_id, task_id, date, hours, description, entry_type, is_billable, approved_by, approved_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		entry.ID, entry.EmployeeID, entry.TenantID, entry.ProjectID, entry.TaskID, entry.Date, entry.Hours, entry.Description, entry.EntryType, entry.IsBillable, entry.ApprovedBy, entry.ApprovedAt, entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}
	return entry, nil
}

// CreatePTORequest creates a new PTO request.
func (r *Phase2Repository) CreatePTORequest(ctx context.Context, req *PTORequest) (*PTORequest, error) {
	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pto_requests (id, employee_id, tenant_id, pto_type, start_date, end_date, days, reason, status, approved_by, approved_at, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		req.ID, req.EmployeeID, req.TenantID, req.PTOType, req.StartDate, req.EndDate, req.Days, req.Reason, req.Status, req.ApprovedBy, req.ApprovedAt, req.Notes, req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create PTO request: %w", err)
	}
	return req, nil
}
