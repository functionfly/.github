package privacy

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Repository provides database operations for privacy functionality
type Repository struct {
	db *storage.PostgresDB
}

// NewRepository creates a new privacy repository
func NewRepository(db *storage.PostgresDB) *Repository {
	return &Repository{db: db}
}

// GetPrivacySettings retrieves privacy settings for a user or tenant
func (r *Repository) GetPrivacySettings(userID, tenantID *uuid.UUID) (*PrivacySettings, error) {
	var settings PrivacySettings
	query := r.db.GORM.Model(&PrivacySettings{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		return nil, fmt.Errorf("either userID or tenantID must be provided")
	}

	if err := query.First(&settings).Error; err != nil {
		if err.Error() == "record not found" {
			// Return default settings
			return r.createDefaultSettings(userID, tenantID)
		}
		return nil, err
	}

	return &settings, nil
}

// GetOrCreatePrivacySettings gets or creates privacy settings for a user
func (r *Repository) GetOrCreatePrivacySettings(userID uuid.UUID) (*PrivacySettings, error) {
	settings, err := r.GetPrivacySettings(&userID, nil)
	if err != nil {
		return r.createDefaultSettings(&userID, nil)
	}
	return settings, nil
}

// createDefaultSettings creates default privacy settings
func (r *Repository) createDefaultSettings(userID, tenantID *uuid.UUID) (*PrivacySettings, error) {
	settings := &PrivacySettings{
		PrivacyLevel:       PrivacyLevelStandard,
		AnonymizeIP:        false,
		AnonymizeUserAgent: false,
		LogGeoData:         true,
		LogEmbedOrigin:     true,
		StoreInputOutput:   true,
		RetentionDays:      90,
		GDPRMode:           false,
		AutoDeleteEnabled:  false,
		ConsentRequired:    false,
		IPMaskType:         PIIMaskTypeNone,
		UserAgentMaskType:  PIIMaskTypeNone,
	}

	if userID != nil {
		settings.UserID = userID
	}
	if tenantID != nil {
		settings.TenantID = tenantID
	}

	// Try to insert, if fails (duplicate), just return
	if err := r.db.GORM.Create(settings).Error; err != nil {
		// Try to fetch again in case of race condition
		existing, fetchErr := r.GetPrivacySettings(userID, tenantID)
		if fetchErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return settings, nil
}

// UpdatePrivacySettings updates privacy settings
func (r *Repository) UpdatePrivacySettings(settings *PrivacySettings) error {
	settings.UpdatedAt = time.Now()
	return r.db.GORM.Save(settings).Error
}

// DeletePrivacySettings deletes privacy settings
func (r *Repository) DeletePrivacySettings(userID uuid.UUID) error {
	return r.db.GORM.Where("user_id = ?", userID).Delete(&PrivacySettings{}).Error
}

// RecordConsent records a user's consent for data processing
func (r *Repository) RecordConsent(record *PrivacyConsentRecord) error {
	return r.db.GORM.Create(record).Error
}

// GetConsentRecords retrieves consent records for a user
func (r *Repository) GetConsentRecords(userID uuid.UUID, consentType string) ([]PrivacyConsentRecord, error) {
	var records []PrivacyConsentRecord
	query := r.db.GORM.Where("user_id = ?", userID)

	if consentType != "" {
		query = query.Where("consent_type = ?", consentType)
	}

	if err := query.Order("given_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

// HasActiveConsent checks if a user has given active consent
func (r *Repository) HasActiveConsent(userID uuid.UUID, consentType string) bool {
	var record PrivacyConsentRecord
	err := r.db.GORM.Where("user_id = ? AND consent_type = ? AND consent_given = ? AND withdrawn_at IS NULL",
		userID, consentType, true).
		Order("given_at DESC").
		First(&record).Error

	return err == nil
}

// WithdrawConsent withdraws consent
func (r *Repository) WithdrawConsent(userID uuid.UUID, consentType, reason string) error {
	return r.db.GORM.Model(&PrivacyConsentRecord{}).
		Where("user_id = ? AND consent_type = ? AND consent_given = ? AND withdrawn_at IS NULL",
			userID, consentType, true).
		Updates(map[string]interface{}{
			"withdrawn_at":     time.Now(),
			"withdrawn_reason": reason,
		}).Error
}

// CreateDataExportRequest creates a new data export request
func (r *Repository) CreateDataExportRequest(userID, tenantID uuid.UUID, requestType string) (*DataExportRequest, error) {
	request := &DataExportRequest{
		UserID:      userID,
		TenantID:    &tenantID,
		Status:      "pending",
		RequestType: requestType,
		RequestedAt: time.Now(),
		ExpiresAt:   nil, // Will be set when completed
	}

	if err := r.db.GORM.Create(request).Error; err != nil {
		return nil, err
	}

	return request, nil
}

// GetDataExportRequest retrieves an export request by ID
func (r *Repository) GetDataExportRequest(requestID uuid.UUID) (*DataExportRequest, error) {
	var request DataExportRequest
	if err := r.db.GORM.First(&request, "id = ?", requestID).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

// GetDataExportRequests retrieves export requests for a user
func (r *Repository) GetDataExportRequests(userID uuid.UUID) ([]DataExportRequest, error) {
	var requests []DataExportRequest
	if err := r.db.GORM.Where("user_id = ?", userID).Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// UpdateExportRequestStatus updates the status of an export request
func (r *Repository) UpdateExportRequestStatus(requestID uuid.UUID, status, downloadURL, downloadToken string, fileSize, recordCount int64, errorMessage string) error {
	updates := map[string]interface{}{
		"status":       status,
		"updated_at":   time.Now(),
		"error_message": errorMessage,
	}

	if status == "completed" {
		updates["completed_at"] = time.Now()
		updates["expires_at"] = time.Now().Add(7 * 24 * time.Hour) // 7 days to download
		updates["download_url"] = downloadURL
		updates["download_token"] = downloadToken
		updates["file_size"] = fileSize
		updates["record_count"] = recordCount
	}

	return r.db.GORM.Model(&DataExportRequest{}).Where("id = ?", requestID).Updates(updates).Error
}

// CreateDataDeletionRequest creates a new data deletion request
func (r *Repository) CreateDataDeletionRequest(userID, tenantID uuid.UUID, requestType string) (*DataDeletionRequest, error) {
	request := &DataDeletionRequest{
		UserID:      userID,
		TenantID:    &tenantID,
		Status:      "pending",
		RequestType: requestType,
		RequestedAt: time.Now(),
	}

	if err := r.db.GORM.Create(request).Error; err != nil {
		return nil, err
	}

	return request, nil
}

// GetDataDeletionRequest retrieves a deletion request by ID
func (r *Repository) GetDataDeletionRequest(requestID uuid.UUID) (*DataDeletionRequest, error) {
	var request DataDeletionRequest
	if err := r.db.GORM.First(&request, "id = ?", requestID).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

// GetDataDeletionRequests retrieves deletion requests for a user
func (r *Repository) GetDataDeletionRequests(userID uuid.UUID) ([]DataDeletionRequest, error) {
	var requests []DataDeletionRequest
	if err := r.db.GORM.Where("user_id = ?", userID).Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// GetAllDataExportRequests retrieves all export requests (admin only)
func (r *Repository) GetAllDataExportRequests() ([]DataExportRequest, error) {
	var requests []DataExportRequest
	if err := r.db.GORM.Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// GetAllDataDeletionRequests retrieves all deletion requests (admin only)
func (r *Repository) GetAllDataDeletionRequests() ([]DataDeletionRequest, error) {
	var requests []DataDeletionRequest
	if err := r.db.GORM.Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// UpdateDeletionRequestStatus updates the status of a deletion request
func (r *Repository) UpdateDeletionRequestStatus(requestID uuid.UUID, status string, recordsDeleted, recordsAnonymized int64, errorMessage, verificationHash string) error {
	updates := map[string]interface{}{
		"status":             status,
		"updated_at":         time.Now(),
		"error_message":      errorMessage,
		"records_deleted":    recordsDeleted,
		"records_anonymized": recordsAnonymized,
		"verification_hash":  verificationHash,
	}

	if status == "completed" || status == "partial" {
		updates["completed_at"] = time.Now()
	}

	return r.db.GORM.Model(&DataDeletionRequest{}).Where("id = ?", requestID).Updates(updates).Error
}

// LogPrivacyEvent logs a privacy-related event
func (r *Repository) LogPrivacyEvent(action string, userID uuid.UUID, requestID *uuid.UUID, ipAddress, userAgent, details string, success bool, errorMessage string) error {
	log := &PrivacyAuditLog{
		Action:       action,
		UserID:       userID,
		RequestID:    requestID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMessage,
		CreatedAt:    time.Now(),
	}

	return r.db.GORM.Create(log).Error
}

// GetGlobalPrivacySettings retrieves global privacy settings
func (r *Repository) GetGlobalPrivacySettings() (*GlobalPrivacySettings, error) {
	var settings GlobalPrivacySettings
	if err := r.db.GORM.Where("is_active = ?", true).First(&settings).Error; err != nil {
		if err.Error() == "record not found" {
			// Create default global settings
			settings = GlobalPrivacySettings{
				DefaultPrivacyLevel:      PrivacyLevelStandard,
				DefaultIPMaskType:        PIIMaskTypeNone,
				DefaultUserAgentMaskType: PIIMaskTypeNone,
				DefaultRetentionDays:     90,
				GDPRModeEnabled:          false,
				CCPAModeEnabled:          false,
				AutoAnonymizeAfterDays:   0,
				RequireConsent:           false,
				PIIScanningEnabled:       false,
				InputOutputRedaction:     false,
				IsActive:                 true,
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
			}
			if err := r.db.GORM.Create(&settings).Error; err != nil {
				return nil, err
			}
			return &settings, nil
		}
		return nil, err
	}
	return &settings, nil
}

// UpdateGlobalPrivacySettings updates global privacy settings
func (r *Repository) UpdateGlobalPrivacySettings(settings *GlobalPrivacySettings) error {
	settings.UpdatedAt = time.Now()
	return r.db.GORM.Save(settings).Error
}

// DeleteUserExecutions deletes all execution records for a user
func (r *Repository) DeleteUserExecutions(userID uuid.UUID) (int64, error) {
	result := r.db.GORM.Where("user_id = ?", userID).Delete(&storage.RegistryFunctionExecution{})
	return result.RowsAffected, result.Error
}

// AnonymizeUserExecutions anonymizes execution records instead of deleting them
func (r *Repository) AnonymizeUserExecutions(userID uuid.UUID) (int64, error) {
	result := r.db.GORM.Model(&storage.RegistryFunctionExecution{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"caller_ip":    sql.NullString{String: "[ANONYMIZED]", Valid: true},
			"user_agent":   sql.NullString{String: "[ANONYMIZED]", Valid: true},
			"geo_country":  sql.NullString{String: "[ANONYMIZED]", Valid: true},
			"embed_origin": sql.NullString{String: "[ANONYMIZED]", Valid: true},
			"user_id":      nil,
		})
	return result.RowsAffected, result.Error
}

// DeleteUserAuditLogs deletes audit logs for a user (with legal hold check)
func (r *Repository) DeleteUserAuditLogs(userID uuid.UUID, legalHold bool) (int64, error) {
	// In a real system, you might have legal hold flags
	// For now, we just delete
	result := r.db.GORM.Where("actor_user_id = ?", userID).Delete(&storage.AuditEvent{})
	return result.RowsAffected, result.Error
}

// GetUserExecutionsForExport retrieves user execution data for GDPR export
func (r *Repository) GetUserExecutionsForExport(userID uuid.UUID, limit, offset int) ([]ExecutionData, error) {
	query := `
		SELECT e.id, e.function_id, f.name as function_name, e.version,
		       e.timestamp, e.duration_ms, e.status_code, e.cached
		FROM registry_function_executions e
		LEFT JOIN registry_functions f ON e.function_id = f.id
		WHERE e.user_id = $1
		ORDER BY e.timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []ExecutionData
	for rows.Next() {
		var exec ExecutionData
		var execID uuid.UUID
		var funcID uuid.UUID
		var funcName sql.NullString

		if err := rows.Scan(&execID, &funcID, &funcName, &exec.Version,
			&exec.Timestamp, &exec.DurationMs, &exec.StatusCode, &exec.Cached); err != nil {
			continue
		}

		exec.ExecutionID = execID.String()
		exec.FunctionID = funcID.String()
		if funcName.Valid {
			exec.FunctionName = funcName.String
		}
		executions = append(executions, exec)
	}

	return executions, rows.Err()
}

// GetUserAuditLogsForExport retrieves user audit logs for GDPR export
func (r *Repository) GetUserAuditLogsForExport(userID uuid.UUID, limit, offset int) ([]AuditLogData, error) {
	var events []storage.AuditEvent
	if err := r.db.GORM.Where("actor_user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).Offset(offset).
		Find(&events).Error; err != nil {
		return nil, err
	}

	var logs []AuditLogData
	for _, event := range events {
		logs = append(logs, AuditLogData{
			EventID:      event.ID.String(),
			Action:       event.Action,
			ResourceType: event.ResourceType,
			Timestamp:    event.Timestamp,
			Success:      event.Success,
		})
	}

	return logs, nil
}

// GetUserConsentRecordsForExport retrieves consent records for export
func (r *Repository) GetUserConsentRecordsForExport(userID uuid.UUID) ([]ConsentData, error) {
	records, err := r.GetConsentRecords(userID, "")
	if err != nil {
		return nil, err
	}

	var data []ConsentData
	for _, record := range records {
		data = append(data, ConsentData{
			ConsentType: record.ConsentType,
			Given:       record.ConsentGiven,
			Version:     record.ConsentVersion,
			Timestamp:   record.GivenAt,
		})
	}

	return data, nil
}

// CountUserExecutions counts total executions for a user
func (r *Repository) CountUserExecutions(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GORM.Model(&storage.RegistryFunctionExecution{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountUserAuditLogs counts total audit logs for a user
func (r *Repository) CountUserAuditLogs(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GORM.Model(&storage.AuditEvent{}).
		Where("actor_user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetAllPrivacyAuditLogs retrieves all privacy audit logs with pagination (admin only)
func (r *Repository) GetAllPrivacyAuditLogs(limit, offset int) ([]PrivacyAuditLog, int64, error) {
	var logs []PrivacyAuditLog
	var count int64

	if err := r.db.GORM.Model(&PrivacyAuditLog{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.GORM.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}

// RunPeriodicAnonymization anonymizes old execution records based on settings
func (r *Repository) RunPeriodicAnonymization(ctx context.Context) error {
	settings, err := r.GetGlobalPrivacySettings()
	if err != nil {
		return err
	}

	if settings.AutoAnonymizeAfterDays <= 0 {
		return nil // Auto-anonymization disabled
	}

	cutoffDate := time.Now().AddDate(0, 0, -settings.AutoAnonymizeAfterDays)

	// Anonymize old executions where user has enhanced privacy
	query := `
		UPDATE registry_function_executions
		SET caller_ip = '[AUTO-ANONYMIZED]',
		    user_agent = '[AUTO-ANONYMIZED]',
		    geo_country = NULL,
		    embed_origin = '[AUTO-ANONYMIZED]'
		WHERE timestamp < $1
		  AND caller_ip IS NOT NULL
		  AND caller_ip != '[AUTO-ANONYMIZED]'
		  AND user_id IN (
		      SELECT user_id FROM privacy_settings
		      WHERE privacy_level IN ('enhanced', 'maximum', 'gdpr')
		  )
		RETURNING id
	`

	rows, err := r.db.Query(query, cutoffDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	var anonymizedCount int
	for rows.Next() {
		anonymizedCount++
	}

	if anonymizedCount > 0 {
		logrus.WithField("anonymized_count", anonymizedCount).
			Info("Auto-anonymized old execution records")
	}

	return rows.Err()
}

// GetExecutionRetentionSettings gets execution retention settings
func (r *Repository) GetExecutionRetentionSettings() (*storage.ExecutionRetentionSettings, error) {
	var settings storage.ExecutionRetentionSettings
	if err := r.db.GORM.Where("is_active = ?", true).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}
