package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AIConfig holds configuration for the AI support service
type AIConfig struct {
	Enabled           bool
	AIProvider        string // "openai", "anthropic", "ollama"
	Model             string
	MaxTokens         int
	Temperature       float64
	AllowedCategories []string // Categories AI can help with
	OpenAIAPIKey      string
	AnthropicAPIKey   string
	OpenAIBaseURL     string
	AnthropicBaseURL  string
}

// DefaultAIConfig returns default AI configuration
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		Enabled:     true,
		AIProvider:  "openai",
		Model:       "gpt-4o-mini",
		MaxTokens:   500,
		Temperature: 0.7,
		AllowedCategories: []string{
			"deployment",
			"configuration",
			"runtime",
			"logs",
			"errors",
			"api",
			"functions",
			"billing",
			"general",
		},
		OpenAIBaseURL:    "https://api.openai.com/v1",
		AnthropicBaseURL: "https://api.anthropic.com/v1",
	}
}

// LoadAIConfigFromEnv loads AI configuration from environment variables
func LoadAIConfigFromEnv() *AIConfig {
	cfg := DefaultAIConfig()
	if v := os.Getenv("AI_SUPPORT_PROVIDER"); v != "" {
		cfg.AIProvider = v
	}
	if v := os.Getenv("AI_SUPPORT_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAIAPIKey = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		cfg.OpenAIBaseURL = v
	}
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		cfg.AnthropicBaseURL = v
	}
	return cfg
}

// AIReplyRequest represents a request for AI-generated support reply
type AIReplyRequest struct {
	ConversationID     uuid.UUID
	UserID             uuid.UUID
	UserMessage        string
	Context            *SupportContext
	Category           string
	ConversationHistory []SupportMessage
}

// AIReplyResponse represents AI-generated response
type AIReplyResponse struct {
	Message          string
	Category         string
	Confidence       float64
	ShouldEscalate   bool
	SuggestedActions []string
}

// GenerateAIReply generates an AI reply for a support conversation.
// When AI API keys are configured, it calls OpenAI/Anthropic directly.
// Otherwise, it uses rule-based pattern matching as a lightweight fallback.
// This is the final fallback called by AIGatewayClient when ai-service is unavailable.
func GenerateAIReply(ctx context.Context, req *AIReplyRequest, cfg *AIConfig) (*AIReplyResponse, error) {
	if cfg == nil {
		cfg = DefaultAIConfig()
	}

	// Try OpenAI first
	if cfg.Enabled && cfg.OpenAIAPIKey != "" {
		resp, err := generateOpenAIReply(ctx, req, cfg)
		if err == nil {
			return resp, nil
		}
		fmt.Printf("OpenAI fallback failed: %v\n", err)
	}

	// Try Anthropic second
	if cfg.Enabled && cfg.AnthropicAPIKey != "" {
		resp, err := generateAnthropicReply(ctx, req, cfg)
		if err == nil {
			return resp, nil
		}
		fmt.Printf("Anthropic fallback failed: %v\n", err)
	}

	// Rule-based fallback when no AI keys configured or AI calls fail
	return generateRuleBasedReply(req), nil
}

// generateOpenAIReply calls OpenAI API for support response
func generateOpenAIReply(ctx context.Context, req *AIReplyRequest, cfg *AIConfig) (*AIReplyResponse, error) {
	systemPrompt := `You are FunctionFly Support Assistant, helping developers with their serverless function platform. Be helpful, concise, and technical. Provide specific solutions when possible. Suggest escalating to human agent if issue is complex. Focus on actionable advice.`

	payload := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": BuildSupportPrompt(req)},
		},
		"max_tokens": cfg.MaxTokens,
		"temperature": cfg.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", cfg.OpenAIBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("OpenAI returned %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &AIReplyResponse{
		Message:   openAIResp.Choices[0].Message.Content,
		Confidence: 0.9,
		Category: req.Category,
		SuggestedActions: []string{
			"View documentation",
			"Contact support",
		},
	}, nil
}

// generateAnthropicReply calls Anthropic API for support response
func generateAnthropicReply(ctx context.Context, req *AIReplyRequest, cfg *AIConfig) (*AIReplyResponse, error) {
	payload := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "user", "content": BuildSupportPrompt(req)},
		},
		"max_tokens": cfg.MaxTokens,
		"temperature": cfg.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", cfg.AnthropicBaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", cfg.AnthropicAPIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("Anthropic returned %d: %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	return &AIReplyResponse{
		Message:   anthropicResp.Content[0].Text,
		Confidence: 0.9,
		Category: req.Category,
		SuggestedActions: []string{
			"View documentation",
			"Contact support",
		},
	}, nil
}

// generateRuleBasedReply provides contextual responses using rule-based pattern matching.
// This is used when no AI API keys are configured or when AI calls fail.
func generateRuleBasedReply(req *AIReplyRequest) *AIReplyResponse {
	response := &AIReplyResponse{
		Category:         req.Category,
		Confidence:       0.85,
		ShouldEscalate:   false,
		SuggestedActions: []string{},
	}

	message := strings.ToLower(req.UserMessage)

	if containsAny(message, "deploy", "deployment", "failed", "fail") {
		response.Message = handleDeploymentIssue(req)
		response.Category = "deployment"
		response.SuggestedActions = []string{
			"Check deployment logs",
			"Verify environment variables",
			"Review function configuration",
		}
	} else if containsAny(message, "error", "exception", "crash", "panic") {
		response.Message = handleRuntimeError(req)
		response.Category = "runtime"
		response.SuggestedActions = []string{
			"View execution logs",
			"Check function timeout settings",
			"Verify memory allocation",
		}
	} else if containsAny(message, "memory", "out of memory", "oom") {
		response.Message = "I see you're experiencing memory issues with your function. Here are some steps:\n\n1. **Check your memory limit**: Go to your function settings and verify the allocated memory.\n2. **Optimize your code**: Look for memory leaks or inefficient data structures.\n3. **Use streaming**: For large datasets, consider streaming instead of loading all data in memory.\n\nWould you like me to help you optimize your function's memory usage?"
		response.Category = "runtime"
		response.SuggestedActions = []string{
			"Increase memory limit",
			"Review code for memory leaks",
			"Enable memory profiling",
		}
	} else if containsAny(message, "timeout", "slow", "taking long") {
		response.Message = "Your function appears to be timing out. Consider:\n\n1. **Increase timeout**: Check your function's timeout setting\n2. **Optimize logic**: Look for inefficient loops or external API calls\n3. **Add caching**: Cache frequently accessed data\n\nWould you like me to check your function's current timeout settings?"
		response.Category = "runtime"
		response.SuggestedActions = []string{
			"Increase function timeout",
			"Profile function performance",
			"Add caching layer",
		}
	} else if containsAny(message, "env", "environment", "variable", "secret") {
		response.Message = "I can help with environment variables and secrets. Here's what you need to know:\n\n1. **Set variables in dashboard**: Go to your function → Settings → Environment Variables\n2. **Use Vault for secrets**: Secrets are encrypted and stored in Vault\n3. **Access in code**: Environment variables are available via `process.env` or `os.environ`\n\nWhat specific issue are you encountering?"
		response.Category = "configuration"
		response.SuggestedActions = []string{
			"Set environment variable",
			"Add secret to Vault",
			"View current variables",
		}
	} else if containsAny(message, "how", "what", "can i", "how do") {
		response.Message = handleHowToQuestion(req)
		response.Category = "general"
		response.SuggestedActions = []string{
			"View documentation",
			"Check API reference",
		}
	} else {
		response.Message = fmt.Sprintf("Thank you for reaching out! I understand you're asking about: %s\n\nI'll help you with this. Could you provide more details about:\n- What you're trying to accomplish\n- Any error messages you're seeing\n- What you've already tried\n\nThis will help me give you a more accurate solution.", req.Category)
		response.Category = "general"
		response.SuggestedActions = []string{
			"Describe your issue in detail",
			"Share relevant logs or code",
		}
	}

	// Check if we should escalate based on confidence
	if response.Confidence < 0.7 {
		response.ShouldEscalate = true
	}

	// Escalate for billing issues
	if containsAny(message, "billing", "payment", "invoice", "charge", "subscription", "plan") {
		response.ShouldEscalate = true
		response.Message = "For billing inquiries, I'll connect you with our billing team who can assist with:\n\n- Account charges and invoices\n- Plan upgrades or downgrades\n- Payment method updates\n- Refund requests\n\nA team member will be with you shortly."
	}

	// Escalate for security concerns
	if containsAny(message, "security", "hack", "breach", "compromise", "unauthorized") {
		response.ShouldEscalate = true
		response.Message = "I understand you have a security concern. This is being flagged for immediate attention from our security team.\n\nPlease stand by while I connect you with a specialist."
		response.SuggestedActions = []string{
			"Change passwords immediately",
			"Review recent activity",
			"Do not share sensitive information",
		}
	}

	return response
}

// BuildSupportPrompt builds a prompt for the AI model
func BuildSupportPrompt(req *AIReplyRequest) string {
	var sb strings.Builder

	sb.WriteString("You are FunctionFly Support Assistant, helping developers with their serverless function platform.\n\n")

	// Add context if available
	if req.Context != nil {
		sb.WriteString("## Context\n")

		if req.Context.FunctionCode != "" {
			sb.WriteString("### Function Code\n")
			sb.WriteString("```\n" + req.Context.FunctionCode + "\n```\n")
		}

		if len(req.Context.FunctionLogs) > 0 {
			sb.WriteString("### Recent Logs\n")
			for i, log := range req.Context.FunctionLogs {
				if i >= 10 {
					break
				}
				sb.WriteString(fmt.Sprintf("- %s\n", log))
			}
		}

		if req.Context.DeploymentError != "" {
			sb.WriteString("### Deployment Error\n")
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n", req.Context.DeploymentError))
		}

		if len(req.Context.ExecutionHistory) > 0 {
			sb.WriteString("### Execution History\n")
			for i, exec := range req.Context.ExecutionHistory {
				if i >= 5 {
					break
				}
				status := "SUCCESS"
				if !exec.Success {
					status = "FAILED: " + exec.Error
				}
				sb.WriteString(fmt.Sprintf("- [%s] %s (duration: %dms)\n", status, exec.Timestamp.Format(time.RFC3339), exec.Duration))
			}
		}

		if req.Context.UserInfo != nil {
			sb.WriteString("### User Info\n")
			sb.WriteString(fmt.Sprintf("- Plan: %s\n", req.Context.UserInfo.Plan))
			sb.WriteString(fmt.Sprintf("- Email: %s\n", req.Context.UserInfo.Email))
		}
	}

	// Add conversation history
	if len(req.ConversationHistory) > 0 {
		sb.WriteString("\n## Conversation History\n")
		for _, msg := range req.ConversationHistory {
			role := "User"
			if msg.AuthorType == AuthorAI {
				role = "Assistant"
			} else if msg.AuthorType == AuthorStaff {
				role = "Staff"
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
		}
	}

	// Add current message
	sb.WriteString(fmt.Sprintf("\n## Current Message\nUser: %s\n", req.UserMessage))

	sb.WriteString("\n## Instructions\n")
	sb.WriteString("- Be helpful, concise, and technical\n")
	sb.WriteString("- Provide specific solutions when possible\n")
	sb.WriteString("- Suggest escalating to human agent if issue is complex\n")
	sb.WriteString("- Focus on actionable advice\n")

	return sb.String()
}

// handleDeploymentIssue generates response for deployment issues
func handleDeploymentIssue(req *AIReplyRequest) string {
	if req.Context != nil && req.Context.DeploymentError != "" {
		return fmt.Sprintf("I can see your deployment failed with the following error:\n\n```\n%s\n```\n\nHere are common causes and solutions:\n\n1. **Build errors**: Check your function code for syntax errors\n2. **Dependency issues**: Verify all dependencies are listed in requirements.txt or package.json\n3. **Invalid configuration**: Check your fly.toml settings\n\nWould you like me to help diagnose the specific error?", req.Context.DeploymentError)
	}

	return "I see you're having trouble with deployment. To help you better, could you tell me:\n\n1. What error message are you seeing?\n2. What runtime are you using?\n3. Did you make any recent changes to your function?"
}

// handleRuntimeError generates response for runtime errors
func handleRuntimeError(req *AIReplyRequest) string {
	if req.Context != nil && len(req.Context.ExecutionHistory) > 0 {
		errors := []string{}
		for _, exec := range req.Context.ExecutionHistory {
			if !exec.Success && exec.Error != "" {
				errors = append(errors, exec.Error)
			}
		}
		if len(errors) > 0 {
			return fmt.Sprintf("I found %d recent error(s) in your function. Here are the most recent:\n\n```\n%s\n```\n\nCommon causes for runtime errors:\n\n1. **Timeout**: Function exceeded its time limit\n2. **Memory exhausted**: Function ran out of memory\n3. **External API failures**: Network timeouts to upstream services\n4. **Code bugs**: Unhandled exceptions in your code\n\nWould you like me to analyze one of these errors in detail?", len(errors), errors[0])
		}
	}

	return "I'm seeing errors in your function. To help diagnose the issue, could you share:\n\n1. The full error message\n2. When the error started occurring\n3. Any recent changes you made"
}

// handleHowToQuestion generates response for how-to questions
func handleHowToQuestion(req *AIReplyRequest) string {
	message := strings.ToLower(req.UserMessage)

	if containsAny(message, "deploy", "deploy function", "deploy my function") {
		return "Here's how to deploy your function on FunctionFly:\n\n```bash\n# Using the CLI\nff deploy\n\n# Or drag and drop in the dashboard\n```\n\nMake sure you have:\n- A valid function file (index.js, index.py, main.go, etc.)\n- Proper configuration in fly.toml\n- All dependencies declared\n\nNeed help with a specific deployment issue?"
	}

	if containsAny(message, "scale", "scaling", "autoscale") {
		return "FunctionFly supports automatic scaling. Here's how to configure it:\n\n1. **Auto-scale**: Enabled by default for Pro plans\n2. **Set limits**: Configure min/max instances in dashboard\n3. **Memory**: Adjust based on your function's needs\n\nFor custom scaling strategies, check the documentation or contact support."
	}

	if containsAny(message, "monitor", "metrics", "logs") {
		return "You can monitor your functions in several ways:\n\n1. **Dashboard**: Real-time metrics in the Functions tab\n2. **CLI**: `ff logs <function-name>`\n3. **API**: Query the metrics endpoint\n\nWhat specific metrics are you looking for?"
	}

	return "Great question! I'd be happy to help. Could you provide more details about what you're trying to accomplish? This will help me give you the most accurate guidance."
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ShouldEscalate determines if a conversation should be escalated to human
func ShouldEscalate(ctx context.Context, conv *SupportConversation, msgCount int) (bool, string) {
	// Escalate if too many AI messages without resolution
	if msgCount > 10 && conv.Status != StatusResolved {
		return true, "Too many AI responses without resolution"
	}

	// Escalate if conversation is old
	if time.Since(conv.CreatedAt) > 30*time.Minute && conv.Status != StatusResolved {
		return true, "Conversation timeout"
	}

	// Escalate critical issues
	if conv.Priority == PriorityCritical {
		return true, "Critical priority"
	}

	return false, ""
}
