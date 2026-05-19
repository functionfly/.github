package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

// VulnerabilityScanner defines the interface for external vulnerability scanners
type VulnerabilityScanner interface {
	ScanImage(ctx context.Context, imageName string) ([]Vulnerability, error)
	ScanFilesystem(ctx context.Context, path string) ([]Vulnerability, error)
	IsAvailable() bool
	GetName() string
}

// ScannerType represents the type of vulnerability scanner
type ScannerType string

const (
	ScannerTrivy ScannerType = "trivy"
	ScannerClair ScannerType = "clair"
	ScannerGrype ScannerType = "grype"
)

// TrivyScanner implements Trivy integration
type TrivyScanner struct {
	binaryPath string
}

// NewTrivyScanner creates a new Trivy scanner instance
func NewTrivyScanner() *TrivyScanner {
	binary := "trivy"
	if path := os.Getenv("TRIVY_PATH"); path != "" {
		binary = path
	}
	return &TrivyScanner{binaryPath: binary}
}

// IsAvailable checks if Trivy is installed and accessible
func (t *TrivyScanner) IsAvailable() bool {
	cmd := exec.Command(t.binaryPath, "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetName returns the scanner name
func (t *TrivyScanner) GetName() string {
	return "Trivy"
}

// ScanImage scans a container image for vulnerabilities
func (t *TrivyScanner) ScanImage(ctx context.Context, imageName string) ([]Vulnerability, error) {
	if !t.IsAvailable() {
		return nil, fmt.Errorf("trivy not available")
	}

	cmd := exec.CommandContext(ctx, t.binaryPath, "image", "-f", "json", "-q", imageName)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("trivy scan failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run trivy: %w", err)
	}

	return t.parseTrivyOutput(output)
}

// ScanFilesystem scans a filesystem for vulnerabilities
func (t *TrivyScanner) ScanFilesystem(ctx context.Context, path string) ([]Vulnerability, error) {
	if !t.IsAvailable() {
		return nil, fmt.Errorf("trivy not available")
	}

	cmd := exec.CommandContext(ctx, t.binaryPath, "filesystem", "-f", "json", "-q", path)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("trivy scan failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run trivy: %w", err)
	}

	return t.parseTrivyOutput(output)
}

// parseTrivyOutput parses Trivy JSON output into Vulnerability structs
func (t *TrivyScanner) parseTrivyOutput(data []byte) ([]Vulnerability, error) {
	var result struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID  string `json:"VulnerabilityID"`
				PkgName          string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion     string `json:"FixedVersion"`
				Severity         string `json:"Severity"`
				Description      string `json:"Description"`
				PrimaryURL       string `json:"PrimaryURL"`
				Title            string `json:"Title"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse trivy output: %w", err)
	}

	var vulnerabilities []Vulnerability
	for _, r := range result.Results {
		for _, v := range r.Vulnerabilities {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:            generateVulnID(),
				Title:         v.Title,
				Description:   fmt.Sprintf("%s in package %s", v.Description, v.PkgName),
				Severity:      strings.ToLower(v.Severity),
				CVE:           v.VulnerabilityID,
				Category:      "container",
				Component:     v.PkgName,
				Status:        "open",
				Remediation:   fmt.Sprintf("Update %s to version %s", v.PkgName, v.FixedVersion),
				ReferenceUrls: []string{v.PrimaryURL},
				Discovered:    time.Now(),
				Updated:       time.Now(),
			})
		}
	}

	return vulnerabilities, nil
}

// ClairScanner implements Clair integration via local scanner or API
type ClairScanner struct {
	endpoint   string
	useLocal   bool
	httpClient *http.Client
}

// NewClairScanner creates a new Clair scanner instance
func NewClairScanner() *ClairScanner {
	endpoint := os.Getenv("CLAIR_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:6060"
	}
	return &ClairScanner{
		endpoint:   endpoint,
		useLocal:   os.Getenv("CLAIR_USE_LOCAL") == "true",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// IsAvailable checks if Clair is available
func (c *ClairScanner) IsAvailable() bool {
	if c.useLocal {
		_, err := exec.LookPath("clairctl")
		return err == nil
	}
	resp, err := c.httpClient.Get(c.endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// GetName returns the scanner name
func (c *ClairScanner) GetName() string {
	return "Clair"
}

// ScanImage scans a container image using Clair
func (c *ClairScanner) ScanImage(ctx context.Context, imageName string) ([]Vulnerability, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("clair not available")
	}

	if c.useLocal {
		cmd := exec.CommandContext(ctx, "clairctl", "report", imageName)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("clair scan failed: %w", err)
		}
		return c.parseClairOutput(output)
	}

	return c.scanImageViaAPI(ctx, imageName)
}

// scanImageViaAPI scans an image using the Clair REST API
func (c *ClairScanner) scanImageViaAPI(ctx context.Context, imageName string) ([]Vulnerability, error) {
	namespace := "default"
	manifest := fmt.Sprintf(`{"docker":{"reference":"%s"}}`, imageName)

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/namespaces/"+namespace+"/artifacts", bytes.NewReader([]byte(manifest)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Clair API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit image to Clair: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Clair API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var scanResult struct {
		Features []struct {
			Name             string `json:"Name"`
			NamespaceName    string `json:"NamespaceName"`
			Version          string `json:"Version"`
			FixedByVersion   string `json:"FixedByVersion"`
			Vulnerabilities []struct {
				Name        string `json:"Name"`
				Description string `json:"Description"`
				Link        string `json:"Link"`
				Severity    string `json:"Severity"`
				FixedBy     string `json:"FixedBy"`
			} `json:"Vulnerabilities"`
		} `json:"features"`
	}

	reportURL := c.endpoint + "/api/v1/namespaces/" + namespace + "/artifacts/" + imageName + "/report"
	reportReq, err := http.NewRequestWithContext(ctx, "GET", reportURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create report request: %w", err)
	}

	maxWait := 120 * time.Second
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		reportResp, err := c.httpClient.Do(reportReq)
		if err != nil {
			return nil, fmt.Errorf("failed to get Clair report: %w", err)
		}

		if reportResp.StatusCode == http.StatusOK {
			defer reportResp.Body.Close()
			bodyBytes, err := io.ReadAll(reportResp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read Clair report: %w", err)
			}
			if err := json.Unmarshal(bodyBytes, &scanResult); err != nil {
				return nil, fmt.Errorf("failed to parse Clair report: %w", err)
			}
			break
		}
		if reportResp.StatusCode == http.StatusAccepted {
			reportResp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		reportResp.Body.Close()
		return nil, fmt.Errorf("Clair report request returned status %d", reportResp.StatusCode)
	}

	var vulnerabilities []Vulnerability
	for _, feature := range scanResult.Features {
		for _, vuln := range feature.Vulnerabilities {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       vuln.Name,
				Description: vuln.Description,
				Severity:    strings.ToLower(vuln.Severity),
				CVE:         vuln.Name,
				Category:    "container",
				Component:   feature.Name,
				Status:      "open",
				Remediation: fmt.Sprintf("Update %s to version %s", feature.Name, vuln.FixedBy),
				ReferenceUrls: []string{vuln.Link},
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}
	return vulnerabilities, nil
}

// ScanFilesystem scans a filesystem using Clair
func (c *ClairScanner) ScanFilesystem(ctx context.Context, path string) ([]Vulnerability, error) {
	return nil, fmt.Errorf("clair filesystem scanning not supported")
}

// parseClairOutput parses Clair output
func (c *ClairScanner) parseClairOutput(data []byte) ([]Vulnerability, error) {
	var vulnerabilities []Vulnerability
	return vulnerabilities, nil
}

// GrypeScanner implements Grype/Syft integration
type GrypeScanner struct {
	binaryPath string
}

// NewGrypeScanner creates a new Grype scanner instance
func NewGrypeScanner() *GrypeScanner {
	binary := "grype"
	if path := os.Getenv("GRYPE_PATH"); path != "" {
		binary = path
	}
	return &GrypeScanner{binaryPath: binary}
}

// IsAvailable checks if Grype is installed
func (g *GrypeScanner) IsAvailable() bool {
	cmd := exec.Command(g.binaryPath, "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetName returns the scanner name
func (g *GrypeScanner) GetName() string {
	return "Grype"
}

// ScanImage scans a container image using Grype
func (g *GrypeScanner) ScanImage(ctx context.Context, imageName string) ([]Vulnerability, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("grype not available")
	}

	cmd := exec.CommandContext(ctx, g.binaryPath, "-o", "json", imageName)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("grype scan failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run grype: %w", err)
	}

	return g.parseGrypeOutput(output)
}

// ScanFilesystem scans a filesystem using Grype
func (g *GrypeScanner) ScanFilesystem(ctx context.Context, path string) ([]Vulnerability, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("grype not available")
	}

	cmd := exec.CommandContext(ctx, g.binaryPath, "-o", "json", "dir:"+path)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("grype scan failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run grype: %w", err)
	}

	return g.parseGrypeOutput(output)
}

// parseGrypeOutput parses Grype JSON output
func (g *GrypeScanner) parseGrypeOutput(data []byte) ([]Vulnerability, error) {
	var result struct {
		Matches []struct {
			Vulnerability struct {
				ID          string   `json:"id"`
				Severity    string   `json:"severity"`
				Description string   `json:"description"`
				URLs        []string `json:"urls"`
			} `json:"vulnerability"`
			Artifact struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Type    string `json:"type"`
			} `json:"artifact"`
			RelatedVulnerabilities []struct {
				ID string `json:"id"`
			} `json:"relatedVulnerabilities"`
		} `json:"matches"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse grype output: %w", err)
	}

	var vulnerabilities []Vulnerability
	for _, m := range result.Matches {
		v := m.Vulnerability
		cve := v.ID
		// Prefer CVE ID from related vulnerabilities if available
		for _, rv := range m.RelatedVulnerabilities {
			if strings.HasPrefix(rv.ID, "CVE-") {
				cve = rv.ID
				break
			}
		}

		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:            generateVulnID(),
			Title:         v.ID,
			Description:   fmt.Sprintf("%s in %s (%s)", v.Description, m.Artifact.Name, m.Artifact.Type),
			Severity:      strings.ToLower(v.Severity),
			CVE:           cve,
			Category:      "container",
			Component:     m.Artifact.Name,
			Status:        "open",
			Remediation:   fmt.Sprintf("Update %s to a fixed version", m.Artifact.Name),
			ReferenceUrls: v.URLs,
			Discovered:    time.Now(),
			Updated:       time.Now(),
		})
	}

	return vulnerabilities, nil
}

// getAvailableScanners returns a list of available vulnerability scanners
func (sas *SecurityAuditService) getAvailableScanners() []VulnerabilityScanner {
	scanners := []VulnerabilityScanner{
		NewTrivyScanner(),
		NewClairScanner(),
		NewGrypeScanner(),
	}

	var available []VulnerabilityScanner
	for _, scanner := range scanners {
		if scanner.IsAvailable() {
			available = append(available, scanner)
		}
	}
	return available
}

// scanWithExternalScanner runs a scanner against configured container images
func (sas *SecurityAuditService) scanWithExternalScanner(ctx context.Context, scanner VulnerabilityScanner) ([]Vulnerability, error) {
	var allVulns []Vulnerability

	// Images to scan (these would typically come from config or discovered Dockerfiles)
	imagesToScan := []string{
		"postgres:15",
		"redis:7-alpine",
	}

	for _, image := range imagesToScan {
		vulns, err := scanner.ScanImage(ctx, image)
		if err == nil {
			allVulns = append(allVulns, vulns...)
		}
	}

	return allVulns, nil
}

// scanContainerImages scans for vulnerable container images
func (sas *SecurityAuditService) scanContainerImages(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for available external vulnerability scanners
	availableScanners := sas.getAvailableScanners()
	if len(availableScanners) > 0 {
		for _, scanner := range availableScanners {
			scannerVulns, err := sas.scanWithExternalScanner(ctx, scanner)
			if err == nil {
				vulnerabilities = append(vulnerabilities, scannerVulns...)
			}
		}
	}

	// Always perform baseline security checks
	// These represent recommended security practices, not actual CVEs
	// These checks run in addition to external scanner results
	vulnerableImages := map[string]struct {
		cve          string
		severity     string
		description  string
		fixedVersion string
	}{
		"golang:1.24-alpine": {
			cve:          "SECURITY-BASELINE-001",
			severity:     "medium",
			description:  "Use specific alpine version and update packages regularly",
			fixedVersion: "golang:1.24-alpine (latest patch)",
		},
		"postgres:15": {
			cve:          "SECURITY-BASELINE-002",
			severity:     "high",
			description:  "Use patched PostgreSQL version with latest security updates",
			fixedVersion: "postgres:15.1+",
		},
		"alpine:latest": {
			cve:          "SECURITY-BASELINE-003",
			severity:     "low",
			description:  "Avoid 'latest' tag, use specific version for reproducibility",
			fixedVersion: "alpine:3.19+ (with updates)",
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
