package privacy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/privacy"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles privacy-related API endpoints
type Handler struct {
	service *privacy.Service
	authSvc *auth.AuthService
}

// NewHandler creates a new privacy handler
func NewHandler(service *privacy.Service, authSvc *auth.AuthService) *Handler {
	return &Handler{
		service: service,
		authSvc: authSvc,
	}
}

// HandleGetPrivacySettings handles GET /v1/privacy/settings
func (h *Handler) HandleGetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	settings, err := h.service.GetPrivacySettings(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get privacy settings", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, settings)
}

// HandleUpdatePrivacySettings handles PUT /v1/privacy/settings
func (h *Handler) HandleUpdatePrivacySettings(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	settings, err := h.service.UpdatePrivacySettings(userID, userID, updates)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update privacy settings", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, settings)
}

// HandleDeletePrivacySettings handles DELETE /v1/privacy/settings
func (h *Handler) HandleDeletePrivacySettings(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	if err := h.service.DeletePrivacySettings(userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete privacy settings", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Privacy settings deleted successfully",
	})
}

// HandleRequestDataExport handles POST /v1/privacy/export
func (h *Handler) HandleRequestDataExport(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	// Tenant ID is optional — individual users may not belong to a tenant
	tenantID, _ := h.getTenantID(r)

	var request struct {
		RequestType string `json:"request_type,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	if request.RequestType == "" {
		request.RequestType = "full"
	}

	exportRequest, err := h.service.RequestDataExport(r.Context(), userID, tenantID, request.RequestType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create export request", err.Error())
		return
	}

	respondWithJSON(w, http.StatusAccepted, exportRequest)
}

// HandleGetExportStatus handles GET /v1/privacy/exports/{id}
func (h *Handler) HandleGetExportStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	exportRequest, err := h.service.GetDataExportStatus(requestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get export status", err.Error())
		return
	}

	// Verify ownership
	if exportRequest.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", "You can only access your own exports")
		return
	}

	respondWithJSON(w, http.StatusOK, exportRequest)
}

// HandleAdminGetExportStatus handles GET /v1/admin/privacy/exports/{id} (admin endpoint - no ownership check)
func (h *Handler) HandleAdminGetExportStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	exportRequest, err := h.service.GetDataExportStatus(requestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get export status", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, exportRequest)
}

// HandleGetExportDownload handles GET /v1/privacy/exports/{id}/download
func (h *Handler) HandleGetExportDownload(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	exportRequest, err := h.service.GetDataExportStatus(requestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get export", err.Error())
		return
	}

	// Verify ownership
	if exportRequest.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", "You can only access your own exports")
		return
	}

	// Check if export is completed
	if exportRequest.Status != "completed" {
		respondWithError(w, http.StatusBadRequest, "Export not ready", fmt.Sprintf("Export status: %s", exportRequest.Status))
		return
	}

	// Check if download token matches
	token := r.URL.Query().Get("token")
	if token == "" || token != exportRequest.DownloadToken {
		respondWithError(w, http.StatusForbidden, "Invalid download token", "")
		return
	}

	// Check if export is expired
	if exportRequest.ExpiresAt != nil && exportRequest.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusGone, "Export expired", "This export has expired. Please request a new export.")
		return
	}

	// Determine how to serve the file based on storage type
	downloadURL := exportRequest.DownloadURL

	// If the download URL is a remote URL (S3/R2/CDN), redirect to it
	if isRemoteURL(downloadURL) {
		// Redirect to the secure storage URL (pre-signed or public)
		http.Redirect(w, r, downloadURL, http.StatusTemporaryRedirect)

		// Log the download
		logrus.WithFields(logrus.Fields{
			"export_id":  requestID,
			"user_id":    userID,
			"remote_url": true,
		}).Info("GDPR export download redirected to secure storage")
		return
	}

	// Local file: stream it directly
	filePath := downloadURL

	// SECURITY: Validate the file path is within the allowed export directory
	// This prevents path traversal attacks (e.g., ../../etc/passwd)
	exportBasePath := os.Getenv("PRIVACY_EXPORT_LOCAL_PATH")
	if exportBasePath == "" {
		exportBasePath = "./exports"
	}
	cleanPath, err := filepath.Abs(filePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file path", "Failed to resolve file path")
		return
	}
	cleanBasePath, err := filepath.Abs(exportBasePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Configuration error", "Failed to resolve export base path")
		return
	}
	cleanPath = filepath.Clean(cleanPath)
	cleanBasePath = filepath.Clean(cleanBasePath)
	if !strings.HasPrefix(cleanPath, cleanBasePath) {
		logrus.WithFields(logrus.Fields{
			"requested_path": filePath,
			"clean_path":      cleanPath,
			"base_path":       cleanBasePath,
		}).Warn("Path traversal attempt detected in GDPR export download")
		respondWithError(w, http.StatusForbidden, "Access denied", "Invalid file path")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		respondWithError(w, http.StatusNotFound, "Export file not found", "The export file may have been deleted or expired.")
		return
	}

	// Open and stream the file
	file, err := os.Open(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read export file", err.Error())
		return
	}
	defer file.Close()

	// Get file info for Content-Length
	fileInfo, err := file.Stat()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read export file", err.Error())
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"gdpr-export-%s.zip\"", requestID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	// Stream the file
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		logrus.WithError(err).Error("Failed to stream export file")
		return
	}

	// Log the download
	logrus.WithFields(logrus.Fields{
		"export_id": requestID,
		"user_id":   userID,
		"file_size": fileInfo.Size(),
		"local":     true,
	}).Info("GDPR export downloaded successfully")
}

// isRemoteURL checks if the URL is a remote (S3/R2/CDN) URL vs a local path
func isRemoteURL(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "s3://")
}

// HandleRequestDataDeletion handles POST /v1/privacy/deletion
func (h *Handler) HandleRequestDataDeletion(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	// Tenant ID is optional — individual users may not belong to a tenant
	tenantID, _ := h.getTenantID(r)

	var request struct {
		RequestType string `json:"request_type,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	if request.RequestType == "" {
		request.RequestType = "full"
	}

	deletionRequest, err := h.service.RequestDataDeletion(r.Context(), userID, tenantID, request.RequestType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create deletion request", err.Error())
		return
	}

	respondWithJSON(w, http.StatusAccepted, deletionRequest)
}

// HandleGetDeletionStatus handles GET /v1/privacy/deletions/{id}
func (h *Handler) HandleGetDeletionStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	deletionRequest, err := h.service.GetDataDeletionStatus(requestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get deletion status", err.Error())
		return
	}

	// Verify ownership
	if deletionRequest.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", "You can only access your own deletion requests")
		return
	}

	respondWithJSON(w, http.StatusOK, deletionRequest)
}

// HandleListExportRequests handles GET /v1/admin/privacy/exports (admin endpoint)
func (h *Handler) HandleListExportRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.GetAllExportRequests()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get export requests", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, requests)
}

// HandleListDeletionRequests handles GET /v1/admin/privacy/deletions (admin endpoint)
func (h *Handler) HandleListDeletionRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.GetAllDeletionRequests()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get deletion requests", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, requests)
}

// HandleAdminGetDeletionStatus handles GET /v1/admin/privacy/deletions/{id} (admin endpoint - no ownership check)
func (h *Handler) HandleAdminGetDeletionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	deletionRequest, err := h.service.GetDataDeletionStatus(requestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get deletion status", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, deletionRequest)
}

// HandleRecordConsent handles POST /v1/privacy/consent
func (h *Handler) HandleRecordConsent(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var request struct {
		ConsentType   string `json:"consent_type"`
		ConsentVersion string `json:"consent_version"`
		ConsentText   string `json:"consent_text"`
		Given         bool   `json:"given"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if request.ConsentType == "" {
		respondWithError(w, http.StatusBadRequest, "Consent type required", "")
		return
	}

	// Get IP and user agent (hashed for audit)
	ipHash := privacy.QuickHash(r.RemoteAddr)
	uaHash := privacy.QuickHash(r.UserAgent())

	record, err := h.service.RecordConsent(
		userID,
		request.ConsentType,
		request.ConsentVersion,
		request.ConsentText,
		ipHash,
		uaHash,
		request.Given,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to record consent", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, record)
}

// HandleGetConsentStatus handles GET /v1/privacy/consent
func (h *Handler) HandleGetConsentStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	consentType := r.URL.Query().Get("type")
	if consentType == "" {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"has_consent": false,
			"consent_type": nil,
			"message": "Consent type not specified",
		})
		return
	}

	hasConsent := h.service.HasActiveConsent(userID, consentType)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"has_consent":   hasConsent,
		"consent_type":  consentType,
	})
}

// HandleWithdrawConsent handles DELETE /v1/privacy/consent/{type}
func (h *Handler) HandleWithdrawConsent(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	vars := mux.Vars(r)
	consentType := vars["type"]

	var request struct {
		Reason string `json:"reason,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	if err := h.service.WithdrawConsent(userID, consentType, request.Reason); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to withdraw consent", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message":      "Consent withdrawn successfully",
		"consent_type": consentType,
	})
}

// HandleGetConsentHistory handles GET /v1/privacy/consent
func (h *Handler) HandleGetConsentHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	consentType := r.URL.Query().Get("type")

	records, err := h.service.GetConsentRecords(userID, consentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get consent history", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, records)
}

// HandleScanForPII handles POST /v1/privacy/pii-scan (admin endpoint)
func (h *Handler) HandleScanForPII(w http.ResponseWriter, r *http.Request) {
	// This is an admin endpoint - verify admin access
	if !h.isAdmin(r) {
		respondWithError(w, http.StatusForbidden, "Admin access required", "")
		return
	}

	var request struct {
		Data   interface{} `json:"data"`
		Redact bool        `json:"redact,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.ScanForPII(request.Data, request.Redact)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "PII scan failed", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// HandleGetGlobalPrivacySettings handles GET /v1/admin/privacy/settings (admin endpoint)
func (h *Handler) HandleGetGlobalPrivacySettings(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		respondWithError(w, http.StatusForbidden, "Admin access required", "")
		return
	}

	settings, err := h.service.GetGlobalPrivacySettings()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get global settings", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, settings)
}

// HandleUpdateGlobalPrivacySettings handles PUT /v1/admin/privacy/settings (admin endpoint)
func (h *Handler) HandleUpdateGlobalPrivacySettings(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		respondWithError(w, http.StatusForbidden, "Admin access required", "")
		return
	}

	userID, _ := h.getUserID(r)

	var updates GlobalPrivacySettingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	settings, err := h.service.GetGlobalPrivacySettings()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get global settings", err.Error())
		return
	}

	// Apply updates
	if updates.DefaultPrivacyLevel != nil {
		settings.DefaultPrivacyLevel = privacy.PrivacyLevel(*updates.DefaultPrivacyLevel)
	}
	if updates.DefaultIPMaskType != nil {
		settings.DefaultIPMaskType = privacy.PIIMaskType(*updates.DefaultIPMaskType)
	}
	if updates.DefaultUserAgentMaskType != nil {
		settings.DefaultUserAgentMaskType = privacy.PIIMaskType(*updates.DefaultUserAgentMaskType)
	}
	if updates.DefaultRetentionDays != nil {
		settings.DefaultRetentionDays = *updates.DefaultRetentionDays
	}
	if updates.GDPRModeEnabled != nil {
		settings.GDPRModeEnabled = *updates.GDPRModeEnabled
	}
	if updates.CCPAModeEnabled != nil {
		settings.CCPAModeEnabled = *updates.CCPAModeEnabled
	}
	if updates.AutoAnonymizeAfterDays != nil {
		settings.AutoAnonymizeAfterDays = *updates.AutoAnonymizeAfterDays
	}
	if updates.RequireConsent != nil {
		settings.RequireConsent = *updates.RequireConsent
	}
	if updates.PIIScanningEnabled != nil {
		settings.PIIScanningEnabled = *updates.PIIScanningEnabled
	}
	if updates.InputOutputRedaction != nil {
		settings.InputOutputRedaction = *updates.InputOutputRedaction
	}

	settings.UpdatedBy = &userID

	if err := h.service.UpdateGlobalPrivacySettings(settings); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update settings", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, settings)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"details": details,
		},
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper methods

// extractClaimsFromRequest parses the Bearer token from r and validates it via authSvc.
func extractClaimsFromRequest(r *http.Request, authSvc *auth.AuthService) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}
	return authSvc.ValidateToken(r.Context(), parts[1])
}

func (h *Handler) getUserID(r *http.Request) (uuid.UUID, error) {
	if h.authSvc != nil {
		claims, err := extractClaimsFromRequest(r, h.authSvc)
		if err != nil {
			return uuid.Nil, err
		}
		return claims.UserID, nil
	}
	return uuid.Nil, fmt.Errorf("auth service not available")
}

func (h *Handler) getTenantID(r *http.Request) (uuid.UUID, error) {
	if h.authSvc != nil {
		claims, err := extractClaimsFromRequest(r, h.authSvc)
		if err != nil {
			return uuid.Nil, err
		}
		if claims.TenantID == uuid.Nil {
			return uuid.Nil, fmt.Errorf("tenant ID not found in claims")
		}
		return claims.TenantID, nil
	}
	return uuid.Nil, fmt.Errorf("auth service not available")
}

func (h *Handler) isAdmin(r *http.Request) bool {
	if h.authSvc != nil {
		claims, err := extractClaimsFromRequest(r, h.authSvc)
		if err != nil {
			return false
		}
		// Check if user has admin role
		return claims.Role == "admin" || claims.Role == "platform_admin"
	}
	return false
}

// GlobalPrivacySettingsUpdate represents updateable fields for global privacy settings
type GlobalPrivacySettingsUpdate struct {
	DefaultPrivacyLevel       *string `json:"default_privacy_level,omitempty"`
	DefaultIPMaskType         *string `json:"default_ip_mask_type,omitempty"`
	DefaultUserAgentMaskType  *string `json:"default_user_agent_mask_type,omitempty"`
	DefaultRetentionDays      *int    `json:"default_retention_days,omitempty"`
	GDPRModeEnabled           *bool   `json:"gdpr_mode_enabled,omitempty"`
	CCPAModeEnabled           *bool   `json:"ccpa_mode_enabled,omitempty"`
	AutoAnonymizeAfterDays    *int    `json:"auto_anonymize_after_days,omitempty"`
	RequireConsent            *bool   `json:"require_consent,omitempty"`
	PIIScanningEnabled        *bool   `json:"pii_scanning_enabled,omitempty"`
	InputOutputRedaction      *bool   `json:"input_output_redaction,omitempty"`
}

// HandleListPrivacyAuditLogs handles GET /v1/admin/privacy/audit (admin endpoint)
func (h *Handler) HandleListPrivacyAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, total, err := h.service.GetAllPrivacyAuditLogs(limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get audit logs", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}
