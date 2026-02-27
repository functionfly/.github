package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Common scanning functions and utilities

func (sas *SecurityAuditService) scanDependencies(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Scan Go module dependencies
	if goVulns, err := sas.scanGoDependencies(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, goVulns...)
	}

	// Scan for vulnerable patterns in dependency usage
	if patternVulns, err := sas.scanDependencyPatterns(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, patternVulns...)
	}

	// Check for outdated dependencies
	if outdatedVulns, err := sas.scanOutdatedDependencies(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, outdatedVulns...)
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) scanInfrastructure(ctx context.Context, target string, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check server configuration and headers
	if headerVulns, err := sas.checkSecurityHeaders(target); err == nil {
		vulnerabilities = append(vulnerabilities, headerVulns...)
	}

	// Check for exposed infrastructure services
	if exposedVulns, err := sas.checkExposedServices(target, config.IncludePorts); err == nil {
		vulnerabilities = append(vulnerabilities, exposedVulns...)
	}

	// Check for common infrastructure misconfigurations
	if configVulns, err := sas.checkInfrastructureMisconfigurations(target); err == nil {
		vulnerabilities = append(vulnerabilities, configVulns...)
	}

	// Check for outdated server software
	if softwareVulns, err := sas.checkServerSoftwareVersions(target); err == nil {
		vulnerabilities = append(vulnerabilities, softwareVulns...)
	}

	// Check for sensitive directories exposure
	if dirVulns, err := sas.checkSensitiveDirectories(); err == nil {
		vulnerabilities = append(vulnerabilities, dirVulns...)
	}

	return vulnerabilities, nil
}

// scanGoDependencies scans Go module dependencies for known vulnerabilities
func (sas *SecurityAuditService) scanGoDependencies(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Known vulnerable Go packages and versions
	vulnerableDeps := map[string][]struct {
		version string
		cve     string
		severity string
		description string
	}{
		"golang.org/x/crypto": {
			{"v0.0.0-20200109152110-bc719b9c8ecf", "CVE-2020-9283", "high", "SSH server code is vulnerable to integer overflow"},
			{"v0.0.0-20200221231518-2c99ac7514ef", "CVE-2020-7919", "high", "RSA decryption oracle vulnerability"},
		},
		"github.com/gorilla/websocket": {
			{"v1.4.2", "CVE-2020-27813", "medium", "WebSocket connections can be intercepted"},
		},
		"github.com/lib/pq": {
			{"v1.10.0", "CVE-2021-27607", "medium", "SQL injection via malicious column names"},
		},
		"golang.org/x/net": {
			{"v0.0.0-20210504180303-3f6846c8bc3d", "CVE-2021-33194", "high", "Request smuggling in net/http"},
		},
	}

	// Read go.mod file
	goModPath := "go.mod"
	if data, err := os.ReadFile(goModPath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "require (") || strings.HasPrefix(line, "\t") {
				// Parse dependency lines
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					depName := parts[0]
					depVersion := parts[1]

					// Check if this dependency has known vulnerabilities
					if vulnVersions, exists := vulnerableDeps[depName]; exists {
						for _, vuln := range vulnVersions {
							if isVulnerableVersion(depVersion, vuln.version) {
								vulnerabilities = append(vulnerabilities, Vulnerability{
									ID:          generateVulnID(),
									Title:       fmt.Sprintf("Vulnerable Dependency: %s", depName),
									Description: fmt.Sprintf("Dependency %s version %s contains known vulnerability %s: %s",
										depName, depVersion, vuln.cve, vuln.description),
									Severity:    vuln.severity,
									CVE:         vuln.cve,
									Category:    "dependency",
									Component:   depName,
									Location:    "go.mod",
									Status:      "open",
									Remediation: fmt.Sprintf("Update %s to a version newer than %s", depName, vuln.version),
									ReferenceUrls: []string{
										fmt.Sprintf("https://cve.mitre.org/cgi-bin/cvename.cgi?name=%s", vuln.cve),
										fmt.Sprintf("https://pkg.go.dev/%s", depName),
									},
									Discovered: time.Now(),
									Updated:    time.Now(),
								})
							}
						}
					}
				}
			}
		}
	}

	return vulnerabilities, nil
}

// scanDependencyPatterns scans for vulnerable usage patterns in dependencies
func (sas *SecurityAuditService) scanDependencyPatterns(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for potentially dangerous dependency usage patterns
	patterns := []struct {
		name        string
		pattern     string
		severity    string
		description string
		remediation string
	}{
		{
			"Unsafe JWT parsing",
			`jwt\.Parse\([^,)]+\)`,
			"medium",
			"JWT token parsed without signature verification",
			"Use jwt.ParseWithClaims with proper validation",
		},
		{
			"Unsafe HTML rendering",
			`template\.HTML\(`,
			"medium",
			"Potentially unsafe HTML rendering without sanitization",
			"Use html.EscapeString or proper HTML sanitization libraries",
		},
		{
			"Unsafe SQL query building",
			`fmt\.Sprintf.*SELECT.*%s`,
			"high",
			"SQL query built using string formatting, vulnerable to injection",
			"Use parameterized queries or prepared statements",
		},
		{
			"Unsafe exec command",
			`exec\.Command\([^,)]*\+`,
			"high",
			"Command execution with string concatenation, vulnerable to injection",
			"Use proper argument arrays instead of string concatenation",
		},
	}

	// Scan Go source files
	sourceFiles := []string{
		"*.go",
		"cmd/**/*.go",
		"internal/**/*.go",
		"pkg/**/*.go",
	}

	for _, pattern := range sourceFiles {
		if matches, err := sas.scanFilesForPattern(pattern, patterns); err == nil {
			vulnerabilities = append(vulnerabilities, matches...)
		}
	}

	return vulnerabilities, nil
}

// scanOutdatedDependencies checks for dependencies that haven't been updated recently
func (sas *SecurityAuditService) scanOutdatedDependencies(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for dependencies that might be outdated
	criticalDeps := map[string]string{
		"golang.org/x/crypto": "v0.17.0", // Latest as of 2024
		"github.com/gorilla/mux": "v1.8.1",
		"github.com/lib/pq": "v1.10.9",
		"github.com/golang-jwt/jwt": "v5.2.1",
	}

	goModPath := "go.mod"
	if data, err := os.ReadFile(goModPath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				depName := parts[0]
				depVersion := parts[1]

				if expectedVersion, isCritical := criticalDeps[depName]; isCritical {
					if compareVersions(depVersion, expectedVersion) < 0 {
						vulnerabilities = append(vulnerabilities, Vulnerability{
							ID:          generateVulnID(),
							Title:       fmt.Sprintf("Outdated Critical Dependency: %s", depName),
							Description: fmt.Sprintf("Dependency %s is outdated (current: %s, recommended: %s)", depName, depVersion, expectedVersion),
							Severity:    "medium",
							Category:    "dependency",
							Component:   depName,
							Location:    "go.mod",
							Status:      "open",
							Remediation: fmt.Sprintf("Update %s to version %s or later", depName, expectedVersion),
							ReferenceUrls: []string{
								fmt.Sprintf("https://pkg.go.dev/%s", depName),
							},
							Discovered: time.Now(),
							Updated:    time.Now(),
						})
					}
				}
			}
		}
	}

	return vulnerabilities, nil
}

// Helper functions

// isVulnerableVersion checks if a version is vulnerable based on a known vulnerable version
func isVulnerableVersion(currentVersion, vulnerableVersion string) bool {
	// Simple version comparison - in production, you'd want proper semver parsing
	current := strings.TrimPrefix(currentVersion, "v")
	vulnerable := strings.TrimPrefix(vulnerableVersion, "v")

	// If versions match exactly, it's vulnerable
	if current == vulnerable {
		return true
	}

	// Check if current version is within a vulnerable range
	// This is a simplified check - production systems should use proper semver
	return strings.HasPrefix(current, vulnerable[:len(vulnerable)-2])
}

// compareVersions compares two semantic versions
func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] != parts2[i] {
			// Simple numeric comparison (doesn't handle pre-release tags)
			if len(parts1[i]) > len(parts2[i]) {
				return 1
			} else if len(parts1[i]) < len(parts2[i]) {
				return -1
			}
			return strings.Compare(parts1[i], parts2[i])
		}
	}

	if len(parts1) > len(parts2) {
		return 1
	} else if len(parts1) < len(parts2) {
		return -1
	}

	return 0
}

// scanFilesForPattern scans files matching a glob pattern for security issues
func (sas *SecurityAuditService) scanFilesForPattern(pattern string, patterns []struct {
	name        string
	pattern     string
	severity    string
	description string
	remediation string
}) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// For this implementation, we'll check common Go files
	// In production, you'd want to use filepath.Glob or a proper file walker
	filesToCheck := []string{
		"main.go", "server.go", "api.go", "database.go",
		"internal/api/server.go", "internal/api/handlers.go",
		"cmd/main.go", "cmd/server/main.go",
	}

	for _, filePath := range filesToCheck {
		if data, err := os.ReadFile(filePath); err == nil {
			content := string(data)
			lines := strings.Split(content, "\n")

			for _, pattern := range patterns {
				re, err := regexp.Compile(pattern.pattern)
				if err != nil {
					continue
				}

				for lineNum, line := range lines {
					if re.MatchString(line) {
						vulnerabilities = append(vulnerabilities, Vulnerability{
							ID:          generateVulnID(),
							Title:       fmt.Sprintf("Security Pattern: %s", pattern.name),
							Description: fmt.Sprintf("%s found in %s:%d", pattern.description, filePath, lineNum+1),
							Severity:    pattern.severity,
							Category:    "code",
							Component:   filePath,
							Location:    fmt.Sprintf("%s:%d", filePath, lineNum+1),
							Status:      "open",
							Remediation: pattern.remediation,
							Discovered: time.Now(),
							Updated:    time.Now(),
						})
					}
				}
			}
		}
	}

	return vulnerabilities, nil
}

// Infrastructure scanning helper methods

// checkSecurityHeaders checks for missing or misconfigured security headers
func (sas *SecurityAuditService) checkSecurityHeaders(target string) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// This is a simplified check - in production you'd make an HTTP request
	// For now, we'll check configuration files for security header settings

	// Check for common configuration files
	configFiles := []string{
		"Caddyfile", "nginx.conf", "apache.conf", "apache2.conf",
		".htaccess", "web.config", "server.go", "main.go",
	}

	requiredHeaders := map[string]string{
		"X-Frame-Options":           "Prevents clickjacking attacks",
		"X-Content-Type-Options":    "Prevents MIME type sniffing",
		"Content-Security-Policy":   "Mitigates XSS and injection attacks",
		"X-XSS-Protection":          "Enables XSS filtering",
		"Strict-Transport-Security": "Enforces HTTPS connections",
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
		21:  {"FTP", "high", "FTP service exposed without encryption", "Use SFTP/SCP or disable FTP"},
		23:  {"Telnet", "critical", "Telnet service exposed (no encryption)", "Use SSH instead of Telnet"},
		25:  {"SMTP", "medium", "SMTP service potentially exposed", "Restrict SMTP access to authorized IPs"},
		445: {"SMB", "high", "SMB/CIFS service exposed", "Restrict SMB access and use SMBv3+"},
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
					Discovered: time.Now(),
					Updated:    time.Now(),
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
		file     string
		software string
		patterns []string
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
			"CVE-2023-XXX",
		},
		{
			"package.json",
			"Node.js",
			[]string{`"node":\s*"([<>~^]?\d+\.\d+)`},
			"18.0",
			"medium",
			"CVE-2023-XXX",
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
							ID:          generateVulnID(),
							Title:       fmt.Sprintf("Outdated %s Version", check.software),
							Description: fmt.Sprintf("%s version %s is outdated (minimum recommended: %s)",
								check.software, version, check.minVersion),
							Severity:    check.severity,
							CVE:         check.cve,
							Category:    "config",
							Component:   check.software,
							Location:    check.file,
							Status:      "open",
							Remediation: fmt.Sprintf("Upgrade %s to version %s or later", check.software, check.minVersion),
							Discovered: time.Now(),
							Updated:    time.Now(),
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

// scanContainers scans for container security issues
func (sas *SecurityAuditService) scanContainers(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Scan Dockerfile security issues
	if dockerfileVulns, err := sas.scanDockerfiles(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, dockerfileVulns...)
	}

	// Scan Docker Compose configuration
	if composeVulns, err := sas.scanDockerCompose(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, composeVulns...)
	}

	// Check for container runtime security issues
	if runtimeVulns, err := sas.scanContainerRuntime(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, runtimeVulns...)
	}

	// Scan for vulnerable base images
	if imageVulns, err := sas.scanContainerImages(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, imageVulns...)
	}

	return vulnerabilities, nil
}

// Container scanning helper methods

// scanDockerfiles scans Dockerfile security issues
func (sas *SecurityAuditService) scanDockerfiles(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	dockerfilePaths := []string{
		"Dockerfile",
		"deploy/docker/Dockerfile.orchestrator",
		"deploy/docker/Dockerfile.health-monitor",
		"deploy/caddy/Dockerfile",
		"edge-targets/fly/Dockerfile",
	}

	securityChecks := []struct {
		pattern     string
		title       string
		description string
		severity    string
		remediation string
	}{
		{
			`FROM.*:latest`,
			"Latest Tag Usage",
			"Using 'latest' tag can lead to unpredictable and potentially vulnerable container deployments",
			"medium",
			"Use specific version tags (e.g., ubuntu:22.04, golang:1.21-alpine)",
		},
		{
			`USER root`,
			"Running as Root User",
			"Container running as root user poses security risk",
			"high",
			"Create non-root user and switch to it with USER directive",
		},
		{
			`ADD `,
			"Using ADD Instead of COPY",
			"ADD can extract tar files and download remote URLs, introducing security risks",
			"low",
			"Use COPY for simple file copying, avoid downloading remote files in Dockerfile",
		},
		{
			`RUN apk add.*--no-cache.*sudo`,
			"Sudo Installed in Container",
			"Installing sudo in container can be used for privilege escalation",
			"high",
			"Remove sudo from container or use gosu/supersu for privilege management",
		},
		{
			`EXPOSE \d+`,
			"Port Exposure Without Documentation",
			"Ports exposed without clear documentation of their purpose",
			"info",
			"Document exposed ports in comments or use LABEL instructions",
		},
		{
			`RUN chmod.*777`,
			"World-Writable Permissions",
			"Files with 777 permissions allow any user to modify them",
			"high",
			"Use minimal required permissions (e.g., 755 for directories, 644 for files)",
		},
		{
			`ENV.*PASSWORD|ENV.*SECRET|ENV.*KEY`,
			"Hardcoded Secrets in Dockerfile",
			"Sensitive information stored as environment variables in Dockerfile",
			"critical",
			"Use secrets management systems or build-time arguments",
		},
	}

	for _, dockerfilePath := range dockerfilePaths {
		if data, err := os.ReadFile(dockerfilePath); err == nil {
			content := string(data)
			lines := strings.Split(content, "\n")

			for _, check := range securityChecks {
				re, err := regexp.Compile(check.pattern)
				if err != nil {
					continue
				}

				for lineNum, line := range lines {
					if re.MatchString(line) {
						vulnerabilities = append(vulnerabilities, Vulnerability{
							ID:          generateVulnID(),
							Title:       fmt.Sprintf("Dockerfile Security Issue: %s", check.title),
							Description: fmt.Sprintf("%s found in %s:%d", check.description, dockerfilePath, lineNum+1),
							Severity:    check.severity,
							Category:    "container",
							Component:   "dockerfile",
							Location:    fmt.Sprintf("%s:%d", dockerfilePath, lineNum+1),
							Status:      "open",
							Remediation: check.remediation,
							ReferenceUrls: []string{
								"https://docs.docker.com/develop/security-best-practices/",
							},
							Discovered: time.Now(),
							Updated:    time.Now(),
						})
					}
				}
			}

			// Check for multi-stage build security
			if strings.Contains(content, "AS builder") && !strings.Contains(content, "FROM builder") {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       "Multi-Stage Build Not Optimized",
					Description: "Multi-stage build defined but builder stage may not be properly isolated",
					Severity:    "low",
					Category:    "dockerfile",
					Component:   "dockerfile",
					Location:    dockerfilePath,
					Status:      "open",
					Remediation: "Ensure builder stage artifacts are not included in final image",
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}

// scanDockerCompose scans Docker Compose configuration for security issues
func (sas *SecurityAuditService) scanDockerCompose(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	composeFiles := []string{
		"docker-compose.yml",
		"docker-compose.monitoring.yml",
		"docker-compose.production.yml",
	}

	for _, composeFile := range composeFiles {
		if data, err := os.ReadFile(composeFile); err == nil {
			content := string(data)

			// Check for exposed ports without restrictions
			if strings.Contains(content, "ports:") {
				lines := strings.Split(content, "\n")
				for lineNum, line := range lines {
					if strings.Contains(line, "ports:") {
						// Look for the next few lines for port mappings
						for i := lineNum + 1; i < len(lines) && i < lineNum+10; i++ {
							portLine := strings.TrimSpace(lines[i])
							if strings.HasPrefix(portLine, "- ") && strings.Contains(portLine, ":") {
								vulnerabilities = append(vulnerabilities, Vulnerability{
									ID:          generateVulnID(),
									Title:       "Exposed Container Port",
									Description: fmt.Sprintf("Container port exposed to host in %s:%d", composeFile, i+1),
									Severity:    "medium",
									Category:    "container",
									Component:   "docker-compose",
									Location:    fmt.Sprintf("%s:%d", composeFile, i+1),
									Status:      "open",
									Remediation: "Restrict port exposure or use docker networks for internal communication",
									Discovered: time.Now(),
									Updated:    time.Now(),
								})
							}
						}
					}
				}
			}

			// Check for privileged containers
			if strings.Contains(content, "privileged: true") {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       "Privileged Container",
					Description: "Container running in privileged mode with full host access",
					Severity:    "critical",
					Category:    "container",
					Component:   "docker-compose",
					Location:    composeFile,
					Status:      "open",
					Remediation: "Remove privileged mode and use specific capabilities instead",
					ReferenceUrls: []string{
						"https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities",
					},
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}

			// Check for default/weak passwords
			if strings.Contains(content, "POSTGRES_PASSWORD") && strings.Contains(content, "postgres") {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       "Weak Database Password",
					Description: "Default or weak PostgreSQL password used in Docker Compose",
					Severity:    "high",
					Category:    "container",
					Component:   "database",
					Location:    composeFile,
					Status:      "open",
					Remediation: "Use strong, randomly generated passwords and environment variables",
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}

			// Check for host networking mode
			if strings.Contains(content, "network_mode: host") {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       "Host Network Mode",
					Description: "Container using host network mode, bypassing network isolation",
					Severity:    "high",
					Category:    "container",
					Component:   "docker-compose",
					Location:    composeFile,
					Status:      "open",
					Remediation: "Use bridge networking mode for better isolation",
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}

// scanContainerRuntime scans for container runtime security issues
func (sas *SecurityAuditService) scanContainerRuntime(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for Docker daemon socket mount (dangerous)
	runtimeChecks := []struct {
		pattern     string
		file        string
		title       string
		description string
		severity    string
		remediation string
	}{
		{
			"/var/run/docker.sock",
			"docker-compose.yml",
			"Docker Socket Mount",
			"Container mounting Docker daemon socket, allowing container escape",
			"critical",
			"Remove Docker socket mount or use Docker-in-Docker with proper isolation",
		},
		{
			"cap_add:",
			"docker-compose.yml",
			"Added Linux Capabilities",
			"Container granted additional Linux capabilities beyond default",
			"high",
			"Review if capabilities are necessary, use minimal required set",
		},
		{
			"security_opt:",
			"docker-compose.yml",
			"Custom Security Options",
			"Container using custom security options that may reduce isolation",
			"medium",
			"Review security options and ensure they don't compromise container security",
		},
	}

	for _, check := range runtimeChecks {
		if data, err := os.ReadFile(check.file); err == nil {
			content := string(data)
			if strings.Contains(content, check.pattern) {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnID(),
					Title:       check.title,
					Description: check.description,
					Severity:    check.severity,
					Category:    "container",
					Component:   "runtime",
					Location:    check.file,
					Status:      "open",
					Remediation: check.remediation,
					Discovered: time.Now(),
					Updated:    time.Now(),
				})
			}
		}
	}

	// Check for running containers with security issues
	// This would require Docker API access in production
	vulnerabilities = append(vulnerabilities, Vulnerability{
		ID:          generateVulnID(),
		Title:       "Container Runtime Security Check",
		Description: "Runtime container security analysis requires Docker API access",
		Severity:    "info",
		Category:    "container",
		Component:   "runtime",
		Location:    "docker",
		Status:      "open",
		Remediation: "Implement Docker API integration for runtime security scanning",
		Discovered: time.Now(),
		Updated:    time.Now(),
	})

	return vulnerabilities, nil
}

// scanContainerImages scans for vulnerable container images
func (sas *SecurityAuditService) scanContainerImages(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for known vulnerable base images
	vulnerableImages := map[string]struct {
		cve        string
		severity   string
		description string
		fixedVersion string
	}{
		"golang:1.24-alpine": {
			cve:         "CVE-2024-XXXX",
			severity:    "medium",
			description: "Alpine Linux vulnerabilities in golang base image",
			fixedVersion: "golang:1.24-alpine (latest patch)",
		},
		"postgres:15": {
			cve:         "CVE-2024-XXXX",
			severity:    "high",
			description: "PostgreSQL authentication bypass vulnerability",
			fixedVersion: "postgres:15.1+",
		},
		"alpine:latest": {
			cve:         "CVE-2024-XXXX",
			severity:    "low",
			description: "Alpine Linux package vulnerabilities",
			fixedVersion: "alpine:latest (with updates)",
		},
	}

	composeFiles := []string{
		"docker-compose.yml",
		"docker-compose.monitoring.yml",
		"docker-compose.production.yml",
	}

	for _, composeFile := range composeFiles {
		if data, err := os.ReadFile(composeFile); err == nil {
			content := string(data)

			for imageName, vuln := range vulnerableImages {
				if strings.Contains(content, imageName) {
					vulnerabilities = append(vulnerabilities, Vulnerability{
						ID:          generateVulnID(),
						Title:       fmt.Sprintf("Vulnerable Container Image: %s", imageName),
						Description: fmt.Sprintf("%s - %s", vuln.description, vuln.cve),
						Severity:    vuln.severity,
						CVE:         vuln.cve,
						Category:    "container",
						Component:   "image",
						Location:    composeFile,
						Status:      "open",
						Remediation: fmt.Sprintf("Update to %s", vuln.fixedVersion),
						ReferenceUrls: []string{
							fmt.Sprintf("https://cve.mitre.org/cgi-bin/cvename.cgi?name=%s", vuln.cve),
						},
						Discovered: time.Now(),
						Updated:    time.Now(),
					})
				}
			}
		}
	}

	// Check Dockerfiles for vulnerable base images
	dockerfilePaths := []string{
		"deploy/docker/Dockerfile.orchestrator",
		"deploy/docker/Dockerfile.health-monitor",
		"deploy/caddy/Dockerfile",
		"edge-targets/fly/Dockerfile",
	}

	for _, dockerfilePath := range dockerfilePaths {
		if data, err := os.ReadFile(dockerfilePath); err == nil {
			content := string(data)

			for imageName, vuln := range vulnerableImages {
				if strings.Contains(content, imageName) {
					vulnerabilities = append(vulnerabilities, Vulnerability{
						ID:          generateVulnID(),
						Title:       fmt.Sprintf("Vulnerable Base Image: %s", imageName),
						Description: fmt.Sprintf("%s - %s", vuln.description, vuln.cve),
						Severity:    vuln.severity,
						CVE:         vuln.cve,
						Category:    "container",
						Component:   "dockerfile",
						Location:    dockerfilePath,
						Status:      "open",
						Remediation: fmt.Sprintf("Update FROM instruction to %s", vuln.fixedVersion),
						ReferenceUrls: []string{
							fmt.Sprintf("https://cve.mitre.org/cgi-bin/cvename.cgi?name=%s", vuln.cve),
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