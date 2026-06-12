package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSubscriberExists   = errors.New("email is already subscribed")
	ErrSubscriberNotFound = errors.New("subscriber not found")
	ErrCampaignNotFound   = errors.New("campaign not found")
)

// CreateNewsletterSubscriber adds a new newsletter subscriber
func (db *PostgresDB) CreateNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string) (*NewsletterSubscriber, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	subscriber := &NewsletterSubscriber{
		ID:        uuid.New(),
		Email:     email,
		Name:      strings.TrimSpace(name),
		Status:    "active",
		Source:    source,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := db.GORM.WithContext(ctx).Create(subscriber).Error; err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, ErrSubscriberExists
		}
		return nil, err
	}

	return subscriber, nil
}

// GetNewsletterSubscriberByEmail retrieves a subscriber by email
func (db *PostgresDB) GetNewsletterSubscriberByEmail(ctx context.Context, email string) (*NewsletterSubscriber, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var subscriber NewsletterSubscriber
	if err := db.GORM.WithContext(ctx).Where("email = ?", email).First(&subscriber).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriberNotFound
		}
		return nil, err
	}
	return &subscriber, nil
}

// GetNewsletterSubscriberByID retrieves a subscriber by ID
func (db *PostgresDB) GetNewsletterSubscriberByID(ctx context.Context, id uuid.UUID) (*NewsletterSubscriber, error) {
	var subscriber NewsletterSubscriber
	if err := db.GORM.WithContext(ctx).Where("id = ?", id).First(&subscriber).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriberNotFound
		}
		return nil, err
	}
	return &subscriber, nil
}

// ListNewsletterSubscribers returns paginated list of subscribers
func (db *PostgresDB) ListNewsletterSubscribers(ctx context.Context, status string, limit, offset int) ([]NewsletterSubscriber, int64, error) {
	var subscribers []NewsletterSubscriber
	var total int64

	query := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{})

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("subscribed_at DESC").Limit(limit).Offset(offset).Find(&subscribers).Error; err != nil {
		return nil, 0, err
	}

	return subscribers, total, nil
}

// GetActiveNewsletterSubscribers returns all active subscribers for sending campaigns
func (db *PostgresDB) GetActiveNewsletterSubscribers(ctx context.Context) ([]NewsletterSubscriber, error) {
	var subscribers []NewsletterSubscriber
	if err := db.GORM.WithContext(ctx).
		Where("status = ?", "active").
		Order("subscribed_at ASC").
		Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// UnsubscribeNewsletterSubscriber marks a subscriber as unsubscribed
func (db *PostgresDB) UnsubscribeNewsletterSubscriber(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	now := time.Now()

	result := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"status":          "unsubscribed",
			"unsubscribed_at": now,
			"updated_at":      now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriberNotFound
	}
	return nil
}

// DeleteNewsletterSubscriber removes a subscriber
func (db *PostgresDB) DeleteNewsletterSubscriber(ctx context.Context, id uuid.UUID) error {
	result := db.GORM.WithContext(ctx).Delete(&NewsletterSubscriber{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriberNotFound
	}
	return nil
}

// MarkNewsletterSubscriberBounced marks a subscriber as bounced
func (db *PostgresDB) MarkNewsletterSubscriberBounced(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	now := time.Now()

	result := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"status":     "bounced",
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriberNotFound
	}
	return nil
}

// ConfirmNewsletterSubscription confirms a subscriber's email (double opt-in)
func (db *PostgresDB) ConfirmNewsletterSubscription(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	now := time.Now()

	result := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).
		Where("email = ? AND status = ?", email, "pending").
		Updates(map[string]interface{}{
			"status":       "active",
			"confirmed_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriberNotFound
	}
	return nil
}

// CreatePendingNewsletterSubscriber creates a subscriber in pending status for double opt-in
func (db *PostgresDB) CreatePendingNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string, confirmationToken string) (*NewsletterSubscriber, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	subscriber := &NewsletterSubscriber{
		ID:                uuid.New(),
		Email:             email,
		Name:              strings.TrimSpace(name),
		Status:            "pending",
		Source:            source,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		ConfirmationToken: &confirmationToken,
	}

	if err := db.GORM.WithContext(ctx).Create(subscriber).Error; err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, ErrSubscriberExists
		}
		return nil, err
	}

	return subscriber, nil
}

// GetNewsletterStats returns aggregate statistics for newsletter
func (db *PostgresDB) GetNewsletterStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_subscribers"] = totalCount

	var activeCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active_subscribers"] = activeCount

	var unsubscribedCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).Where("status = ?", "unsubscribed").Count(&unsubscribedCount).Error; err != nil {
		return nil, err
	}
	stats["unsubscribed"] = unsubscribedCount

	var bouncedCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).Where("status = ?", "bounced").Count(&bouncedCount).Error; err != nil {
		return nil, err
	}
	stats["bounced"] = bouncedCount

	var recentSubscribers int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterSubscriber{}).
		Where("subscribed_at > ?", time.Now().AddDate(0, 0, -30)).Count(&recentSubscribers).Error; err != nil {
		return nil, err
	}
	stats["subscribers_last_30_days"] = recentSubscribers

	return stats, nil
}

// CreateNewsletterCampaign creates a new newsletter campaign
func (db *PostgresDB) CreateNewsletterCampaign(ctx context.Context, campaign *NewsletterCampaign) (*NewsletterCampaign, error) {
	if campaign.ID == uuid.Nil {
		campaign.ID = uuid.New()
	}

	if err := db.GORM.WithContext(ctx).Create(campaign).Error; err != nil {
		return nil, err
	}

	return campaign, nil
}

// GetNewsletterCampaignByID retrieves a campaign by ID
func (db *PostgresDB) GetNewsletterCampaignByID(ctx context.Context, id uuid.UUID) (*NewsletterCampaign, error) {
	var campaign NewsletterCampaign
	if err := db.GORM.WithContext(ctx).Where("id = ?", id).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCampaignNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

// ListNewsletterCampaigns returns paginated list of campaigns
func (db *PostgresDB) ListNewsletterCampaigns(ctx context.Context, status string, limit, offset int) ([]NewsletterCampaign, int64, error) {
	var campaigns []NewsletterCampaign
	var total int64

	query := db.GORM.WithContext(ctx).Model(&NewsletterCampaign{})

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// UpdateNewsletterCampaign updates a campaign
func (db *PostgresDB) UpdateNewsletterCampaign(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*NewsletterCampaign, error) {
	updates["updated_at"] = time.Now()

	if err := db.GORM.WithContext(ctx).Model(&NewsletterCampaign{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	return db.GetNewsletterCampaignByID(ctx, id)
}

// CreateNewsletterCampaignEmail creates a campaign email record
func (db *PostgresDB) CreateNewsletterCampaignEmail(ctx context.Context, campaignEmail *NewsletterCampaignEmail) error {
	if campaignEmail.ID == uuid.Nil {
		campaignEmail.ID = uuid.New()
	}
	return db.GORM.WithContext(ctx).Create(campaignEmail).Error
}

// UpdateNewsletterCampaignEmailStatus updates the status of a campaign email
func (db *PostgresDB) UpdateNewsletterCampaignEmailStatus(ctx context.Context, id uuid.UUID, status string, emailID string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if emailID != "" {
		updates["email_id"] = emailID
	}

	switch status {
	case "sent":
		now := time.Now()
		updates["sent_at"] = now
	case "delivered":
		now := time.Now()
		updates["delivered_at"] = now
	case "opened":
		now := time.Now()
		updates["opened_at"] = now
	case "clicked":
		now := time.Now()
		updates["clicked_at"] = now
	}

	return db.GORM.WithContext(ctx).Model(&NewsletterCampaignEmail{}).Where("id = ?", id).Updates(updates).Error
}

// GetNewsletterCampaignEmailsByCampaign returns all email records for a campaign
func (db *PostgresDB) GetNewsletterCampaignEmailsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]NewsletterCampaignEmail, error) {
	var emails []NewsletterCampaignEmail
	if err := db.GORM.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Order("created_at ASC").
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// UpdateCampaignStats updates campaign aggregate statistics
func (db *PostgresDB) UpdateCampaignStats(ctx context.Context, campaignID uuid.UUID) error {
	var sentCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterCampaignEmail{}).
		Where("campaign_id = ? AND status IN (?, ?, ?, ?)", campaignID, "sent", "delivered", "opened", "clicked").
		Count(&sentCount).Error; err != nil {
		return err
	}

	var openCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterCampaignEmail{}).
		Where("campaign_id = ? AND status IN (?, ?)", campaignID, "opened", "clicked").
		Count(&openCount).Error; err != nil {
		return err
	}

	var clickCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterCampaignEmail{}).
		Where("campaign_id = ? AND status = ?", campaignID, "clicked").
		Count(&clickCount).Error; err != nil {
		return err
	}

	var bounceCount int64
	if err := db.GORM.WithContext(ctx).Model(&NewsletterCampaignEmail{}).
		Where("campaign_id = ? AND status = ?", campaignID, "bounced").
		Count(&bounceCount).Error; err != nil {
		return err
	}

	return db.GORM.WithContext(ctx).Model(&NewsletterCampaign{}).Where("id = ?", campaignID).Updates(map[string]interface{}{
		"sent_count":   sentCount,
		"open_count":   openCount,
		"click_count":  clickCount,
		"bounce_count": bounceCount,
		"updated_at":   time.Now(),
	}).Error
}
