package trustapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookRepository handles webhook database operations
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// ============================================
// Webhook CRUD Operations
// ============================================

// CreateWebhook creates a new webhook configuration
func (r *WebhookRepository) CreateWebhook(webhook *TrustWebhook) error {
	if webhook.ID == uuid.Nil {
		webhook.ID = uuid.New()
	}

	// Generate public webhook ID
	b := make([]byte, 12)
	rand.Read(b)
	webhook.WebhookID = "whk_" + hex.EncodeToString(b)

	// Generate secret if not provided
	if webhook.Secret == "" {
		secretBytes := make([]byte, 32)
		rand.Read(secretBytes)
		webhook.Secret = hex.EncodeToString(secretBytes)
	}

	// Set defaults
	if webhook.Status == "" {
		webhook.Status = string(WebhookStatusActive)
	}
	if webhook.Method == "" {
		webhook.Method = "POST"
	}
	if webhook.EventFilter == "" {
		webhook.EventFilter = "specific"
	}
	if webhook.MaxRetries == 0 {
		webhook.MaxRetries = 3
	}
	if webhook.RetryDelaySecs == 0 {
		webhook.RetryDelaySecs = 60
	}
	if webhook.TimeoutSecs == 0 {
		webhook.TimeoutSecs = 30
	}

	// Marshal arrays to JSON
	eventsJSON, _ := json.Marshal(webhook.Events)
	webhook.Events = eventsJSON

	if len(webhook.FunctionFilter) > 0 {
		filterJSON, _ := json.Marshal(webhook.FunctionFilter)
		webhook.FunctionFilter = filterJSON
	}

	if len(webhook.CustomHeaders) > 0 {
		headersJSON, _ := json.Marshal(webhook.CustomHeaders)
		webhook.CustomHeaders = headersJSON
	}

	return r.db.Create(webhook).Error
}

// GetWebhookByID retrieves a webhook by ID
func (r *WebhookRepository) GetWebhookByID(id uuid.UUID) (*TrustWebhook, error) {
	var webhook TrustWebhook
	err := r.db.First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// GetWebhookByWebhookID retrieves a webhook by its public ID
func (r *WebhookRepository) GetWebhookByWebhookID(webhookID string) (*TrustWebhook, error) {
	var webhook TrustWebhook
	err := r.db.Where("webhook_id = ?", webhookID).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// ListWebhooksForOwner lists all webhooks for an owner
func (r *WebhookRepository) ListWebhooksForOwner(
	ownerID uuid.UUID,
	ownerType string,
	status string,
	limit, offset int,
) ([]TrustWebhook, int64, error) {
	var webhooks []TrustWebhook
	var total int64

	query := r.db.Model(&TrustWebhook{}).Where("owner_id = ? AND owner_type = ?", ownerID, ownerType)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&webhooks).Error; err != nil {
		return nil, 0, err
	}

	return webhooks, total, nil
}

// UpdateWebhook updates a webhook
func (r *WebhookRepository) UpdateWebhook(webhook *TrustWebhook) error {
	// Marshal arrays to JSON if they've changed
	if webhook.Events != nil {
		eventsJSON, _ := json.Marshal(webhook.Events)
		webhook.Events = eventsJSON
	}
	if webhook.FunctionFilter != nil {
		filterJSON, _ := json.Marshal(webhook.FunctionFilter)
		webhook.FunctionFilter = filterJSON
	}
	if webhook.CustomHeaders != nil {
		headersJSON, _ := json.Marshal(webhook.CustomHeaders)
		webhook.CustomHeaders = headersJSON
	}

	return r.db.Save(webhook).Error
}

// DeleteWebhook deletes a webhook (hard delete)
func (r *WebhookRepository) DeleteWebhook(webhookID uuid.UUID) error {
	return r.db.Delete(&TrustWebhook{}, webhookID).Error
}

// ============================================
// Active Webhook Queries
// ============================================

// GetActiveWebhooksForEvent gets all active webhooks subscribed to an event type
func (r *WebhookRepository) GetActiveWebhooksForEvent(eventType WebhookEventType) ([]TrustWebhook, error) {
	var webhooks []TrustWebhook

	err := r.db.Where("status = ?", WebhookStatusActive).
		Find(&webhooks).Error

	if err != nil {
		return nil, err
	}

	// Filter by event subscription in memory (PostgreSQL JSONB query could be used for optimization)
	var filtered []TrustWebhook
	for _, w := range webhooks {
		if w.IsSubscribedToEvent(eventType) {
			filtered = append(filtered, w)
		}
	}

	return filtered, nil
}

// GetActiveWebhooksForEventAndFunction gets webhooks for an event and specific function
func (r *WebhookRepository) GetActiveWebhooksForEventAndFunction(
	eventType WebhookEventType,
	functionID uuid.UUID,
) ([]TrustWebhook, error) {
	webhooks, err := r.GetActiveWebhooksForEvent(eventType)
	if err != nil {
		return nil, err
	}

	// Filter by function
	var filtered []TrustWebhook
	for _, w := range webhooks {
		if w.IsFilteredForFunction(functionID) {
			filtered = append(filtered, w)
		}
	}

	return filtered, nil
}

// ============================================
// Webhook Status Management
// ============================================

// RecordWebhookFailure records a failed delivery attempt
func (r *WebhookRepository) RecordWebhookFailure(webhookID uuid.UUID, errorMsg string) error {
	now := time.Now()
	return r.db.Model(&TrustWebhook{}).
		Where("id = ?", webhookID).
		Updates(map[string]interface{}{
			"fail_count":   gorm.Expr("fail_count + 1"),
			"last_failure": &now,
			"status":       gorm.Expr("CASE WHEN fail_count + 1 >= 5 THEN ? ELSE status END", WebhookStatusFailed),
		}).Error
}

// RecordWebhookSuccess records a successful delivery
func (r *WebhookRepository) RecordWebhookSuccess(webhookID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&TrustWebhook{}).
		Where("id = ?", webhookID).
		Updates(map[string]interface{}{
			"fail_count":   0,
			"last_success": &now,
			"status":       WebhookStatusActive,
		}).Error
}

// DisableWebhook disables a webhook due to too many failures
func (r *WebhookRepository) DisableWebhook(webhookID uuid.UUID) error {
	return r.db.Model(&TrustWebhook{}).
		Where("id = ?", webhookID).
		Update("status", WebhookStatusDisabled).Error
}

// ResetWebhookFailures resets the failure count for a webhook
func (r *WebhookRepository) ResetWebhookFailures(webhookID uuid.UUID) error {
	return r.db.Model(&TrustWebhook{}).
		Where("id = ?", webhookID).
		Update("fail_count", 0).Error
}

// ============================================
// Delivery Operations
// ============================================

// CreateDelivery creates a new webhook delivery record
func (r *WebhookRepository) CreateDelivery(delivery *WebhookDelivery) error {
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}

	// Generate delivery ID
	b := make([]byte, 12)
	rand.Read(b)
	delivery.DeliveryID = "del_" + hex.EncodeToString(b)

	// Calculate payload hash for integrity
	if delivery.Payload != nil {
		hash := sha256.Sum256(delivery.Payload)
		delivery.PayloadHash = hex.EncodeToString(hash[:])
	}

	return r.db.Create(delivery).Error
}

// GetDeliveryByID retrieves a delivery by ID
func (r *WebhookRepository) GetDeliveryByID(id uuid.UUID) (*WebhookDelivery, error) {
	var delivery WebhookDelivery
	err := r.db.First(&delivery, id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

// GetDeliveryByDeliveryID retrieves a delivery by its public ID
func (r *WebhookRepository) GetDeliveryByDeliveryID(deliveryID string) (*WebhookDelivery, error) {
	var delivery WebhookDelivery
	err := r.db.Where("delivery_id = ?", deliveryID).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

// UpdateDeliveryStatus updates the status of a delivery
func (r *WebhookRepository) UpdateDeliveryStatus(
	deliveryID uuid.UUID,
	status WebhookDeliveryStatus,
	responseStatusCode *int,
	responseHeaders json.RawMessage,
	responseBody string,
	responseTimeMs int,
	errorMsg string,
) error {
	updates := map[string]interface{}{
		"status": string(status),
	}

	if responseStatusCode != nil {
		updates["response_status_code"] = *responseStatusCode
	}
	if responseHeaders != nil {
		updates["response_headers"] = responseHeaders
	}
	if responseBody != "" {
		updates["response_body"] = responseBody
	}
	if responseTimeMs > 0 {
		updates["response_time_ms"] = responseTimeMs
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	switch status {
	case WebhookDeliveryStatusSent:
		now := time.Now()
		updates["sent_at"] = &now
	case WebhookDeliveryStatusDelivered:
		now := time.Now()
		updates["delivered_at"] = &now
	}

	return r.db.Model(&WebhookDelivery{}).
		Where("id = ?", deliveryID).
		Updates(updates).Error
}

// ScheduleRetry schedules a retry for a failed delivery
func (r *WebhookRepository) ScheduleRetry(deliveryID uuid.UUID, nextAttemptNumber int, delaySecs int) error {
	nextRetry := time.Now().Add(time.Duration(delaySecs) * time.Second)

	return r.db.Model(&WebhookDelivery{}).
		Where("id = ?", deliveryID).
		Updates(map[string]interface{}{
			"status":         string(WebhookDeliveryStatusRetrying),
			"attempt_number": nextAttemptNumber,
			"next_retry_at":  &nextRetry,
		}).Error
}

// ListDeliveriesForWebhook lists all deliveries for a webhook
func (r *WebhookRepository) ListDeliveriesForWebhook(
	webhookID uuid.UUID,
	status string,
	limit, offset int,
) ([]WebhookDelivery, int64, error) {
	var deliveries []WebhookDelivery
	var total int64

	query := r.db.Model(&WebhookDelivery{}).Where("webhook_id = ?", webhookID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}

	return deliveries, total, nil
}

// ListPendingDeliveries lists all pending deliveries that need to be sent
func (r *WebhookRepository) ListPendingDeliveries(limit int) ([]WebhookDelivery, error) {
	var deliveries []WebhookDelivery

	err := r.db.Where("status = ?", WebhookDeliveryStatusPending).
		Where("scheduled_at IS NULL OR scheduled_at <= ?", time.Now()).
		Limit(limit).
		Find(&deliveries).Error

	return deliveries, err
}

// ListRetriesDue lists deliveries that need retry
func (r *WebhookRepository) ListRetriesDue(limit int) ([]WebhookDelivery, error) {
	var deliveries []WebhookDelivery

	err := r.db.Where("status = ?", WebhookDeliveryStatusRetrying).
		Where("next_retry_at <= ?", time.Now()).
		Limit(limit).
		Find(&deliveries).Error

	return deliveries, err
}

// CleanOldDeliveries removes deliveries older than the specified time
func (r *WebhookRepository) CleanOldDeliveries(olderThan time.Time) error {
	return r.db.Where("created_at < ?", olderThan).
		Delete(&WebhookDelivery{}).Error
}

// GetDeliveryStats gets statistics for webhook deliveries
func (r *WebhookRepository) GetDeliveryStats(webhookID uuid.UUID) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	// Total deliveries
	var total int64
	r.db.Model(&WebhookDelivery{}).Where("webhook_id = ?", webhookID).Count(&total)
	stats["total_deliveries"] = total

	// Successful deliveries
	var delivered int64
	r.db.Model(&WebhookDelivery{}).Where("webhook_id = ? AND status = ?", webhookID, WebhookDeliveryStatusDelivered).Count(&delivered)
	stats["successful_deliveries"] = delivered

	// Failed deliveries
	var failed int64
	r.db.Model(&WebhookDelivery{}).Where("webhook_id = ? AND status = ?", webhookID, WebhookDeliveryStatusFailed).Count(&failed)
	stats["failed_deliveries"] = failed

	// Success rate
	if total > 0 {
		stats["success_rate"] = float64(delivered) / float64(total) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	return stats, nil
}
