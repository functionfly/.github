package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// LogComplianceEvent logs a compliance-related audit event
func (acs *AuditComplianceService) LogComplianceEvent(ctx context.Context, event *ComplianceAuditEvent) error {
	// Convert to standard audit event
	auditEvent := &storage.AuditEvent{
		Action:       fmt.Sprintf("compliance.%s", event.Action),
		ResourceType: string(event.Framework),
		ResourceID:   nil, // We'll store the requirement ID in metadata
		BeforeState:  event.BeforeState,
		AfterState:   event.AfterState,
		Success:      event.Success,
	}

	// Add compliance-specific metadata
	if auditEvent.AfterState == nil {
		auditEvent.AfterState = make(map[string]interface{})
	}
	afterState := auditEvent.AfterState.(map[string]interface{})
	afterState["compliance_framework"] = event.Framework
	afterState["compliance_section"] = event.Section
	afterState["requirement_id"] = event.RequirementID
	afterState["severity"] = event.Severity

	return acs.db.LogAuditEvent(ctx, auditEvent)
}

// AssessCompliance performs a compliance assessment for a given framework
func (acs *AuditComplianceService) AssessCompliance(ctx context.Context, framework ComplianceFramework) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:             uuid.New().String(),
		Framework:      framework,
		AssessmentDate: time.Now(),
		Requirements:   []ComplianceRequirement{},
		CriticalFindings: []string{},
		Recommendations:  []string{},
		Assessor:       "automated_system",
	}

	var requirements []ComplianceRequirement
	var err error

	switch framework {
	case GDPR:
		requirements, err = acs.assessGDPRCompliance(ctx)
	case SOC2:
		requirements, err = acs.assessSOC2Compliance(ctx)
	case ISO27001:
		requirements, err = acs.assessISO27001Compliance(ctx)
	case PCI_DSS:
		requirements, err = acs.assessPCIDSSCompliance(ctx)
	case HIPAA:
		requirements, err = acs.assessHIPAACompliance(ctx)
	case CCPA:
		requirements, err = acs.assessCCPACompliance(ctx)
	default:
		return nil, fmt.Errorf("unsupported compliance framework: %s", framework)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to assess %s compliance: %w", framework, err)
	}

	report.Requirements = requirements

	// Calculate overall score
	totalScore := 0.0
	totalWeight := 0.0

	for _, req := range requirements {
		weight := acs.getRequirementWeight(req.Severity)
		totalWeight += weight

		if req.Status == "compliant" {
			totalScore += weight
		} else if req.Status == "conditional" {
			totalScore += weight * 0.5 // Partial credit for conditional compliance
		}

		if req.Severity == "critical" && req.Status != "compliant" {
			report.CriticalFindings = append(report.CriticalFindings, req.Requirement)
		}
	}

	if totalWeight > 0 {
		report.OverallScore = (totalScore / totalWeight) * 100
	}

	// Determine overall status
	if report.OverallScore >= 95 {
		report.Status = "compliant"
	} else if report.OverallScore >= 80 {
		report.Status = "conditional"
	} else {
		report.Status = "non_compliant"
	}

	// Generate recommendations
	report.Recommendations = acs.generateRecommendations(framework, requirements)

	// Set next assessment date (quarterly)
	report.NextAssessment = report.AssessmentDate.AddDate(0, 3, 0)

	// Log the compliance assessment
	event := &ComplianceAuditEvent{
		Framework:    framework,
		Action:       "assessment_completed",
		BeforeState:  nil,
		AfterState:   map[string]interface{}{
			"score":     report.OverallScore,
			"status":    report.Status,
			"findings":  len(report.CriticalFindings),
		},
		Success:   true,
		Timestamp: time.Now(),
	}

	if err := acs.LogComplianceEvent(ctx, event); err != nil {
		acs.logger.WithError(err).Warn("Failed to log compliance assessment")
	}

	return report, nil
}

// getRequirementWeight returns weight based on severity
func (acs *AuditComplianceService) getRequirementWeight(severity string) float64 {
	switch severity {
	case "critical":
		return 10.0
	case "high":
		return 7.0
	case "medium":
		return 4.0
	case "low":
		return 1.0
	default:
		return 1.0
	}
}

// generateRecommendations generates compliance recommendations
func (acs *AuditComplianceService) generateRecommendations(framework ComplianceFramework, requirements []ComplianceRequirement) []string {
	recommendations := []string{}

	for _, req := range requirements {
		if req.Status != "compliant" {
			switch framework {
			case GDPR:
				recommendations = append(recommendations, fmt.Sprintf("Implement %s: %s", req.Section, req.Requirement))
			case SOC2:
				recommendations = append(recommendations, fmt.Sprintf("Address SOC 2 %s requirement", req.Section))
			case ISO27001:
				recommendations = append(recommendations, fmt.Sprintf("Strengthen ISO 27001 %s controls", req.Section))
			case PCI_DSS:
				recommendations = append(recommendations, fmt.Sprintf("Enhance PCI DSS %s implementation", req.Section))
			}
		}
	}

	// Add general recommendations
	if framework == GDPR && !acs.db.IsEncryptionEnabled() {
		recommendations = append(recommendations, "Enable database encryption to comply with GDPR Article 32")
	}

	return recommendations
}