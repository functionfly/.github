package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
)

func (h *Handler) HandleGetTeamHealth(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListTeamHealthOpts{
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
	if d := q.Get("department_id"); d != "" {
		if n, err := strconv.ParseInt(d, 10, 64); err == nil {
			opts.DepartmentID = &n
		}
	}

	metrics, total, err := h.repo.GetTeamHealthMetrics(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to get team health metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team health metrics"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metrics,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

func (h *Handler) HandleGetBurnoutRisk(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	deptIDStr := r.URL.Query().Get("department_id")
	if deptIDStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("department_id is required"))
		return
	}

	deptID, err := strconv.ParseInt(deptIDStr, 10, 64)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid department_id"))
		return
	}

	recalculate := r.URL.Query().Get("recalculate") == "true"
	if recalculate {
		metric, err := h.repo.CalculateTeamHealth(r.Context(), claims.TenantID, deptID)
		if err != nil {
			h.log.WithError(err).Error("Failed to calculate team health")
			apierror.WriteError(w, apierror.NewInternal("Failed to calculate team health"))
			return
		}

		saved, err := h.repo.CreateTeamHealthMetric(r.Context(), metric)
		if err != nil {
			h.log.WithError(err).Error("Failed to save team health metric")
			apierror.WriteError(w, apierror.NewInternal("Failed to save team health"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metric": saved,
		})
		return
	}

	metric, err := h.repo.GetLatestTeamHealth(r.Context(), claims.TenantID, deptID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get latest team health")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team health"))
		return
	}
	if metric == nil {
		apierror.WriteError(w, apierror.NewNotFound("No team health data found. Use ?recalculate=true to generate."))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metric": metric,
	})
}
