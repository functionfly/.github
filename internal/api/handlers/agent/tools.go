package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/gorilla/mux"
)

// HandleListTools returns all registered tools available to the authenticated agent.
// GET /agent/tools
func (h *Handler) HandleListTools(w http.ResponseWriter, r *http.Request) {
	agentID, _, err := h.authenticateAgent(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
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
	extra := map[string]interface{}{
		"ok":      true,
		"tools":   toolsOut,
		"agent_id": agentID,
	}
	costMap := make(map[string]float64)
	for _, t := range h.toolRegistry.List() {
		costMap[t.Name()] = h.toolRegistry.ToolCost(t.Name())
	}
	extra["costs_usd"] = costMap
	writeJSON(w, http.StatusOK, extra)
}

// HandleExecuteTool executes a named tool with the given params.
// POST /agent/tools/{tool_name}/call
func (h *Handler) HandleExecuteTool(w http.ResponseWriter, r *http.Request) {
	agentID, tenantID, err := h.authenticateAgent(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}
	if h.toolRegistry == nil {
		writeError(w, http.StatusNotImplemented, "TOOLS_DISABLED", "tool registry not configured")
		return
	}

	vars := mux.Vars(r)
	toolName := vars["tool_name"]
	if toolName == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tool_name is required")
		return
	}

	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid params body")
		return
	}
	if params == nil {
		params = map[string]interface{}{}
	}

	sessionID := r.Header.Get("X-Agent-Session-ID")
	callDepth := 0
	if dStr := r.Header.Get("X-Agent-Call-Depth"); dStr != "" {
		if d, err := parseIntHeader(dStr); err == nil {
			callDepth = d
		}
	}

	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
	tenantIDStr := tenantID.String()

	res, err := h.toolRegistry.ExecuteTool(r.Context(), toolName, params, agentID, tenantIDStr, sessionID, executionID, callDepth)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          false,
			"tool":        toolName,
			"error":       err.Error(),
			"execution_id": executionID,
		})
		return
	}

	errCode := ""
	if res.Error != nil {
		errCode = res.Error.Code
	}
	status := http.StatusOK
	if !res.OK {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]interface{}{
		"ok":          res.OK,
		"tool":        toolName,
		"data":        res.Data,
		"error":       errCode,
		"error_msg":   toolErrMsg(res),
		"duration_ms": res.DurationMs,
		"execution_id": executionID,
	})
}

func parseIntHeader(v string) (int, error) {
	var out int
	_, err := fmt.Sscanf(v, "%d", &out)
	return out, err
}

func toolErrMsg(r *tools.ToolResult) string {
	if r.Error == nil {
		return ""
	}
	return r.Error.Message
}
