package version

import (
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

func (h *Handler) requireFunctionOwner(
	w http.ResponseWriter,
	r *http.Request,
	functionID uuid.UUID,
) (*storageregistry.RegistryFunction, *auth.Claims, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		registry.WriteUnauthorized(w)
		return nil, nil, false
	}
	if h.registryRepo == nil {
		http.Error(w, "Registry not configured", http.StatusInternalServerError)
		return nil, nil, false
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return nil, nil, false
		}
		http.Error(w, "Failed to load function", http.StatusInternalServerError)
		return nil, nil, false
	}

	if !registry.IsRegistryFunctionOwner(fn, claims) {
		registry.WriteForbidden(w)
		return nil, nil, false
	}

	return fn, claims, true
}

func (h *Handler) requireFunctionView(
	w http.ResponseWriter,
	r *http.Request,
	functionID uuid.UUID,
) (*storageregistry.RegistryFunction, *auth.Claims, bool) {
	claims := middleware.GetUserFromContext(r)

	if h.registryRepo == nil {
		http.Error(w, "Registry not configured", http.StatusInternalServerError)
		return nil, claims, false
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return nil, claims, false
		}
		http.Error(w, "Failed to load function", http.StatusInternalServerError)
		return nil, claims, false
	}

	if !registry.CanViewRegistryFunction(fn, claims) {
		registry.WriteForbidden(w)
		return nil, claims, false
	}

	return fn, claims, true
}

func initiatedByFromRequest(r *http.Request) *uuid.UUID {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return nil
	}
	id := claims.UserID
	return &id
}
