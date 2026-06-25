package employee

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListFlags(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListFeatureFlagsOpts{
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
	if e := q.Get("is_enabled"); e != "" {
		if e == "true" {
			t := true
			opts.IsEnabled = &t
		} else if e == "false" {
			f := false
			opts.IsEnabled = &f
		}
	}

	flags, total, err := h.repo.ListFeatureFlags(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list feature flags")
		apierror.WriteError(w, apierror.NewInternal("Failed to list feature flags"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flags":   flags,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

func (h *Handler) HandleGetFlag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	key := mux.Vars(r)["key"]
	if key == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Flag key is required"))
		return
	}

	flag, err := h.repo.GetFeatureFlagByKey(r.Context(), claims.TenantID, key)
	if err != nil {
		h.log.WithError(err).Error("Failed to get feature flag")
		apierror.WriteError(w, apierror.NewInternal("Failed to get feature flag"))
		return
	}
	if flag == nil {
		apierror.WriteError(w, apierror.NewNotFound("Feature flag not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flag": flag,
	})
}

type createFlagRequest struct {
	Key            string                 `json:"key"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	FlagType       string                 `json:"flag_type,omitempty"`
	IsEnabled      bool                   `json:"is_enabled,omitempty"`
	RolloutPct     int                    `json:"rollout_pct,omitempty"`
	Variants       map[string]interface{} `json:"variants,omitempty"`
	TargetAudience map[string]interface{} `json:"target_audience,omitempty"`
}

func (h *Handler) HandleCreateFlag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Key == "" || req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("key and name are required"))
		return
	}

	flag := &storage.FeatureFlag{
		TenantID:    claims.TenantID,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		FlagType:    "boolean",
		IsEnabled:   req.IsEnabled,
		RolloutPct:  req.RolloutPct,
	}
	if req.FlagType != "" {
		flag.FlagType = req.FlagType
	}
	if req.Variants != nil {
		flag.Variants = storage.JSONMap(req.Variants)
	}
	if req.TargetAudience != nil {
		flag.TargetAudience = storage.JSONMap(req.TargetAudience)
	}

	created, err := h.repo.CreateFeatureFlag(r.Context(), flag)
	if err != nil {
		h.log.WithError(err).Error("Failed to create feature flag")
		apierror.WriteError(w, apierror.NewInternal("Failed to create feature flag"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flag": created,
	})
}

func (h *Handler) HandleUpdateFlag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid flag ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if err := h.repo.UpdateFeatureFlag(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update feature flag")
		apierror.WriteError(w, apierror.NewInternal("Failed to update feature flag"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleEvaluateFlag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	key := mux.Vars(r)["key"]
	if key == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Flag key is required"))
		return
	}

	flag, err := h.repo.GetFeatureFlagByKey(r.Context(), claims.TenantID, key)
	if err != nil {
		h.log.WithError(err).Error("Failed to evaluate feature flag")
		apierror.WriteError(w, apierror.NewInternal("Failed to evaluate feature flag"))
		return
	}
	if flag == nil {
		apierror.WriteError(w, apierror.NewNotFound("Feature flag not found"))
		return
	}

	result := map[string]interface{}{
		"key":        flag.Key,
		"enabled":    false,
		"flag_type":  flag.FlagType,
	}

	if !flag.IsEnabled {
		result["enabled"] = false
	} else {
		switch flag.FlagType {
		case "boolean":
			result["enabled"] = true
		case "percentage":
			result["enabled"] = rand.Intn(100) < flag.RolloutPct
		case "variant":
			result["enabled"] = true
			if flag.Variants != nil {
				result["variants"] = flag.Variants
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
