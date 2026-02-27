package security

import (
	"context"

	"github.com/functionfly/functionfly/internal/security/compliance"
)

// convertComplianceIssuesToVulnerabilities converts compliance issues to vulnerabilities
func convertComplianceIssuesToVulnerabilities(issues []compliance.ComplianceIssue) []Vulnerability {
	vulnerabilities := make([]Vulnerability, len(issues))
	for i, issue := range issues {
		vulnerabilities[i] = Vulnerability{
			ID:          issue.ID,
			Title:       issue.Title,
			Description: issue.Description,
			Severity:    issue.Severity,
			Category:    issue.Category,
			Component:   issue.Component,
			Status:      issue.Status,
			Remediation: issue.Remediation,
			Discovered:  issue.Discovered,
			Updated:     issue.Updated,
		}
	}
	return vulnerabilities
}

// Compliance checking methods

// checkSOC2Compliance performs SOC2 compliance checks
func (sas *SecurityAuditService) checkSOC2Compliance(ctx context.Context) []Vulnerability {
	checker := compliance.NewSOC2Checker(sas.db, sas.logger)
	issues := checker.CheckCompliance(ctx)
	return convertComplianceIssuesToVulnerabilities(issues)
}

// checkISO27001Compliance performs ISO 27001 compliance checks
func (sas *SecurityAuditService) checkISO27001Compliance(ctx context.Context) []Vulnerability {
	checker := compliance.NewISO27001Checker(sas.db, sas.logger)
	issues := checker.CheckCompliance(ctx)
	return convertComplianceIssuesToVulnerabilities(issues)
}

// checkGDPRCompliance performs GDPR compliance checks
func (sas *SecurityAuditService) checkGDPRCompliance(ctx context.Context) []Vulnerability {
	checker := compliance.NewGDPRChecker(sas.db, sas.logger)
	issues := checker.CheckCompliance(ctx)
	return convertComplianceIssuesToVulnerabilities(issues)
}

// checkPCIDSSCompliance performs PCI DSS compliance checks
func (sas *SecurityAuditService) checkPCIDSSCompliance(ctx context.Context) []Vulnerability {
	checker := compliance.NewPCIDSSChecker(sas.db, sas.logger)
	issues := checker.CheckCompliance(ctx)
	return convertComplianceIssuesToVulnerabilities(issues)
}