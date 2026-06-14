package registry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetMCPSettings serves GET /v1/functions/{author}/{name}/mcp.
// Returns the per-function MCP configuration (or 404 if not configured).
// Auth required.
func (h *Handler) HandleGetMCPSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil || fn == nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}
	if !userOwnsMCPSettings(user, fn) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	settings, err := h.repo.GetMCPSettings(r.Context(), fn.ID)
	if err != nil {
		logrus.WithError(err).Error("failed to get MCP settings")
		apierror.WriteError(w, apierror.NewInternal("Internal error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if settings == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"function_id":         fn.ID,
			"enabled":             false,
			"transports":          []string{"streamable-http"},
			"expose_input_schema": true,
			"rate_limit_per_min":  60,
		})
		return
	}
	_ = json.NewEncoder(w).Encode(settings)
}

// HandleUpdateMCPSettings serves PATCH /v1/functions/{author}/{name}/mcp.
// Accepts a partial MCPSettingsInput in the body. Auth required.
func (h *Handler) HandleUpdateMCPSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil || fn == nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}
	if !userOwnsMCPSettings(user, fn) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 16*1024))
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Failed to read body"))
		return
	}
	defer r.Body.Close()

	var in registry.MCPSettingsInput
	if err := json.Unmarshal(body, &in); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON: "+err.Error()))
		return
	}
	if err := in.Validate(); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid MCP settings: "+err.Error()))
		return
	}

	actorID := user.UserID
	settings, err := h.repo.UpsertMCPSettings(r.Context(), fn.ID, in, &actorID)
	if err != nil {
		logrus.WithError(err).Error("failed to update MCP settings")
		apierror.WriteError(w, apierror.NewInternal("Internal error"))
		return
	}

	// Invalidate caches so the change is visible immediately. We invalidate
	// the per-function cache and the registry Redis search/list caches
	// (via the public GetRegistryCache() helper) to keep the /v1/mcp/tools
	// index and /.well-known/functionfly.json fresh.
	if h.cacheService != nil {
		go func() {
			ctx := r.Context()
			_ = h.cacheService.InvalidateFunction(fn.ID.String())
			if rc := h.cacheService.GetRegistryCache(); rc != nil {
				_ = rc.InvalidateSearchResults(ctx)
				_ = rc.InvalidateListResults(ctx)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

// userOwnsMCPSettings returns true if the user is allowed to mutate the
// function's MCP settings. Rules:
//
//   - FunctionFly platform admin (Role == "admin" / "platform_admin") — yes.
//   - Function owner (OwnerUserID == user.UserID) — yes.
//   - Same-tenant user with PermissionPublish or write capability — yes.
//   - "functionfly"-authored function — only platform admins.
//
// This matches the spirit of the existing publish auth check, but kept
// deliberately narrow: MCP settings are sensitive (rate limit, transport
// allowlist, override) so we default to deny.
func userOwnsMCPSettings(user *auth.Claims, fn *registry.RegistryFunction) bool {
	if user == nil || fn == nil {
		return false
	}
	// Platform admins can manage anything.
	if user.Role == "admin" || user.Role == "platform_admin" {
		return true
	}
	// The owner can manage their own function.
	if fn.OwnerUserID != nil && *fn.OwnerUserID == user.UserID {
		return true
	}
	// Same-tenant users with explicit publish permission.
	if fn.TenantID != nil && *fn.TenantID == user.TenantID {
		for _, p := range user.Permissions {
			if p == "publish" || p == "write" || p == "functions:write" {
				return true
			}
		}
	}
	// Public "functionfly" namespace is admin-only.
	if fn.Author == "functionfly" {
		return false
	}
	// Default: deny.
	_ = uuid.Nil
	return false
}
