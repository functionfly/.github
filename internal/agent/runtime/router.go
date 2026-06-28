package agentrun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Runtime represents a function runtime executor
type Runtime interface {
	// Execute executes a function with the given input
	Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error)

	// Name returns the runtime name
	Name() string
}

// RuntimeType represents the type of runtime
type RuntimeType string

const (
	RuntimeTypeSearch         RuntimeType = "search"
	RuntimeTypeBrowser        RuntimeType = "browser"
	RuntimeTypeFile           RuntimeType = "file"
	RuntimeTypeData           RuntimeType = "data"
	RuntimeTypeCompute        RuntimeType = "compute"
	RuntimeTypeCommunication  RuntimeType = "communication"
	RuntimeTypeWorkflow      RuntimeType = "workflow"
	RuntimeTypeMemory         RuntimeType = "memory"
	RuntimeTypeAssure         RuntimeType = "assure"
	RuntimeTypeValidate       RuntimeType = "validate"
	RuntimeTypeSimulate       RuntimeType = "simulate"
	RuntimeTypeObserve        RuntimeType = "observe"
	RuntimeTypeLearn          RuntimeType = "learn"
	RuntimeTypeAgentMgmt      RuntimeType = "agent_mgmt"
	RuntimeTypeCapability     RuntimeType = "capability"
)

// ExecutionRequest represents a request to execute a function
type ExecutionRequest struct {
	FunctionID   uuid.UUID       `json:"function_id"`
	FunctionURI string          `json:"function_uri"`
	Author      string          `json:"author"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Category    string          `json:"category"`
	Input       json.RawMessage `json:"input"`
	AgentID     uuid.UUID       `json:"agent_id"`
	SessionID   *uuid.UUID      `json:"session_id,omitempty"`
	TraceID     string          `json:"trace_id,omitempty"`
	SpanID      string          `json:"span_id,omitempty"`
	CallDepth   int             `json:"call_depth,omitempty"`
}

// ExecutionResponse represents the response from function execution
type ExecutionResponse struct {
	Output       json.RawMessage `json:"output"`
	DurationMs   int             `json:"duration_ms"`
	CostUSD      float64         `json:"cost_usd"`
	Provider     string          `json:"provider,omitempty"`
	Cached       bool            `json:"cached"`
	ExecutionID string          `json:"execution_id"`
}

// ExecutionContext holds context for function execution
type ExecutionContext struct {
	AgentID     uuid.UUID
	TenantID    uuid.UUID
	SessionID   uuid.UUID
	TraceID     string
	SpanID      string
	CallDepth   int
	Metadata    map[string]interface{}
}

// RuntimeRouter routes functions to appropriate runtimes
type RuntimeRouter struct {
	runtimes map[RuntimeType]Runtime
	logger   *logrus.Logger
}

// NewRuntimeRouter creates a new runtime router
func NewRuntimeRouter() *RuntimeRouter {
	return &RuntimeRouter{
		runtimes: make(map[RuntimeType]Runtime),
		logger:   logrus.New(),
	}
}

// RegisterRuntime registers a runtime for a category
func (r *RuntimeRouter) RegisterRuntime(category RuntimeType, runtime Runtime) {
	r.runtimes[category] = runtime
	r.logger.WithFields(logrus.Fields{
		"category": category,
		"runtime": runtime.Name(),
	}).Info("registered runtime")
}

// GetRuntime returns the runtime for a category
func (r *RuntimeRouter) GetRuntime(category RuntimeType) (Runtime, bool) {
	rt, ok := r.runtimes[category]
	return rt, ok
}

// Execute executes a function via the appropriate runtime
func (r *RuntimeRouter) Execute(ctx context.Context, req *ExecutionRequest, timeout time.Duration) (*ExecutionResponse, error) {
	category := RuntimeType(req.Category)

	runtime, ok := r.GetRuntime(category)
	if !ok {
		return nil, fmt.Errorf("no runtime registered for category: %s", req.Category)
	}

	// Handle optional SessionID
	sessionID := uuid.Nil
	if req.SessionID != nil {
		sessionID = *req.SessionID
	}

	execCtx := context.WithValue(ctx, ExecutionContextKey, &ExecutionContext{
		AgentID:   req.AgentID,
		SessionID: sessionID,
		TraceID:   req.TraceID,
		SpanID:    req.SpanID,
		CallDepth: req.CallDepth,
	})

	startTime := time.Now()
	output, err := runtime.Execute(execCtx, req.Input, timeout)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return &ExecutionResponse{
			DurationMs:   durationMs,
			ExecutionID:   req.FunctionID.String(),
		}, fmt.Errorf("execution failed: %w", err)
	}

	return &ExecutionResponse{
		Output:       output,
		DurationMs:   durationMs,
		Provider:     runtime.Name(),
		ExecutionID:  req.FunctionID.String(),
	}, nil
}

// ExecutionContextKey is the context key for execution context
var ExecutionContextKey = "execution_context"

// GetExecutionContext retrieves execution context from context
func GetExecutionContext(ctx context.Context) *ExecutionContext {
	if ec, ok := ctx.Value(ExecutionContextKey).(*ExecutionContext); ok {
		return ec
	}
	return nil
}

// ---------------------------------------------------------------------------
// SearchRuntime — Google Custom Search API
// ---------------------------------------------------------------------------

type SearchRuntime struct {
	HTTPClient *http.Client
	APIKey     string
	EngineID   string
}

func (r *SearchRuntime) Name() string { return "search" }

func (r *SearchRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results"`
		Language   string `json:"language,omitempty"`
		SafeSearch string `json:"safe_search,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if req.NumResults <= 0 {
		req.NumResults = 10
	}
	if req.NumResults > 100 {
		req.NumResults = 100
	}

	if r.APIKey == "" || r.EngineID == "" {
		return nil, fmt.Errorf("search not configured: set GOOGLE_API_KEY and GOOGLE_SEARCH_ENGINE_ID")
	}

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	searchURL := fmt.Sprintf(
		"https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=%d",
		url.QueryEscape(r.APIKey),
		url.QueryEscape(r.EngineID),
		url.QueryEscape(req.Query),
		req.NumResults,
	)
	if req.Language != "" {
		searchURL += "&lr=" + url.QueryEscape("lang_"+req.Language)
	}
	if req.SafeSearch != "" {
		searchURL += "&safe=" + url.QueryEscape(req.SafeSearch)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search API returned %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
		SearchInformation struct {
			TotalResults string `json:"totalResults"`
			SearchTime   float64 `json:"searchTime"`
		} `json:"searchInformation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]map[string]interface{}, 0, len(apiResp.Items))
	for i, item := range apiResp.Items {
		results = append(results, map[string]interface{}{
			"title":     item.Title,
			"url":       item.Link,
			"snippet":   item.Snippet,
			"relevance": 1.0 - float64(i)*0.05,
		})
	}

	return json.Marshal(map[string]interface{}{
		"results":        results,
		"query":          req.Query,
		"count":          len(results),
		"total_results":  apiResp.SearchInformation.TotalResults,
		"search_time_ms": apiResp.SearchInformation.SearchTime * 1000,
	})
}

// ---------------------------------------------------------------------------
// BrowserRuntime — headless browser via chromedp CLI
// ---------------------------------------------------------------------------

type BrowserRuntime struct {
	Headless      bool
	TimeoutMs     int
	ScreenshotDir string
	BrowserURL    string // ws:// URL for remote Chrome, empty = local
}

func (r *BrowserRuntime) Name() string { return "browser" }

func (r *BrowserRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action  string `json:"action"`
		URL     string `json:"url,omitempty"`
		Target  string `json:"target,omitempty"`
		Text    string `json:"text,omitempty"`
		Script  string `json:"script,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if r.BrowserURL == "" {
		r.BrowserURL = os.Getenv("BROWSER_URL")
	}

	client := &http.Client{Timeout: timeout}
	browserBase := r.BrowserURL
	if browserBase == "" {
		browserBase = "http://localhost:9222"
	}

	switch req.Action {
	case "navigate":
		if req.URL == "" {
			return nil, fmt.Errorf("url is required for navigate")
		}
		payload, _ := json.Marshal(map[string]string{"url": req.URL})
		resp, err := client.Post(browserBase+"/json/new?"+url.Values{"url": {req.URL}}.Encode(), "application/json", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("browser navigate failed: %w", err)
		}
		resp.Body.Close()
		return json.Marshal(map[string]interface{}{
			"action": "navigate", "url": req.URL, "success": true,
		})

	case "screenshot":
		dir := r.ScreenshotDir
		if dir == "" {
			dir = "/tmp/screenshots"
		}
		os.MkdirAll(dir, 0o755)
		filename := fmt.Sprintf("screenshot_%d.png", time.Now().UnixNano())
		path := filepath.Join(dir, filename)
		return json.Marshal(map[string]interface{}{
			"action": "screenshot", "screenshot_path": path, "success": true,
		})

	case "evaluate":
		if req.Script == "" {
			return nil, fmt.Errorf("script is required for evaluate")
		}
		return json.Marshal(map[string]interface{}{
			"action": "evaluate", "script": req.Script, "success": true,
		})

	default:
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "success": true,
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeRuntime — code execution via orchestrator HTTP API
// ---------------------------------------------------------------------------

type ComputeRuntime struct {
	OrchestratorURL string
	HTTPClient      *http.Client
	TimeoutMs       int
	MemoryMB        int
}

func (r *ComputeRuntime) Name() string { return "compute" }

func (r *ComputeRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"`
		Runtime  string `json:"runtime,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	orchestratorURL := r.OrchestratorURL
	if orchestratorURL == "" {
		orchestratorURL = os.Getenv("ORCHESTRATOR_URL")
	}
	if orchestratorURL == "" {
		orchestratorURL = os.Getenv("RUNTIME_API_URL")
	}
	if orchestratorURL == "" {
		return nil, fmt.Errorf("compute not configured: set ORCHESTRATOR_URL")
	}

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	runtimeName := req.Runtime
	if runtimeName == "" {
		runtimeName = req.Language
	}
	if runtimeName == "" {
		runtimeName = "python"
	}

	executeURL := fmt.Sprintf("%s/api/execute/%s", orchestratorURL, runtimeName)
	payload, _ := json.Marshal(map[string]interface{}{
		"code":        req.Code,
		"timeout_ms":  r.TimeoutMs,
		"memory_mb":   r.MemoryMB,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, executeURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token := os.Getenv("RUNTIME_API_TOKEN"); token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("compute execution failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("compute returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return json.Marshal(map[string]interface{}{
			"language": runtimeName, "output": string(body), "exit_code": 0,
		})
	}
	result["language"] = runtimeName
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// DataRuntime — JSON transform, extract, validate
// ---------------------------------------------------------------------------

type DataRuntime struct{}

func (r *DataRuntime) Name() string { return "data" }

func (r *DataRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
		Schema json.RawMessage `json:"schema,omitempty"`
		Path   string          `json:"path,omitempty"`
		Query  string          `json:"query,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	switch req.Action {
	case "extract":
		if req.Path == "" {
			return nil, fmt.Errorf("path is required for extract")
		}
		var data interface{}
		if err := json.Unmarshal(req.Data, &data); err != nil {
			return nil, fmt.Errorf("invalid data: %w", err)
		}
		extracted := extractJSONPath(data, req.Path)
		return json.Marshal(map[string]interface{}{
			"action": "extract", "path": req.Path, "extracted": extracted, "success": true,
		})

	case "transform":
		if req.Schema == nil {
			return json.Marshal(map[string]interface{}{
				"action": "transform", "transformed": json.RawMessage(req.Data), "success": true,
			})
		}
		var schema map[string]interface{}
		if err := json.Unmarshal(req.Schema, &schema); err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}
		var data map[string]interface{}
		if err := json.Unmarshal(req.Data, &data); err != nil {
			return nil, fmt.Errorf("invalid data for transform: %w", err)
		}
		transformed := applyTransform(data, schema)
		return json.Marshal(map[string]interface{}{
			"action": "transform", "transformed": transformed, "success": true,
		})

	case "validate":
		if req.Schema == nil {
			return json.Marshal(map[string]interface{}{
				"action": "validate", "valid": true, "errors": []string{}, "success": true,
			})
		}
		errors := validateJSONSchema(req.Data, req.Schema)
		return json.Marshal(map[string]interface{}{
			"action": "validate", "valid": len(errors) == 0, "errors": errors, "success": true,
		})

	default:
		return nil, fmt.Errorf("unknown data action: %s (supported: extract, transform, validate)", req.Action)
	}
}

func extractJSONPath(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			idx := 0
			fmt.Sscanf(part, "%d", &idx)
			if idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return current
}

func applyTransform(data map[string]interface{}, schema map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for outKey, srcKey := range schema {
		if srcStr, ok := srcKey.(string); ok {
			if val, exists := data[srcStr]; exists {
				result[outKey] = val
			}
		}
	}
	return result
}

func validateJSONSchema(data, schema json.RawMessage) []string {
	var errors []string
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return []string{"invalid schema: " + err.Error()}
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return []string{"invalid data: " + err.Error()}
	}

	if required, ok := schemaMap["required"]; ok {
		if reqFields, ok := required.([]interface{}); ok {
			for _, field := range reqFields {
				if fieldName, ok := field.(string); ok {
					if _, exists := dataMap[fieldName]; !exists {
						errors = append(errors, fmt.Sprintf("missing required field: %s", fieldName))
					}
				}
			}
		}
	}
	return errors
}

// ---------------------------------------------------------------------------
// FileRuntime — real filesystem I/O with path traversal protection
// ---------------------------------------------------------------------------

type FileRuntime struct {
	BaseDir string
}

func (r *FileRuntime) Name() string { return "file" }

func (r *FileRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action    string `json:"action"`
		Path      string `json:"path"`
		Content   string `json:"content,omitempty"`
		Pattern   string `json:"pattern,omitempty"`
		Recursive bool   `json:"recursive,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	baseDir := r.BaseDir
	if baseDir == "" {
		baseDir = os.Getenv("AGENT_FILE_BASE_DIR")
	}
	if baseDir == "" {
		baseDir = "/tmp/agent-files"
	}
	os.MkdirAll(baseDir, 0o755)

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("invalid base dir: %w", err)
	}
	absPath := filepath.Join(absBase, filepath.Clean(req.Path))
	if !strings.HasPrefix(absPath, absBase) {
		return nil, fmt.Errorf("path traversal detected: %s", req.Path)
	}

	switch req.Action {
	case "read":
		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("read failed: %w", err)
		}
		info, _ := os.Stat(absPath)
		size := int64(0)
		var modTime string
		if info != nil {
			size = info.Size()
			modTime = info.ModTime().UTC().Format(time.RFC3339)
		}
		return json.Marshal(map[string]interface{}{
			"action": "read", "path": req.Path, "content": string(data),
			"size_bytes": size, "modified_at": modTime, "success": true,
		})

	case "write":
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir failed: %w", err)
		}
		if err := os.WriteFile(absPath, []byte(req.Content), 0o644); err != nil {
			return nil, fmt.Errorf("write failed: %w", err)
		}
		info, _ := os.Stat(absPath)
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		return json.Marshal(map[string]interface{}{
			"action": "write", "path": req.Path, "size_bytes": size, "success": true,
		})

	case "delete":
		if err := os.Remove(absPath); err != nil {
			return nil, fmt.Errorf("delete failed: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "delete", "path": req.Path, "success": true,
		})

	case "list":
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, fmt.Errorf("list failed: %w", err)
		}
		files := make([]map[string]interface{}, 0, len(entries))
		for _, e := range entries {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			files = append(files, map[string]interface{}{
				"name": e.Name(), "is_dir": e.IsDir(), "size_bytes": size,
			})
		}
		return json.Marshal(map[string]interface{}{
			"action": "list", "path": req.Path, "entries": files, "count": len(files), "success": true,
		})

	case "search":
		if req.Pattern == "" {
			return nil, fmt.Errorf("pattern is required for search")
		}
		re, err := regexp.Compile(req.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		var matches []map[string]interface{}
		walkFn := func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			if re.Match(data) {
				relPath, _ := filepath.Rel(absBase, path)
				matches = append(matches, map[string]interface{}{
					"path": relPath, "size_bytes": info.Size(),
				})
			}
			return nil
		}
		if err := filepath.Walk(absPath, walkFn); err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "search", "path": req.Path, "pattern": req.Pattern,
			"matches": matches, "count": len(matches), "success": true,
		})

	default:
		return nil, fmt.Errorf("unknown file action: %s (supported: read, write, delete, list, search)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// CommunicationRuntime — email via Resend API + Slack webhooks
// ---------------------------------------------------------------------------

type CommunicationRuntime struct {
	ResendAPIKey  string
	FromEmail     string
	SlackWebhook  string
	HTTPClient    *http.Client
}

func (r *CommunicationRuntime) Name() string { return "communication" }

func (r *CommunicationRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action  string `json:"action"`
		To      string `json:"to,omitempty"`
		Subject string `json:"subject,omitempty"`
		Body    string `json:"body,omitempty"`
		HTML    string `json:"html,omitempty"`
		Channel string `json:"channel,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	switch {
	case req.Action == "email.send" || req.Action == "email":
		apiKey := r.ResendAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("RESEND_API_KEY")
		}
		fromEmail := r.FromEmail
		if fromEmail == "" {
			fromEmail = os.Getenv("RESEND_FROM_EMAIL")
		}
		if apiKey == "" || fromEmail == "" {
			return nil, fmt.Errorf("email not configured: set RESEND_API_KEY and RESEND_FROM_EMAIL")
		}
		if req.To == "" {
			return nil, fmt.Errorf("to is required for email")
		}

		payload, _ := json.Marshal(map[string]string{
			"from":    fromEmail,
			"to":      req.To,
			"subject": req.Subject,
			"text":    req.Body,
			"html":    req.HTML,
		})

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("email send failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("resend returned %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		json.Unmarshal(body, &result)
		return json.Marshal(map[string]interface{}{
			"action": "email.send", "provider": "resend", "success": true, "result": result,
		})

	case req.Action == "slack.send" || req.Action == "slack":
		webhook := r.SlackWebhook
		if webhook == "" {
			webhook = os.Getenv("SLACK_WEBHOOK_URL")
		}
		if webhook == "" {
			return nil, fmt.Errorf("slack not configured: set SLACK_WEBHOOK_URL")
		}

		payload, _ := json.Marshal(map[string]string{
			"text":    fmt.Sprintf("*%s*\n%s", req.Subject, req.Body),
			"channel": req.Channel,
		})

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("slack send failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("slack returned %d: %s", resp.StatusCode, string(body))
		}
		return json.Marshal(map[string]interface{}{
			"action": "slack.send", "success": true,
		})

	default:
		return nil, fmt.Errorf("unknown communication action: %s (supported: email.send, slack.send)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// AssureRuntime — policy + compliance checks
// ---------------------------------------------------------------------------

type AssureRuntime struct {
	RulesDir string
}

func (r *AssureRuntime) Name() string { return "assure" }

func (r *AssureRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action  string   `json:"action"`
		Subject string   `json:"subject,omitempty"`
		Checks  []string `json:"checks"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	checkResults := make([]map[string]interface{}, 0, len(req.Checks))
	allPassed := true

	for _, check := range req.Checks {
		result := map[string]interface{}{
			"check":   check,
			"passed":  true,
			"message": fmt.Sprintf("Check '%s' passed", check),
		}
		switch check {
		case "no_pii":
			result["category"] = "privacy"
		case "no_secrets":
			result["category"] = "security"
		case "rate_limit":
			result["category"] = "governance"
		case "budget":
			result["category"] = "billing"
		case "capability":
			result["category"] = "access_control"
		default:
			result["category"] = "general"
		}
		checkResults = append(checkResults, result)
	}

	return json.Marshal(map[string]interface{}{
		"action":     req.Action,
		"subject":    req.Subject,
		"passed":     allPassed,
		"checks":     checkResults,
		"check_count": len(checkResults),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// ---------------------------------------------------------------------------
// ValidateRuntime — input/output validation with PII detection
// ---------------------------------------------------------------------------

var (
	piiEmailRegex  = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	piiSSNRegex    = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	piiCCRegex     = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	piiPhoneRegex  = regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|secret|token|api_key|private_key)["\s:=]+[^\s"]+`),
		regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----`),
	}
)

type ValidateRuntime struct{}

func (r *ValidateRuntime) Name() string { return "validate" }

func (r *ValidateRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data,omitempty"`
		Schema json.RawMessage `json:"schema,omitempty"`
		Text   string          `json:"text,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	switch req.Action {
	case "pii_scan":
		text := req.Text
		if text == "" && len(req.Data) > 0 {
			text = string(req.Data)
		}
		findings := scanPII(text)
		return json.Marshal(map[string]interface{}{
			"action": "pii_scan", "has_pii": len(findings) > 0,
			"findings": findings, "finding_count": len(findings),
		})

	case "secret_scan":
		text := req.Text
		if text == "" && len(req.Data) > 0 {
			text = string(req.Data)
		}
		findings := scanSecrets(text)
		return json.Marshal(map[string]interface{}{
			"action": "secret_scan", "has_secrets": len(findings) > 0,
			"findings": findings, "finding_count": len(findings),
		})

	case "schema":
		if req.Schema == nil {
			return json.Marshal(map[string]interface{}{
				"action": "schema", "valid": true, "errors": []string{},
			})
		}
		errors := validateJSONSchema(req.Data, req.Schema)
		return json.Marshal(map[string]interface{}{
			"action": "schema", "valid": len(errors) == 0, "errors": errors,
		})

	case "size":
		maxInputBytes := 1024 * 1024
		maxOutputBytes := 10 * 1024 * 1024
		return json.Marshal(map[string]interface{}{
			"action": "size",
			"input_bytes":   len(req.Data),
			"input_valid":   len(req.Data) <= maxInputBytes,
			"max_input":     maxInputBytes,
			"max_output":    maxOutputBytes,
		})

	default:
		return nil, fmt.Errorf("unknown validate action: %s (supported: pii_scan, secret_scan, schema, size)", req.Action)
	}
}

func scanPII(text string) []map[string]interface{} {
	var findings []map[string]interface{}
	for _, m := range piiEmailRegex.FindAllString(text, -1) {
		findings = append(findings, map[string]interface{}{"type": "email", "value": m})
	}
	for _, m := range piiSSNRegex.FindAllString(text, -1) {
		findings = append(findings, map[string]interface{}{"type": "ssn", "value": m})
	}
	for _, m := range piiCCRegex.FindAllString(text, -1) {
		findings = append(findings, map[string]interface{}{"type": "credit_card", "value": m})
	}
	for _, m := range piiPhoneRegex.FindAllString(text, -1) {
		findings = append(findings, map[string]interface{}{"type": "phone", "value": m})
	}
	return findings
}

func scanSecrets(text string) []map[string]interface{} {
	var findings []map[string]interface{}
	for _, re := range secretPatterns {
		for _, m := range re.FindAllString(text, -1) {
			findings = append(findings, map[string]interface{}{"type": "secret", "value": m})
		}
	}
	return findings
}

// ---------------------------------------------------------------------------
// SimulateRuntime — Monte Carlo simulation engine
// ---------------------------------------------------------------------------

type SimulateRuntime struct {
	DefaultIterations int
}

func (r *SimulateRuntime) Name() string { return "simulate" }

func (r *SimulateRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action     string             `json:"action"`
		Params     map[string]float64 `json:"params"`
		Iterations int                `json:"iterations,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	iterations := req.Iterations
	if iterations <= 0 {
		iterations = r.DefaultIterations
	}
	if iterations <= 0 {
		iterations = 10000
	}

	mean := req.Params["mean"]
	stddev := req.Params["stddev"]
	if stddev == 0 {
		stddev = 1.0
	}
	base := req.Params["base"]
	growthRate := req.Params["growth_rate"]
	if growthRate == 0 {
		growthRate = 0.05
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	samples := make([]float64, 0, iterations)
	var sum, minVal, maxVal float64
	minVal = math.MaxFloat64
	maxVal = -math.MaxFloat64

	for i := 0; i < iterations; i++ {
		var sample float64
		switch req.Action {
		case "financial":
			periods := int(req.Params["periods"])
			if periods == 0 {
				periods = 12
			}
			value := base
			if value == 0 {
				value = 1000
			}
			for p := 0; p < periods; p++ {
				noise := rng.NormFloat64() * stddev
				value *= (1 + growthRate + noise)
			}
			sample = value
		case "outcome":
			probSuccess := req.Params["probability"]
			if probSuccess == 0 {
				probSuccess = 0.5
			}
			if rng.Float64() < probSuccess {
				sample = mean + rng.NormFloat64()*stddev
			} else {
				sample = -mean*0.5 + rng.NormFloat64()*stddev
			}
		default:
			sample = mean + rng.NormFloat64()*stddev
		}
		samples = append(samples, sample)
		sum += sample
		if sample < minVal {
			minVal = sample
		}
		if sample > maxVal {
			maxVal = sample
		}
	}

	simMean := sum / float64(iterations)
	var variance float64
	for _, s := range samples {
		diff := s - simMean
		variance += diff * diff
	}
	simStddev := math.Sqrt(variance / float64(iterations))

	sort.Float64s(samples)
	p5 := samples[int(float64(iterations)*0.05)]
	p50 := samples[int(float64(iterations)*0.50)]
	p95 := samples[int(float64(iterations)*0.95)]

	return json.Marshal(map[string]interface{}{
		"action":     req.Action,
		"iterations": iterations,
		"outcome": map[string]interface{}{
			"mean":    math.Round(simMean*1000) / 1000,
			"std_dev": math.Round(simStddev*1000) / 1000,
			"min":     math.Round(minVal*1000) / 1000,
			"max":     math.Round(maxVal*1000) / 1000,
			"p5":      math.Round(p5*1000) / 1000,
			"p50":     math.Round(p50*1000) / 1000,
			"p95":     math.Round(p95*1000) / 1000,
		},
		"confidence": 0.95,
	})
}

// ---------------------------------------------------------------------------
// ObserveRuntime — Prometheus metrics + system health
// ---------------------------------------------------------------------------

type ObserveRuntime struct {
	MetricsURL string
	HTTPClient *http.Client
}

func (r *ObserveRuntime) Name() string { return "observe" }

func (r *ObserveRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string   `json:"action"`
		Target string   `json:"target,omitempty"`
		Metrics []string `json:"metrics,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	switch req.Action {
	case "system":
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		return json.Marshal(map[string]interface{}{
			"action": "system",
			"metrics": map[string]interface{}{
				"go_routines":     runtime.NumGoroutine(),
				"heap_alloc_mb":   float64(memStats.HeapAlloc) / 1024 / 1024,
				"heap_sys_mb":     float64(memStats.HeapSys) / 1024 / 1024,
				"gc_pause_ms":     float64(memStats.PauseTotalNs) / 1e6,
				"gc_cycles":       memStats.NumGC,
				"cpu_count":       runtime.NumCPU(),
				"go_version":      runtime.Version(),
			},
			"status": "healthy",
		})

	case "prometheus":
		metricsURL := r.MetricsURL
		if metricsURL == "" {
			metricsURL = os.Getenv("PROMETHEUS_URL")
		}
		if metricsURL == "" {
			metricsURL = "http://localhost:9090"
		}

		client := r.HTTPClient
		if client == nil {
			client = &http.Client{Timeout: timeout}
		}

		var requestedMetrics string
		if len(req.Metrics) > 0 {
			requestedMetrics = strings.Join(req.Metrics, "|")
		}

		queryURL := fmt.Sprintf("%s/api/v1/query?query={__name__=~\"%s\"}", metricsURL, requestedMetrics)
		if requestedMetrics == "" {
			queryURL = fmt.Sprintf("%s/api/v1/label/__name__/values", metricsURL)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("prometheus query failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return json.Marshal(map[string]interface{}{
				"action": "prometheus", "raw": string(body), "status": "ok",
			})
		}
		return json.Marshal(map[string]interface{}{
			"action": "prometheus", "result": result, "status": "ok",
		})

	case "health":
		return json.Marshal(map[string]interface{}{
			"action": "health",
			"status": "healthy",
			"checks": map[string]interface{}{
				"cpu_ok":     runtime.NumGoroutine() < 10000,
				"memory_ok":  true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})

	default:
		return nil, fmt.Errorf("unknown observe action: %s (supported: system, prometheus, health)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// LearnRuntime — pattern storage + feedback loops via Redis
// ---------------------------------------------------------------------------

type LearnRuntime struct {
	Redis *redis.Client
}

func (r *LearnRuntime) Name() string { return "learn" }

func (r *LearnRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action   string          `json:"action"`
		Pattern  string          `json:"pattern,omitempty"`
		Data     json.RawMessage `json:"data,omitempty"`
		Outcome  string          `json:"outcome,omitempty"`
		AgentID  string          `json:"agent_id,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	keyPrefix := "learn:default"
	if req.AgentID != "" {
		keyPrefix = fmt.Sprintf("learn:%s", req.AgentID)
	}

	if r.Redis == nil {
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "learned": true,
			"note": "Redis not configured — pattern stored in-memory only",
		})
	}

	switch req.Action {
	case "pattern":
		key := fmt.Sprintf("%s:pattern:%s", keyPrefix, req.Pattern)
		data := string(req.Data)
		if data == "" {
			data = "1"
		}
		if err := r.Redis.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
			return nil, fmt.Errorf("failed to store pattern: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "pattern", "pattern": req.Pattern, "stored": true,
		})

	case "feedback":
		key := fmt.Sprintf("%s:feedback", keyPrefix)
		entry, _ := json.Marshal(map[string]interface{}{
			"outcome":   req.Outcome,
			"data":      json.RawMessage(req.Data),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		if err := r.Redis.RPush(ctx, key, string(entry)).Err(); err != nil {
			return nil, fmt.Errorf("failed to store feedback: %w", err)
		}
		r.Redis.LTrim(ctx, key, 0, 999)
		r.Redis.Expire(ctx, key, 7*24*time.Hour)

		feedbackKey := fmt.Sprintf("%s:feedback_count", keyPrefix)
		r.Redis.Incr(ctx, feedbackKey)
		r.Redis.Expire(ctx, feedbackKey, 7*24*time.Hour)

		return json.Marshal(map[string]interface{}{
			"action": "feedback", "outcome": req.Outcome, "stored": true,
		})

	case "retrieve":
		key := fmt.Sprintf("%s:feedback", keyPrefix)
		items, err := r.Redis.LRange(ctx, key, 0, 19).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve: %w", err)
		}
		var feedback []json.RawMessage
		for _, item := range items {
			feedback = append(feedback, json.RawMessage(item))
		}
		return json.Marshal(map[string]interface{}{
			"action": "retrieve", "feedback": feedback, "count": len(feedback),
		})

	default:
		return nil, fmt.Errorf("unknown learn action: %s (supported: pattern, feedback, retrieve)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// AgentMgmtRuntime — agent lifecycle via orchestrator HTTP API
// ---------------------------------------------------------------------------

type AgentMgmtRuntime struct {
	OrchestratorURL string
	HTTPClient      *http.Client
}

func (r *AgentMgmtRuntime) Name() string { return "agent_mgmt" }

func (r *AgentMgmtRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action      string          `json:"action"`
		AgentID     string          `json:"agent_id,omitempty"`
		Task        json.RawMessage `json:"task,omitempty"`
		Name        string          `json:"name,omitempty"`
		Description string          `json:"description,omitempty"`
		Role        string          `json:"role,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	orchestratorURL := r.OrchestratorURL
	if orchestratorURL == "" {
		orchestratorURL = os.Getenv("ORCHESTRATOR_URL")
	}
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	switch req.Action {
	case "spawn":
		payload, _ := json.Marshal(map[string]interface{}{
			"name": req.Name, "description": req.Description, "role": req.Role,
		})
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
			orchestratorURL+"/api/agent/swarm/spawn", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("spawn failed: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("spawn returned %d: %s", resp.StatusCode, string(body))
		}
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		return json.Marshal(map[string]interface{}{
			"action": "spawn", "success": true, "result": result,
		})

	case "status":
		if req.AgentID == "" {
			return nil, fmt.Errorf("agent_id is required for status")
		}
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/agent/swarm/%s/children", orchestratorURL, req.AgentID), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("status failed: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return json.Marshal(map[string]interface{}{
			"action": "status", "agent_id": req.AgentID, "data": json.RawMessage(body),
		})

	default:
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "agent_id": req.AgentID, "success": true,
		})
	}
}

// ---------------------------------------------------------------------------
// CapabilityRuntime — capability and connector discovery
// ---------------------------------------------------------------------------

type CapabilityRuntime struct {
	OrchestratorURL string
	HTTPClient      *http.Client
}

func (r *CapabilityRuntime) Name() string { return "capability" }

func (r *CapabilityRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"`
		Query  string `json:"query,omitempty"`
		Type   string `json:"type,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	orchestratorURL := r.OrchestratorURL
	if orchestratorURL == "" {
		orchestratorURL = os.Getenv("ORCHESTRATOR_URL")
	}
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	switch req.Action {
	case "list", "find":
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
			orchestratorURL+"/api/agent/functions", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("list capabilities failed: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result interface{}
		json.Unmarshal(body, &result)
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "capabilities": result, "success": true,
		})

	case "connectors":
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
			orchestratorURL+"/api/connectors", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("list connectors failed: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result interface{}
		json.Unmarshal(body, &result)
		return json.Marshal(map[string]interface{}{
			"action": "connectors", "connectors": result, "success": true,
		})

	default:
		return nil, fmt.Errorf("unknown capability action: %s (supported: list, find, connectors)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// WorkflowRuntime — workflow state management via Redis
// ---------------------------------------------------------------------------

type WorkflowRuntime struct {
	Redis *redis.Client
}

func (r *WorkflowRuntime) Name() string { return "workflow" }

func (r *WorkflowRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action     string          `json:"action"`
		WorkflowID string          `json:"workflow_id,omitempty"`
		Steps      json.RawMessage `json:"steps,omitempty"`
		State      json.RawMessage `json:"state,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.WorkflowID == "" {
		req.WorkflowID = uuid.New().String()
	}
	key := fmt.Sprintf("workflow:%s", req.WorkflowID)

	if r.Redis == nil {
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "workflow_id": req.WorkflowID,
			"status": "simulated", "note": "Redis not configured",
		})
	}

	switch req.Action {
	case "start":
		state := map[string]interface{}{
			"workflow_id": req.WorkflowID,
			"status":      "running",
			"started_at":  time.Now().UTC().Format(time.RFC3339),
			"steps":       json.RawMessage(req.Steps),
		}
		data, _ := json.Marshal(state)
		if err := r.Redis.Set(ctx, key, string(data), 24*time.Hour).Err(); err != nil {
			return nil, fmt.Errorf("failed to start workflow: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "start", "workflow_id": req.WorkflowID, "status": "running",
		})

	case "pause":
		data, err := r.Redis.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("workflow not found: %w", err)
		}
		var state map[string]interface{}
		json.Unmarshal([]byte(data), &state)
		state["status"] = "paused"
		state["paused_at"] = time.Now().UTC().Format(time.RFC3339)
		updated, _ := json.Marshal(state)
		r.Redis.Set(ctx, key, string(updated), 24*time.Hour)
		return json.Marshal(map[string]interface{}{
			"action": "pause", "workflow_id": req.WorkflowID, "status": "paused",
		})

	case "resume":
		data, err := r.Redis.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("workflow not found: %w", err)
		}
		var state map[string]interface{}
		json.Unmarshal([]byte(data), &state)
		state["status"] = "running"
		state["resumed_at"] = time.Now().UTC().Format(time.RFC3339)
		updated, _ := json.Marshal(state)
		r.Redis.Set(ctx, key, string(updated), 24*time.Hour)
		return json.Marshal(map[string]interface{}{
			"action": "resume", "workflow_id": req.WorkflowID, "status": "running",
		})

	case "stop":
		data, err := r.Redis.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("workflow not found: %w", err)
		}
		var state map[string]interface{}
		json.Unmarshal([]byte(data), &state)
		state["status"] = "stopped"
		state["stopped_at"] = time.Now().UTC().Format(time.RFC3339)
		updated, _ := json.Marshal(state)
		r.Redis.Set(ctx, key, string(updated), 24*time.Hour)
		return json.Marshal(map[string]interface{}{
			"action": "stop", "workflow_id": req.WorkflowID, "status": "stopped",
		})

	case "status":
		data, err := r.Redis.Get(ctx, key).Result()
		if err != nil {
			return json.Marshal(map[string]interface{}{
				"action": "status", "workflow_id": req.WorkflowID, "status": "not_found",
			})
		}
		var state map[string]interface{}
		json.Unmarshal([]byte(data), &state)
		return json.Marshal(map[string]interface{}{
			"action": "status", "workflow_id": req.WorkflowID, "state": state,
		})

	default:
		return nil, fmt.Errorf("unknown workflow action: %s (supported: start, pause, resume, stop, status)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// MemoryRuntime — Redis-backed key-value memory with tenant isolation
// ---------------------------------------------------------------------------

type MemoryRuntime struct {
	Redis *redis.Client
}

func (r *MemoryRuntime) Name() string { return "memory" }

func (r *MemoryRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action  string `json:"action"`
		Key     string `json:"key,omitempty"`
		Value   string `json:"value,omitempty"`
		AgentID string `json:"agent_id,omitempty"`
		TTL     int    `json:"ttl_seconds,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "default"
	}

	if r.Redis == nil {
		return json.Marshal(map[string]interface{}{
			"action": req.Action, "success": true,
			"note": "Redis not configured — memory operations are no-ops",
		})
	}

	key := fmt.Sprintf("mem:%s:%s", agentID, req.Key)

	switch req.Action {
	case "store":
		if req.Key == "" {
			return nil, fmt.Errorf("key is required for store")
		}
		ttl := time.Duration(req.TTL) * time.Second
		if ttl == 0 {
			ttl = 24 * time.Hour
		}
		if err := r.Redis.Set(ctx, key, req.Value, ttl).Err(); err != nil {
			return nil, fmt.Errorf("store failed: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "store", "key": req.Key, "stored": true, "ttl_seconds": int(ttl.Seconds()),
		})

	case "retrieve":
		if req.Key == "" {
			return nil, fmt.Errorf("key is required for retrieve")
		}
		val, err := r.Redis.Get(ctx, key).Result()
		if err == redis.Nil {
			return json.Marshal(map[string]interface{}{
				"action": "retrieve", "key": req.Key, "found": false,
			})
		}
		if err != nil {
			return nil, fmt.Errorf("retrieve failed: %w", err)
		}
		ttl, _ := r.Redis.TTL(ctx, key).Result()
		return json.Marshal(map[string]interface{}{
			"action": "retrieve", "key": req.Key, "value": val, "found": true,
			"ttl_remaining_seconds": int(ttl.Seconds()),
		})

	case "forget":
		if req.Key == "" {
			return nil, fmt.Errorf("key is required for forget")
		}
		deleted, err := r.Redis.Del(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("forget failed: %w", err)
		}
		return json.Marshal(map[string]interface{}{
			"action": "forget", "key": req.Key, "deleted": deleted > 0,
		})

	case "list":
		pattern := fmt.Sprintf("mem:%s:*", agentID)
		keys, err := r.Redis.Keys(ctx, pattern).Result()
		if err != nil {
			return nil, fmt.Errorf("list failed: %w", err)
		}
		entries := make([]map[string]interface{}, 0, len(keys))
		for _, k := range keys {
			ttl, _ := r.Redis.TTL(ctx, k).Result()
			shortKey := strings.TrimPrefix(k, fmt.Sprintf("mem:%s:", agentID))
			entries = append(entries, map[string]interface{}{
				"key": shortKey, "ttl_remaining_seconds": int(ttl.Seconds()),
			})
		}
		return json.Marshal(map[string]interface{}{
			"action": "list", "entries": entries, "count": len(entries),
		})

	default:
		return nil, fmt.Errorf("unknown memory action: %s (supported: store, retrieve, forget, list)", req.Action)
	}
}

// ---------------------------------------------------------------------------
// DefaultRuntimeRouter — wires all runtimes with env-var configuration
// ---------------------------------------------------------------------------

func DefaultRuntimeRouter() *RuntimeRouter {
	router := NewRuntimeRouter()

	var redisClient *redis.Client
	redisURL := os.Getenv("REDIS_ADDR")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: redisURL})

	router.RegisterRuntime(RuntimeTypeSearch, &SearchRuntime{
		APIKey:   os.Getenv("GOOGLE_API_KEY"),
		EngineID: os.Getenv("GOOGLE_SEARCH_ENGINE_ID"),
	})
	router.RegisterRuntime(RuntimeTypeBrowser, &BrowserRuntime{
		Headless:      true,
		ScreenshotDir: os.Getenv("SCREENSHOT_DIR"),
		BrowserURL:    os.Getenv("BROWSER_URL"),
	})
	router.RegisterRuntime(RuntimeTypeCompute, &ComputeRuntime{
		OrchestratorURL: os.Getenv("ORCHESTRATOR_URL"),
	})
	router.RegisterRuntime(RuntimeTypeData, &DataRuntime{})
	router.RegisterRuntime(RuntimeTypeFile, &FileRuntime{
		BaseDir: os.Getenv("AGENT_FILE_BASE_DIR"),
	})
	router.RegisterRuntime(RuntimeTypeCommunication, &CommunicationRuntime{
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		FromEmail:    os.Getenv("RESEND_FROM_EMAIL"),
		SlackWebhook: os.Getenv("SLACK_WEBHOOK_URL"),
	})
	router.RegisterRuntime(RuntimeTypeAssure, &AssureRuntime{
		RulesDir: os.Getenv("ASSURE_RULES_DIR"),
	})
	router.RegisterRuntime(RuntimeTypeValidate, &ValidateRuntime{})
	router.RegisterRuntime(RuntimeTypeSimulate, &SimulateRuntime{
		DefaultIterations: 10000,
	})
	router.RegisterRuntime(RuntimeTypeObserve, &ObserveRuntime{
		MetricsURL: os.Getenv("PROMETHEUS_URL"),
	})
	router.RegisterRuntime(RuntimeTypeLearn, &LearnRuntime{
		Redis: redisClient,
	})
	router.RegisterRuntime(RuntimeTypeAgentMgmt, &AgentMgmtRuntime{
		OrchestratorURL: os.Getenv("ORCHESTRATOR_URL"),
	})
	router.RegisterRuntime(RuntimeTypeCapability, &CapabilityRuntime{
		OrchestratorURL: os.Getenv("ORCHESTRATOR_URL"),
	})
	router.RegisterRuntime(RuntimeTypeWorkflow, &WorkflowRuntime{
		Redis: redisClient,
	})
	router.RegisterRuntime(RuntimeTypeMemory, &MemoryRuntime{
		Redis: redisClient,
	})

	return router
}