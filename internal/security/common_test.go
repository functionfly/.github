package security

import (
	"context"
	"testing"
	"time"
)

func TestScanDependencies(t *testing.T) {
	sas := &SecurityAuditService{}

	vulns, err := sas.scanDependencies(context.Background())
	if err != nil {
		t.Fatalf("scanDependencies failed: %v", err)
	}

	// Should not return nil, even if empty
	if vulns == nil {
		t.Error("scanDependencies returned nil, should return empty slice")
	}

	// Check that vulnerabilities have required fields
	for i, vuln := range vulns {
		if vuln.ID == "" {
			t.Errorf("vulnerability %d missing ID", i)
		}
		if vuln.Title == "" {
			t.Errorf("vulnerability %d missing Title", i)
		}
		if vuln.Discovered.IsZero() {
			t.Errorf("vulnerability %d missing Discovered timestamp", i)
		}
		if vuln.Updated.IsZero() {
			t.Errorf("vulnerability %d missing Updated timestamp", i)
		}
	}
}

func TestScanInfrastructure(t *testing.T) {
	sas := &SecurityAuditService{}

	config := ScanConfig{
		IncludePorts: []int{80, 443},
		Timeout:      5 * time.Second,
	}

	vulns, err := sas.scanInfrastructure(context.Background(), "localhost", config)
	if err != nil {
		t.Fatalf("scanInfrastructure failed: %v", err)
	}

	// Should not return nil, even if empty
	if vulns == nil {
		t.Error("scanInfrastructure returned nil, should return empty slice")
	}

	// Check that vulnerabilities have required fields
	for i, vuln := range vulns {
		if vuln.ID == "" {
			t.Errorf("vulnerability %d missing ID", i)
		}
		if vuln.Title == "" {
			t.Errorf("vulnerability %d missing Title", i)
		}
		if vuln.Discovered.IsZero() {
			t.Errorf("vulnerability %d missing Discovered timestamp", i)
		}
		if vuln.Updated.IsZero() {
			t.Errorf("vulnerability %d missing Updated timestamp", i)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.1.0", "v1.0.0", 1},
		{"v1.0.0", "v1.1.0", -1},
		{"v2.0.0", "v1.9.9", 1},
		{"v1.0.1", "v1.0.0", 1},
	}

	for _, test := range tests {
		result := compareVersions(test.v1, test.v2)
		if result != test.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", test.v1, test.v2, result, test.expected)
		}
	}
}

func TestScanContainers(t *testing.T) {
	sas := &SecurityAuditService{}

	vulns, err := sas.scanContainers(context.Background())
	if err != nil {
		t.Fatalf("scanContainers failed: %v", err)
	}

	// Should not return nil, even if empty
	if vulns == nil {
		t.Error("scanContainers returned nil, should return empty slice")
	}

	// Check that vulnerabilities have required fields
	for i, vuln := range vulns {
		if vuln.ID == "" {
			t.Errorf("vulnerability %d missing ID", i)
		}
		if vuln.Title == "" {
			t.Errorf("vulnerability %d missing Title", i)
		}
		if vuln.Discovered.IsZero() {
			t.Errorf("vulnerability %d missing Discovered timestamp", i)
		}
		if vuln.Updated.IsZero() {
			t.Errorf("vulnerability %d missing Updated timestamp", i)
		}
	}
}