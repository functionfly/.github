package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	agentpolicy "github.com/functionfly/functionfly/internal/agent/policy"
	agentquota "github.com/functionfly/functionfly/internal/agent/quota"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AgentExecuteRequest is the request body for agent function execution
type AgentExecuteRequest struct {
	Input     json.RawMessage `json:"input"`
	SessionID string          `json:"session_id,omitempty"`
	CallDepth int             `json:"call_depth,omitempty"`
}

// AgentExecuteResponse is the response from agent function execution
type AgentExecuteResponse struct {
	OK          bool            `json:"ok"`
	Data        json.RawMessage `json:"data,omitempty"`
	ExecutionID string          `json:"execution_id"`
	SessionID   string          `json:"session_id,omitempty"`
	DurationMs  int             `json:"duration_ms"`
	Version     string          `json:"version"`
	CostUSD     float64         `json:"cost_usd"`
	CallDepth   int             `json:"call_depth"`
	Cached      bool            `json:"cached"`
}

// HandleExecute executes a function as an agent with full quota, policy, and attribution
// POST /v1/agent/execute/{author}/{name}
// POST /v1/agent/execute/{author}/{name}/{version}
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 1. Authenticate agent (via X-Agent-API-Key header or JWT)
	agentID, tenantID, err := h.authenticateAgent(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	// 2. Parse request
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

	// 3. Get agent identity and plan tier
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "AGENT_NOT_FOUND", "agent not found")
		return
	}

	// 4. Acquire concurrency slot
	pool, err := h.scheduler.AcquireSlot(r.Context(), agentID, agent.PlanTier)
	if err != nil {
		writeError(w, http.StatusTooManyRequests, "CONCURRENCY_EXCEEDED", err.Error())
		return
	}
	defer pool.Release()

	// 5. Check quota (rate limits + spend caps)
	quotaResult, err := h.quotaEnforcer.CheckAndConsume(r.Context(), agentID, functionURI, 0)
	if err != nil {
		retryAfter := 0
		if quotaResult != nil {
			retryAfter = quotaResult.RetryAfterSecs
		}
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
		if qErr, ok := err.(*agentquota.QuotaViolationError); ok {
			writeError(w, http.StatusTooManyRequests, string(qErr.Code), qErr.Message)
		} else {
			writeError(w, http.StatusTooManyRequests, "QUOTA_EXCEEDED", err.Error())
		}
		return
	}

	// 6. Check behavioral policy
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = r.Header.Get("X-Agent-Session-ID")
	}

	policyReq := &agentpolicy.AgentExecutionRequest{
		AgentID:     agentID,
		SessionID:   sessionID,
		FunctionURI: functionURI,
		Input:       req.Input,
		CallDepth:   req.CallDepth,
	}

	policyResult, err := h.policyEngine.CheckPolicy(r.Context(), agentID, policyReq)
	if err != nil {
		logrus.WithError(err).Warn("policy check error")
	}

	if policyResult != nil && !policyResult.Allowed {
		violation := policyResult.Violation
		record := &attribution.AgentExecutionRecord{
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.Nil,
			FunctionURI: functionURI,
			ExecutionID: generateExecutionID(),
			SessionID:   sessionID,
			CallDepth:   req.CallDepth,
			InputHash:   attribution.HashInput(req.Input),
			LatencyMs:   int(time.Since(startTime).Milliseconds()),
			Outcome:     attribution.OutcomePolicyViolation,
		}
		if violation != nil {
			code := string(violation.Code)
			record.PolicyViolation = &code
		}
		h.attributionRepo.RecordExecution(r.Context(), record)

		writeError(w, http.StatusForbidden, string(violation.Code), violation.Message)
		return
	}

	// 7. Look up function in registry
	fn, err := h.registryRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", fmt.Sprintf("function %s/%s not found", author, name))
		return
	}

	fnVersion, err := h.registryRepo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "VERSION_NOT_FOUND", "no versions available")
		return
	}

	// 8. Execute via the existing registry execution path
	// We delegate to the existing execution infrastructure and wrap with attribution
	executionID := generateExecutionID()
	execStart := time.Now()

	// Build the execution request for the existing handler
	execResult, execErr := h.executeViaRegistry(r, author, name, fnVersion.Version, req.Input)

	latencyMs := int(time.Since(execStart).Milliseconds())
	outcome := attribution.OutcomeSuccess
	var errorCode *string
	if execErr != nil {
		outcome = attribution.OutcomeError
		code := execErr.Error()
		errorCode = &code
	}

	// 9. Record attribution
	var outputHash string
	if execResult != nil {
		outputHash = attribution.HashOutput(execResult)
	}

	record := &attribution.AgentExecutionRecord{
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionID:  fn.ID,
		FunctionURI: fmt.Sprintf("fx://%s/%s@%s", author, name, fnVersion.Version),
		ExecutionID: executionID,
		SessionID:   sessionID,
		CallDepth:   req.CallDepth,
		InputHash:   attribution.HashInput(req.Input),
		OutputHash:  outputHash,
		LatencyMs:   latencyMs,
		Outcome:     outcome,
		ErrorCode:   errorCode,
		Timestamp:   time.Now(),
	}
	h.attributionRepo.RecordExecution(r.Context(), record)

	// 10. Update session cost if session is active
	if sessionID != "" {
		h.attributionRepo.IncrementSessionCost(r.Context(), sessionID, record.CostUSD)
	}

	// 11. Record spend
	h.quotaEnforcer.RecordSpend(r.Context(), agentID, record.CostUSD)

	if execErr != nil {
		writeError(w, http.StatusInternalServerError, "EXECUTION_FAILED", execErr.Error())
		return
	}

	writeJSON(w, http.StatusOK, AgentExecuteResponse{
		OK:          true,
		Data:        execResult,
		ExecutionID: executionID,
		SessionID:   sessionID,
		DurationMs:  latencyMs,
		Version:     fnVersion.Version,
		CostUSD:     record.CostUSD,
		CallDepth:   req.CallDepth,
	})
}

// authenticateAgent extracts agent identity from the request.
// Supports both X-Agent-API-Key header (agent auth) and JWT (user auth).
func (h *Handler) authenticateAgent(r *http.Request) (agentID string, tenantID uuid.UUID, err error) {
	// Try agent API key first
	if apiKey := r.Header.Get("X-Agent-API-Key"); apiKey != "" {
		agent, err := h.identityRepo.GetAgentByAPIKeyHash(r.Context(), apiKey)
		if err != nil {
			return "", uuid.Nil, fmt.Errorf("invalid agent API key")
		}
		return agent.AgentID, agent.TenantID, nil
	}

	// Fall back to JWT user auth
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return "", uuid.Nil, fmt.Errorf("authentication required: provide X-Agent-API-Key or Bearer token")
	}

	// For JWT auth, agent_id must be provided in the request header
	agentID = r.Header.Get("X-Agent-ID")
	if agentID == "" {
		return "", uuid.Nil, fmt.Errorf("X-Agent-ID header required when using JWT authentication")
	}

	return agentID, claims.TenantID, nil
}

// executeViaRegistry delegates execution to the existing registry execution infrastructure.
// This is a thin wrapper that calls the existing Wasm execution path.
func (h *Handler) executeViaRegistry(r *http.Request, author, name, version string, input json.RawMessage) (json.RawMessage, error) {
	// Build an internal execution request to the registry
	// The actual Wasm execution is handled by the existing registry handler infrastructure.
	// For now, we call the registry's GetFunctionByAuthorName and execute via the existing path.
	// In a full implementation, this would call the Wasm runtime directly.

	// This is a placeholder that returns the input as-is for testing.
	// The real implementation would invoke the Wasm sandbox.
	_ = author
	_ = name
	_ = version

	// Return the input echoed back as a placeholder
	// TODO: Wire to actual Wasm execution engine
	return input, nil
}

func generateExecutionID() string {
	return "exec_" + fmt.Sprintf("%d", time.Now().UnixNano())
}
