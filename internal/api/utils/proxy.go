package utils

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ProxyResult captures the outcome of a proxy operation
type ProxyResult struct {
	Success      bool          // Whether the proxy operation succeeded
	StatusCode   int           // HTTP status code returned by backend
	LatencyMs    int64         // Response time in milliseconds
	BackendID    uuid.UUID     // ID of the backend that handled the request
	Outcome      string        // "success", "failure", "timeout", etc.
	ErrorMessage string        // Error message if failed
}

// proxyToBackend proxies the request to the selected backend with failover support
func ProxyToBackend(w http.ResponseWriter, r *http.Request, primaryBackend *storage.Backend, failoverBackends []*storage.Backend, requestID string) *ProxyResult {
	backends := []*storage.Backend{primaryBackend}
	backends = append(backends, failoverBackends...)

	// For non-idempotent methods, only try the primary backend
	isIdempotent := IsIdempotentMethod(r.Method)
	if !isIdempotent && len(backends) > 1 {
		backends = backends[:1]
	}

	var lastErr error
	result := &ProxyResult{
		Success:      false,
		BackendID:    primaryBackend.ID, // Default to primary, will be updated if failover succeeds
		Outcome:      "failure",
	}
	for i, backend := range backends {
		// Create a new request for each backend attempt
		backendURL := backend.URL
		if !strings.HasSuffix(backendURL, "/") {
			backendURL += "/"
		}

		// Build the target URL
		targetURL := backendURL + strings.TrimPrefix(r.URL.Path, "/"+mux.Vars(r)["appSlug"]+"/")
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Create proxy request
		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Add FunctionFly headers
		proxyReq.Header.Set("X-FunctionFly-Request-ID", requestID)
		proxyReq.Header.Set("X-FunctionFly-App-Slug", mux.Vars(r)["appSlug"])

		// Set timeout for backend request
		client := &http.Client{
			Timeout: 30 * time.Second, // Configurable timeout
		}

		start := time.Now()
		resp, err := client.Do(proxyReq)
		latency := time.Since(start)

		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"backend_url": backend.URL,
				"attempt":     i + 1,
				"latency_ms":  latency.Milliseconds(),
				"request_id":  requestID,
			}).Warn("Backend request failed")

			lastErr = err
			// Update result with failure information
			result.LatencyMs = latency.Milliseconds()
			result.BackendID = backend.ID
			result.Outcome = "backend_error"
			result.ErrorMessage = err.Error()
			continue
		}

		// Check if response is successful or if we should retry
		if resp.StatusCode >= 500 && i < len(backends)-1 {
			logrus.WithFields(logrus.Fields{
				"backend_url": backend.URL,
				"status_code": resp.StatusCode,
				"attempt":     i + 1,
				"request_id":  requestID,
			}).Warn("Backend returned 5xx, trying failover")

			resp.Body.Close()
			continue
		}

		// Success - populate result and copy response
		result.Success = true
		result.StatusCode = resp.StatusCode
		result.LatencyMs = latency.Milliseconds()
		result.BackendID = backend.ID
		result.Outcome = "success"

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
		resp.Body.Close()

		logrus.WithFields(logrus.Fields{
			"backend_url": backend.URL,
			"status_code": resp.StatusCode,
			"latency_ms":  latency.Milliseconds(),
			"attempt":     i + 1,
			"request_id":  requestID,
		}).Info("Successfully proxied request")

		return result
	}

	// All backends failed
	result.Outcome = "all_backends_failed"
	if lastErr != nil {
		result.ErrorMessage = lastErr.Error()
	}

	logrus.WithError(lastErr).WithFields(logrus.Fields{
		"app_slug":   mux.Vars(r)["appSlug"],
		"backends":   len(backends),
		"request_id": requestID,
	}).Error("All backends failed")

	http.Error(w, "All backends unavailable", http.StatusServiceUnavailable)
	return result
}


// isIdempotentMethod checks if an HTTP method is considered idempotent for fast failover
func IsIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}