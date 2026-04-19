package security

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// checkSecurityHeaders checks for missing or misconfigured security headers
// by making a live HTTP request to target (when non-empty) and by scanning config files.
func (sas *SecurityAuditService) checkSecurityHeaders(target string) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	requiredHeaders := map[string]string{
		"X-Frame-Options":           "Prevents clickjacking attacks",
		"X-Content-Type-Options":    "Prevents MIME type sniffing",
		"Content-Security-Policy":   "Mitigates XSS and injection attacks",
		"X-XSS-Protection":          "Enables XSS filtering",
		"Strict-Transport-Security": "Enforces HTTPS connections",
	}

	// Live check: HTTP request to target when provided
	if target != "" {
		liveVulns, err := sas.checkSecurityHeadersLive(target, requiredHeaders)
		if err == nil {
			vulnerabilities = append(vulnerabilities, liveVulns...)
		}
	}

	// Config-based check: scan configuration files for header settings
	configFiles := []string{
		"Caddyfile", "nginx.conf", "apache.conf", "apache2.conf",
		".htaccess", "web.config", "server.go", "main.go",
	}
	for _, configFile := range configFiles {
		if data, err := os.ReadFile(configFile); err == nil {
			content := strings.ToLower(string(data))
			for header, description := range requiredHeaders {
				headerLower := strings.ToLower(header)
				if !strings.Contains(content, headerLower) {
					vulnerabilities = append(vulnerabilities, Vulnerability{
						ID:          generateVulnID(),
						Title:       fmt.Sprintf("Missing Security Header: %s", header),
						Description: fmt.Sprintf("Security header %s is not configured: %s", header, description),
						Severity:    "medium",
						Category:    "config",
						Component:   "web_server",
						Location:    configFile,
						Status:      "open",
						Remediation: fmt.Sprintf("Add %s header to server configuration", header),
						ReferenceUrls: []string{
							"https://owasp.org/www-project-secure-headers/",
						},
						Discovered: time.Now(),
						Updated:    time.Now(),
					})
				}
			}
		}
	}

	return vulnerabilities, nil
}

// checkSecurityHeadersLive performs an HTTP request to the target and reports missing security headers.
func (sas *SecurityAuditService) checkSecurityHeadersLive(target string, requiredHeaders map[string]string) ([]Vulnerability, error) {
	baseURL := target
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + target
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, baseURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}
	resp, err := client.Do(req)
	if err != nil {
		// Try HTTP if HTTPS failed (e.g. local dev or no TLS)
		if strings.HasPrefix(baseURL, "https://") {
			host := strings.TrimPrefix(strings.TrimPrefix(target, "https://"), "http://")
			baseURL = "http://" + host
			req2, _ := http.NewRequestWithContext(ctx, http.MethodHead, baseURL, nil)
			resp, err = client.Do(req2)
		}
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	var vulns []Vulnerability
	for header, description := range requiredHeaders {
		if resp.Header.Get(header) == "" {
			vulns = append(vulns, Vulnerability{
				ID:          generateVulnID(),
				Title:       fmt.Sprintf("Missing Security Header: %s", header),
				Description: fmt.Sprintf("Response from %s is missing %s: %s", baseURL, header, description),
				Severity:    "medium",
				Category:    "config",
				Component:   "web_server",
				Location:    baseURL,
				Status:      "open",
				Remediation: fmt.Sprintf("Add %s header to server configuration", header),
				ReferenceUrls: []string{
					"https://owasp.org/www-project-secure-headers/",
				},
				Discovered: time.Now(),
				Updated:    time.Now(),
			})
		}
	}
	return vulns, nil
}

// checkExposedServices checks for potentially dangerous exposed services
func (sas *SecurityAuditService) checkExposedServices(target string, ports []int) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Default ports to check if none specified
	if len(ports) == 0 {
		ports = []int{21, 23, 25, 53, 110, 143, 445, 993, 995, 3389, 5900}
	}

	// Dangerous exposed services
	dangerousServices := map[int]struct {
		name        string
		severity    string
		description string
		remediation string
	}{
		21:   {"FTP", "high", "FTP service exposed without encryption", "Use SFTP/SCP or disable FTP"},
		23:   {"Telnet", "critical", "Telnet service exposed (no encryption)", "Use SSH instead of Telnet"},
		25:   {"SMTP", "medium", "SMTP service potentially exposed", "Restrict SMTP access to authorized IPs"},
		445:  {"SMB", "high", "SMB/CIFS service exposed", "Restrict SMB access and use SMBv3+"},
		3389: {"RDP", "high", "Remote Desktop exposed to internet", "Restrict RDP access with firewall rules"},
		5900: {"VNC", "high", "VNC service exposed without encryption", "Use encrypted VNC or restrict access"},
	}

	for _, port := range ports {
		if service, exists := dangerousServices[port]; exists {
			if sas.isPortOpen(target, port) {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       fmt.Sprintf("Dangerous Service Exposed: %s (Port %d)", service.name, port),
					Description: fmt.Sprintf("%s: %s", service.description, service.name),
					Severity:    service.severity,
					Category:    "network",
					Component:   service.name,
					Location:    fmt.Sprintf("%s:%d", target, port),
					Status:      "open",
					Remediation: service.remediation,
					ReferenceUrls: []string{
						"https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks",
					},
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}

// checkInfrastructureMisconfigurations checks for common infrastructure misconfigurations
func (sas *SecurityAuditService) checkInfrastructureMisconfigurations(target string) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for common misconfiguration patterns in config files
	configChecks := []struct {
		file        string
		pattern     string
		title       string
		description string
		severity    string
		remediation string
	}{
		{
			"docker-compose.yml",
			"ports:\\s*-\\s*\"\\d+:",
			"Docker Port Exposed",
			"Docker container ports exposed to host without restrictions",
			"medium",
			"Use specific IP bindings or docker networks to restrict access",
		},
		{
			"Dockerfile",
			"FROM.*:latest",
			"Docker Latest Tag Used",
			"Using 'latest' tag can lead to unpredictable deployments",
			"low",
			"Use specific version tags for reproducible builds",
		},
		{
			".env",
			"DEBUG.*=.*true",
			"Debug Mode Enabled",
			"Debug mode enabled in production environment",
			"high",
			"Disable debug mode in production",
		},
		{
			"nginx.conf",
			"autoindex on",
			"Directory Listing Enabled",
			"Directory listing enabled in web server configuration",
			"medium",
			"Disable directory listing (autoindex off)",
		},
	}

	for _, check := range configChecks {
		if data, err := os.ReadFile(check.file); err == nil {
			content := string(data)
			re, err := regexp.Compile(check.pattern)
			if err != nil {
				continue
			}

			if re.MatchString(content) {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       check.title,
					Description: check.description,
					Severity:    check.severity,
					Category:    "config",
					Component:   check.file,
					Location:    check.file,
					Status:      "open",
					Remediation: check.remediation,
					Discovered:  time.Now(),
					Updated:     time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}

// checkServerSoftwareVersions checks for outdated server software
func (sas *SecurityAuditService) checkServerSoftwareVersions(target string) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for known vulnerable software versions in config files
	versionChecks := []struct {
		file       string
		software   string
		patterns   []string
		minVersion string
		severity   string
		cve        string
	}{
		{
			"go.mod",
			"Go",
			[]string{`go (\d+\.\d+)`, `go (\d+\.\d+\.\d+)`},
			"1.21",
			"medium",
			"SECURITY-VERSION-001",
		},
		{
			"package.json",
			"Node.js",
			[]string{`"node":\s*"([<>~^]?\d+\.\d+)`},
			"18.0",
			"medium",
			"SECURITY-VERSION-002",
		},
	}

	for _, check := range versionChecks {
		if data, err := os.ReadFile(check.file); err == nil {
			content := string(data)

			for _, pattern := range check.patterns {
				re, err := regexp.Compile(pattern)
				if err != nil {
					continue
				}

				matches := re.FindStringSubmatch(content)
				if len(matches) > 1 {
					version := matches[1]
					if compareVersions("v"+version, "v"+check.minVersion) < 0 {
						vulnerabilities = append(vulnerabilities, Vulnerability{
							ID:    generateVulnID(),
							Title: fmt.Sprintf("Outdated %s Version", check.software),
							Description: fmt.Sprintf("%s version %s is outdated (minimum recommended: %s)",
								check.software, version, check.minVersion),
							Severity:    check.severity,
							CVE:         check.cve,
							Category:    "config",
							Component:   check.software,
							Location:    check.file,
							Status:      "open",
							Remediation: fmt.Sprintf("Upgrade %s to version %s or later", check.software, check.minVersion),
							Discovered:  time.Now(),
							Updated:     time.Now(),
						})
					}
				}
			}
		}
	}

	return vulnerabilities, nil
}

// checkSensitiveDirectories checks for sensitive directories that might be exposed
func (sas *SecurityAuditService) checkSensitiveDirectories() ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for common directories that might expose sensitive information
	sensitiveDirs := []string{
		".git", ".env", "node_modules", ".DS_Store",
		"backup", "temp", "tmp", "logs", "config",
	}

	for _, dir := range sensitiveDirs {
		if _, err := os.Stat(dir); err == nil {
			// Directory exists, check if it should be exposed
			if dir == ".git" || dir == ".env" || strings.Contains(dir, "backup") {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       fmt.Sprintf("Sensitive Directory Exposed: %s", dir),
					Description: fmt.Sprintf("Directory '%s' contains sensitive information and should not be web-accessible", dir),
					Severity:    "high",
					Category:    "config",
					Component:   "filesystem",
					Location:    dir,
					Status:      "open",
					Remediation: "Configure web server to deny access to sensitive directories",
					ReferenceUrls: []string{
						"https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/",
					},
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}
