package testing

import "strings"

type SecurityScanner struct {
	patterns map[string][]string
}

func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{patterns: map[string][]string{
		"critical": {"eval(", "exec(", "compile(", "os.system", "subprocess", "__import__(", "importlib", "child_process.exec", "new Function("},
		"network":  {"requests.", "urllib", "httpx", "fetch(", "axios", "net/http"},
		"resource": {"while true", "for (;;)", "malloc(", "bytearray(10**", "make([]byte, 1<<"},
	}}
}

func (s *SecurityScanner) Scan(code string) (bool, []string) {
	lower := strings.ToLower(code)
	findings := []string{}
	for severity, patterns := range s.patterns {
		for _, pattern := range patterns {
			if strings.Contains(lower, strings.ToLower(pattern)) {
				findings = append(findings, severity+":"+pattern)
			}
		}
	}
	return len(findings) == 0, findings
}
