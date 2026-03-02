package wellknown

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// mockRegistry implements registryReader for tests
type mockRegistry struct {
	searchFunc    func(query, category, runtime string, minRating float64, limit, offset int) ([]registry.RegistryFunction, int, error)
	latestVersion func(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error)
}

func (m *mockRegistry) SearchFunctions(query, category, runtime string, minRating float64, limit, offset int) ([]registry.RegistryFunction, int, error) {
	if m.searchFunc != nil {
		return m.searchFunc(query, category, runtime, minRating, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockRegistry) GetLatestFunctionVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error) {
	if m.latestVersion != nil {
		return m.latestVersion(functionID)
	}
	return nil, nil
}

func TestHandleWellKnown_EmptyRegistry(t *testing.T) {
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return nil, 0, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" && !strings.Contains(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=300" {
		t.Errorf("expected Cache-Control public, max-age=300, got %s", rec.Header().Get("Cache-Control"))
	}

	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if manifest.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version 1.0, got %s", manifest.SchemaVersion)
	}
	if manifest.Provider != "functionfly" {
		t.Errorf("expected provider functionfly, got %s", manifest.Provider)
	}
	if manifest.TotalFunctions != 0 {
		t.Errorf("expected total_functions 0, got %d", manifest.TotalFunctions)
	}
	if len(manifest.Functions) != 0 {
		t.Errorf("expected 0 functions, got %d", len(manifest.Functions))
	}
}

func TestHandleWellKnown_WithFunctions(t *testing.T) {
	fnID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{
						ID:               fnID,
						Author:           "acme",
						Name:             "slugify",
						Title:            sql.NullString{String: "Slugify", Valid: true},
						Description:      sql.NullString{String: "Converts text to slug", Valid: true},
						Category:         sql.NullString{String: "text", Valid: true},
						LatestVersion:    sql.NullString{String: "1.0.0", Valid: true},
						PricePerCall:     0,
						ReliabilityScore: 0.98,
						Tags:             json.RawMessage(`["text","utility"]`),
					},
				}, 1, nil
			},
			latestVersion: func(id uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				if id != fnID {
					return nil, nil
				}
				return &registry.RegistryFunctionVersion{
					Version:       "1.0.0",
					Deterministic: true,
					SideEffects:   "none",
					Manifest:      json.RawMessage(`{"input":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}}`),
				}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	req.Host = "api.example.com"
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if manifest.TotalFunctions != 1 {
		t.Errorf("expected total_functions 1, got %d", manifest.TotalFunctions)
	}
	if len(manifest.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(manifest.Functions))
	}
	fn := manifest.Functions[0]
	if fn.URI != "fx://acme/slugify" {
		t.Errorf("expected uri fx://acme/slugify, got %s", fn.URI)
	}
	if fn.Name != "acme_slugify" {
		t.Errorf("expected name acme_slugify, got %s", fn.Name)
	}
	if fn.Title != "Slugify" {
		t.Errorf("expected title Slugify, got %s", fn.Title)
	}
	if fn.Description != "Converts text to slug" {
		t.Errorf("expected description, got %s", fn.Description)
	}
	if fn.Deterministic != true {
		t.Errorf("expected deterministic true, got %v", fn.Deterministic)
	}
	if fn.SideEffects != "none" {
		t.Errorf("expected side_effects none, got %s", fn.SideEffects)
	}
	if len(fn.ToolSchema) == 0 {
		t.Error("expected non-empty tool_schema")
	}
	var toolSchema struct {
		Type     string `json:"type"`
		Function struct {
			Name        string      `json:"name"`
			Description string      `json:"description"`
			Parameters  interface{} `json:"parameters"`
		} `json:"function"`
	}
	if err := json.Unmarshal(fn.ToolSchema, &toolSchema); err != nil {
		t.Fatalf("tool_schema not valid JSON: %v", err)
	}
	if toolSchema.Type != "function" {
		t.Errorf("expected tool_schema.type function, got %s", toolSchema.Type)
	}
	if toolSchema.Function.Name != "acme_slugify" {
		t.Errorf("expected tool_schema.function.name acme_slugify, got %s", toolSchema.Function.Name)
	}
}

func TestHandleWellKnown_QueryParams(t *testing.T) {
	var gotLimit, gotOffset int
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, category, _ string, _ float64, limit, offset int) ([]registry.RegistryFunction, int, error) {
				gotLimit = limit
				gotOffset = offset
				return nil, 0, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json?limit=10&offset=5&category=text", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotLimit != 10 {
		t.Errorf("expected limit=10 to be passed to SearchFunctions, got %d", gotLimit)
	}
	if gotOffset != 5 {
		t.Errorf("expected offset=5 to be passed to SearchFunctions, got %d", gotOffset)
	}
}

func TestHandleWellKnown_OPTIONS(t *testing.T) {
	h := &Handler{registryRepo: &mockRegistry{}}
	req := httptest.NewRequest(http.MethodOptions, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestFnHasTag(t *testing.T) {
	tests := []struct {
		name     string
		tagsJSON json.RawMessage
		param    string
		want     bool
	}{
		{"nil tags", nil, "a", false},
		{"empty array", json.RawMessage(`[]`), "a", false},
		{"single match", json.RawMessage(`["text"]`), "text", true},
		{"comma list", json.RawMessage(`["a","b"]`), "b", true},
		{"case insensitive", json.RawMessage(`["Text"]`), "text", true},
		{"no match", json.RawMessage(`["x"]`), "y", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fnHasTag(tt.tagsJSON, tt.param)
			if got != tt.want {
				t.Errorf("fnHasTag(%s, %q) = %v, want %v", tt.tagsJSON, tt.param, got, tt.want)
			}
		})
	}
}

func TestBuildOpenAIToolSchema(t *testing.T) {
	// No manifest -> empty object params
	out := buildOpenAIToolSchema("my_fn", "desc", nil)
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["type"] != "function" {
		t.Errorf("expected type function, got %v", m["type"])
	}
	fn, _ := m["function"].(map[string]interface{})
	if fn["name"] != "my_fn" || fn["description"] != "desc" {
		t.Errorf("unexpected name/description: %v", fn)
	}
}

func TestHandleWellKnown_MethodNotAllowed(t *testing.T) {
	h := &Handler{registryRepo: &mockRegistry{}}
	req := httptest.NewRequest(http.MethodPost, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, OPTIONS" {
		t.Errorf("expected Allow: GET, OPTIONS, got %s", rec.Header().Get("Allow"))
	}
	var body struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.OK {
		t.Error("expected ok false")
	}
	if body.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected error code METHOD_NOT_ALLOWED, got %s", body.Error.Code)
	}
}

func TestHandleWellKnown_SearchFailure(t *testing.T) {
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return nil, 0, errors.New("search failed")
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	var body struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "SEARCH_FAILED" {
		t.Errorf("expected error code SEARCH_FAILED, got %s", body.Error.Code)
	}
}

func TestHandleWellKnown_AuthorFilter(t *testing.T) {
	fnID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{ID: fnID, Author: "acme", Name: "slugify", LatestVersion: sql.NullString{String: "1.0.0", Valid: true}},
					{ID: uuid.New(), Author: "other", Name: "echo", LatestVersion: sql.NullString{String: "1.0.0", Valid: true}},
				}, 2, nil
			},
			latestVersion: func(id uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				if id != fnID {
					return &registry.RegistryFunctionVersion{Version: "1.0.0", SideEffects: "none"}, nil
				}
				return &registry.RegistryFunctionVersion{Version: "1.0.0", SideEffects: "none"}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json?author=acme", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(manifest.Functions) != 1 {
		t.Errorf("expected 1 function after author filter, got %d", len(manifest.Functions))
	}
	if manifest.TotalFunctions != 1 {
		t.Errorf("expected total_functions 1 when author filter applied, got %d", manifest.TotalFunctions)
	}
	if manifest.Functions[0].URI != "fx://acme/slugify" {
		t.Errorf("expected uri fx://acme/slugify, got %s", manifest.Functions[0].URI)
	}
}

func TestHandleWellKnown_TagsFilter(t *testing.T) {
	fnID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{ID: fnID, Author: "acme", Name: "slugify", Tags: json.RawMessage(`["text","utility"]`), LatestVersion: sql.NullString{Valid: true}},
					{ID: uuid.New(), Author: "acme", Name: "echo", Tags: json.RawMessage(`["other"]`), LatestVersion: sql.NullString{Valid: true}},
				}, 2, nil
			},
			latestVersion: func(id uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				return &registry.RegistryFunctionVersion{SideEffects: "none"}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json?tags=utility", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(manifest.Functions) != 1 {
		t.Errorf("expected 1 function after tags filter, got %d", len(manifest.Functions))
	}
	if manifest.Functions[0].Name != "acme_slugify" {
		t.Errorf("expected name acme_slugify, got %s", manifest.Functions[0].Name)
	}
}

func TestHandleWellKnown_LimitCappedAtMax(t *testing.T) {
	var gotLimit int
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, limit, _ int) ([]registry.RegistryFunction, int, error) {
				gotLimit = limit
				return nil, 0, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json?limit=999", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotLimit != maxLimit {
		t.Errorf("expected limit capped to %d, got %d", maxLimit, gotLimit)
	}
}

func TestHandleWellKnown_InvalidLimitUsesDefault(t *testing.T) {
	var gotLimit int
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, limit, _ int) ([]registry.RegistryFunction, int, error) {
				gotLimit = limit
				return nil, 0, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json?limit=invalid", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotLimit != defaultLimit {
		t.Errorf("expected default limit %d, got %d", defaultLimit, gotLimit)
	}
}

func TestHandleWellKnown_CORSHeadersOnGET(t *testing.T) {
	h := &Handler{registryRepo: &mockRegistry{}}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Access-Control-Allow-Methods") != "GET, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, OPTIONS, got %s", rec.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestHandleWellKnown_ManifestTopLevelFields(t *testing.T) {
	h := &Handler{registryRepo: &mockRegistry{}}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if manifest.APIBase == "" {
		t.Error("expected non-empty api_base")
	}
	if manifest.ExecutionEndpoint != "POST /v1/fx/{author}/{name}" {
		t.Errorf("unexpected execution_endpoint: %s", manifest.ExecutionEndpoint)
	}
	if manifest.AgentEndpoint != "POST /v1/agent/execute/{author}/{name}" {
		t.Errorf("unexpected agent_endpoint: %s", manifest.AgentEndpoint)
	}
	if manifest.DiscoveryEndpoint != "GET /v1/agent/discover" {
		t.Errorf("unexpected discovery_endpoint: %s", manifest.DiscoveryEndpoint)
	}
	if manifest.GeneratedAt.IsZero() {
		t.Error("expected non-zero generated_at")
	}
}

func TestHandleWellKnown_SkippedFunctionNoLatestVersion(t *testing.T) {
	goodID := uuid.New()
	badID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{ID: badID, Author: "bad", Name: "nover", LatestVersion: sql.NullString{}},
					{ID: goodID, Author: "good", Name: "ok", LatestVersion: sql.NullString{String: "1.0.0", Valid: true}},
				}, 2, nil
			},
			latestVersion: func(id uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				if id == badID {
					return nil, errors.New("no version")
				}
				return &registry.RegistryFunctionVersion{Version: "1.0.0", SideEffects: "none"}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(manifest.Functions) != 1 {
		t.Errorf("expected 1 function (one skipped), got %d", len(manifest.Functions))
	}
	if manifest.Functions[0].Name != "good_ok" {
		t.Errorf("expected remaining function good_ok, got %s", manifest.Functions[0].Name)
	}
}

func TestHandleWellKnown_ToolNameSanitized(t *testing.T) {
	fnID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{
						ID:            fnID,
						Author:        "acme-corp",
						Name:          "my-fn.v2",
						LatestVersion: sql.NullString{String: "1.0.0", Valid: true},
					},
				}, 1, nil
			},
			latestVersion: func(id uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				if id != fnID {
					return nil, nil
				}
				return &registry.RegistryFunctionVersion{Version: "1.0.0", SideEffects: "none"}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(manifest.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(manifest.Functions))
	}
	// Sanitized: only [a-zA-Z0-9_]; dashes/dots become underscores
	name := manifest.Functions[0].Name
	for _, r := range name {
		if r != '_' && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			t.Errorf("tool name %q contains invalid rune %q", name, r)
		}
	}
	if !strings.Contains(name, "acme") || !strings.Contains(name, "corp") {
		t.Errorf("expected sanitized name to contain acme and corp, got %s", name)
	}
}

func TestHandleWellKnown_ExecutionURLsUseRequestHost(t *testing.T) {
	fnID := uuid.New()
	h := &Handler{
		registryRepo: &mockRegistry{
			searchFunc: func(_, _, _ string, _ float64, _, _ int) ([]registry.RegistryFunction, int, error) {
				return []registry.RegistryFunction{
					{ID: fnID, Author: "acme", Name: "fn", LatestVersion: sql.NullString{Valid: true}},
				}, 1, nil
			},
			latestVersion: func(uuid.UUID) (*registry.RegistryFunctionVersion, error) {
				return &registry.RegistryFunctionVersion{SideEffects: "none"}, nil
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/functionfly.json", nil)
	req.Host = "custom.api.example.com"
	rec := httptest.NewRecorder()
	h.HandleWellKnown(rec, req)

	var manifest FunctionFlyManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	fn := manifest.Functions[0]
	if !strings.Contains(fn.ExecutionURL, "custom.api.example.com") {
		t.Errorf("expected execution_url to contain request host, got %s", fn.ExecutionURL)
	}
	if !strings.Contains(fn.AgentExecutionURL, "custom.api.example.com") {
		t.Errorf("expected agent_execution_url to contain request host, got %s", fn.AgentExecutionURL)
	}
}

func TestBuildOpenAIToolSchema_WithManifestInputSchema(t *testing.T) {
	manifest := json.RawMessage(`{"input":{"schema":{"type":"object","properties":{"x":{"type":"string"}},"required":["x"]}}`)
	out := buildOpenAIToolSchema("test_fn", "test desc", manifest)
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	fn, _ := m["function"].(map[string]interface{})
	params, _ := fn["parameters"].(map[string]interface{})
	if params["type"] != "object" {
		t.Errorf("expected parameters.type object, got %v", params["type"])
	}
	props, _ := params["properties"].(map[string]interface{})
	if props == nil {
		t.Error("expected parameters.properties")
	}
}

func TestFnHasTag_EmptyRequestedTagIgnored(t *testing.T) {
	// Comma with empty segment should not match everything
	got := fnHasTag(json.RawMessage(`["a"]`), "a,,b")
	if !got {
		t.Error("expected true when one of requested tags matches")
	}
}
