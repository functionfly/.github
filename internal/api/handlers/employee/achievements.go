package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListAchievementDefinitions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	defs, err := h.repo.ListAchievementDefinitions(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to list achievement definitions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list achievements"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"achievements": defs,
	})
}

func (h *Handler) HandleGetAchievementProgress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeIDStr := mux.Vars(r)["employeeId"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	progress, err := h.repo.GetAchievementProgress(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get achievement progress")
		apierror.WriteError(w, apierror.NewInternal("Failed to get achievement progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"progress": progress,
	})
}

func (h *Handler) HandleCheckAchievements(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeIDStr := mux.Vars(r)["employeeId"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	if err := h.repo.CheckAndAwardAchievements(r.Context(), employeeID); err != nil {
		h.log.WithError(err).Error("Failed to check achievements")
		apierror.WriteError(w, apierror.NewInternal("Failed to check achievements"))
		return
	}

	progress, err := h.repo.GetAchievementProgress(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get achievement progress after check")
		apierror.WriteError(w, apierror.NewInternal("Failed to get achievement progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"progress": progress,
	})
}

type createAchievementDefinitionRequest struct {
	Slug              string  `json:"slug"`
	Name              string  `json:"name"`
	Description       *string `json:"description,omitempty"`
	Icon              *string `json:"icon,omitempty"`
	Category          string  `json:"category"`
	CriteriaType      string  `json:"criteria_type"`
	CriteriaThreshold int     `json:"criteria_threshold"`
	Points            int     `json:"points"`
	Tier              int     `json:"tier"`
}

func (h *Handler) HandleCreateAchievementDefinition(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	var req createAchievementDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Slug == "" || req.Name == "" || req.CriteriaType == "" || req.CriteriaThreshold <= 0 {
		apierror.WriteError(w, apierror.NewBadRequest("slug, name, criteria_type, and criteria_threshold are required"))
		return
	}

	def := &types.AchievementDefinition{
		TenantID:          claims.TenantID,
		Slug:              req.Slug,
		Name:              req.Name,
		Description:       req.Description,
		Icon:              req.Icon,
		Category:          req.Category,
		CriteriaType:      req.CriteriaType,
		CriteriaThreshold: req.CriteriaThreshold,
		Points:            req.Points,
		Tier:              req.Tier,
		IsActive:          true,
	}

	created, err := h.repo.CreateAchievementDefinition(r.Context(), def)
	if err != nil {
		h.log.WithError(err).Error("Failed to create achievement definition")
		apierror.WriteError(w, apierror.NewInternal("Failed to create achievement"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"achievement": created,
	})
}

func (h *Handler) HandleSeedAchievements(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	if err := h.repo.SeedAchievementDefinitions(r.Context(), claims.TenantID); err != nil {
		h.log.WithError(err).Error("Failed to seed achievement definitions")
		apierror.WriteError(w, apierror.NewInternal("Failed to seed achievements"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Achievement definitions seeded successfully",
	})
}
