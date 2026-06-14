package privacy

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service provides privacy-related business logic
type Service struct {
	repo          *Repository
	anonymizer    *Anonymizer
	detector      *PIIDetector
	geoIP         *GeoIPService
	exportStorage ExportStorage
}

// NewService creates a new privacy service
func NewService(repo *Repository, salt string) *Service {
	svc := &Service{
		repo:       repo,
		anonymizer: NewAnonymizer(salt),
		detector:   NewPIIDetector(),
	}

	// Initialize GeoIP service if license key is available
	geoIPConfig := DefaultGeoIPConfig()
	if geoIPConfig.LicenseKey != "" {
		geoIPService, err := NewGeoIPService(geoIPConfig)
		if err != nil {
			logrus.WithError(err).Warn("Failed to initialize GeoIP service, using fallback region detection")
		} else {
			svc.geoIP = geoIPService
			logrus.Info("GeoIP service initialized for accurate region detection")
		}
	} else {
		logrus.Info("MAXMIND_LICENSE_KEY not set, using simplified region detection. " +
			"For accurate region detection, get a free key at https://www.maxmind.com/en/geolite2/signup")
	}

	// Initialize export storage (S3/R2 for production, local filesystem for dev)
	storage, err := NewExportStorage()
	if err != nil {
		logrus.WithError(err).Warn("Failed to initialize export storage, using local fallback")
		// Fallback to local storage
		storage, _ = NewLocalStorage("./exports", "")
	}
	svc.exportStorage = storage

	return svc
}

// SetExportStorage allows setting a custom export storage (useful for testing)
func (s *Service) SetExportStorage(storage ExportStorage) {
	s.exportStorage = storage
}

// SetGeoIPService allows setting a custom GeoIP service (useful for testing)
func (s *Service) SetGeoIPService(geoIP *GeoIPService) {
	s.geoIP = geoIP
}

// GetRegionFromIP returns a privacy-preserving region code
// Uses GeoIPService if available, otherwise falls back to simplified detection
func (s *Service) GetRegionFromIP(ip string) string {
	if s.geoIP != nil {
		return s.geoIP.GetRegionFromIP(ip)
	}
	return GetRegionFromIP(ip) // fallback to anonymizer.go function
}

// GetPrivacySettings retrieves privacy settings for a user
func (s *Service) GetPrivacySettings(userID uuid.UUID) (*PrivacySettings, error) {
	return s.repo.GetOrCreatePrivacySettings(userID)
}

// UpdatePrivacySettings updates a user's privacy settings
func (s *Service) UpdatePrivacySettings(userID, updaterID uuid.UUID, updates map[string]interface{}) (*PrivacySettings, error) {
	settings, err := s.repo.GetOrCreatePrivacySettings(userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if level, ok := updates["privacy_level"].(string); ok {
		settings.PrivacyLevel = PrivacyLevel(level)
	}
	if anonymizeIP, ok := updates["anonymize_ip"].(bool); ok {
		settings.AnonymizeIP = anonymizeIP
	}
	if anonymizeUA, ok := updates["anonymize_user_agent"].(bool); ok {
		settings.AnonymizeUserAgent = anonymizeUA
	}
	if logGeo, ok := updates["log_geo_data"].(bool); ok {
		settings.LogGeoData = logGeo
	}
	if logEmbed, ok := updates["log_embed_origin"].(bool); ok {
		settings.LogEmbedOrigin = logEmbed
	}
	if storeIO, ok := updates["store_input_output"].(bool); ok {
		settings.StoreInputOutput = storeIO
	}
	if retention, ok := updates["retention_days"].(int); ok {
		if retention >= 1 && retention <= 365 {
			settings.RetentionDays = retention
		}
	}
	if gdprMode, ok := updates["gdpr_mode"].(bool); ok {
		settings.GDPRMode = gdprMode
		if gdprMode {
			settings.PrivacyLevel = PrivacyLevelGDPR
		}
	}
	if autoDelete, ok := updates["auto_delete_enabled"].(bool); ok {
		settings.AutoDeleteEnabled = autoDelete
	}
	if ipMaskType, ok := updates["ip_mask_type"].(string); ok {
		settings.IPMaskType = PIIMaskType(ipMaskType)
	}
	if uaMaskType, ok := updates["user_agent_mask_type"].(string); ok {
		settings.UserAgentMaskType = PIIMaskType(uaMaskType)
	}

	settings.UpdatedBy = &updaterID

	if err := s.repo.UpdatePrivacySettings(settings); err != nil {
		return nil, err
	}

	// Log the privacy settings change
	s.repo.LogPrivacyEvent(
		"settings_updated",
		userID,
		nil,
		"",
		"",
		fmt.Sprintf("Privacy level changed to %s", settings.PrivacyLevel),
		true,
		"",
	)

	return settings, nil
}

// DeletePrivacySettings deletes a user's privacy settings
func (s *Service) DeletePrivacySettings(userID uuid.UUID) error {
	return s.repo.DeletePrivacySettings(userID)
}

// AnonymizeExecutionData anonymizes execution data based on privacy settings
func (s *Service) AnonymizeExecutionData(ip, userAgent, embedOrigin string, settings *PrivacySettings) (string, string, string) {
	if settings == nil {
		return ip, userAgent, embedOrigin
	}

	anonymizedIP := ip
	anonymizedUA := userAgent
	anonymizedOrigin := embedOrigin

	// Anonymize IP
	if settings.AnonymizeIP {
		anonymizedIP = s.anonymizer.AnonymizeIP(ip, settings.IPMaskType)
	}

	// Anonymize User Agent
	if settings.AnonymizeUserAgent {
		anonymizedUA = s.anonymizer.AnonymizeUserAgent(userAgent, settings.UserAgentMaskType)
	}

	// Anonymize embed origin in maximum/GDPR mode
	if settings.PrivacyLevel == PrivacyLevelMaximum || settings.PrivacyLevel == PrivacyLevelGDPR {
		anonymizedOrigin = s.anonymizer.AnonymizeEmbedOrigin(embedOrigin, PIIMaskTypePartial)
	}

	return anonymizedIP, anonymizedUA, anonymizedOrigin
}

// ShouldLogGeoData checks if geo data should be logged
func (s *Service) ShouldLogGeoData(settings *PrivacySettings) bool {
	if settings == nil {
		return true
	}
	return settings.LogGeoData
}

// ShouldLogEmbedOrigin checks if embed origin should be logged
func (s *Service) ShouldLogEmbedOrigin(settings *PrivacySettings) bool {
	if settings == nil {
		return true
	}
	return settings.LogEmbedOrigin
}

// ShouldStoreInputOutput checks if input/output should be stored
func (s *Service) ShouldStoreInputOutput(settings *PrivacySettings) bool {
	if settings == nil {
		return true
	}
	return settings.StoreInputOutput
}

// ScanForPII scans data for PII and optionally redacts it
func (s *Service) ScanForPII(data interface{}, redact bool) (*PIIDetectionResult, error) {
	s.detector.SetRedactMode(redact)
	result := s.detector.DetectPII(data)
	return result, nil
}

// SanitizeInputOutput sanitizes input/output data for storage
func (s *Service) SanitizeInputOutput(input, output []byte) ([]byte, []byte, bool) {
	globalSettings, err := s.repo.GetGlobalPrivacySettings()
	if err != nil {
		// Default to not sanitizing if we can't get settings
		return input, output, false
	}

	if !globalSettings.PIIScanningEnabled {
		return input, output, false
	}

	sanitizedInput := input
	sanitizedOutput := output
	hasPII := false

	// Scan input
	if len(input) > 0 {
		result, err := s.detector.DetectPIIInJSON(input)
		if err == nil && result.HasPII {
			hasPII = true
			if globalSettings.InputOutputRedaction {
				sanitizedInput = []byte(result.RedactedData)
			}
		}
	}

	// Scan output
	if len(output) > 0 {
		result, err := s.detector.DetectPIIInJSON(output)
		if err == nil && result.HasPII {
			hasPII = true
			if globalSettings.InputOutputRedaction {
				sanitizedOutput = []byte(result.RedactedData)
			}
		}
	}

	return sanitizedInput, sanitizedOutput, hasPII
}

// RecordConsent records a user's consent
func (s *Service) RecordConsent(userID uuid.UUID, consentType, consentVersion, consentText, ipHash, uaHash string, given bool) (*PrivacyConsentRecord, error) {
	record := &PrivacyConsentRecord{
		UserID:        userID,
		ConsentType:   consentType,
		ConsentGiven:  given,
		ConsentVersion: consentVersion,
		ConsentText:   consentText,
		IPHash:        ipHash,
		UserAgentHash: uaHash,
		GivenAt:       time.Now(),
	}

	if err := s.repo.RecordConsent(record); err != nil {
		return nil, err
	}

	// Log the consent event
	s.repo.LogPrivacyEvent(
		"consent_given",
		userID,
		&record.ID,
		"",
		"",
		fmt.Sprintf("Consent %s for %s (version %s)", map[bool]string{true: "given", false: "withdrawn"}[given], consentType, consentVersion),
		true,
		"",
	)

	return record, nil
}

// WithdrawConsent withdraws a user's consent
func (s *Service) WithdrawConsent(userID uuid.UUID, consentType, reason string) error {
	if err := s.repo.WithdrawConsent(userID, consentType, reason); err != nil {
		return err
	}

	// Log the consent withdrawal
	s.repo.LogPrivacyEvent(
		"consent_withdrawn",
		userID,
		nil,
		"",
		"",
		fmt.Sprintf("Consent withdrawn for %s: %s", consentType, reason),
		true,
		"",
	)

	return nil
}

// HasActiveConsent checks if a user has active consent
func (s *Service) HasActiveConsent(userID uuid.UUID, consentType string) bool {
	return s.repo.HasActiveConsent(userID, consentType)
}

// GetConsentRecords retrieves consent records for a user
func (s *Service) GetConsentRecords(userID uuid.UUID, consentType string) ([]PrivacyConsentRecord, error) {
	return s.repo.GetConsentRecords(userID, consentType)
}

// GetGlobalPrivacySettings retrieves global privacy settings
func (s *Service) GetGlobalPrivacySettings() (*GlobalPrivacySettings, error) {
	return s.repo.GetGlobalPrivacySettings()
}

// UpdateGlobalPrivacySettings updates global privacy settings
func (s *Service) UpdateGlobalPrivacySettings(settings *GlobalPrivacySettings) error {
	return s.repo.UpdateGlobalPrivacySettings(settings)
}

// RequestDataExport initiates a GDPR data export request
func (s *Service) RequestDataExport(ctx context.Context, userID, tenantID uuid.UUID, requestType string) (*DataExportRequest, error) {
	// Check for existing pending requests
	existingRequests, err := s.repo.GetDataExportRequests(userID)
	if err != nil {
		return nil, err
	}

	for _, req := range existingRequests {
		if req.Status == "pending" || req.Status == "processing" {
			return nil, fmt.Errorf("an export request is already in progress")
		}
	}

	request, err := s.repo.CreateDataExportRequest(userID, tenantID, requestType)
	if err != nil {
		return nil, err
	}

	// Log the export request
	s.repo.LogPrivacyEvent(
		"export_requested",
		userID,
		&request.ID,
		"",
		"",
		fmt.Sprintf("Data export requested: %s", requestType),
		true,
		"",
	)

	// Start async processing
	go s.processDataExport(ctx, request.ID, userID, requestType)

	return request, nil
}

// processDataExport processes a data export request asynchronously
func (s *Service) processDataExport(ctx context.Context, requestID, userID uuid.UUID, requestType string) {
	// Update status to processing
	s.repo.UpdateExportRequestStatus(requestID, "processing", "", "", 0, 0, "")

	// Build the data package
	dataPackage, err := s.buildGDPRDataPackage(ctx, userID, requestType)
	if err != nil {
		s.repo.UpdateExportRequestStatus(requestID, "failed", "", "", 0, 0, err.Error())
		s.repo.LogPrivacyEvent(
			"export_failed",
			userID,
			&requestID,
			"",
			"",
			"",
			false,
			err.Error(),
		)
		return
	}

	// Create zip archive
	zipData, recordCount, err := s.createExportArchive(dataPackage)
	if err != nil {
		s.repo.UpdateExportRequestStatus(requestID, "failed", "", "", 0, 0, err.Error())
		s.repo.LogPrivacyEvent(
			"export_failed",
			userID,
			&requestID,
			"",
			"",
			"",
			false,
			err.Error(),
		)
		return
	}

	// Upload to secure storage (S3 with pre-signed URL for secure download)
	downloadURL, downloadToken, err := s.storeExportFile(requestID, zipData)
	if err != nil {
		s.repo.UpdateExportRequestStatus(requestID, "failed", "", "", 0, 0, err.Error())
		return
	}

	// Update request as completed
	fileSize := int64(len(zipData))
	s.repo.UpdateExportRequestStatus(requestID, "completed", downloadURL, downloadToken, fileSize, recordCount, "")

	// Log completion
	s.repo.LogPrivacyEvent(
		"export_completed",
		userID,
		&requestID,
		"",
		"",
		fmt.Sprintf("Export completed with %d records", recordCount),
		true,
		"",
	)
}

// buildGDPRDataPackage builds a complete data package for export
func (s *Service) buildGDPRDataPackage(ctx context.Context, userID uuid.UUID, requestType string) (*GDPRDataPackage, error) {
	package_ := &GDPRDataPackage{
		ExportID:    uuid.New().String(),
		UserID:      userID.String(),
		GeneratedAt: time.Now(),
		Version:     "1.0",
	}

	// Get user profile
	user, err := s.repo.db.GetUserByID(ctx, userID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get user profile for export")
	} else {
		package_.Profile = map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"username":       user.Username,
			"name":           user.Name,
			"created_at":     user.CreatedAt,
			"email_verified": user.EmailVerified,
		}
	}

	// Get executions based on request type
	if requestType == "full" || requestType == "executions" {
		// Count first
		count, err := s.repo.CountUserExecutions(userID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to count user executions")
		}

		// Get in batches
		batchSize := 1000
		for offset := 0; offset < int(count); offset += batchSize {
			executions, err := s.repo.GetUserExecutionsForExport(userID, batchSize, offset)
			if err != nil {
				logrus.WithError(err).Warn("Failed to get executions for export")
				break
			}
			package_.Executions = append(package_.Executions, executions...)
		}
	}

	// Get audit logs
	if requestType == "full" || requestType == "audit" {
		count, err := s.repo.CountUserAuditLogs(userID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to count audit logs")
		}

		batchSize := 1000
		for offset := 0; offset < int(count); offset += batchSize {
			logs, err := s.repo.GetUserAuditLogsForExport(userID, batchSize, offset)
			if err != nil {
				logrus.WithError(err).Warn("Failed to get audit logs for export")
				break
			}
			package_.AuditLogs = append(package_.AuditLogs, logs...)
		}
	}

	// Get consent records
	consentRecords, err := s.repo.GetUserConsentRecordsForExport(userID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get consent records for export")
	} else {
		package_.ConsentRecords = consentRecords
	}

	return package_, nil
}

// createExportArchive creates a zip archive of the data package
func (s *Service) createExportArchive(dataPackage *GDPRDataPackage) ([]byte, int64, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add manifest
	manifest := map[string]interface{}{
		"export_id":     dataPackage.ExportID,
		"user_id":       dataPackage.UserID,
		"generated_at":  dataPackage.GeneratedAt,
		"version":       dataPackage.Version,
		"record_counts": map[string]int{
			"profile":          1,
			"executions":       len(dataPackage.Executions),
			"audit_logs":       len(dataPackage.AuditLogs),
			"consent_records":  len(dataPackage.ConsentRecords),
		},
	}

	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	manifestFile, _ := zipWriter.Create("manifest.json")
	manifestFile.Write(manifestData)

	// Add profile
	if dataPackage.Profile != nil {
		profileData, _ := json.MarshalIndent(dataPackage.Profile, "", "  ")
		profileFile, _ := zipWriter.Create("profile.json")
		profileFile.Write(profileData)
	}

	// Add executions
	if len(dataPackage.Executions) > 0 {
		execData, _ := json.MarshalIndent(dataPackage.Executions, "", "  ")
		execFile, _ := zipWriter.Create("executions.json")
		execFile.Write(execData)
	}

	// Add audit logs
	if len(dataPackage.AuditLogs) > 0 {
		auditData, _ := json.MarshalIndent(dataPackage.AuditLogs, "", "  ")
		auditFile, _ := zipWriter.Create("audit_logs.json")
		auditFile.Write(auditData)
	}

	// Add consent records
	if len(dataPackage.ConsentRecords) > 0 {
		consentData, _ := json.MarshalIndent(dataPackage.ConsentRecords, "", "  ")
		consentFile, _ := zipWriter.Create("consent_records.json")
		consentFile.Write(consentData)
	}

	zipWriter.Close()

	recordCount := int64(len(dataPackage.Executions) + len(dataPackage.AuditLogs) + len(dataPackage.ConsentRecords))
	if dataPackage.Profile != nil {
		recordCount++
	}

	return buf.Bytes(), recordCount, nil
}

// storeExportFile stores the export file and returns download info
// Uses S3-compatible storage (S3, R2, MinIO) in production, local filesystem for development
func (s *Service) storeExportFile(requestID uuid.UUID, data []byte) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Upload to configured storage (S3/R2 or local)
	downloadURL, err := s.exportStorage.Upload(ctx, requestID, data, "application/zip")
	if err != nil {
		return "", "", fmt.Errorf("failed to store export: %w", err)
	}

	// Generate a secure download token for verification
	downloadToken := s.anonymizer.hashString(requestID.String() + time.Now().String())

	logrus.WithFields(logrus.Fields{
		"request_id":     requestID,
		"download_url":   downloadURL,
		"size":           len(data),
		"storage_type":   fmt.Sprintf("%T", s.exportStorage),
	}).Info("Export file stored successfully")

	return downloadURL, downloadToken[:32], nil
}

// RequestDataDeletion initiates a GDPR right-to-erasure request
func (s *Service) RequestDataDeletion(ctx context.Context, userID, tenantID uuid.UUID, requestType string) (*DataDeletionRequest, error) {
	request, err := s.repo.CreateDataDeletionRequest(userID, tenantID, requestType)
	if err != nil {
		return nil, err
	}

	// Log the deletion request
	s.repo.LogPrivacyEvent(
		"deletion_requested",
		userID,
		&request.ID,
		"",
		"",
		fmt.Sprintf("Data deletion requested: %s", requestType),
		true,
		"",
	)

	// Start async processing
	go s.processDataDeletion(ctx, request.ID, userID, tenantID, requestType)

	return request, nil
}

// processDataDeletion processes a data deletion request
func (s *Service) processDataDeletion(ctx context.Context, requestID, userID, tenantID uuid.UUID, requestType string) {
	var totalDeleted, totalAnonymized int64

	// Update status to processing
	s.repo.UpdateDeletionRequestStatus(requestID, "processing", 0, 0, "", "")

	// Delete executions
	if requestType == "full" || requestType == "executions" {
		// Check privacy settings for deletion preference
		settings, err := s.repo.GetOrCreatePrivacySettings(userID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get privacy settings for deletion")
		}

		if settings != nil && settings.PrivacyLevel == PrivacyLevelMaximum {
			// Hard delete for maximum privacy
			deleted, err := s.repo.DeleteUserExecutions(userID)
			if err != nil {
				logrus.WithError(err).Error("Failed to delete user executions")
			}
			totalDeleted += deleted
		} else {
			// Anonymize for other privacy levels (keep aggregate data)
			anonymized, err := s.repo.AnonymizeUserExecutions(userID)
			if err != nil {
				logrus.WithError(err).Error("Failed to anonymize user executions")
			}
			totalAnonymized += anonymized
		}
	}

	// Delete audit logs (with legal hold check in production)
	if requestType == "full" || requestType == "audit_logs" {
		deleted, err := s.repo.DeleteUserAuditLogs(userID, false)
		if err != nil {
			logrus.WithError(err).Error("Failed to delete user audit logs")
		}
		totalDeleted += deleted
	}

	// Generate verification hash
	verificationHash := s.anonymizer.hashString(fmt.Sprintf("%s:%d:%d:%s", requestID, totalDeleted, totalAnonymized, time.Now()))

	// Determine final status
	status := "completed"
	if totalDeleted == 0 && totalAnonymized == 0 {
		// Check if there was any data to delete
		execCount, _ := s.repo.CountUserExecutions(userID)
		auditCount, _ := s.repo.CountUserAuditLogs(userID)
		if execCount > 0 || auditCount > 0 {
			status = "failed"
		}
	} else if totalDeleted > 0 && totalAnonymized > 0 {
		status = "partial"
	}

	// Update request as completed
	s.repo.UpdateDeletionRequestStatus(requestID, status, totalDeleted, totalAnonymized, "", verificationHash)

	// Log completion
	s.repo.LogPrivacyEvent(
		"deletion_completed",
		userID,
		&requestID,
		"",
		"",
		fmt.Sprintf("Data deletion completed: %d deleted, %d anonymized", totalDeleted, totalAnonymized),
		status != "failed",
		"",
	)
}

// GetDataDeletionStatus gets the status of a deletion request
func (s *Service) GetDataDeletionStatus(requestID uuid.UUID) (*DataDeletionRequest, error) {
	return s.repo.GetDataDeletionRequest(requestID)
}

// GetDataExportStatus gets the status of an export request
func (s *Service) GetDataExportStatus(requestID uuid.UUID) (*DataExportRequest, error) {
	return s.repo.GetDataExportRequest(requestID)
}

// GetAllExportRequests gets all export requests (admin only)
func (s *Service) GetAllExportRequests() ([]DataExportRequest, error) {
	return s.repo.GetAllDataExportRequests()
}

// GetAllDeletionRequests gets all deletion requests (admin only)
func (s *Service) GetAllDeletionRequests() ([]DataDeletionRequest, error) {
	return s.repo.GetAllDataDeletionRequests()
}

// GetAllPrivacyAuditLogs gets all privacy audit logs (admin only)
func (s *Service) GetAllPrivacyAuditLogs(limit, offset int) ([]PrivacyAuditLog, int64, error) {
	return s.repo.GetAllPrivacyAuditLogs(limit, offset)
}

// RunPeriodicAnonymization runs the periodic anonymization task
func (s *Service) RunPeriodicAnonymization(ctx context.Context) error {
	return s.repo.RunPeriodicAnonymization(ctx)
}

// GetPrivacyHeaders extracts privacy-related headers from request
func (s *Service) GetPrivacyHeaders(headers map[string]string) *PrivacyHeaders {
	privacyHeaders := &PrivacyHeaders{
		DoNotTrack:      false,
		GDPRApplies:     false,
		CCPAApplies:     false,
		ConsentGiven:    false,
		PrivacyLevel:    "standard",
		RequestAnonymization: false,
	}

	// Check DNT header
	if dnt, ok := headers["Dnt"]; ok && (dnt == "1" || dnt == "true") {
		privacyHeaders.DoNotTrack = true
	}

	// Check GDPR applies header
	if gdpr, ok := headers["Gdpr"]; ok && (gdpr == "1" || gdpr == "true") {
		privacyHeaders.GDPRApplies = true
	}

	// Check CCPA applies header
	if ccpa, ok := headers["Ccpa"]; ok && (ccpa == "1" || ccpa == "true") {
		privacyHeaders.CCPAApplies = true
	}

	// Check consent header
	if consent, ok := headers["Consent"]; ok && (consent == "given" || consent == "true") {
		privacyHeaders.ConsentGiven = true
	}

	// Check privacy level header
	if level, ok := headers["Privacy-Level"]; ok {
		privacyHeaders.PrivacyLevel = level
	}

	return privacyHeaders
}
