package apikeys

import (
	"context"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleList handles GET /api/v1/api-keys
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Parse query parameters
	filters := &apikey.ListFilters{
		TenantID: &claims.TenantID,
	}

	// key_type filter
	if keyType := r.URL.Query().Get("key_type"); keyType != "" {
		kt := apikey.KeyType(keyType)
		filters.KeyType = &kt
	}

	// is_active filter
	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		active := isActive == "true"
		filters.IsActive = &active
	}

	// search filter
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	// team_id filter
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		if teamIDStr == "personal" {
			filters.TeamIDIsNull = true
		} else if teamID, err := uuid.Parse(teamIDStr); err == nil {
			filters.TeamID = &teamID
		}
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Fetch API keys
	ctx := context.Background()
	keys, total, err := h.repo.List(ctx, filters, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list API keys")
		return
	}

	// Build response (without plaintext!)
	var items []interface{}
	for _, key := range keys {
		items = append(items, key.ToResponse())
	}

	// Calculate pagination meta
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := map[string]interface{}{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}

	h.writeSuccess(w, items, meta)
}

// RegisterListRoutes registers the list route
func RegisterListRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo, nil)
	router.HandleFunc("/api-keys", h.HandleList).Methods("GET", "OPTIONS")
}
