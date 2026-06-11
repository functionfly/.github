package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListTools returns all registered tools available to the authenticated agent.
// GET /v1/agent/tools
func (h *Handler) HandleListTools(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	if h.toolRegistry == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    true,
			"tools": []tools.ToolDefinition{},
		})
		return
	}

	toolsOut := h.toolRegistry.ListDefinitions()

	// Build response with costs
	costMap := make(map[string]float64, len(toolsOut))
	for _, t := range toolsOut {
		costMap[t.Name] = t.CostUSD
	}

	// Filter tools based on agent's policy - but for now return all definitions
	// In production, you'd check policyEngine.IsToolAllowed for each tool

	logrus.WithFields(logrus.Fields{
		"tenant_id": claims.TenantID,
		"agent_id":  claims.UserID,
		"tool_count": len(toolsOut),
	}).Debug("listing tools")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"tools":    toolsOut,
		"costs_usd": costMap,
	})
}

// HandleGetTool returns a specific tool's definition
// GET /v1/agent/tools/{tool_name}
func (h *Handler) HandleGetTool(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	vars := mux.Vars(r)
	toolName := vars["tool_name"]
	if toolName == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tool_name is required")
		return
	}

	if h.toolRegistry == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found")
		return
	}

	tool, ok := h.toolRegistry.Get(toolName)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found")
		return
	}

	// Check policy - is this tool allowed for this agent?
	// allowed, violation := h.policyEngine.IsToolAllowed(r.Context(), agentID, toolName)
	// if !allowed { ... }

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"tool":        tool.Definition(),
		"cost_usd":    tool.Definition().CostUSD,
	})
}

// HandleExecuteTool executes a named tool with the given params.
// POST /v1/agent/tools/{tool_name}/call
func (h *Handler) HandleExecuteTool(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	if h.toolRegistry == nil {
		writeError(w, http.StatusServiceUnavailable, "TOOLS_DISABLED", "tool registry not configured")
		return
	}

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

	// Extract execution context from headers
	sessionID := r.Header.Get("X-Agent-Session-ID")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "X-Agent-Session-ID header required")
		return
	}

	callDepth := 0
	if dStr := r.Header.Get("X-Agent-Call-Depth"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 {
			callDepth = d
		}
	}

	// Limit call depth to prevent infinite recursion
	const maxCallDepth = 10
	if callDepth > maxCallDepth {
		writeError(w, http.StatusBadRequest, "CALL_DEPTH_EXCEEDED",
			fmt.Sprintf("call depth %d exceeds maximum of %d", callDepth, maxCallDepth))
		return
	}

	// Generate execution ID for tracing
	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())

	// Build execution context
	execCtx := tools.ExecutionContext{
		AgentID:     claims.UserID.String(),
		TenantID:    claims.TenantID.String(),
		SessionID:   sessionID,
		ExecutionID: executionID,
		CallDepth:   callDepth,
		Metadata: map[string]interface{}{
			"ip_address": getClientIP(r),
			"user_agent": r.UserAgent(),
		},
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"tool":         toolName,
		"agent_id":     claims.UserID,
		"tenant_id":    claims.TenantID,
		"session_id":   sessionID,
		"call_depth":   callDepth,
	}).Info("executing tool")

	// Execute tool
	result := h.toolRegistry.ExecuteTool(r.Context(), toolName, params, execCtx)

	// Log the execution
	if result.OK {
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"tool":         toolName,
			"duration_ms":  result.DurationMs,
			"cost_usd":     result.CostUSD,
		}).Info("tool execution succeeded")
	} else {
		errCode := ""
		if result.Error != nil {
			errCode = result.Error.Code
		}
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"tool":         toolName,
			"error_code":   errCode,
			"duration_ms":  result.DurationMs,
		}).Warn("tool execution failed")
	}

	// Determine HTTP status
	status := http.StatusOK
	if !result.OK {
		status = http.StatusBadRequest
	}

	errCode := ""
	errMsg := ""
	if result.Error != nil {
		errCode = result.Error.Code
		errMsg = result.Error.Message
	}

	writeJSON(w, status, map[string]interface{}{
		"ok":           result.OK,
		"tool":         toolName,
		"data":         result.Data,
		"error_code":   errCode,
		"error_msg":    errMsg,
		"duration_ms":  result.DurationMs,
		"cost_usd":     result.CostUSD,
		"execution_id": executionID,
	})
}

// HandleListToolCalls returns tool call history for the agent
// GET /v1/agent/tools/calls
func (h *Handler) HandleListToolCalls(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	// Parse pagination
	limit := 20
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

	sessionID := r.URL.Query().Get("session_id")

	logrus.WithFields(logrus.Fields{
		"tenant_id": claims.TenantID,
		"agent_id":  claims.UserID,
		"session_id": sessionID,
		"limit":     limit,
		"offset":    offset,
	}).Debug("listing tool calls")

	records, total, err := h.attributionRepo.ListToolCalls(r.Context(), claims.UserID.String(), sessionID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to list tool calls")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tool calls")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"calls":  records,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
