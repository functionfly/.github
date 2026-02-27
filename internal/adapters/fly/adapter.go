package fly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/storage"
)

const (
	ProviderName   = "fly"
	RequestTimeout = 30 * time.Second
)

// FlyAdapter implements the ProviderAdapter interface for Fly.io
type FlyAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewFlyAdapter creates a new Fly.io adapter
func NewFlyAdapter() *FlyAdapter {
	return &FlyAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *FlyAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Fly.io specific configuration
func (a *FlyAdapter) ValidateConfig(region, urlStr string) error {
	// Validate region
	validRegions := a.GetRegions()
	regionValid := false
	for _, r := range validRegions {
		if r == region {
			regionValid = true
			break
		}
	}
	if !regionValid {
		return fmt.Errorf("invalid region '%s', valid regions: %v", region, validRegions)
	}

	// Validate URL format - should be a fly.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a fly.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".fly.dev") &&
		!strings.Contains(parsedURL.Host, ".") &&
		!strings.HasSuffix(parsedURL.Host, ".internal") {
		return fmt.Errorf("URL must be a fly.dev subdomain, .internal domain, or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Fly.io regions
func (a *FlyAdapter) GetRegions() []string {
	return []string{
		"ams", // Amsterdam
		"arn", // Stockholm
		"atl", // Atlanta
		"bog", // Bogotá
		"bos", // Boston
		"bru", // Brussels
		"cdg", // Paris
		"den", // Denver
		"dfw", // Dallas
		"ewr", // New Jersey
		"eze", // Buenos Aires
		"fra", // Frankfurt
		"gig", // Rio de Janeiro
		"gru", // São Paulo
		"hkg", // Hong Kong
		"iad", // Ashburn
		"jnb", // Johannesburg
		"lax", // Los Angeles
		"lhr", // London
		"mad", // Madrid
		"mia", // Miami
		"nrt", // Tokyo
		"ord", // Chicago
		"phx", // Phoenix
		"pma", // Palm Beach
		"sea", // Seattle
		"sfo", // San Francisco
		"sin", // Singapore
		"syd", // Sydney
		"tsn", // Toronto
		"waw", // Warsaw
		"yyz", // Toronto
	}
}

// HealthCheck performs health checks for Fly.io deployments
func (a *FlyAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
	startTime := time.Now()

	// Check /healthz endpoint
	healthURL := strings.TrimSuffix(backend.URL, "/") + "/healthz"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Add Fly.io specific headers
	req.Header.Set("User-Agent", "FunctionFly-HealthCheck/1.0")
	req.Header.Set("fly-client-ip", "0.0.0.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("health check request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latencyMs := int(time.Since(startTime).Milliseconds())

	// Fly.io should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("health check returned status %d", resp.StatusCode)
	}

	return result, nil
}

// SignRequest adds Fly.io specific headers/signatures to requests
func (a *FlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Fly.io specific headers
	req.Header.Set("fly-client-ip", "0.0.0.0")
	req.Header.Set("fly-forwarded-proto", "https")

	// Use HMAC signing if configured
	if backend.SharedSecret != "" {
		return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
	}

	return nil
}

// GetRequestTimeout returns the recommended timeout for requests to Fly.io
func (a *FlyAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}
