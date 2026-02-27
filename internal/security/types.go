package security

import (
	"time"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // "critical", "high", "medium", "low", "info"
	CVSS        float64                `json:"cvss_score,omitempty"`
	CVE         string                 `json:"cve,omitempty"`
	Category    string                 `json:"category"` // "injection", "auth", "crypto", "config", "network"
	Component   string                 `json:"component"`
	Location    string                 `json:"location,omitempty"`
	Status      string                 `json:"status"` // "open", "fixed", "accepted", "false_positive"
	Remediation string                 `json:"remediation,omitempty"`
	ReferenceUrls []string               `json:"reference_urls,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Discovered  time.Time              `json:"discovered"`
	Updated     time.Time              `json:"updated"`
}

// SecurityScan represents a complete security scan
type SecurityScan struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"` // "penetration_test", "vulnerability_scan", "compliance_check"
	Status       string          `json:"status"` // "running", "completed", "failed"
	Target       string          `json:"target"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Duration     time.Duration   `json:"duration,omitempty"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Summary      ScanSummary     `json:"summary"`
	Config       ScanConfig      `json:"config"`
}

// ScanSummary provides a summary of scan results
type ScanSummary struct {
	TotalVulnerabilities int            `json:"total_vulnerabilities"`
	CriticalCount        int            `json:"critical_count"`
	HighCount            int            `json:"high_count"`
	MediumCount          int            `json:"medium_count"`
	LowCount             int            `json:"low_count"`
	InfoCount            int            `json:"info_count"`
	RiskScore            float64        `json:"risk_score"`
	Coverage             float64        `json:"coverage_percentage"`
	ComplianceScore      float64        `json:"compliance_score,omitempty"`
}

// ScanConfig defines scan configuration
type ScanConfig struct {
	IncludePorts    []int  `json:"include_ports,omitempty"`
	ExcludePaths    []string `json:"exclude_paths,omitempty"`
	AuthCredentials map[string]string `json:"auth_credentials,omitempty"`
	Timeout         time.Duration `json:"timeout"`
	MaxConcurrency  int `json:"max_concurrency"`
	Depth           int `json:"depth"` // for crawling/web scanning
}