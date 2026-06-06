package security

import (
	"context"
	"testing"
)

func TestPluginAnalyzer_AnalyzeManifest_NilManifest(t *testing.T) {
	a := NewPluginAnalyzer()
	result, err := a.AnalyzeManifest(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Safe {
		t.Error("nil manifest should be marked unsafe")
	}
	if result.Score != 0 {
		t.Errorf("nil manifest score should be 0, got %.1f", result.Score)
	}
	if len(result.Issues) == 0 {
		t.Error("nil manifest should produce at least one issue")
	}
	if result.Issues[0].Severity != "critical" {
		t.Errorf("nil manifest issue should be critical, got %q", result.Issues[0].Severity)
	}
}

func TestPluginAnalyzer_AnalyzeManifest_CleanPlugin(t *testing.T) {
	a := NewPluginAnalyzer()
	manifest := &PluginManifest{
		ID:      "test-plugin",
		Name:    "test-plugin",
		Version: "1.0.0",
		Permissions: map[string]interface{}{
			"read": true,
		},
		NetworkHosts:    []string{"api.example.com"},
		FilesystemScope: "workspace",
	}
	result, err := a.AnalyzeManifest(context.Background(), manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Safe {
		t.Errorf("clean plugin should be safe, got %v", result.Issues)
	}
	if result.Score < 90 {
		t.Errorf("clean plugin should score >= 90, got %.1f", result.Score)
	}
}

func TestPluginAnalyzer_AnalyzeManifest_DangerousNetworkHosts(t *testing.T) {
	a := NewPluginAnalyzer()
	manifest := &PluginManifest{
		Name: "data-exfil",
		NetworkHosts: []string{
			"pastebin.com",
			"transfer.sh",
		},
	}
	result, err := a.AnalyzeManifest(context.Background(), manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NetworkAnalysis.DataExfilRisk {
		t.Error("expected DataExfilRisk=true for known exfil hosts")
	}
	if len(result.NetworkAnalysis.DangerousHosts) != 2 {
		t.Errorf("expected 2 dangerous hosts, got %d", len(result.NetworkAnalysis.DangerousHosts))
	}
}

func TestPluginAnalyzer_AnalyzeManifest_InternalNetworkAccess(t *testing.T) {
	tests := []struct {
		host     string
		internal bool
	}{
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"localhost", true},
		{"api.example.com", false},
		{"8.8.8.8", false},
	}
	a := NewPluginAnalyzer()
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			manifest := &PluginManifest{Name: "net-test", NetworkHosts: []string{tt.host}}
			result, _ := a.AnalyzeManifest(context.Background(), manifest)
			if result.NetworkAnalysis.InternalAccess != tt.internal {
				t.Errorf("host %q: expected InternalAccess=%v, got %v",
					tt.host, tt.internal, result.NetworkAnalysis.InternalAccess)
			}
		})
	}
}

func TestPluginAnalyzer_AnalyzeManifest_PermissionRiskLevels(t *testing.T) {
	tests := []struct {
		perm     string
		expected string
	}{
		{"terminal", "high"},
		{"gpu", "high"},
		{"api_keys", "high"},
		{"secrets", "high"},
		{"network", "medium"},
		{"filesystem", "medium"},
		{"agents", "low"},
	}
	a := NewPluginAnalyzer()
	for _, tt := range tests {
		t.Run(tt.perm, func(t *testing.T) {
			manifest := &PluginManifest{
				Name:        "perm-test",
				Permissions: map[string]interface{}{tt.perm: true},
			}
			result, _ := a.AnalyzeManifest(context.Background(), manifest)
			found := false
			for _, risk := range result.PermissionRisks {
				if risk.Permission == tt.perm {
					found = true
					if risk.RiskLevel != tt.expected {
						t.Errorf("permission %q: expected risk %q, got %q",
							tt.perm, tt.expected, risk.RiskLevel)
					}
				}
			}
			if !found {
				t.Errorf("permission %q not found in risk analysis", tt.perm)
			}
		})
	}
}

func TestPluginAnalyzer_AnalyzeManifest_HighRiskPermissionsLowerScore(t *testing.T) {
	a := NewPluginAnalyzer()
	clean := &PluginManifest{
		Name:        "clean",
		Permissions: map[string]interface{}{"read": true},
	}
	risky := &PluginManifest{
		Name:        "risky",
		Permissions: map[string]interface{}{"terminal": true, "secrets": true, "api_keys": true},
	}
	cleanResult, _ := a.AnalyzeManifest(context.Background(), clean)
	riskyResult, _ := a.AnalyzeManifest(context.Background(), risky)
	if riskyResult.Score >= cleanResult.Score {
		t.Errorf("risky plugin (%.1f) should score lower than clean (%.1f)",
			riskyResult.Score, cleanResult.Score)
	}
}

func TestPluginAnalyzer_AnalyzeManifest_FilesystemScopes(t *testing.T) {
	tests := []struct {
		scope        string
		readOps      bool
		writeOps     bool
		pathTravel   bool
	}{
		{"full", true, true, true},
		{"workspace", true, true, false},
		{"read-only", true, false, false},
		{"none", false, false, false},
		{"unknown-scope", true, true, true},
	}
	a := NewPluginAnalyzer()
	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			manifest := &PluginManifest{Name: "fs-test", FilesystemScope: tt.scope}
			result, _ := a.AnalyzeManifest(context.Background(), manifest)
			fs := result.FilesystemAnalysis
			if fs.ReadOperations != tt.readOps {
				t.Errorf("scope %q: readOps expected %v, got %v", tt.scope, tt.readOps, fs.ReadOperations)
			}
			if fs.WriteOperations != tt.writeOps {
				t.Errorf("scope %q: writeOps expected %v, got %v", tt.scope, tt.writeOps, fs.WriteOperations)
			}
			if fs.PathTraversal != tt.pathTravel {
				t.Errorf("scope %q: pathTraversal expected %v, got %v", tt.scope, tt.pathTravel, fs.PathTraversal)
			}
		})
	}
}

func TestPluginAnalyzer_AnalyzeManifest_ScoreCannotGoNegative(t *testing.T) {
	a := NewPluginAnalyzer()
	manifest := &PluginManifest{
		Name: "everything-risky",
		Permissions: map[string]interface{}{
			"terminal":   true,
			"gpu":        true,
			"api_keys":   true,
			"secrets":    true,
			"network":    true,
			"filesystem": true,
		},
		NetworkHosts: []string{"pastebin.com", "10.0.0.1", "127.0.0.1", "192.168.1.1"},
	}
	result, _ := a.AnalyzeManifest(context.Background(), manifest)
	if result.Score < 0 {
		t.Errorf("score should be clamped to 0, got %.1f", result.Score)
	}
	if result.Safe {
		t.Error("extremely risky plugin should not be marked safe")
	}
}

func TestPluginAnalyzer_AnalyzeCode_CryptoMinerDetection(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"stratum url", "const url = 'stratum+tcp://1.2.3.4:3333'", true},
		{"cryptonight", "import { cryptonight } from 'crypto'", true},
		{"mining keyword", "function startMining() {}", true},
		{"wallet pattern", "var wallet = '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa'", true},
		{"clean code", "function add(a, b) { return a + b }", false},
		{"empty", "", false},
	}
	a := NewPluginAnalyzer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := a.AnalyzeCode(context.Background(), tt.code)
			hasCrypto := false
			for _, issue := range result.Issues {
				if issue.Type == IssueTypeCryptoMiner {
					hasCrypto = true
					break
				}
			}
			if hasCrypto != tt.expected {
				t.Errorf("code %q: expected crypto detected=%v, got %v", tt.name, tt.expected, hasCrypto)
			}
			if hasCrypto && result.Safe {
				t.Error("plugin with crypto miner should not be safe")
			}
		})
	}
}

func TestPluginAnalyzer_AnalyzeCode_SecretDetection(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"openai key", "const key = 'sk-' + 'abcdefghijklmnopqrstuvwxyz1234567890'", true},
		{"github pat", "const token = 'ghp_' + 'abcdefghijklmnopqrstuvwxyz1234567890abcd'", true},
		{"aws key", "aws_access_key_id = 'AKIAIOSFODNN7EXAMPLE'", false},
		{"api key pattern", `const apiKey = "myApiKeyabcdef1234567890abcdef1234"`, true},
		{"password assignment", `password = "supersecret123"`, true},
		{"clean code", `const x = 42`, false},
		{"empty", "", false},
	}
	a := NewPluginAnalyzer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := a.AnalyzeCode(context.Background(), tt.code)
			hasSecret := false
			for _, issue := range result.Issues {
				if issue.Type == IssueTypeSecret {
					hasSecret = true
					break
				}
			}
			if hasSecret != tt.expected {
				t.Errorf("code %q: expected secret detected=%v, got %v", tt.name, tt.expected, hasSecret)
			}
		})
	}
}

func TestPluginAnalyzer_AnalyzeCode_ScoreClamping(t *testing.T) {
	a := NewPluginAnalyzer()
	dirty := `const key = "sk-abcdefghijklmnopqrstuvwxyz1234567890";
const token = "ghp_abcdefghijklmnopqrstuvwxyz1234567890abcd";
function startMining() {}`
	result, _ := a.AnalyzeCode(context.Background(), dirty)
	if result.Score < 0 {
		t.Errorf("score should be clamped to 0, got %.1f", result.Score)
	}
}

func TestPluginAnalyzer_AnalyzeCode_CryptoMinerIsCritical(t *testing.T) {
	a := NewPluginAnalyzer()
	code := "const url = 'stratum+tcp://1.2.3.4:3333'"
	result, _ := a.AnalyzeCode(context.Background(), code)
	if len(result.Issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	if result.Issues[0].Severity != "critical" {
		t.Errorf("crypto miner should be critical severity, got %q", result.Issues[0].Severity)
	}
	if result.Issues[0].Type != IssueTypeCryptoMiner {
		t.Errorf("expected crypto miner type, got %q", result.Issues[0].Type)
	}
}

func TestPluginAnalyzer_AnalyzeManifest_DataExfilWarning(t *testing.T) {
	a := NewPluginAnalyzer()
	manifest := &PluginManifest{
		Name:         "exfil",
		NetworkHosts: []string{"pastebin.com"},
	}
	result, _ := a.AnalyzeManifest(context.Background(), manifest)
	found := false
	for _, w := range result.Warnings {
		if contains(w, "exfiltrate") {
			found = true
		}
	}
	if !found {
		t.Error("expected data exfiltration warning in result")
	}
}

func TestPluginAnalyzer_AnalyzeManifest_SafeThreshold(t *testing.T) {
	a := NewPluginAnalyzer()

	highRisk := &PluginManifest{
		Name: "all-bad",
		Permissions: map[string]interface{}{
			"terminal": true,
			"secrets":  true,
			"api_keys": true,
			"gpu":      true,
		},
		NetworkHosts:    []string{"pastebin.com"},
		FilesystemScope: "full",
	}
	result, _ := a.AnalyzeManifest(context.Background(), highRisk)
	if result.Safe {
		t.Errorf("high-risk plugin (score %.1f) should not be safe", result.Score)
	}
}

func TestGetRecommendation(t *testing.T) {
	tests := []struct {
		perm  string
		empty bool
	}{
		{"terminal", false},
		{"gpu", false},
		{"api_keys", false},
		{"secrets", false},
		{"filesystem", false},
		{"agents", false},
		{"unknown-perm", false},
	}
	for _, tt := range tests {
		t.Run(tt.perm, func(t *testing.T) {
			rec := getRecommendation(tt.perm)
			if tt.empty && rec != "" {
				t.Errorf("expected empty recommendation for %q, got %q", tt.perm, rec)
			}
			if !tt.empty && rec == "" {
				t.Errorf("expected non-empty recommendation for %q", tt.perm)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
