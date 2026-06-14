package version

import (
	"context"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewInternal("Registry not configured"))
		return nil, nil, false
	}

	fn, err := h.registryRepo.GetFunctionByID(context.Background(), functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return nil, nil, false
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to load function"))
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
		apierror.WriteError(w, apierror.NewInternal("Registry not configured"))
		return nil, claims, false
	}

	fn, err := h.registryRepo.GetFunctionByID(context.Background(), functionID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "no rows") {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return nil, claims, false
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to load function"))
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
