package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/aikeys"
	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/chat"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type agentChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

type agentChatResponse struct {
	OK               bool                  `json:"ok"`
	Message          string                `json:"message"`
	Model            string                `json:"model"`
	Thinking         *chat.ThinkingContent `json:"thinking,omitempty"`
	PromptTokens     int                   `json:"prompt_tokens,omitempty"`
	CompletionTokens int                   `json:"completion_tokens,omitempty"`
	TotalTokens      int                   `json:"total_tokens,omitempty"`
}

// agentSessionID returns a deterministic UUID for an agent's chat session.
func agentSessionID(agentID uuid.UUID) uuid.UUID {
	h := sha256.Sum256([]byte("agent-chat-session:" + agentID.String()))
	s := hex.EncodeToString(h[:16])
	u, _ := uuid.Parse(s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:])
	return u
}

// shouldThink determines if thinking should be enabled based on mode and message content.
func shouldThink(message, mode string) bool {
	if mode == "always" {
		return true
	}
	if mode == "off" || mode == "" {
		return false
	}
	// auto mode
	if len(message) > 200 {
		return true
	}
	if strings.Contains(message, "```") {
		return true
	}
	keywords := []string{"analyze", "debug", "explain", "compare", "design",
		"architect", "evaluate", "think", "reason", "why", "how",
		"optimize", "review", "trace", "investigate"}
	lower := strings.ToLower(message)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// HandleAgentChat sends a message to the agent via the FlyMind AI service
// using the agent's configured model. Persists messages to chat history.
// POST /v1/agent/{agent_id}/chat
func (h *Handler) HandleAgentChat(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req agentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "message is required")
		return
	}

	model := agent.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	systemPrompt := fmt.Sprintf(
		"You are %s, an AI agent. %s Respond concisely and helpfully.",
		agent.Name, agent.Description,
	)

	// Resolve session ID: use provided session_id, or fall back to deterministic default
	var sessionID uuid.UUID
	if req.SessionID != "" {
		sid, err := uuid.Parse(req.SessionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SESSION", "invalid session_id")
			return
		}
		sessionID = sid
	} else {
		sessionID = agentSessionID(agent.ID)
	}

	// Execute tools if the message looks like it needs them
	toolResults := h.executeAgentTools(r.Context(), req.Message, agent, claims)

	messageWithContext := req.Message
	if toolResults != "" {
		messageWithContext = fmt.Sprintf("%s\n\n[Tool Results]\n%s", req.Message, toolResults)
		systemPrompt += "\n\nYou have access to tools. When tool results are provided in [Tool Results], use them to answer the user's question. Cite sources when available."
	}

	// Build thinking config from agent settings
	var thinking *chat.ThinkingConfig
	if shouldThink(req.Message, agent.ThinkingMode) {
		budget := agent.ThinkingBudget
		if budget <= 0 {
			budget = 10000
		}
		thinking = &chat.ThinkingConfig{
			Mode:         agent.ThinkingMode,
			BudgetTokens: budget,
		}
	}

	// Persist user message
	modelCopy := model
	h.saveChatMessage(r.Context(), sessionID, "user", req.Message, &modelCopy, nil)

	// Try BYOK direct path: call the LLM provider directly, bypassing FlyMind
	if h.byokRepo != nil {
		reply, thinkingContent, byokResp, ok := h.tryBYOKDirect(r.Context(), model, systemPrompt, messageWithContext, claims, thinking)
		if ok {
			h.saveChatMessage(r.Context(), sessionID, "assistant", reply, &modelCopy, thinkingContent)

			// Record execution with token data
			if h.attributionRepo != nil && byokResp != nil {
				execRecord := &attribution.AgentExecutionRecord{
					AgentID:          agent.AgentID,
					TenantID:         agent.TenantID,
					FunctionID:       uuid.Nil,
					FunctionURI:      fmt.Sprintf("chat://%s", model),
					ExecutionID:      generateExecutionID(),
					SessionID:        sessionID.String(),
					LatencyMs:        0,
					Outcome:          attribution.OutcomeSuccess,
					ModelName:        byokResp.Model,
					Provider:         chat.ProviderFromModel(model),
					PromptTokens:     byokResp.PromptTokens,
					CompletionTokens: byokResp.CompletionTokens,
					TotalTokens:      byokResp.TotalTokens,
					ReasoningTokens:  byokResp.ReasoningTokens,
					Timestamp:        time.Now(),
				}
				h.attributionRepo.RecordExecution(r.Context(), execRecord)
			}

			writeJSON(w, http.StatusOK, agentChatResponse{
				OK:               true,
				Message:          reply,
				Model:            model,
				Thinking:         thinkingContent,
				PromptTokens:     byokResp.PromptTokens,
				CompletionTokens: byokResp.CompletionTokens,
				TotalTokens:      byokResp.TotalTokens,
			})
			return
		}
	}

	// Fall back to FlyMind
	aiBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiBaseURL == "" {
		aiBaseURL = "http://localhost:18081"
	}
	aiAPIKey := os.Getenv("AI_SERVICE_API_KEY")

	contextMap := map[string]string{
		"agent_id":      agent.AgentID,
		"agent_name":    agent.Name,
		"system_prompt": systemPrompt,
	}
	if thinking != nil {
		contextMap["thinking_mode"] = thinking.Mode
		contextMap["thinking_budget"] = strconv.Itoa(thinking.BudgetTokens)
	}

	aiReq := map[string]interface{}{
		"session_id": sessionID.String(),
		"message":    messageWithContext,
		"model":      model,
		"tenant_id":  claims.TenantID.String(),
		"user_id":    claims.UserID.String(),
		"context":    contextMap,
	}

	body, _ := json.Marshal(aiReq)
	httpReq, err := http.NewRequestWithContext(r.Context(), "POST", aiBaseURL+"/api/chat/message", bytes.NewBuffer(body))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "failed to create AI request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		httpReq.Header.Set("X-API-Key", aiAPIKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_UNREACHABLE", "AI service is unreachable")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		writeJSON(w, resp.StatusCode, map[string]interface{}{
			"ok":      false,
			"message": fmt.Sprintf("AI service returned status %d", resp.StatusCode),
			"raw":     string(respBody),
		})
		return
	}

	var aiResp struct {
		Message         string `json:"message"`
		Model           string `json:"model"`
		ThinkingContent string `json:"thinking_content"`
		ThinkingTokens  int    `json:"thinking_tokens"`
	}
	reply := string(respBody)
	if err := json.Unmarshal(respBody, &aiResp); err == nil && aiResp.Message != "" {
		reply = aiResp.Message
	}

	var flyMindThinking *chat.ThinkingContent
	if aiResp.ThinkingContent != "" {
		flyMindThinking = &chat.ThinkingContent{
			Content: aiResp.ThinkingContent,
			Tokens:  aiResp.ThinkingTokens,
		}
	}

	// Persist assistant reply
	h.saveChatMessage(r.Context(), sessionID, "assistant", reply, &modelCopy, flyMindThinking)

	writeJSON(w, http.StatusOK, agentChatResponse{
		OK:       true,
		Message:  reply,
		Model:    model,
		Thinking: flyMindThinking,
	})
}

// tryBYOKDirect attempts to call the LLM provider directly using a BYOK key.
// Returns (reply, thinkingContent, byokResponse, true) on success, ("", nil, nil, false) if no BYOK key is available.
func (h *Handler) tryBYOKDirect(ctx context.Context, model, systemPrompt, userMessage string, claims *auth.Claims, thinking *chat.ThinkingConfig) (string, *chat.ThinkingContent, *chat.BYOKResponse, bool) {
	if h.byokRepo == nil {
		return "", nil, nil, false
	}

	provider := chat.ProviderFromModel(model)
	if provider == "" {
		return "", nil, nil, false
	}

	tid := claims.TenantID
	key, err := h.byokRepo.GetByTenantAndProvider(ctx, tid, provider)
	if err != nil || key == nil || key.Status != "active" {
		// Try token-plan variants
		switch provider {
		case "mimo":
			key, err = h.byokRepo.GetByTenantAndProvider(ctx, tid, "mimo-token-plan")
		case "minimax":
			key, err = h.byokRepo.GetByTenantAndProvider(ctx, tid, "minimax-token-plan")
		default:
			return "", nil, nil, false
		}
		if err != nil || key == nil || key.Status != "active" {
			return "", nil, nil, false
		}
	}

	plaintext, err := aikeys.DecryptKey(key.EncryptedKey, key.KeyNonce, tid)
	if err != nil {
		logrus.WithError(err).Warn("tryBYOKDirect: failed to decrypt key")
		return "", nil, nil, false
	}

	// Resolve base URL (token plans may have region-specific endpoints)
	callProvider := provider
	baseURL := ""
	if key.Provider == "mimo-token-plan" {
		callProvider = "mimo"
		region := extractRegionFromHealthMessage(key.HealthMessage)
		if base, ok := aikeys.TokenPlanRegionURLs[region]; ok {
			baseURL = base
		}
	} else if key.Provider == "minimax-token-plan" {
		callProvider = "minimax"
	}

	resp, err := chat.CallLLM(ctx, chat.BYOKRequest{
		Provider:     callProvider,
		APIKey:       string(plaintext),
		BaseURL:      baseURL,
		Model:        model,
		SystemPrompt: systemPrompt,
		UserMessage:  userMessage,
		MaxTokens:    1024,
		Temperature:  0.7,
		Thinking:     thinking,
	})
	if err != nil {
		logrus.WithError(err).WithField("provider", callProvider).Warn("tryBYOKDirect: LLM call failed")
		return "", nil, nil, false
	}

	return resp.Content, resp.ThinkingContent, resp, true
}

// HandleAgentChatClear deletes all chat messages for an agent session.
// DELETE /v1/agent/{agent_id}/chat
func (h *Handler) HandleAgentChatClear(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	if h.rawDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}

	// Resolve session ID from query param, or fall back to deterministic default
	var sessionID uuid.UUID
	if sidStr := r.URL.Query().Get("session_id"); sidStr != "" {
		sid, err := uuid.Parse(sidStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SESSION", "invalid session_id")
			return
		}
		sessionID = sid
	} else {
		sessionID = agentSessionID(agent.ID)
	}

	if _, err := h.rawDB.ExecContext(r.Context(),
		`DELETE FROM ai_chat_messages WHERE session_id = $1`, sessionID,
	); err != nil {
		logrus.WithError(err).WithField("session_id", sessionID).Error("HandleAgentChatClear: failed to delete messages")
		writeError(w, http.StatusInternalServerError, "CLEAR_FAILED", "failed to clear chat")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// HandleAgentChatHistory returns chat history for an agent.
// GET /v1/agent/{agent_id}/chat/history?limit=50&offset=0
func (h *Handler) HandleAgentChatHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// Resolve session ID from query param, or fall back to deterministic default
	var sessionID uuid.UUID
	if sidStr := r.URL.Query().Get("session_id"); sidStr != "" {
		sid, err := uuid.Parse(sidStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SESSION", "invalid session_id")
			return
		}
		sessionID = sid
	} else {
		sessionID = agentSessionID(agent.ID)
	}

	db := h.rawDB
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "messages": []map[string]interface{}{}})
		return
	}

	rows, err := db.QueryContext(r.Context(),
		`SELECT role, content, COALESCE(model, ''), COALESCE(metadata, '{}'), created_at FROM ai_chat_messages WHERE session_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		sessionID, limit, offset,
	)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "messages": []interface{}{}})
		return
	}
	defer rows.Close()

	type chatMsg struct {
		Role      string                 `json:"role"`
		Content   string                 `json:"content"`
		Model     string                 `json:"model,omitempty"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
		CreatedAt time.Time              `json:"created_at"`
	}

	var messages []chatMsg
	for rows.Next() {
		var m chatMsg
		var metadataStr string
		if err := rows.Scan(&m.Role, &m.Content, &m.Model, &metadataStr, &m.CreatedAt); err == nil {
			if metadataStr != "" && metadataStr != "{}" {
				var meta map[string]interface{}
				if json.Unmarshal([]byte(metadataStr), &meta) == nil {
					m.Metadata = meta
				}
			}
			messages = append(messages, m)
		}
	}

	if messages == nil {
		messages = []chatMsg{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"messages": messages,
	})
}

// saveChatMessage persists a chat message to the database.
func (h *Handler) saveChatMessage(ctx context.Context, sessionID uuid.UUID, role, content string, model *string, thinking *chat.ThinkingContent) {
	if h.rawDB == nil {
		logrus.Warn("saveChatMessage: rawDB is nil, cannot persist chat message")
		return
	}

	// Build metadata with thinking content
	metadata := "{}"
	if thinking != nil && thinking.Content != "" {
		meta := map[string]interface{}{
			"thinking": map[string]interface{}{
				"content": thinking.Content,
				"tokens":  thinking.Tokens,
			},
		}
		if b, err := json.Marshal(meta); err == nil {
			metadata = string(b)
		}
	}

	// Ensure session exists
	if _, err := h.rawDB.ExecContext(ctx,
		`INSERT INTO ai_chat_sessions (id, user_id, tenant_id, title, context_type, context_reference, is_active, created_at, updated_at)
		 VALUES ($1, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 'Agent Chat', 'agent', NULL, true, NOW(), NOW())
		 ON CONFLICT (id) DO NOTHING`,
		sessionID,
	); err != nil {
		logrus.WithError(err).WithField("session_id", sessionID).Error("saveChatMessage: failed to upsert session")
		return
	}

	if _, err := h.rawDB.ExecContext(ctx,
		`INSERT INTO ai_chat_messages (id, session_id, role, content, tokens_used, model, metadata, created_at)
		 VALUES ($1, $2, $3, $4, 0, $5, $6, NOW())`,
		uuid.New(), sessionID, role, content, model, metadata,
	); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"role":       role,
		}).Error("saveChatMessage: failed to insert message")
	}
}

// isToolEnabled checks if a tool is enabled in the agent's capabilities.
func isToolEnabled(agent *identity.AgentIdentity, toolKey string) bool {
	if agent.Capabilities == nil {
		return false
	}
	val, ok := agent.Capabilities[toolKey]
	return ok && (val == "true" || val == "enabled")
}

// executeAgentTools detects tool-worthy intents and executes them, returning results as text.
func (h *Handler) executeAgentTools(ctx context.Context, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	if !isToolEnabled(agent, "web_search") {
		return ""
	}

	msg := strings.ToLower(message)

	searchTriggers := []string{
		"search", "find", "look up", "what is", "what are", "best", "compare",
		"who is", "where", "how to", "latest", "news", "price", "buy",
		"developments", "recent", "trending", "current", "update", "updates",
	}
	needsSearch := false
	for _, trigger := range searchTriggers {
		if strings.Contains(msg, trigger) {
			needsSearch = true
			break
		}
	}

	if !needsSearch {
		return ""
	}

	// Call the AI service's tool execution endpoint
	aiBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiBaseURL == "" {
		aiBaseURL = "http://localhost:18081"
	}
	aiAPIKey := os.Getenv("AI_SERVICE_API_KEY")

	toolReq := map[string]interface{}{
		"tool": "web_search",
		"params": map[string]interface{}{
			"query":        message,
			"max_results":  5,
		},
		"agent_id":  agent.AgentID,
		"tenant_id": claims.TenantID.String(),
	}

	body, _ := json.Marshal(toolReq)
	req, err := http.NewRequestWithContext(ctx, "POST", aiBaseURL+"/api/tools/execute", bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		req.Header.Set("X-API-Key", aiAPIKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback: execute search directly
		return h.executeSearchDirect(ctx, message)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK   bool        `json:"ok"`
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.OK {
		return h.executeSearchDirect(ctx, message)
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

// executeSearchDirect runs a web search directly without the AI service.
func (h *Handler) executeSearchDirect(ctx context.Context, query string) string {
	searchURL := os.Getenv("AGENT_WEB_SEARCH_URL")
	if searchURL == "" {
		return ""
	}

	reqBody := fmt.Sprintf(`{"query":%q,"max_results":5}`, query)
	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, strings.NewReader(reqBody))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	data, _ := json.MarshalIndent(parsed, "", "  ")
	return string(data)
}

func extractRegionFromHealthMessage(healthMessage string) string {
	const prefix = "region:"
	if len(healthMessage) > len(prefix) && healthMessage[:len(prefix)] == prefix {
		return healthMessage[len(prefix):]
	}
	return ""
}
