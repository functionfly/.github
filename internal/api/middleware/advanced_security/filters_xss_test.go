package advanced_security

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestXSSQueryPattern_eventHandlerRequiresWordBoundary(t *testing.T) {
	p := regexp.MustCompile(`(?i)\bon\w+\s*=`)

	benign := []string{
		"component=all&period=30d&resolution=day",
		"component=all&period=24h",
		"provider=all&period=24h&percentile=p95",
	}
	for _, raw := range benign {
		if p.MatchString(raw) {
			t.Errorf("false positive on benign RawQuery %q", raw)
		}
	}

	malicious := []string{
		"x=onclick=alert(1)",
		"q=<script>alert(1)</script>",
	}
	for _, raw := range malicious {
		u := &url.URL{RawQuery: raw}
		r := &http.Request{URL: u}
		if !xssPatternsForTest().Detect(r) {
			t.Errorf("expected XSS detection for RawQuery %q", raw)
		}
	}
}

func xssPatternsForTest() *XSSFilter {
	return &XSSFilter{patterns: []*regexp.Regexp{
		regexp.MustCompile(`(?i)(<script|<iframe|<object|<embed)`),
		regexp.MustCompile(`(?i)(javascript:|vbscript:|data:)`),
		regexp.MustCompile(`(?i)\bon\w+\s*=`),
	}}
}
