package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// BillingOperationalRepository handles operational readiness for billing
type BillingOperationalRepository struct {
	db *gorm.DB
}

// NewBillingOperationalRepository creates a new billing operational repository
func NewBillingOperationalRepository(db *gorm.DB) *BillingOperationalRepository {
	return &BillingOperationalRepository{db: db}
}

// ==================== Stored Webhook Payloads ====================

// StoreWebhookPayload stores a raw webhook payload for replay capability
func (r *BillingOperationalRepository) StoreWebhookPayload(ctx context.Context, eventID, eventType string, payload json.RawMessage, signature string) (*StoredWebhookPayload, error) {
	// Calculate webhook secret hash for audit (first 8 chars of signature hash)
	secretHash := ""
	if signature != "" {
		hash := sha256.Sum256([]byte(signature))
		secretHash = hex.EncodeToString(hash[:])[:16]
	}

	stored := &StoredWebhookPayload{
		StripeEventID:     eventID,
		EventType:         eventType,
		Payload:           payload,
		Signature:         signature,
		ProcessingStatus:  "pending",
		Attempts:          0,
		WebhookSecretHash: secretHash,
		ExpiresAt:         time.Now().Add(30 * 24 * time.Hour), // 30-day retention
	}

	if err := r.db.WithContext(ctx).Create(stored).Error; err != nil {
		return nil, fmt.Errorf("failed to store webhook payload: %w", err)
	}

	return stored, nil
}

// GetStoredWebhookPayload retrieves a stored webhook payload by ID
func (r *BillingOperationalRepository) GetStoredWebhookPayload(ctx context.Context, id uuid.UUID) (*StoredWebhookPayload, error) {
	var payload StoredWebhookPayload
	if err := r.db.WithContext(ctx).First(&payload, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get webhook payload: %w", err)
	}
	return &payload, nil
}

// GetStoredWebhookPayloadByEventID retrieves a stored webhook payload by Stripe event ID
func (r *BillingOperationalRepository) GetStoredWebhookPayloadByEventID(ctx context.Context, eventID string) (*StoredWebhookPayload, error) {
	var payload StoredWebhookPayload
	if err := r.db.WithContext(ctx).Where("stripe_event_id = ?", eventID).First(&payload).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get webhook payload: %w", err)
	}
	return &payload, nil
}

// ListStoredWebhookPayloads lists stored webhook payloads with optional filtering
func (r *BillingOperationalRepository) ListStoredWebhookPayloads(ctx context.Context, status string, eventType string, limit, offset int) ([]StoredWebhookPayload, int64, error) {
	var payloads []StoredWebhookPayload
	var total int64

	query := r.db.WithContext(ctx).Model(&StoredWebhookPayload{})

	if status != "" {
		query = query.Where("processing_status = ?", status)
	}
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook payloads: %w", err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payloads).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook payloads: %w", err)
	}

	return payloads, total, nil
}

// UpdateWebhookPayloadStatus updates the processing status of a webhook payload
func (r *BillingOperationalRepository) UpdateWebhookPayloadStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"processing_status": status,
	}

	if status == "processed" {
		now := time.Now()
		updates["processed_at"] = &now
	}

	if errorMsg != "" {
		updates["processing_error"] = errorMsg
	}

	if err := r.db.WithContext(ctx).Model(&StoredWebhookPayload{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update webhook payload status: %w", err)
	}

	return nil
}

// MarkWebhookPayloadReplayed marks a webhook payload as replayed
func (r *BillingOperationalRepository) MarkWebhookPayloadReplayed(ctx context.Context, id uuid.UUID, replayedBy uuid.UUID, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"processing_status": "replayed",
		"replayed_at":       &now,
		"replayed_by":       replayedBy,
		"replay_reason":     reason,
	}

	if err := r.db.WithContext(ctx).Model(&StoredWebhookPayload{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to mark webhook payload as replayed: %w", err)
	}

	return nil
}

// CleanupExpiredWebhookPayloads removes expired webhook payloads (called by scheduled job)
func (r *BillingOperationalRepository) CleanupExpiredWebhookPayloads(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&StoredWebhookPayload{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired webhook payloads: %w", result.Error)
	}

	deletedCount := result.RowsAffected
	if deletedCount > 0 {
		logrus.WithField("deleted_count", deletedCount).Info("Cleaned up expired webhook payloads")
	}

	return deletedCount, nil
}

// IncrementWebhookPayloadAttempts increments the attempt counter for a webhook payload
func (r *BillingOperationalRepository) IncrementWebhookPayloadAttempts(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&StoredWebhookPayload{}).Where("id = ?", id).Update("attempts", gorm.Expr("attempts + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment webhook payload attempts: %w", err)
	}
	return nil
}

// ==================== Webhook Replay Requests ====================

// CreateWebhookReplayRequest creates a new webhook replay request
func (r *BillingOperationalRepository) CreateWebhookReplayRequest(ctx context.Context, payloadID uuid.UUID, requestedBy uuid.UUID, reason string) (*WebhookReplayRequest, error) {
	req := &WebhookReplayRequest{
		WebhookPayloadID: payloadID,
		RequestedBy:      requestedBy,
		Reason:           reason,
		Status:           "pending",
	}

	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook replay request: %w", err)
	}

	return req, nil
}

// GetWebhookReplayRequest retrieves a webhook replay request by ID
func (r *BillingOperationalRepository) GetWebhookReplayRequest(ctx context.Context, id uuid.UUID) (*WebhookReplayRequest, error) {
	var req WebhookReplayRequest
	if err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get webhook replay request: %w", err)
	}
	return &req, nil
}

// ListWebhookReplayRequests lists webhook replay requests
func (r *BillingOperationalRepository) ListWebhookReplayRequests(ctx context.Context, status string, limit, offset int) ([]WebhookReplayRequest, int64, error) {
	var requests []WebhookReplayRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&WebhookReplayRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook replay requests: %w", err)
	}

	if err := query.Order("requested_at DESC").Limit(limit).Offset(offset).Find(&requests).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook replay requests: %w", err)
	}

	return requests, total, nil
}

// CompleteWebhookReplayRequest marks a replay request as completed or failed
func (r *BillingOperationalRepository) CompleteWebhookReplayRequest(ctx context.Context, id uuid.UUID, status string, resultMessage string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":         status,
		"result_message": resultMessage,
		"completed_at":   &now,
	}

	if err := r.db.WithContext(ctx).Model(&WebhookReplayRequest{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to complete webhook replay request: %w", err)
	}

	return nil
}

// ==================== Tax Exemption Certificates ====================

// CreateTaxExemptionCertificate creates a new tax exemption certificate
func (r *BillingOperationalRepository) CreateTaxExemptionCertificate(ctx context.Context, cert *TaxExemptionCertificate) (*TaxExemptionCertificate, error) {
	if err := r.db.WithContext(ctx).Create(cert).Error; err != nil {
		return nil, fmt.Errorf("failed to create tax exemption certificate: %w", err)
	}
	return cert, nil
}

// GetTaxExemptionCertificate retrieves a tax exemption certificate by ID
func (r *BillingOperationalRepository) GetTaxExemptionCertificate(ctx context.Context, id uuid.UUID) (*TaxExemptionCertificate, error) {
	var cert TaxExemptionCertificate
	if err := r.db.WithContext(ctx).First(&cert, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tax exemption certificate: %w", err)
	}
	return &cert, nil
}

// ListTaxExemptionCertificates lists tax exemption certificates for a tenant
func (r *BillingOperationalRepository) ListTaxExemptionCertificates(ctx context.Context, tenantID uuid.UUID, status string) ([]TaxExemptionCertificate, error) {
	var certs []TaxExemptionCertificate

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, fmt.Errorf("failed to list tax exemption certificates: %w", err)
	}

	return certs, nil
}

// ListPendingTaxExemptionCertificates lists all pending certificates (for admin review)
func (r *BillingOperationalRepository) ListPendingTaxExemptionCertificates(ctx context.Context, limit, offset int) ([]TaxExemptionCertificate, int64, error) {
	var certs []TaxExemptionCertificate
	var total int64

	query := r.db.WithContext(ctx).Model(&TaxExemptionCertificate{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count pending certificates: %w", err)
	}

	if err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&certs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list pending certificates: %w", err)
	}

	return certs, total, nil
}

// ReviewTaxExemptionCertificate approves or rejects a tax exemption certificate
func (r *BillingOperationalRepository) ReviewTaxExemptionCertificate(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, approved bool, notes, rejectionReason string) (*TaxExemptionCertificate, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":           map[bool]string{true: "approved", false: "rejected"}[approved],
		"reviewed_by":      reviewerID,
		"reviewed_at":      &now,
		"review_notes":     notes,
		"rejection_reason": rejectionReason,
	}

	if err := r.db.WithContext(ctx).Model(&TaxExemptionCertificate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to review tax exemption certificate: %w", err)
	}

	return r.GetTaxExemptionCertificate(ctx, id)
}

// MarkTaxExemptionCertificateExpired marks expired certificates based on valid_until date
func (r *BillingOperationalRepository) MarkTaxExemptionCertificateExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&TaxExemptionCertificate{}).
		Where("status = ? AND valid_until < ?", "approved", now).
		Updates(map[string]interface{}{
			"status": "expired",
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark expired certificates: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// UpdateTaxExemptionStripeID updates the Stripe exemption ID after syncing
func (r *BillingOperationalRepository) UpdateTaxExemptionStripeID(ctx context.Context, id uuid.UUID, stripeID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"stripe_exemption_id":  stripeID,
		"applied_to_stripe_at": &now,
	}

	if err := r.db.WithContext(ctx).Model(&TaxExemptionCertificate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update tax exemption stripe ID: %w", err)
	}
	return nil
}

// ==================== EU VAT Validations ====================

// CreateEUVATValidation creates a new EU VAT validation record
func (r *BillingOperationalRepository) CreateEUVATValidation(ctx context.Context, validation *EUVATValidation) (*EUVATValidation, error) {
	if err := r.db.WithContext(ctx).Create(validation).Error; err != nil {
		return nil, fmt.Errorf("failed to create EU VAT validation: %w", err)
	}
	return validation, nil
}

// GetEUVATValidation retrieves an EU VAT validation by ID
func (r *BillingOperationalRepository) GetEUVATValidation(ctx context.Context, id uuid.UUID) (*EUVATValidation, error) {
	var validation EUVATValidation
	if err := r.db.WithContext(ctx).First(&validation, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get EU VAT validation: %w", err)
	}
	return &validation, nil
}

// GetEUVATValidationByVATID retrieves the latest EU VAT validation by VAT ID
func (r *BillingOperationalRepository) GetEUVATValidationByVATID(ctx context.Context, vatID string) (*EUVATValidation, error) {
	var validation EUVATValidation
	if err := r.db.WithContext(ctx).Where("vat_id = ?", vatID).Order("created_at DESC").First(&validation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get EU VAT validation: %w", err)
	}
	return &validation, nil
}

// ListEUVATValidations lists EU VAT validations for a tenant
func (r *BillingOperationalRepository) ListEUVATValidations(ctx context.Context, tenantID uuid.UUID, limit int) ([]EUVATValidation, error) {
	var validations []EUVATValidation
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Limit(limit).Find(&validations).Error; err != nil {
		return nil, fmt.Errorf("failed to list EU VAT validations: %w", err)
	}
	return validations, nil
}

// UpdateEUVATValidationStatus updates the status of a VAT validation
func (r *BillingOperationalRepository) UpdateEUVATValidationStatus(ctx context.Context, id uuid.UUID, status string, errorCode, errorMsg string) error {
	updates := map[string]interface{}{
		"status":        status,
		"error_code":    errorCode,
		"error_message": errorMsg,
	}

	if err := r.db.WithContext(ctx).Model(&EUVATValidation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update VAT validation status: %w", err)
	}
	return nil
}

// UpdateEUVATValidationVIESResponse updates the VIES response details
func (r *BillingOperationalRepository) UpdateEUVATValidationVIESResponse(ctx context.Context, id uuid.UUID, isValid bool, requestID, responseCode, traderName, traderAddress string) error {
	updates := map[string]interface{}{
		"is_valid":            isValid,
		"vies_request_id":     requestID,
		"vies_response_code":  responseCode,
		"vies_trader_name":    traderName,
		"vies_trader_address": traderAddress,
		"status":              map[bool]string{true: "valid", false: "invalid"}[isValid],
	}

	if err := r.db.WithContext(ctx).Model(&EUVATValidation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update VIES response: %w", err)
	}
	return nil
}

// ScheduleVATValidationRetry schedules a retry for a failed validation
func (r *BillingOperationalRepository) ScheduleVATValidationRetry(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error {
	updates := map[string]interface{}{
		"retry_count":   retryCount,
		"next_retry_at": &nextRetryAt,
	}

	if err := r.db.WithContext(ctx).Model(&EUVATValidation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to schedule VAT validation retry: %w", err)
	}
	return nil
}

// GetPendingVATValidations retrieves VAT validations that need retry
func (r *BillingOperationalRepository) GetPendingVATValidations(ctx context.Context, limit int) ([]EUVATValidation, error) {
	var validations []EUVATValidation
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("status IN (?, ?) AND (next_retry_at IS NULL OR next_retry_at <= ?)", "pending", "error", now).
		Where("retry_count < ?", 5). // Max 5 retries
		Order("created_at ASC").
		Limit(limit).
		Find(&validations).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending VAT validations: %w", err)
	}

	return validations, nil
}

// MarkVATValidationAppliedToSettings marks the validation as applied to tenant settings
func (r *BillingOperationalRepository) MarkVATValidationAppliedToSettings(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	updates := map[string]interface{}{
		"applied_to_settings": true,
		"applied_at":          &now,
	}

	if err := r.db.WithContext(ctx).Model(&EUVATValidation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to mark VAT validation as applied: %w", err)
	}
	return nil
}
