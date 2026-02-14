package cloudflare

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
	ProviderName = "workers"
	RequestTimeout = 30 * time.Second
)

// CloudflareAdapter implements the ProviderAdapter interface for Cloudflare Workers
type CloudflareAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewCloudflareAdapter creates a new Cloudflare adapter
func NewCloudflareAdapter() *CloudflareAdapter {
	return &CloudflareAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *CloudflareAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Cloudflare Workers specific configuration
func (a *CloudflareAdapter) ValidateConfig(region, urlStr string) error {
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

	// Validate URL format - should be workers.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a workers.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".workers.dev") &&
	   !strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a workers.dev subdomain or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Cloudflare Workers regions
func (a *CloudflareAdapter) GetRegions() []string {
	return []string{
		"us-east-1",    // Northern Virginia
		"us-west-1",    // Northern California
		"eu-west-1",    // Ireland
		"ap-southeast-1", // Singapore
		"ap-northeast-1", // Tokyo
		"ap-southeast-2", // Sydney
		"sa-east-1",    // São Paulo
		"af-south-1",   // Cape Town
	}
}

// HealthCheck performs health checks for Cloudflare Workers
func (a *CloudflareAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
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

	// Cloudflare Workers should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract version info from headers (optional)
	if version := resp.Header.Get("X-Workers-Version"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Cloudflare-specific headers and request signing
func (a *CloudflareAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Cloudflare-specific headers that Workers expect
	req.Header.Set("CF-Ray", "") // Will be set by Cloudflare
	req.Header.Set("CF-Connecting-IP", "") // Will be set by Cloudflare
	req.Header.Set("CF-IPCountry", "") // Will be set by Cloudflare

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Cloudflare Workers requests
func (a *CloudflareAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}