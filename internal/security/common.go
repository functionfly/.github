package security

import (
	"context"
)

// Common scanning orchestration functions
// Detailed implementations are in scan_dependencies.go, scan_infrastructure.go, and scan_containers.go

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
