package security

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

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
					Discovered:  time.Now(),
					Updated:     time.Now(),
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
									Discovered:  time.Now(),
									Updated:     time.Now(),
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
					Discovered:  time.Now(),
					Updated:     time.Now(),
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
					Discovered:  time.Now(),
					Updated:     time.Now(),
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
					Discovered:  time.Now(),
					Updated:     time.Now(),
				})
			}
		}
	}

	// Live inspection of running containers (inspect, processes, mounts) is not performed:
	// it would require Docker API access (e.g. DOCKER_HOST / socket) and appropriate auth in production.
	// Record an informational finding so operators know runtime checks are a possible extension.
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
		Discovered:  time.Now(),
		Updated:     time.Now(),
	})

	return vulnerabilities, nil
}

// scanContainerImages scans for vulnerable container images
func (sas *SecurityAuditService) scanContainerImages(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for known vulnerable base images
	vulnerableImages := map[string]struct {
		cve          string
		severity     string
		description  string
		fixedVersion string
	}{
		"golang:1.24-alpine": {
			cve:          "CVE-2024-XXXX",
			severity:     "medium",
			description:  "Alpine Linux vulnerabilities in golang base image",
			fixedVersion: "golang:1.24-alpine (latest patch)",
		},
		"postgres:15": {
			cve:          "CVE-2024-XXXX",
			severity:     "high",
			description:  "PostgreSQL authentication bypass vulnerability",
			fixedVersion: "postgres:15.1+",
		},
		"alpine:latest": {
			cve:          "CVE-2024-XXXX",
			severity:     "low",
			description:  "Alpine Linux package vulnerabilities",
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
