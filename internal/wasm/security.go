//go:build cgo

package wasm

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
		},
	}
}

func (s *RuntimeScanner) ScanSource(sourceCode string) *RuntimeScanResult {
	result := &RuntimeScanResult{
		Blocked: false,
		Threats: []RuntimeThreat{},
	}

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