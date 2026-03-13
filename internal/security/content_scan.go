package security

import (
	"regexp"
	"strings"
)

// MessageScanResult holds the result of scanning message content for secrets and dangerous hints.
type MessageScanResult struct {
	Valid          bool     // false if suspected secrets are present
	Warnings       []string // non-blocking warnings (length, dangerous phrases)
	SecretsFound   []string // types of suspected secrets (e.g. "AWS Access Key")
	DangerousHints []string // matched dangerous capability phrases
}

// secretPattern describes a named regex pattern for detecting secrets in text.
var secretPatterns = []struct {
	name  string
	regex string
}{
	{"AWS Access Key", `AKIA[0-9A-Z]{16}`},
	{"JWT Token", `eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`},
	{"Private Key (PEM)", `-----BEGIN [A-Z\s]+PRIVATE KEY-----`},
	{"Stripe Secret Key", `sk_(test|live)_[A-Za-z0-9]{24,}`},
	{"GitHub Token", `ghp_[A-Za-z0-9]{36}`},
	{"GitHub OAuth/App", `gho_[A-Za-z0-9]{36}`},
	{"Slack Token", `xox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[A-Za-z0-9-]+`},
	{"Generic API Key (Bearer)", `(?i)bearer\s+[A-Za-z0-9\-_.]{20,}`},
	{"Database URL with password", `(postgres|mysql|mongodb|redis)://[^:\s]+:[^@\s]+@`},
}

// dangerousCapabilityPhrases are substrings that may indicate unsafe execution or capability requests.
var dangerousCapabilityPhrases = []string{
	"disable sandbox", "disable_sandbox", "run as root", "runas root",
	"execute shell", "exec shell", "system shell", "eval(",
	"bypass verification", "skip verification", "ignore ssl",
	"load arbitrary code", "arbitrary code execution", "remote code execution",
	"drop database", "delete from users", "truncate table",
	"curl | bash", "wget -O- | sh", "pipe to bash",
}

var (
	compiledSecretRegexes []struct {
		name  string
		regex *regexp.Regexp
	}
)

func init() {
	for _, p := range secretPatterns {
		re, err := regexp.Compile(p.regex)
		if err != nil {
			continue
		}
		compiledSecretRegexes = append(compiledSecretRegexes, struct {
			name  string
			regex *regexp.Regexp
		}{p.name, re})
	}
}

// ScanMessageContent scans user-provided message content for suspected secrets and dangerous capability hints.
// It is intended for conversation/DM message validation before sending.
func ScanMessageContent(content string) MessageScanResult {
	var result MessageScanResult
	result.Valid = true

	contentLower := strings.ToLower(content)

	// Scan for secret-like patterns
	for _, p := range compiledSecretRegexes {
		if p.regex.FindString(content) != "" {
			result.SecretsFound = append(result.SecretsFound, p.name)
			result.Valid = false
		}
	}

	// Scan for dangerous capability phrases (warning only; do not block by default)
	for _, phrase := range dangerousCapabilityPhrases {
		if strings.Contains(contentLower, phrase) {
			result.DangerousHints = append(result.DangerousHints, phrase)
		}
	}
	if len(result.DangerousHints) > 0 {
		result.Warnings = append(result.Warnings,
			"Do not request execution of unsafe commands or capability bypasses in this conversation.")
	}

	return result
}
