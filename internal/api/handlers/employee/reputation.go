package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetReputation(w http.ResponseWriter, r *http.Request) {
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

	recalculate := r.URL.Query().Get("recalculate") == "true"
	if recalculate {
		scores, err := h.repo.CalculateReputation(r.Context(), employeeID, claims.TenantID)
		if err != nil {
			h.log.WithError(err).Error("Failed to calculate reputation")
			apierror.WriteError(w, apierror.NewInternal("Failed to calculate reputation"))
			return
		}

		for _, s := range scores {
			h.repo.CreateReputationScore(r.Context(), s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"scores": scores,
		})
		return
	}

	scores, err := h.repo.GetReputationScores(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get reputation scores")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation scores"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scores": scores,
	})
}

func (h *Handler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = "engineering"
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	scores, total, err := h.repo.GetReputationLeaderboard(r.Context(), claims.TenantID, category, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to get reputation leaderboard")
		apierror.WriteError(w, apierror.NewInternal("Failed to get leaderboard"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": scores,
		"category":    category,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (h *Handler) HandleGetReputationTrends(w http.ResponseWriter, r *http.Request) {
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

	category := r.URL.Query().Get("category")

	history, err := h.repo.GetReputationHistory(r.Context(), employeeID, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to get reputation history")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation trends"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}
