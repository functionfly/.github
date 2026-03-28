//go:build cgo

package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestAIInferenceRequestValidation tests AI inference request validation
func TestAIInferenceRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *AIInferenceRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &AIInferenceRequest{
				Model:  "gpt-4",
				Input:  json.RawMessage(`"hello world"`),
			},
			wantErr: false,
		},
		{
			name: "empty model",
			req: &AIInferenceRequest{
				Model:  "",
				Input:  json.RawMessage(`"hello world"`),
			},
			wantErr: true,
		},
		{
			name: "empty input",
			req: &AIInferenceRequest{
				Model:  "gpt-4",
				Input:  nil,
			},
			wantErr: true,
		},
		{
			name: "with parameters",
			req: &AIInferenceRequest{
				Model: "gpt-4",
				Input: json.RawMessage(`"hello"`),
				Parameters: map[string]interface{}{
					"temperature": 0.7,
					"max_tokens":  100,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAIInferenceRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAIInferenceRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAIInferenceResponseSerialization tests JSON serialization of AI inference responses
func TestAIInferenceResponseSerialization(t *testing.T) {
	resp := &AIInferenceResponse{
		Output:    json.RawMessage(`"Hello, how can I help you?"`),
		LatencyMs: 150,
		Cost:      0.002,
		ModelUsed: "gpt-4",
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal back
	var decoded AIInferenceResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if decoded.LatencyMs != resp.LatencyMs {
		t.Errorf("LatencyMs = %d, want %d", decoded.LatencyMs, resp.LatencyMs)
	}
	if decoded.Cost != resp.Cost {
		t.Errorf("Cost = %f, want %f", decoded.Cost, resp.Cost)
	}
	if decoded.ModelUsed != resp.ModelUsed {
		t.Errorf("ModelUsed = %s, want %s", decoded.ModelUsed, resp.ModelUsed)
	}
}

// TestAIInferenceErrorResponse tests error response handling
func TestAIInferenceErrorResponse(t *testing.T) {
	resp := &AIInferenceResponse{
		LatencyMs: 50,
		Error: &AIInferenceError{
			Code:    "HTTP_500",
			Message: "Internal server error",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var decoded AIInferenceResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if decoded.Error == nil {
		t.Fatal("Expected error to be present")
	}
	if decoded.Error.Code != "HTTP_500" {
		t.Errorf("Error code = %s, want HTTP_500", decoded.Error.Code)
	}
}

// mockAIGateway creates a mock AI Gateway server
func mockAIGateway(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// TestAIInferenceWithMockGateway tests AI inference with a mock Gateway
func TestAIInferenceWithMockGateway(t *testing.T) {
	// Create mock gateway
	gateway := mockAIGateway(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req AIInferenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Check model
		if req.Model == "" {
			http.Error(w, "Model required", http.StatusBadRequest)
			return
		}

		// Return mock response
		resp := AIInferenceResponse{
			Output:    json.RawMessage(`"Mock AI response"`),
			LatencyMs: 100,
			Cost:      0.001,
			ModelUsed: req.Model,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer gateway.Close()

	// Create config with mock gateway URL
	config := &AIInferenceConfig{
		Enabled:         true,
		GatewayURL:      gateway.URL,
		MaxModelSizeMB:  100,
		DefaultModel:    "gpt-4",
		TimeoutSeconds: 30,
		EnableCaching:  true,
	}

	// Test inference
	ctx := context.Background()
	resp, err := AIInference(ctx, config, "gpt-4", []byte(`"test input"`), "")
	if err != nil {
		t.Fatalf("AIInference() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	// Latency might be 0 if mock is too fast - that's acceptable
	if resp.LatencyMs < 0 {
		t.Errorf("Expected non-negative latency, got %d", resp.LatencyMs)
	}
	if resp.Error != nil {
		t.Errorf("Expected no error in response, got: %v", resp.Error)
	}
}

// TestAIInferenceDisabled tests that disabled AI inference returns an error
func TestAIInferenceDisabled(t *testing.T) {
	config := &AIInferenceConfig{
		Enabled: false,
	}

	ctx := context.Background()
	_, err := AIInference(ctx, config, "gpt-4", []byte(`"test"`), "")
	if err == nil {
		t.Error("Expected error when AI inference is disabled")
	}
}

// TestAIInferenceTimeout tests that timeout is respected
func TestAIInferenceTimeout(t *testing.T) {
	// Create a gateway that delays response (longer than timeout)
	gateway := mockAIGateway(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // 3 second delay
		resp := AIInferenceResponse{
			Output: json.RawMessage(`"slow response"`),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer gateway.Close()

	config := &AIInferenceConfig{
		Enabled:         true,
		GatewayURL:      gateway.URL,
		TimeoutSeconds:  1, // 1 second timeout - should timeout before gateway responds
		MaxModelSizeMB:  100,
	}

	ctx := context.Background()
	_, err := AIInference(ctx, config, "gpt-4", []byte(`"test"`), "")
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// TestAIInferenceGatewayError tests handling of gateway errors
func TestAIInferenceGatewayError(t *testing.T) {
	// Create a gateway that returns an error
	gateway := mockAIGateway(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Gateway error", http.StatusInternalServerError)
	})
	defer gateway.Close()

	config := &AIInferenceConfig{
		Enabled:         true,
		GatewayURL:      gateway.URL,
		TimeoutSeconds:  30,
		MaxModelSizeMB:  100,
	}

	ctx := context.Background()
	_, err := AIInference(ctx, config, "gpt-4", []byte(`"test"`), "")
	if err == nil {
		t.Error("Expected error from gateway")
	}
}

// TestAISignatureResultSerialization tests AISignatureResult serialization
func TestAISignatureResultSerialization(t *testing.T) {
	result := &AISignatureResult{
		Language:      "python",
		Code:          "def hello(): return 'world'",
		Signature:     "def hello() -> str",
		Documentation: "A simple hello function",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal signature result: %v", err)
	}

	var decoded AISignatureResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal signature result: %v", err)
	}

	if decoded.Language != result.Language {
		t.Errorf("Language = %s, want %s", decoded.Language, result.Language)
	}
	if decoded.Code != result.Code {
		t.Errorf("Code = %s, want %s", decoded.Code, result.Code)
	}
}

// TestAIInferenceConfigDefaults tests that AI inference config has proper defaults
func TestAIInferenceConfigDefaults(t *testing.T) {
	config := NewDefaultSecurityConfig()

	if config.AIInference.Enabled != DefaultAIInferenceEnabled {
		t.Errorf("Enabled = %v, want %v", config.AIInference.Enabled, DefaultAIInferenceEnabled)
	}
	if config.AIInference.GatewayURL != DefaultAIGatewayURL {
		t.Errorf("GatewayURL = %s, want %s", config.AIInference.GatewayURL, DefaultAIGatewayURL)
	}
	if config.AIInference.DefaultModel != DefaultAIDefaultModel {
		t.Errorf("DefaultModel = %s, want %s", config.AIInference.DefaultModel, DefaultAIDefaultModel)
	}
	if config.AIInference.TimeoutSeconds != DefaultAITimeoutSeconds {
		t.Errorf("TimeoutSeconds = %d, want %d", config.AIInference.TimeoutSeconds, DefaultAITimeoutSeconds)
	}
	if config.AIInference.EnableCaching != DefaultAIEnableCaching {
		t.Errorf("EnableCaching = %v, want %v", config.AIInference.EnableCaching, DefaultAIEnableCaching)
	}
}

// TestAIInferenceConfigClone tests that AI inference config is properly cloned
func TestAIInferenceConfigClone(t *testing.T) {
	config := NewDefaultSecurityConfig()
	config.AIInference.Enabled = true
	config.AIInference.GatewayURL = "http://custom-gateway:8082"

	cloned := config.Clone()

	if cloned.AIInference.Enabled != config.AIInference.Enabled {
		t.Errorf("Cloned Enabled = %v, want %v", cloned.AIInference.Enabled, config.AIInference.Enabled)
	}
	if cloned.AIInference.GatewayURL != config.AIInference.GatewayURL {
		t.Errorf("Cloned GatewayURL = %s, want %s", cloned.AIInference.GatewayURL, config.AIInference.GatewayURL)
	}
}

// TestLogAIInference tests the audit logging function
func TestLogAIInference(t *testing.T) {
	// Should not panic
	LogAIInference("gpt-4", 100, 200, 150, 0.002, nil)
	LogAIInference("gpt-4", 100, 0, 50, 0, fmt.Errorf("test error"))
}

// TestAIInferenceWithParams tests AI inference with parameters
func TestAIInferenceWithParams(t *testing.T) {
	var receivedReq AIInferenceRequest

	gateway := mockAIGateway(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedReq)

		resp := AIInferenceResponse{
			Output:    json.RawMessage(`"response"`),
			LatencyMs: 100,
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer gateway.Close()

	config := &AIInferenceConfig{
		Enabled:         true,
		GatewayURL:      gateway.URL,
		TimeoutSeconds:  30,
		MaxModelSizeMB:  100,
	}

	params := `{"temperature": 0.7, "max_tokens": 100}`
	ctx := context.Background()

	_, err := AIInference(ctx, config, "gpt-4", []byte(`"test"`), params)
	if err != nil {
		t.Fatalf("AIInference() error = %v", err)
	}

	if receivedReq.Parameters == nil {
		t.Error("Expected parameters to be set")
	}
	if temp, ok := receivedReq.Parameters["temperature"].(float64); !ok || temp != 0.7 {
		t.Errorf("temperature = %v, want 0.7", receivedReq.Parameters["temperature"])
	}
}
