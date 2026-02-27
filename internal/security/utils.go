package security

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// Utility functions

func (sas *SecurityAuditService) generateScanSummary(vulnerabilities []Vulnerability, coverage float64) ScanSummary {
	summary := ScanSummary{
		TotalVulnerabilities: len(vulnerabilities),
		Coverage:             coverage,
	}

	riskScore := 0.0

	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case "critical":
			summary.CriticalCount++
			riskScore += 10.0
		case "high":
			summary.HighCount++
			riskScore += 7.0
		case "medium":
			summary.MediumCount++
			riskScore += 4.0
		case "low":
			summary.LowCount++
			riskScore += 1.0
		case "info":
			summary.InfoCount++
			riskScore += 0.1
		}
	}

	// Normalize risk score (0-100)
	if summary.TotalVulnerabilities > 0 {
		summary.RiskScore = (riskScore / float64(summary.TotalVulnerabilities*10)) * 100
		if summary.RiskScore > 100 {
			summary.RiskScore = 100
		}
	}

	return summary
}

func (sas *SecurityAuditService) logScanResults(scan *SecurityScan) {
	sas.logger.WithFields(logrus.Fields{
		"scan_id":                scan.ID,
		"type":                   scan.Type,
		"vulnerabilities_found":  len(scan.Vulnerabilities),
		"critical_count":         scan.Summary.CriticalCount,
		"high_count":             scan.Summary.HighCount,
		"risk_score":             scan.Summary.RiskScore,
		"duration":               scan.Duration,
	}).Info("Security scan completed")
}

func (sas *SecurityAuditService) isPortOpen(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ID generation functions

func generateScanID() string {
	return fmt.Sprintf("scan_%d", time.Now().UnixNano())
}

func generateVulnID() string {
	return fmt.Sprintf("vuln_%d", time.Now().UnixNano())
}