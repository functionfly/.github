package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type AIServiceClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *logrus.Logger
}

func NewAIServiceClient(baseURL, apiKey string, logger *logrus.Logger) *AIServiceClient {
	if logger == nil {
		logger = logrus.New()
	}
	if baseURL == "" {
		baseURL = "http://localhost:18081"
	}
	return &AIServiceClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger,
	}
}

type AIServiceRequest struct {
	SessionID string            `json:"session_id"`
	Message  string            `json:"message"`
	History  []ChatMessage     `json:"history,omitempty"`
	Model    string            `json:"model,omitempty"`
	TenantID string            `json:"tenant_id,omitempty"`
	UserID   string            `json:"user_id,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

type AIServiceResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Intent    string `json:"intent,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Model     string `json:"model,omitempty"`
	Usage     struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func (c *AIServiceClient) ChatMessage(ctx context.Context, req *AIServiceRequest) (*AIServiceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat/message", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	c.logger.WithFields(logrus.Fields{
		"ai_url":  c.baseURL,
		"session": req.SessionID,
	}).Debug("Calling ai-service for chat response")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.WithError(err).Warn("AI service unavailable, using fallback")
		return &AIServiceResponse{
			Message: "I'm having trouble connecting to my AI service. Please try again.",
			Model:   req.Model,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &AIServiceResponse{
			Message: fmt.Sprintf("AI service returned status %d", resp.StatusCode),
			Model:   req.Model,
		}, nil
	}

	var aiResp AIServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		c.logger.WithError(err).Warn("Failed to decode AI response")
		return &AIServiceResponse{
			Message: "I apologize, but I encountered an error processing your request.",
			Model:   req.Model,
		}, nil
	}

	if aiResp.Model == "" {
		aiResp.Model = req.Model
	}

	c.logger.WithFields(logrus.Fields{
		"session_id":  aiResp.SessionID,
		"intent":     aiResp.Intent,
		"confidence": aiResp.Confidence,
	}).Debug("AI response received")

	return &aiResp, nil
}

func (c *AIServiceClient) GenerateFunctionCode(ctx context.Context, req *CreateFunctionRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/composer/generate", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call composer service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("composer service returned status %d", resp.StatusCode)
	}

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Code, nil
}
