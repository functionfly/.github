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
	ProviderName = "fly"
	RequestTimeout = 20 * time.Second // Fly.io can have variable latency
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

	// Validate URL format - should be fly.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a fly.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".fly.dev") &&
	   !strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a fly.dev subdomain or custom domain, got: %s", parsedURL.Host)
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
		"cdg", // Paris
		"den", // Denver
		"dfw", // Dallas
		"ewr", // Newark
		"fra", // Frankfurt
		"gru", // São Paulo
		"hkg", // Hong Kong
		"iad", // Ashburn
		"lax", // Los Angeles
		"lhr", // London
		"maa", // Chennai
		"mad", // Madrid
		"mia", // Miami
		"nrt", // Tokyo
		"ord", // Chicago
		"otp", // Bucharest
		"qro", // Querétaro
		"sea", // Seattle
		"sin", // Singapore
		"sjc", // San Jose
		"syd", // Sydney
		"waw", // Warsaw
		"yul", // Montreal
		"yyc", // Calgary
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
	req.Header.Set("Fly-Client-IP", "") // Will be set by Fly proxy
	req.Header.Set("Fly-Forwarded-Port", "")
	req.Header.Set("Fly-Forwarded-Proto", "")
	req.Header.Set("Fly-Region", backend.Region)

	// Add request signing
	err = a.SignRequest(req, backend, startTime)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to sign request: %v", err),
		}, nil
	}

	resp, err := a.client.Do(req)
	latencyMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("health check failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Fly.io should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract Fly.io specific headers
	if region := resp.Header.Get("Fly-Region"); region != "" {
		result.Region = region
	}
	if version := resp.Header.Get("Fly-App-Version"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Fly.io specific headers and request signing
func (a *FlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Fly.io specific headers that may be expected
	req.Header.Set("Fly-Request-Id", "") // Will be set by Fly proxy

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Fly.io requests
func (a *FlyAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}