//go:build cgo

package wasm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AIInferenceRequest represents a request to the AI Gateway for inference
type AIInferenceRequest struct {
	// Model is the AI model to use (e.g., "gpt-4", "claude-3")
	Model string `json:"model"`

	// Input is the text or data to send to the model
	Input json.RawMessage `json:"input"`

	// Parameters are model-specific parameters (temperature, max_tokens, etc.)
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Streaming indicates whether to use streaming responses
	Streaming bool `json:"streaming,omitempty"`
}

// AIInferenceResponse represents a response from the AI Gateway
type AIInferenceResponse struct {
	// Output is the model's response
	Output json.RawMessage `json:"output"`

	// LatencyMs is the inference latency in milliseconds
	LatencyMs int64 `json:"latency_ms"`

	// Cost is the estimated cost in USD
	Cost float64 `json:"cost"`

	// ModelUsed is the actual model that processed the request
	ModelUsed string `json:"model_used,omitempty"`

	// Error contains error information if the inference failed
	Error *AIInferenceError `json:"error,omitempty"`
}

// AIInferenceError represents an error from AI inference
type AIInferenceError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`
}

// AISignatureResult represents a code generation response from AI inference
// Used when the AI is generating code, signatures, or structured outputs
type AISignatureResult struct {
	// Language is the programming language of the generated code
	Language string `json:"language,omitempty"`

	// Code is the generated code
	Code string `json:"code,omitempty"`

	// Signature is the function/signature extracted from the code
	Signature string `json:"signature,omitempty"`

	// Documentation is any associated documentation
	Documentation string `json:"documentation,omitempty"`
}

// ValidateAIInferenceRequest validates an AI inference request
func ValidateAIInferenceRequest(req *AIInferenceRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Input) == 0 {
		return fmt.Errorf("input is required")
	}
	if len(req.Input) > 10*1024*1024 { // 10MB max input
		return fmt.Errorf("input too large: %d bytes (max 10MB)", len(req.Input))
	}
	return nil
}

// AIInference makes an inference request to the AI Gateway
// This is the core function that WASM modules call via the ai_infer host function
func AIInference(ctx context.Context, config *AIInferenceConfig, model string, input []byte, params string) (*AIInferenceResponse, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("ai inference is not enabled")
	}

	// Use default model if not specified
	if model == "" {
		model = config.DefaultModel
	}

	// Parse parameters if provided
	var parameters map[string]interface{}
	if params != "" {
		if err := json.Unmarshal([]byte(params), &parameters); err != nil {
			return nil, fmt.Errorf("invalid parameters JSON: %w", err)
		}
	}

	// Build request
	aiReq := AIInferenceRequest{
		Model:      model,
		Input:      json.RawMessage(input),
		Parameters: parameters,
		Streaming:  false, // Non-streaming for simpler response handling
	}

	// Validate request
	if err := ValidateAIInferenceRequest(&aiReq); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Serialize request
	reqBody, err := json.Marshal(aiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := config.GatewayURL + "/v1/inference"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Set timeout
	timeout := time.Duration(config.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	latencyMs := time.Since(startTime).Milliseconds()

	// Read response body
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, int64(config.MaxModelSizeMB)*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &AIInferenceResponse{
			LatencyMs: latencyMs,
			Error: &AIInferenceError{
				Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
				Message: fmt.Sprintf("AI Gateway returned status %d: %s", resp.StatusCode, string(respBody)),
			},
		}, fmt.Errorf("AI Gateway returned status %d", resp.StatusCode)
	}

	// Parse response
	var aiResp AIInferenceResponse
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	aiResp.LatencyMs = latencyMs

	return &aiResp, nil
}

// LogAIInference logs an AI inference call for audit purposes
func LogAIInference(model string, inputSize int, outputSize int, latencyMs int64, cost float64, err error) {
	if err != nil {
		log.Printf("[AI Inference] model=%s input_size=%d output_size=%d latency_ms=%d cost=%.6f error=%v",
			model, inputSize, outputSize, latencyMs, cost, err)
	} else {
		log.Printf("[AI Inference] model=%s input_size=%d output_size=%d latency_ms=%d cost=%.6f success=true",
			model, inputSize, outputSize, latencyMs, cost)
	}
}
