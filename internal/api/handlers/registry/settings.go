package registry

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
)

// FunctionSettingsResponse is the response for GET /v1/functions/{author}/{name}/settings
type FunctionSettingsResponse struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Author               string            `json:"author"`
	Description          string            `json:"description"`
	IsPublic             bool              `json:"isPublic"`
	IsPublished          bool              `json:"isPublished"`
	AllowAnonymousInvoke bool              `json:"allowAnonymousInvoke"`
	CorsEnabled          bool              `json:"corsEnabled"`
	CorsOrigins          []string          `json:"corsOrigins"`
	Timeout              int               `json:"timeout"`
	Memory               int               `json:"memory"`
	Runtime              string            `json:"runtime"`
	Providers            []string          `json:"providers"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
	Secrets              []string          `json:"secrets"`
	CustomDomains        []string          `json:"customDomains"`
}

// FunctionSettingsPatchRequest is the body for PATCH /v1/functions/{author}/{name}/settings
type FunctionSettingsPatchRequest struct {
	CustomDomains *[]string `json:"customDomains,omitempty"`
	// Other fields (description, cors, etc.) can be added here when needed
}

// HandleGetFunctionSettings returns GET /v1/functions/{author}/{name}/settings
func (h *Handler) HandleGetFunctionSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	customDomains := []string{}
	if len(fn.Settings) > 0 {
		var settings map[string]interface{}
		if err := json.Unmarshal(fn.Settings, &settings); err == nil {
			if cd, ok := settings["custom_domains"].([]interface{}); ok {
				for _, v := range cd {
					if s, ok := v.(string); ok {
						customDomains = append(customDomains, s)
					}
				}
			}
		}
	}

	desc := ""
	if fn.Description.Valid {
		desc = fn.Description.String
	}
	resp := FunctionSettingsResponse{
		ID:                   fn.ID.String(),
		Name:                 fn.Name,
		Author:               fn.Author,
		Description:          desc,
		IsPublic:             fn.Visibility == "public",
		IsPublished:          fn.LatestVersion.Valid,
		AllowAnonymousInvoke: fn.Visibility == "public",
		CorsEnabled:          false,
		CorsOrigins:          []string{},
		Timeout:              30,
		Memory:               128,
		Runtime:              "python3.11",
		Providers:            []string{},
		EnvironmentVariables: map[string]string{},
		Secrets:              []string{},
		CustomDomains:        customDomains,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandlePatchFunctionSettings handles PATCH /v1/functions/{author}/{name}/settings
func (h *Handler) HandlePatchFunctionSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req FunctionSettingsPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	plan := middleware.GetTenantPlan(r)
	if plan == "" && h.backendRepo != nil {
		if sub, err := h.backendRepo.GetSubscriptionByTenantID(user.TenantID); err == nil && sub != nil && sub.PricingTier != nil {
			plan = sub.PricingTier.Name
		}
	}
	if plan == "" {
		plan = plans.PlanStarter
	}

	checker := plans.NewFeatureChecker(plan)
	if !checker.HasFeature(plans.FeatureCustomDomains) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Custom domains are not available on your plan. Upgrade to Starter or higher.",
		})
		return
	}

	if req.CustomDomains != nil {
		maxDomains := plans.GetMaxCustomDomains(plan)
		if maxDomains >= 0 && len(*req.CustomDomains) > maxDomains {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Custom domain limit exceeded for your plan",
				"limit":   maxDomains,
				"current": len(*req.CustomDomains),
			})
			return
		}

		// Merge with existing settings so we don't wipe other keys when column exists
		settings := make(map[string]interface{})
		if len(fn.Settings) > 0 {
			_ = json.Unmarshal(fn.Settings, &settings)
		}
		settings["custom_domains"] = *req.CustomDomains

		if err := h.repo.UpdateFunctionSettings(fn.ID, settings); err != nil {
			http.Error(w, "Failed to update settings", http.StatusInternalServerError)
			return
		}
	}

	// Return updated settings (same as GET)
	h.HandleGetFunctionSettings(w, r)
}
