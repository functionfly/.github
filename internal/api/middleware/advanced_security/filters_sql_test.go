package advanced_security

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSQLInjectionQueryPatterns_allowsBenignQueries(t *testing.T) {
	filter := &SQLInjectionFilter{patterns: sqlInjectionQueryPatterns()}
	cases := []string{
		"limit=10&offset=0",
		"redirect=http%3A%2F%2Flocalhost%3A3000%2Fcallback%3Fcode%3Dabc%26state%3Dxyz",
		"filter=%28status+eq+open%29",
		"q=hello+world",
		"sort=created_at&order=desc",
		"scope=read%3Awrite",
	}
	for _, raw := range cases {
		u := &url.URL{RawQuery: raw}
		r := &http.Request{URL: u}
		if filter.Detect(r) {
			t.Errorf("false positive for benign RawQuery %q", raw)
		}
	}
}

func TestSQLInjectionQueryPatterns_blocksClassicPayloads(t *testing.T) {
	filter := &SQLInjectionFilter{patterns: sqlInjectionQueryPatterns()}
	cases := []string{
		"id=1+union+select+*+from+users",
		"q=delete+from+users",
		"sort=1+or+1%3D1",
		"x=%27+or+1%3D1",
	}
	for _, raw := range cases {
		u := &url.URL{RawQuery: raw}
		r := &http.Request{URL: u}
		if !filter.Detect(r) {
			t.Errorf("expected detection for malicious RawQuery %q", raw)
		}
	}
}
