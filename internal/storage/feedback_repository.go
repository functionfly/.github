package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FeedbackRepository handles feedback-related database operations
type FeedbackRepository struct {
	db *PostgresDB
}

// NewFeedbackRepository creates a new feedback repository
func NewFeedbackRepository(db *PostgresDB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// CreateFeedback creates a new feedback submission
func (r *FeedbackRepository) CreateFeedback(ctx context.Context, feedback *Feedback) (*Feedback, error) {
	feedback.ID = uuid.New()
	feedback.Status = "submitted"
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	var ipAddr interface{} = feedback.IPAddress
	if feedback.IPAddress == "" {
		ipAddr = nil
	}
	var priority interface{} = feedback.Priority
	if feedback.Priority == "" {
		priority = nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO feedback (id, user_id, user_email, feedback_type, subject, message, priority, browser_info, status, ip_address, user_agent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		feedback.ID, feedback.UserID, feedback.UserEmail, feedback.FeedbackType, feedback.Subject,
		feedback.Message, priority, feedback.BrowserInfo, feedback.Status,
		ipAddr, feedback.UserAgent, feedback.CreatedAt, feedback.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit feedback: %w", err)
	}

	return feedback, nil
}

// GetFeedbackByID retrieves a feedback submission by ID
func (r *FeedbackRepository) GetFeedbackByID(ctx context.Context, id uuid.UUID) (*Feedback, error) {
	feedback := &Feedback{}
	var userID, userEmail sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, user_email, feedback_type, subject, message, priority, browser_info, status, ip_address, user_agent, created_at, updated_at
		FROM feedback WHERE id = $1`, id).Scan(
		&feedback.ID, &userID, &userEmail, &feedback.FeedbackType, &feedback.Subject,
		&feedback.Message, &feedback.Priority, &feedback.BrowserInfo, &feedback.Status,
		&feedback.IPAddress, &feedback.UserAgent, &feedback.CreatedAt, &feedback.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feedback not found")
		}
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	if userID.Valid {
		uid := uuid.MustParse(userID.String)
		feedback.UserID = &uid
	}
	if userEmail.Valid {
		feedback.UserEmail = &userEmail.String
	}

	attachments, err := r.GetFeedbackAttachments(ctx, id)
	if err != nil {
		fmt.Printf("Warning: failed to load attachments for feedback %s: %v\n", id, err)
	} else {
		feedback.Attachments = attachments
	}

	return feedback, nil
}

// GetFeedbackByUser retrieves feedback submissions for a user (authenticated or anonymous by email)
func (r *FeedbackRepository) GetFeedbackByUser(ctx context.Context, userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error) {
	var feedbacks []Feedback
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, user_email, feedback_type, subject, message, priority, browser_info, status, ip_address, user_agent, created_at, updated_at
			FROM feedback
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{*userID, limit, offset}
	} else if userEmail != nil {
		query = `
			SELECT id, user_id, user_email, feedback_type, subject, message, priority, browser_info, status, ip_address, user_agent, created_at, updated_at
			FROM feedback
			WHERE user_email = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{*userEmail, limit, offset}
	} else {
		return feedbacks, fmt.Errorf("either userID or userEmail must be provided")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedback: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		feedback := Feedback{}
		var userIDVal, userEmailVal sql.NullString

		err := rows.Scan(
			&feedback.ID, &userIDVal, &userEmailVal, &feedback.FeedbackType, &feedback.Subject,
			&feedback.Message, &feedback.Priority, &feedback.BrowserInfo, &feedback.Status,
			&feedback.IPAddress, &feedback.UserAgent, &feedback.CreatedAt, &feedback.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback: %w", err)
		}

		if userIDVal.Valid {
			uid := uuid.MustParse(userIDVal.String)
			feedback.UserID = &uid
		}
		if userEmailVal.Valid {
			feedback.UserEmail = &userEmailVal.String
		}

		attachments, err := r.GetFeedbackAttachments(ctx, feedback.ID)
		if err != nil {
			fmt.Printf("Warning: failed to load attachments for feedback %s: %v\n", feedback.ID, err)
		} else {
			feedback.Attachments = attachments
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// ListFeedback retrieves feedback submissions with pagination (admin only).
func (r *FeedbackRepository) ListFeedback(ctx context.Context, limit, offset int, statusFilter *string, typeFilter *string) ([]Feedback, error) {
	var feedbacks []Feedback
	var query string
	var args []interface{}
	var where []string
	var n int

	if statusFilter != nil {
		n++
		where = append(where, fmt.Sprintf("status = $%d", n))
	}
	if typeFilter != nil {
		n++
		where = append(where, fmt.Sprintf("feedback_type = $%d", n))
	}

	baseSelect := `
		SELECT id, user_id, user_email, feedback_type, subject, message, priority, browser_info, status, ip_address, user_agent, created_at, updated_at
		FROM feedback`
	if len(where) > 0 {
		query = baseSelect + " WHERE " + strings.Join(where, " AND ") + " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", n+1) + " OFFSET $" + fmt.Sprintf("%d", n+2)
		for _, a := range []*string{statusFilter, typeFilter} {
			if a != nil {
				args = append(args, *a)
			}
		}
		args = append(args, limit, offset)
	} else {
		query = baseSelect + " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedback: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		feedback := Feedback{}
		var userIDVal, userEmailVal sql.NullString

		err := rows.Scan(
			&feedback.ID, &userIDVal, &userEmailVal, &feedback.FeedbackType, &feedback.Subject,
			&feedback.Message, &feedback.Priority, &feedback.BrowserInfo, &feedback.Status,
			&feedback.IPAddress, &feedback.UserAgent, &feedback.CreatedAt, &feedback.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback: %w", err)
		}

		if userIDVal.Valid {
			uid := uuid.MustParse(userIDVal.String)
			feedback.UserID = &uid
		}
		if userEmailVal.Valid {
			feedback.UserEmail = &userEmailVal.String
		}

		attachments, err := r.GetFeedbackAttachments(ctx, feedback.ID)
		if err != nil {
			fmt.Printf("Warning: failed to load attachments for feedback %s: %v\n", feedback.ID, err)
		} else {
			feedback.Attachments = attachments
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// UpdateFeedbackStatus updates the status of a feedback submission
func (r *FeedbackRepository) UpdateFeedbackStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE feedback
		SET status = $1, updated_at = $2
		WHERE id = $3`,
		status, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to update feedback status: %w", err)
	}

	return nil
}

// CreateFeedbackAttachment creates a new feedback attachment
func (r *FeedbackRepository) CreateFeedbackAttachment(ctx context.Context, attachment *FeedbackAttachment) (*FeedbackAttachment, error) {
	attachment.ID = uuid.New()
	attachment.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback_attachments (id, feedback_id, filename, content_type, size, s3_key, s3_bucket, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		attachment.ID, attachment.FeedbackID, attachment.Filename, attachment.ContentType,
		attachment.Size, attachment.S3Key, attachment.S3Bucket, attachment.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create feedback attachment: %w", err)
	}

	return attachment, nil
}

// GetFeedbackAttachments retrieves all attachments for a feedback submission
func (r *FeedbackRepository) GetFeedbackAttachments(ctx context.Context, feedbackID uuid.UUID) ([]FeedbackAttachment, error) {
	var attachments []FeedbackAttachment

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, feedback_id, filename, content_type, size, s3_key, s3_bucket, created_at
		FROM feedback_attachments
		WHERE feedback_id = $1
		ORDER BY created_at ASC`, feedbackID)

	if err != nil {
		return nil, fmt.Errorf("failed to query feedback attachments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		attachment := FeedbackAttachment{}
		err := rows.Scan(
			&attachment.ID, &attachment.FeedbackID, &attachment.Filename, &attachment.ContentType,
			&attachment.Size, &attachment.S3Key, &attachment.S3Bucket, &attachment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback attachment: %w", err)
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// GetFeedbackAttachmentByID retrieves a single attachment by ID (for download).
func (r *FeedbackRepository) GetFeedbackAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*FeedbackAttachment, error) {
	var att FeedbackAttachment
	err := r.db.QueryRowContext(ctx, `
		SELECT id, feedback_id, filename, content_type, size, s3_key, s3_bucket, created_at
		FROM feedback_attachments
		WHERE id = $1`, attachmentID).Scan(
		&att.ID, &att.FeedbackID, &att.Filename, &att.ContentType,
		&att.Size, &att.S3Key, &att.S3Bucket, &att.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &att, nil
}

// GetFeedbackStats returns statistics about feedback submissions
func (r *FeedbackRepository) GetFeedbackStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedback").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total feedback count: %w", err)
	}
	stats["total"] = totalCount

	statusRows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) as count
		FROM feedback
		GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback status stats: %w", err)
	}
	defer statusRows.Close()

	statusStats := make(map[string]int)
	for statusRows.Next() {
		var status string
		var count int
		err := statusRows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status stats: %w", err)
		}
		statusStats[status] = count
	}
	stats["status_breakdown"] = statusStats

	typeRows, err := r.db.QueryContext(ctx, `
		SELECT feedback_type, COUNT(*) as count
		FROM feedback
		GROUP BY feedback_type`)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback type stats: %w", err)
	}
	defer typeRows.Close()

	typeStats := make(map[string]int)
	for typeRows.Next() {
		var feedbackType string
		var count int
		err := typeRows.Scan(&feedbackType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %w", err)
		}
		typeStats[feedbackType] = count
	}
	stats["type_breakdown"] = typeStats

	return stats, nil
}

// GetFeedbackTrends returns feedback submission trends over time
func (r *FeedbackRepository) GetFeedbackTrends(ctx context.Context) (map[string]interface{}, error) {
	trends := make(map[string]interface{})

	dailyQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM feedback
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at)`
	dailyRows, err := r.db.QueryContext(ctx, dailyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily trends: %w", err)
	}
	defer dailyRows.Close()

	dailyTrends := make([]map[string]interface{}, 0)
	for dailyRows.Next() {
		var date string
		var count int
		err := dailyRows.Scan(&date, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily trends: %w", err)
		}
		dailyTrends = append(dailyTrends, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}
	trends["daily"] = dailyTrends

	weeklyQuery := `
		SELECT DATE_TRUNC('week', created_at) as week, COUNT(*) as count
		FROM feedback
		WHERE created_at >= CURRENT_DATE - INTERVAL '12 weeks'
		GROUP BY DATE_TRUNC('week', created_at)
		ORDER BY DATE_TRUNC('week', created_at)`
	weeklyRows, err := r.db.QueryContext(ctx, weeklyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly trends: %w", err)
	}
	defer weeklyRows.Close()

	weeklyTrends := make([]map[string]interface{}, 0)
	for weeklyRows.Next() {
		var week string
		var count int
		err := weeklyRows.Scan(&week, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weekly trends: %w", err)
		}
		dailyTrends = append(dailyTrends, map[string]interface{}{
			"week":  week,
			"count": count,
		})
	}
	trends["weekly"] = weeklyTrends

	monthlyQuery := `
		SELECT DATE_TRUNC('month', created_at) as month, COUNT(*) as count
		FROM feedback
		WHERE created_at >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)`
	monthlyRows, err := r.db.QueryContext(ctx, monthlyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly trends: %w", err)
	}
	defer monthlyRows.Close()

	monthlyTrends := make([]map[string]interface{}, 0)
	for monthlyRows.Next() {
		var month string
		var count int
		err := monthlyRows.Scan(&month, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly trends: %w", err)
		}
		monthlyTrends = append(monthlyTrends, map[string]interface{}{
			"month": month,
			"count": count,
		})
	}
	trends["monthly"] = monthlyTrends

	priorityQuery := `
		SELECT priority, COUNT(*) as count
		FROM feedback
		GROUP BY priority`
	priorityRows, err := r.db.QueryContext(ctx, priorityQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}
	defer priorityRows.Close()

	priorityStats := make(map[string]int)
	for priorityRows.Next() {
		var priority string
		var count int
		err := priorityRows.Scan(&priority, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan priority stats: %w", err)
		}
		priorityStats[priority] = count
	}
	trends["priority_breakdown"] = priorityStats

	responseTimeQuery := `
		SELECT AVG(EXTRACT(EPOCH FROM (updated_at - created_at))/86400) as avg_days
		FROM feedback
		WHERE status IN ('resolved', 'closed')`
	var avgResponseTime *float64
	err = r.db.QueryRowContext(ctx, responseTimeQuery).Scan(&avgResponseTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get response time: %w", err)
	}
	trends["avg_response_time_days"] = avgResponseTime

	return trends, nil
}

// GetFeedbackAnalytics returns comprehensive analytics data
func (r *FeedbackRepository) GetFeedbackAnalytics(ctx context.Context) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	basicStats, err := r.GetFeedbackStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}
	analytics["stats"] = basicStats

	trends, err := r.GetFeedbackTrends(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trends: %w", err)
	}
	analytics["trends"] = trends

	topIssuesQuery := `
		SELECT LOWER(subject) as subject_lower, COUNT(*) as count
		FROM feedback
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY LOWER(subject)
		ORDER BY count DESC
		LIMIT 10`
	topIssuesRows, err := r.db.QueryContext(ctx, topIssuesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get top issues: %w", err)
	}
	defer topIssuesRows.Close()

	topIssues := make([]map[string]interface{}, 0)
	for topIssuesRows.Next() {
		var subject string
		var count int
		err := topIssuesRows.Scan(&subject, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top issues: %w", err)
		}
		topIssues = append(topIssues, map[string]interface{}{
			"subject": subject,
			"count":   count,
		})
	}
	analytics["top_issues"] = topIssues

	userEngagementQuery := `
		SELECT
			CASE
				WHEN user_id IS NOT NULL THEN 'authenticated'
				ELSE 'anonymous'
			END as user_type,
			COUNT(*) as count
		FROM feedback
		GROUP BY user_type`
	userEngagementRows, err := r.db.QueryContext(ctx, userEngagementQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get user engagement: %w", err)
	}
	defer userEngagementRows.Close()

	userEngagement := make(map[string]int)
	for userEngagementRows.Next() {
		var userType string
		var count int
		err := userEngagementRows.Scan(&userType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user engagement: %w", err)
		}
		userEngagement[userType] = count
	}
	analytics["user_engagement"] = userEngagement

	return analytics, nil
}
