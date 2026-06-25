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

func (h *Handler) HandleListOpportunities(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListMarketplaceOpportunitiesOpts{
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
	if t := q.Get("type"); t != "" {
		opts.Type = &t
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}

	opps, total, err := h.repo.ListMarketplaceOpportunities(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list opportunities")
		apierror.WriteError(w, apierror.NewInternal("Failed to list opportunities"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"opportunities": opps,
		"total":         total,
		"limit":         opts.Limit,
		"offset":        opts.Offset,
	})
}

type createOpportunityRequest struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	OpportunityType string   `json:"opportunity_type,omitempty"`
	SkillsRequired  []string `json:"skills_required,omitempty"`
	HoursPerWeek    *float64 `json:"hours_per_week,omitempty"`
	DurationWeeks   *int     `json:"duration_weeks,omitempty"`
	IsRemote        *bool    `json:"is_remote,omitempty"`
	MaxApplicants   *int     `json:"max_applicants,omitempty"`
	DepartmentID    *int64   `json:"department_id,omitempty"`
}

func (h *Handler) HandleCreateOpportunity(w http.ResponseWriter, r *http.Request) {
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

	var req createOpportunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" || req.Description == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title and description are required"))
		return
	}

	opp := &storage.MarketplaceOpportunity{
		TenantID:        claims.TenantID,
		PostedBy:        emp.ID,
		Title:           req.Title,
		Description:     req.Description,
		OpportunityType: "project",
		Status:          "open",
		IsRemote:        true,
		DepartmentID:    req.DepartmentID,
		HoursPerWeek:    req.HoursPerWeek,
		DurationWeeks:   req.DurationWeeks,
		MaxApplicants:   req.MaxApplicants,
	}
	if req.OpportunityType != "" {
		opp.OpportunityType = req.OpportunityType
	}
	if req.IsRemote != nil {
		opp.IsRemote = *req.IsRemote
	}
	if req.SkillsRequired != nil {
		opp.SkillsRequired = storage.JSONMap{"skills": req.SkillsRequired}
	}

	created, err := h.repo.CreateMarketplaceOpportunity(r.Context(), opp)
	if err != nil {
		h.log.WithError(err).Error("Failed to create opportunity")
		apierror.WriteError(w, apierror.NewInternal("Failed to create opportunity"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"opportunity": created,
	})
}

type applyOpportunityRequest struct {
	Message string `json:"message,omitempty"`
}

func (h *Handler) HandleApplyOpportunity(w http.ResponseWriter, r *http.Request) {
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

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid opportunity ID"))
		return
	}

	opp, err := h.repo.GetMarketplaceOpportunityByID(r.Context(), id)
	if err != nil || opp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Opportunity not found"))
		return
	}
	if opp.Status != "open" {
		apierror.WriteError(w, apierror.NewBadRequest("Opportunity is not open for applications"))
		return
	}

	var req applyOpportunityRequest
	json.NewDecoder(r.Body).Decode(&req)

	app := &storage.MarketplaceApplication{
		OpportunityID: id,
		ApplicantID:   emp.ID,
		Status:        "pending",
	}
	if req.Message != "" {
		app.Message = &req.Message
	}

	created, err := h.repo.CreateMarketplaceApplication(r.Context(), app)
	if err != nil {
		h.log.WithError(err).Error("Failed to create application")
		apierror.WriteError(w, apierror.NewInternal("Failed to create application"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"application": created,
	})
}

type reviewApplicationRequest struct {
	Status string `json:"status"`
}

func (h *Handler) HandleReviewApplication(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	appID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid application ID"))
		return
	}

	app, err := h.repo.GetMarketplaceApplicationByID(r.Context(), appID)
	if err != nil || app == nil {
		apierror.WriteError(w, apierror.NewNotFound("Application not found"))
		return
	}

	var req reviewApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Status != "accepted" && req.Status != "rejected" {
		apierror.WriteError(w, apierror.NewBadRequest("Status must be accepted or rejected"))
		return
	}

	if err := h.repo.UpdateMarketplaceApplicationStatus(r.Context(), appID, req.Status); err != nil {
		h.log.WithError(err).Error("Failed to review application")
		apierror.WriteError(w, apierror.NewInternal("Failed to review application"))
		return
	}

	if req.Status == "accepted" {
		opp, err := h.repo.GetMarketplaceOpportunityByID(r.Context(), app.OpportunityID)
		if err == nil && opp != nil {
			h.repo.UpdateMarketplaceOpportunity(r.Context(), opp.ID, map[string]interface{}{
				"status": "filled",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
