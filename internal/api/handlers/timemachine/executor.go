package timemachine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/storage"
	tmengine "github.com/functionfly/functionfly/internal/timemachine"
)

type SandboxExecutorAdapter struct {
	repo   storage.Repository
	client *http.Client
}

func NewSandboxExecutorAdapter(repo storage.Repository) *SandboxExecutorAdapter {
	return &SandboxExecutorAdapter{
		repo:   repo,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

var _ tmengine.Executor = (*SandboxExecutorAdapter)(nil)

func (a *SandboxExecutorAdapter) Execute(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error) {
	if len(fnVersion.WasmBinary) > 0 {
		return a.executeViaSandbox(fnVersion, input, timeoutMs)
	}

	if fnVersion.BackendID != nil {
		return a.executeViaBackend(fnVersion, input, timeoutMs)
	}

	if fnVersion.DeploymentID != nil {
		return a.executeViaDeployment(fnVersion, input, timeoutMs)
	}

	return nil, 0, fmt.Errorf("function version %s has no executable path (no WASM, backend, or deployment)", fnVersion.Version)
}

func (a *SandboxExecutorAdapter) executeViaSandbox(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error) {
	start := time.Now()
	output, err := execution.ExecuteLocally(fnVersion, input, fnVersion.MemoryMB, timeoutMs)
	duration := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, duration, fmt.Errorf("sandbox execution failed: %w", err)
	}
	return output, duration, nil
}

func (a *SandboxExecutorAdapter) executeViaBackend(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error) {
	if fnVersion.BackendID == nil {
		return nil, 0, fmt.Errorf("no backend configured")
	}

	backend, err := a.repo.GetBackendByID(context.Background(), *fnVersion.BackendID)
	if err != nil {
		return nil, 0, fmt.Errorf("backend lookup failed: %w", err)
	}
	if backend == nil {
		return nil, 0, fmt.Errorf("backend not found")
	}

	execURL := strings.TrimSuffix(backend.URL, "/") + "/execute"
	return a.httpExecute(execURL, input, timeoutMs)
}

func (a *SandboxExecutorAdapter) executeViaDeployment(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error) {
	deployment, err := a.repo.GetActiveDeploymentForFunction(context.Background(), fnVersion.FunctionID)
	if err != nil {
		return nil, 0, fmt.Errorf("deployment lookup failed: %w", err)
	}
	if deployment == nil || deployment.DeployedURL == nil {
		return nil, 0, fmt.Errorf("no active deployment for function %s", fnVersion.FunctionID)
	}

	execURL := strings.TrimSuffix(*deployment.DeployedURL, "/") + "/execute"
	return a.httpExecute(execURL, input, timeoutMs)
}

func (a *SandboxExecutorAdapter) httpExecute(url string, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error) {
	reqBody := map[string]interface{}{"input": json.RawMessage(input)}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := a.client.Do(req)
	duration := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, duration, fmt.Errorf("http execute: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, duration, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, duration, fmt.Errorf("execution returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), duration, nil
}
