package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// createExecuteFunction creates the main execution function that can be cached.
// When RuntimeRouter is wired, it routes through the router which handles engine
// selection, tier resolution, eager bundling, and fallback. Otherwise it falls
// back to the legacy direct-execution path.
func (h *Handler) createExecuteFunction(fnVersion *storage.RegistryFunctionVersion, execReq functionregistry.ExecutionRequest, r *http.Request, fn *storage.RegistryFunction, maxMemoryMB, maxCPUTimeMs int, resourceUsage **ResourceUsage) func() (json.RawMessage, error) {
	return func() (json.RawMessage, error) {
		// 1. Remote execution paths (backend / deployment) — always bypass router.
		if fnVersion.BackendID != nil {
			backend, err := h.BackendRepo.GetBackendByID(*fnVersion.BackendID)
			if err != nil {
				return nil, fmt.Errorf("backend not found: %w", err)
			}
			execURL := fmt.Sprintf("%s/execute", strings.TrimSuffix(backend.URL, "/"))
			return executeOnBackend(execURL, string(execReq.Input), fnVersion.TimeoutMs)
		}
		if fnVersion.DeploymentID != nil {
			deployment, err := h.BackendRepo.GetActiveDeploymentForFunction(r.Context(), fn.ID)
			if err != nil || deployment == nil || deployment.DeployedURL == nil {
				return nil, fmt.Errorf("function is not deployed")
			}
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
		}

		// 2. Router-driven local execution (preferred when wired).
		if h.RuntimeRouter != nil {
			logrus.Info("Execution path: RuntimeRouter is WIRED, using router-driven execution")
			// Resolve tier from tenant plan.
			tier := resolveTierFromRequest(fn, h.BackendRepo)
			runtime := resolveRuntimeFromVersion(fnVersion)

			req := ExecutionRequest{
				FunctionVersion: fnVersion,
				Function:        fn,
				Input:           execReq.Input,
				MaxMemoryMB:     maxMemoryMB,
				MaxCPUTimeMs:    maxCPUTimeMs,
				TenantID:        "",
				Tier:            tier,
				Runtime:         runtime,
			}
			if fn != nil && fn.TenantID != nil {
				req.TenantID = fn.TenantID.String()
			}

			logrus.WithFields(logrus.Fields{
				"input_len": len(req.Input),
				"input":     string(req.Input),
			}).Debug("Router-driven execution: input before router")

			res, err := h.RuntimeRouter.Execute(r.Context(), req)
			if err != nil {
				if res.ResourceUsage != nil {
					*resourceUsage = res.ResourceUsage
				}
				return nil, err
			}
			if res.ResourceUsage != nil {
				*resourceUsage = res.ResourceUsage
			}
			return res.Output, nil
		}

		// 3. Legacy direct-execution fallback (router not wired).
		logrus.Info("Execution path: RuntimeRouter is NIL, using LEGACY execution")
		logrus.WithFields(logrus.Fields{
			"wasm_binary_len": len(fnVersion.WasmBinary),
			"runtime":         fnVersion.Runtime,
			"source_code_valid": fnVersion.SourceCode.Valid,
		}).Info("Legacy execution path: checking fnVersion state")
		if len(fnVersion.WasmBinary) > 0 {
			output, execErr := executeLocallyWithLimits(fnVersion, execReq.Input, maxMemoryMB, maxCPUTimeMs, fn, h.BackendRepo)
			if execErr != nil {
				if execError, ok := execErr.(*ExecutionError); ok {
					*resourceUsage = execError.ResourceUsage
				}
				return nil, execErr
			}
			return output, nil
		} else if fnVersion.SourceCode.Valid && fnVersion.SourceCode.String != "" {
			return executeWithLazyBundling(fnVersion, execReq.Input, maxMemoryMB, maxCPUTimeMs, resourceUsage, fn, h.BackendRepo)
		}

		return nil, fmt.Errorf("function is not executable")
	}
}

