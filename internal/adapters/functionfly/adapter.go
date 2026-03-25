package functionfly

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
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
	ProviderName   = "functionfly-edge"
	WASMProviderName = "functionfly-wasm"
	RequestTimeout = 30 * time.Second
	HealthPath     = "/healthz"
	// AppNameMaxLen and allowed pattern to prevent path traversal and injection
	AppNameMaxLen     = 128
	appNamePatternStr = `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
)

var appNameRegex = regexp.MustCompile(appNamePatternStr)

// FunctionFlyAdapter implements ProviderAdapter and DeploymentAdapter for FunctionFly Edge.
// Edge is a zero-deployment provider: functions are served from the edge URL; deploy only validates and returns the URL.
// For WASM runtime, it pushes the WASM artifact to the WASM edge service.
type FunctionFlyAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
	// wasmClient is used for WASM deployments (nil means use default client)
	wasmClient *http.Client
}

// NewFunctionFlyAdapter creates a new FunctionFly Edge adapter with default HTTP client.
func NewFunctionFlyAdapter() *FunctionFlyAdapter {
	return NewFunctionFlyAdapterWithClient(&http.Client{Timeout: RequestTimeout})
}

// NewFunctionFlyAdapterWithClient creates a new FunctionFly Edge adapter with a custom HTTP client (e.g. for tests or custom TLS).
func NewFunctionFlyAdapterWithClient(client *http.Client) *FunctionFlyAdapter {
	if client == nil {
		client = &http.Client{Timeout: RequestTimeout}
	}
	return &FunctionFlyAdapter{
		signer: &signing.RequestSigner{},
		client: client,
	}
}

func (a *FunctionFlyAdapter) GetName() string { return ProviderName }

// ValidateConfig validates region and edge URL. URL must be HTTPS and a valid base URL.
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
	return []string{"us-east-1", "eu-west-1", "ap-southeast-1"}
}

// HealthCheck performs an HTTP GET to backend.URL/healthz with signed request, respects context and reports latency.
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
	resp, err := a.client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("health check failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()
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

// SignRequest signs the request with the backend's shared secret (HMAC-SHA256). Safe if backend is nil.
func (a *FunctionFlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	secret := ""
	if backend != nil {
		secret = backend.SharedSecret
	}
	return a.signer.SignRequest(req, secret, timestamp)
}

func (a *FunctionFlyAdapter) GetRequestTimeout() time.Duration { return RequestTimeout }

// validateAppName returns an error if name is invalid (injection or path traversal risk).
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

// baseURLFromSpec returns the edge base URL from provider config or default.
func baseURLFromSpec(spec *common.DeploymentSpec) string {
	if spec != nil && spec.ProviderConfig != nil {
		if u, ok := spec.ProviderConfig["url"].(string); ok && u != "" {
			return strings.TrimSuffix(u, "/")
		}
	}
	return "https://edge.functionfly.com"
}

// wasmURLFromSpec returns the WASM edge service URL from provider config or default.
func wasmURLFromSpec(spec *common.DeploymentSpec) string {
	if spec != nil && spec.ProviderConfig != nil {
		if u, ok := spec.ProviderConfig["wasm_url"].(string); ok && u != "" {
			return strings.TrimSuffix(u, "/")
		}
	}
	return "http://localhost:8080" // Default to local for development
}

// deployWASM pushes the WASM artifact to the WASM edge service.
func (a *FunctionFlyAdapter) deployWASM(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	wasmURL := wasmURLFromSpec(spec)

	// Prepare the WASM artifact (base64 encoded)
	artifact := spec.Artifact
	if len(artifact) == 0 {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "WASM artifact is required for WASM runtime",
		}, nil
	}

	// Create deployment request to WASM edge service
	deployURL := fmt.Sprintf("%s/deploy/%s", wasmURL, spec.AppName)

	// Wrap the WASM bytes in JSON with base64 encoding
	encoded := base64.StdEncoding.EncodeToString(artifact)
	payload := fmt.Sprintf(`{"wasm":"%s"}`, encoded)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deployURL, strings.NewReader(payload))
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to create WASM deployment request: %v", err),
		}, nil
	}
	req.Header.Set("Content-Type", "application/json")

	// Use WASM client or default client
	client := a.wasmClient
	if client == nil {
		client = a.client
	}

	resp, err := client.Do(req)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to deploy WASM function: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("WASM deployment failed: %s", string(body)),
		}, nil
	}

	// Return successful deployment
	deploymentURL := fmt.Sprintf("%s/%s", wasmURL, spec.AppName)
	return &common.DeploymentResult{
		Status:        common.DeploymentStatusSuccess,
		Message:       "WASM function deployed successfully",
		DeploymentURL: deploymentURL,
		DeploymentID:  spec.AppName,
		Metadata: map[string]interface{}{
			"provider": WASMProviderName,
			"runtime":  string(common.RuntimeWASM),
			"endpoint": deploymentURL,
			"deployed": true,
		},
	}, nil
}

// Deploy validates the spec and returns the deployment URL.
// For WASM runtime, the artifact is pushed to the WASM edge service.
// For other runtimes, FunctionFly Edge is zero-deploy; no artifact is pushed.
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

	// Handle WASM runtime - deploy the artifact to WASM edge service
	if spec.Runtime == common.RuntimeWASM || spec.Runtime == common.RuntimeRust {
		return a.deployWASM(ctx, spec)
	}

	// Default: zero-deploy edge proxy
	base := baseURLFromSpec(spec)
	// Build URL safely: base is already validated or default; AppName is validated
	deploymentURL := base + "/" + spec.AppName
	return &common.DeploymentResult{
		Status:        common.DeploymentStatusSuccess,
		Message:       "Function available via FunctionFly Edge",
		DeploymentURL: deploymentURL,
		DeploymentID:  spec.AppName,
		Metadata: map[string]interface{}{
			"provider": ProviderName,
			"endpoint": deploymentURL,
			"noDeploy": true,
		},
	}, nil
}

// SetEnv is a no-op for FunctionFly Edge; env is managed by the orchestrator/vault at invoke time.
func (a *FunctionFlyAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	return nil
}

// BindRoutes is a no-op for FunctionFly Edge; routing is by path prefix on the edge.
func (a *FunctionFlyAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	return nil
}

// GetDeploymentStatus returns success for any non-empty deploymentID (edge serves by app name).
func (a *FunctionFlyAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	if deploymentID == "" {
		return common.DeploymentStatusFailed, fmt.Errorf("deployment ID is required")
	}
	return common.DeploymentStatusSuccess, nil
}

// Rollback is a no-op; FunctionFly Edge applies updates instantly.
func (a *FunctionFlyAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	return &common.DeploymentResult{
		Status:  common.DeploymentStatusSuccess,
		Message: "No rollback needed — FunctionFly Edge provides instant updates",
	}, nil
}
