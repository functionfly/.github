package github

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler() (*Handler, *storage.GitHubRepository) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Use a nil DB — most tests don't hit the DB. Tests that trigger
	// goroutines with DB access (e.g. webhook) use newTestHandlerWithMockDB.
	repo := storage.NewGitHubRepository(nil)

	h := &Handler{
		repo:       nil,
		githubRepo: repo,
		logger:     logger,
		vaultKey:   "12345678901234567890123456789012",
		baseURL:    "http://localhost:3000",
	}

	return h, repo
}

func newTestHandlerWithMockDB(t *testing.T) (*Handler, *storage.GitHubRepository, sqlmock.Sqlmock) {
	t.Helper()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repo := storage.NewGitHubRepository(db)

	h := &Handler{
		repo:       nil,
		githubRepo: repo,
		logger:     logger,
		vaultKey:   "12345678901234567890123456789012",
		baseURL:    "http://localhost:3000",
	}

	return h, repo, mock
}

func setAuthContext(r *http.Request, userID, tenantID uuid.UUID) *http.Request {
	claims := &auth.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
	}
	return middleware.SetUserInContext(r, claims)
}

func TestHandler_RequireAuth(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("returns claims when authenticated", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		req := httptest.NewRequest("GET", "/", nil)
		req = setAuthContext(req, userID, tenantID)
		w := httptest.NewRecorder()

		claims := h.requireAuth(w, req)
		require.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		claims := h.requireAuth(w, req)
		assert.Nil(t, claims)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "unauthorized", resp["error"])
	})
}

func TestHandler_RequireAuthOrToken(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("returns claims from middleware context", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		req := httptest.NewRequest("GET", "/", nil)
		req = setAuthContext(req, userID, tenantID)
		w := httptest.NewRecorder()

		claims := h.requireAuthOrToken(w, req)
		require.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("returns 401 without token and no context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		claims := h.requireAuthOrToken(w, req)
		assert.Nil(t, claims)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_RespondJSON(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("writes JSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("writes error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.respondError(w, http.StatusBadRequest, "bad_request", "Invalid input")

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "bad_request", resp["error"])
		assert.Equal(t, "Invalid input", resp["message"])
	})
}

func TestHandler_ParseUUID(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("parses valid UUID from mux vars", func(t *testing.T) {
		id := uuid.New()
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"id": id.String()})

		parsed, err := h.parseUUID(req, "id")
		require.NoError(t, err)
		assert.Equal(t, id, parsed)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "not-a-uuid"})

		_, err := h.parseUUID(req, "id")
		assert.Error(t, err)
	})
}

func TestHandler_HandleConnect_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/connect", nil)
	w := httptest.NewRecorder()

	h.HandleConnect(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleConnect_Authenticated(t *testing.T) {
	h, _, mock := newTestHandlerWithMockDB(t)

	mock.ExpectExec("INSERT INTO oauth_states").WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest("GET", "/connect", nil)
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()

	h.HandleConnect(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp ConnectResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.URL, "github.com/login/oauth/authorize")
	assert.Contains(t, resp.URL, "client_id=")
	assert.Contains(t, resp.URL, "scope=")
	assert.Contains(t, resp.URL, "state=")
}

func TestHandler_HandleCallback_MissingParams(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("missing code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?state=abc", nil)
		w := httptest.NewRecorder()
		h.HandleCallback(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing state", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=abc", nil)
		w := httptest.NewRecorder()
		h.HandleCallback(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing both", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback", nil)
		w := httptest.NewRecorder()
		h.HandleCallback(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_HandleCallback_InvalidState(t *testing.T) {
	h, _, mock := newTestHandlerWithMockDB(t)

	mock.ExpectQuery("SELECT state, user_id, tenant_id, provider, expires_at FROM oauth_states").
		WithArgs("nonexistent", sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/callback?code=abc&state=nonexistent", nil)
	w := httptest.NewRecorder()
	h.HandleCallback(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_state", resp["error"])
}

func TestHandler_HandleGetConnection_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/connection", nil)
	w := httptest.NewRecorder()
	h.HandleGetConnection(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleListRepos_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/repos", nil)
	w := httptest.NewRecorder()
	h.HandleListRepos(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleListImports_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/imports", nil)
	w := httptest.NewRecorder()
	h.HandleListImports(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleImport_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(ImportRequest{
		RepoID:       uuid.New(),
		FunctionName: "test-func",
		Branch:       "main",
	})
	req := httptest.NewRequest("POST", "/imports", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.HandleImport(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleImport_MissingRepoID(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(ImportRequest{
		FunctionName: "test-func",
	})
	req := httptest.NewRequest("POST", "/imports", bytes.NewReader(body))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleImport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "missing_repo", resp["error"])
}

func TestHandler_HandleImport_MissingFunctionName(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(ImportRequest{
		RepoID: uuid.New(),
	})
	req := httptest.NewRequest("POST", "/imports", bytes.NewReader(body))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleImport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "missing_name", resp["error"])
}

func TestHandler_HandleImport_InvalidBody(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/imports", bytes.NewReader([]byte("not json")))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleImport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_HandleBulkImport_EmptyImports(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(BulkImportRequest{
		Imports: []ImportRequest{},
	})
	req := httptest.NewRequest("POST", "/imports/bulk", bytes.NewReader(body))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleBulkImport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "empty_imports", resp["error"])
}

func TestHandler_HandleBulkImport_TooManyImports(t *testing.T) {
	h, _ := newTestHandler()

	imports := make([]ImportRequest, 21)
	for i := range imports {
		imports[i] = ImportRequest{
			RepoID:       uuid.New(),
			FunctionName: "func",
		}
	}
	body, _ := json.Marshal(BulkImportRequest{Imports: imports})
	req := httptest.NewRequest("POST", "/imports/bulk", bytes.NewReader(body))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleBulkImport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "too_many_imports", resp["error"])
}

func TestHandler_HandleDisconnect_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("DELETE", "/connection", nil)
	w := httptest.NewRecorder()
	h.HandleDisconnect(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleRefreshToken_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/connection/refresh", nil)
	w := httptest.NewRecorder()
	h.HandleRefreshToken(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleListTemplates_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/templates", nil)
	w := httptest.NewRecorder()
	h.HandleListTemplates(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleCreateTemplate_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(TemplateRequest{Name: "test"})
	req := httptest.NewRequest("POST", "/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.HandleCreateTemplate(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleCreateTemplate_MissingName(t *testing.T) {
	h, _ := newTestHandler()

	body, _ := json.Marshal(TemplateRequest{
		Config: json.RawMessage(`{}`),
	})
	req := httptest.NewRequest("POST", "/templates", bytes.NewReader(body))
	req = setAuthContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()
	h.HandleCreateTemplate(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "missing_name", resp["error"])
}

func TestHandler_HandleWebhook_PingEvent(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/webhook", nil)
	req.Header.Set("X-GitHub-Event", "ping")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "pong", resp["status"])
}

func TestHandler_HandleWebhook_MissingEventHeader(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_HandleWebhook_WrongMethod(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/webhook", nil)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandler_HandleWebhook_IgnoredEvent(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader([]byte("{}")))
	req.Header.Set("X-GitHub-Event", "issues")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ignored", resp["status"])
}

func TestHandler_HandleWebhook_InvalidPushPayload(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader([]byte("not json")))
	req.Header.Set("X-GitHub-Event", "push")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_HandleWebhook_InvalidPRPayload(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader([]byte("not json")))
	req.Header.Set("X-GitHub-Event", "pull_request")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_HandleWebhook_ValidPushPayload(t *testing.T) {
	h, _, mock := newTestHandlerWithMockDB(t)

	// The webhook handler spawns a goroutine that queries for the repo.
	// Return no rows so the goroutine exits cleanly.
	rows := sqlmock.NewRows([]string{"id", "connection_id", "github_repo_id", "full_name", "name", "owner",
		"description", "default_branch", "language", "languages", "is_private", "is_fork", "is_archived", "topics",
		"stars_count", "forks_count", "size_kb", "pushed_at", "html_url", "clone_url", "ssh_url",
		"detected_functions", "detected_runtime", "has_functionfly_json", "import_status", "metadata", "last_scanned_at",
		"created_at", "updated_at"})
	mock.ExpectQuery("SELECT .+ FROM github_repos WHERE github_repo_id").WillReturnRows(rows)

	pushEvent := githubsvc.WebhookPushEvent{
		Ref:    "refs/heads/main",
		Before: "aaa",
		After:  "bbb",
	}
	pushEvent.Repository.ID = 12345
	pushEvent.Repository.FullName = "owner/repo"

	body, _ := json.Marshal(pushEvent)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-GitHub-Delivery", "delivery-123")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "accepted", resp["status"])

	// Wait briefly for the goroutine to complete
	_ = mock
}

func TestHandler_HandleImportProgress_Unauthenticated(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/imports/"+uuid.New().String()+"/progress", nil)
	w := httptest.NewRecorder()
	h.HandleImportProgress(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_MapImportResponse(t *testing.T) {
	h, _ := newTestHandler()

	importID := uuid.New()
	functionID := uuid.New()
	versionID := uuid.New()
	commitSHA := "abc123"

	imp := &storage.GitHubImport{
		ID:                importID,
		RepoID:            uuid.New(),
		FunctionName:      "my-func",
		SourceBranch:      "main",
		Visibility:        "private",
		AutoSyncEnabled:   true,
		SyncBranches:      json.RawMessage(`["main"]`),
		Status:            "completed",
		Progress:          100,
		FunctionID:        &functionID,
		FunctionVersionID: &versionID,
		CommitSHA:         &commitSHA,
		FilesImported:     42,
		TotalSizeBytes:    1024000,
	}

	resp := h.mapImportResponse(imp)
	assert.Equal(t, importID, resp.ID)
	assert.Equal(t, "my-func", resp.FunctionName)
	assert.Equal(t, "main", resp.SourceBranch)
	assert.Equal(t, "private", resp.Visibility)
	assert.True(t, resp.AutoSyncEnabled)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 100, resp.Progress)
	assert.Equal(t, &functionID, resp.FunctionID)
	assert.Equal(t, &versionID, resp.FunctionVersionID)
	assert.Equal(t, &commitSHA, resp.CommitSHA)
	assert.Equal(t, 42, resp.FilesImported)
	assert.Equal(t, int64(1024000), resp.TotalSizeBytes)
}

func TestHandler_MapRepoResponse(t *testing.T) {
	h, _ := newTestHandler()

	repoID := uuid.New()
	repo := &storage.GitHubRepo{
		ID:                 repoID,
		FullName:           "owner/repo",
		Name:               "repo",
		Owner:              "owner",
		DefaultBranch:      "main",
		Language:           strPtr("Go"),
		Languages:          json.RawMessage(`{"Go": 100}`),
		IsPrivate:          false,
		StarsCount:         42,
		ForksCount:         7,
		HtmlURL:            "https://github.com/owner/repo",
		DetectedFunctions:  json.RawMessage(`[]`),
		DetectedRuntime:    strPtr("go1.22"),
		HasFunctionflyJSON: true,
		ImportStatus:       "imported",
	}

	resp := h.mapRepoResponse(repo)
	assert.Equal(t, repoID, resp.ID)
	assert.Equal(t, "owner/repo", resp.FullName)
	assert.Equal(t, "repo", resp.Name)
	assert.Equal(t, "owner", resp.Owner)
	assert.Equal(t, "main", resp.DefaultBranch)
	assert.Equal(t, "Go", *resp.Language)
	assert.False(t, resp.IsPrivate)
	assert.Equal(t, 42, resp.StarsCount)
	assert.Equal(t, 7, resp.ForksCount)
	assert.Equal(t, "go1.22", *resp.DetectedRuntime)
	assert.True(t, resp.HasFunctionflyJSON)
	assert.Equal(t, "imported", resp.ImportStatus)
}

func TestHandler_MapTemplateResponse(t *testing.T) {
	h, _ := newTestHandler()

	tmplID := uuid.New()
	tmpl := &storage.GitHubImportTemplate{
		ID:             tmplID,
		Name:           "My Template",
		Description:    strPtr("A test template"),
		Config:         json.RawMessage(`{"runtime": "node18"}`),
		DetectionRules: json.RawMessage(`{}`),
		IsDefault:      true,
		UsageCount:     5,
	}

	resp := h.mapTemplateResponse(tmpl)
	assert.Equal(t, tmplID, resp.ID)
	assert.Equal(t, "My Template", resp.Name)
	assert.Equal(t, "A test template", *resp.Description)
	assert.True(t, resp.IsDefault)
	assert.Equal(t, 5, resp.UsageCount)
}

func TestHandler_GetProgressChan(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("creates new channel", func(t *testing.T) {
		id := uuid.New()
		ch := h.getProgressChan(id)
		assert.NotNil(t, ch)
	})

	t.Run("returns same channel for same ID", func(t *testing.T) {
		id := uuid.New()
		ch1 := h.getProgressChan(id)
		ch2 := h.getProgressChan(id)
		assert.Equal(t, ch1, ch2)
	})
}

func TestHandler_CompleteProgress(t *testing.T) {
	h, _ := newTestHandler()

	t.Run("closes and removes channel", func(t *testing.T) {
		id := uuid.New()
		ch := h.getProgressChan(id)
		h.completeProgress(id)

		// Channel should be closed
		_, ok := <-ch
		assert.False(t, ok)

		// Getting a new channel should create a new one
		ch2 := h.getProgressChan(id)
		assert.NotEqual(t, ch, ch2)
	})

	t.Run("safe to call twice", func(t *testing.T) {
		id := uuid.New()
		h.getProgressChan(id)
		h.completeProgress(id)
		h.completeProgress(id) // should not panic
	})
}

func strPtr(s string) *string {
	return &s
}
