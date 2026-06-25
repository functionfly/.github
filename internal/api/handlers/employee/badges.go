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

func (h *Handler) HandleCreateBadge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	var req struct {
		Slug        string  `json:"slug"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		IconURL     *string `json:"icon_url"`
		Category    string  `json:"category"`
		Points      int     `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Slug == "" || req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("slug and name are required"))
		return
	}

	badge, err := h.repo.CreateDigitalBadge(r.Context(), &storage.DigitalBadge{
		TenantID:    claims.TenantID,
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		Category:    req.Category,
		Points:      req.Points,
		IsActive:    true,
	})
	if err != nil {
		h.log.WithError(err).Error("Failed to create badge")
		apierror.WriteError(w, apierror.NewInternal("Failed to create badge"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"badge":   badge,
		"success": true,
	})
}

func (h *Handler) HandleListBadges(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListBadgesOpts{
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
	if c := q.Get("category"); c != "" {
		opts.Category = &c
	}

	badges, total, err := h.repo.ListDigitalBadges(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list badges")
		apierror.WriteError(w, apierror.NewInternal("Failed to list badges"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"badges": badges,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

type awardBadgeRequest struct {
	BadgeSlug  string  `json:"badge_slug"`
	EmployeeID string  `json:"employee_id"`
	AwardedBy  *string `json:"awarded_by,omitempty"`
}

func (h *Handler) HandleAwardBadge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req awardBadgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.BadgeSlug == "" || req.EmployeeID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("badge_slug and employee_id are required"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	badge, err := h.repo.GetDigitalBadgeBySlug(r.Context(), claims.TenantID, req.BadgeSlug)
	if err != nil || badge == nil {
		apierror.WriteError(w, apierror.NewNotFound("Badge not found"))
		return
	}

	eb := &storage.EmployeeBadge{
		EmployeeID: employeeID,
		BadgeID:    badge.ID,
	}
	if req.AwardedBy != nil {
		aid, err := uuid.Parse(*req.AwardedBy)
		if err == nil {
			eb.AwardedBy = &aid
		}
	}

	awarded, err := h.repo.AwardEmployeeBadge(r.Context(), eb)
	if err != nil {
		h.log.WithError(err).Error("Failed to award badge")
		apierror.WriteError(w, apierror.NewInternal("Failed to award badge"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"badge_award": awarded,
	})
}

func (h *Handler) HandleGetMyBadges(w http.ResponseWriter, r *http.Request) {
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

	badges, err := h.repo.GetEmployeeBadges(r.Context(), emp.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee badges")
		apierror.WriteError(w, apierror.NewInternal("Failed to get badges"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"badges": badges,
	})
}

func (h *Handler) HandleRevokeBadge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeID, err := uuid.Parse(mux.Vars(r)["employeeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	badgeID, err := uuid.Parse(mux.Vars(r)["badgeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid badge ID"))
		return
	}

	if err := h.repo.RevokeEmployeeBadge(r.Context(), employeeID, badgeID); err != nil {
		h.log.WithError(err).Error("Failed to revoke badge")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke badge"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
