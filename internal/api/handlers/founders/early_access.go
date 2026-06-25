package founders

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

type EarlyAccessHandler struct {
	repo  storage.Repository
	log   *logrus.Logger
}

func NewEarlyAccessHandler(repo storage.Repository, log *logrus.Logger) *EarlyAccessHandler {
	return &EarlyAccessHandler{
		repo:  repo,
		log:   log,
	}
}

func (h *EarlyAccessHandler) HandleListFeatures(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	features, err := h.repo.GetFounderEarlyAccessFeatures(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to list early access features")
		http.Error(w, "failed to list features", http.StatusInternalServerError)
		return
	}

	userAccess, err := h.repo.GetUserFounderEarlyAccess(r.Context(), claims.UserID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get user early access")
		http.Error(w, "failed to get user access", http.StatusInternalServerError)
		return
	}

	claimedSlugs := make(map[string]bool)
	for _, access := range userAccess {
		claimedSlugs[access.FeatureSlug] = true
	}

	type FeatureResponse struct {
		Slug        string  `json:"slug"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		IsClaimed   bool    `json:"is_claimed"`
		LaunchedAt  *string `json:"launched_at,omitempty"`
	}

	responses := make([]FeatureResponse, 0, len(features))
	for _, feature := range features {
		resp := FeatureResponse{
			Slug:        feature.Slug,
			Name:        feature.Name,
			Description: feature.Description,
			IsClaimed:   claimedSlugs[feature.Slug],
		}
		if feature.LaunchedAt != nil {
			s := feature.LaunchedAt.Format("2006-01-02")
			resp.LaunchedAt = &s
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"features": responses,
	})
}

func (h *EarlyAccessHandler) HandleClaimAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	slug := r.URL.Query().Get(":slug")
	if slug == "" {
		http.Error(w, "feature slug is required", http.StatusBadRequest)
		return
	}

	feature, err := h.repo.GetFounderEarlyAccessFeatureBySlug(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get feature")
		http.Error(w, "failed to get feature", http.StatusInternalServerError)
		return
	}

	if feature == nil {
		http.Error(w, "feature not found", http.StatusNotFound)
		return
	}

	if !feature.IsActive {
		http.Error(w, "feature is not available", http.StatusBadRequest)
		return
	}

	hasClaimed, err := h.repo.HasUserClaimedEarlyAccess(r.Context(), claims.UserID, slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to check early access")
		http.Error(w, "failed to check access", http.StatusInternalServerError)
		return
	}

	if hasClaimed {
		http.Error(w, "already claimed", http.StatusConflict)
		return
	}

	if err := h.repo.ClaimFounderEarlyAccess(r.Context(), claims.UserID, feature); err != nil {
		h.log.WithError(err).Error("Failed to claim early access")
		http.Error(w, "failed to claim access", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Early access claimed successfully",
		"feature": map[string]interface{}{
			"slug": feature.Slug,
			"name": feature.Name,
		},
	})
}

func (h *EarlyAccessHandler) HandleGetUserAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	access, err := h.repo.GetUserFounderEarlyAccess(r.Context(), claims.UserID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get user early access")
		http.Error(w, "failed to get access", http.StatusInternalServerError)
		return
	}

	type AccessResponse struct {
		Slug       string    `json:"slug"`
		Name       string    `json:"feature_name"`
		ClaimedAt  string    `json:"claimed_at"`
	}

	responses := make([]AccessResponse, 0, len(access))
	for _, a := range access {
		responses = append(responses, AccessResponse{
			Slug:      a.FeatureSlug,
			Name:      a.FeatureName,
			ClaimedAt: a.AccessedAt.Format("2006-01-02"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"claimed_features": responses,
	})
}
