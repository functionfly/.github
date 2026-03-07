package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateEmailEvent creates a new email event record
func (db *PostgresDB) CreateEmailEvent(ctx context.Context, event *EmailEvent) error {
	return db.GORM.WithContext(ctx).Create(event).Error
}

// GetEmailEvents retrieves email events with filters
func (db *PostgresDB) GetEmailEvents(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*EmailEvent, error) {
	var events []*EmailEvent
	query := db.GORM.WithContext(ctx).Preload("User")

	// Apply filters
	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		query = query.Where("user_id = ?", userID)
	}
	if email, ok := filters["user_email"].(string); ok {
		query = query.Where("user_email = ?", email)
	}
	if eventType, ok := filters["event_type"].(string); ok {
		query = query.Where("event_type = ?", eventType)
	}
	if reviewed, ok := filters["reviewed"].(bool); ok {
		query = query.Where("reviewed = ?", reviewed)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Order by timestamp descending (most recent first)
	query = query.Order("timestamp DESC")

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get email events: %w", err)
	}

	return events, nil
}

// GetEmailEventsByUserEmail retrieves email events for a specific email address
func (db *PostgresDB) GetEmailEventsByUserEmail(ctx context.Context, email string, limit, offset int) ([]*EmailEvent, error) {
	var events []*EmailEvent
	query := db.GORM.WithContext(ctx).Where("user_email = ?", email).Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get email events by user email: %w", err)
	}

	return events, nil
}

// GetPendingBounceReviews retrieves bounce/complaint events that need admin review
func (db *PostgresDB) GetPendingBounceReviews(ctx context.Context, limit, offset int) ([]*EmailEvent, error) {
	var events []*EmailEvent
	query := db.GORM.WithContext(ctx).
		Where("event_type IN (?, ?)", "email.bounced", "email.complained").
		Where("reviewed = ?", false).
		Preload("User").
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending bounce reviews: %w", err)
	}

	return events, nil
}

// MarkEmailEventReviewed marks an email event as reviewed by an admin
func (db *PostgresDB) MarkEmailEventReviewed(ctx context.Context, eventID int64, reviewedBy uuid.UUID) error {
	now := time.Now()
	return db.GORM.WithContext(ctx).
		Model(&EmailEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"reviewed":    true,
			"reviewed_by": reviewedBy,
			"reviewed_at": now,
		}).Error
}

// GetEmailEventStats returns aggregated statistics for email events
func (db *PostgresDB) GetEmailEventStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Base query
	query := db.GORM.WithContext(ctx).Model(&EmailEvent{})

	// Apply filters
	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		query = query.Where("user_id = ?", userID)
	}
	if email, ok := filters["user_email"].(string); ok {
		query = query.Where("user_email = ?", email)
	}

	// Total events
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}
	stats["total_events"] = totalCount

	// Event type breakdown
	var eventTypeStats []struct {
		EventType string `gorm:"column:event_type"`
		Count     int64  `gorm:"column:count"`
	}
	if err := query.
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Scan(&eventTypeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get event type stats: %w", err)
	}

	eventTypes := make(map[string]int64)
	for _, stat := range eventTypeStats {
		eventTypes[stat.EventType] = stat.Count
	}
	stats["event_types"] = eventTypes

	// Bounce statistics
	var bounceCount int64
	if err := db.GORM.WithContext(ctx).
		Model(&EmailEvent{}).
		Where("event_type = ?", "email.bounced").
		Count(&bounceCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count bounces: %w", err)
	}
	stats["total_bounces"] = bounceCount

	// Pending review count
	var pendingReviewCount int64
	if err := db.GORM.WithContext(ctx).
		Model(&EmailEvent{}).
		Where("event_type IN (?, ?)", "email.bounced", "email.complained").
		Where("reviewed = ?", false).
		Count(&pendingReviewCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending reviews: %w", err)
	}
	stats["pending_reviews"] = pendingReviewCount

	// Delivery rate (delivered / (sent + bounced))
	var deliveredCount int64
	if err := db.GORM.WithContext(ctx).
		Model(&EmailEvent{}).
		Where("event_type = ?", "email.delivered").
		Count(&deliveredCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count delivered: %w", err)
	}
	stats["delivered_count"] = deliveredCount

	// Calculate delivery rate
	if totalCount > 0 {
		deliveryRate := float64(deliveredCount) / float64(totalCount) * 100
		stats["delivery_rate"] = fmt.Sprintf("%.2f%%", deliveryRate)
	} else {
		stats["delivery_rate"] = "N/A"
	}

	return stats, nil
}
