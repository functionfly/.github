package registry

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req FunctionSettingsPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	plan := middleware.GetTenantPlan(r)
	if plan == "" && h.backendRepo != nil {
		if sub, err := h.backendRepo.GetSubscriptionByTenantID(context.Background(), user.TenantID); err == nil && sub != nil && sub.PricingTier != nil {
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

		if err := h.repo.UpdateFunctionSettings(context.Background(), fn.ID, settings); err != nil {
			apierror.WriteError(w, apierror.NewInternal("Failed to update settings"))
			return
		}
	}

	// Return updated settings (same as GET)
	h.HandleGetFunctionSettings(w, r)
}

type EnvVarsRequest map[string]string

type EnvVarResponse struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type EnvVarsResponse struct {
	EnvironmentVariables map[string]string `json:"environmentVariables"`
}

func (h *Handler) HandleGetEnvVars(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	envVars := map[string]string{}
	if len(fn.Settings) > 0 {
		var settings map[string]interface{}
		if err := json.Unmarshal(fn.Settings, &settings); err == nil {
			if env, ok := settings["environment_variables"].(map[string]interface{}); ok {
				for k, v := range env {
					if s, ok := v.(string); ok {
						envVars[k] = s
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(EnvVarsResponse{EnvironmentVariables: envVars})
}

func (h *Handler) HandlePutEnvVars(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req EnvVarsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	settings := make(map[string]interface{})
	if len(fn.Settings) > 0 {
		_ = json.Unmarshal(fn.Settings, &settings)
	}
	settings["environment_variables"] = req

	if err := h.repo.UpdateFunctionSettings(context.Background(), fn.ID, settings); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update environment variables"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(EnvVarsResponse{EnvironmentVariables: req})
}

func (h *Handler) HandleDeleteEnvVar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	key := vars["key"]
	if author == "" || name == "" || key == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author, name, and key are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	settings := make(map[string]interface{})
	if len(fn.Settings) > 0 {
		_ = json.Unmarshal(fn.Settings, &settings)
	}

	if env, ok := settings["environment_variables"].(map[string]interface{}); ok {
		delete(env, key)
		settings["environment_variables"] = env
	}

	if err := h.repo.UpdateFunctionSettings(context.Background(), fn.ID, settings); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to delete environment variable"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SecretsRequest map[string]string

type SecretsResponse struct {
	Secrets []string `json:"secrets"`
}

func (h *Handler) HandleGetSecrets(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	secrets := []string{}
	if len(fn.Settings) > 0 {
		var settings map[string]interface{}
		if err := json.Unmarshal(fn.Settings, &settings); err == nil {
			if sec, ok := settings["secrets"].([]interface{}); ok {
				for _, v := range sec {
					if s, ok := v.(string); ok {
						secrets = append(secrets, s)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SecretsResponse{Secrets: secrets})
}

func (h *Handler) HandlePutSecrets(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req SecretsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	secretKeys := make([]string, 0, len(req))
	for k := range req {
		secretKeys = append(secretKeys, k)
	}

	settings := make(map[string]interface{})
	if len(fn.Settings) > 0 {
		_ = json.Unmarshal(fn.Settings, &settings)
	}
	settings["secrets"] = secretKeys

	if err := h.repo.UpdateFunctionSettings(context.Background(), fn.ID, settings); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update secrets"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SecretsResponse{Secrets: secretKeys})
}

func (h *Handler) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	key := vars["key"]
	if author == "" || name == "" || key == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author, name, and key are required"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	settings := make(map[string]interface{})
	if len(fn.Settings) > 0 {
		_ = json.Unmarshal(fn.Settings, &settings)
	}

	if sec, ok := settings["secrets"].([]interface{}); ok {
		newSecrets := make([]string, 0, len(sec))
		for _, v := range sec {
			if s, ok := v.(string); ok && s != key {
				newSecrets = append(newSecrets, s)
			}
		}
		settings["secrets"] = newSecrets
	}

	if err := h.repo.UpdateFunctionSettings(context.Background(), fn.ID, settings); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to delete secret"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
