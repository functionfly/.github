package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// aiServiceLikelyDown returns true when the error is a typical "nothing listening" / network issue,
// as opposed to a 4xx/5xx or JSON error from a running ai-service.
func aiServiceLikelyDown(err error) bool {
	if err == nil {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
	}
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such host") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "connection reset by peer")
}

// AIChatClientConfig holds configuration for the AI chat client
type AIChatClientConfig struct {
	BaseURL     string        // AI service base URL (e.g., http://localhost:8000)
	APIKey      string        // API key for authentication
	Timeout     time.Duration // Request timeout
	Model       string        // Model to use
	MaxTokens   int           // Max tokens in response
	Temperature float64       // Temperature for randomness
	Enabled     bool          // Whether AI is enabled
}

// DefaultAIChatClientConfig returns default configuration
func DefaultAIChatClientConfig() *AIChatClientConfig {
	return &AIChatClientConfig{
		BaseURL:     "", // Must be set via AI_SERVICE_URL env var
		APIKey:      "",
		Timeout:     30 * time.Second,
		Model:       "gpt-4o-mini",
		MaxTokens:   500,
		Temperature: 0.7,
		Enabled:     true,
	}
}

// AIServiceClient implements AIChatClient by calling the ai-service REST API
type AIServiceClient struct {
	config AIChatClientConfig
	client *http.Client
	logger *logrus.Logger
}

// NewAIServiceClient creates a new AI service client
func NewAIServiceClient(config *AIChatClientConfig, logger *logrus.Logger) *AIServiceClient {
	if config == nil {
		config = DefaultAIChatClientConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &AIServiceClient{
		config: *config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// chatRequest represents the request body for the ai-service chat endpoint
type chatRequest struct {
	SessionID string            `json:"session_id"`
	Message  string            `json:"message"`
	TenantID string            `json:"tenant_id,omitempty"`
	UserID   string            `json:"user_id,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

// chatResponse represents the response from the ai-service chat endpoint
type chatResponse struct {
	Message    string  `json:"message"`
	SessionID  string  `json:"session_id"`
	Model     string  `json:"model,omitempty"`
	Tokens    int     `json:"tokens,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Intent    string  `json:"intent,omitempty"`
}

// GenerateSupportResponse generates a support response using the AI service
// This implements the AIChatClient interface
func (c *AIServiceClient) GenerateSupportResponse(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("AI support is disabled")
	}

	// Build context map for the AI service
	contextMap := make(map[string]string)
	if req.Context != nil {
		if req.Context.FunctionCode != "" {
			contextMap["function_code"] = req.Context.FunctionCode
		}
		if req.Context.DeploymentError != "" {
			contextMap["deployment_error"] = req.Context.DeploymentError
		}
		if len(req.Context.FunctionLogs) > 0 {
			contextMap["function_logs"] = fmt.Sprintf("%v", req.Context.FunctionLogs)
		}
		if len(req.Context.EnvironmentVars) > 0 {
			// Mask sensitive env vars
			envVars := make(map[string]string)
			for k, v := range req.Context.EnvironmentVars {
				if isSensitiveEnvVar(k) {
					envVars[k] = "***REDACTED***"
				} else {
					envVars[k] = v
				}
			}
			contextMap["environment_vars"] = fmt.Sprintf("%v", envVars)
		}
		if req.Context.UserInfo != nil {
			contextMap["user_plan"] = req.Context.UserInfo.Plan
			contextMap["user_email"] = req.Context.UserInfo.Email
		}
	}

	// Build conversation history as part of the message
	message := req.UserMessage
	if len(req.History) > 0 {
		history := "Previous conversation:\n"
		for _, msg := range req.History {
			role := "User"
			if msg.AuthorType == AuthorAI {
				role = "Assistant"
			} else if msg.AuthorType == AuthorStaff {
				role = "Staff"
			}
			history += fmt.Sprintf("%s: %s\n", role, msg.Content)
		}
		message = history + "\nCurrent message: " + req.UserMessage
	}

	// Create chat message request for ai-service
	chatReq := chatRequest{
		SessionID: req.ConversationID.String(),
		Message:   message,
		Context:   contextMap,
	}
	if req.UserID != uuid.Nil {
		chatReq.UserID = req.UserID.String()
	}

	// Serialize request
	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/chat/message", c.config.BaseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	c.logger.WithFields(logrus.Fields{
		"conversation_id": req.ConversationID,
		"ai_url":          c.config.BaseURL,
	}).Debug("Calling ai-service for support response")

	// Execute request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		// Avoid WARN here: AIGatewayClient logs one clear INFO/WARN when applying fallback.
		c.logger.WithFields(logrus.Fields{
			"ai_url": c.config.BaseURL,
			"error":  err.Error(),
		}).Debug("ai-service HTTP request failed")
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.logger.WithField("status", resp.StatusCode).Warn("ai-service returned non-200 status")
		return nil, fmt.Errorf("ai-service returned status %d", resp.StatusCode)
	}

	// Decode response
	var aiResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		c.logger.WithError(err).Warn("Failed to decode ai-service response")
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	// Build response matching the AIResponse interface
	supportResp := &AIResponse{
		Message:    aiResp.Message,
		Confidence: aiResp.Confidence,
		Model:      aiResp.Model,
		Actions:    []string{},
	}

	// Add suggested actions based on intent
	switch aiResp.Intent {
	case "deployment":
		supportResp.Actions = []string{
			"Check deployment logs",
			"Verify environment variables",
			"Review function configuration",
		}
	case "runtime", "error":
		supportResp.Actions = []string{
			"View execution logs",
			"Check function timeout settings",
			"Verify memory allocation",
		}
	case "configuration", "env":
		supportResp.Actions = []string{
			"Set environment variable",
			"Add secret to Vault",
			"View current variables",
		}
	case "billing":
		supportResp.Actions = []string{
			"View billing dashboard",
			"Contact billing team",
		}
	default:
		supportResp.Actions = []string{
			"View documentation",
			"Contact support",
		}
	}

	c.logger.WithFields(logrus.Fields{
		"conversation_id": req.ConversationID,
		"confidence":      aiResp.Confidence,
		"intent":          aiResp.Intent,
	}).Debug("Received AI response from ai-service")

	return supportResp, nil
}

// isSensitiveEnvVar checks if an environment variable name suggests it contains secrets
func isSensitiveEnvVar(name string) bool {
	sensitivePrefixes := []string{
		"api_key", "apikey", "secret", "password", "passwd", "token",
		"auth", "credential", "private", "access_key", "aws_",
	}
	// Convert to lowercase for comparison
	nameLower := []rune(name)
	for i := range nameLower {
		if nameLower[i] >= 'A' && nameLower[i] <= 'Z' {
			nameLower[i] = nameLower[i] + 32
		}
	}
	nameStr := string(nameLower)

	for _, prefix := range sensitivePrefixes {
		if len(nameStr) >= len(prefix) && nameStr[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// AIGatewayClient wraps AIServiceClient to provide additional support-specific functionality
type AIGatewayClient struct {
	aiClient *AIServiceClient
	logger   *logrus.Logger
}

// NewAIGatewayClient creates a new AI gateway client for support
func NewAIGatewayClient(config *AIChatClientConfig, logger *logrus.Logger) *AIGatewayClient {
	if logger == nil {
		logger = logrus.New()
	}
	return &AIGatewayClient{
		aiClient: NewAIServiceClient(config, logger),
		logger:   logger,
	}
}

// GenerateSupportResponse implements AIChatClient: tries ai-service, then rule-based fallback.
func (c *AIGatewayClient) GenerateSupportResponse(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	return c.GenerateWithFallback(ctx, req)
}

// GenerateWithFallback generates an AI response with fallback to rule-based responses
func (c *AIGatewayClient) GenerateWithFallback(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	// Try AI service first
	resp, err := c.aiClient.GenerateSupportResponse(ctx, req)
	if err != nil {
		if aiServiceLikelyDown(err) {
			c.logger.WithFields(logrus.Fields{
				"ai_url": c.aiClient.config.BaseURL,
				"error":  err.Error(),
			}).Info("ai-service not reachable; using built-in rule-based assistant (start FlyMind / ai-service or set AI_SERVICE_URL)")
		} else {
			c.logger.WithError(err).Warn("AI service returned an error; using rule-based fallback")
		}

		// Use rule-based fallback from ai.go
		fallbackReq := &AIReplyRequest{
			ConversationID: req.ConversationID,
			UserID:         req.UserID,
			UserMessage:    req.UserMessage,
			Context: &SupportContext{
				FunctionCode:     req.Context.FunctionCode,
				FunctionLogs:     req.Context.FunctionLogs,
				DeploymentError:  req.Context.DeploymentError,
				EnvironmentVars:  req.Context.EnvironmentVars,
				ExecutionHistory: req.Context.ExecutionHistory,
				UserInfo:         req.Context.UserInfo,
			},
			Category:            "general",
			ConversationHistory: convertHistory(req.History),
		}

		cfg := LoadAIConfigFromEnv()
		fallbackResp, fallbackErr := GenerateAIReply(ctx, fallbackReq, cfg)
		if fallbackErr != nil {
			return nil, fmt.Errorf("both AI service and fallback failed: AI error: %w, fallback error: %v", err, fallbackErr)
		}

		model := "rule-based"
		if cfg.OpenAIAPIKey != "" {
			model = cfg.Model + " (OpenAI)"
		} else if cfg.AnthropicAPIKey != "" {
			model = cfg.Model + " (Anthropic)"
		}

		return &AIResponse{
			Message:    fallbackResp.Message,
			Confidence: fallbackResp.Confidence,
			Model:      model,
			Actions:    fallbackResp.SuggestedActions,
		}, nil
	}

	return resp, nil
}

// convertHistory converts []*SupportMessage to []SupportMessage for the fallback
func convertHistory(history []*SupportMessage) []SupportMessage {
	result := make([]SupportMessage, len(history))
	for i, msg := range history {
		result[i] = *msg
	}
	return result
}
