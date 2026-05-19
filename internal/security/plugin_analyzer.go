package security

import (
	"context"
	"regexp"
	"strings"
)

type PluginManifest struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Permissions   map[string]interface{} `json:"permissions"`
	Entrypoints   map[string]string      `json:"entrypoints"`
	Capabilities  []string               `json:"capabilities"`
	SandboxTier   string                `json:"sandbox_tier"`
	NetworkHosts  []string               `json:"network_hosts"`
	FilesystemScope string               `json:"filesystem_scope"`
}

type PluginSecurityIssue struct {
	ID          string `json:"id"`
	Type        IssueType `json:"type"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Remediation string `json:"remediation"`
}

type IssueType string

const (
	IssueTypeMalware       IssueType = "malware"
	IssueTypeSecret        IssueType = "secret"
	IssueTypeCryptoMiner   IssueType = "crypto_miner"
	IssueTypeSuspiciousPerm IssueType = "suspicious_permission"
	IssueTypeNetworkRisk   IssueType = "network_risk"
	IssueTypeFilesystemRisk IssueType = "filesystem_risk"
)

type PluginAnalysisResult struct {
	Safe               bool                   `json:"safe"`
	Score              float64                `json:"score"`
	Issues             []PluginSecurityIssue  `json:"issues"`
	PermissionRisks    []PermissionRisk       `json:"permission_risks"`
	NetworkAnalysis    NetworkAnalysis        `json:"network_analysis"`
	FilesystemAnalysis FilesystemAnalysis     `json:"filesystem_analysis"`
	Warnings           []string               `json:"warnings"`
}

type PermissionRisk struct {
	Permission   string  `json:"permission"`
	RiskLevel    string  `json:"risk_level"`
	Description  string  `json:"description"`
	Recommendation string `json:"recommendation"`
}

type NetworkAnalysis struct {
	AllowedHosts    []string `json:"allowed_hosts"`
	DangerousHosts  []string `json:"dangerous_hosts"`
	InternalAccess  bool     `json:"internal_access"`
	DataExfilRisk   bool     `json:"data_exfiltration_risk"`
}

type FilesystemAnalysis struct {
	Scope           string   `json:"scope"`
	ReadOperations  bool     `json:"read_operations"`
	WriteOperations bool     `json:"write_operations"`
	PathTraversal   bool     `json:"path_traversal_risk"`
}

var (
	cryptoPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(stratum|tcp://|http://).*\d+\.\d+\.\d+\.\d+:\d+`),
		regexp.MustCompile(`(?i)cryptonight|equihash|ethash|keccak|scrypt|lyra2`),
		regexp.MustCompile(`(?i)wallet.*[0-9a-zA-Z]{20,}`),
		regexp.MustCompile(`(?i)mining|miner|hashrate|nethash`),
	}

	pluginSecretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|token|auth).*['"][0-9a-zA-Z]{16,}['"]`),
		regexp.MustCompile(`(?i)password\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)(aws[_-]?access|aws[_-]?secret).*['"][0-9a-zA-Z]{16,}['"]`),
		regexp.MustCompile(`sk-[0-9a-zA-Z]{20,}`),
		regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`),
		regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z]{10,}`),
	}

	networkRiskDomains = []string{
		"pastebin.com",
		"transfer.sh",
		"ipinfo.io",
		"ip-api.com",
		"check.torproject.org",
	}

	suspiciousPermissions = map[string]string{
		"terminal":     "Allows arbitrary command execution",
		"gpu":          "Allows GPU access for compute workloads",
		"agents":       "Allows creating/managing AI agents",
		"api_keys":     "Allows reading API keys",
		"secrets":      "Allows access to secrets vault",
		"filesystem":   "full filesystem access",
	}
)

type PluginAnalyzer struct {
	enabled bool
}

func NewPluginAnalyzer() *PluginAnalyzer {
	return &PluginAnalyzer{enabled: true}
}

func (a *PluginAnalyzer) AnalyzeManifest(ctx context.Context, manifest *PluginManifest) (*PluginAnalysisResult, error) {
	result := &PluginAnalysisResult{
		Safe:            true,
		Score:           100.0,
		Issues:          []PluginSecurityIssue{},
		PermissionRisks: []PermissionRisk{},
		Warnings:       []string{},
	}

	if manifest == nil {
		result.Safe = false
		result.Score = 0
		result.Issues = append(result.Issues, PluginSecurityIssue{
			ID:          "INVALID_MANIFEST",
			Type:        IssueTypeMalware,
			Severity:    "critical",
			Title:       "Invalid or missing manifest",
			Description: "Plugin manifest is required and must be valid JSON",
			Remediation: "Provide a valid plugin manifest with required fields",
		})
		return result, nil
	}

	result.PermissionRisks = a.analyzePermissions(manifest.Permissions)
	result.NetworkAnalysis = a.analyzeNetwork(manifest.NetworkHosts)
	result.FilesystemAnalysis = a.analyzeFilesystem(manifest.FilesystemScope)

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			result.Score -= 50
		case "high":
			result.Score -= 25
		case "medium":
			result.Score -= 10
		case "low":
			result.Score -= 5
		}
		if result.Score < 0 {
			result.Score = 0
		}
	}

	for _, permRisk := range result.PermissionRisks {
		switch permRisk.RiskLevel {
		case "high":
			result.Score -= 15
		case "medium":
			result.Score -= 8
		case "low":
			result.Score -= 3
		}
	}

	if result.NetworkAnalysis.DataExfilRisk {
		result.Score -= 20
		result.Warnings = append(result.Warnings, "Plugin may be able to exfiltrate data to external servers")
	}

	if result.Score < 30 {
		result.Safe = false
	}

	return result, nil
}

func (a *PluginAnalyzer) AnalyzeCode(ctx context.Context, code string) (*PluginAnalysisResult, error) {
	result := &PluginAnalysisResult{
		Safe:    true,
		Score:   100.0,
		Issues:  []PluginSecurityIssue{},
		Warnings: []string{},
	}

	if code == "" {
		return result, nil
	}

	result.Issues = append(result.Issues, a.detectCryptoMiners(code)...)
	result.Issues = append(result.Issues, a.detectSecrets(code)...)

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			result.Score -= 40
		case "high":
			result.Score -= 20
		case "medium":
			result.Score -= 10
		case "low":
			result.Score -= 5
		}
		if result.Score < 0 {
			result.Score = 0
		}
	}

	if result.Score < 30 {
		result.Safe = false
	}

	return result, nil
}

func (a *PluginAnalyzer) analyzePermissions(permissions map[string]interface{}) []PermissionRisk {
	risks := []PermissionRisk{}

	for perm := range permissions {
		if desc, ok := suspiciousPermissions[perm]; ok {
			riskLevel := "low"
			if perm == "terminal" || perm == "gpu" || perm == "api_keys" || perm == "secrets" {
				riskLevel = "high"
			} else if perm == "network" || perm == "filesystem" {
				riskLevel = "medium"
			}

			risks = append(risks, PermissionRisk{
				Permission:    perm,
				RiskLevel:     riskLevel,
				Description:   desc,
				Recommendation: getRecommendation(perm),
			})
		}
	}

	return risks
}

func (a *PluginAnalyzer) analyzeNetwork(hosts []string) NetworkAnalysis {
	analysis := NetworkAnalysis{
		AllowedHosts:   []string{},
		DangerousHosts: []string{},
	}

	for _, host := range hosts {
		isDangerous := false
		hostLower := strings.ToLower(host)

		for _, dangerous := range networkRiskDomains {
			if strings.Contains(hostLower, dangerous) {
				isDangerous = true
				break
			}
		}

		if isDangerous {
			analysis.DangerousHosts = append(analysis.DangerousHosts, host)
			analysis.DataExfilRisk = true
		} else {
			analysis.AllowedHosts = append(analysis.AllowedHosts, host)
		}

		if strings.HasPrefix(host, "10.") ||
			strings.HasPrefix(host, "172.16.") || strings.HasPrefix(host, "172.31.") ||
			strings.HasPrefix(host, "192.168.") ||
			strings.HasPrefix(host, "127.") ||
			strings.HasPrefix(host, "localhost") {
			analysis.InternalAccess = true
		}
	}

	return analysis
}

func (a *PluginAnalyzer) analyzeFilesystem(scope string) FilesystemAnalysis {
	analysis := FilesystemAnalysis{
		Scope: scope,
	}

	switch scope {
	case "full":
		analysis.ReadOperations = true
		analysis.WriteOperations = true
		analysis.PathTraversal = true
	case "workspace":
		analysis.ReadOperations = true
		analysis.WriteOperations = true
		analysis.PathTraversal = false
	case "read-only":
		analysis.ReadOperations = true
		analysis.WriteOperations = false
	case "none":
		analysis.ReadOperations = false
		analysis.WriteOperations = false
	default:
		analysis.ReadOperations = true
		analysis.WriteOperations = true
		analysis.PathTraversal = true
	}

	return analysis
}

func (a *PluginAnalyzer) detectCryptoMiners(code string) []PluginSecurityIssue {
	issues := []PluginSecurityIssue{}

	for i, pattern := range cryptoPatterns {
		if pattern.MatchString(code) {
			issues = append(issues, PluginSecurityIssue{
				ID:          "CRYPTO_MINER",
				Type:        IssueTypeCryptoMiner,
				Severity:    "critical",
				Title:       "Potential cryptocurrency miner detected",
				Description: "Code pattern matches known cryptocurrency mining behavior",
				Location:    "pattern_match",
				Remediation: "Remove any cryptocurrency mining code from the plugin",
			})
			_ = i
			break
		}
	}

	return issues
}

func (a *PluginAnalyzer) detectSecrets(code string) []PluginSecurityIssue {
	issues := []PluginSecurityIssue{}

	for _, pattern := range pluginSecretPatterns {
		if pattern.MatchString(code) {
			issues = append(issues, PluginSecurityIssue{
				ID:          "EXPOSED_SECRET",
				Type:        IssueTypeSecret,
				Severity:    "high",
				Title:       "Potential exposed secret detected",
				Description: "Code contains a pattern that may be an API key, token, or password",
				Location:    "pattern_match",
				Remediation: "Remove hardcoded secrets and use environment variables or the secrets vault instead",
			})
			break
		}
	}

	return issues
}

func getRecommendation(permission string) string {
	switch permission {
	case "terminal":
		return "Consider using a sandboxed execution environment instead of direct terminal access"
	case "gpu":
		return "GPU access should only be requested if the plugin genuinely needs ML/compute workloads"
	case "api_keys":
		return "Use the secrets vault to store API keys rather than direct access"
	case "secrets":
		return "Ensure the plugin really needs vault access and follows the principle of least privilege"
	case "filesystem":
		return "Request specific directories rather than broad filesystem access"
	case "agents":
		return "Clearly document why the plugin needs to create/manages agents"
	default:
		return "Review if this permission is necessary for the plugin's functionality"
	}
}