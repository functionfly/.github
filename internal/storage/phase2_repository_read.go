package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetChatSessionByID retrieves a chat session by ID.
func (r *Phase2Repository) GetChatSessionByID(ctx context.Context, id uuid.UUID) (*AIChatSession, error) {
	sess := &AIChatSession{}
	var contextRef sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, title, context_type, context_reference, is_active, created_at, updated_at
		FROM ai_chat_sessions WHERE id = $1`, id).Scan(
		&sess.ID, &sess.UserID, &sess.TenantID, &sess.Title, &sess.ContextType, &contextRef, &sess.IsActive, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	if contextRef.Valid {
		ref, err := uuid.Parse(contextRef.String)
		if err == nil {
			sess.ContextReference = &ref
		}
	}
	return sess, nil
}

// ListChatSessions lists AI chat sessions for a user.
func (r *Phase2Repository) ListChatSessions(ctx context.Context, userID uuid.UUID, opts ListAIChatSessionsOpts) ([]*AIChatSession, int, error) {
	where := "WHERE user_id = $1 AND is_active = TRUE"
	args := []interface{}{userID}
	argIdx := 2

	if opts.ContextType != nil {
		where += fmt.Sprintf(" AND context_type = $%d", argIdx)
		args = append(args, *opts.ContextType)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM ai_chat_sessions %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat sessions: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, user_id, tenant_id, title, context_type, context_reference, is_active, created_at, updated_at
		FROM ai_chat_sessions %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*AIChatSession
	for rows.Next() {
		sess := &AIChatSession{}
		var contextRef sql.NullString
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.TenantID, &sess.Title, &sess.ContextType, &contextRef, &sess.IsActive, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat session: %w", err)
		}
		if contextRef.Valid {
			ref, err := uuid.Parse(contextRef.String)
			if err == nil {
				sess.ContextReference = &ref
			}
		}
		sessions = append(sessions, sess)
	}
	return sessions, total, nil
}

// ListChatMessages lists messages for a chat session.
func (r *Phase2Repository) ListChatMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*AIChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, role, content, tokens_used, model, metadata, created_at
		FROM ai_chat_messages WHERE session_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*AIChatMessage
	for rows.Next() {
		msg := &AIChatMessage{}
		var model sql.NullString
		var metadataBytes []byte
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.TokensUsed, &model, &metadataBytes, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		if model.Valid {
			msg.Model = &model.String
		}
		if metadataBytes != nil {
			var m JSONMap
			if err := json.Unmarshal(metadataBytes, &m); err == nil {
				msg.Metadata = m
			}
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetPerformanceGoalByID retrieves a performance goal by ID.
func (r *Phase2Repository) GetPerformanceGoalByID(ctx context.Context, id uuid.UUID) (*PerformanceGoal, error) {
	goal := &PerformanceGoal{}
	var description sql.NullString
	var targetDate sql.NullTime
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, title, description, category, status, priority, target_date, completed_at, progress_pct, created_at, updated_at
		FROM performance_goals WHERE id = $1`, id).Scan(
		&goal.ID, &goal.EmployeeID, &goal.TenantID, &goal.Title, &description, &goal.Category, &goal.Status, &goal.Priority, &targetDate, &completedAt, &goal.ProgressPct, &goal.CreatedAt, &goal.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get performance goal: %w", err)
	}
	if description.Valid {
		goal.Description = &description.String
	}
	if targetDate.Valid {
		goal.TargetDate = &targetDate.Time
	}
	if completedAt.Valid {
		goal.CompletedAt = &completedAt.Time
	}
	return goal, nil
}

// ListPerformanceGoals lists performance goals for an employee.
func (r *Phase2Repository) ListPerformanceGoals(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceGoalsOpts) ([]*PerformanceGoal, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Category != nil {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *opts.Category)
		argIdx++
	}
	if opts.Priority != nil {
		where += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, *opts.Priority)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM performance_goals %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count performance goals: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, title, description, category, status, priority, target_date, completed_at, progress_pct, created_at, updated_at
		FROM performance_goals %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list performance goals: %w", err)
	}
	defer rows.Close()

	var goals []*PerformanceGoal
	for rows.Next() {
		goal := &PerformanceGoal{}
		var description sql.NullString
		var targetDate sql.NullTime
		var completedAt sql.NullTime
		if err := rows.Scan(&goal.ID, &goal.EmployeeID, &goal.TenantID, &goal.Title, &description, &goal.Category, &goal.Status, &goal.Priority, &targetDate, &completedAt, &goal.ProgressPct, &goal.CreatedAt, &goal.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan performance goal: %w", err)
		}
		if description.Valid {
			goal.Description = &description.String
		}
		if targetDate.Valid {
			goal.TargetDate = &targetDate.Time
		}
		if completedAt.Valid {
			goal.CompletedAt = &completedAt.Time
		}
		goals = append(goals, goal)
	}
	return goals, total, nil
}

// GetPerformanceReviewByID retrieves a performance review by ID.
func (r *Phase2Repository) GetPerformanceReviewByID(ctx context.Context, id uuid.UUID) (*PerformanceReview, error) {
	rev := &PerformanceReview{}
	var strengths, areasForImprovement, comments sql.NullString
	var overallRating sql.NullInt64
	var submittedAt, acknowledgedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, reviewer_id, tenant_id, review_period, review_type, status, strengths, areas_for_improvement, overall_rating, comments, submitted_at, acknowledged_at, created_at, updated_at
		FROM performance_reviews WHERE id = $1`, id).Scan(
		&rev.ID, &rev.EmployeeID, &rev.ReviewerID, &rev.TenantID, &rev.ReviewPeriod, &rev.ReviewType, &rev.Status,
		&strengths, &areasForImprovement, &overallRating, &comments, &submittedAt, &acknowledgedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get performance review: %w", err)
	}
	if strengths.Valid {
		rev.Strengths = &strengths.String
	}
	if areasForImprovement.Valid {
		rev.AreasForImprovement = &areasForImprovement.String
	}
	if overallRating.Valid {
		rating := int(overallRating.Int64)
		rev.OverallRating = &rating
	}
	if comments.Valid {
		rev.Comments = &comments.String
	}
	if submittedAt.Valid {
		rev.SubmittedAt = &submittedAt.Time
	}
	if acknowledgedAt.Valid {
		rev.AcknowledgedAt = &acknowledgedAt.Time
	}
	return rev, nil
}

// ListPerformanceReviews lists performance reviews for an employee.
func (r *Phase2Repository) ListPerformanceReviews(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceReviewsOpts) ([]*PerformanceReview, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.ReviewType != nil {
		where += fmt.Sprintf(" AND review_type = $%d", argIdx)
		args = append(args, *opts.ReviewType)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Period != nil {
		where += fmt.Sprintf(" AND review_period = $%d", argIdx)
		args = append(args, *opts.Period)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM performance_reviews %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count performance reviews: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, reviewer_id, tenant_id, review_period, review_type, status, strengths, areas_for_improvement, overall_rating, comments, submitted_at, acknowledged_at, created_at, updated_at
		FROM performance_reviews %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list performance reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*PerformanceReview
	for rows.Next() {
		rev := &PerformanceReview{}
		var strengths, areasForImprovement, comments sql.NullString
		var overallRating sql.NullInt64
		var submittedAt, acknowledgedAt sql.NullTime
		if err := rows.Scan(&rev.ID, &rev.EmployeeID, &rev.ReviewerID, &rev.TenantID, &rev.ReviewPeriod, &rev.ReviewType, &rev.Status,
			&strengths, &areasForImprovement, &overallRating, &comments, &submittedAt, &acknowledgedAt, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan performance review: %w", err)
		}
		if strengths.Valid {
			rev.Strengths = &strengths.String
		}
		if areasForImprovement.Valid {
			rev.AreasForImprovement = &areasForImprovement.String
		}
		if overallRating.Valid {
			rating := int(overallRating.Int64)
			rev.OverallRating = &rating
		}
		if comments.Valid {
			rev.Comments = &comments.String
		}
		if submittedAt.Valid {
			rev.SubmittedAt = &submittedAt.Time
		}
		if acknowledgedAt.Valid {
			rev.AcknowledgedAt = &acknowledgedAt.Time
		}
		reviews = append(reviews, rev)
	}
	return reviews, total, nil
}

// ListPeerFeedback lists peer feedback for an employee.
func (r *Phase2Repository) ListPeerFeedback(ctx context.Context, toEmployeeID uuid.UUID, limit, offset int) ([]*PeerFeedback, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, review_id, from_employee_id, to_employee_id, tenant_id, feedback_text, rating, is_anonymous, created_at
		FROM peer_feedback WHERE to_employee_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		toEmployeeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list peer feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []*PeerFeedback
	for rows.Next() {
		fb := &PeerFeedback{}
		var reviewID sql.NullString
		var rating sql.NullInt64
		if err := rows.Scan(&fb.ID, &reviewID, &fb.FromEmployeeID, &fb.ToEmployeeID, &fb.TenantID, &fb.FeedbackText, &rating, &fb.IsAnonymous, &fb.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan peer feedback: %w", err)
		}
		if reviewID.Valid {
			rid, err := uuid.Parse(reviewID.String)
			if err == nil {
				fb.ReviewID = &rid
			}
		}
		if rating.Valid {
			r := int(rating.Int64)
			fb.Rating = &r
		}
		feedbacks = append(feedbacks, fb)
	}
	return feedbacks, nil
}

// GetTimeEntryByID retrieves a time entry by ID.
func (r *Phase2Repository) GetTimeEntryByID(ctx context.Context, id uuid.UUID) (*TimeEntry, error) {
	entry := &TimeEntry{}
	var projectID, taskID, approvedBy sql.NullString
	var description sql.NullString
	var approvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, project_id, task_id, date, hours, description, entry_type, is_billable, approved_by, approved_at, created_at, updated_at
		FROM time_entries WHERE id = $1`, id).Scan(
		&entry.ID, &entry.EmployeeID, &entry.TenantID, &projectID, &taskID, &entry.Date, &entry.Hours, &description, &entry.EntryType, &entry.IsBillable, &approvedBy, &approvedAt, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}
	if projectID.Valid {
		pid, err := uuid.Parse(projectID.String)
		if err == nil {
			entry.ProjectID = &pid
		}
	}
	if taskID.Valid {
		tid, err := uuid.Parse(taskID.String)
		if err == nil {
			entry.TaskID = &tid
		}
	}
	if description.Valid {
		entry.Description = &description.String
	}
	if approvedBy.Valid {
		aid, err := uuid.Parse(approvedBy.String)
		if err == nil {
			entry.ApprovedBy = &aid
		}
	}
	if approvedAt.Valid {
		entry.ApprovedAt = &approvedAt.Time
	}
	return entry, nil
}

// ListTimeEntries lists time entries for an employee.
func (r *Phase2Repository) ListTimeEntries(ctx context.Context, employeeID uuid.UUID, opts ListTimeEntriesOpts) ([]*TimeEntry, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *opts.ProjectID)
		argIdx++
	}
	if opts.EntryType != nil {
		where += fmt.Sprintf(" AND entry_type = $%d", argIdx)
		args = append(args, *opts.EntryType)
		argIdx++
	}
	if opts.StartDate != nil {
		where += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, *opts.StartDate)
		argIdx++
	}
	if opts.EndDate != nil {
		where += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, *opts.EndDate)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM time_entries %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count time entries: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY date DESC, created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, project_id, task_id, date, hours, description, entry_type, is_billable, approved_by, approved_at, created_at, updated_at
		FROM time_entries %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list time entries: %w", err)
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		entry := &TimeEntry{}
		var projectID, taskID, approvedBy sql.NullString
		var description sql.NullString
		var approvedAt sql.NullTime
		if err := rows.Scan(&entry.ID, &entry.EmployeeID, &entry.TenantID, &projectID, &taskID, &entry.Date, &entry.Hours, &description, &entry.EntryType, &entry.IsBillable, &approvedBy, &approvedAt, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan time entry: %w", err)
		}
		if projectID.Valid {
			pid, err := uuid.Parse(projectID.String)
			if err == nil {
				entry.ProjectID = &pid
			}
		}
		if taskID.Valid {
			tid, err := uuid.Parse(taskID.String)
			if err == nil {
				entry.TaskID = &tid
			}
		}
		if description.Valid {
			entry.Description = &description.String
		}
		if approvedBy.Valid {
			aid, err := uuid.Parse(approvedBy.String)
			if err == nil {
				entry.ApprovedBy = &aid
			}
		}
		if approvedAt.Valid {
			entry.ApprovedAt = &approvedAt.Time
		}
		entries = append(entries, entry)
	}
	return entries, total, nil
}

// GetPTORequestByID retrieves a PTO request by ID.
func (r *Phase2Repository) GetPTORequestByID(ctx context.Context, id uuid.UUID) (*PTORequest, error) {
	req := &PTORequest{}
	var reason, notes sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, pto_type, start_date, end_date, days, reason, status, approved_by, approved_at, notes, created_at, updated_at
		FROM pto_requests WHERE id = $1`, id).Scan(
		&req.ID, &req.EmployeeID, &req.TenantID, &req.PTOType, &req.StartDate, &req.EndDate, &req.Days, &reason, &req.Status, &approvedBy, &approvedAt, &notes, &req.CreatedAt, &req.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PTO request: %w", err)
	}
	if reason.Valid {
		req.Reason = &reason.String
	}
	if notes.Valid {
		req.Notes = &notes.String
	}
	if approvedBy.Valid {
		aid, err := uuid.Parse(approvedBy.String)
		if err == nil {
			req.ApprovedBy = &aid
		}
	}
	if approvedAt.Valid {
		req.ApprovedAt = &approvedAt.Time
	}
	return req, nil
}

// ListPTORequests lists PTO requests for an employee.
func (r *Phase2Repository) ListPTORequests(ctx context.Context, employeeID uuid.UUID, opts ListPTORequestsOpts) ([]*PTORequest, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.PTOType != nil {
		where += fmt.Sprintf(" AND pto_type = $%d", argIdx)
		args = append(args, *opts.PTOType)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM pto_requests %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count PTO requests: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, pto_type, start_date, end_date, days, reason, status, approved_by, approved_at, notes, created_at, updated_at
		FROM pto_requests %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list PTO requests: %w", err)
	}
	defer rows.Close()

	var requests []*PTORequest
	for rows.Next() {
		req := &PTORequest{}
		var reason, notes sql.NullString
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		if err := rows.Scan(&req.ID, &req.EmployeeID, &req.TenantID, &req.PTOType, &req.StartDate, &req.EndDate, &req.Days, &reason, &req.Status, &approvedBy, &approvedAt, &notes, &req.CreatedAt, &req.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan PTO request: %w", err)
		}
		if reason.Valid {
			req.Reason = &reason.String
		}
		if notes.Valid {
			req.Notes = &notes.String
		}
		if approvedBy.Valid {
			aid, err := uuid.Parse(approvedBy.String)
			if err == nil {
				req.ApprovedBy = &aid
			}
		}
		if approvedAt.Valid {
			req.ApprovedAt = &approvedAt.Time
		}
		requests = append(requests, req)
	}
	return requests, total, nil
}
