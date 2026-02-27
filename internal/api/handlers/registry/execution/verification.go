package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// verifyReplay re-executes a function with the same input and verifies the output matches
func (h *Handler) verifyReplay(fnVersion *storage.RegistryFunctionVersion, originalInput json.RawMessage, originalOutput json.RawMessage, originalDuration int) *ReplayVerificationResult {
	result := &ReplayVerificationResult{
		Status:           VerificationPending,
		OriginalOutput:   originalOutput,
		OriginalDuration: originalDuration,
		VerifiedAt:       time.Now(),
	}

	// Define the execution function (same as in HandleExecute)
	executeFn := func() (json.RawMessage, error) {
		// Check if function has WASM binary (local execution)
		if len(fnVersion.WasmBinary) > 0 {
			// Execute using sandbox - all runtimes use the same path
			return executeLocally(fnVersion, originalInput)
		} else if fnVersion.BackendID != nil {
			// Execute via backend
			backend, err := h.BackendRepo.GetBackendByID(*fnVersion.BackendID)
			if err != nil {
				return nil, fmt.Errorf("backend not found: %w", err)
			}

			// Create execution URL
			execURL := fmt.Sprintf("%s/execute", strings.TrimSuffix(backend.URL, "/"))
			return executeOnBackend(execURL, string(originalInput), fnVersion.TimeoutMs)
		} else if fnVersion.DeploymentID != nil {
			// Execute via deployment (similar to playground)
			deployment, err := h.BackendRepo.GetActiveDeploymentForFunction(context.Background(), fnVersion.FunctionID)
			if err != nil || deployment == nil || deployment.DeployedURL == nil {
				return nil, fmt.Errorf("function is not deployed")
			}

			// Forward request to deployment
			execURL := *deployment.DeployedURL + "/execute"
			client := &http.Client{Timeout: time.Duration(fnVersion.TimeoutMs) * time.Millisecond}

			reqBody := map[string]interface{}{"input": originalInput}
			jsonBody, _ := json.Marshal(reqBody)

			req, err := http.NewRequest("POST", execURL, bytes.NewReader(jsonBody))
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

	// Execute the function for verification
	startTime := time.Now()
	replayedOutput, err := executeFn()
	result.ReplayedDuration = int(time.Since(startTime).Milliseconds())

	if err != nil {
		result.Status = VerificationFailed
		result.Error = err.Error()
		result.OutputMatches = false
		return result
	}

	result.ReplayedOutput = replayedOutput

	// Compare outputs - for deterministic functions, outputs should be identical
	result.OutputMatches = outputsEqual(originalOutput, replayedOutput)

	if result.OutputMatches {
		result.Status = VerificationVerified
	} else {
		result.Status = VerificationFailed
		result.Error = "output mismatch: replay produced different result"
	}

	return result
}
