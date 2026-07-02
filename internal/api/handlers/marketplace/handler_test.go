package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func newAuthedRequest(method, url string, body *bytes.Buffer) *http.Request {
	return newAuthedRequestWithTenant(method, url, body, uuid.New().String())
}

func newAuthedRequestWithTenant(method, url string, body *bytes.Buffer, tenantID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	tid, _ := uuid.Parse(tenantID)
	claims := &auth.Claims{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		TenantID: tid,
		Role:     "user",
	}
	return middleware.SetUserInContext(req, claims)
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"1.0", true},
		{"1.2.3", true},
		{"10.20.30", true},
		{"0.0.1", true},
		{"v2.1", true},
		{"1.0.0-beta", false},
		{"abc", false},
		{"1.0.0.0", false},
		{"1", false},
		{".1.0", false},
		{"1..0", false},
		{"a.b.c", false},
		{"1.x.0", false},
		{"", false},
		{".", false},
		{"1.0.", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidSemver(tt.input)
			if got != tt.expected {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMapToStruct(t *testing.T) {
	type Sample struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Count   int    `json:"count"`
	}

	t.Run("valid map", func(t *testing.T) {
		input := map[string]interface{}{
			"name":    "test",
			"version": "1.0.0",
			"count":   42,
		}
		var dst Sample
		if err := mapToStruct(input, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Name != "test" || dst.Version != "1.0.0" || dst.Count != 42 {
			t.Errorf("got %+v", dst)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var dst Sample
		if err := mapToStruct(nil, &dst); err == nil {
			t.Error("expected error for nil map")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		var dst Sample
		if err := mapToStruct(map[string]interface{}{}, &dst); err != nil {
			t.Errorf("empty map should not error: %v", err)
		}
		if dst.Name != "" {
			t.Errorf("expected empty struct, got %+v", dst)
		}
	})
}

func TestValidateExtensionRequest(t *testing.T) {
	tests := []struct {
		name        string
		extName     string
		version     string
		manifest    map[string]interface{}
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid clean plugin",
			extName:     "good-plugin",
			version:     "1.0.0",
			manifest:    map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "missing name",
			extName:     "",
			version:     "1.0.0",
			manifest:    map[string]interface{}{},
			expectError: true,
			errorSubstr: "name is required",
		},
		{
			name:        "missing version",
			extName:     "good",
			version:     "",
			manifest:    map[string]interface{}{},
			expectError: true,
			errorSubstr: "version is required",
		},
		{
			name:        "invalid version",
			extName:     "good",
			version:     "not-semver",
			manifest:    map[string]interface{}{},
			expectError: true,
			errorSubstr: "semver",
		},
		{
			name:        "name too long",
			extName:     strings.Repeat("a", 256),
			version:     "1.0.0",
			manifest:    map[string]interface{}{},
			expectError: true,
			errorSubstr: "255 characters",
		},
		{
			name:        "high-risk permissions trigger security failure",
			extName:     "dangerous",
			version:     "1.0.0",
			manifest:    map[string]interface{}{},
			expectError: true,
			errorSubstr: "security analysis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "high-risk permissions trigger security failure" {
				tt.manifest = map[string]interface{}{
					"permissions": map[string]interface{}{
						"terminal":   true,
						"secrets":    true,
						"api_keys":   true,
						"gpu":        true,
						"network":    true,
						"filesystem": true,
					},
					"network_hosts":    []string{"pastebin.com"},
					"filesystem_scope": "full",
				}
			}

			err := validateExtensionRequest(tt.extName, tt.version, tt.manifest)
			if tt.expectError && err == nil {
				t.Errorf("expected error containing %q, got nil", tt.errorSubstr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.expectError && err != nil && tt.errorSubstr != "" {
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errorSubstr)
				}
			}
		})
	}
}

type mockRepo struct {
	extensions map[string]*Extension
	ratings    map[string]*Rating
	updates    []ExtensionUpdate
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		extensions: make(map[string]*Extension),
		ratings:    make(map[string]*Rating),
	}
}

func (m *mockRepo) List(ctx context.Context, params ListParams) ([]Extension, error) {
	result := []Extension{}
	for _, e := range m.extensions {
		result = append(result, *e)
	}
	return result, nil
}

func (m *mockRepo) Get(ctx context.Context, id string) (*Extension, error) {
	if e, ok := m.extensions[id]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockRepo) Create(ctx context.Context, ext *Extension) error {
	if ext.ID == "" {
		ext.ID = "ext-1"
	}
	m.extensions[ext.ID] = ext
	return nil
}

func (m *mockRepo) Update(ctx context.Context, ext *Extension) error {
	m.extensions[ext.ID] = ext
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.extensions, id)
	return nil
}

func (m *mockRepo) IncrementInstallCount(ctx context.Context, id string) error {
	if e, ok := m.extensions[id]; ok {
		e.InstallCount++
	}
	return nil
}

func (m *mockRepo) GetInstallCounts(ctx context.Context, ids []string) (map[string]int, error) {
	result := make(map[string]int)
	for _, id := range ids {
		if e, ok := m.extensions[id]; ok {
			result[id] = e.InstallCount
		}
	}
	return result, nil
}

func (m *mockRepo) GetCategories(ctx context.Context) ([]CategoryCount, error) {
	return []CategoryCount{}, nil
}

func (m *mockRepo) UpsertRating(ctx context.Context, rating *Rating) error {
	key := rating.ExtensionID + ":" + rating.TenantID
	m.ratings[key] = rating
	return nil
}

func (m *mockRepo) GetRating(ctx context.Context, extensionID, tenantID string) (*Rating, error) {
	key := extensionID + ":" + tenantID
	if r, ok := m.ratings[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockRepo) ListRatings(ctx context.Context, extensionID string, limit int) ([]Rating, error) {
	result := []Rating{}
	for _, r := range m.ratings {
		if r.ExtensionID == extensionID {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *mockRepo) FindUpdates(ctx context.Context, installed []InstalledPlugin) ([]ExtensionUpdate, error) {
	return m.updates, nil
}

func (m *mockRepo) CreatePluginFromExtension(ctx context.Context, tenantID, extensionID string) (*Extension, error) {
	ext, ok := m.extensions[extensionID]
	if !ok {
		return nil, nil
	}
	return ext, nil
}

func TestHandleListExtensions_QueryParamParsing(t *testing.T) {
	repo := newMockRepo()
	repo.extensions["1"] = &Extension{ID: "1", Name: "test", Version: "1.0.0", Tags: []string{"ci"}}
	h := NewHandler(repo)

	t.Run("with tags param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/marketplace/extensions?tags=ci,github", nil)
		w := httptest.NewRecorder()
		h.HandleListExtensions(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("with sort param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/marketplace/extensions?sort=trending", nil)
		w := httptest.NewRecorder()
		h.HandleListExtensions(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("with limit and offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/marketplace/extensions?limit=10&offset=20", nil)
		w := httptest.NewRecorder()
		h.HandleListExtensions(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("with invalid limit uses default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/marketplace/extensions?limit=abc", nil)
		w := httptest.NewRecorder()
		h.HandleListExtensions(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("limit too large capped at 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/marketplace/extensions?limit=999999", nil)
		w := httptest.NewRecorder()
		h.HandleListExtensions(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestHandleCreateExtension_ValidationErrors(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	t.Run("missing tenant id", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"test","version":"1.0.0"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/marketplace/extensions", body)
		w := httptest.NewRecorder()
		h.HandleCreateExtension(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		body := bytes.NewBufferString(`{invalid json`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions", body)
		w := httptest.NewRecorder()
		h.HandleCreateExtension(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"","version":"1.0.0"}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions", body)
		w := httptest.NewRecorder()
		h.HandleCreateExtension(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid semver", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"good","version":"abc"}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions", body)
		w := httptest.NewRecorder()
		h.HandleCreateExtension(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("valid request", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"good-plugin","version":"1.0.0","manifest":{}}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions", body)
		w := httptest.NewRecorder()
		h.HandleCreateExtension(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if _, ok := resp["extension"]; !ok {
			t.Error("response should include extension")
		}
	})
}

func TestHandleRateExtension(t *testing.T) {
	repo := newMockRepo()
	repo.extensions["ext-1"] = &Extension{ID: "ext-1", Name: "test", Version: "1.0.0"}
	h := NewHandler(repo)

	t.Run("missing tenant", func(t *testing.T) {
		body := bytes.NewBufferString(`{"rating":5,"review":"great"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/marketplace/extensions/ext-1/rate", body)
		w := httptest.NewRecorder()
		h.HandleRateExtension(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing extension id", func(t *testing.T) {
		body := bytes.NewBufferString(`{"rating":5,"review":"great"}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions//rate", body)
		w := httptest.NewRecorder()
		h.HandleRateExtension(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid rating value", func(t *testing.T) {
		body := bytes.NewBufferString(`{"rating":10,"review":"too high"}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions/ext-1/rate", body)
		req = mux.SetURLVars(req, map[string]string{"id": "ext-1"})
		w := httptest.NewRecorder()
		h.HandleRateExtension(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for rating > 5, got %d", w.Code)
		}
	})

	t.Run("valid rating", func(t *testing.T) {
		body := bytes.NewBufferString(`{"rating":5,"review":"excellent"}`)
		req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions/ext-1/rate", body)
		req = mux.SetURLVars(req, map[string]string{"id": "ext-1"})
		w := httptest.NewRecorder()
		h.HandleRateExtension(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestHandleGetMyRating(t *testing.T) {
	repo := newMockRepo()
	repo.extensions["ext-1"] = &Extension{ID: "ext-1"}
	tenantID := uuid.New().String()
	repo.ratings["ext-1:"+tenantID] = &Rating{ExtensionID: "ext-1", TenantID: tenantID, Rating: 4, Review: "good"}
	h := NewHandler(repo)

	t.Run("returns existing rating", func(t *testing.T) {
		body := bytes.NewBuffer(nil)
		req := newAuthedRequestWithTenant(http.MethodGet, "/api/marketplace/extensions/ext-1/my-rating", body, tenantID)
		req = mux.SetURLVars(req, map[string]string{"id": "ext-1"})
		w := httptest.NewRecorder()
		h.HandleGetMyRating(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		r, ok := resp["rating"].(map[string]interface{})
		if !ok {
			t.Fatal("response should include rating object")
		}
		if r["rating"] != float64(4) {
			t.Errorf("expected rating 4, got %v", r["rating"])
		}
	})

	t.Run("returns null for missing rating", func(t *testing.T) {
		req := newAuthedRequest(http.MethodGet, "/api/marketplace/extensions/ext-1/my-rating", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "ext-1"})
		w := httptest.NewRecorder()
		h.HandleGetMyRating(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `"rating":null`) && !strings.Contains(body, `"rating": null`) {
			t.Errorf("expected null rating in body, got %s", body)
		}
	})
}

func TestHandleCheckUpdates(t *testing.T) {
	repo := newMockRepo()
	repo.updates = []ExtensionUpdate{
		{InstalledPluginID: "p1", InstalledVersion: "1.0.0", ExtensionID: "ext-1", LatestVersion: "2.0.0"},
	}
	h := NewHandler(repo)

	body := bytes.NewBufferString(`[{"id":"p1","name":"test","version":"1.0.0"}]`)
	req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions/check-updates", body)
	w := httptest.NewRecorder()
	h.HandleCheckUpdates(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInstallExtensionWithPlugin_MissingID(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions//install", nil)
	req = mux.SetURLVars(req, map[string]string{"id": ""})
	w := httptest.NewRecorder()
	h.HandleInstallExtensionWithPlugin(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleInstallExtensionWithPlugin_NotFound(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions/missing/install", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "missing"})
	w := httptest.NewRecorder()
	h.HandleInstallExtensionWithPlugin(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInstallExtensionWithPlugin_Unauthorized(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/extensions/ext-1/install", nil)
	w := httptest.NewRecorder()
	h.HandleInstallExtensionWithPlugin(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandleInstallExtensionWithPlugin_Success(t *testing.T) {
	repo := newMockRepo()
	repo.extensions["ext-1"] = &Extension{
		ID:      "ext-1",
		Name:    "test-plugin",
		Version: "1.0.0",
		Manifest: map[string]interface{}{
			"permissions": map[string]interface{}{},
		},
	}
	h := NewHandler(repo)

	req := newAuthedRequest(http.MethodPost, "/api/marketplace/extensions/ext-1/install", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "ext-1"})
	w := httptest.NewRecorder()
	h.HandleInstallExtensionWithPlugin(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
