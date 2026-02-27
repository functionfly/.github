package security

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// API security scanning functions

func (sas *SecurityAuditService) scanAPISecurity(ctx context.Context, target string, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for API rate limiting
	if noRateLimit, err := sas.checkRateLimiting(target); err == nil && noRateLimit {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Missing API Rate Limiting",
			Description: "API endpoints are not protected by rate limiting",
			Severity:    "high",
			Category:    "network",
			Component:   "api_gateway",
			Location:    target,
			Status:      "open",
			Remediation: "Implement rate limiting on all API endpoints",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	// Check for API authentication
	if weakAuth, err := sas.checkAPIAuthentication(target); err == nil && weakAuth {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Weak API Authentication",
			Description: "API authentication mechanism is insufficient",
			Severity:    "critical",
			Category:    "auth",
			Component:   "api_gateway",
			Location:    target,
			Status:      "open",
			Remediation: "Implement strong authentication (JWT, OAuth2) for API access",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkRateLimiting(target string) (bool, error) {
	// Test rate limiting by making multiple requests
	// If no rate limiting is detected, return true (vulnerability exists)

	const numRequests = 10
	const delayBetweenRequests = 100 * time.Millisecond

	rateLimitedCount := 0
	successCount := 0

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < numRequests; i++ {
		resp, err := client.Get("http://" + target)
		if err != nil {
			// Connection error, might indicate server overload or blocking
			rateLimitedCount++
			continue
		}

		statusCode := resp.StatusCode
		resp.Body.Close()

		// Check for rate limiting indicators
		if statusCode == 429 { // Too Many Requests
			rateLimitedCount++
		} else if statusCode == 503 || statusCode == 502 { // Service Unavailable/Bad Gateway (possible rate limiting)
			rateLimitedCount++
		} else if statusCode >= 200 && statusCode < 400 {
			successCount++
		}

		// Check for rate limit headers
		if resp.Header.Get("X-RateLimit-Remaining") == "0" ||
			resp.Header.Get("Retry-After") != "" ||
			resp.Header.Get("X-RateLimit-Reset") != "" {
			rateLimitedCount++
		}

		// Small delay between requests to avoid being too aggressive
		time.Sleep(delayBetweenRequests)
	}

	// If we got very few rate limiting responses, rate limiting is likely missing
	// Allow for some tolerance (e.g., 2-3 rate limited responses out of 10 might be acceptable)
	if rateLimitedCount < 3 && successCount >= 7 {
		return true, nil // Rate limiting is missing (vulnerability exists)
	}

	return false, nil // Rate limiting appears to be present
}

func (sas *SecurityAuditService) checkAPIAuthentication(target string) (bool, error) {
	// Test API authentication weaknesses
	// Return true if weak/missing authentication is detected

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Common API endpoints to test
	endpoints := []string{
		"/api/v1/users",
		"/api/users",
		"/api/admin",
		"/api/data",
		"/graphql",
		"/api/health", // Sometimes unprotected
	}

	weakCredentials := []struct {
		username string
		password string
	}{
		{"admin", "admin"},
		{"admin", "password"},
		{"admin", "123456"},
		{"root", "root"},
		{"user", "user"},
		{"test", "test"},
		{"api", "api"},
		{"guest", "guest"},
	}

	for _, endpoint := range endpoints {
		url := "http://" + target + endpoint

		// Test 1: Check if endpoint is accessible without authentication
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			// If we get a 200 response, authentication might be missing
			if resp.StatusCode == 200 {
				// Additional check: look for sensitive data in response
				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)

				// Look for indicators of successful access to protected data
				if strings.Contains(bodyStr, "user") ||
					strings.Contains(bodyStr, "data") ||
					strings.Contains(bodyStr, "admin") ||
					strings.Contains(bodyStr, "secret") ||
					len(body) > 100 { // Substantial response likely contains data
					return true, nil // Authentication is weak/missing
				}
			}
		}

		// Test 2: Try basic authentication with weak credentials
		for _, creds := range weakCredentials {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				continue
			}

			req.SetBasicAuth(creds.username, creds.password)

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// If weak credentials work, authentication is weak
				if resp.StatusCode == 200 {
					return true, nil // Weak authentication detected
				}
			}
		}

		// Test 3: Check for common auth bypass patterns
		bypassUrls := []string{
			url + "?api_key=admin",
			url + "?token=admin",
			url + "?auth=admin",
			url + "?key=123456",
		}

		for _, bypassUrl := range bypassUrls {
			resp, err := client.Get(bypassUrl)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					body, _ := io.ReadAll(resp.Body)
					bodyStr := string(body)

					// Check if we got actual data (not an error page)
					if len(bodyStr) > 50 && !strings.Contains(strings.ToLower(bodyStr), "error") {
						return true, nil // Auth bypass possible
					}
				}
			}
		}
	}

	// Test 4: Check for exposed API keys or tokens in headers
	testResp, err := client.Get("http://" + target)
	if err == nil {
		defer testResp.Body.Close()

		// Look for API keys in response headers
		headersToCheck := []string{
			"X-API-Key",
			"Authorization",
			"X-Auth-Token",
			"API-Key",
		}

		for _, header := range headersToCheck {
			if testResp.Header.Get(header) != "" {
				return true, nil // Exposed authentication credentials
			}
		}
	}

	return false, nil // No weak authentication detected
}