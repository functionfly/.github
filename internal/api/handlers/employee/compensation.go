package employee

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetCompensation(w http.ResponseWriter, r *http.Request) {
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

	isAdmin := claims.HasPermission("admin") || claims.Role == "admin"
	isSelf := claims.UserID == employeeID
	if !isAdmin && !isSelf {
		apierror.WriteError(w, apierror.NewForbidden("Admin or self access required"))
		return
	}

	h.log.WithFields(map[string]interface{}{
		"viewer_id":   claims.UserID,
		"employee_id": employeeID,
		"is_admin":    isAdmin,
	}).Info("Compensation data accessed")

	comp, err := h.repo.GetActiveCompensation(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get compensation")
		apierror.WriteError(w, apierror.NewInternal("Failed to get compensation"))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()
	_ = h.repo.LogCompensationAccess(r.Context(), &types.CompensationAccessLog{
		AccessorID: claims.UserID,
		TargetID:   employeeID,
		Action:     "view",
		IPAddress:  &ip,
		UserAgent:  &ua,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"compensation": comp,
	})
}

type updateCompensationRequest struct {
	BaseSalary    *int64  `json:"base_salary,omitempty"`
	Currency      *string `json:"currency,omitempty"`
	BonusTarget   *int64  `json:"bonus_target,omitempty"`
	EquityPackage *string `json:"equity_package,omitempty"`
	EffectiveDate *string `json:"effective_date,omitempty"`
}

func (h *Handler) HandleUpdateCompensation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	employeeID, err := uuid.Parse(mux.Vars(r)["employeeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	var req updateCompensationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.BaseSalary == nil {
		apierror.WriteError(w, apierror.NewBadRequest("base_salary is required"))
		return
	}

	currency := "USD"
	if req.Currency != nil {
		currency = *req.Currency
	}

	effDate := time.Now()
	if req.EffectiveDate != nil {
		if t, err := time.Parse("2006-01-02", *req.EffectiveDate); err == nil {
			effDate = t
		}
	}

	rec := &types.CompensationRecord{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		TenantID:        claims.TenantID,
		BaseSalaryCents: *req.BaseSalary,
		Currency:        currency,
		PayFrequency:    "biweekly",
		EffectiveDate:   effDate,
		CreatedBy:       claims.UserID,
	}

	created, err := h.repo.CreateCompensationRecord(r.Context(), rec)
	if err != nil {
		h.log.WithError(err).Error("Failed to update compensation")
		apierror.WriteError(w, apierror.NewInternal("Failed to update compensation"))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()
	_ = h.repo.LogCompensationAccess(r.Context(), &types.CompensationAccessLog{
		AccessorID: claims.UserID,
		TargetID:   employeeID,
		Action:     "update",
		IPAddress:  &ip,
		UserAgent:  &ua,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"compensation": created,
	})
}

func (h *Handler) HandleListEquityGrants(w http.ResponseWriter, r *http.Request) {
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

	isAdmin := claims.HasPermission("admin") || claims.Role == "admin"
	isSelf := claims.UserID == employeeID
	if !isAdmin && !isSelf {
		apierror.WriteError(w, apierror.NewForbidden("Admin or self access required"))
		return
	}

	grants, err := h.repo.ListEquityGrants(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to list equity grants")
		apierror.WriteError(w, apierror.NewInternal("Failed to list equity grants"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"equity_grants": grants,
	})
}

func (h *Handler) HandleGetCompensationAudit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	employeeID, err := uuid.Parse(mux.Vars(r)["employeeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	_ = employeeID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"audit_log": []interface{}{},
	})
}
