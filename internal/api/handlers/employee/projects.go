package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListProjects(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListProjectsOpts{
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
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if s := q.Get("search"); s != "" {
		opts.Search = &s
	}

	projects, total, err := h.repo.ListProjects(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list projects")
		apierror.WriteError(w, apierror.NewInternal("Failed to list projects"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"projects": projects,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}

func (h *Handler) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid project ID"))
		return
	}

	project, err := h.repo.GetProjectByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get project")
		apierror.WriteError(w, apierror.NewInternal("Failed to get project"))
		return
	}
	if project == nil {
		apierror.WriteError(w, apierror.NewNotFound("Project not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"project": project,
	})
}

type createProjectRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	DepartmentID string   `json:"department_id,omitempty"`
	LeadID       string   `json:"lead_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

func (h *Handler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Project name is required"))
		return
	}

	now := time.Now()
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug) + "-" + uuid.New().String()[:8]

	// Resolve employee ID from user ID
	emp, _ := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	ownerID := claims.UserID
	if emp != nil {
		ownerID = emp.ID
	}

	proj := &types.Project{
		ID:        uuid.New(),
		TenantID:  claims.TenantID,
		Name:      req.Name,
		Slug:      slug,
		OwnerID:   ownerID,
		Status:    "active",
		Priority:  "medium",
		Tags:      types.JSONMap{},
		Metadata:  types.JSONMap{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Description != "" {
		proj.Description = &req.Description
	}

	created, err := h.repo.CreateProject(r.Context(), proj)
	if err != nil {
		h.log.WithError(err).Error("Failed to create project")
		apierror.WriteError(w, apierror.NewInternal("Failed to create project"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"project": created,
	})
}

type updateProjectRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Status      *string  `json:"status,omitempty"`
	LeadID      *string  `json:"lead_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (h *Handler) HandleUpdateProject(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid project ID"))
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Tags != nil {
		updates["tags"] = types.JSONMap{"tags": req.Tags}
	}
	updates["updated_at"] = time.Now()

	if len(updates) == 1 {
		apierror.WriteError(w, apierror.NewBadRequest("No fields to update"))
		return
	}

	if err := h.repo.UpdateProject(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update project")
		apierror.WriteError(w, apierror.NewInternal("Failed to update project"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
