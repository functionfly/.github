package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/sirupsen/logrus"
)

// AIProxyHandler provides proxy endpoints to the FlyMind AI Service
// This allows the dashboard to make AI requests through the orchestrator API
// without exposing the AI service directly to the frontend
type AIProxyHandler struct {
	aiServiceURL string
	httpClient   *http.Client
}

// NewAIProxyHandler creates a new AI proxy handler
func NewAIProxyHandler() *AIProxyHandler {
	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = "http://localhost:18081" // Default AI service URL
	}

	return &AIProxyHandler{
		aiServiceURL: aiURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for AI generation
		},
	}
}

// proxyRequest forwards a request to the AI service
func (h *AIProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, aiPath string) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Build the target URL
	targetURL := h.aiServiceURL + aiPath

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create a new request to the AI service
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		logrus.WithError(err).Error("Failed to create proxy request")
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set("Content-Type", "application/json")

	// Add user context to the request for the AI service to use
	// This allows the AI service to track usage per user/tenant
	proxyReq.Header.Set("X-User-ID", user.UserID.String())
	proxyReq.Header.Set("X-Tenant-ID", user.TenantID.String())

	// Send the request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to proxy request to AI service")
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
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
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
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
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
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
