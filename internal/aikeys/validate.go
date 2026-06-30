package aikeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ValidationResponse is the result of validating a provider key.
type ValidationResponse struct {
	IsValid bool   `json:"is_valid"`
	Message string `json:"message,omitempty"`
}

// SupportedProviders returns all providers available for BYOK.
func SupportedProviders() []SupportedProvider {
	return []SupportedProvider{
		{ID: "openai", Name: "OpenAI", Description: "GPT-4o, o1, GPT-4.1, and embeddings", KeyFormat: "sk-...", KeyPrefix: "sk-"},
		{ID: "anthropic", Name: "Anthropic", Description: "Claude Sonnet 4.6, Claude 3.5, Opus", KeyFormat: "sk-ant-...", KeyPrefix: "sk-ant-"},
		{ID: "openrouter", Name: "OpenRouter", Description: "Access 100+ models through a single key", KeyFormat: "sk-or-...", KeyPrefix: "sk-or-", IsMetaProvider: true},
		{ID: "fireworks", Name: "Fireworks AI", Description: "Structured output, function calling", KeyFormat: "fw_...", KeyPrefix: "fw_"},
		{ID: "groq", Name: "Groq", Description: "Ultra-low latency LPU inference", KeyFormat: "gsk_...", KeyPrefix: "gsk_"},
		{ID: "deepinfra", Name: "DeepInfra", Description: "Cost-optimized serverless inference", KeyFormat: "Any (min 20 chars)"},
		{ID: "together", Name: "Together AI", Description: "Wide model catalog, batch at 50% off", KeyFormat: "Any (min 20 chars)"},
		{ID: "mimo", Name: "MiMo (Xiaomi)", Description: "Long-context reasoning, 1M context", KeyFormat: "Any (min 10 chars)"},
		{ID: "stepfun", Name: "StepFun", Description: "Reasoning + vision models", KeyFormat: "Any (min 10 chars)"},
	}
}

// ValidateProviderKey validates a key for a specific provider.
// It first checks format, then makes a lightweight API call.
func ValidateProviderKey(provider, apiKey string) ValidationResponse {
	if err := validateFormat(provider, apiKey); err != nil {
		return ValidationResponse{IsValid: false, Message: err.Error()}
	}
	return testProviderAPI(provider, apiKey)
}

// validateFormat checks key format (prefix/length) for fast rejection.
func validateFormat(provider, apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	switch provider {
	case "openai":
		if !strings.HasPrefix(apiKey, "sk-") || len(apiKey) < 20 {
			return fmt.Errorf("OpenAI keys start with 'sk-' and are at least 20 characters")
		}
	case "anthropic":
		if !strings.HasPrefix(apiKey, "sk-ant-") || len(apiKey) < 20 {
			return fmt.Errorf("anthropic keys start with 'sk-ant-' and are at least 20 characters")
		}
	case "openrouter":
		if !strings.HasPrefix(apiKey, "sk-or-") || len(apiKey) < 20 {
			return fmt.Errorf("openrouter keys start with 'sk-or-' and are at least 20 characters")
		}
	case "fireworks":
		if !strings.HasPrefix(apiKey, "fw_") || len(apiKey) < 10 {
			return fmt.Errorf("fireworks keys start with 'fw_' and are at least 10 characters")
		}
	case "groq":
		if !strings.HasPrefix(apiKey, "gsk_") || len(apiKey) < 10 {
			return fmt.Errorf("groq keys start with 'gsk_' and are at least 10 characters")
		}
	case "deepinfra", "together":
		if len(apiKey) < 20 {
			return fmt.Errorf("%s keys must be at least 20 characters", provider)
		}
	case "mimo", "stepfun":
		if len(apiKey) < 10 {
			return fmt.Errorf("%s keys must be at least 10 characters", provider)
		}
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
	return nil
}

// testProviderAPI makes a lightweight API call to verify the key works.
func testProviderAPI(provider, apiKey string) ValidationResponse {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint, method, headers := providerTestConfig(provider, apiKey)

	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return ValidationResponse{IsValid: false, Message: fmt.Sprintf("request error: %v", err)}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ValidationResponse{IsValid: false, Message: fmt.Sprintf("connection failed: %v", err)}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return ValidationResponse{IsValid: true, Message: fmt.Sprintf("%s key validated successfully", provider)}
	}

	// Try to extract error message
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		msg := errResp.Error.Message
		if msg == "" && len(errResp.Errors) > 0 {
			msg = errResp.Errors[0].Message
		}
		if msg != "" {
			return ValidationResponse{IsValid: false, Message: msg}
		}
	}

	return ValidationResponse{IsValid: false, Message: fmt.Sprintf("validation failed (HTTP %d)", resp.StatusCode)}
}

// providerTestConfig returns the endpoint, method, and headers for testing each provider.
func providerTestConfig(provider, apiKey string) (endpoint, method string, headers map[string]string) {
	headers = map[string]string{"Content-Type": "application/json"}

	switch provider {
	case "openai":
		endpoint = "https://api.openai.com/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "anthropic":
		endpoint = "https://api.anthropic.com/v1/models"
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
	case "openrouter":
		endpoint = "https://openrouter.ai/api/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "fireworks":
		endpoint = "https://api.fireworks.ai/inference/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "groq":
		endpoint = "https://api.groq.com/openai/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "deepinfra":
		endpoint = "https://api.deepinfra.com/v1/openai/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "together":
		endpoint = "https://api.together.xyz/v2/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "mimo":
		endpoint = "https://api.mimo.ai/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	case "stepfun":
		endpoint = "https://api.stepfun.com/v1/models"
		headers["Authorization"] = "Bearer " + apiKey
	default:
		endpoint = ""
	}

	return endpoint, "GET", headers
}
