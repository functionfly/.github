package agentruntime

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	agentrun "github.com/functionfly/functionfly/internal/agent/runtime"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles agent function API requests
type Handler struct {
	repo        *storage.AgentFunctionRepository
	router      *agentrun.RuntimeRouter
	billingCtrl BillingController
	logger      *logrus.Logger
}

// BillingController interface for billing operations
type BillingController interface {
	ReserveCredits(ctx context.Context, agentID uuid.UUID, functionID uuid.UUID, estimatedCost float64) error
	SettleCredits(ctx context.Context, agentID uuid.UUID, functionID uuid.UUID, actualCost float64) error
	GetCreditBalance(ctx context.Context, agentID uuid.UUID) (float64, error)
}

// NewHandler creates a new agent function handler
func NewHandler(repo *storage.AgentFunctionRepository, billingCtrl BillingController) *Handler {
	return &Handler{
		repo:        repo,
		router:      agentrun.DefaultRuntimeRouter(),
		billingCtrl: billingCtrl,
		logger:      logrus.New(),
	}
}

// SetRuntimeRouter sets a custom runtime router
func (h *Handler) SetRuntimeRouter(router *agentrun.RuntimeRouter) {
	h.router = router
}

// HandleListFunctions handles GET /v1/agent/functions
// Lists all agent functions with optional filtering
func (h *Handler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	category := r.URL.Query().Get("category")
	exclusive := r.URL.Query().Get("exclusive")
	verified := r.URL.Query().Get("verified")
	capabilities := r.URL.Query().Get("capabilities")

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var functions []storage.AgentFunctionDefinition
	var total int64

	// Filter by capabilities
	if capabilities != "" {
		var caps []string
		json.Unmarshal([]byte(capabilities), &caps)
		af, t, e := h.repo.ListByCapabilities(r.Context(), caps, limit, offset)
		if e != nil {
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list functions")
			return
		}
		total = t
		for _, f := range af {
			if def := f.ToDefinition(); def != nil {
				functions = append(functions, *def)
			}
		}
	} else {
		// Filter by category
		var catPtr *string
		if category != "" {
			catPtr = &category
		}

		var exclPtr, verPtr *bool
		if exclusive == "true" {
			b := true
			exclPtr = &b
		}
		if verified == "true" {
			b := true
			verPtr = &b
		}

		af, t, e := h.repo.ListAll(r.Context(), catPtr, exclPtr, verPtr, limit, offset)
		if e != nil {
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list functions")
			return
		}
		total = t
		for _, f := range af {
			if def := f.ToDefinition(); def != nil {
				functions = append(functions, *def)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"functions": functions,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// HandleGetFunction handles GET /v1/agent/functions/{author}/{name}
// Gets a specific agent function by author and name
func (h *Handler) HandleGetFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// This would typically look up the function in the registry
	// For now, return a placeholder response
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"author": author,
		"name":  name,
	})
}

// AgentExecuteRequest is the request body for agent function execution
type AgentExecuteRequest struct {
	Input     json.RawMessage `json:"input"`
	SessionID string          `json:"session_id,omitempty"`
	CallDepth int             `json:"call_depth,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	SpanID    string          `json:"span_id,omitempty"`
}

// AgentExecuteResponse is the response from agent function execution
type AgentExecuteResponse struct {
	Output        json.RawMessage `json:"output,omitempty"`
	ExecutionID  string          `json:"execution_id"`
	DurationMs   int             `json:"duration_ms"`
	CostUSD      float64         `json:"cost_usd"`
	QuotaRemaining float64       `json:"quota_remaining,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// HandleExecuteFunction handles POST /v1/agent/execute/{author}/{name}
// Executes an agent function
func (h *Handler) HandleExecuteFunction(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := vars["version"] // may be empty

	var req AgentExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	functionURI := fmt.Sprintf("fx://%s/%s", author, name)
	if version != "" {
		functionURI += "@" + version
	}

	// Get agent ID from context (set by auth middleware)
	agentIDStr := r.Header.Get("X-Agent-ID")
	if agentIDStr == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "X-Agent-ID header required")
		return
	}
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AGENT_ID", "invalid agent ID format")
		return
	}

	// Get session ID
	sessionIDStr := req.SessionID
	if sessionIDStr == "" {
		sessionIDStr = r.Header.Get("X-Agent-Session-ID")
	}
	var sessionID *uuid.UUID
	if sessionIDStr != "" {
		sid, err := uuid.Parse(sessionIDStr)
		if err == nil {
			sessionID = &sid
		}
	}

	// Look up function (simplified - would normally query registry)
	fn, err := h.lookupFunction(author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", fmt.Sprintf("function %s/%s not found", author, name))
		return
	}

	// Get function category for runtime routing
	category := h.getFunctionCategory(fn)

	// Execute via runtime router
	execReq := &agentrun.ExecutionRequest{
		FunctionID:  fn.ID,
		FunctionURI: functionURI,
		Author:      author,
		Name:        name,
		Version:     version,
		Category:    category,
		Input:       req.Input,
		AgentID:     agentID,
		SessionID:   sessionID,
		TraceID:     req.TraceID,
		SpanID:      req.SpanID,
		CallDepth:   req.CallDepth,
	}

	// Default timeout of 30 seconds
	timeout := 30 * time.Second
	execResp, err := h.router.Execute(r.Context(), execReq, timeout)
	if err != nil {
		// Record failed execution
		h.recordExecution(r.Context(), agentID, fn.ID, sessionID, req.Input, nil, err.Error(), int(time.Since(startTime).Milliseconds()), 0)

		writeError(w, http.StatusInternalServerError, "EXECUTION_FAILED", err.Error())
		return
	}

	// Calculate cost based on function pricing
	costUSD := h.calculateCost(fn, execResp.DurationMs)

	// Record execution
	h.recordExecution(r.Context(), agentID, fn.ID, sessionID, req.Input, execResp.Output, "", execResp.DurationMs, costUSD)

	// Build metadata
	metadata := map[string]interface{}{
		"provider":   execResp.Provider,
		"cached":     execResp.Cached,
		"latency_ms": execResp.DurationMs,
	}
	metadataJSON, _ := json.Marshal(metadata)

	writeJSON(w, http.StatusOK, AgentExecuteResponse{
		Output:      execResp.Output,
		ExecutionID: execResp.ExecutionID,
		DurationMs:  execResp.DurationMs,
		CostUSD:     costUSD,
		Metadata:    metadataJSON,
	})
}

// HandleToolCall handles POST /v1/agent/tools/{tool_name}/call
// Executes a named tool with the given parameters
func (h *Handler) HandleToolCall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	toolName := vars["tool_name"]
	if toolName == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tool_name is required")
		return
	}

	// Parse request body
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if params == nil {
		params = map[string]interface{}{}
	}

	// Get agent ID
	agentIDStr := r.Header.Get("X-Agent-ID")
	if agentIDStr == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "X-Agent-ID header required")
		return
	}
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AGENT_ID", "invalid agent ID format")
		return
	}

	// Get session ID
	sessionIDStr := r.Header.Get("X-Agent-Session-ID")
	var sessionID *uuid.UUID
	if sessionIDStr != "" {
		sid, err := uuid.Parse(sessionIDStr)
		if err == nil {
			sessionID = &sid
		}
	}

	// Parse call depth
	callDepth := 0
	if dStr := r.Header.Get("X-Agent-Call-Depth"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 {
			callDepth = d
		}
	}

	// Limit call depth
	const maxCallDepth = 10
	if callDepth > maxCallDepth {
		writeError(w, http.StatusBadRequest, "CALL_DEPTH_EXCEEDED",
			fmt.Sprintf("call depth %d exceeds maximum of %d", callDepth, maxCallDepth))
		return
	}

	// Generate execution ID
	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())

	// Map tool name to category
	category := h.mapToolToCategory(toolName)

	// Build execution request
	paramsJSON, _ := json.Marshal(params)
	execReq := &agentrun.ExecutionRequest{
		FunctionID:  uuid.Nil, // Tool calls don't have function IDs
		FunctionURI: fmt.Sprintf("tool://%s", toolName),
		Author:     "functionfly",
		Name:       toolName,
		Category:   category,
		Input:      paramsJSON,
		AgentID:    agentID,
		SessionID:  sessionID,
		CallDepth:  callDepth,
	}

	// Execute
	timeout := 30 * time.Second
	execResp, err := h.router.Execute(r.Context(), execReq, timeout)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":           false,
			"tool":         toolName,
			"error_code":   "EXECUTION_FAILED",
			"error_msg":    err.Error(),
			"duration_ms":  0,
			"cost_usd":     0,
			"execution_id": executionID,
		})
		return
	}

	// Calculate cost
	costUSD := h.calculateToolCost(toolName, execResp.DurationMs)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"tool":         toolName,
		"data":         execResp.Output,
		"duration_ms":  execResp.DurationMs,
		"cost_usd":     costUSD,
		"execution_id": executionID,
	})
}

// lookupFunction looks up a function by author and name
// This is a simplified version - would normally query the registry
func (h *Handler) lookupFunction(author, name string) (*storage.RegistryFunction, error) {
	// Simplified - in production this would query the registry
	return &storage.RegistryFunction{
		ID:          uuid.New(),
		Author:      author,
		Name:        name,
		LatestVersion: sql.NullString{String: "1.0.0", Valid: true},
		Description:  sql.NullString{String: fmt.Sprintf("Function %s/%s", author, name), Valid: true},
	}, nil
}

// getFunctionCategory returns the category for a function
func (h *Handler) getFunctionCategory(fn *storage.RegistryFunction) string {
	// Simplified - would normally look up from agent_functions table
	// Default to compute for unknown functions
	return string(agentrun.RuntimeTypeCompute)
}

// mapToolToCategory maps a tool name to its runtime category
func (h *Handler) mapToolToCategory(toolName string) string {
	// Map tool names to categories based on the plan
	switch {
	case hasPrefix(toolName, "search."):
		return string(agentrun.RuntimeTypeSearch)
	case hasPrefix(toolName, "browser."):
		return string(agentrun.RuntimeTypeBrowser)
	case hasPrefix(toolName, "file."):
		return string(agentrun.RuntimeTypeFile)
	case hasPrefix(toolName, "data."):
		return string(agentrun.RuntimeTypeData)
	case hasPrefix(toolName, "compute."):
		return string(agentrun.RuntimeTypeCompute)
	case hasPrefix(toolName, "email.") || hasPrefix(toolName, "sms.") || hasPrefix(toolName, "slack.") || hasPrefix(toolName, "calendar."):
		return string(agentrun.RuntimeTypeCommunication)
	case hasPrefix(toolName, "workflow."):
		return string(agentrun.RuntimeTypeWorkflow)
	case hasPrefix(toolName, "memory."):
		return string(agentrun.RuntimeTypeMemory)
	case hasPrefix(toolName, "assure."):
		return string(agentrun.RuntimeTypeAssure)
	case hasPrefix(toolName, "validate."):
		return string(agentrun.RuntimeTypeValidate)
	case hasPrefix(toolName, "simulate."):
		return string(agentrun.RuntimeTypeSimulate)
	case hasPrefix(toolName, "observe."):
		return string(agentrun.RuntimeTypeObserve)
	case hasPrefix(toolName, "learn."):
		return string(agentrun.RuntimeTypeLearn)
	case hasPrefix(toolName, "agent."):
		return string(agentrun.RuntimeTypeAgentMgmt)
	case hasPrefix(toolName, "capability."):
		return string(agentrun.RuntimeTypeCapability)
	default:
		return string(agentrun.RuntimeTypeCompute)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// calculateCost calculates the cost for a function execution
func (h *Handler) calculateCost(fn *storage.RegistryFunction, durationMs int) float64 {
	// Base pricing from plan
	basePrice := 0.001

	// Category-based multipliers
	category := h.getFunctionCategory(fn)
	switch category {
	case string(agentrun.RuntimeTypeSearch):
		basePrice = 0.001
	case string(agentrun.RuntimeTypeBrowser):
		basePrice = 0.01
	case string(agentrun.RuntimeTypeCompute):
		basePrice = 0.005
	case string(agentrun.RuntimeTypeCommunication):
		basePrice = 0.01
	case string(agentrun.RuntimeTypeWorkflow):
		basePrice = 0.002
	case string(agentrun.RuntimeTypeMemory):
		basePrice = 0.001
	case string(agentrun.RuntimeTypeAssure):
		basePrice = 0.02
	case string(agentrun.RuntimeTypeValidate):
		basePrice = 0.01
	case string(agentrun.RuntimeTypeSimulate):
		basePrice = 0.05
	case string(agentrun.RuntimeTypeObserve):
		basePrice = 0.002
	case string(agentrun.RuntimeTypeLearn):
		basePrice = 0.01
	case string(agentrun.RuntimeTypeAgentMgmt):
		basePrice = 0.005
	case string(agentrun.RuntimeTypeCapability):
		basePrice = 0.001
	}

	// Add per-second billing for compute-intensive functions
	if category == string(agentrun.RuntimeTypeCompute) || category == string(agentrun.RuntimeTypeSimulate) {
		seconds := float64(durationMs) / 1000.0
		basePrice += seconds * 0.001
	}

	return basePrice
}

// calculateToolCost calculates the cost for a tool execution
func (h *Handler) calculateToolCost(toolName string, durationMs int) float64 {
	// Simplified tool pricing - use compute as default category
	return h.calculateCost(&storage.RegistryFunction{}, durationMs)
}

// recordExecution records a function execution
func (h *Handler) recordExecution(ctx context.Context, agentID, functionID uuid.UUID, sessionID *uuid.UUID, input, output json.RawMessage, errStr string, durationMs int, costUSD float64) {
	execution := &storage.AgentFunctionExecution{
		AgentID:    agentID,
		FunctionID: functionID,
		SessionID:  sessionID,
		Input:      input,
		Output:     output,
		Error:      errStr,
		DurationMs: durationMs,
		CostUSD:    costUSD,
	}

	h.repo.RecordExecution(ctx, execution)
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}