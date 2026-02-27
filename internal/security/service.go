package security

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SecurityAuditService handles security auditing and penetration testing
type SecurityAuditService struct {
	db     storage.Repository
	logger *logrus.Logger
}

// NewSecurityAuditService creates a new security audit service
func NewSecurityAuditService(db storage.Repository) *SecurityAuditService {
	return &SecurityAuditService{
		db:     db,
		logger: logrus.New(),
	}
}

// PenetrationTest performs comprehensive penetration testing
func (sas *SecurityAuditService) PenetrationTest(ctx context.Context, target string, config ScanConfig) (*SecurityScan, error) {
	scan := &SecurityScan{
		ID:        generateScanID(),
		Type:      "penetration_test",
		Status:    "running",
		Target:    target,
		StartedAt: time.Now(),
		Config:    config,
		Vulnerabilities: []Vulnerability{},
	}

	sas.logger.WithFields(logrus.Fields{
		"scan_id": scan.ID,
		"target":  target,
		"type":    scan.Type,
	}).Info("Starting penetration test")

	// Perform comprehensive security testing
	vulnerabilities := []Vulnerability{}
	successfulScans := 0
	totalScans := 5 // Network, Web, API, Database, Configuration

	// 1. Network reconnaissance
	if networkVulns, err := sas.scanNetworkSecurity(ctx, target, config); err == nil {
		vulnerabilities = append(vulnerabilities, networkVulns...)
		successfulScans++
	}

	// 2. Web application security
	if webVulns, err := sas.scanWebSecurity(ctx, target, config); err == nil {
		vulnerabilities = append(vulnerabilities, webVulns...)
		successfulScans++
	}

	// 3. API security testing
	if apiVulns, err := sas.scanAPISecurity(ctx, target, config); err == nil {
		vulnerabilities = append(vulnerabilities, apiVulns...)
		successfulScans++
	}

	// 4. Database security
	if dbVulns, err := sas.scanDatabaseSecurity(ctx, config); err == nil {
		vulnerabilities = append(vulnerabilities, dbVulns...)
		successfulScans++
	}

	// 5. Configuration analysis
	if configVulns, err := sas.scanConfigurationSecurity(ctx, config); err == nil {
		vulnerabilities = append(vulnerabilities, configVulns...)
		successfulScans++
	}

	// Calculate coverage based on successful scan executions
	coverage := float64(successfulScans) / float64(totalScans) * 100.0

	scan.Vulnerabilities = vulnerabilities
	scan.CompletedAt = &time.Time{}
	*scan.CompletedAt = time.Now()
	scan.Duration = scan.CompletedAt.Sub(scan.StartedAt)
	scan.Status = "completed"
	scan.Summary = sas.generateScanSummary(vulnerabilities, coverage)

	// Log scan results
	sas.logScanResults(scan)

	return scan, nil
}

// VulnerabilityScan performs automated vulnerability scanning
func (sas *SecurityAuditService) VulnerabilityScan(ctx context.Context, target string, config ScanConfig) (*SecurityScan, error) {
	scan := &SecurityScan{
		ID:        generateScanID(),
		Type:      "vulnerability_scan",
		Status:    "running",
		Target:    target,
		StartedAt: time.Now(),
		Config:    config,
		Vulnerabilities: []Vulnerability{},
	}

	sas.logger.WithFields(logrus.Fields{
		"scan_id": scan.ID,
		"target":  target,
		"type":    scan.Type,
	}).Info("Starting vulnerability scan")

	// Perform vulnerability scanning
	vulnerabilities := []Vulnerability{}
	successfulScans := 0
	totalScans := 3 // Dependencies, Infrastructure, Containers

	// Check for known vulnerabilities in dependencies
	if depVulns, err := sas.scanDependencies(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, depVulns...)
		successfulScans++
	}

	// Infrastructure vulnerability scanning
	if infraVulns, err := sas.scanInfrastructure(ctx, target, config); err == nil {
		vulnerabilities = append(vulnerabilities, infraVulns...)
		successfulScans++
	}

	// Container security scanning
	if containerVulns, err := sas.scanContainers(ctx); err == nil {
		vulnerabilities = append(vulnerabilities, containerVulns...)
		successfulScans++
	}

	// Calculate coverage based on successful scan executions
	coverage := float64(successfulScans) / float64(totalScans) * 100.0

	scan.Vulnerabilities = vulnerabilities
	scan.CompletedAt = &time.Time{}
	*scan.CompletedAt = time.Now()
	scan.Duration = scan.CompletedAt.Sub(scan.StartedAt)
	scan.Status = "completed"
	scan.Summary = sas.generateScanSummary(vulnerabilities, coverage)

	sas.logScanResults(scan)

	return scan, nil
}

// ComplianceCheck performs compliance auditing
func (sas *SecurityAuditService) ComplianceCheck(ctx context.Context, framework string) (*SecurityScan, error) {
	scan := &SecurityScan{
		ID:        generateScanID(),
		Type:      "compliance_check",
		Status:    "running",
		Target:    framework,
		StartedAt: time.Now(),
		Config:    ScanConfig{Timeout: 30 * time.Minute},
		Vulnerabilities: []Vulnerability{},
	}

	sas.logger.WithFields(logrus.Fields{
		"scan_id": scan.ID,
		"framework": framework,
		"type":    scan.Type,
	}).Info("Starting compliance check")

	vulnerabilities := []Vulnerability{}
	coverage := 100.0 // Compliance checks are all-or-nothing

	switch framework {
	case "soc2", "SOC2":
		vulnerabilities = sas.checkSOC2Compliance(ctx)
	case "iso27001", "ISO27001":
		vulnerabilities = sas.checkISO27001Compliance(ctx)
	case "gdpr", "GDPR":
		vulnerabilities = sas.checkGDPRCompliance(ctx)
	case "pci-dss", "PCI-DSS":
		vulnerabilities = sas.checkPCIDSSCompliance(ctx)
	default:
		return nil, fmt.Errorf("unsupported compliance framework: %s", framework)
	}

	scan.Vulnerabilities = vulnerabilities
	scan.CompletedAt = &time.Time{}
	*scan.CompletedAt = time.Now()
	scan.Duration = scan.CompletedAt.Sub(scan.StartedAt)
	scan.Status = "completed"
	scan.Summary = sas.generateScanSummary(vulnerabilities, coverage)

	sas.logScanResults(scan)

	return scan, nil
}

// SaveScanResult saves scan results to database
func (sas *SecurityAuditService) SaveScanResult(ctx context.Context, scan *SecurityScan) error {
	// Convert scan to audit event
	resourceID := uuid.New()
	event := &storage.AuditEvent{
		Action:     fmt.Sprintf("security.%s", scan.Type),
		ResourceType: "security_scan",
		ResourceID:  &resourceID,
		BeforeState: nil,
		AfterState:  scan,
		Success:     scan.Status == "completed",
	}

	// Add metadata about vulnerabilities found
	if len(scan.Vulnerabilities) > 0 {
		event.AfterState = map[string]interface{}{
			"scan": scan,
			"vulnerability_summary": scan.Summary,
		}
	}

	return sas.db.LogAuditEvent(ctx, event)
}

// GetScanResults retrieves scan results
func (sas *SecurityAuditService) GetScanResults(ctx context.Context, limit, offset int) ([]*SecurityScan, error) {
	// Query scan results from database
	scans, err := sas.db.ListSecurityScans(limit, offset, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan results: %w", err)
	}

	// Convert database models to service models
	result := make([]*SecurityScan, len(scans))
	for i, scan := range scans {
		// Convert ScanSummary from map to struct
		summary := ScanSummary{}
		if scan.Summary != nil {
			if totalVulns, ok := scan.Summary["total_vulnerabilities"]; ok {
				if v, ok := totalVulns.(float64); ok {
					summary.TotalVulnerabilities = int(v)
				}
			}
			if criticalCount, ok := scan.Summary["critical_count"]; ok {
				if v, ok := criticalCount.(float64); ok {
					summary.CriticalCount = int(v)
				}
			}
			if highCount, ok := scan.Summary["high_count"]; ok {
				if v, ok := highCount.(float64); ok {
					summary.HighCount = int(v)
				}
			}
			if mediumCount, ok := scan.Summary["medium_count"]; ok {
				if v, ok := mediumCount.(float64); ok {
					summary.MediumCount = int(v)
				}
			}
			if lowCount, ok := scan.Summary["low_count"]; ok {
				if v, ok := lowCount.(float64); ok {
					summary.LowCount = int(v)
				}
			}
			if infoCount, ok := scan.Summary["info_count"]; ok {
				if v, ok := infoCount.(float64); ok {
					summary.InfoCount = int(v)
				}
			}
			if riskScore, ok := scan.Summary["risk_score"]; ok {
				if v, ok := riskScore.(float64); ok {
					summary.RiskScore = v
				}
			}
			if coverage, ok := scan.Summary["coverage_percentage"]; ok {
				if v, ok := coverage.(float64); ok {
					summary.Coverage = v
				}
			}
			if complianceScore, ok := scan.Summary["compliance_score"]; ok {
				if v, ok := complianceScore.(float64); ok {
					summary.ComplianceScore = v
				}
			}
		}

		// Convert ScanConfig from map to struct
		config := ScanConfig{}
		if scan.Config != nil {
			if ports, ok := scan.Config["include_ports"]; ok {
				if portsSlice, ok := ports.([]interface{}); ok {
					for _, p := range portsSlice {
						if port, ok := p.(float64); ok {
							config.IncludePorts = append(config.IncludePorts, int(port))
						}
					}
				}
			}
			if paths, ok := scan.Config["exclude_paths"]; ok {
				if pathsSlice, ok := paths.([]interface{}); ok {
					for _, p := range pathsSlice {
						if path, ok := p.(string); ok {
							config.ExcludePaths = append(config.ExcludePaths, path)
						}
					}
				}
			}
			if creds, ok := scan.Config["auth_credentials"]; ok {
				if credsMap, ok := creds.(map[string]interface{}); ok {
					config.AuthCredentials = make(map[string]string)
					for k, v := range credsMap {
						if str, ok := v.(string); ok {
							config.AuthCredentials[k] = str
						}
					}
				}
			}
			if timeout, ok := scan.Config["timeout"]; ok {
				if t, ok := timeout.(float64); ok {
					config.Timeout = time.Duration(t) * time.Second
				}
			}
			if concurrency, ok := scan.Config["max_concurrency"]; ok {
				if c, ok := concurrency.(float64); ok {
					config.MaxConcurrency = int(c)
				}
			}
			if depth, ok := scan.Config["depth"]; ok {
				if d, ok := depth.(float64); ok {
					config.Depth = int(d)
				}
			}
		}

		// Calculate duration if completed
		var duration time.Duration
		if scan.CompletedAt != nil {
			if scan.DurationMs != nil {
				duration = time.Duration(*scan.DurationMs) * time.Millisecond
			} else {
				duration = scan.CompletedAt.Sub(scan.StartedAt)
			}
		}

		// Load vulnerabilities for this scan
		scanVulns, err := sas.db.GetVulnerabilities(map[string]interface{}{
			"scan_id": scan.ID,
		})
		if err != nil {
			sas.logger.WithError(err).WithField("scan_id", scan.ID).Warn("Failed to load vulnerabilities for scan")
			scanVulns = []*storage.Vulnerability{} // Continue with empty list
		}

		// Convert database vulnerabilities to service vulnerabilities
		vulnerabilities := make([]Vulnerability, len(scanVulns))
		for j, vuln := range scanVulns {
			// Handle pointer fields
			var cvss float64
			if vuln.CVSS != nil {
				cvss = *vuln.CVSS
			}

			var cve string
			if vuln.CVE != nil {
				cve = *vuln.CVE
			}

			var location string
			if vuln.Location != nil {
				location = *vuln.Location
			}

			var remediation string
			if vuln.Remediation != nil {
				remediation = *vuln.Remediation
			}

			vulnerabilities[j] = Vulnerability{
				ID:          vuln.ID.String(),
				Title:       vuln.Title,
				Description: vuln.Description,
				Severity:    vuln.Severity,
				CVSS:        cvss,
				CVE:         cve,
				Category:    vuln.Category,
				Component:   vuln.Component,
				Location:    location,
				Status:      vuln.Status,
				Remediation: remediation,
				ReferenceUrls: vuln.ReferenceUrls,
				Metadata:    vuln.Metadata,
				Discovered:  vuln.DiscoveredAt,
				Updated:     vuln.UpdatedAt,
			}
		}

		result[i] = &SecurityScan{
			ID:             scan.ID.String(),
			Type:           scan.ScanType,
			Status:         scan.Status,
			Target:         scan.Target,
			StartedAt:      scan.StartedAt,
			CompletedAt:    scan.CompletedAt,
			Duration:       duration,
			Vulnerabilities: vulnerabilities,
			Summary:        summary,
			Config:         config,
		}
	}

	return result, nil
}

// GetVulnerabilities retrieves vulnerabilities with filtering
func (sas *SecurityAuditService) GetVulnerabilities(ctx context.Context, filters map[string]interface{}) ([]*Vulnerability, error) {
	// Query vulnerabilities from database based on filters
	vulns, err := sas.db.GetVulnerabilities(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerabilities: %w", err)
	}

	// Convert database models to service models
	result := make([]*Vulnerability, len(vulns))
	for i, vuln := range vulns {
		// Handle pointer fields
		var cvss float64
		if vuln.CVSS != nil {
			cvss = *vuln.CVSS
		}

		var cve string
		if vuln.CVE != nil {
			cve = *vuln.CVE
		}

		var location string
		if vuln.Location != nil {
			location = *vuln.Location
		}

		var remediation string
		if vuln.Remediation != nil {
			remediation = *vuln.Remediation
		}

		result[i] = &Vulnerability{
			ID:          vuln.ID.String(),
			Title:       vuln.Title,
			Description: vuln.Description,
			Severity:    vuln.Severity,
			CVSS:        cvss,
			CVE:         cve,
			Category:    vuln.Category,
			Component:   vuln.Component,
			Location:    location,
			Status:      vuln.Status,
			Remediation: remediation,
			ReferenceUrls: vuln.ReferenceUrls,
			Metadata:    vuln.Metadata,
			Discovered:  vuln.DiscoveredAt,
			Updated:     vuln.UpdatedAt,
		}
	}

	return result, nil
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (sas *SecurityAuditService) UpdateVulnerabilityStatus(ctx context.Context, vulnID, status, remediation string) error {
	// Convert string ID to UUID
	vulnUUID, err := uuid.Parse(vulnID)
	if err != nil {
		return fmt.Errorf("invalid vulnerability ID: %w", err)
	}

	// Get current vulnerability state for audit logging
	currentVuln, err := sas.db.GetVulnerabilityByID(vulnUUID)
	if err != nil {
		return fmt.Errorf("failed to get vulnerability: %w", err)
	}

	// Update vulnerability status in database
	updates := map[string]interface{}{
		"status": status,
	}
	if remediation != "" {
		updates["remediation"] = remediation
	}

	_, err = sas.db.UpdateVulnerability(vulnUUID, updates)
	if err != nil {
		return fmt.Errorf("failed to update vulnerability status: %w", err)
	}

	// Log the change as audit event
	event := &storage.AuditEvent{
		Action:       "security.vulnerability_update",
		ResourceType: "vulnerability",
		ResourceID:   &vulnUUID,
		BeforeState:  map[string]interface{}{"status": currentVuln.Status, "remediation": currentVuln.Remediation},
		AfterState:   map[string]interface{}{"status": status, "remediation": remediation},
		Success:      true,
	}

	return sas.db.LogAuditEvent(ctx, event)
}