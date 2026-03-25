package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminAuditHandler handles admin audit log endpoints
type AdminAuditHandler struct {
	db *sql.DB
}

// NewAdminAuditHandler creates a new admin audit handler
func NewAdminAuditHandler(db *sql.DB) *AdminAuditHandler {
	return &AdminAuditHandler{
		db: db,
	}
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID           uuid.UUID              `json:"id"`
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       string                `json:"action"`
	ResourceType string                `json:"resource_type"`
	ResourceID   *string               `json:"resource_id,omitempty"`
	IPAddress    *string               `json:"ip_address,omitempty"`
	UserAgent    *string               `json:"user_agent,omitempty"`
	RequestID    *uuid.UUID            `json:"request_id,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
}

// HandleListAuditLogs lists audit log entries with filtering
// GET /v1/admin/audit
func (h *AdminAuditHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse query parameters
	queryParams := r.URL.Query()

	// Pagination
	limitStr := queryParams.Get("limit")
	limit := 50 // Default
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
	userIDFilter := queryParams.Get("user_id")
	actionFilter := queryParams.Get("action")
	resourceTypeFilter := queryParams.Get("resource_type")
	dateFrom := queryParams.Get("date_from")
	dateTo := queryParams.Get("date_to")

	// Build query
	query := `
		SELECT id, user_id, action, resource_type, resource_id,
		       ip_address, user_agent, request_id, changes, created_at
		FROM admin_audit_log
		WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if userIDFilter != "" {
		query += " AND user_id = $" + strconv.Itoa(argNum)
		if uid, err := uuid.Parse(userIDFilter); err == nil {
			args = append(args, uid)
			argNum++
		}
	}

	if actionFilter != "" {
		query += " AND action = $" + strconv.Itoa(argNum)
		args = append(args, actionFilter)
		argNum++
	}

	if resourceTypeFilter != "" {
		query += " AND resource_type = $" + strconv.Itoa(argNum)
		args = append(args, resourceTypeFilter)
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

	// Count total
	countQuery := "SELECT COUNT(*) FROM admin_audit_log WHERE 1=1"
	countArgs := []interface{}{}
	countArgNum := 1

	if userIDFilter != "" {
		countQuery += " AND user_id = $" + strconv.Itoa(countArgNum)
		if uid, err := uuid.Parse(userIDFilter); err == nil {
			countArgs = append(countArgs, uid)
			countArgNum++
		}
	}
	if actionFilter != "" {
		countQuery += " AND action = $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, actionFilter)
		countArgNum++
	}
	if resourceTypeFilter != "" {
		countQuery += " AND resource_type = $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, resourceTypeFilter)
		countArgNum++
	}
	if dateFrom != "" {
		countQuery += " AND created_at >= $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, dateFrom)
		countArgNum++
	}
	if dateTo != "" {
		countQuery += " AND created_at <= $" + strconv.Itoa(countArgNum)
		countArgs = append(countArgs, dateTo)
	}

	var total int
	if err := h.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		logrus.WithError(err).Error("Failed to count audit logs")
		http.Error(w, "Failed to count audit logs", http.StatusInternalServerError)
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
		logrus.WithError(err).Error("Failed to list audit logs")
		http.Error(w, "Failed to list audit logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var entry AuditLogEntry
		var userID sql.NullString
		var resourceID sql.NullString
		var ipAddress sql.NullString
		var userAgent sql.NullString
		var requestID sql.NullString
		var changes []byte

		if err := rows.Scan(
			&entry.ID,
			&userID,
			&entry.Action,
			&entry.ResourceType,
			&resourceID,
			&ipAddress,
			&userAgent,
			&requestID,
			&changes,
			&entry.CreatedAt,
		); err != nil {
			logrus.WithError(err).Warn("Failed to scan audit log entry")
			continue
		}

		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			entry.UserID = &uid
		}
		if resourceID.Valid {
			entry.ResourceID = &resourceID.String
		}
		if ipAddress.Valid {
			entry.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			entry.UserAgent = &userAgent.String
		}
		if requestID.Valid {
			rid, _ := uuid.Parse(requestID.String)
			entry.RequestID = &rid
		}
		if changes != nil {
			if err := json.Unmarshal(changes, &entry.Changes); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal audit log changes")
			}
		}

		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []AuditLogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"audit_logs": entries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// HandleGetAuditLog gets a specific audit log entry
// GET /v1/admin/audit/:id
func (h *AdminAuditHandler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	// Check permission
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !auth.IsAdminRole(claims.Role) {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var entry AuditLogEntry
	var userID sql.NullString
	var resourceID sql.NullString
	var ipAddress sql.NullString
	var userAgent sql.NullString
	var requestID sql.NullString
	var changes []byte

	err = h.db.QueryRowContext(ctx, `
		SELECT id, user_id, action, resource_type, resource_id,
		       ip_address, user_agent, request_id, changes, created_at
		FROM admin_audit_log
		WHERE id = $1`, id).Scan(
		&entry.ID,
		&userID,
		&entry.Action,
		&entry.ResourceType,
		&resourceID,
		&ipAddress,
		&userAgent,
		&requestID,
		&changes,
		&entry.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Audit log entry not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get audit log entry")
		http.Error(w, "Failed to get audit log entry", http.StatusInternalServerError)
		return
	}

	if userID.Valid {
		uid, _ := uuid.Parse(userID.String)
		entry.UserID = &uid
	}
	if resourceID.Valid {
		entry.ResourceID = &resourceID.String
	}
	if ipAddress.Valid {
		entry.IPAddress = &ipAddress.String
	}
	if userAgent.Valid {
		entry.UserAgent = &userAgent.String
	}
	if requestID.Valid {
		rid, _ := uuid.Parse(requestID.String)
		entry.RequestID = &rid
	}
	if changes != nil {
		if err := json.Unmarshal(changes, &entry.Changes); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal audit log changes")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// CreateAuditLog creates a new audit log entry
func (h *AdminAuditHandler) CreateAuditLog(ctx context.Context, userID uuid.UUID, action, resourceType string, resourceID *string, ipAddress *string, userAgent *string, requestID *uuid.UUID, changes map[string]interface{}) error {
	if h.db == nil {
		return nil
	}

	query := `
		INSERT INTO admin_audit_log (id, user_id, action, resource_type, resource_id, ip_address, user_agent, request_id, changes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`

	var changesJSON []byte
	var err error
	if changes != nil {
		changesJSON, err = json.Marshal(changes)
		if err != nil {
			logrus.WithError(err).Warn("Failed to marshal audit log changes")
		}
	}

	_, err = h.db.ExecContext(ctx, query,
		uuid.New(),
		userID,
		action,
		resourceType,
		resourceID,
		ipAddress,
		userAgent,
		requestID,
		changesJSON,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create audit log entry")
		return err
	}

	return nil
}

// BuildFilterURL builds a URL with the current filters applied
func BuildFilterURL(baseURL string, filters map[string]string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	for k, v := range filters {
		if v != "" {
			q.Set(k, v)
		} else {
			q.Del(k)
		}
	}
	u.RawQuery = q.Encode()

	return u.String()
}
