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