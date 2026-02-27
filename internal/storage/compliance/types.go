package compliance

import (
	"time"

	"github.com/google/uuid"
)

// ComplianceFramework represents different compliance frameworks
type ComplianceFramework string

const (
	GDPR      ComplianceFramework = "gdpr"
	SOC2      ComplianceFramework = "soc2"
	ISO27001  ComplianceFramework = "iso27001"
	PCI_DSS   ComplianceFramework = "pci_dss"
	HIPAA     ComplianceFramework = "hipaa"
	CCPA      ComplianceFramework = "ccpa"
)

// ComplianceRequirement represents a compliance requirement
type ComplianceRequirement struct {
	ID          string              `json:"id"`
	Framework   ComplianceFramework `json:"framework"`
	Section     string              `json:"section"`
	Requirement string              `json:"requirement"`
	Description string              `json:"description"`
	Severity    string              `json:"severity"` // "critical", "high", "medium", "low"
	Status      string              `json:"status"`   // "compliant", "non_compliant", "not_applicable"
	LastChecked time.Time           `json:"last_checked"`
	Evidence    []string            `json:"evidence,omitempty"`
}

// ComplianceReport represents a compliance assessment report
type ComplianceReport struct {
	ID               string                 `json:"id"`
	Framework        ComplianceFramework    `json:"framework"`
	AssessmentDate   time.Time              `json:"assessment_date"`
	OverallScore     float64                `json:"overall_score"`
	Status           string                 `json:"status"` // "compliant", "conditional", "non_compliant"
	Requirements     []ComplianceRequirement `json:"requirements"`
	CriticalFindings []string               `json:"critical_findings"`
	Recommendations  []string               `json:"recommendations"`
	NextAssessment   time.Time              `json:"next_assessment"`
	Assessor         string                 `json:"assessor"`
}

// ComplianceAuditEvent represents a compliance-specific audit event
type ComplianceAuditEvent struct {
	Framework    ComplianceFramework     `json:"framework"`
	Section      string                  `json:"section"`
	RequirementID string                 `json:"requirement_id"`
	Action       string                  `json:"action"`
	Severity     string                  `json:"severity"`
	BeforeState  interface{}             `json:"before_state,omitempty"`
	AfterState   interface{}             `json:"after_state,omitempty"`
	Success      bool                    `json:"success"`
	UserID       *uuid.UUID              `json:"user_id,omitempty"`
	IPAddress    string                  `json:"ip_address,omitempty"`
	UserAgent    string                  `json:"user_agent,omitempty"`
	Timestamp    time.Time               `json:"timestamp"`
}