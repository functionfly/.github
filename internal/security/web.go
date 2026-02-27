package security

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Web security scanning functions

func (sas *SecurityAuditService) scanWebSecurity(ctx context.Context, target string, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for common web vulnerabilities
	headers := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Content-Security-Policy",
		"Strict-Transport-Security",
	}

	for _, header := range headers {
		if missing, err := sas.checkSecurityHeader(target, header); err == nil && missing {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       fmt.Sprintf("Missing Security Header: %s", header),
				Description: fmt.Sprintf("The %s security header is not present in HTTP responses", header),
				Severity:    "medium",
				Category:    "config",
				Component:   "web_server",
				Location:    target,
				Status:      "open",
				Remediation: fmt.Sprintf("Add %s header to all HTTP responses", header),
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	// Check for directory listing
	if dirListing, err := sas.checkDirectoryListing(target); err == nil && dirListing {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Directory Listing Enabled",
			Description: "Web server is configured to show directory contents",
			Severity:    "medium",
			Category:    "config",
			Component:   "web_server",
			Location:    target,
			Status:      "open",
			Remediation: "Disable directory listing in web server configuration",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkSecurityHeader(target, header string) (bool, error) {
	// Make HTTP request and check for header
	resp, err := http.Get("http://" + target)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.Header.Get(header) == "" {
		return true, nil // Header is missing
	}
	return false, nil
}

func (sas *SecurityAuditService) checkDirectoryListing(target string) (bool, error) {
	// Check for directory listing vulnerability
	resp, err := http.Get("http://" + target)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	bodyStr := string(body)

	// Look for directory listing indicators in response
	// Common patterns from different web servers

	// Apache-style directory listing
	if strings.Contains(bodyStr, "<title>Index of") ||
		strings.Contains(bodyStr, "Parent Directory") ||
		strings.Contains(bodyStr, "Last modified") ||
		strings.Contains(bodyStr, "Size") ||
		strings.Contains(bodyStr, "Description") {
		return true, nil
	}

	// IIS-style directory listing
	if strings.Contains(bodyStr, "<dir>") ||
		strings.Contains(bodyStr, "<directory>") ||
		strings.Contains(bodyStr, "Directory Listing") {
		return true, nil
	}

	// nginx autoindex style
	if strings.Contains(bodyStr, "Index of /") ||
		strings.Contains(bodyStr, "directory listing") {
		return true, nil
	}

	// Look for file extension patterns that indicate directory listing
	// Multiple files with common extensions
	extensions := []string{".html", ".php", ".js", ".css", ".txt", ".md", ".json", ".xml"}
	fileCount := 0
	for _, ext := range extensions {
		if strings.Contains(bodyStr, ext) {
			fileCount++
		}
	}

	// If we find multiple file extensions, likely directory listing
	if fileCount >= 3 {
		// Additional check: look for file size patterns (e.g., "123 bytes", "1.2 KB")
		if strings.Contains(bodyStr, "bytes") ||
			strings.Contains(bodyStr, "KB") ||
			strings.Contains(bodyStr, "MB") {
			return true, nil
		}
	}

	return false, nil
}