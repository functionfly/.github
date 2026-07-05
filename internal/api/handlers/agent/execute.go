package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	agentpolicy "github.com/functionfly/functionfly/internal/agent/policy"
	agentquota "github.com/functionfly/functionfly/internal/agent/quota"
	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/paperclip/costbridge"
	"github.com/functionfly/functionfly/internal/storage"
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
		writeErrorFromErr(r, w, http.StatusUnauthorized, "UNAUTHORIZED", "authenticate agent for execute", err)
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
		writeErrorFromErr(r, w, http.StatusTooManyRequests, "CONCURRENCY_EXCEEDED", "acquire concurrency slot", err)
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
			writeErrorFromErr(r, w, http.StatusTooManyRequests, "QUOTA_EXCEEDED", "check quota", err)
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
		if err := h.attributionRepo.RecordExecution(r.Context(), record); err != nil {
			logrus.WithError(err).Warn("failed to record execution")
		}

		writeError(w, http.StatusForbidden, string(violation.Code), violation.Message)
		return
	}

	// 7. Look up function in registry
	fn, err := h.registryRepo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", fmt.Sprintf("function %s/%s not found", author, name))
		return
	}
	chargeUSD := fn.PricePerCall

	if chargeUSD > 0 {
		controls, controlsErr := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
		if controlsErr != nil {
			writeError(w, http.StatusInternalServerError, "BILLING_CONTROLS_FAILED", "failed to load billing controls")
			return
		}
		if controls.CreditBalanceUSD < chargeUSD {
			writeError(w, http.StatusPaymentRequired, "INSUFFICIENT_CREDIT_BALANCE", fmt.Sprintf("insufficient balance: need $%.4f", chargeUSD))
			return
		}
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
	executionCostUSD := 0.0
	var errorCode *string
	if execErr != nil {
		outcome = attribution.OutcomeError
		code := execErr.Error()
		errorCode = &code
	} else {
		executionCostUSD = chargeUSD

		if chargeUSD > 0 {
			consumeUpdate, consumeErr := h.billingCtrl.ConsumeCredits(r.Context(), agentID, chargeUSD)
			if consumeErr != nil {
				logrus.WithError(consumeErr).WithFields(logrus.Fields{
					"agent_id":   agentID,
					"execution":  executionID,
					"charge_usd": chargeUSD,
				}).Warn("failed to consume credits after execution")
				writeError(w, http.StatusPaymentRequired, "CREDIT_CONSUME_FAILED", consumeErr.Error())
				return
			}

			lowBalanceThreshold := walletLowBalanceThresholdUSD()
			if h.notificationSvc != nil &&
				consumeUpdate != nil &&
				consumeUpdate.PreviousUSD > lowBalanceThreshold &&
				consumeUpdate.CurrentUSD <= lowBalanceThreshold {
				userIDs, usersErr := h.userRepo.ListUserIDsByTenant(r.Context(), tenantID)
				if usersErr != nil {
					logrus.WithError(usersErr).WithField("tenant_id", tenantID).Warn("failed to list tenant users for low-balance notification")
				} else if len(userIDs) > 0 {
					err = h.notificationSvc.Broadcast(r.Context(), notification.BroadcastRequest{
						UserIDs:  userIDs,
						Type:     notification.TypeBillingWalletLowBalance,
						Category: notification.CategoryBilling,
						Title:    "Low Wallet Balance",
						Body: fmt.Sprintf(
							"Wallet balance for agent %s is low ($%.2f). It is now at or below your alert threshold of $%.2f.",
							agentID,
							consumeUpdate.CurrentUSD,
							lowBalanceThreshold,
						),
						Data: notification.JSONMap{
							"agent_id":      agentID,
							"balance_usd":   consumeUpdate.CurrentUSD,
							"threshold_usd": lowBalanceThreshold,
						},
						Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
						Priority: notification.PriorityHigh,
					})
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"tenant_id": tenantID,
							"agent_id":  agentID,
						}).Warn("failed to broadcast low-balance notification")
					}
				}
			}
		}
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
		CostUSD:     executionCostUSD,
		Outcome:     outcome,
		ErrorCode:   errorCode,
		Timestamp:   time.Now(),
	}
	if err := h.attributionRepo.RecordExecution(r.Context(), record); err != nil {
		logrus.WithError(err).Warn("failed to record execution")
	}

	// 10. Update session cost if session is active
	if sessionID != "" {
		if err := h.attributionRepo.IncrementSessionCost(r.Context(), sessionID, record.CostUSD); err != nil {
			logrus.WithError(err).Warn("failed to increment session cost")
		}
	}

	// 11. Record spend
	if err := h.quotaEnforcer.RecordSpend(r.Context(), agentID, record.CostUSD); err != nil {
		logrus.WithError(err).Warn("failed to record spend")
	}

	// 12. Push cost to Paperclip for budget enforcement (if configured)
	if record.CostUSD >= 0 {
		cfg := costbridge.FromEnv()
		_ = costbridge.ReportCost(r.Context(), cfg, record.CostUSD, map[string]string{
			"execution_id": executionID,
			"function_uri": record.FunctionURI,
			"agent_id":     agentID,
		})
	}

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
// This uses the registry repository to get the function and then executes it via WASM.
func (h *Handler) executeViaRegistry(r *http.Request, author, name, version string, input json.RawMessage) (json.RawMessage, error) {
	// Get function by author and name
	fn, err := h.registryRepo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		return nil, fmt.Errorf("function not found: %s/%s: %w", author, name, err)
	}

	// Get the specific version or latest
	var fnVersion *storage.RegistryFunctionVersion
	if version != "" {
		fnVersion, err = h.registryRepo.GetFunctionVersion(fn.ID, version)
		if err != nil {
			return nil, fmt.Errorf("version not found: %s: %w", version, err)
		}
	} else {
		// Get latest version
		versions, err := h.registryRepo.ListFunctionVersions(fn.ID)
		if err != nil || len(versions) == 0 {
			return nil, fmt.Errorf("no versions available for function")
		}
		// Find the latest version (highest version number)
		for _, v := range versions {
			if fnVersion == nil || v.Version > fnVersion.Version {
				vCopy := v
				fnVersion = (*storage.RegistryFunctionVersion)(&vCopy)
			}
		}
	}

	// Check if the function has WASM binary
	if fnVersion == nil || len(fnVersion.WasmBinary) == 0 {
		return nil, fmt.Errorf("function has no WASM binary to execute")
	}

	// Execute the function using WASM sandbox
	// Default timeout of 30 seconds if not specified
	timeoutMs := fnVersion.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 30000
	}

	// Execute via the sandbox executor
	executor, err := execution.NewSandboxExecutor()
	if err != nil {
		return nil, fmt.Errorf("create sandbox executor: %w", err)
	}
	result, err := executor.ExecuteFunction(fnVersion, input, timeoutMs)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return result, nil
}

func generateExecutionID() string {
	return "exec_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func walletLowBalanceThresholdUSD() float64 {
	const defaultThreshold = 5.0
	raw := os.Getenv("AGENT_WALLET_LOW_BALANCE_USD")
	if raw == "" {
		return defaultThreshold
	}

	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || parsed <= 0 {
		return defaultThreshold
	}
	return parsed
}
