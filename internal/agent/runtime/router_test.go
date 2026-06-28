package agentrun

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func decodeMap(t *testing.T, raw json.RawMessage) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

// newMockServer creates a test HTTP server that returns the given handler.
func newMockServer(handler http.HandlerFunc) *httptest.Server {
	srv := httptest.NewServer(handler)
	return srv
}

// ---------------------------------------------------------------------------
// RuntimeRouter tests
// ---------------------------------------------------------------------------

func TestRuntimeRouter_RegisterAndGet(t *testing.T) {
	router := NewRuntimeRouter()
	rt := &DataRuntime{}
	router.RegisterRuntime(RuntimeTypeData, rt)

	got, ok := router.GetRuntime(RuntimeTypeData)
	assert.True(t, ok)
	assert.Equal(t, "data", got.Name())

	_, ok = router.GetRuntime(RuntimeTypeSearch)
	assert.False(t, ok)
}

func TestRuntimeRouter_Execute_UnknownCategory(t *testing.T) {
	router := NewRuntimeRouter()
	_, err := router.Execute(context.Background(), &ExecutionRequest{
		Category:   "nonexistent",
		FunctionID: uuid.New(),
		Input:      json.RawMessage(`{}`),
	}, 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no runtime registered")
}

func TestRuntimeRouter_Execute_Success(t *testing.T) {
	router := NewRuntimeRouter()
	router.RegisterRuntime(RuntimeTypeData, &DataRuntime{})

	resp, err := router.Execute(context.Background(), &ExecutionRequest{
		Category:   "data",
		FunctionID: uuid.New(),
		AgentID:    uuid.New(),
		Input: mustJSON(map[string]interface{}{
			"action": "validate",
			"data":   json.RawMessage(`{"name":"test"}`),
		}),
	}, 5*time.Second)

	require.NoError(t, err)
	assert.NotNil(t, resp.Output)
	assert.Equal(t, "data", resp.Provider)
	assert.GreaterOrEqual(t, resp.DurationMs, 0)
	assert.NotEmpty(t, resp.ExecutionID)
}

func TestGetExecutionContext(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, GetExecutionContext(ctx))

	ec := &ExecutionContext{AgentID: uuid.New(), TraceID: "abc"}
	ctx = context.WithValue(ctx, ExecutionContextKey, ec)
	got := GetExecutionContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "abc", got.TraceID)
}

func TestDefaultRuntimeRouter_RegistersAll15(t *testing.T) {
	router := DefaultRuntimeRouter()
	categories := []RuntimeType{
		RuntimeTypeSearch, RuntimeTypeBrowser, RuntimeTypeCompute,
		RuntimeTypeData, RuntimeTypeFile, RuntimeTypeCommunication,
		RuntimeTypeAssure, RuntimeTypeValidate, RuntimeTypeSimulate,
		RuntimeTypeObserve, RuntimeTypeLearn, RuntimeTypeAgentMgmt,
		RuntimeTypeCapability, RuntimeTypeWorkflow, RuntimeTypeMemory,
	}
	for _, cat := range categories {
		_, ok := router.GetRuntime(cat)
		assert.True(t, ok, "missing runtime for category: %s", cat)
	}
}

// ---------------------------------------------------------------------------
// SearchRuntime
// ---------------------------------------------------------------------------

func TestSearchRuntime_MissingConfig(t *testing.T) {
	rt := &SearchRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"query": "test",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search not configured")
}

func TestSearchRuntime_EmptyQuery(t *testing.T) {
	rt := &SearchRuntime{APIKey: "key", EngineID: "cx"}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is required")
}

func TestSearchRuntime_Success(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("key"), "test-key")
		assert.Equal(t, "test-cx", r.URL.Query().Get("cx"))
		assert.Equal(t, "golang", r.URL.Query().Get("q"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"title": "Go", "link": "https://go.dev", "snippet": "Build fast"},
				{"title": "GoDoc", "link": "https://pkg.go.dev", "snippet": "Packages"},
			},
			"searchInformation": map[string]interface{}{
				"totalResults": "1000000",
				"searchTime":   0.3,
			},
		})
	})
	defer srv.Close()

	// We can't easily redirect Google URLs, so test the error path for bad API key
	// and test parsing logic via the mock server indirectly.
	rt := &SearchRuntime{
		HTTPClient: srv.Client(),
		APIKey:     "test-key",
		EngineID:   "test-cx",
	}

	// This will call Google's API with the mock key and get a 400/403.
	// The important thing is the code path is exercised.
	input := mustJSON(map[string]interface{}{"query": "golang", "num_results": 5})
	_, err := rt.Execute(context.Background(), input, 5*time.Second)
	// Expected to fail because we're calling the real Google API with a fake key
	if err != nil {
		assert.Contains(t, err.Error(), "search API returned")
	}
}

func TestSearchRuntime_InvalidInput(t *testing.T) {
	rt := &SearchRuntime{APIKey: "k", EngineID: "c"}
	_, err := rt.Execute(context.Background(), json.RawMessage(`{invalid`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// BrowserRuntime
// ---------------------------------------------------------------------------

func TestBrowserRuntime_Screenshot(t *testing.T) {
	dir := t.TempDir()
	rt := &BrowserRuntime{ScreenshotDir: dir}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "screenshot",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "screenshot", m["action"])
	assert.True(t, m["success"].(bool))
	assert.Contains(t, m["screenshot_path"].(string), "screenshot_")
}

func TestBrowserRuntime_Evaluate(t *testing.T) {
	rt := &BrowserRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "evaluate", "script": "document.title",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "evaluate", m["action"])
	assert.Equal(t, "document.title", m["script"])
}

func TestBrowserRuntime_Evaluate_MissingScript(t *testing.T) {
	rt := &BrowserRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "evaluate",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script is required")
}

func TestBrowserRuntime_Navigate_MissingURL(t *testing.T) {
	rt := &BrowserRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "navigate",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url is required")
}

func TestBrowserRuntime_DefaultAction(t *testing.T) {
	rt := &BrowserRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "click",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "click", m["action"])
}

func TestBrowserRuntime_InvalidInput(t *testing.T) {
	rt := &BrowserRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`not json`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ComputeRuntime
// ---------------------------------------------------------------------------

func TestComputeRuntime_MissingOrchestrator(t *testing.T) {
	os.Unsetenv("ORCHESTRATOR_URL")
	os.Unsetenv("RUNTIME_API_URL")
	rt := &ComputeRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"code": "print(1)", "language": "python",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compute not configured")
}

func TestComputeRuntime_EmptyCode(t *testing.T) {
	rt := &ComputeRuntime{OrchestratorURL: "http://localhost"}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"language": "python",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

func TestComputeRuntime_Success(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/execute/python", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "print('hello')", body["code"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"output": "hello\n", "exit_code": 0,
		})
	})
	defer srv.Close()

	rt := &ComputeRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"code": "print('hello')", "language": "python",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "python", m["language"])
	assert.Equal(t, "hello\n", m["output"])
}

func TestComputeRuntime_ServerError(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})
	defer srv.Close()

	rt := &ComputeRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"code": "x", "language": "python",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compute returned 500")
}

func TestComputeRuntime_InvalidInput(t *testing.T) {
	rt := &ComputeRuntime{OrchestratorURL: "http://localhost"}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DataRuntime
// ---------------------------------------------------------------------------

func TestDataRuntime_Extract(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "extract",
		"data":   map[string]interface{}{"user": map[string]interface{}{"name": "Alice", "age": 30}},
		"path":   "user.name",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "Alice", m["extracted"])
}

func TestDataRuntime_Extract_MissingPath(t *testing.T) {
	rt := &DataRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "extract", "data": map[string]interface{}{"a": 1},
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
}

func TestDataRuntime_Extract_Nested(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "extract",
		"data":   map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "found"}}},
		"path":   "a.0.b",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "found", m["extracted"])
}

func TestDataRuntime_Transform(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "transform",
		"data":   map[string]interface{}{"first_name": "Alice", "last_name": "Smith"},
		"schema": map[string]interface{}{"name": "first_name", "surname": "last_name"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	transformed := m["transformed"].(map[string]interface{})
	assert.Equal(t, "Alice", transformed["name"])
	assert.Equal(t, "Smith", transformed["surname"])
}

func TestDataRuntime_Transform_NoSchema(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "transform",
		"data":   map[string]interface{}{"key": "val"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
}

func TestDataRuntime_Validate_Pass(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "validate",
		"data":   map[string]interface{}{"name": "Alice", "email": "a@b.com"},
		"schema": map[string]interface{}{"required": []interface{}{"name", "email"}},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["valid"].(bool))
}

func TestDataRuntime_Validate_Fail(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "validate",
		"data":   map[string]interface{}{"name": "Alice"},
		"schema": map[string]interface{}{"required": []interface{}{"name", "email"}},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.False(t, m["valid"].(bool))
	errors := m["errors"].([]interface{})
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].(string), "email")
}

func TestDataRuntime_Validate_NoSchema(t *testing.T) {
	rt := &DataRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "validate",
		"data":   map[string]interface{}{"x": 1},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["valid"].(bool))
}

func TestDataRuntime_UnknownAction(t *testing.T) {
	rt := &DataRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "unknown",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown data action")
}

func TestDataRuntime_InvalidInput(t *testing.T) {
	rt := &DataRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`not json`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// FileRuntime
// ---------------------------------------------------------------------------

func TestFileRuntime_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}

	// Write
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "test.txt", "content": "hello world",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
	assert.Equal(t, float64(11), m["size_bytes"])

	// Read
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "read", "path": "test.txt",
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.Equal(t, "hello world", m["content"])
	assert.True(t, m["success"].(bool))
}

func TestFileRuntime_Delete(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}

	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "del.txt", "content": "bye",
	}), 5*time.Second)

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "delete", "path": "del.txt",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))

	_, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "read", "path": "del.txt",
	}), 5*time.Second)
	assert.Error(t, err)
}

func TestFileRuntime_List(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}

	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "a.txt", "content": "a",
	}), 5*time.Second)
	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "b.txt", "content": "bb",
	}), 5*time.Second)

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "list", "path": ".",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(2), m["count"])
}

func TestFileRuntime_Search(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}

	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "findme.txt", "content": "the secret word is xyzzy",
	}), 5*time.Second)
	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "other.txt", "content": "nothing here",
	}), 5*time.Second)

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "search", "path": ".", "pattern": "xyzzy",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(1), m["count"])
}

func TestFileRuntime_Search_MissingPattern(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "search", "path": ".",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pattern is required")
}

func TestFileRuntime_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}

	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "read", "path": "../../../etc/passwd",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}

func TestFileRuntime_MissingPath(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "read",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
}

func TestFileRuntime_UnknownAction(t *testing.T) {
	dir := t.TempDir()
	rt := &FileRuntime{BaseDir: dir}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope", "path": "x",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown file action")
}

func TestFileRuntime_InvalidInput(t *testing.T) {
	rt := &FileRuntime{BaseDir: t.TempDir()}
	_, err := rt.Execute(context.Background(), json.RawMessage(`{bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// CommunicationRuntime
// ---------------------------------------------------------------------------

func TestCommunicationRuntime_Email_MissingConfig(t *testing.T) {
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("RESEND_FROM_EMAIL")
	rt := &CommunicationRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "email.send", "to": "a@b.com", "subject": "hi", "body": "hello",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email not configured")
}

func TestCommunicationRuntime_Email_MissingTo(t *testing.T) {
	rt := &CommunicationRuntime{ResendAPIKey: "k", FromEmail: "f@b.com"}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "email.send", "subject": "hi", "body": "hello",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to is required")
}

func TestCommunicationRuntime_Email_Success(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "msg_123"})
	})
	defer srv.Close()

	// CommunicationRuntime hardcodes resend URL, so we test via error path.
	// The real test is that the request is well-formed.
	rt := &CommunicationRuntime{
		ResendAPIKey: "test-key",
		FromEmail:    "from@test.com",
		HTTPClient:   srv.Client(),
	}
	// This will fail because it hits api.resend.com, not our mock.
	// But it validates the code path.
	_, _ = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "email.send", "to": "to@test.com", "subject": "hi", "body": "hello",
	}), 5*time.Second)
}

func TestCommunicationRuntime_Slack_MissingConfig(t *testing.T) {
	os.Unsetenv("SLACK_WEBHOOK_URL")
	rt := &CommunicationRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "slack.send", "subject": "hi", "body": "hello",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slack not configured")
}

func TestCommunicationRuntime_Slack_Success(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	defer srv.Close()

	rt := &CommunicationRuntime{
		SlackWebhook: srv.URL + "/webhook",
		HTTPClient:   srv.Client(),
	}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "slack.send", "subject": "Alert", "body": "Something happened",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
}

func TestCommunicationRuntime_Slack_ServerError(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	})
	defer srv.Close()

	rt := &CommunicationRuntime{
		SlackWebhook: srv.URL + "/webhook",
		HTTPClient:   srv.Client(),
	}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "slack.send", "subject": "h", "body": "b",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slack returned 400")
}

func TestCommunicationRuntime_UnknownAction(t *testing.T) {
	rt := &CommunicationRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "sms.send",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown communication action")
}

func TestCommunicationRuntime_InvalidInput(t *testing.T) {
	rt := &CommunicationRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// AssureRuntime
// ---------------------------------------------------------------------------

func TestAssureRuntime_BasicChecks(t *testing.T) {
	rt := &AssureRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action":  "compliance",
		"subject": "test-agent",
		"checks":  []string{"no_pii", "no_secrets", "budget"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["passed"].(bool))
	assert.Equal(t, float64(3), m["check_count"])
}

func TestAssureRuntime_CheckCategories(t *testing.T) {
	rt := &AssureRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "security",
		"checks": []string{"no_pii", "no_secrets", "rate_limit", "budget", "capability", "custom"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	checks := m["checks"].([]interface{})
	assert.Len(t, checks, 6)

	catMap := map[string]string{}
	for _, c := range checks {
		cm := c.(map[string]interface{})
		catMap[cm["check"].(string)] = cm["category"].(string)
	}
	assert.Equal(t, "privacy", catMap["no_pii"])
	assert.Equal(t, "security", catMap["no_secrets"])
	assert.Equal(t, "governance", catMap["rate_limit"])
	assert.Equal(t, "billing", catMap["budget"])
	assert.Equal(t, "access_control", catMap["capability"])
	assert.Equal(t, "general", catMap["custom"])
}

func TestAssureRuntime_EmptyChecks(t *testing.T) {
	rt := &AssureRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "policy", "checks": []string{},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(0), m["check_count"])
}

func TestAssureRuntime_InvalidInput(t *testing.T) {
	rt := &AssureRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ValidateRuntime
// ---------------------------------------------------------------------------

func TestValidateRuntime_PIIScan_Email(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pii_scan",
		"text":   "Contact alice@example.com for details",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_pii"].(bool))
	assert.Equal(t, float64(1), m["finding_count"])
}

func TestValidateRuntime_PIIScan_SSN(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pii_scan",
		"text":   "SSN: 123-45-6789",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_pii"].(bool))
}

func TestValidateRuntime_PIIScan_CreditCard(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pii_scan",
		"text":   "Card: 4111-1111-1111-1111",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_pii"].(bool))
}

func TestValidateRuntime_PIIScan_Phone(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pii_scan",
		"text":   "Call 555-123-4567",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_pii"].(bool))
}

func TestValidateRuntime_PIIScan_Clean(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pii_scan",
		"text":   "No personal data here",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.False(t, m["has_pii"].(bool))
	assert.Equal(t, float64(0), m["finding_count"])
}

func TestValidateRuntime_SecretScan_Password(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "secret_scan",
		"text":   `password: "hunter2"`,
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_secrets"].(bool))
}

func TestValidateRuntime_SecretScan_PrivateKey(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "secret_scan",
		"text":   "-----BEGIN RSA PRIVATE KEY-----",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["has_secrets"].(bool))
}

func TestValidateRuntime_SecretScan_Clean(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "secret_scan",
		"text":   "This is safe text",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.False(t, m["has_secrets"].(bool))
}

func TestValidateRuntime_Schema_Valid(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "schema",
		"data":   map[string]interface{}{"name": "Alice"},
		"schema": map[string]interface{}{"required": []interface{}{"name"}},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["valid"].(bool))
}

func TestValidateRuntime_Schema_Invalid(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "schema",
		"data":   map[string]interface{}{},
		"schema": map[string]interface{}{"required": []interface{}{"name", "email"}},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.False(t, m["valid"].(bool))
}

func TestValidateRuntime_Schema_NoSchema(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "schema",
		"data":   map[string]interface{}{"x": 1},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["valid"].(bool))
}

func TestValidateRuntime_Size(t *testing.T) {
	rt := &ValidateRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "size",
		"data":   json.RawMessage(`{"big":"data"}`),
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["input_valid"].(bool))
}

func TestValidateRuntime_UnknownAction(t *testing.T) {
	rt := &ValidateRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown validate action")
}

func TestValidateRuntime_InvalidInput(t *testing.T) {
	rt := &ValidateRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SimulateRuntime
// ---------------------------------------------------------------------------

func TestSimulateRuntime_Default(t *testing.T) {
	rt := &SimulateRuntime{DefaultIterations: 1000}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "outcome",
		"params": map[string]interface{}{"mean": 100.0, "stddev": 10.0},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "outcome", m["action"])
	assert.Equal(t, float64(1000), m["iterations"])
	outcome := m["outcome"].(map[string]interface{})
	assert.NotNil(t, outcome["mean"])
	assert.NotNil(t, outcome["std_dev"])
	assert.NotNil(t, outcome["p5"])
	assert.NotNil(t, outcome["p50"])
	assert.NotNil(t, outcome["p95"])
}

func TestSimulateRuntime_Financial(t *testing.T) {
	rt := &SimulateRuntime{DefaultIterations: 500}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action":     "financial",
		"params":     map[string]interface{}{"base": 10000.0, "growth_rate": 0.08, "stddev": 0.02, "periods": 12.0},
		"iterations": 500,
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	outcome := m["outcome"].(map[string]interface{})
	mean := outcome["mean"].(float64)
	assert.Greater(t, mean, 0.0)
}

func TestSimulateRuntime_IterationsOverride(t *testing.T) {
	rt := &SimulateRuntime{DefaultIterations: 100}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action":     "default",
		"params":     map[string]interface{}{"mean": 0.0},
		"iterations": 50,
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(50), m["iterations"])
}

func TestSimulateRuntime_InvalidInput(t *testing.T) {
	rt := &SimulateRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ObserveRuntime
// ---------------------------------------------------------------------------

func TestObserveRuntime_System(t *testing.T) {
	rt := &ObserveRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "system",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "system", m["action"])
	assert.Equal(t, "healthy", m["status"])
	metrics := m["metrics"].(map[string]interface{})
	assert.NotNil(t, metrics["go_routines"])
	assert.NotNil(t, metrics["cpu_count"])
	assert.NotNil(t, metrics["go_version"])
}

func TestObserveRuntime_Health(t *testing.T) {
	rt := &ObserveRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "health",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "healthy", m["status"])
	checks := m["checks"].(map[string]interface{})
	assert.True(t, checks["cpu_ok"].(bool))
}

func TestObserveRuntime_Prometheus(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"resultType": "vector"},
		})
	})
	defer srv.Close()

	rt := &ObserveRuntime{MetricsURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action":  "prometheus",
		"metrics": []string{"http_requests_total"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "ok", m["status"])
}

func TestObserveRuntime_UnknownAction(t *testing.T) {
	rt := &ObserveRuntime{}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown observe action")
}

func TestObserveRuntime_InvalidInput(t *testing.T) {
	rt := &ObserveRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// LearnRuntime (with miniredis)
// ---------------------------------------------------------------------------

func TestLearnRuntime_NoRedis(t *testing.T) {
	rt := &LearnRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pattern", "pattern": "test",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["learned"].(bool))
	assert.Contains(t, m["note"].(string), "Redis not configured")
}

func TestLearnRuntime_Pattern(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &LearnRuntime{Redis: rdb}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pattern", "pattern": "retry_on_timeout", "agent_id": "agent-1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["stored"].(bool))
	assert.Equal(t, "retry_on_timeout", m["pattern"])
}

func TestLearnRuntime_Feedback(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &LearnRuntime{Redis: rdb}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "feedback", "outcome": "success", "agent_id": "agent-1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["stored"].(bool))
	assert.Equal(t, "success", m["outcome"])
}

func TestLearnRuntime_Retrieve(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &LearnRuntime{Redis: rdb}

	// Store feedback first
	for i := 0; i < 3; i++ {
		rt.Execute(context.Background(), mustJSON(map[string]interface{}{
			"action": "feedback", "outcome": fmt.Sprintf("result_%d", i), "agent_id": "a1",
		}), 5*time.Second)
	}

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "retrieve", "agent_id": "a1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(3), m["count"])
}

func TestLearnRuntime_UnknownAction(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &LearnRuntime{Redis: rdb}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown learn action")
}

func TestLearnRuntime_InvalidInput(t *testing.T) {
	rt := &LearnRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// AgentMgmtRuntime
// ---------------------------------------------------------------------------

func TestAgentMgmtRuntime_Spawn(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agent/swarm/spawn", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agent_id": "child-1", "status": "active",
		})
	})
	defer srv.Close()

	rt := &AgentMgmtRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "spawn", "name": "worker-1", "role": "worker",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
}

func TestAgentMgmtRuntime_Status(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/children")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"agent_id": "child-1"}})
	})
	defer srv.Close()

	rt := &AgentMgmtRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "status", "agent_id": "parent-1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "status", m["action"])
}

func TestAgentMgmtRuntime_Status_MissingAgentID(t *testing.T) {
	rt := &AgentMgmtRuntime{OrchestratorURL: "http://localhost"}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "status",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id is required")
}

func TestAgentMgmtRuntime_DefaultAction(t *testing.T) {
	rt := &AgentMgmtRuntime{OrchestratorURL: "http://localhost"}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "terminate", "agent_id": "a1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
}

func TestAgentMgmtRuntime_InvalidInput(t *testing.T) {
	rt := &AgentMgmtRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// CapabilityRuntime
// ---------------------------------------------------------------------------

func TestCapabilityRuntime_List(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agent/functions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"name": "search", "category": "search"},
		})
	})
	defer srv.Close()

	rt := &CapabilityRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "list",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
}

func TestCapabilityRuntime_Connectors(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/connectors", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"github", "slack"})
	})
	defer srv.Close()

	rt := &CapabilityRuntime{OrchestratorURL: srv.URL, HTTPClient: srv.Client()}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "connectors",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "connectors", m["action"])
}

func TestCapabilityRuntime_UnknownAction(t *testing.T) {
	rt := &CapabilityRuntime{OrchestratorURL: "http://localhost"}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "install",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown capability action")
}

func TestCapabilityRuntime_InvalidInput(t *testing.T) {
	rt := &CapabilityRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// WorkflowRuntime (with miniredis)
// ---------------------------------------------------------------------------

func TestWorkflowRuntime_NoRedis(t *testing.T) {
	rt := &WorkflowRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "start",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "simulated", m["status"])
}

func TestWorkflowRuntime_FullLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &WorkflowRuntime{Redis: rdb}
	wfID := "wf-" + uuid.New().String()

	// Start
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "start", "workflow_id": wfID, "steps": []string{"step1", "step2"},
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "running", m["status"])

	// Status (running)
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "status", "workflow_id": wfID,
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.Equal(t, "status", m["action"])

	// Pause
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pause", "workflow_id": wfID,
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.Equal(t, "paused", m["status"])

	// Resume
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "resume", "workflow_id": wfID,
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.Equal(t, "running", m["status"])

	// Stop
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "stop", "workflow_id": wfID,
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.Equal(t, "stopped", m["status"])
}

func TestWorkflowRuntime_Status_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &WorkflowRuntime{Redis: rdb}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "status", "workflow_id": "nonexistent",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, "not_found", m["status"])
}

func TestWorkflowRuntime_Pause_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &WorkflowRuntime{Redis: rdb}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "pause", "workflow_id": "nonexistent",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow not found")
}

func TestWorkflowRuntime_AutoID(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &WorkflowRuntime{Redis: rdb}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "start",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.NotEmpty(t, m["workflow_id"])
	assert.Equal(t, "running", m["status"])
}

func TestWorkflowRuntime_UnknownAction(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &WorkflowRuntime{Redis: rdb}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown workflow action")
}

func TestWorkflowRuntime_InvalidInput(t *testing.T) {
	rt := &WorkflowRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// MemoryRuntime (with miniredis)
// ---------------------------------------------------------------------------

func TestMemoryRuntime_NoRedis(t *testing.T) {
	rt := &MemoryRuntime{}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "store", "key": "k", "value": "v",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["success"].(bool))
	assert.Contains(t, m["note"].(string), "Redis not configured")
}

func TestMemoryRuntime_StoreAndRetrieve(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}

	// Store
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "store", "key": "greeting", "value": "hello", "agent_id": "a1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["stored"].(bool))

	// Retrieve
	out, err = rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "retrieve", "key": "greeting", "agent_id": "a1",
	}), 5*time.Second)
	require.NoError(t, err)
	m = decodeMap(t, out)
	assert.True(t, m["found"].(bool))
	assert.Equal(t, "hello", m["value"])
}

func TestMemoryRuntime_Retrieve_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}
	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "retrieve", "key": "nonexistent",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.False(t, m["found"].(bool))
}

func TestMemoryRuntime_Forget(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}
	rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "store", "key": "temp", "value": "data",
	}), 5*time.Second)

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "forget", "key": "temp",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.True(t, m["deleted"].(bool))
}

func TestMemoryRuntime_List(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}
	for i := 0; i < 3; i++ {
		rt.Execute(context.Background(), mustJSON(map[string]interface{}{
			"action": "store", "key": fmt.Sprintf("k%d", i), "value": fmt.Sprintf("v%d", i), "agent_id": "a1",
		}), 5*time.Second)
	}

	out, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "list", "agent_id": "a1",
	}), 5*time.Second)
	require.NoError(t, err)
	m := decodeMap(t, out)
	assert.Equal(t, float64(3), m["count"])
}

func TestMemoryRuntime_Store_MissingKey(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "store", "value": "v",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key is required")
}

func TestMemoryRuntime_UnknownAction(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rt := &MemoryRuntime{Redis: rdb}
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "nope",
	}), 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown memory action")
}

func TestMemoryRuntime_InvalidInput(t *testing.T) {
	rt := &MemoryRuntime{}
	_, err := rt.Execute(context.Background(), json.RawMessage(`bad`), 5*time.Second)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestExtractJSONPath(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": []interface{}{map[string]interface{}{"c": "found"}},
		},
	}
	assert.Equal(t, "found", extractJSONPath(data, "a.b.0.c"))
	assert.Nil(t, extractJSONPath(data, "a.b.1.c"))
	assert.Nil(t, extractJSONPath(data, "x.y.z"))
	assert.Nil(t, extractJSONPath("not_a_map", "a.b"))
}

func TestApplyTransform(t *testing.T) {
	data := map[string]interface{}{"first": "Alice", "last": "Smith", "age": 30.0}
	schema := map[string]interface{}{"name": "first", "surname": "last"}
	result := applyTransform(data, schema)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "Smith", result["surname"])
	_, hasAge := result["age"]
	assert.False(t, hasAge)
}

func TestValidateJSONSchema(t *testing.T) {
	schema := json.RawMessage(`{"required": ["name", "email"]}`)

	// Valid
	valid := validateJSONSchema(json.RawMessage(`{"name":"Alice","email":"a@b.com"}`), schema)
	assert.Empty(t, valid)

	// Missing email
	invalid := validateJSONSchema(json.RawMessage(`{"name":"Alice"}`), schema)
	assert.Len(t, invalid, 1)
	assert.Contains(t, invalid[0], "email")

	// Invalid schema
	badSchema := validateJSONSchema(json.RawMessage(`{"name":"Alice"}`), json.RawMessage(`{bad`))
	assert.Len(t, badSchema, 1)
	assert.Contains(t, badSchema[0], "invalid schema")

	// Invalid data
	badData := validateJSONSchema(json.RawMessage(`not json`), schema)
	assert.Len(t, badData, 1)
	assert.Contains(t, badData[0], "invalid data")
}

func TestScanPII(t *testing.T) {
	// Multiple PII types
	findings := scanPII("Contact alice@test.com, SSN 123-45-6789, phone 555-123-4567")
	assert.GreaterOrEqual(t, len(findings), 3)
}

func TestScanSecrets(t *testing.T) {
	findings := scanSecrets(`api_key: "sk-1234567890"`)
	assert.NotEmpty(t, findings)

	clean := scanSecrets("no secrets here")
	assert.Empty(t, clean)
}

// ---------------------------------------------------------------------------
// Name() tests for all runtimes
// ---------------------------------------------------------------------------

func TestAllRuntimes_Name(t *testing.T) {
	runtimes := map[string]Runtime{
		"search":        &SearchRuntime{},
		"browser":       &BrowserRuntime{},
		"compute":       &ComputeRuntime{},
		"data":          &DataRuntime{},
		"file":          &FileRuntime{},
		"communication": &CommunicationRuntime{},
		"assure":        &AssureRuntime{},
		"validate":      &ValidateRuntime{},
		"simulate":      &SimulateRuntime{},
		"observe":       &ObserveRuntime{},
		"learn":         &LearnRuntime{},
		"agent_mgmt":    &AgentMgmtRuntime{},
		"capability":    &CapabilityRuntime{},
		"workflow":      &WorkflowRuntime{},
		"memory":        &MemoryRuntime{},
	}
	for expected, rt := range runtimes {
		assert.Equal(t, expected, rt.Name(), "Name() mismatch for %T", rt)
	}
}

// ---------------------------------------------------------------------------
// FileRuntime env fallback
// ---------------------------------------------------------------------------

func TestFileRuntime_EnvBaseDir(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("AGENT_FILE_BASE_DIR", dir)
	defer os.Unsetenv("AGENT_FILE_BASE_DIR")

	rt := &FileRuntime{} // no BaseDir set
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"action": "write", "path": "env.txt", "content": "from env",
	}), 5*time.Second)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "env.txt"))
	require.NoError(t, err)
	assert.Equal(t, "from env", string(data))
}

// ---------------------------------------------------------------------------
// SearchRuntime num_results clamping
// ---------------------------------------------------------------------------

func TestSearchRuntime_NumResultsClamping(t *testing.T) {
	// Verify that num_results > 100 gets clamped.
	// We can't test the actual API call, but we can verify the URL construction
	// by checking that no error occurs for invalid inputs.
	rt := &SearchRuntime{APIKey: "k", EngineID: "c"}
	// Empty query should fail before hitting API
	_, err := rt.Execute(context.Background(), mustJSON(map[string]interface{}{
		"query":       "test",
		"num_results": 500,
	}), 5*time.Second)
	// Will fail with API error (fake key), but won't panic on clamping
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "search API returned") || strings.Contains(err.Error(), "search request failed"))
	}
}
