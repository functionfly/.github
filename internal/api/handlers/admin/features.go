package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListFeatures lists all available features
func (h *Handler) HandleListFeatures(w http.ResponseWriter, r *http.Request) {
	features := plans.GetAllFeatures()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"features": features,
	})
}

// HandleListPlanFeatures lists features for a specific plan
func (h *Handler) HandleListPlanFeatures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plan := vars["plan"]

	if plan == "" {
		http.Error(w, "Plan is required", http.StatusBadRequest)
		return
	}

	// Validate plan exists
	features := plans.GetFeaturesForPlan(plan)
	if features == nil {
		http.Error(w, "Invalid plan", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":     plan,
		"features": features,
	})
}

// HandleGetPlanInfo gets detailed information about a plan
func (h *Handler) HandleGetPlanInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plan := vars["plan"]

	if plan == "" {
		http.Error(w, "Plan is required", http.StatusBadRequest)
		return
	}

	features := plans.GetFeaturesForPlan(plan)
	if features == nil {
		http.Error(w, "Invalid plan", http.StatusBadRequest)
		return
	}

	// Get detailed feature information
	featureDetails := make([]plans.Feature, 0, len(features))
	for _, f := range features {
		if def, ok := plans.GetFeatureDefinition(f); ok {
			featureDetails = append(featureDetails, def)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":              plan,
		"features":          features,
		"feature_details":   featureDetails,
		"feature_count":    len(features),
	})
}

// HandleCheckFeature checks if a plan has a specific feature
func (h *Handler) HandleCheckFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plan := vars["plan"]
	feature := vars["feature"]

	if plan == "" || feature == "" {
		http.Error(w, "Plan and feature are required", http.StatusBadRequest)
		return
	}

	checker := plans.NewFeatureChecker(plan)
	available := checker.HasFeature(feature)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":         plan,
		"feature":      feature,
		"available":    available,
		"required_plan": plans.GetRequiredPlan(feature),
	})
}

// HandleGetAllPlansInfo gets information about all plans
func (h *Handler) HandleGetAllPlansInfo(w http.ResponseWriter, r *http.Request) {
	plansInfo := plans.GetAllPlanInfo()

	// Build response with detailed information
	type PlanDetail struct {
		Plan           string   `json:"plan"`
		Features       []string `json:"features"`
		FeatureCount   int      `json:"feature_count"`
		IsEnterprise   bool     `json:"is_enterprise"`
		IsPro          bool     `json:"is_pro"`
		IsAgent        bool     `json:"is_agent"`
	}

	details := make([]PlanDetail, 0, len(plansInfo))
	for _, p := range plansInfo {
		detail := PlanDetail{
			Plan:         p.Plan,
			Features:     p.Features,
			FeatureCount: len(p.Features),
			IsEnterprise: p.Plan == plans.PlanEnterprise || p.Plan == plans.PlanAgentEnterprise,
			IsPro:        p.Plan == plans.PlanPro || p.Plan == plans.PlanAgentPro,
			IsAgent:      plans.IsAgentTier(p.Plan),
		}
		details = append(details, detail)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plans": details,
	})
}

// HandleGetFeatureDefinition gets the definition of a specific feature
func (h *Handler) HandleGetFeatureDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feature := vars["feature"]

	if feature == "" {
		http.Error(w, "Feature is required", http.StatusBadRequest)
		return
	}

	def, ok := plans.GetFeatureDefinition(feature)
	if !ok {
		http.Error(w, "Feature not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(def)
}

// HandleCheckTenantFeatures checks features for a specific tenant
func (h *Handler) HandleCheckTenantFeatures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	if tenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	// Get tenant from repository
	tenant, err := h.repo.GetTenantByID(tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}

	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	plan := tenant.Plan
	if plan == "" {
		plan = plans.PlanStarter
	}

	checker := plans.NewFeatureChecker(plan)

	// Parse requested features from query params
	features := r.URL.Query()["features"]
	if len(features) == 0 {
		// Return all features if none specified
		features = plans.GetFeaturesForPlan(plan)
	}

	results := checker.CheckFeatures(features)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":       tenantID,
		"plan":            plan,
		"plan_info":       plans.GetAllPlanInfo(),
		"feature_results": results,
	})
}
