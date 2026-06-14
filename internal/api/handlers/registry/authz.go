package registry

import (
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// IsPlatformAdmin returns true for platform admin roles.
func IsPlatformAdmin(claims *auth.Claims) bool {
	if claims == nil {
		return false
	}
	return claims.Role == "admin" || claims.Role == "super_admin"
}

// CanViewRegistryFunction reports whether the caller may read function metadata and non-source resources.
// A function with trust_score = 0 is fully revoked and cannot be viewed or invoked.
func CanViewRegistryFunction(fn *storageregistry.RegistryFunction, claims *auth.Claims) bool {
	if fn == nil {
		return false
	}
	// Fully revoked functions (trust_score = 0) are invisible to everyone except platform admins.
	if fn.TrustScore == 0 && !IsPlatformAdmin(claims) {
		return false
	}
	vis := strings.ToLower(strings.TrimSpace(fn.Visibility))
	switch vis {
	case "", "public", "unlisted":
		return true
	case "private":
		if claims == nil {
			return false
		}
		if IsPlatformAdmin(claims) {
			return true
		}
		if fn.OwnerUserID != nil && *fn.OwnerUserID == claims.UserID {
			return true
		}
		if fn.TenantID != nil && claims.TenantID == *fn.TenantID {
			return true
		}
		return false
	default:
		return true
	}
}

// IsRegistryFunctionOwner reports whether the caller may mutate a registry function.
func IsRegistryFunctionOwner(fn *storageregistry.RegistryFunction, claims *auth.Claims) bool {
	if fn == nil || claims == nil {
		return false
	}
	if IsPlatformAdmin(claims) {
		return true
	}
	if fn.OwnerUserID == nil {
		return false
	}
	return *fn.OwnerUserID == claims.UserID
}

// AuthorMatchesPublisher validates that the authenticated user may publish under the given author namespace.
func AuthorMatchesPublisher(claims *auth.Claims, author string) bool {
	if claims == nil || strings.TrimSpace(author) == "" {
		return false
	}
	if IsPlatformAdmin(claims) {
		return true
	}
	if strings.EqualFold(author, "functionfly") {
		return false
	}
	if claims.Username != "" && strings.EqualFold(claims.Username, author) {
		return true
	}
	return strings.EqualFold(claims.Email, author)
}

// WriteForbidden writes a 403 response.
func WriteForbidden(w http.ResponseWriter) {
	apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
}

// WriteUnauthorized writes a 401 response.
func WriteUnauthorized(w http.ResponseWriter) {
	apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
}

// RequireRegistryFunctionOwner loads a function by ID and ensures the caller owns it (or is admin).
func (h *Handler) RequireRegistryFunctionOwner(
	w http.ResponseWriter,
	r *http.Request,
	functionID uuid.UUID,
) (*storageregistry.RegistryFunction, *auth.Claims, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		WriteUnauthorized(w)
		return nil, nil, false
	}

	fn, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return nil, nil, false
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to load function"))
		return nil, nil, false
	}

	if !IsRegistryFunctionOwner(fn, claims) {
		WriteForbidden(w)
		return nil, nil, false
	}

	return fn, claims, true
}

// RequireRegistryFunctionView loads a function and ensures the caller may read it.
func (h *Handler) RequireRegistryFunctionView(
	w http.ResponseWriter,
	r *http.Request,
	functionID uuid.UUID,
) (*storageregistry.RegistryFunction, *auth.Claims, bool) {
	claims := middleware.GetUserFromContext(r)
	// OptionalAuth on /v1 may attach claims; public routes still work without auth for public functions.

	fn, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return nil, claims, false
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to load function"))
		return nil, claims, false
	}

	if !CanViewRegistryFunction(fn, claims) {
		WriteForbidden(w)
		return nil, claims, false
	}

	return fn, claims, true
}

// EnrichFunctionInfoForViewer adds owner hints for the authenticated viewer (dashboard management UI).
func EnrichFunctionInfoForViewer(info map[string]interface{}, fn *storageregistry.RegistryFunction, claims *auth.Claims) {
	if info == nil || fn == nil {
		return
	}
	info["visibility"] = fn.Visibility
	if claims != nil && fn.OwnerUserID != nil && *fn.OwnerUserID == claims.UserID {
		info["is_owner"] = true
		info["owner_user_id"] = fn.OwnerUserID.String()
	} else if claims != nil && IsPlatformAdmin(claims) {
		info["is_owner"] = false
		info["is_platform_admin"] = true
	}
}
