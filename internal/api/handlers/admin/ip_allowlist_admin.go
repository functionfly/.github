package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminIPAllowlistHandler handles admin-level IP allowlist management
type AdminIPAllowlistHandler struct {
	db                   *sql.DB
	ipAllowlistMiddleware *middleware.IPAllowlistMiddleware
}

// NewAdminIPAllowlistHandler creates a new admin IP allowlist handler
func NewAdminIPAllowlistHandler(db *sql.DB, ipAllowlistMiddleware *middleware.IPAllowlistMiddleware) *AdminIPAllowlistHandler {
	return &AdminIPAllowlistHandler{
		db:                   db,
		ipAllowlistMiddleware: ipAllowlistMiddleware,
	}
}

// IPAccessEntry represents an entry in the admin IP allowlist
type IPAccessEntry struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	CIDR        string     `json:"cidr"`
	Description string     `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	IsWhitelist bool       `json:"is_whitelist"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// HandleListIPAllowlist lists all IP allowlist entries
// GET /v1/admin/ip-allowlist
func (h *AdminIPAllowlistHandler) HandleListIPAllowlist(w http.ResponseWriter, r *http.Request) {
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
	query := `
		SELECT id, name, cidr, COALESCE(description, ''), is_active, is_whitelist,
		       created_by, created_at, updated_at
		FROM ip_allowlist
		ORDER BY created_at DESC`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to list IP allowlist entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to list IP allowlist entries"))
		return
	}
	defer rows.Close()

	var entries []IPAccessEntry
	for rows.Next() {
		var entry IPAccessEntry
		var description string
		var createdBy sql.NullString

		if err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.CIDR,
			&description,
			&entry.IsActive,
			&entry.IsWhitelist,
			&createdBy,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			logrus.WithError(err).Warn("Failed to scan IP allowlist entry")
			continue
		}

		entry.Description = description
		if createdBy.Valid {
			id, _ := uuid.Parse(createdBy.String)
			entry.CreatedBy = &id
		}

		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []IPAccessEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ip_allowlist": entries,
		"total":        len(entries),
	})
}

// HandleGetIPAllowlist gets a specific IP allowlist entry
// GET /v1/admin/ip-allowlist/:id
func (h *AdminIPAllowlistHandler) HandleGetIPAllowlist(w http.ResponseWriter, r *http.Request) {
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
	var entry IPAccessEntry
	var description string
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, `
		SELECT id, name, cidr, COALESCE(description, ''), is_active, is_whitelist,
		       created_by, created_at, updated_at
		FROM ip_allowlist WHERE id = $1`, id).Scan(
		&entry.ID, &entry.Name, &entry.CIDR, &description,
		&entry.IsActive, &entry.IsWhitelist, &createdBy,
		&entry.CreatedAt, &entry.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
			return
		}
		logrus.WithError(err).Error("Failed to get IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist entry"))
		return
	}

	entry.Description = description
	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		entry.CreatedBy = &uid
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleCreateIPAllowlist creates a new IP allowlist entry
// POST /v1/admin/ip-allowlist
func (h *AdminIPAllowlistHandler) HandleCreateIPAllowlist(w http.ResponseWriter, r *http.Request) {
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
		Name        string `json:"name"`
		CIDR        string `json:"cidr"`
		Description string `json:"description"`
		IsWhitelist *bool  `json:"is_whitelist"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Name is required"))
		return
	}

	if req.CIDR == "" {
		apierror.WriteError(w, apierror.NewBadRequest("CIDR is required"))
		return
	}

	// Validate CIDR format
	if _, _, err := net.ParseCIDR(req.CIDR); err != nil {
		// Try parsing as a single IP
		ip := net.ParseIP(req.CIDR)
		if ip == nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid CIDR format"))
			return
		}
		// Convert single IP to CIDR notation
		if ip.To4() != nil {
			req.CIDR = ip.String() + "/32"
		} else {
			req.CIDR = ip.String() + "/128"
		}
	}

	isWhitelist := true
	if req.IsWhitelist != nil {
		isWhitelist = *req.IsWhitelist
	}

	ctx := r.Context()
	query := `
		INSERT INTO ip_allowlist (id, name, cidr, description, is_active, is_whitelist, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, NOW(), NOW())
		RETURNING id, name, cidr, COALESCE(description, ''), is_active, is_whitelist, created_by, created_at, updated_at`

	entry := IPAccessEntry{}
	var description string
	var createdBy sql.NullString

	err := h.db.QueryRowContext(ctx, query,
		uuid.New(),
		req.Name,
		req.CIDR,
		req.Description,
		isWhitelist,
		claims.UserID,
	).Scan(
		&entry.ID,
		&entry.Name,
		&entry.CIDR,
		&description,
		&entry.IsActive,
		&entry.IsWhitelist,
		&createdBy,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to create IP allowlist entry"))
		return
	}

	entry.Description = description
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		entry.CreatedBy = &id
	}

	// Invalidate cache
	if h.ipAllowlistMiddleware != nil {
		h.ipAllowlistMiddleware.HandleInvalidateCache(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// HandleUpdateIPAllowlist updates an IP allowlist entry
// PUT /v1/admin/ip-allowlist/:id
func (h *AdminIPAllowlistHandler) HandleUpdateIPAllowlist(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Name        *string `json:"name"`
		CIDR        *string `json:"cidr"`
		Description *string `json:"description"`
		IsWhitelist *bool   `json:"is_whitelist"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	ctx := r.Context()

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if req.Name != nil {
		updates = append(updates, "name = $"+string(rune('0'+argNum)))
		args = append(args, *req.Name)
		argNum++
	}

	if req.CIDR != nil {
		// Validate CIDR
		_, _, err := net.ParseCIDR(*req.CIDR)
		if err != nil {
			ip := net.ParseIP(*req.CIDR)
			if ip == nil {
				apierror.WriteError(w, apierror.NewBadRequest("Invalid CIDR format"))
				return
			}
			if ip.To4() != nil {
				*req.CIDR = ip.String() + "/32"
			} else {
				*req.CIDR = ip.String() + "/128"
			}
		}
		updates = append(updates, "cidr = $"+string(rune('0'+argNum)))
		args = append(args, *req.CIDR)
		argNum++
	}

	if req.Description != nil {
		updates = append(updates, "description = $"+string(rune('0'+argNum)))
		args = append(args, *req.Description)
		argNum++
	}

	if req.IsWhitelist != nil {
		updates = append(updates, "is_whitelist = $"+string(rune('0'+argNum)))
		args = append(args, *req.IsWhitelist)
		argNum++
	}

	if len(updates) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("No fields to update"))
		return
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := "UPDATE ip_allowlist SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += " WHERE id = $" + string(rune('0'+argNum))

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to update IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to update IP allowlist entry"))
		return
	}

	rowsAff, _ := result.RowsAffected()
	if rowsAff == 0 {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
		return
	}

	// Fetch updated entry
	var entry IPAccessEntry
	var description string
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, `
		SELECT id, name, cidr, COALESCE(description, ''), is_active, is_whitelist,
		       created_by, created_at, updated_at
		FROM ip_allowlist WHERE id = $1`, id).Scan(
		&entry.ID, &entry.Name, &entry.CIDR, &description,
		&entry.IsActive, &entry.IsWhitelist, &createdBy,
		&entry.CreatedAt, &entry.UpdatedAt,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch updated IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to fetch updated entry"))
		return
	}

	entry.Description = description
	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		entry.CreatedBy = &uid
	}

	// Invalidate cache
	if h.ipAllowlistMiddleware != nil {
		h.ipAllowlistMiddleware.HandleInvalidateCache(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleDeleteIPAllowlist deletes an IP allowlist entry
// DELETE /v1/admin/ip-allowlist/:id
func (h *AdminIPAllowlistHandler) HandleDeleteIPAllowlist(w http.ResponseWriter, r *http.Request) {
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
	result, err := h.db.ExecContext(ctx, "DELETE FROM ip_allowlist WHERE id = $1", id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete IP allowlist entry"))
		return
	}

	rowsAff, _ := result.RowsAffected()
	if rowsAff == 0 {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
		return
	}

	// Invalidate cache
	if h.ipAllowlistMiddleware != nil {
		h.ipAllowlistMiddleware.HandleInvalidateCache(w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleIPAllowlist toggles the active status of an IP allowlist entry
// PATCH /v1/admin/ip-allowlist/:id/toggle
func (h *AdminIPAllowlistHandler) HandleToggleIPAllowlist(w http.ResponseWriter, r *http.Request) {
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

	// Toggle the is_active status
	var entry IPAccessEntry
	var description string
	var createdBy sql.NullString

	err = h.db.QueryRowContext(ctx, `
		UPDATE ip_allowlist
		SET is_active = NOT is_active, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, cidr, COALESCE(description, ''), is_active, is_whitelist, created_by, created_at, updated_at`,
		id).Scan(
		&entry.ID, &entry.Name, &entry.CIDR, &description,
		&entry.IsActive, &entry.IsWhitelist, &createdBy,
		&entry.CreatedAt, &entry.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
			return
		}
		logrus.WithError(err).Error("Failed to toggle IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to toggle IP allowlist entry"))
		return
	}

	entry.Description = description
	if createdBy.Valid {
		uid, _ := uuid.Parse(createdBy.String)
		entry.CreatedBy = &uid
	}

	// Invalidate cache
	if h.ipAllowlistMiddleware != nil {
		h.ipAllowlistMiddleware.HandleInvalidateCache(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleCheckMyIPAccess checks if the current user's IP is allowed
// GET /v1/admin/ip-allowlist/check-access
func (h *AdminIPAllowlistHandler) HandleCheckMyIPAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	clientIP := middleware.ExtractClientIP(r)

	if h.ipAllowlistMiddleware == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ip":        clientIP,
			"allowed":   true,
			"reason":    "IP allowlist not configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	allowed, err := h.ipAllowlistMiddleware.IsIPAllowed(ctx, clientIP)
	if err != nil {
		logrus.WithError(err).Warn("Error checking IP access")
		allowed = true // Fail open
	}

	response := map[string]interface{}{
		"ip":      clientIP,
		"allowed": allowed,
	}

	if !allowed {
		response["reason"] = "IP address not in allowlist"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

