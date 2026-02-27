package registry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// MigrationChange represents a change in the API
type MigrationChange struct {
	Field     string `json:"field"`
	OldType   string `json:"old_type"`
	NewType   string `json:"new_type"`
	Migration string `json:"migration"`
}

// MigrationEndpoint represents migration info for an endpoint
type MigrationEndpoint struct {
	Endpoint        string            `json:"endpoint"`
	Method          string            `json:"method"`
	BreakingChanges []MigrationChange `json:"breaking_changes"`
	Additions       []string          `json:"additions"`
	Removals        []string          `json:"removals"`
}

// MigrationGuide represents a complete migration guide
type MigrationGuide struct {
	CurrentVersion   string              `json:"current_version"`
	SuccessorVersion string              `json:"successor_version"`
	DeprecationDate  time.Time           `json:"deprecation_date"`
	SunsetDate       time.Time           `json:"sunset_date"`
	Changes          []MigrationEndpoint `json:"changes"`
	QuickStart       string              `json:"quick_start"`
	FAQs             []FAQ               `json:"faqs"`
}

// FAQ represents a frequently asked question
type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// MigrationHandler handles migration guide requests
type MigrationHandler struct{}

// NewMigrationHandler creates a new migration handler
func NewMigrationHandler() *MigrationHandler {
	return &MigrationHandler{}
}

// HandleGetMigrationGuide returns the complete migration guide
func (h *MigrationHandler) HandleGetMigrationGuide(w http.ResponseWriter, r *http.Request) {
	guide := MigrationGuide{
		CurrentVersion:   "v1",
		SuccessorVersion: "v2",
		DeprecationDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		SunsetDate:       time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		QuickStart:       "Replace /v1/ with /v2/ in your API calls. Update field names from snake_case to camelCase.",
		Changes: []MigrationEndpoint{
			{
				Endpoint: "/registry/functions",
				Method:   "GET",
				BreakingChanges: []MigrationChange{
					{
						Field:     "popularity_score",
						OldType:   "int",
						NewType:   "float",
						Migration: "Cast to float client-side: floatValue = parseFloat(popularityScore)",
					},
					{
						Field:     "deterministic_score",
						OldType:   "snake_case",
						NewType:   "camelCase",
						Migration: "Use deterministicScore instead of deterministic_score",
					},
				},
				Additions: []string{
					"trust_score (float)",
					"execution_count (int)",
					"last_executed_at (timestamp)",
				},
				Removals: []string{},
			},
			{
				Endpoint: "/registry/functions/{author}/{name}",
				Method:   "GET",
				BreakingChanges: []MigrationChange{
					{
						Field:     "latest_version",
						OldType:   "snake_case",
						NewType:   "camelCase",
						Migration: "Use latestVersion instead of latest_version",
					},
				},
				Additions: []string{
					"trust_score (float)",
					"verified (boolean)",
					"signature_info (object)",
				},
				Removals: []string{},
			},
			{
				Endpoint:        "/registry/search",
				Method:          "GET",
				BreakingChanges: []MigrationChange{},
				Additions: []string{
					"relevance_score (float)",
					"highlights (array)",
				},
				Removals: []string{},
			},
		},
		FAQs: []FAQ{
			{
				Question: "When will v1 be removed?",
				Answer:   "v1 is scheduled for removal on January 1, 2027 (Sunset Date).",
			},
			{
				Question: "Do I need to update my code?",
				Answer:   "Yes, you need to update field names from snake_case to camelCase and handle the new float type for popularityScore.",
			},
			{
				Question: "Will v1 stop working immediately?",
				Answer:   "No, v1 will continue to work but will return deprecation warnings. Plan your migration to v2 before the sunset date.",
			},
			{
				Question: "How do I handle the popularity score type change?",
				Answer:   "Update your code to handle float values instead of integers. Example: const score = parseFloat(data.popularityScore)",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Sunset", "Sat, 01 Jan 2027 00:00:00 GMT")

	json.NewEncoder(w).Encode(guide)
}

// HandleGetEndpointMigration returns migration info for a specific endpoint
func (h *MigrationHandler) HandleGetEndpointMigration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpoint := vars["endpoint"]
	method := r.URL.Query().Get("method")

	if method == "" {
		method = "GET"
	}

	// Find migration info for this endpoint
	migration := h.findMigration(endpoint, method)
	if migration == nil {
		http.Error(w, "No migration information found for this endpoint", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(migration)
}

// HandleGetVersionInfo returns information about available API versions
func (h *MigrationHandler) HandleGetVersionInfo(w http.ResponseWriter, r *http.Request) {
	versions := []map[string]interface{}{
		{
			"version":          "v1",
			"status":           "deprecated",
			"deprecation_date": "2026-01-01",
			"sunset_date":      "2027-01-01",
			"successor":        "v2",
		},
		{
			"version":          "v2",
			"status":           "current",
			"deprecation_date": nil,
			"sunset_date":      nil,
			"successor":        nil,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": versions,
		"default":  "v2",
	})
}

// findMigration finds migration info for an endpoint
func (h *MigrationHandler) findMigration(endpoint, method string) *MigrationEndpoint {
	migrations := map[string]*MigrationEndpoint{
		"GET:/v1/registry/functions": {
			Endpoint: "/registry/functions",
			Method:   "GET",
			BreakingChanges: []MigrationChange{
				{
					Field:     "popularity_score",
					OldType:   "int",
					NewType:   "float",
					Migration: "Cast to float client-side",
				},
			},
			Additions: []string{"trust_score", "execution_count", "last_executed_at"},
			Removals:  []string{},
		},
		"GET:/v1/registry/functions/{author}/{name}": {
			Endpoint: "/registry/functions/{author}/{name}",
			Method:   "GET",
			BreakingChanges: []MigrationChange{
				{
					Field:     "latest_version",
					OldType:   "snake_case",
					NewType:   "camelCase",
					Migration: "Use latestVersion",
				},
			},
			Additions: []string{"trust_score", "verified", "signature_info"},
			Removals:  []string{},
		},
		"GET:/v1/registry/search": {
			Endpoint:        "/registry/search",
			Method:          "GET",
			BreakingChanges: []MigrationChange{},
			Additions:       []string{"relevance_score", "highlights"},
			Removals:        []string{},
		},
	}

	key := method + ":" + endpoint
	if m, ok := migrations[key]; ok {
		return m
	}

	// Try with version prefix
	key = method + ":/v1" + endpoint
	return migrations[key]
}
