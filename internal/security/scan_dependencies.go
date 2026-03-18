package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// scanGoDependencies scans Go module dependencies for known vulnerabilities
func (sas *SecurityAuditService) scanGoDependencies(ctx context.Context) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Known vulnerable Go packages and versions
	vulnerableDeps := map[string][]struct {
		version     string
		cve         string
		severity    string
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
									Severity:  vuln.severity,
									CVE:       vuln.cve,
									Category:  "dependency",
									Component: depName,
									Location:  "go.mod",
									Status:    "open",
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
		"golang.org/x/crypto":      "v0.17.0", // Latest as of 2024
		"github.com/gorilla/mux":   "v1.8.1",
		"github.com/lib/pq":        "v1.10.9",
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

// isVulnerableVersion checks if a version is vulnerable based on a known vulnerable version.
// Returns true if current is equal to or older than the known vulnerable version (semver order).
func isVulnerableVersion(currentVersion, vulnerableVersion string) bool {
	cur := canonicalSemver(currentVersion)
	vuln := canonicalSemver(vulnerableVersion)
	if cur == "" || vuln == "" {
		return currentVersion == vulnerableVersion
	}
	return semver.Compare(cur, vuln) <= 0
}

// compareVersions compares two semantic versions. Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2.
func compareVersions(v1, v2 string) int {
	c1 := canonicalSemver(v1)
	c2 := canonicalSemver(v2)
	if c1 == "" || c2 == "" {
		return strings.Compare(v1, v2)
	}
	return semver.Compare(c1, c2)
}

// canonicalSemver returns a version in canonical form for semver comparison (e.g. "v1.0.0").
// Handles Go pseudo-versions. Returns empty string if the version is not valid.
func canonicalSemver(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if v[0] != 'v' {
		v = "v" + v
	}
	if semver.IsValid(v) {
		return v
	}
	return ""
}

// scanFilesForPattern scans files matching a glob pattern for security issues.
// Pattern may be "*.go", "cmd/*.go", or "dir/**/*.go" (/**/ matches any directory depth).
func (sas *SecurityAuditService) scanFilesForPattern(globPattern string, patterns []struct {
	name        string
	pattern     string
	severity    string
	description string
	remediation string
}) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	files, err := globGoFiles(globPattern)
	if err != nil {
		return nil, err
	}

	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		content := string(data)
		lines := strings.Split(content, "\n")

		for _, p := range patterns {
			re, err := regexp.Compile(p.pattern)
			if err != nil {
				continue
			}

			for lineNum, line := range lines {
				if re.MatchString(line) {
					vulnerabilities = append(vulnerabilities, Vulnerability{
						ID:          generateVulnID(),
						Title:       fmt.Sprintf("Security Pattern: %s", p.name),
						Description: fmt.Sprintf("%s found in %s:%d", p.description, filePath, lineNum+1),
						Severity:    p.severity,
						Category:    "code",
						Component:   filePath,
						Location:    fmt.Sprintf("%s:%d", filePath, lineNum+1),
						Status:      "open",
						Remediation: p.remediation,
						Discovered:  time.Now(),
						Updated:     time.Now(),
					})
				}
			}
		}
	}

	return vulnerabilities, nil
}

// globGoFiles returns Go file paths matching the pattern.
// Supports "*.go" (all .go files under current dir via WalkDir), "dir/*.go" (filepath.Glob),
// and "dir/**/*.go" (WalkDir under dir). Skips vendor, .git, and node_modules.
func globGoFiles(pattern string) ([]string, error) {
	walkRoot := ""
	switch {
	case pattern == "*.go" || pattern == "":
		walkRoot = "."
	case strings.Contains(pattern, "**"):
		prefix := strings.TrimSuffix(strings.Split(pattern, "**")[0], "/")
		if prefix == "" {
			prefix = "."
		}
		walkRoot = prefix
	}
	if walkRoot != "" {
		var out []string
		err := filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				base := filepath.Base(path)
				if base == "vendor" || base == ".git" || base == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) == ".go" {
				out = append(out, path)
			}
			return nil
		})
		return out, err
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() && filepath.Ext(m) == ".go" {
			out = append(out, m)
		}
	}
	return out, nil
}
