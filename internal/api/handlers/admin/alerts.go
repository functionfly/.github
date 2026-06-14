package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AlertHandler handles security alert rule endpoints
type AlertHandler struct {
	db *sql.DB
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(db *sql.DB) *AlertHandler {
	return &AlertHandler{
		db: db,
	}
}

// SecurityAlertRule represents a security alert rule
type SecurityAlertRule struct {
	ID                   uuid.UUID              `json:"id"`
	Name                 string                `json:"name"`
	AlertType            string                `json:"alert_type"`
	Threshold            int                   `json:"threshold"`
	WindowSeconds        int                   `json:"window_seconds"`
	Severity             string                `json:"severity"`
	IsEnabled            bool                  `json:"is_enabled"`
	NotificationChannels []string              `json:"notification_channels"`
	CreatedBy            *uuid.UUID            `json:"created_by,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

// Valid alert types
var ValidAlertTypes = map[string]bool{
	"failed_login_threshold": true,
	"rate_limit_exceeded":    true,
	"ip_blocked":             true,
	"suspicious_activity":     true,
	"session_anomaly":        true,
}

// Valid severity levels
var ValidSeverityLevels = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

// HandleListSecurityAlerts lists all configured security alert rules
// GET /v1/admin/security-alerts
func (h *AlertHandler) HandleListSecurityAlerts(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden: insufficient permissions"))
		return
	}

	// Parse query parameters
	queryParams := r.URL.Query()

	// Pagination
	limitStr := queryParams.Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offsetStr := queryParams.Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Filters
	alertType := queryParams.Get("alert_type")
	severity := queryParams.Get("severity")
	isEnabled := queryParams.Get("is_enabled")

	// Build query
	query := `
		SELECT id, name, alert_type, threshold, window_seconds, severity,
		       is_enabled, notification_channels, created_by, created_at, updated_at
		FROM security_alert_rules
		WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if alertType != "" {
		query += " AND alert_type = $" + strconv.Itoa(argNum)
		args = append(args, alertType)
		argNum++
	}

	if severity != "" {
		query += " AND severity = $" + strconv.Itoa(argNum)
		args = append(args, severity)
		argNum++
	}

	if isEnabled == "true" {
		query += " AND is_enabled = TRUE"
	} else if isEnabled == "false" {
		query += " AND is_enabled = FALSE"
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM security_alert_rules WHERE 1=1`
	countArgs := []interface{}{}
	countArgNum := 1

	if alertType != "" {
		countQuery += " AND alert_type = $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, alertType)
		countArgNum++
	}

	if severity != "" {
		countQuery += " AND severity = $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, severity)
		countArgNum++
	}

	if isEnabled == "true" {
		countQuery += " AND is_enabled = TRUE"
	} else if isEnabled == "false" {
		countQuery += " AND is_enabled = FALSE"
	}

	var total int
	if err := h.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		logrus.WithError(err).Error("Failed to count security alerts")
		apierror.WriteError(w, apierror.NewInternal("Failed to count security alerts"))
		return
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += " LIMIT $" + strconv.Itoa(argNum)
	args = append(args, limit)
	argNum++
	query += " OFFSET $" + strconv.Itoa(argNum)
	args = append(args, offset)

	ctx := r.Context()
	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to list security alerts")
		apierror.WriteError(w, apierror.NewInternal("Failed to list security alerts"))
		return
	}
	defer rows.Close()

	var alerts []SecurityAlertRule
	for rows.Next() {
		var alert SecurityAlertRule
		var notificationChannels []byte
		var createdBy sql.NullString

		if err := rows.Scan(
			&alert.ID,
			&alert.Name,
			&alert.AlertType,
			&alert.Threshold,
			&alert.WindowSeconds,
			&alert.Severity,
			&alert.IsEnabled,
			&notificationChannels,
			&createdBy,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		); err != nil {
			logrus.WithError(err).Warn("Failed to scan security alert")
			continue
		}

		if notificationChannels != nil {
			if err := json.Unmarshal(notificationChannels, &alert.NotificationChannels); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal notification channels")
				alert.NotificationChannels = []string{}
			}
		}

		if createdBy.Valid {
			uid, _ := uuid.Parse(createdBy.String)
			alert.CreatedBy = &uid
		}

		alerts = append(alerts, alert)
	}

	if alerts == nil {
		alerts = []SecurityAlertRule{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"security_alerts": alerts,
		"total":           total,
		"limit":           limit,
		"offset":          offset,
	})
}

// HandleCreateSecurityAlert creates a new security alert rule
// POST /v1/admin/security-alerts
func (h *AlertHandler) HandleCreateSecurityAlert(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden: insufficient permissions"))
		return
	}

	var req struct {
		Name                 string   `json:"name"`
		AlertType            string   `json:"alert_type"`
		Threshold            int      `json:"threshold"`
		WindowSeconds        int      `json:"window_seconds"`
		Severity             string   `json:"severity"`
		IsEnabled            *bool    `json:"is_enabled"`
		NotificationChannels []string `json:"notification_channels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Validate required fields
	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Alert name is required"))
		return
	}

	if req.AlertType == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Alert type is required"))
		return
	}

	if !ValidAlertTypes[req.AlertType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert type. Valid types: failed_login_threshold, rate_limit_exceeded, ip_blocked, suspicious_activity, session_anomaly"))
		return
	}

	if req.Threshold <= 0 {
		apierror.WriteError(w, apierror.NewBadRequest("Threshold must be a positive integer"))
		return
	}

	// Default window to 300 seconds (5 minutes) if not specified
	if req.WindowSeconds <= 0 {
		req.WindowSeconds = 300
	}

	// Default severity to medium if not specified
	if req.Severity == "" {
		req.Severity = "medium"
	} else if !ValidSeverityLevels[req.Severity] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid severity level. Valid levels: low, medium, high, critical"))
		return
	}

	// Default is_enabled to true if not specified
	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	// Default notification channels to empty array
	if req.NotificationChannels == nil {
		req.NotificationChannels = []string{}
	}

	notificationChannelsJSON, err := json.Marshal(req.NotificationChannels)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal notification channels")
		apierror.WriteError(w, apierror.NewInternal("Failed to create security alert"))
		return
	}

	ctx := r.Context()
	query := `
		INSERT INTO security_alert_rules (name, alert_type, threshold, window_seconds, severity, is_enabled, notification_channels, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, name, alert_type, threshold, window_seconds, severity, is_enabled, notification_channels, created_by, created_at, updated_at`

	var alert SecurityAlertRule
	var notificationChannels []byte
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, query,
		req.Name,
		req.AlertType,
		req.Threshold,
		req.WindowSeconds,
		req.Severity,
		isEnabled,
		notificationChannelsJSON,
		claims.UserID,
	).Scan(
		&alert.ID,
		&alert.Name,
		&alert.AlertType,
		&alert.Threshold,
		&alert.WindowSeconds,
		&alert.Severity,
		&alert.IsEnabled,
		&notificationChannels,
		&createdBy,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create security alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to create security alert"))
		return
	}

	if notificationChannels != nil {
		if err := json.Unmarshal(notificationChannels, &alert.NotificationChannels); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal notification channels")
			alert.NotificationChannels = []string{}
		}
	}

	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		alert.CreatedBy = &uid
	}

	logrus.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"alert_type": alert.AlertType,
		"created_by": claims.UserID,
	}).Info("Security alert rule created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// HandleUpdateSecurityAlert updates an existing security alert rule
// PUT /v1/admin/security-alerts/:id
func (h *AlertHandler) HandleUpdateSecurityAlert(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	_, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	var req struct {
		Name                 *string  `json:"name,omitempty"`
		AlertType            *string  `json:"alert_type,omitempty"`
		Threshold            *int     `json:"threshold,omitempty"`
		WindowSeconds        *int     `json:"window_seconds,omitempty"`
		Severity             *string  `json:"severity,omitempty"`
		IsEnabled            *bool    `json:"is_enabled,omitempty"`
		NotificationChannels []string `json:"notification_channels,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Validate alert_type if provided
	if req.AlertType != nil && !ValidAlertTypes[*req.AlertType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert type. Valid types: failed_login_threshold, rate_limit_exceeded, ip_blocked, suspicious_activity, session_anomaly"))
		return
	}

	// Validate severity if provided
	if req.Severity != nil && !ValidSeverityLevels[*req.Severity] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid severity level. Valid levels: low, medium, high, critical"))
		return
	}

	// Validate threshold if provided
	if req.Threshold != nil && *req.Threshold <= 0 {
		apierror.WriteError(w, apierror.NewBadRequest("Threshold must be a positive integer"))
		return
	}

	ctx := r.Context()

	// Check if alert exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM security_alert_rules WHERE id = $1)`
	if err := h.db.QueryRowContext(ctx, checkQuery, idStr).Scan(&exists); err != nil {
		logrus.WithError(err).Error("Failed to check if security alert exists")
		apierror.WriteError(w, apierror.NewInternal("Failed to update security alert"))
		return
	}

	if !exists {
		apierror.WriteError(w, apierror.NewNotFound("Security alert not found"))
		return
	}

	// Build dynamic update query
	updateQuery := `UPDATE security_alert_rules SET updated_at = NOW()`
	args := []interface{}{}
	argNum := 1

	if req.Name != nil {
		updateQuery += ", name = $" + strconv.Itoa(argNum)
		args = append(args, *req.Name)
		argNum++
	}

	if req.AlertType != nil {
		updateQuery += ", alert_type = $" + strconv.Itoa(argNum)
		args = append(args, *req.AlertType)
		argNum++
	}

	if req.Threshold != nil {
		updateQuery += ", threshold = $" + strconv.Itoa(argNum)
		args = append(args, *req.Threshold)
		argNum++
	}

	if req.WindowSeconds != nil {
		updateQuery += ", window_seconds = $" + strconv.Itoa(argNum)
		args = append(args, *req.WindowSeconds)
		argNum++
	}

	if req.Severity != nil {
		updateQuery += ", severity = $" + strconv.Itoa(argNum)
		args = append(args, *req.Severity)
		argNum++
	}

	if req.IsEnabled != nil {
		updateQuery += ", is_enabled = $" + strconv.Itoa(argNum)
		args = append(args, *req.IsEnabled)
		argNum++
	}

	if req.NotificationChannels != nil {
		notificationChannelsJSON, err := json.Marshal(req.NotificationChannels)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshal notification channels")
			apierror.WriteError(w, apierror.NewInternal("Failed to update security alert"))
			return
		}
		updateQuery += ", notification_channels = $" + strconv.Itoa(argNum)
		args = append(args, notificationChannelsJSON)
		argNum++
	}

	updateQuery += " WHERE id = $" + strconv.Itoa(argNum)
	args = append(args, idStr)

	updateQuery += " RETURNING id, name, alert_type, threshold, window_seconds, severity, is_enabled, notification_channels, created_by, created_at, updated_at"

	var alert SecurityAlertRule
	var notificationChannels []byte
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, updateQuery, args...).Scan(
		&alert.ID,
		&alert.Name,
		&alert.AlertType,
		&alert.Threshold,
		&alert.WindowSeconds,
		&alert.Severity,
		&alert.IsEnabled,
		&notificationChannels,
		&createdBy,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to update security alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to update security alert"))
		return
	}

	if notificationChannels != nil {
		if err := json.Unmarshal(notificationChannels, &alert.NotificationChannels); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal notification channels")
			alert.NotificationChannels = []string{}
		}
	}

	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		alert.CreatedBy = &uid
	}

	logrus.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"updated_by": claims.UserID,
	}).Info("Security alert rule updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// HandleDeleteSecurityAlert deletes a security alert rule
// DELETE /v1/admin/security-alerts/:id
func (h *AlertHandler) HandleDeleteSecurityAlert(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	alertID, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	ctx := r.Context()

	// Check if alert exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM security_alert_rules WHERE id = $1)`
	if err := h.db.QueryRowContext(ctx, checkQuery, idStr).Scan(&exists); err != nil {
		logrus.WithError(err).Error("Failed to check if security alert exists")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete security alert"))
		return
	}

	if !exists {
		apierror.WriteError(w, apierror.NewNotFound("Security alert not found"))
		return
	}

	// Delete the alert
	deleteQuery := `DELETE FROM security_alert_rules WHERE id = $1`
	_, err = h.db.ExecContext(ctx, deleteQuery, idStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete security alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete security alert"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"alert_id":   alertID,
		"deleted_by": claims.UserID,
	}).Info("Security alert rule deleted")

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetSecurityAlert gets a specific security alert rule
// GET /v1/admin/security-alerts/:id
func (h *AlertHandler) HandleGetSecurityAlert(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	_, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	ctx := r.Context()

	query := `
		SELECT id, name, alert_type, threshold, window_seconds, severity,
		       is_enabled, notification_channels, created_by, created_at, updated_at
		FROM security_alert_rules
		WHERE id = $1`

	var alert SecurityAlertRule
	var notificationChannels []byte
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, query, idStr).Scan(
		&alert.ID,
		&alert.Name,
		&alert.AlertType,
		&alert.Threshold,
		&alert.WindowSeconds,
		&alert.Severity,
		&alert.IsEnabled,
		&notificationChannels,
		&createdBy,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			apierror.WriteError(w, apierror.NewNotFound("Security alert not found"))
			return
		}
		logrus.WithError(err).Error("Failed to get security alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to get security alert"))
		return
	}

	if notificationChannels != nil {
		if err := json.Unmarshal(notificationChannels, &alert.NotificationChannels); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal notification channels")
			alert.NotificationChannels = []string{}
		}
	}

	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		alert.CreatedBy = &uid
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// RegisterRoutes registers security alert routes
func (h *AlertHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/security-alerts", h.HandleListSecurityAlerts).Methods("GET", "OPTIONS")
	router.HandleFunc("/security-alerts", h.HandleCreateSecurityAlert).Methods("POST", "OPTIONS")
	router.HandleFunc("/security-alerts/{id}", h.HandleGetSecurityAlert).Methods("GET", "OPTIONS")
	router.HandleFunc("/security-alerts/{id}", h.HandleUpdateSecurityAlert).Methods("PUT", "OPTIONS")
	router.HandleFunc("/security-alerts/{id}", h.HandleDeleteSecurityAlert).Methods("DELETE", "OPTIONS")
}
