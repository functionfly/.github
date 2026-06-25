package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetOrgChart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	orgChart, err := h.repo.GetOrgChart(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get org chart")
		apierror.WriteError(w, apierror.NewInternal("Failed to get org chart"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"org_chart": orgChart,
	})
}

func (h *Handler) HandleGetDirectReports(w http.ResponseWriter, r *http.Request) {
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

	reports, err := h.repo.GetDirectReports(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get direct reports")
		apierror.WriteError(w, apierror.NewInternal("Failed to get direct reports"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
	})
}
