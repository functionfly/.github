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

// SecurityEventHandler handles security event endpoints
type SecurityEventHandler struct {
	db *sql.DB
}

// NewSecurityEventHandler creates a new security event handler
func NewSecurityEventHandler(db *sql.DB) *SecurityEventHandler {
	return &SecurityEventHandler{
		db: db,
	}
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID           uuid.UUID              `json:"id"`
	EventType    string                `json:"event_type"`
	IPAddress    *string               `json:"ip_address,omitempty"`
	UserAgent    *string               `json:"user_agent,omitempty"`
	UserID       *uuid.UUID            `json:"user_id,omitempty"`
	Email        *string               `json:"email,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IsReviewed   bool                  `json:"is_reviewed"`
	ReviewedBy   *uuid.UUID            `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time            `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
}

// HandleListSecurityEvents lists security events
// GET /v1/admin/security-events
func (h *SecurityEventHandler) HandleListSecurityEvents(w http.ResponseWriter, r *http.Request) {
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
	eventType := queryParams.Get("event_type")
	isReviewedStr := queryParams.Get("is_reviewed")
	dateFrom := queryParams.Get("date_from")
	dateTo := queryParams.Get("date_to")

	// Combine events from multiple sources
	// 1. Failed login attempts (auth_events table)
	// 2. Rate limited requests (from Redis - not directly queryable, so we skip for now)
	// 3. Blocked IPs (auth_events table with event_type = 'login_blocked')

	// Build query for failed login attempts
	query := `
		SELECT id, event_type, ip_address, user_agent, user_id, email, details,
		       FALSE as is_reviewed, NULL::uuid as reviewed_by, NULL::timestamptz as reviewed_at,
		       created_at
		FROM auth_events
		WHERE event_type IN ('login_failed', 'login_blocked', 'mfa_failed', 'password_reset_failed')`
	args := []interface{}{}
	argNum := 1

	if eventType != "" {
		query += " AND event_type = $" + strconv.Itoa(argNum)
		args = append(args, eventType)
		argNum++
	}

	if dateFrom != "" {
		query += " AND created_at >= $" + strconv.Itoa(argNum)
		args = append(args, dateFrom)
		argNum++
	}

	if dateTo != "" {
		query += " AND created_at <= $" + strconv.Itoa(argNum)
		args = append(args, dateTo)
		argNum++
	}

	// Filter by reviewed status
	if isReviewedStr == "true" {
		query += " AND FALSE = TRUE" // No auth_events have is_reviewed column
	} else if isReviewedStr == "false" {
		// Include all auth_events since they don't have review status
	}

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM auth_events
		WHERE event_type IN ('login_failed', 'login_blocked', 'mfa_failed', 'password_reset_failed')`
	countArgs := []interface{}{}

	if eventType != "" {
		countQuery += " AND event_type = $1"
		countArgs = append(countArgs, eventType)
	}
	if dateFrom != "" {
		if eventType != "" {
			countQuery += " AND created_at >= $2"
		} else {
			countQuery += " AND created_at >= $1"
		}
		countArgs = append(countArgs, dateFrom)
	}
	if dateTo != "" {
		paramNum := len(countArgs) + 1
		countQuery += " AND created_at <= $" + strconv.Itoa(paramNum)
		countArgs = append(countArgs, dateTo)
	}

	var total int
	if err := h.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		logrus.WithError(err).Error("Failed to count security events")
		apierror.WriteError(w, apierror.NewInternal("Failed to count security events"))
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
		logrus.WithError(err).Error("Failed to list security events")
		apierror.WriteError(w, apierror.NewInternal("Failed to list security events"))
		return
	}
	defer rows.Close()

	var events []SecurityEvent
	for rows.Next() {
		var event SecurityEvent
		var ipAddress, userAgent, email sql.NullString
		var userID sql.NullString
		var details []byte
		var isReviewed sql.NullBool
		var reviewedBy sql.NullString
		var reviewedAt sql.NullString

		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&ipAddress,
			&userAgent,
			&userID,
			&email,
			&details,
			&isReviewed,
			&reviewedBy,
			&reviewedAt,
			&event.CreatedAt,
		); err != nil {
			logrus.WithError(err).Warn("Failed to scan security event")
			continue
		}

		if ipAddress.Valid {
			event.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = &userAgent.String
		}
		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			event.UserID = &uid
		}
		if email.Valid {
			event.Email = &email.String
		}
		if details != nil {
			if err := json.Unmarshal(details, &event.Details); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal event details")
			}
		}
		if isReviewed.Valid {
			event.IsReviewed = isReviewed.Bool
		}
		if reviewedBy.Valid {
			uid, _ := uuid.Parse(reviewedBy.String)
			event.ReviewedBy = &uid
		}
		if reviewedAt.Valid {
			t, _ := time.Parse(time.RFC3339, reviewedAt.String)
			event.ReviewedAt = &t
		}

		events = append(events, event)
	}

	if events == nil {
		events = []SecurityEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"security_events": events,
		"total":           total,
		"limit":           limit,
		"offset":          offset,
	})
}

// HandleReviewSecurityEvent marks a security event as reviewed
// POST /v1/admin/security-events/:id/review
func (h *SecurityEventHandler) HandleReviewSecurityEvent(w http.ResponseWriter, r *http.Request) {
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid ID"))
		return
	}

	ctx := r.Context()

	// Update the security event as reviewed
	// Since auth_events doesn't have is_reviewed column, we'll add a note/comment instead
	// For now, just mark it as reviewed by updating details JSON
	query := `
		UPDATE auth_events
		SET details = COALESCE(details, '{}') || jsonb_build_object('reviewed', true, 'reviewed_by', $1::text, 'reviewed_at', NOW()::text)
		WHERE id = $2
		RETURNING id`

	var eventID uuid.UUID
	err = h.db.QueryRowContext(ctx, query, claims.UserID, id).Scan(&eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			apierror.WriteError(w, apierror.NewNotFound("Security event not found"))
			return
		}
		logrus.WithError(err).Error("Failed to review security event")
		apierror.WriteError(w, apierror.NewInternal("Failed to review security event"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         eventID,
		"is_reviewed": true,
		"reviewed_by": claims.UserID,
		"reviewed_at": time.Now(),
	})
}

// HandleCreateSecurityEvents creates multiple security events from client-side tracking
// POST /v1/admin/security/events
func (h *SecurityEventHandler) HandleCreateSecurityEvents(w http.ResponseWriter, r *http.Request) {
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

	var requestBody struct {
		Events []struct {
			EventType         string                 `json:"event_type"`
			Timestamp         string                 `json:"timestamp"`
			IPAddress         string                 `json:"ip_address,omitempty"`
			UserAgent         string                 `json:"user_agent,omitempty"`
			DeviceFingerprint string                 `json:"device_fingerprint,omitempty"`
			SessionID         string                 `json:"session_id,omitempty"`
			Metadata          map[string]interface{} `json:"metadata,omitempty"`
		} `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if len(requestBody.Events) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"created": 0,
		})
		return
	}

	ctx := r.Context()
	createdCount := 0

	for _, event := range requestBody.Events {
		// Parse timestamp
		eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			eventTime = time.Now()
		}

		// Build metadata JSON including device fingerprint if provided
		metadata := event.Metadata
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		if event.DeviceFingerprint != "" {
			metadata["device_fingerprint"] = event.DeviceFingerprint
		}
		metadataJSON, _ := json.Marshal(metadata)

		// session_id column is UUID — skip non-UUID values (e.g. "jwt-..." synthetic IDs)
		var sessionID interface{}
		if event.SessionID != "" {
			if _, err := uuid.Parse(event.SessionID); err == nil {
				sessionID = event.SessionID
			}
		}

		// Insert into auth_events table
		// Note: success column is NOT NULL, default to true for security events
		query := `
			INSERT INTO auth_events (id, event_type, success, ip_address, user_agent, session_id, metadata, created_at)
			VALUES ($1, $2, TRUE, $3, $4, $5, $6, $7)`

		_, err = h.db.ExecContext(ctx, query,
			uuid.New(),
			event.EventType,
			nullString(event.IPAddress),
			nullString(event.UserAgent),
			sessionID,
			metadataJSON,
			eventTime,
		)
		if err != nil {
			logrus.WithError(err).WithField("event_type", event.EventType).Warn("Failed to create security event")
			continue
		}
		createdCount++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": createdCount,
	})
}

// nullString returns a sql.NullString for optional string fields
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// HandleGetSecurityEventStats returns security event statistics
// GET /v1/admin/security-events/stats
func (h *SecurityEventHandler) HandleGetSecurityEventStats(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()

	// Get counts by event type for the last 24 hours
	query := `
		SELECT
			event_type,
			COUNT(*) as count
		FROM auth_events
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		AND event_type IN ('login_failed', 'login_blocked', 'mfa_failed', 'password_reset_failed')
		GROUP BY event_type
		ORDER BY count DESC`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to get security event stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get security event stats"))
		return
	}
	defer rows.Close()

	stats := make(map[string]int)
	var total int
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			logrus.WithError(err).Warn("Failed to scan security event stat")
			continue
		}
		stats[eventType] = count
		total += count
	}

	// Get unique IPs with failed logins
	ipQuery := `
		SELECT COUNT(DISTINCT ip_address)
		FROM auth_events
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		AND event_type IN ('login_failed', 'login_blocked')`
	var uniqueIPs int
	if err := h.db.QueryRowContext(ctx, ipQuery).Scan(&uniqueIPs); err != nil {
		logrus.WithError(err).Warn("Failed to count unique IPs")
	}

	// Get blocked IPs count
	blockedQuery := `
		SELECT COUNT(DISTINCT ip_address)
		FROM auth_events
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		AND event_type = 'login_blocked'`
	var blockedIPs int
	if err := h.db.QueryRowContext(ctx, blockedQuery).Scan(&blockedIPs); err != nil {
		logrus.WithError(err).Warn("Failed to count blocked IPs")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_events":      total,
		"events_by_type":     stats,
		"unique_ips":        uniqueIPs,
		"blocked_ips":        blockedIPs,
		"period":             "24 hours",
		"generated_at":       time.Now(),
	})
}
