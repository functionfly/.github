package execution

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
)

// createExecuteFunction creates the main execution function that can be cached
func (h *Handler) createExecuteFunction(fnVersion *storage.RegistryFunctionVersion, execReq functionregistry.ExecutionRequest, r *http.Request, fn *storage.RegistryFunction, maxMemoryMB, maxCPUTimeMs int, resourceUsage **ResourceUsage) func() (json.RawMessage, error) {
	return func() (json.RawMessage, error) {
		// Check if function has WASM binary (local execution)
		if len(fnVersion.WasmBinary) > 0 {
			// Execute using sandbox executor - all runtimes use the same path
			output, execErr := executeLocallyWithLimits(fnVersion, execReq.Input, maxMemoryMB, maxCPUTimeMs, fn, h.BackendRepo)
			if execErr != nil {
				if execError, ok := execErr.(*ExecutionError); ok {
					*resourceUsage = execError.ResourceUsage
				}
				return nil, execErr
			}
			return output, nil
		} else if fnVersion.SourceCode.Valid && fnVersion.SourceCode.String != "" {
			// Lazy bundling: bundle source code to WASM at execution time
			// This is triggered when publish didn't bundle (for faster publish)
			return executeWithLazyBundling(fnVersion, execReq.Input, maxMemoryMB, maxCPUTimeMs, resourceUsage, fn, h.BackendRepo)
		} else if fnVersion.BackendID != nil {
			// Execute via specific assigned backend
			backend, err := h.BackendRepo.GetBackendByID(*fnVersion.BackendID)
			if err != nil {
				return nil, fmt.Errorf("backend not found: %w", err)
			}

			// Create execution URL
			execURL := fmt.Sprintf("%s/execute", strings.TrimSuffix(backend.URL, "/"))
			return executeOnBackend(execURL, string(execReq.Input), fnVersion.TimeoutMs)
		} else if fnVersion.DeploymentID != nil {
			// Execute via deployment (similar to playground)
			deployment, err := h.BackendRepo.GetActiveDeploymentForFunction(r.Context(), fn.ID)
			if err != nil || deployment == nil || deployment.DeployedURL == nil {
				return nil, fmt.Errorf("function is not deployed")
			}

			// Forward request to deployment
			execURL := *deployment.DeployedURL + "/execute"
			client := &http.Client{Timeout: time.Duration(fnVersion.TimeoutMs) * time.Millisecond}

			reqBody := map[string]interface{}{"input": execReq.Input}
			jsonBody, _ := json.Marshal(reqBody)

			req, err := http.NewRequestWithContext(r.Context(), "POST", execURL, bytes.NewReader(jsonBody))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("deployment execution failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, fmt.Errorf("deployment returned status %d: %s", resp.StatusCode, string(body))
			}

			return json.RawMessage(body), nil
		} else {
			return nil, fmt.Errorf("function is not executable")
		}
	}
}

// executeWithCaching executes a function with caching support
func (h *Handler) executeWithCaching(eligibility cache.EligibilityResult, input json.RawMessage, executeFn func() (json.RawMessage, error)) (json.RawMessage, error, bool) {
	cacheResult, err := h.CacheService.GetOrExecute(eligibility, input, executeFn)
	if err != nil {
		return nil, err, false
	}
	return cacheResult.Output, nil, cacheResult.FromCache
}

// determineOutcome determines the execution outcome and error code
func determineOutcome(executionErr error, statusCode int) (string, string) {
	outcome := "success"
	errorCode := ""
	if executionErr != nil || statusCode >= 400 {
		outcome = "error"
		if executionErr != nil {
			errorCode = "execution_failed"
		} else {
			errorCode = "http_error"
		}
	}
	return outcome, errorCode
}

// toNullString converts a string pointer to sql.NullString
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
