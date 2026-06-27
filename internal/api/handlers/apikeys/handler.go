// Package apikeys provides HTTP handlers for API key management
package apikeys

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler contains API key HTTP handlers
type Handler struct {
	repo      *apikey.Repository
	tenantGetter TenantPlanGetter
	keyGen   *apikey.Generator
	hasher   *apikey.Hasher
	validate *apikey.Validator
}

// TenantPlanGetter provides tenant plan information (optional)
type TenantPlanGetter interface {
	GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error)
}

// NewHandler creates a new API key handler
// tenantGetter is optional - pass nil if tenant plan checking is not needed
func NewHandler(repo *apikey.Repository, tenantGetter TenantPlanGetter) *Handler {
	return &Handler{
		repo:        repo,
		tenantGetter: tenantGetter,
		keyGen:      apikey.NewGenerator(),
		hasher:      apikey.NewHasher(),
		validate:    apikey.NewValidator(),
	}
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeSuccess writes a success JSON response with the standard format
func (h *Handler) writeSuccess(w http.ResponseWriter, data interface{}, meta ...map[string]interface{}) {
	resp := map[string]interface{}{
		"data": data,
	}
	if len(meta) > 0 {
		resp["meta"] = meta[0]
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// writeError writes an error JSON response
func (h *Handler) writeError(w http.ResponseWriter, status int, code string, message string) {
	resp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	h.writeJSON(w, status, resp)
}

// getUserClaims extracts user claims from request context
func getUserClaims(r *http.Request) (*UserClaims, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return nil, false
	}
	return &UserClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
	}, true
}

// getIDFromPath extracts a UUID from the URL path variable
func getIDFromPath(r *http.Request, varName string) (uuid.UUID, bool, bool) {
	vars := extractPathVars(r)
	idStr, ok := vars[varName]
	if !ok {
		return uuid.Nil, false, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, false, true
	}
	return id, true, true
}

// UserClaims represents user information extracted from JWT
type UserClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Email    string
}

// extractPathVars extracts path variables from the request (e.g. "id" from /api-keys/{id}).
// Uses gorilla/mux route vars; returns a non-nil map (empty if request wasn't matched by mux).
func extractPathVars(r *http.Request) map[string]string {
	v := mux.Vars(r)
	if v == nil {
		return make(map[string]string)
	}
	return v
}
