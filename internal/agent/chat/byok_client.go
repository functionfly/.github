package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var providerBaseURLs = map[string]string{
	"openai":           "https://api.openai.com/v1",
	"anthropic":        "https://api.anthropic.com/v1",
	"openrouter":       "https://openrouter.ai/api/v1",
	"groq":             "https://api.groq.com/openai/v1",
	"fireworks":        "https://api.fireworks.ai/inference/v1",
	"deepinfra":        "https://api.deepinfra.com/v1/openai",
	"together":         "https://api.together.xyz/v2",
	"mimo":             "https://api.mimo.ai/v1",
	"stepfun":          "https://api.stepfun.com/v1",
	"minimax":          "https://api.minimax.io/v1",
	"minimax-token-plan": "https://api.minimax.io/v1",
}

// ProviderFromModel maps a model name to its primary provider.
func ProviderFromModel(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4"):
		return "openai"
	case strings.HasPrefix(m, "claude-"):
		return "anthropic"
	case strings.HasPrefix(m, "groq/") || strings.Contains(m, "llama") || strings.Contains(m, "mixtral"):
		return "groq"
	case strings.HasPrefix(m, "minimax"):
		return "minimax"
	case strings.HasPrefix(m, "mimo") || strings.HasPrefix(m, "xiaomi"):
		return "mimo"
	case strings.HasPrefix(m, "step-") || strings.HasPrefix(m, "stepfun"):
		return "stepfun"
	default:
		if strings.Contains(m, "/") {
			return "openrouter"
		}
		return ""
	}
}

// BYOKRequest holds the parameters for a direct LLM call.
type BYOKRequest struct {
	Provider     string
	APIKey       string
	BaseURL      string
	Model        string
	SystemPrompt string
	UserMessage  string
	MaxTokens    int
	Temperature  float64
	Thinking     *ThinkingConfig
}

// ThinkingConfig configures provider-native thinking/reasoning.
type ThinkingConfig struct {
	Mode         string `json:"mode"`          // off | auto | always
	BudgetTokens int    `json:"budget_tokens"` // max tokens for thinking
}

// ThinkingContent holds thinking output from the LLM.
type ThinkingContent struct {
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

// BYOKResponse is the parsed LLM response.
type BYOKResponse struct {
	Content          string           `json:"content"`
	Model            string           `json:"model"`
	ThinkingContent  *ThinkingContent `json:"thinking_content,omitempty"`
	PromptTokens     int              `json:"prompt_tokens,omitempty"`
	CompletionTokens int              `json:"completion_tokens,omitempty"`
	TotalTokens      int              `json:"total_tokens,omitempty"`
	ReasoningTokens  int              `json:"reasoning_tokens,omitempty"`
}

// CallLLM calls the LLM provider directly using the BYOK key.
// Returns the assistant's reply text.
func CallLLM(ctx context.Context, req BYOKRequest) (*BYOKResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 1024
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	baseURL := req.BaseURL
	if baseURL == "" {
		base, ok := providerBaseURLs[req.Provider]
		if !ok {
			return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
		}
		baseURL = base
	}

	if req.Provider == "anthropic" {
		return callAnthropic(ctx, req, baseURL)
	}
	return callOpenAICompatible(ctx, req, baseURL)
}

func callOpenAICompatible(ctx context.Context, req BYOKRequest, baseURL string) (*BYOKResponse, error) {
	payload := map[string]interface{}{
		"model":       req.Model,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserMessage},
		},
	}

	// Apply thinking/reasoning for o-series models
	if req.Thinking != nil && req.Thinking.Mode != "off" {
		modelLower := strings.ToLower(req.Model)
		isReasoning := strings.HasPrefix(modelLower, "o1") || strings.HasPrefix(modelLower, "o3") || strings.HasPrefix(modelLower, "o4")
		if isReasoning {
			delete(payload, "temperature")
			if req.Thinking.BudgetTokens >= 20000 {
				payload["reasoning_effort"] = "high"
			} else if req.Thinking.BudgetTokens >= 10000 {
				payload["reasoning_effort"] = "medium"
			} else {
				payload["reasoning_effort"] = "low"
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, string(errBody))
	}

	var openAIResp struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content,omitempty"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			CompletionTokensDetails struct {
				ReasoningTokens int `json:"reasoning_tokens"`
			} `json:"completion_tokens_details"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from provider")
	}

	result := &BYOKResponse{
		Content:          openAIResp.Choices[0].Message.Content,
		Model:            openAIResp.Model,
		PromptTokens:     openAIResp.Usage.PromptTokens,
		CompletionTokens: openAIResp.Usage.CompletionTokens,
		TotalTokens:      openAIResp.Usage.TotalTokens,
	}

	// Extract thinking content from DeepSeek R1 reasoning_content
	if openAIResp.Choices[0].Message.ReasoningContent != "" {
		result.ThinkingContent = &ThinkingContent{
			Content: openAIResp.Choices[0].Message.ReasoningContent,
		}
	}

	// Extract reasoning tokens from OpenAI o-series
	if openAIResp.Usage.CompletionTokensDetails.ReasoningTokens > 0 {
		result.ReasoningTokens = openAIResp.Usage.CompletionTokensDetails.ReasoningTokens
		if result.ThinkingContent == nil {
			result.ThinkingContent = &ThinkingContent{}
		}
		result.ThinkingContent.Tokens = openAIResp.Usage.CompletionTokensDetails.ReasoningTokens
	}

	return result, nil
}

func callAnthropic(ctx context.Context, req BYOKRequest, baseURL string) (*BYOKResponse, error) {
	payload := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"system":     req.SystemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": req.UserMessage},
		},
	}

	// Apply extended thinking
	if req.Thinking != nil && req.Thinking.Mode != "off" {
		payload["thinking"] = map[string]interface{}{
			"type":         "enabled",
			"budget_tokens": req.Thinking.BudgetTokens,
		}
		payload["temperature"] = 1.0
	} else {
		payload["temperature"] = req.Temperature
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, string(errBody))
	}

	var anthropicResp struct {
		Model   string `json:"model"`
		Content []struct {
			Type     string `json:"type"`
			Thinking string `json:"thinking,omitempty"`
			Text     string `json:"text,omitempty"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	result := &BYOKResponse{
		Model:            anthropicResp.Model,
		PromptTokens:     anthropicResp.Usage.InputTokens,
		CompletionTokens: anthropicResp.Usage.OutputTokens,
		TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
	}

	for _, block := range anthropicResp.Content {
		switch block.Type {
		case "thinking":
			if block.Thinking != "" {
				result.ThinkingContent = &ThinkingContent{
					Content: block.Thinking,
				}
			}
		case "text":
			result.Content = block.Text
		}
	}

	// If no explicit text block found, use first content block
	if result.Content == "" && len(anthropicResp.Content) > 0 {
		if anthropicResp.Content[0].Text != "" {
			result.Content = anthropicResp.Content[0].Text
		}
	}

	return result, nil
}
