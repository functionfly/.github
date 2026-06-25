package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
)

func (h *Handler) HandleGetSkillsGraph(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	limit := 50
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

	recalculate := q.Get("recalculate") == "true"
	if recalculate {
		skills, err := h.repo.CalculateSkillsGraph(r.Context(), claims.TenantID)
		if err != nil {
			h.log.WithError(err).Error("Failed to calculate skills graph")
			apierror.WriteError(w, apierror.NewInternal("Failed to calculate skills graph"))
			return
		}

		for _, s := range skills {
			h.repo.CreateSkillsGraphEntry(r.Context(), s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"skills": skills,
			"total":  len(skills),
		})
		return
	}

	skills, total, err := h.repo.GetSkillsGraph(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to get skills graph")
		apierror.WriteError(w, apierror.NewInternal("Failed to get skills graph"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": skills,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetSkillGap(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	gaps, err := h.repo.GetSkillGaps(r.Context(), claims.TenantID, limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to get skill gaps")
		apierror.WriteError(w, apierror.NewInternal("Failed to get skill gaps"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gaps":  gaps,
		"total": len(gaps),
	})
}
