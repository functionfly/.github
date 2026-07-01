package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/aikeys"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// AIProxyHandler provides proxy endpoints to the FlyMind AI Service
// This allows the dashboard to make AI requests through the orchestrator API
// without exposing the AI service directly to the frontend
type AIProxyHandler struct {
	aiServiceURL string
	httpClient   *http.Client
	byokRepo     *aikeys.Repository
}

// NewAIProxyHandler creates a new AI proxy handler
func NewAIProxyHandler(byokRepo *aikeys.Repository) *AIProxyHandler {
	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		logrus.Warn("AI_SERVICE_URL not set - AI proxy will not be available")
		aiURL = "" // Must be set explicitly
	}

	return &AIProxyHandler{
		aiServiceURL: aiURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for AI generation
		},
		byokRepo: byokRepo,
	}
}

// proxyRequest forwards a request to the AI service
func (h *AIProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, aiPath string) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Build the target URL
	targetURL := h.aiServiceURL + aiPath

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read request body")
		apierror.WriteError(w, apierror.NewBadRequest("Failed to read request body"))
		return
	}
	defer r.Body.Close()

	// Create a new request to the AI service
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		logrus.WithError(err).Error("Failed to create proxy request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create proxy request"))
		return
	}

	// Copy headers
	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set("Content-Type", "application/json")

	// Add user context to the request for the AI service to use
	proxyReq.Header.Set("X-User-ID", user.UserID.String())
	proxyReq.Header.Set("X-Tenant-ID", user.TenantID.String())

	// Inject BYOK key if user has one for the requested provider
	h.injectBYOKHeader(proxyReq, user.TenantID.String(), body)

	// Send the request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to proxy request to AI service")
		apierror.WriteError(w, apierror.NewServiceUnavailable("AI service unavailable"))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		logrus.WithError(err).Warn("Failed to copy response body")
	}
}

// HandleGenerateFunction proxies function generation requests to the AI service
func (h *AIProxyHandler) HandleGenerateFunction(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/api/composer/generate")
}

// HandleGenerateFunctionStream proxies streaming function generation requests
func (h *AIProxyHandler) HandleGenerateFunctionStream(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Build the target URL with query params (excluding the _token query param)
	// The AI service doesn't need the token - we handle auth at the proxy level
	queryParams := r.URL.Query()
	queryParams.Del("_token") // Remove the auth token from query params
	targetURL := h.aiServiceURL + "/api/composer/generate/stream?" + queryParams.Encode()

	// Create a new request to the AI service
	proxyReq, err := http.NewRequest(r.Method, targetURL, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create proxy request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create proxy request"))
		return
	}

	// Copy headers
	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set("X-User-ID", user.UserID.String())
	proxyReq.Header.Set("X-Tenant-ID", user.TenantID.String())

	// Send the request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to proxy request to AI service")
		apierror.WriteError(w, apierror.NewServiceUnavailable("AI service unavailable"))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// For streaming responses, we need to flush the response writer
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Copy response body (streaming)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			break
		}
	}
}

// HealthResponse represents the AI service health response
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	AIHealthy bool   `json:"ai_healthy"`
	AIMessage string `json:"ai_message,omitempty"`
}

// HandleHealth returns the health status of the AI service proxy
func (h *AIProxyHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Check AI service health
	resp, err := h.httpClient.Get(h.aiServiceURL + "/health")
	aiHealthy := err == nil && resp != nil && resp.StatusCode == http.StatusOK
	aiMessage := ""
	if err != nil {
		aiMessage = err.Error()
	} else if resp != nil {
		aiMessage = resp.Status
		resp.Body.Close()
	}

	health := HealthResponse{
		Status:    "healthy",
		Service:   "ai-proxy",
		Version:   "1.0.0",
		AIHealthy: aiHealthy,
		AIMessage: aiMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// HandleRefineFunction proxies function refinement requests to the AI service
func (h *AIProxyHandler) HandleRefineFunction(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/api/ai/composer/refine")
}

// HandleRefineFunctionStream proxies streaming function refinement requests
func (h *AIProxyHandler) HandleRefineFunctionStream(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Build the target URL with query params (excluding the _token query param)
	queryParams := r.URL.Query()
	queryParams.Del("_token")
	targetURL := h.aiServiceURL + "/api/ai/composer/refine/stream?" + queryParams.Encode()

	// Create a new request to the AI service
	proxyReq, err := http.NewRequest(r.Method, targetURL, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create proxy request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create proxy request"))
		return
	}

	// Copy headers
	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set("X-User-ID", user.UserID.String())
	proxyReq.Header.Set("X-Tenant-ID", user.TenantID.String())

	// Send the request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to proxy request to AI service")
		apierror.WriteError(w, apierror.NewServiceUnavailable("AI service unavailable"))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			break
		}
	}
}

// HandleAIStatus returns the status of all AI namespace features
func (h *AIProxyHandler) HandleAIStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"namespace": "ai",
		"version":   "1.0.0",
		"features": map[string]interface{}{
			"composer": map[string]interface{}{
				"path":        "/ai/composer/*",
				"status":      "active",
				"description": "AI function generation",
				"endpoints": []string{
					"POST /v1/ai/composer/generate",
					"GET /v1/ai/composer/generate/stream",
					"POST /v1/ai/composer/refine",
				},
			},
		},
		"message": "AI namespace is active. Use /v1/ai/composer/* for current features.",
	}

	json.NewEncoder(w).Encode(status)
}

// injectBYOKHeader checks if the tenant has a BYOK key for the provider
// being requested and injects the key into the proxy request headers.
func (h *AIProxyHandler) injectBYOKHeader(proxyReq *http.Request, tenantID string, body []byte) {
	if h.byokRepo == nil {
		proxyReq.Header.Set("X-Key-Source", "platform")
		return
	}

	// Extract provider from request body
	provider := extractProviderFromBody(body)
	if provider == "" {
		proxyReq.Header.Set("X-Key-Source", "platform")
		return
	}

	// Parse tenant UUID
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		proxyReq.Header.Set("X-Key-Source", "platform")
		return
	}

	// Look up BYOK key — try exact provider first, then token plan fallbacks
	key, err := h.byokRepo.GetByTenantAndProvider(proxyReq.Context(), tid, provider)
	if err != nil || key == nil || key.Status != "active" {
		// For mimo requests, also check for mimo-token-plan key
		if provider == "mimo" {
			key, err = h.byokRepo.GetByTenantAndProvider(proxyReq.Context(), tid, "mimo-token-plan")
			if err != nil || key == nil || key.Status != "active" {
				proxyReq.Header.Set("X-Key-Source", "platform")
				return
			}
		} else if provider == "minimax" {
			// For minimax requests, also check for minimax-token-plan key
			key, err = h.byokRepo.GetByTenantAndProvider(proxyReq.Context(), tid, "minimax-token-plan")
			if err != nil || key == nil || key.Status != "active" {
				proxyReq.Header.Set("X-Key-Source", "platform")
				return
			}
		} else {
			proxyReq.Header.Set("X-Key-Source", "platform")
			return
		}
	}

	// Decrypt the key
	plaintext, err := aikeys.DecryptKey(key.EncryptedKey, key.KeyNonce, tid)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("Failed to decrypt BYOK key, falling back to platform")
		proxyReq.Header.Set("X-Key-Source", "platform")
		return
	}

	// Inject BYOK headers
	proxyReq.Header.Set("X-BYOK-Key", string(plaintext))
	proxyReq.Header.Set("X-BYOK-Provider", provider)
	proxyReq.Header.Set("X-Key-Source", "byok")

	// For token plan keys, inject the regional base URL and override provider
	if key.Provider == "mimo-token-plan" {
		region := extractRegionFromHealthMessage(key.HealthMessage)
		if base, ok := aikeys.TokenPlanRegionURLs[region]; ok {
			proxyReq.Header.Set("X-BYOK-Base-URL", base)
		}
		proxyReq.Header.Set("X-BYOK-Provider", "mimo")
	} else if key.Provider == "minimax-token-plan" {
		proxyReq.Header.Set("X-BYOK-Provider", "minimax")
	}

	// Update last_used_at in background
	go func() {
		_ = h.byokRepo.UpdateLastUsed(context.Background(), key.ID)
	}()
}

// extractRegionFromHealthMessage extracts the region code from a health_message string.
func extractRegionFromHealthMessage(healthMessage string) string {
	const prefix = "region:"
	if len(healthMessage) > len(prefix) && healthMessage[:len(prefix)] == prefix {
		return healthMessage[len(prefix):]
	}
	return ""
}

// extractProviderFromBody tries to extract the "provider" field from a JSON request body.
func extractProviderFromBody(body []byte) string {
	var req struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Provider
}
