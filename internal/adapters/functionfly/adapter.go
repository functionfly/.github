package functionfly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/storage"
)

const (
	ProviderName     = "functionfly-edge"
	WASMProviderName = "functionfly-wasm"
	RequestTimeout   = 30 * time.Second
	HealthPath       = "/healthz"
	AppNameMaxLen    = 128
	appNamePatternStr = `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
)

var appNameRegex = regexp.MustCompile(appNamePatternStr)

type FunctionFlyAdapter struct {
	signer           *signing.RequestSigner
	client           *http.Client
	deploymentClient EdgeDeploymentClientInterface
	edgeURL          string
	apiKey           string
}

func NewFunctionFlyAdapter() *FunctionFlyAdapter {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return NewFunctionFlyAdapterWithClient(&http.Client{
		Timeout:   RequestTimeout,
		Transport: transport,
	})
}

func NewFunctionFlyAdapterWithClient(client *http.Client) *FunctionFlyAdapter {
	if client == nil {
		client = &http.Client{Timeout: RequestTimeout}
	}
	return &FunctionFlyAdapter{
		signer: &signing.RequestSigner{},
		client: client,
	}
}

func NewFunctionFlyAdapterWithDeploymentClient(edgeURL, apiKey string) *FunctionFlyAdapter {
	adapter := &FunctionFlyAdapter{
		signer:           &signing.RequestSigner{},
		client:           &http.Client{Timeout: RequestTimeout},
		deploymentClient: NewDeploymentClient(edgeURL, apiKey),
		edgeURL:          edgeURL,
		apiKey:          apiKey,
	}
	return adapter
}

func (a *FunctionFlyAdapter) WithDeploymentClient(client EdgeDeploymentClientInterface) *FunctionFlyAdapter {
	a.deploymentClient = client
	return a
}

func (a *FunctionFlyAdapter) GetName() string { return ProviderName }

func (a *FunctionFlyAdapter) ValidateConfig(region, rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required for %s provider", ProviderName)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS for %s provider", ProviderName)
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host for %s provider", ProviderName)
	}
	validRegions := a.GetRegions()
	regionOK := false
	for _, r := range validRegions {
		if r == region {
			regionOK = true
			break
		}
	}
	if !regionOK {
		return fmt.Errorf("invalid region %q, valid: %v", region, validRegions)
	}
	return nil
}

func (a *FunctionFlyAdapter) GetRegions() []string {
	return []string{"eu-central-1", "us-east-1"}
}

func (a *FunctionFlyAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
	if backend == nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: "backend is nil",
		}, nil
	}
	base := strings.TrimSuffix(backend.URL, "/")
	if base == "" {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: "backend URL is empty",
		}, nil
	}
	healthURL := base + HealthPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}
	start := time.Now()
	if err := a.SignRequest(req, backend, start); err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to sign request: %v", err),
		}, nil
	}

	var resp *http.Response
	var doErr error
	var latencyMs int
	for attempt := 0; attempt < 2; attempt++ {
		attemptStart := time.Now()
		resp, doErr = a.client.Do(req)
		latencyMs = int(time.Since(attemptStart).Milliseconds())
		if doErr == nil {
			break
		}
		errStr := doErr.Error()
		if attempt == 0 && (strings.Contains(errStr, "EOF") || strings.Contains(errStr, "connection reset") || strings.Contains(errStr, "broken pipe")) {
			time.Sleep(100 * time.Millisecond)
			newReq, newReqErr := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			if newReqErr != nil {
				doErr = newReqErr
				break
			}
			req = newReq
			if signErr := a.SignRequest(req, backend, time.Now()); signErr != nil {
				return &common.HealthCheckResult{
					OK:           false,
					ErrorMessage: fmt.Sprintf("failed to sign retry request: %v", signErr),
				}, nil
			}
			continue
		}
		break
	}

	if doErr != nil {
		return &common.HealthCheckResult{
			OK:           false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("health check failed: %v", doErr),
		}, nil
	}
		defer func() { _ = resp.Body.Close() }()

	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}
	if v := resp.Header.Get("X-FFLY-Version"); v != "" {
		result.Version = v
	}
	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}
	return result, nil
}

func (a *FunctionFlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	secret := ""
	if backend != nil {
		secret = backend.SharedSecret
	}
	return a.signer.SignRequest(req, secret, timestamp)
}

func (a *FunctionFlyAdapter) GetRequestTimeout() time.Duration { return RequestTimeout }

func validateAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name is required")
	}
	if len(name) > AppNameMaxLen {
		return fmt.Errorf("app name must be at most %d characters", AppNameMaxLen)
	}
	if !appNameRegex.MatchString(name) {
		return fmt.Errorf("app name must match %s", appNamePatternStr)
	}
	return nil
}

func baseURLFromSpec(spec *common.DeploymentSpec) string {
	if spec != nil && spec.ProviderConfig != nil {
		if u, ok := spec.ProviderConfig["url"].(string); ok && u != "" {
			return strings.TrimSuffix(u, "/")
		}
	}
	return "https://edge.functionfly.com"
}

func wasmURLFromSpec(spec *common.DeploymentSpec) string {
	if spec != nil && spec.ProviderConfig != nil {
		if u, ok := spec.ProviderConfig["wasm_url"].(string); ok && u != "" {
			return strings.TrimSuffix(u, "/")
		}
	}
	return "http://localhost:8080"
}

func apiKeyFromSpec(spec *common.DeploymentSpec) string {
	if spec != nil && spec.ProviderConfig != nil {
		if key, ok := spec.ProviderConfig["api_key"].(string); ok && key != "" {
			return key
		}
	}
	return ""
}

func (a *FunctionFlyAdapter) getDeploymentClient(spec *common.DeploymentSpec) EdgeDeploymentClientInterface {
	if a.deploymentClient != nil {
		return a.deploymentClient
	}
	edgeURL := baseURLFromSpec(spec)
	apiKey := apiKeyFromSpec(spec)
	return NewDeploymentClient(edgeURL, apiKey)
}

func (a *FunctionFlyAdapter) getClientFromProviderConfig(providerConfig map[string]interface{}) EdgeDeploymentClientInterface {
	if a.deploymentClient != nil {
		return a.deploymentClient
	}
	edgeURL := "https://edge.functionfly.com"
	if providerConfig != nil {
		if u, ok := providerConfig["url"].(string); ok && u != "" {
			edgeURL = strings.TrimSuffix(u, "/")
		}
	}
	apiKey := ""
	if providerConfig != nil {
		if key, ok := providerConfig["api_key"].(string); ok && key != "" {
			apiKey = key
		}
	}
	return NewDeploymentClient(edgeURL, apiKey)
}

func (a *FunctionFlyAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	if spec == nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "deployment spec is required",
		}, nil
	}
	if err := validateAppName(spec.AppName); err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: err.Error(),
		}, nil
	}

	if spec.Runtime == common.RuntimeWASM || spec.Runtime == common.RuntimeRust {
		return a.deployWASM(ctx, spec)
	}

	client := a.getDeploymentClient(spec)

	result, err := client.RegisterFunction(ctx, spec)
	if err != nil {
		base := baseURLFromSpec(spec)
		deploymentURL := base + "/" + spec.AppName
		return &common.DeploymentResult{
			Status:        common.DeploymentStatusSuccess,
			Message:       fmt.Sprintf("Function registered (API unavailable: %v), accessible via edge", err),
			DeploymentURL: deploymentURL,
			DeploymentID:  spec.AppName,
			Metadata: map[string]interface{}{
				"provider":   ProviderName,
				"endpoint":   deploymentURL,
				"registered": false,
				"error":     err.Error(),
			},
		}, nil
	}

	deploymentURL := result.DeploymentURL
	if deploymentURL == "" {
		deploymentURL = baseURLFromSpec(spec) + "/" + spec.AppName
	}

	return &common.DeploymentResult{
		DeploymentID:  result.DeploymentID,
		Status:        result.Status,
		Message:       result.Message,
		DeploymentURL: deploymentURL,
		Metadata: map[string]interface{}{
			"provider":   ProviderName,
			"endpoint":   deploymentURL,
			"registered": true,
		},
	}, nil
}

func (a *FunctionFlyAdapter) deployWASM(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	wasmURL := wasmURLFromSpec(spec)

	artifact := spec.Artifact
	if len(artifact) == 0 {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "WASM artifact is required for WASM runtime",
		}, nil
	}

	var wasmClient EdgeDeploymentClientInterface
	if a.deploymentClient != nil {
		wasmClient = a.deploymentClient
	} else {
		wasmClient = &DeploymentClient{
			httpClient: &http.Client{Timeout: 60 * time.Second},
			edgeURL:    wasmURL,
			apiKey:     apiKeyFromSpec(spec),
		}
	}

	result, err := wasmClient.RegisterFunction(ctx, spec)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("WASM deployment failed: %v", err),
		}, nil
	}

	deploymentURL := wasmURL + "/" + spec.AppName
	if result.DeploymentURL != "" {
		deploymentURL = result.DeploymentURL
	}

	return &common.DeploymentResult{
		Status:        result.Status,
		Message:       result.Message,
		DeploymentURL: deploymentURL,
		DeploymentID:  spec.AppName,
		Metadata: map[string]interface{}{
			"provider": WASMProviderName,
			"runtime":  string(common.RuntimeWASM),
			"endpoint": deploymentURL,
		},
	}, nil
}

func (a *FunctionFlyAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	if err := validateAppName(deploymentID); err != nil {
		return err
	}

	client := a.getClientFromProviderConfig(providerConfig)

	if err := client.SetEnvironment(ctx, deploymentID, envVars, secrets); err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}

	return nil
}

func (a *FunctionFlyAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	if err := validateAppName(deploymentID); err != nil {
		return err
	}

	if len(routes) == 0 {
		return nil
	}

	client := a.getClientFromProviderConfig(providerConfig)

	if err := client.BindRoutes(ctx, deploymentID, routes); err != nil {
		return fmt.Errorf("failed to bind routes: %w", err)
	}

	return nil
}

func (a *FunctionFlyAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	if err := validateAppName(deploymentID); err != nil {
		return common.DeploymentStatusFailed, err
	}

	client := a.getClientFromProviderConfig(providerConfig)

	status, err := client.GetFunctionStatus(ctx, deploymentID)
	if err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("status check failed: %w", err)
	}

	if !status.Exists {
		return common.DeploymentStatusFailed, fmt.Errorf("function not found: %s", deploymentID)
	}

	return status.Status, nil
}

func (a *FunctionFlyAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	if spec == nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "deployment spec is required for rollback",
		}, nil
	}

	if err := validateAppName(spec.AppName); err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: err.Error(),
		}, nil
	}

	client := a.getDeploymentClient(spec)

	currentStatus, err := client.GetFunctionStatus(ctx, spec.AppName)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to get current status: %v", err),
		}, nil
	}

	if !currentStatus.Exists {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("function %s not found, cannot rollback", spec.AppName),
		}, nil
	}

	if spec.Version != "" && currentStatus.Version == spec.Version {
		return &common.DeploymentResult{
			Status:        common.DeploymentStatusSuccess,
			Message:       fmt.Sprintf("function %s is already at version %s", spec.AppName, spec.Version),
			DeploymentURL: baseURLFromSpec(spec) + "/" + spec.AppName,
			Metadata: map[string]interface{}{
				"no_change": true,
				"version":   spec.Version,
			},
		}, nil
	}

	result, err := client.RegisterFunction(ctx, spec)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("rollback failed: %v", err),
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID:  result.DeploymentID,
		Status:        result.Status,
		Message:      fmt.Sprintf("Rollback to %s: %s", spec.Version, result.Message),
		DeploymentURL: result.DeploymentURL,
	}, nil
}

func (a *FunctionFlyAdapter) Delete(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) error {
	if err := validateAppName(deploymentID); err != nil {
		return err
	}

	edgeURL := "https://edge.functionfly.com"
	if providerConfig != nil {
		if u, ok := providerConfig["url"].(string); ok && u != "" {
			edgeURL = strings.TrimSuffix(u, "/")
		}
	}

	apiKey := ""
	if providerConfig != nil {
		if key, ok := providerConfig["api_key"].(string); ok && key != "" {
			apiKey = key
		}
	}

	client := NewDeploymentClient(edgeURL, apiKey)

	if err := client.DeleteFunction(ctx, deploymentID); err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	return nil
}
