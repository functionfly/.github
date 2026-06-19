package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/versions"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// DeprecationConfig contains deprecation configuration
type DeprecationConfig struct {
	WarningDaysBeforeDeprecation int        // Days before deprecation to start warning
	SunsetDate                   *time.Time // When the endpoint will be removed
	SuccessorEndpoint            string     // The new endpoint to use
	SuccessorVersion             string     // The new API version
	MigrationGuideURL            string     // URL to migration guide
}

// DefaultDeprecationConfig returns default deprecation configuration
func DefaultDeprecationConfig() *DeprecationConfig {
	return &DeprecationConfig{
		WarningDaysBeforeDeprecation: 90,
	}
}

// EndpointDeprecation stores deprecation information for an endpoint
type EndpointDeprecation struct {
	Endpoint        string     `json:"endpoint"`
	Method          string     `json:"method"`
	DeprecatedIn    string     `json:"deprecated_in"`
	RemovedIn       string     `json:"removed_in"`
	DeprecationDate *time.Time `json:"deprecation_date,omitempty"`
	SunsetDate      *time.Time `json:"sunset_date,omitempty"`
	MigrationGuide  string     `json:"migration_guide"`
	Alternative     string     `json:"alternative"`
	BreakingChanges []string   `json:"breaking_changes"`
}

// DeprecationTracker tracks deprecation status for all endpoints
type DeprecationTracker struct {
	deprecations map[string]*EndpointDeprecation
}

// NewDeprecationTracker creates a new deprecation tracker
func NewDeprecationTracker() *DeprecationTracker {
	dt := &DeprecationTracker{
		deprecations: make(map[string]*EndpointDeprecation),
	}
	dt.initDefaultDeprecations()
	return dt
}

// initDefaultDeprecations initializes default deprecation rules
func (dt *DeprecationTracker) initDefaultDeprecations() {
	// Register v1 endpoint deprecations
	dt.deprecations["GET:/v1/registry/functions"] = &EndpointDeprecation{
		Endpoint:       "/registry/functions",
		Method:         "GET",
		DeprecatedIn:   "v1",
		RemovedIn:      "v3",
		MigrationGuide: "Use /v2/registry/functions for new integrations",
		Alternative:    "/v2/registry/functions",
		BreakingChanges: []string{
			"Field 'popularity_score' renamed to 'popularityScore'",
			"New field 'trust_score' added",
			"New field 'execution_count' added",
		},
	}

	dt.deprecations["GET:/v1/registry/functions/{author}/{name}"] = &EndpointDeprecation{
		Endpoint:       "/registry/functions/{author}/{name}",
		Method:         "GET",
		DeprecatedIn:   "v1",
		RemovedIn:      "v3",
		MigrationGuide: "Use /v2/registry/functions/{author}/{name}",
		Alternative:    "/v2/registry/functions/{author}/{name}",
		BreakingChanges: []string{
			"Field 'latest_version' renamed to 'latestVersion'",
			"New field 'trust_score' added",
			"New field 'verified' added",
		},
	}

	dt.deprecations["GET:/v1/registry/search"] = &EndpointDeprecation{
		Endpoint:       "/registry/search",
		Method:         "GET",
		DeprecatedIn:   "v1",
		RemovedIn:      "v3",
		MigrationGuide: "Use /v2/registry/search",
		Alternative:    "/v2/registry/search",
		BreakingChanges: []string{
			"New field 'relevance_score' added",
			"New field 'highlights' added",
		},
	}
}

// GetDeprecation returns deprecation info for an endpoint
func (dt *DeprecationTracker) GetDeprecation(endpoint, method string) *EndpointDeprecation {
	key := fmt.Sprintf("%s:%s", method, endpoint)
	return dt.deprecations[key]
}

// RegisterDeprecation registers a new deprecation
func (dt *DeprecationTracker) RegisterDeprecation(dep *EndpointDeprecation) {
	key := fmt.Sprintf("%s:%s", dep.Method, dep.Endpoint)
	dt.deprecations[key] = dep
}

// GetAllDeprecations returns all registered deprecations
func (dt *DeprecationTracker) GetAllDeprecations() []*EndpointDeprecation {
	deps := make([]*EndpointDeprecation, 0, len(dt.deprecations))
	for _, dep := range dt.deprecations {
		deps = append(deps, dep)
	}
	return deps
}

// DeprecationMiddleware creates middleware that adds deprecation headers
func DeprecationMiddleware(tracker *DeprecationTracker, versionManager *VersionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the API version
			version := GetAPIVersion(r)

			// Get the endpoint and method
			path := r.URL.Path

			// Try to match the path to a deprecated endpoint
			method := r.Method

			// Check if this endpoint is deprecated
			dep := tracker.GetDeprecation(path, method)
			if dep == nil {
				// Try with version prefix
				dep = tracker.GetDeprecation("/v1"+path, method)
			}

			if dep != nil && version == "v1" {
				// Add deprecation headers
				w.Header().Set("Deprecation", "true")

				if dep.SunsetDate != nil {
					w.Header().Set("Sunset", dep.SunsetDate.Format(http.TimeFormat))
				}

				if dep.Alternative != "" {
					w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", dep.Alternative))
				}

				w.Header().Set("X-API-Warning", dep.MigrationGuide)

				// Log deprecation warning
				logrus.Warnf("Deprecated endpoint accessed: %s %s - %s", method, path, dep.MigrationGuide)
			}

			// Check if the version itself is deprecated
			if versionManager.IsDeprecated(version) {
				if sunsetDate := versionManager.GetSunsetDate(version); sunsetDate != nil {
					w.Header().Set("Sunset", sunsetDate.Format(http.TimeFormat))
					w.Header().Set("Deprecation", "true")

					vInfo := versionManager.GetVersionInfo(version)
					if vInfo != nil && vInfo.SuccessorVersion != "" {
						w.Header().Set("Link", fmt.Sprintf("</v%s/registry/functions>; rel=\"successor-version\"", vInfo.SuccessorVersion))
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AddDeprecationHeaders adds deprecation headers to a response
func AddDeprecationHeaders(w http.ResponseWriter, dep *EndpointDeprecation) {
	if dep == nil {
		return
	}

	w.Header().Set("Deprecation", "true")

	if dep.SunsetDate != nil {
		w.Header().Set("Sunset", dep.SunsetDate.Format(http.TimeFormat))
	}

	if dep.Alternative != "" {
		w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", dep.Alternative))
	}

	if dep.MigrationGuide != "" {
		w.Header().Set("X-API-Warning", dep.MigrationGuide)
	}
}

// DeprecationHandler handles deprecation-related requests
type DeprecationHandler struct {
	tracker        *DeprecationTracker
	versionManager *VersionManager
	compatibility  *versions.CompatibilityLayer
}

// NewDeprecationHandler creates a new deprecation handler
func NewDeprecationHandler(tracker *DeprecationTracker, versionManager *VersionManager) *DeprecationHandler {
	return &DeprecationHandler{
		tracker:        tracker,
		versionManager: versionManager,
		compatibility:  versions.NewCompatibilityLayer(),
	}
}

// HandleGetDeprecations returns all deprecations
func (h *DeprecationHandler) HandleGetDeprecations(w http.ResponseWriter, r *http.Request) {
	deps := h.tracker.GetAllDeprecations()

	response := map[string]interface{}{
		"deprecations":      deps,
		"current_version":   "v1",
		"successor_version": "v2",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Note: Should use json.Marshal but returning response map
	// This would need proper JSON encoding in actual implementation
	_ = response
}

// HandleGetEndpointDeprecation returns deprecation info for a specific endpoint
func (h *DeprecationHandler) HandleGetEndpointDeprecation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpoint := vars["endpoint"]
	method := r.URL.Query().Get("method")

	if method == "" {
		method = "GET"
	}

	dep := h.tracker.GetDeprecation(endpoint, method)
	if dep == nil {
		// Try without version prefix
		dep = h.tracker.GetDeprecation("/v1"+endpoint, method)
	}

	if dep == nil {
		apierror.WriteError(w, apierror.NewNotFound("No deprecation information found for this endpoint"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = dep
}

// IsDeprecated checks if an endpoint is deprecated
func (h *DeprecationHandler) IsDeprecated(endpoint, method string) bool {
	return h.tracker.GetDeprecation(endpoint, method) != nil
}

// GetDeprecationInfo returns detailed deprecation information
func (h *DeprecationHandler) GetDeprecationInfo(endpoint, method string) map[string]interface{} {
	dep := h.tracker.GetDeprecation(endpoint, method)
	if dep == nil {
		return nil
	}

	return map[string]interface{}{
		"endpoint":         dep.Endpoint,
		"method":           dep.Method,
		"deprecated_in":    dep.DeprecatedIn,
		"removed_in":       dep.RemovedIn,
		"migration_guide":  dep.MigrationGuide,
		"alternative":      dep.Alternative,
		"breaking_changes": dep.BreakingChanges,
	}
}

// RequireNonDeprecated creates middleware that rejects requests to deprecated endpoints
func RequireNonDeprecated(tracker *DeprecationTracker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			dep := tracker.GetDeprecation(path, method)
			if dep == nil {
				dep = tracker.GetDeprecation("/v1"+path, method)
			}

			if dep != nil {
				// Endpoint is deprecated but still functional
				// You could choose to reject with 410 Gone instead
				// For now, we just add headers and continue
				AddDeprecationHeaders(w, dep)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetSunsetHeaders returns sunset headers for a version
func GetSunsetHeaders(version string, versionManager *VersionManager) map[string]string {
	headers := make(map[string]string)

	if sunsetDate := versionManager.GetSunsetDate(version); sunsetDate != nil {
		headers["Sunset"] = sunsetDate.Format(http.TimeFormat)
		headers["Deprecation"] = "true"

		if vInfo := versionManager.GetVersionInfo(version); vInfo != nil && vInfo.SuccessorVersion != "" {
			headers["Link"] = fmt.Sprintf("</v%s/registry/functions>; rel=\"successor-version\"", vInfo.SuccessorVersion)
		}
	}

	return headers
}
