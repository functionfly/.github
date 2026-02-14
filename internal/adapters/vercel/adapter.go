package vercel

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
	ProviderName = "vercel"
	RequestTimeout = 25 * time.Second // Vercel can be slower than Workers
)

// VercelAdapter implements the ProviderAdapter interface for Vercel
type VercelAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewVercelAdapter creates a new Vercel adapter
func NewVercelAdapter() *VercelAdapter {
	return &VercelAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *VercelAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Vercel specific configuration
func (a *VercelAdapter) ValidateConfig(region, urlStr string) error {
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

	// Validate URL format - should be vercel.app domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a vercel.app subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".vercel.app") &&
	   !strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a vercel.app subdomain or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Vercel regions
// Vercel Edge Functions run in multiple regions, Serverless Functions in specific regions
func (a *VercelAdapter) GetRegions() []string {
	return []string{
		"arn1", // Stockholm (ARN)
		"bom1", // Mumbai (BOM)
		"cdg1", // Paris (CDG)
		"cle1", // Cleveland (CLE)
		"cpt1", // Cape Town (CPT)
		"dub1", // Dublin (DUB)
		"fra1", // Frankfurt (FRA)
		"gru1", // São Paulo (GRU)
		"hkg1", // Hong Kong (HKG)
		"hnd1", // Tokyo (HND)
		"iad1", // Washington DC (IAD)
		"icn1", // Seoul (ICN)
		"jnb1", // Johannesburg (JNB)
		"lax1", // Los Angeles (LAX)
		"lhr1", // London (LHR)
		"pdx1", // Portland (PDX)
		"sfo1", // San Francisco (SFO)
		"sin1", // Singapore (SIN)
		"syd1", // Sydney (SYD)
	}
}

// HealthCheck performs health checks for Vercel deployments
func (a *VercelAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
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

	// Add Vercel-specific headers
	req.Header.Set("User-Agent", "FunctionFly-HealthCheck/1.0")

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

	// Vercel should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract Vercel-specific headers
	if region := resp.Header.Get("x-vercel-region"); region != "" {
		result.Region = region
	}
	if version := resp.Header.Get("x-vercel-deployment-id"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Vercel-specific headers and request signing
func (a *VercelAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Vercel-specific headers that may be expected
	req.Header.Set("x-forwarded-host", req.Host)
	req.Header.Set("x-forwarded-proto", "https")

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Vercel requests
func (a *VercelAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}