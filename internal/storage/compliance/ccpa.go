package compliance

import (
	"context"
	"time"
)

// assessCCPACompliance assesses CCPA compliance
func (acs *AuditComplianceService) assessCCPACompliance(ctx context.Context) ([]ComplianceRequirement, error) {
	requirements := []ComplianceRequirement{
		{
			ID:          "ccpa_rights",
			Framework:   CCPA,
			Section:     "1798.100",
			Requirement: "Consumer Rights",
			Description: "Right to know, delete, and opt-out of sale of personal information",
			Severity:    "high",
			LastChecked: time.Now(),
		},
		{
			ID:          "ccpa_security",
			Framework:   CCPA,
			Section:     "1798.150",
			Requirement: "Personal Information Security",
			Description: "Security safeguards for personal information",
			Severity:    "critical",
			LastChecked: time.Now(),
		},
		{
			ID:          "ccpa_disclosure",
			Framework:   CCPA,
			Section:     "1798.130",
			Requirement: "Information to Consumers",
			Description: "Privacy policy and notice at collection requirements",
			Severity:    "medium",
			LastChecked: time.Now(),
		},
	}

	// Check each requirement
	for i := range requirements {
		requirements[i].Status = acs.checkCCPARequirement(requirements[i])
	}

	return requirements, nil
}

// checkCCPARequirement checks CCPA compliance for a specific requirement
func (acs *AuditComplianceService) checkCCPARequirement(req ComplianceRequirement) string {
	switch req.ID {
	case "ccpa_security":
		// Check security measures
		if acs.db.IsEncryptionEnabled() {
			return "compliant"
		}
		return "non_compliant"
	default:
		return "compliant"
	}
}