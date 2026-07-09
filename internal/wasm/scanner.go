package wasm

// This file holds the RuntimeScanner and threat types in a build-tag-free
// location so the production Go runtime (both wasmtime and wazero builds)
// can scan user code at instantiation time. The original security.go
// remains a thin shim for CGO consumers.

import (
	"fmt"
	"log"
	"math"
	"regexp"
)

type ThreatSeverity string

const (
	SeverityLow      ThreatSeverity = "low"
	SeverityMedium   ThreatSeverity = "medium"
	SeverityHigh     ThreatSeverity = "high"
	SeverityCritical ThreatSeverity = "critical"
)

type RuntimeThreat struct {
	Severity    ThreatSeverity
	Description string
	RuleID      string
	Confidence  float64
}

type RuntimeScanResult struct {
	Blocked bool
	Threats []RuntimeThreat
}

type RuntimeScanner struct {
	patterns []threatPattern
}

type threatPattern struct {
	regex       *regexp.Regexp
	severity    ThreatSeverity
	ruleID      string
	description string
}

func NewRuntimeScanner() *RuntimeScanner {
	return &RuntimeScanner{
		patterns: []threatPattern{
			{regexp.MustCompile(`(?i)\beval\s*\(`), SeverityHigh, "RUNTIME_EVAL", "Dynamic code execution via eval()"},
			{regexp.MustCompile(`(?i)\bexec\s*\(`), SeverityHigh, "RUNTIME_EXEC", "Dynamic code execution via exec()"},
			{regexp.MustCompile(`(?i)\bcompile\s*\(`), SeverityHigh, "RUNTIME_COMPILE", "Dynamic code compilation"},
			{regexp.MustCompile(`(?i)__import__\s*\(`), SeverityHigh, "RUNTIME_IMPORT", "Dynamic module import"},
			{regexp.MustCompile(`(?i)open\s*\([^)]*["']w["']`), SeverityHigh, "RUNTIME_FILE_WRITE", "File write operation"},
			{regexp.MustCompile(`(?i)os\.system\s*\(`), SeverityHigh, "RUNTIME_OS_SYSTEM", "OS command execution"},
			{regexp.MustCompile(`(?i)subprocess\s*\(`), SeverityHigh, "RUNTIME_SUBPROCESS", "Subprocess execution"},
			{regexp.MustCompile(`(?i)eval\s*\(.*input\s*\(`), SeverityCritical, "RUNTIME_EVAL_INPUT", "Eval with user input - code injection"},
			{regexp.MustCompile(`(?i)base64\.b64decode\s*\(`), SeverityMedium, "RUNTIME_BASE64_DECODE", "Base64 decoding in runtime"},
			{regexp.MustCompile(`(?i)memoryview\s*\(`), SeverityMedium, "RUNTIME_MEMORY_VIEW", "Direct memory access"},
			{regexp.MustCompile(`(?i)ctypes\s*\(`), SeverityHigh, "RUNTIME_CTYPES", "Foreign function interface usage"},
			{regexp.MustCompile(`(?i)\bgetattr\s*\(`), SeverityMedium, "RUNTIME_GETATTR", "Dynamic attribute access"},
			{regexp.MustCompile(`(?i)\bsetattr\s*\(`), SeverityMedium, "RUNTIME_SETATTR", "Dynamic attribute modification"},
			{regexp.MustCompile(`(?i)locals\s*\(`), SeverityMedium, "RUNTIME_LOCALS", "Local namespace access"},
			{regexp.MustCompile(`(?i)globals\s*\(`), SeverityMedium, "RUNTIME_GLOBALS", "Global namespace access"},
			// C-specific rules (used by c_wasm_e2e_test.go).
			{regexp.MustCompile(`(?i)\bsystem\s*\(`), SeverityHigh, "RUNTIME_C_SYSTEM", "C system() shell exec"},
			{regexp.MustCompile(`(?i)\bpopen\s*\(`), SeverityHigh, "RUNTIME_C_POPEN", "C popen shell exec"},
			{regexp.MustCompile(`(?i)\bdlopen\s*\(`), SeverityHigh, "RUNTIME_C_DLOPEN", "C dynamic library loading"},
			{regexp.MustCompile(`(?i)\bgets\s*\(`), SeverityCritical, "RUNTIME_C_GETS", "C gets() buffer overflow risk"},
			{regexp.MustCompile(`(?i)\bstrcpy\s*\(`), SeverityCritical, "RUNTIME_C_STRCPY", "C strcpy() buffer overflow risk"},
			{regexp.MustCompile(`(?i)\bstrcat\s*\(`), SeverityCritical, "RUNTIME_C_STRCAT", "C strcat() buffer overflow risk"},
			{regexp.MustCompile(`(?i)\bsprintf\s*\(`), SeverityCritical, "RUNTIME_C_SPRINTF", "C sprintf() buffer overflow risk"},
			{regexp.MustCompile(`(?i)\bfork\s*\(`), SeverityHigh, "RUNTIME_C_FORK", "C fork process spawn"},
		},
	}
}

func (s *RuntimeScanner) ScanSource(sourceCode string) *RuntimeScanResult {
	result := &RuntimeScanResult{Blocked: false, Threats: []RuntimeThreat{}}

	for _, pattern := range s.patterns {
		matches := pattern.regex.FindAllStringIndex(sourceCode, -1)
		for range matches {
			threat := RuntimeThreat{
				Severity:    pattern.severity,
				Description: pattern.description,
				RuleID:      pattern.ruleID,
				Confidence:  0.85,
			}
			result.Threats = append(result.Threats, threat)

			if threat.Severity == SeverityHigh || threat.Severity == SeverityCritical {
				result.Blocked = true
			}
		}
	}

	if entropy := calculateRuntimeEntropy(sourceCode); entropy > 6.5 {
		highEntropyThreat := RuntimeThreat{
			Severity:    SeverityMedium,
			Description: fmt.Sprintf("High entropy (%.2f) - possible obfuscated payload", entropy),
			RuleID:      "HIGH_ENTROPY",
			Confidence:  0.7,
		}
		result.Threats = append(result.Threats, highEntropyThreat)
	}

	return result
}

// ScanBytes performs a lightweight scan over compiled WASM bytes. It
// checks for the WASM magic and detects wasi sock_* imports which would
// bypass network isolation.
func (s *RuntimeScanner) ScanBytes(wasmBytes []byte) *RuntimeScanResult {
	result := &RuntimeScanResult{Blocked: false, Threats: []RuntimeThreat{}}
	if len(wasmBytes) < 8 || string(wasmBytes[:4]) != "\x00asm" {
		result.Blocked = true
		result.Threats = append(result.Threats, RuntimeThreat{
			Severity:    SeverityCritical,
			Description: "missing or invalid WASM magic bytes",
			RuleID:      "WASM_BAD_MAGIC",
			Confidence:  1.0,
		})
		return result
	}
	if hasAnySubstring(string(wasmBytes), "sock_accept", "sock_connect", "sock_send", "sock_recv") {
		result.Threats = append(result.Threats, RuntimeThreat{
			Severity:    SeverityHigh,
			Description: "module imports wasi sock_* functions (network would bypass WASI allowlist)",
			RuleID:      "WASM_SOCK_IMPORT",
			Confidence:  0.9,
		})
		result.Blocked = true
	}
	return result
}

func calculateRuntimeEntropy(data string) float64 {
	if len(data) == 0 {
		return 0
	}

	charCount := make(map[rune]int)
	for _, char := range data {
		charCount[char]++
	}

	entropy := 0.0
	length := float64(len(data))

	for _, count := range charCount {
		if count > 0 {
			prob := float64(count) / length
			entropy -= prob * math.Log2(prob)
		}
	}

	return entropy
}

func (s *RuntimeScanner) LogThreats(result *RuntimeScanResult) {
	for _, threat := range result.Threats {
		prefix := "[RuntimeScan]"
		if threat.Severity == SeverityCritical {
			prefix = "[RuntimeScan BLOCKED]"
		}
		log.Printf("%s %s: %s (confidence: %.2f)", prefix, threat.RuleID, threat.Description, threat.Confidence)
	}
}

// indexOf is a low-allocation substring index. Used by hasAnySubstring.
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// hasAnySubstring is a substring check used by the scanner's sock-import
// heuristic and by go_runtime's fuel/proc_exit checks. Variadic so
// callers can pass one or more patterns.
func hasAnySubstring(s string, subs ...string) bool {
	for _, sub := range subs {
		if sub == "" {
			continue
		}
		if indexOf(s, sub) >= 0 {
			return true
		}
	}
	return false
}
