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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/aikeys"
	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/chat"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/tools"
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
	SessionID        string                `json:"session_id,omitempty"`
	Message          string                `json:"message"`
	Model            string                `json:"model"`
	Thinking         *chat.ThinkingContent `json:"thinking,omitempty"`
	PromptTokens     int                   `json:"prompt_tokens,omitempty"`
	CompletionTokens int                   `json:"completion_tokens,omitempty"`
	TotalTokens      int                   `json:"total_tokens,omitempty"`
	KeySource        string                `json:"key_source,omitempty"`
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

	description := agent.Description
	if description == "" {
		description = fmt.Sprintf("A general-purpose AI assistant agent.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, an AI agent running on the %s model. Your agent ID is %s. %s", agent.Name, model, agent.AgentID, description))
	sb.WriteString(fmt.Sprintf("\n\nIdentity: Your name is %s. Your model is %s. You are designed to assist users based on your description.", agent.Name, model))

	if agent.Capabilities != nil && len(agent.Capabilities) > 0 {
		sb.WriteString("\n\nYou have access to the following tools and capabilities:")
		for key, val := range agent.Capabilities {
			if val == "true" || val == "enabled" {
				switch key {
				case "web_search":
					sb.WriteString("\n- Web Search: Search the internet for current information, news, and answers.")
				case "database_query":
					sb.WriteString("\n- Database Query: Run read-only SQL SELECT queries against the database.")
				case "file_read":
					sb.WriteString("\n- File Read: Read files from the agent workspace filesystem.")
				case "file_write":
					sb.WriteString("\n- File Write: Create or write files to the agent workspace filesystem.")
				case "http_request":
					sb.WriteString("\n- HTTP Request: Make HTTP API requests to external URLs and APIs.")
				case "code_execution":
					sb.WriteString("\n- Code Execution: Execute code in sandboxed environments (Python, Node.js, Go, Bash).")
				case "image_generation":
					sb.WriteString("\n- Image Generation: Generate images from text prompts using AI.")
				case "text_to_speech":
					sb.WriteString("\n- Text to Speech: Convert text to speech audio.")
				case "email":
					sb.WriteString("\n- Email: Send emails and notifications to users.")
				case "notification":
					sb.WriteString("\n- Notification: Send push notifications and alerts.")
				default:
					sb.WriteString(fmt.Sprintf("\n- %s", key))
				}
			}
		}
		sb.WriteString("\n\nTo use a tool, respond with a JSON block: {\"tool\": \"tool_name\", \"args\": { ... }}. Tool results will be provided in [Tool Results].")
	}

	sb.WriteString("\n\nRespond concisely and helpfully. When asked about your identity, always state your name, model, and purpose.")

	systemPrompt := sb.String()

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
	h.saveChatMessage(r.Context(), sessionID, "user", req.Message, &modelCopy, nil, claims.UserID.String(), claims.TenantID.String(), agent.AgentID)

	// Try BYOK direct path: call the LLM provider directly, bypassing FlyMind
	// BYOK is the primary path — users bring their own keys
	if h.byokRepo != nil {
		reply, thinkingContent, byokResp, ok := h.tryBYOKDirect(r.Context(), model, systemPrompt, messageWithContext, claims, thinking)
		if ok {
			h.saveChatMessage(r.Context(), sessionID, "assistant", reply, &modelCopy, thinkingContent, claims.UserID.String(), claims.TenantID.String(), agent.AgentID)

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
				if err := h.attributionRepo.RecordExecution(r.Context(), execRecord); err != nil {
				logrus.WithError(err).Warn("failed to record execution")
			}
			}

			writeJSON(w, http.StatusOK, agentChatResponse{
				OK:               true,
				SessionID:        sessionID.String(),
				Message:          reply,
				Model:            model,
				Thinking:         thinkingContent,
				PromptTokens:     byokResp.PromptTokens,
				CompletionTokens: byokResp.CompletionTokens,
				TotalTokens:      byokResp.TotalTokens,
				KeySource:        "byok",
			})
			return
		}
	}

	// Fall back to FlyMind (uses OpenRouter free models when no platform key is configured)
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
	defer func() { _ = resp.Body.Close() }()

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
	h.saveChatMessage(r.Context(), sessionID, "assistant", reply, &modelCopy, flyMindThinking, claims.UserID.String(), claims.TenantID.String(), agent.AgentID)

	writeJSON(w, http.StatusOK, agentChatResponse{
		OK:        true,
		SessionID: sessionID.String(),
		Message:   reply,
		Model:     model,
		Thinking:  flyMindThinking,
		KeySource: "platform",
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
	defer func() { _ = rows.Close() }()

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
func (h *Handler) saveChatMessage(ctx context.Context, sessionID uuid.UUID, role, content string, model *string, thinking *chat.ThinkingContent, userID, tenantID, agentID string) {
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
		`INSERT INTO ai_chat_sessions (id, user_id, tenant_id, title, agent_id, context_type, context_reference, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, 'Agent Chat', $4, 'agent', NULL, true, NOW(), NOW())
		 ON CONFLICT (id) DO UPDATE SET user_id = EXCLUDED.user_id, agent_id = EXCLUDED.agent_id, tenant_id = EXCLUDED.tenant_id`,
		sessionID, userID, tenantID, agentID,
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
	msg := strings.ToLower(message)
	var results []string

	if isToolEnabled(agent, "web_search") {
		if result := h.tryWebSearch(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Web Search]\n"+result)
		}
	}

	if isToolEnabled(agent, "database_query") {
		if result := h.tryDatabaseQuery(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Database Query]\n"+result)
		}
	}

	if isToolEnabled(agent, "file_read") {
		if result := h.tryFileRead(ctx, msg, message); result != "" {
			results = append(results, "[File Read]\n"+result)
		}
	}

	if isToolEnabled(agent, "file_write") {
		if result := h.tryFileWrite(ctx, msg, message); result != "" {
			results = append(results, "[File Write]\n"+result)
		}
	}

	if isToolEnabled(agent, "http_request") {
		if result := h.tryHTTPRequest(ctx, msg, message); result != "" {
			results = append(results, "[HTTP Request]\n"+result)
		}
	}

	if isToolEnabled(agent, "code_execution") {
		if result := h.tryCodeExecution(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Code Execution]\n"+result)
		}
	}

	if isToolEnabled(agent, "image_generation") {
		if result := h.tryImageGeneration(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Image Generation]\n"+result)
		}
	}

	if isToolEnabled(agent, "text_to_speech") {
		if result := h.tryTextToSpeech(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Text to Speech]\n"+result)
		}
	}

	if isToolEnabled(agent, "email") {
		if result := h.tryEmailTool(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Email]\n"+result)
		}
	}

	if isToolEnabled(agent, "notification") {
		if result := h.tryNotificationTool(ctx, msg, message, agent, claims); result != "" {
			results = append(results, "[Notification]\n"+result)
		}
	}

	return strings.Join(results, "\n\n")
}

func (h *Handler) tryWebSearch(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
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

	aiBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiBaseURL == "" {
		aiBaseURL = "http://localhost:18081"
	}
	aiAPIKey := os.Getenv("AI_SERVICE_API_KEY")

	toolReq := map[string]interface{}{
		"tool": "web_search",
		"params": map[string]interface{}{
			"query":       message,
			"max_results": 5,
		},
		"agent_id":  agent.AgentID,
		"tenant_id": claims.TenantID.String(),
	}

	body, _ := json.Marshal(toolReq)
	req, err := http.NewRequestWithContext(ctx, "POST", aiBaseURL+"/api/tools/execute", bytes.NewBuffer(body))
	if err != nil {
		return h.executeSearchDirect(ctx, message)
	}
	req.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		req.Header.Set("X-API-Key", aiAPIKey)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return h.executeSearchDirect(ctx, message)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
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

func (h *Handler) tryDatabaseQuery(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	queryTriggers := []string{
		"query", "select", "show me", "count", "list all", "how many",
		"database", "table", "sql", "records", "rows",
	}
	needsQuery := false
	for _, trigger := range queryTriggers {
		if strings.Contains(msg, trigger) {
			needsQuery = true
			break
		}
	}
	if !needsQuery {
		return ""
	}

	if h.rawDB == nil {
		return "database not available"
	}

	lowerMsg := strings.TrimSpace(msg)
	if !strings.HasPrefix(lowerMsg, "select") && !strings.HasPrefix(lowerMsg, "show") {
		return "only SELECT queries are allowed"
	}

	query := strings.TrimSpace(message)
	if idx := strings.Index(strings.ToLower(query), "select"); idx >= 0 {
		query = query[idx:]
	}
	if idx := strings.Index(strings.ToLower(query), "show"); idx >= 0 {
		query = query[idx:]
	}

	upper := strings.ToUpper(query)
	if strings.Contains(upper, "INSERT") || strings.Contains(upper, "UPDATE") ||
		strings.Contains(upper, "DELETE") || strings.Contains(upper, "DROP") ||
		strings.Contains(upper, "ALTER") || strings.Contains(upper, "CREATE") ||
		strings.Contains(upper, "TRUNCATE") {
		return "only read-only queries are allowed"
	}

	rows, err := h.rawDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Sprintf("query error: %s", err.Error())
	}
	defer func() { _ = rows.Close() }()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}

	if len(results) == 0 {
		return "no results"
	}
	if len(results) > 20 {
		results = results[:20]
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return fmt.Sprintf("%d rows returned:\n%s", len(results), string(data))
}

func (h *Handler) tryFileRead(ctx context.Context, msg, message string) string {
	readTriggers := []string{
		"read file", "show file", "cat ", "open file", "get contents",
		"read the", "show the contents", "what's in", "file contents",
	}
	needsRead := false
	for _, trigger := range readTriggers {
		if strings.Contains(msg, trigger) {
			needsRead = true
			break
		}
	}
	if !needsRead {
		return ""
	}

	workspace := os.Getenv("AGENT_WORKSPACE_DIR")
	if workspace == "" {
		workspace = "/tmp/agent-workspace"
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		logrus.WithError(err).Warn("failed to create workspace directory")
	}

	var filePath string
	lower := strings.ToLower(message)
	for _, prefix := range []string{"read file ", "show file ", "cat ", "open file ", "read "} {
		if idx := strings.Index(lower, prefix); idx >= 0 {
			rest := message[idx+len(prefix):]
			rest = strings.TrimSpace(rest)
			rest = strings.Trim(rest, "\"'`")
			if rest != "" {
				filePath = rest
				break
			}
		}
	}
	if filePath == "" {
		return "no file path specified"
	}
	if !strings.HasPrefix(filePath, "/") {
		filePath = workspace + "/" + filePath
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("error reading file: %s", err.Error())
	}
	if len(data) > 10000 {
		data = data[:10000]
		return fmt.Sprintf("file: %s (truncated to 10KB)\n%s", filePath, string(data))
	}
	return fmt.Sprintf("file: %s\n%s", filePath, string(data))
}

func (h *Handler) tryFileWrite(ctx context.Context, msg, message string) string {
	writeTriggers := []string{
		"write file", "create file", "save to", "write to",
		"save file", "generate file", "make file",
	}
	needsWrite := false
	for _, trigger := range writeTriggers {
		if strings.Contains(msg, trigger) {
			needsWrite = true
			break
		}
	}
	if !needsWrite {
		return ""
	}

	workspace := os.Getenv("AGENT_WORKSPACE_DIR")
	if workspace == "" {
		workspace = "/tmp/agent-workspace"
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		logrus.WithError(err).Warn("failed to create workspace directory")
	}

	lower := strings.ToLower(message)
	var filePath string
	for _, prefix := range []string{"write file ", "create file ", "save to ", "write to ", "save file "} {
		if idx := strings.Index(lower, prefix); idx >= 0 {
			rest := message[idx+len(prefix):]
			rest = strings.TrimSpace(rest)
			rest = strings.Trim(rest, "\"'`")
			parts := strings.SplitN(rest, " ", 2)
			if len(parts) > 0 {
				filePath = parts[0]
			}
			break
		}
	}
	if filePath == "" {
		return "no file path specified"
	}
	if !strings.HasPrefix(filePath, "/") {
		filePath = workspace + "/" + filePath
	}

	dir := filePath[:strings.LastIndex(filePath, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Sprintf("error creating directory: %s", err.Error())
	}

	content := message
	if idx := strings.Index(message, "content:"); idx >= 0 {
		content = strings.TrimSpace(message[idx+8:])
	} else if idx := strings.Index(message, "Content:"); idx >= 0 {
		content = strings.TrimSpace(message[idx+8:])
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Sprintf("error writing file: %s", err.Error())
	}
	return fmt.Sprintf("file written: %s (%d bytes)", filePath, len(content))
}

func (h *Handler) tryHTTPRequest(ctx context.Context, msg, message string) string {
	httpTriggers := []string{
		"http request", "fetch url", "api call", "hit endpoint",
		"get request", "post request", "call api", "request to",
		"curl", "fetch from",
	}
	needsHTTP := false
	for _, trigger := range httpTriggers {
		if strings.Contains(msg, trigger) {
			needsHTTP = true
			break
		}
	}
	if !needsHTTP {
		return ""
	}

	var url string
	lower := strings.ToLower(message)
	for _, prefix := range []string{"http request to ", "fetch url ", "api call to ", "get ", "post ", "request to ", "curl ", "fetch from "} {
		if idx := strings.Index(lower, prefix); idx >= 0 {
			rest := message[idx+len(prefix):]
			rest = strings.TrimSpace(rest)
			rest = strings.Trim(rest, "\"'`")
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				url = parts[0]
			}
			break
		}
	}
	if url == "" {
		for _, word := range strings.Fields(message) {
			if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
				url = strings.Trim(word, "\"'`,.")
				break
			}
		}
	}
	if url == "" {
		return "no URL found in message"
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	method := "GET"
	if strings.Contains(msg, "post ") || strings.Contains(msg, "post request") {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Sprintf("error creating request: %s", err.Error())
	}
	req.Header.Set("User-Agent", "FunctionFly-Agent/1.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("request error: %s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 5000 {
		body = body[:5000]
	}
	return fmt.Sprintf("HTTP %s %s -> %d\n%s", method, url, resp.StatusCode, string(body))
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

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

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

func (h *Handler) tryCodeExecution(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	triggers := []string{"run code", "execute code", "run python", "run node", "run script", "code execution", "run this", "execute this"}
	needsExec := false
	for _, t := range triggers {
		if strings.Contains(msg, t) {
			needsExec = true
			break
		}
	}
	if !needsExec {
		return ""
	}

	result := h.toolRegistry.ExecuteTool(ctx, "code_execution", map[string]interface{}{
		"language": "python",
		"code":     message,
	}, tools.ExecutionContext{
		AgentID:  agent.AgentID,
		TenantID: claims.TenantID.String(),
	})
	if result == nil || !result.OK {
		return ""
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

func (h *Handler) tryImageGeneration(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	triggers := []string{"generate image", "create image", "draw", "make an image", "image of", "picture of", "generate a picture"}
	needsGen := false
	for _, t := range triggers {
		if strings.Contains(msg, t) {
			needsGen = true
			break
		}
	}
	if !needsGen {
		return ""
	}

	result := h.toolRegistry.ExecuteTool(ctx, "image_generation", map[string]interface{}{
		"prompt": message,
		"size":   "1024x1024",
	}, tools.ExecutionContext{
		AgentID:  agent.AgentID,
		TenantID: claims.TenantID.String(),
	})
	if result == nil || !result.OK {
		return ""
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

func (h *Handler) tryTextToSpeech(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	triggers := []string{"text to speech", "convert to audio", "speak this", "read aloud", "tts", "say this", "audio of"}
	needsTTS := false
	for _, t := range triggers {
		if strings.Contains(msg, t) {
			needsTTS = true
			break
		}
	}
	if !needsTTS {
		return ""
	}

	result := h.toolRegistry.ExecuteTool(ctx, "text_to_speech", map[string]interface{}{
		"text":  message,
		"voice": "alloy",
	}, tools.ExecutionContext{
		AgentID:  agent.AgentID,
		TenantID: claims.TenantID.String(),
	})
	if result == nil || !result.OK {
		return ""
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

func (h *Handler) tryEmailTool(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	triggers := []string{"send email", "send mail", "email to", "email notification"}
	needsEmail := false
	for _, t := range triggers {
		if strings.Contains(msg, t) {
			needsEmail = true
			break
		}
	}
	if !needsEmail {
		return ""
	}

	result := h.toolRegistry.ExecuteTool(ctx, "email", map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Agent notification",
		"body":    message,
	}, tools.ExecutionContext{
		AgentID:  agent.AgentID,
		TenantID: claims.TenantID.String(),
	})
	if result == nil || !result.OK {
		return ""
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

func (h *Handler) tryNotificationTool(ctx context.Context, msg, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	triggers := []string{"send notification", "notify", "push notification", "alert me", "send alert"}
	needsNotify := false
	for _, t := range triggers {
		if strings.Contains(msg, t) {
			needsNotify = true
			break
		}
	}
	if !needsNotify {
		return ""
	}

	result := h.toolRegistry.ExecuteTool(ctx, "notification", map[string]interface{}{
		"title":    "Agent Notification",
		"message":  message,
		"severity": "info",
	}, tools.ExecutionContext{
		AgentID:  agent.AgentID,
		TenantID: claims.TenantID.String(),
	})
	if result == nil || !result.OK {
		return ""
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

func extractRegionFromHealthMessage(healthMessage string) string {
	const prefix = "region:"
	if len(healthMessage) > len(prefix) && healthMessage[:len(prefix)] == prefix {
		return healthMessage[len(prefix):]
	}
	return ""
}

type workspaceEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	ModTime  string `json:"mod_time"`
}

// HandleBrowseWorkspace lists files in the agent's workspace.
// GET /v1/agent/{agent_id}/workspace?path=
func (h *Handler) HandleBrowseWorkspace(w http.ResponseWriter, r *http.Request) {
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

	relPath := r.URL.Query().Get("path")
	base, _ := tools.EnsureWorkspace(agent.AgentID, claims.TenantID.String())

	target := base
	if relPath != "" {
		clean := filepath.Clean("/" + relPath)
		target = filepath.Join(base, clean)
		if !strings.HasPrefix(target, base) {
			writeError(w, http.StatusBadRequest, "INVALID_PATH", "path escapes workspace")
			return
		}
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	var result []workspaceEntry
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(base, filepath.Join(target, e.Name()))
		result = append(result, workspaceEntry{
			Name:    e.Name(),
			Path:    "/" + rel,
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"path":     "/" + relPath,
		"workspace": base,
		"entries":  result,
	})
}

// HandleWorkspaceManifest returns the workspace manifest.
// GET /v1/agent/{agent_id}/workspace/manifest
func (h *Handler) HandleWorkspaceManifest(w http.ResponseWriter, r *http.Request) {
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

	base, _ := tools.EnsureWorkspace(agent.AgentID, claims.TenantID.String())
	manifestPath := filepath.Join(base, tools.ManifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "manifest not found")
		return
	}

	var manifest interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		logrus.WithError(err).Warn("failed to unmarshal manifest")
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"manifest": manifest,
	})
}

// HandleWorkspaceHistory returns the workspace execution history.
// GET /v1/agent/{agent_id}/workspace/history
func (h *Handler) HandleWorkspaceHistory(w http.ResponseWriter, r *http.Request) {
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

	base, _ := tools.EnsureWorkspace(agent.AgentID, claims.TenantID.String())
	historyPath := filepath.Join(base, tools.HistoryFile)
	data, err := os.ReadFile(historyPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "history": []interface{}{}})
		return
	}

	var entries []interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var entry interface{}
		if json.Unmarshal([]byte(line), &entry) == nil {
			entries = append(entries, entry)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"history": entries,
	})
}
