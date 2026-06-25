package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListPackages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListPackageRegistryOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if t := q.Get("registry_type"); t != "" {
		opts.RegistryType = &t
	}
	if s := q.Get("search"); s != "" {
		opts.Search = &s
	}

	packages, total, err := h.repo.ListPackages(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list packages")
		apierror.WriteError(w, apierror.NewInternal("Failed to list packages"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"packages": packages,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}

func (h *Handler) HandleGetPackage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid package ID"))
		return
	}

	pkg, err := h.repo.GetPackageByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get package")
		apierror.WriteError(w, apierror.NewInternal("Failed to get package"))
		return
	}
	if pkg == nil {
		apierror.WriteError(w, apierror.NewNotFound("Package not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"package": pkg,
	})
}

type publishPackageRequest struct {
	Name          string  `json:"name"`
	Scope         *string `json:"scope,omitempty"`
	Description   *string `json:"description,omitempty"`
	RegistryType  string  `json:"registry_type,omitempty"`
	RepositoryURL *string `json:"repository_url,omitempty"`
	IsInternal    *bool   `json:"is_internal,omitempty"`
}

func (h *Handler) HandlePublishPackage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req publishPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name is required"))
		return
	}

	registryType := "npm"
	if req.RegistryType != "" {
		registryType = req.RegistryType
	}

	pkg := &storage.PackageRegistry{
		TenantID:     claims.TenantID,
		Name:         req.Name,
		Scope:        req.Scope,
		Description:  req.Description,
		RegistryType: registryType,
		IsInternal:   true,
		PublishedBy:  &emp.ID,
	}
	if req.RepositoryURL != nil {
		pkg.RepositoryURL = req.RepositoryURL
	}
	if req.IsInternal != nil {
		pkg.IsInternal = *req.IsInternal
	}

	created, err := h.repo.CreatePackage(r.Context(), pkg)
	if err != nil {
		h.log.WithError(err).Error("Failed to publish package")
		apierror.WriteError(w, apierror.NewInternal("Failed to publish package"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"package": created,
	})
}

func (h *Handler) HandleListVersions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid package ID"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListPackageVersionsOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}

	versions, total, err := h.repo.ListPackageVersions(r.Context(), id, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list package versions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list package versions"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": versions,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}
