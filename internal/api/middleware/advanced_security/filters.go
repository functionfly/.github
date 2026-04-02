package advanced_security

import (
	"net/http"
	"regexp"
)

// SQLInjectionFilter detects SQL injection attempts
type SQLInjectionFilter struct {
	patterns []*regexp.Regexp
}

// XSSFilter detects XSS attempts
type XSSFilter struct {
	patterns []*regexp.Regexp
}

// PathTraversalFilter detects path traversal attempts
type PathTraversalFilter struct {
	patterns []*regexp.Regexp
}

// Filter implementations
func (sqlf *SQLInjectionFilter) Detect(r *http.Request) bool {
	for _, pattern := range sqlf.patterns {
		if pattern.MatchString(r.URL.RawQuery) {
			return true
		}
	}
	return false
}

func (xssf *XSSFilter) Detect(r *http.Request) bool {
	for _, pattern := range xssf.patterns {
		if pattern.MatchString(r.URL.RawQuery) {
			return true
		}
	}
	return false
}

func (ptf *PathTraversalFilter) Detect(r *http.Request) bool {
	for _, pattern := range ptf.patterns {
		if pattern.MatchString(r.URL.Path) {
			return true
		}
	}
	return false
}

// sqlInjectionQueryPatterns returns regex patterns for detecting SQL injection in query strings.
func sqlInjectionQueryPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bUNION\b.*\bSELECT\b`),
		regexp.MustCompile(`(?i)\bDELETE\b.*\bFROM\b`),
		regexp.MustCompile(`(?i)\bINSERT\b.*\bINTO\b`),
		regexp.MustCompile(`(?i)\bUPDATE\b.*\bSET\b`),
		regexp.MustCompile(`(?i)\bDROP\b.*\bTABLE\b`),
		regexp.MustCompile(`(?i)\bEXEC(UTE)?\b`),
		regexp.MustCompile(`(?i)\bWAITFOR\b.*\bDELAY\b`),
		regexp.MustCompile(`(?i)\bBENCHMARK\b`),
		regexp.MustCompile(`(?i)\bSLEEP\b`),
		regexp.MustCompile(`(?i)\bLOAD_FILE\b`),
		regexp.MustCompile(`(?i)\bINTO\s+(OUTFILE|DUMPFILE)\b`),
		regexp.MustCompile(`(?i)--\s*$`),
		regexp.MustCompile(`(?i);\s*\bDROP\b`),
		regexp.MustCompile(`(?i)\bOR\b\s+1\s*=\s*1`),
		regexp.MustCompile(`(?i)\bAND\b\s+1\s*=\s*1`),
	}
}