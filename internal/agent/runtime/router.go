package agentrun

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// SearchRuntime implements search function execution
type SearchRuntime struct {
	// Configuration for search providers
	BaseURL   string
	APIKey    string
	TimeoutMs int
}

// Name returns the runtime name
func (r *SearchRuntime) Name() string {
	return "search"
}

// Execute executes a search function
func (r *SearchRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.NumResults == 0 {
		req.NumResults = 10
	}

	// Simulate search execution
	results := []map[string]interface{}{
		{
			"title":       "Search Result 1",
			"url":         "https://example.com/1",
			"snippet":     "This is a sample search result",
			"relevance":   0.95,
		},
		{
			"title":       "Search Result 2",
			"url":         "https://example.com/2",
			"snippet":     "Another sample result",
			"relevance":   0.85,
		},
	}

	response := map[string]interface{}{
		"results": results,
		"query":   req.Query,
		"count":   len(results),
	}

	return json.Marshal(response)
}

// BrowserRuntime implements browser function execution
type BrowserRuntime struct {
	// Configuration for headless browser
	Headless    bool
	TimeoutMs   int
	ScreenshotDir string
}

// Name returns the runtime name
func (r *BrowserRuntime) Name() string {
	return "browser"
}

// Execute executes a browser function
func (r *BrowserRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // navigate, snapshot, click, type, screenshot, extract
		URL    string `json:"url"`
		Target string `json:"target,omitempty"`
		Text   string `json:"text,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"success": true,
	}

	switch req.Action {
	case "navigate":
		response["url"] = req.URL
		response["final_url"] = req.URL // Would follow redirects
	case "snapshot":
		response["snapshot"] = map[string]interface{}{
			"role": "main",
			"name": "Page",
		}
	case "screenshot":
		response["screenshot_url"] = "/screenshots/session_123.png"
	}

	return json.Marshal(response)
}

// ComputeRuntime implements compute function execution
type ComputeRuntime struct {
	// Configuration for code execution
	TimeoutMs int
	MemoryMB  int
}

// Name returns the runtime name
func (r *ComputeRuntime) Name() string {
	return "compute"
}

// Execute executes a compute function
func (r *ComputeRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"` // python, javascript
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Simulate code execution
	response := map[string]interface{}{
		"language": req.Language,
		"output":   "Executed successfully",
		"exit_code": 0,
	}

	return json.Marshal(response)
}

// DataRuntime implements data function execution
type DataRuntime struct{}

// Name returns the runtime name
func (r *DataRuntime) Name() string {
	return "data"
}

// Execute executes a data function
func (r *DataRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // extract, transform, validate
		Data   interface{} `json:"data"`
		Schema json.RawMessage `json:"schema,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"success": true,
	}

	switch req.Action {
	case "extract":
		response["extracted"] = map[string]interface{}{"key": "value"}
	case "transform":
		response["transformed"] = req.Data
	case "validate":
		response["valid"] = true
	}

	return json.Marshal(response)
}

// FileRuntime implements file function execution
type FileRuntime struct {
	// Base directory for file operations
	BaseDir string
}

// Name returns the runtime name
func (r *FileRuntime) Name() string {
	return "file"
}

// Execute executes a file function
func (r *FileRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // read, write, edit, delete, search
		Path   string `json:"path"`
		Content string `json:"content,omitempty"`
		Pattern string `json:"pattern,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"path":   req.Path,
		"success": true,
	}

	return json.Marshal(response)
}

// CommunicationRuntime implements communication function execution
type CommunicationRuntime struct {
	// Email/SMS/Slack provider configuration
	EmailProvider string
	SMSProvider   string
	SlackToken    string
}

// Name returns the runtime name
func (r *CommunicationRuntime) Name() string {
	return "communication"
}

// Execute executes a communication function
func (r *CommunicationRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action    string `json:"action"` // email.send, sms.send, slack.send, calendar.schedule
		To        string `json:"to,omitempty"`
		Subject   string `json:"subject,omitempty"`
		Body      string `json:"body,omitempty"`
		Channel   string `json:"channel,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"success": true,
		"message_id": fmt.Sprintf("msg_%d", time.Now().UnixNano()),
	}

	return json.Marshal(response)
}

// AssureRuntime implements assurance function execution
type AssureRuntime struct {
	// Compliance engine configuration
	RulesDir string
}

// Name returns the runtime name
func (r *AssureRuntime) Name() string {
	return "assure"
}

// Execute executes an assure function
func (r *AssureRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action   string `json:"action"` // compliance, security, policy, approval
		Context  string `json:"context"`
		Checks   []string `json:"checks"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"context": req.Context,
		"passed": true,
		"checks": req.Checks,
	}

	return json.Marshal(response)
}

// ValidateRuntime implements validation function execution
type ValidateRuntime struct{}

// Name returns the runtime name
func (r *ValidateRuntime) Name() string {
	return "validate"
}

// Execute executes a validate function
func (r *ValidateRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action   string `json:"action"` // hypothesis, assumption, market, requirements
		Input    interface{} `json:"input"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"valid": true,
		"confidence": 0.85,
	}

	return json.Marshal(response)
}

// SimulateRuntime implements simulation function execution
type SimulateRuntime struct {
	// Monte Carlo configuration
	Iterations int
}

// Name returns the runtime name
func (r *SimulateRuntime) Name() string {
	return "simulate"
}

// Execute executes a simulate function
func (r *SimulateRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action   string `json:"action"` // outcome, workflow, financial, system
		Params   map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"simulations": 1000,
		"outcome": map[string]interface{}{
			"mean": 100,
			"std_dev": 15,
			"confidence": 0.95,
		},
	}

	return json.Marshal(response)
}

// ObserveRuntime implements observation function execution
type ObserveRuntime struct{}

// Name returns the runtime name
func (r *ObserveRuntime) Name() string {
	return "observe"
}

// Execute executes an observe function
func (r *ObserveRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // system, user, workflow, events
		Target string `json:"target,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"target": req.Target,
		"metrics": map[string]interface{}{
			"status": "healthy",
			"latency_ms": 50,
		},
	}

	return json.Marshal(response)
}

// LearnRuntime implements learning function execution
type LearnRuntime struct{}

// Name returns the runtime name
func (r *LearnRuntime) Name() string {
	return "learn"
}

// Execute executes a learn function
func (r *LearnRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // pattern, feedback, success, failure
		Data   interface{} `json:"data"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"learned": true,
		"confidence": 0.9,
	}

	return json.Marshal(response)
}

// AgentMgmtRuntime implements agent management function execution
type AgentMgmtRuntime struct{}

// Name returns the runtime name
func (r *AgentMgmtRuntime) Name() string {
	return "agent_mgmt"
}

// Execute executes an agent management function
func (r *AgentMgmtRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action   string `json:"action"` // spawn, delegate, coordinate, terminate
		AgentID  string `json:"agent_id,omitempty"`
		Task    interface{} `json:"task,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"success": true,
	}

	return json.Marshal(response)
}

// CapabilityRuntime implements capability discovery function execution
type CapabilityRuntime struct{}

// Name returns the runtime name
func (r *CapabilityRuntime) Name() string {
	return "capability"
}

// Execute executes a capability function
func (r *CapabilityRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // list, find, install
		Query string `json:"query,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"capabilities": []string{"web_access", "file_operations", "data_processing"},
	}

	return json.Marshal(response)
}

// WorkflowRuntime implements workflow function execution
type WorkflowRuntime struct{}

// Name returns the runtime name
func (r *WorkflowRuntime) Name() string {
	return "workflow"
}

// Execute executes a workflow function
func (r *WorkflowRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // start, pause, resume, stop
		WorkflowID string `json:"workflow_id,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"workflow_id": req.WorkflowID,
		"status": "running",
	}

	return json.Marshal(response)
}

// MemoryRuntime implements memory function execution
type MemoryRuntime struct{}

// Name returns the runtime name
func (r *MemoryRuntime) Name() string {
	return "memory"
}

// Execute executes a memory function
func (r *MemoryRuntime) Execute(ctx context.Context, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	var req struct {
		Action string `json:"action"` // store, retrieve, update, forget
		Key    string `json:"key,omitempty"`
		Value  interface{} `json:"value,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	response := map[string]interface{}{
		"action": req.Action,
		"success": true,
	}

	return json.Marshal(response)
}

// DefaultRuntimeRouter creates a runtime router with default runtimes
func DefaultRuntimeRouter() *RuntimeRouter {
	router := NewRuntimeRouter()

	// Register all default runtimes
	router.RegisterRuntime(RuntimeTypeSearch, &SearchRuntime{})
	router.RegisterRuntime(RuntimeTypeBrowser, &BrowserRuntime{})
	router.RegisterRuntime(RuntimeTypeCompute, &ComputeRuntime{})
	router.RegisterRuntime(RuntimeTypeData, &DataRuntime{})
	router.RegisterRuntime(RuntimeTypeFile, &FileRuntime{})
	router.RegisterRuntime(RuntimeTypeCommunication, &CommunicationRuntime{})
	router.RegisterRuntime(RuntimeTypeAssure, &AssureRuntime{})
	router.RegisterRuntime(RuntimeTypeValidate, &ValidateRuntime{})
	router.RegisterRuntime(RuntimeTypeSimulate, &SimulateRuntime{})
	router.RegisterRuntime(RuntimeTypeObserve, &ObserveRuntime{})
	router.RegisterRuntime(RuntimeTypeLearn, &LearnRuntime{})
	router.RegisterRuntime(RuntimeTypeAgentMgmt, &AgentMgmtRuntime{})
	router.RegisterRuntime(RuntimeTypeCapability, &CapabilityRuntime{})
	router.RegisterRuntime(RuntimeTypeWorkflow, &WorkflowRuntime{})
	router.RegisterRuntime(RuntimeTypeMemory, &MemoryRuntime{})

	return router
}