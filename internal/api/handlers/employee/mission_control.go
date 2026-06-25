package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
)

func (h *Handler) HandleGetMissionControl(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	snapshot, err := h.repo.GetLatestMissionControlSnapshot(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get mission control snapshot")
		apierror.WriteError(w, apierror.NewInternal("Failed to get mission control snapshot"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snapshot": snapshot,
	})
}

func (h *Handler) HandleRefreshSnapshot(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	snapshot, err := h.repo.GenerateMissionControlSnapshot(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate mission control snapshot")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate snapshot"))
		return
	}

	saved, err := h.repo.CreateMissionControlSnapshot(r.Context(), snapshot)
	if err != nil {
		h.log.WithError(err).Error("Failed to save mission control snapshot")
		apierror.WriteError(w, apierror.NewInternal("Failed to save snapshot"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snapshot": saved,
	})
}

func (h *Handler) HandleListSnapshots(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	limit := 20
	offset := 0
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	snapshots, total, err := h.repo.ListMissionControlSnapshots(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list snapshots")
		apierror.WriteError(w, apierror.NewInternal("Failed to list snapshots"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snapshots": snapshots,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}
