package registry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// DeprecationInfo represents deprecation information for an endpoint
type DeprecationInfo struct {
	Endpoint        string     `json:"endpoint"`
	Method          string     `json:"method"`
	DeprecatedIn    string     `json:"deprecated_in"`
	RemovedIn       string     `json:"removed_in"`
	DeprecationDate *time.Time `json:"deprecation_date,omitempty"`
	SunsetDate      *time.Time `json:"sunset_date,omitempty"`
	MigrationGuide  string     `json:"migration_guide"`
	Alternative     string     `json:"alternative"`
	BreakingChanges []string   `json:"breaking_changes"`
	NewFields       []string   `json:"new_fields"`
	TypeChanges     []string   `json:"type_changes"`
}

// ptrTime returns a pointer to a time.Time
func ptrTime(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}

// DeprecationHandler handles deprecation-related requests
type DeprecationHandler struct {
	versionManager *middleware.VersionManager
}

// NewDeprecationHandler creates a new deprecation handler
func NewDeprecationHandler(versionManager *middleware.VersionManager) *DeprecationHandler {
	return &DeprecationHandler{
		versionManager: versionManager,
	}
}

// HandleGetAllDeprecations returns all deprecations across all endpoints
func (h *DeprecationHandler) HandleGetAllDeprecations(w http.ResponseWriter, r *http.Request) {
	deprecations := []DeprecationInfo{
		{
			Endpoint:       "/registry/functions",
			Method:         "GET",
			DeprecatedIn:   "v1",
			RemovedIn:      "v3",
			MigrationGuide: "Use /v2/registry/functions for new integrations. The v2 endpoint returns camelCase field names and includes additional metadata fields.",
			Alternative:    "/v2/registry/functions",
			BreakingChanges: []string{
				"Field 'popularity_score' renamed to 'popularityScore' (type changed to float)",
				"Field 'deterministic_score' renamed to 'deterministicScore'",
			},
			NewFields: []string{
				"trust_score",
				"execution_count",
				"last_executed_at",
			},
			TypeChanges: []string{
				"popularity_score: int -> popularityScore: float",
			},
		},
		{
			Endpoint:       "/registry/functions/{author}/{name}",
			Method:         "GET",
			DeprecatedIn:   "v1",
			RemovedIn:      "v3",
			MigrationGuide: "Use /v2/registry/functions/{author}/{name} for new integrations. Includes trust_score and verified fields.",
			Alternative:    "/v2/registry/functions/{author}/{name}",
			BreakingChanges: []string{
				"Field 'latest_version' renamed to 'latestVersion'",
			},
			NewFields: []string{
				"trust_score",
				"verified",
				"signature_info",
			},
		},
		{
			Endpoint:       "/registry/search",
			Method:         "GET",
			DeprecatedIn:   "v1",
			RemovedIn:      "v3",
			MigrationGuide: "Use /v2/registry/search for new integrations. Includes relevance scoring and highlighting.",
			Alternative:    "/v2/registry/search",
			NewFields: []string{
				"relevance_score",
				"highlights",
			},
		},
	}

	response := map[string]interface{}{
		"current_version":   "v1",
		"successor_version": "v2",
		"deprecation_date":  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		"sunset_date":       time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		"deprecations":      deprecations,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Sunset", "Sat, 01 Jan 2027 00:00:00 GMT")
	w.Header().Set("Link", "</v2/registry/deprecations>; rel=\"successor-version\"")

	json.NewEncoder(w).Encode(response)
}

// HandleGetEndpointDeprecation returns deprecation info for a specific endpoint
func (h *DeprecationHandler) HandleGetEndpointDeprecation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpoint := vars["endpoint"]
	method := r.URL.Query().Get("method")

	if method == "" {
		method = "GET"
	}

	// Find deprecation info for this endpoint
	dep := h.findDeprecation(endpoint, method)
	if dep == nil {
		http.Error(w, "No deprecation information found for this endpoint", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Deprecation", "true")

	if dep.SunsetDate != nil {
		w.Header().Set("Sunset", dep.SunsetDate.Format(http.TimeFormat))
	}

	json.NewEncoder(w).Encode(dep)
}

// HandleGetVersionDeprecation returns deprecation info for the current API version
func (h *DeprecationHandler) HandleGetVersionDeprecation(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "v1"
	}

	info := h.versionManager.GetVersionCompatibilityInfo(version)

	w.Header().Set("Content-Type", "application/json")

	if h.versionManager.IsDeprecated(version) {
		w.Header().Set("Deprecation", "true")
		if sunsetDate := h.versionManager.GetSunsetDate(version); sunsetDate != nil {
			w.Header().Set("Sunset", sunsetDate.Format(http.TimeFormat))
		}
	}

	json.NewEncoder(w).Encode(info)
}

// findDeprecation finds deprecation info for an endpoint
func (h *DeprecationHandler) findDeprecation(endpoint, method string) *DeprecationInfo {
	deprecations := map[string]*DeprecationInfo{
		"GET:/v1/registry/functions": {
			Endpoint:        "/registry/functions",
			Method:          "GET",
			DeprecatedIn:    "v1",
			RemovedIn:       "v3",
			DeprecationDate: ptrTime(2025, 10, 1),
			SunsetDate:      ptrTime(2027, 1, 1),
			MigrationGuide:  "Use /v2/registry/functions for new integrations.",
			Alternative:     "/v2/registry/functions",
			BreakingChanges: []string{
				"popularity_score -> popularityScore (type changed to float)",
			},
			NewFields: []string{
				"trust_score",
				"execution_count",
				"last_executed_at",
			},
		},
		"GET:/v1/registry/functions/{author}/{name}": {
			Endpoint:       "/registry/functions/{author}/{name}",
			Method:         "GET",
			DeprecatedIn:   "v1",
			RemovedIn:      "v3",
			MigrationGuide: "Use /v2/registry/functions/{author}/{name}",
			Alternative:    "/v2/registry/functions/{author}/{name}",
			BreakingChanges: []string{
				"latest_version -> latestVersion",
			},
			NewFields: []string{
				"trust_score",
				"verified",
				"signature_info",
			},
		},
		"GET:/v1/registry/search": {
			Endpoint:       "/registry/search",
			Method:         "GET",
			DeprecatedIn:   "v1",
			RemovedIn:      "v3",
			MigrationGuide: "Use /v2/registry/search",
			Alternative:    "/v2/registry/search",
			NewFields: []string{
				"relevance_score",
				"highlights",
			},
		},
	}

	key := method + ":" + endpoint
	if dep, ok := deprecations[key]; ok {
		return dep
	}

	// Try with version prefix
	key = method + ":/v1" + endpoint
	return deprecations[key]
}
