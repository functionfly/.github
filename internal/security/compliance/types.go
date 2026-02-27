package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// generateVulnID generates a unique vulnerability ID
func generateVulnID() string {
	return fmt.Sprintf("vuln_%d", time.Now().UnixNano())
}

// ComplianceCheck represents a single compliance check
type ComplianceCheck struct {
	Title       string
	Description string
	Severity    string
	CheckFunc   func() bool
}

// ComplianceIssue represents a compliance issue found during checking
type ComplianceIssue struct {
	ID          string
	Title       string
	Description string
	Severity    string
	Category    string
	Component   string
	Status      string
	Remediation string
	Discovered  time.Time
	Updated     time.Time
}

// ComplianceChecker interface for different compliance frameworks
type ComplianceChecker interface {
	CheckCompliance(ctx context.Context) []ComplianceIssue
}

// Note: Vulnerability type is defined in the parent security package

// ComplianceManager orchestrates all compliance checkers
type ComplianceManager struct {
	checkers map[string]ComplianceChecker
}

// NewComplianceManager creates a new compliance manager
func NewComplianceManager(db storage.Repository, logger *logrus.Logger) *ComplianceManager {
	return &ComplianceManager{
		checkers: map[string]ComplianceChecker{
			"soc2":     NewSOC2Checker(db, logger),
			"iso27001": NewISO27001Checker(db, logger),
			"gdpr":     NewGDPRChecker(db, logger),
			"pci-dss":  NewPCIDSSChecker(db, logger),
		},
	}
}

// CheckCompliance runs compliance checks for the specified framework
func (cm *ComplianceManager) CheckCompliance(ctx context.Context, framework string) ([]ComplianceIssue, error) {
	checker, exists := cm.checkers[framework]
	if !exists {
		return nil, fmt.Errorf("unsupported compliance framework: %s", framework)
	}

	return checker.CheckCompliance(ctx), nil
}

// GetSupportedFrameworks returns a list of supported compliance frameworks
func (cm *ComplianceManager) GetSupportedFrameworks() []string {
	frameworks := make([]string, 0, len(cm.checkers))
	for framework := range cm.checkers {
		frameworks = append(frameworks, framework)
	}
	return frameworks
}