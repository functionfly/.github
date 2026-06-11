package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type mockRepo struct {
	plugins        map[string]*Plugin
	versions       map[string][]PluginVersion
	permissions    map[string][]PluginPermission
	sandboxes      map[string]*PluginSandbox
	analyticsCalls []*PluginAnalytics

	listErr      error
	getErr       error
	createErr    error
	deleteErr    error
	setStatusErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		plugins:     map[string]*Plugin{},
		versions:    map[string][]PluginVersion{},
		permissions: map[string][]PluginPermission{},
		sandboxes:   map[string]*PluginSandbox{},
	}
}

func (m *mockRepo) List(_ context.Context, _ ListPluginsParams) ([]Plugin, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		out = append(out, *p)
	}
	return out, nil
}

func (m *mockRepo) Get(_ context.Context, _, pluginID string) (*Plugin, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.plugins[pluginID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (m *mockRepo) Create(_ context.Context, plugin *Plugin) error {
	if m.createErr != nil {
		return m.createErr
	}
	if plugin.ID == "" {
		plugin.ID = uuid.NewString()
	}
	m.plugins[plugin.ID] = plugin
	return nil
}

func (m *mockRepo) Update(_ context.Context, plugin *Plugin) error {
	if _, ok := m.plugins[plugin.ID]; !ok {
		return errors.New("plugin not found")
	}
	m.plugins[plugin.ID] = plugin
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _, pluginID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.plugins, pluginID)
	return nil
}

func (m *mockRepo) SetStatus(_ context.Context, _, pluginID string, status PluginStatus) error {
	if m.setStatusErr != nil {
		return m.setStatusErr
	}
	p, ok := m.plugins[pluginID]
	if !ok {
		return errors.New("plugin not found")
	}
	p.Status = status
	return nil
}

func (m *mockRepo) SetError(_ context.Context, _, pluginID string, errMsg string) error {
	p, ok := m.plugins[pluginID]
	if !ok {
		return errors.New("plugin not found")
	}
	if errMsg == "" {
		p.ErrorMessage = nil
	} else {
		p.ErrorMessage = &errMsg
	}
	return nil
}

func (m *mockRepo) UpdateConfig(_ context.Context, _, pluginID string, config map[string]string) error {
	p, ok := m.plugins[pluginID]
	if !ok {
		return errors.New("plugin not found")
	}
	p.Config = config
	return nil
}

func (m *mockRepo) GetEnabledByType(_ context.Context, _ string, pluginType PluginType) (*Plugin, error) {
	for _, p := range m.plugins {
		if p.PluginType == pluginType && p.Status == PluginStatusEnabled {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetSandbox(_ context.Context, pluginID string) (*PluginSandbox, error) {
	return m.sandboxes[pluginID], nil
}

func (m *mockRepo) UpsertSandbox(_ context.Context, sandbox *PluginSandbox) error {
	m.sandboxes[sandbox.PluginID] = sandbox
	return nil
}

func (m *mockRepo) ListPermissions(_ context.Context, pluginID string) ([]PluginPermission, error) {
	return m.permissions[pluginID], nil
}

func (m *mockRepo) SetPermission(_ context.Context, perm *PluginPermission) error {
	m.permissions[perm.PluginID] = append(m.permissions[perm.PluginID], *perm)
	return nil
}

func (m *mockRepo) CreateVersion(_ context.Context, version *PluginVersion) error {
	m.versions[version.PluginID] = append(m.versions[version.PluginID], *version)
	return nil
}

func (m *mockRepo) ListVersions(_ context.Context, pluginID string) ([]PluginVersion, error) {
	return m.versions[pluginID], nil
}

func (m *mockRepo) GetPreviousVersion(_ context.Context, pluginID, currentVersion string) (*PluginVersion, error) {
	versions := m.versions[pluginID]
	for i := range versions {
		if versions[i].Version != currentVersion {
			return &versions[i], nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetTelemetrySummary(_ context.Context, _, _ string, _ string) (*TelemetrySummary, error) {
	return &TelemetrySummary{Executions: 1, Errors: 0, ErrorRate: 0}, nil
}

func (m *mockRepo) RecordAnalytics(_ context.Context, analytics *PluginAnalytics) error {
	m.analyticsCalls = append(m.analyticsCalls, analytics)
	return nil
}

func newCtxRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	r := httptest.NewRequest(method, path, &buf)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	tenantID := uuid.New()
	claims := &auth.Claims{
		UserID:   uuid.New(),
		TenantID: tenantID,
	}
	return middleware.SetUserInContext(r, claims)
}

func newTestRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/plugins", h.HandleListPlugins).Methods("GET")
	r.HandleFunc("/plugins", h.HandleInstallPlugin).Methods("POST")
	r.HandleFunc("/plugins/{id}", h.HandleGetPlugin).Methods("GET")
	r.HandleFunc("/plugins/{id}", h.HandleUninstallPlugin).Methods("DELETE")
	r.HandleFunc("/plugins/{id}/enable", h.HandleEnablePlugin).Methods("POST")
	r.HandleFunc("/plugins/{id}/disable", h.HandleDisablePlugin).Methods("POST")
	r.HandleFunc("/plugins/{id}/pause", h.HandlePausePlugin).Methods("POST")
	r.HandleFunc("/plugins/{id}/rollback", h.HandleRollbackPlugin).Methods("POST")
	r.HandleFunc("/plugins/{id}/permissions", h.HandleGetPermissions).Methods("GET")
	r.HandleFunc("/plugins/{id}/permissions", h.HandleSetPermission).Methods("POST")
	r.HandleFunc("/plugins/{id}/versions", h.HandleListVersions).Methods("GET")
	r.HandleFunc("/plugins/{id}/sandbox", h.HandleGetSandbox).Methods("GET")
	return r
}

func doRequest(t *testing.T, h *Handler, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, r)
	return w
}

func TestHandleListPlugins_Success(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test", Status: PluginStatusEnabled}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Plugins []Plugin `json:"plugins"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Plugins) != 1 || resp.Plugins[0].ID != "p1" {
		t.Fatalf("expected plugin p1, got %+v", resp.Plugins)
	}
}

func TestHandleListPlugins_Unauthorized(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	r := httptest.NewRequest(http.MethodGet, "/plugins", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleInstallPlugin_Success(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	req := InstallPluginRequest{
		PluginType: PluginTypeUI,
		Name:       "hello-world",
		Version:    "1.0.0",
		AuthorName: "Tester",
		Category:   "ui",
	}
	r := newCtxRequest(t, http.MethodPost, "/plugins", req)
	w := doRequest(t, h, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if len(repo.plugins) != 1 {
		t.Fatalf("expected 1 plugin stored, got %d", len(repo.plugins))
	}
	if len(repo.versions) == 0 {
		t.Fatalf("expected a version to be recorded")
	}
}

func TestHandleInstallPlugin_MissingFields(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	req := map[string]string{"name": ""}
	r := newCtxRequest(t, http.MethodPost, "/plugins", req)
	w := doRequest(t, h, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetPlugin_NotFound(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins/missing", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleUninstallPlugin_Success(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test"}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodDelete, "/plugins/p1", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, ok := repo.plugins["p1"]; ok {
		t.Fatalf("expected plugin to be removed")
	}
}

func TestHandleEnablePlugin_UpdatesStatus(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test", Status: PluginStatusDisabled}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodPost, "/plugins/p1/enable", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.plugins["p1"].Status != PluginStatusEnabled {
		t.Fatalf("expected enabled, got %s", repo.plugins["p1"].Status)
	}
}

func TestHandlePausePlugin_UpdatesStatus(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test", Status: PluginStatusEnabled}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodPost, "/plugins/p1/pause", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if repo.plugins["p1"].Status != PluginStatusPaused {
		t.Fatalf("expected paused, got %s", repo.plugins["p1"].Status)
	}
}

func TestHandleRollbackPlugin_NoPrevious(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test", Version: "1.0.0"}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodPost, "/plugins/p1/rollback", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleRollbackPlugin_ToSpecificVersion(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test", Version: "2.0.0"}
	repo.versions["p1"] = []PluginVersion{
		{PluginID: "p1", Version: "1.0.0", Manifest: map[string]interface{}{"v": "1"}},
		{PluginID: "p1", Version: "2.0.0", Manifest: map[string]interface{}{"v": "2"}},
	}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodPost, "/plugins/p1/rollback?to_version=1.0.0", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.plugins["p1"].Version != "1.0.0" {
		t.Fatalf("expected rolled back to 1.0.0, got %s", repo.plugins["p1"].Version)
	}
}

func TestHandleSetPermission_StoresPermission(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test"}
	h := NewHandler(repo)

	body := SetPermissionRequest{
		PermissionType:   "api",
		PermissionAction: "read",
		Resource:         "functions",
		Granted:          true,
	}
	r := newCtxRequest(t, http.MethodPost, "/plugins/p1/permissions", body)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(repo.permissions["p1"]) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(repo.permissions["p1"]))
	}
	if !repo.permissions["p1"][0].Granted {
		t.Fatalf("expected permission granted=true")
	}
}

func TestHandleListVersions_EmptyWhenNoVersions(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test"}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins/p1/versions", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Versions []PluginVersion `json:"versions"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Versions == nil {
		t.Fatalf("expected non-nil empty versions slice")
	}
	if len(resp.Versions) != 0 {
		t.Fatalf("expected 0 versions, got %d", len(resp.Versions))
	}
}

func TestHandleGetSandbox_NilWhenAbsent(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Test"}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins/p1/sandbox", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Sandbox *PluginSandbox `json:"sandbox"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Sandbox != nil {
		t.Fatalf("expected nil sandbox, got %+v", resp.Sandbox)
	}
}

func TestHandleListPlugins_FiltersByStatus(t *testing.T) {
	repo := newMockRepo()
	repo.plugins["p1"] = &Plugin{ID: "p1", Name: "Enabled", Status: PluginStatusEnabled, PluginType: PluginTypeUI}
	repo.plugins["p2"] = &Plugin{ID: "p2", Name: "Disabled", Status: PluginStatusDisabled, PluginType: PluginTypeUI}
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins?status=disabled", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Plugins []Plugin `json:"plugins"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Plugins) != 1 || resp.Plugins[0].ID != "p2" {
		t.Fatalf("expected only p2, got %+v", resp.Plugins)
	}
}

func TestHandleListPlugins_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(repo)

	r := newCtxRequest(t, http.MethodGet, "/plugins?status=invalid", nil)
	w := doRequest(t, h, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
