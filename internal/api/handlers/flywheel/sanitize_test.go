package flywheel

import (
	"strings"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "empty string",
			input: "",
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty, got %q", result)
				}
			},
		},
		{
			name:  "plain text preserved",
			input: "Hello, World!",
			check: func(t *testing.T, result string) {
				if result != "Hello, World!" {
					t.Errorf("expected 'Hello, World!', got %q", result)
				}
			},
		},
		{
			name:  "script tag removed",
			input: "Hello <script>alert('xss')</script> World",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "script") {
					t.Errorf("script tag not removed, got %q", result)
				}
				if strings.Contains(result, "alert") {
					t.Errorf("script content not removed, got %q", result)
				}
			},
		},
		{
			name:  "iframe tag removed",
			input: "Check <iframe src=\"http://evil.com\"></iframe> this",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "iframe") {
					t.Errorf("iframe tag not removed, got %q", result)
				}
				if strings.Contains(result, "evil.com") {
					t.Errorf("iframe src not removed, got %q", result)
				}
			},
		},
		{
			name:  "object tag removed",
			input: "<object data=\"evil.swf\"></object>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "object") {
					t.Errorf("object tag not removed, got %q", result)
				}
			},
		},
		{
			name:  "embed tag removed",
			input: "<embed src=\"evil.swf\">",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "embed") {
					t.Errorf("embed tag not removed, got %q", result)
				}
			},
		},
		{
			name:  "onclick event handler neutralized",
			input: "<button onclick=\"alert('xss')\">Click me</button>",
			check: func(t *testing.T, result string) {
				// Event handler should be neutralized (replaced with data-removed)
				if strings.Contains(result, "onclick=") {
					t.Errorf("onclick not neutralized, got %q", result)
				}
			},
		},
		{
			name:  "onerror event handler neutralized",
			input: "<img src=\"x\" onerror=\"alert('xss')\">",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "onerror=") {
					t.Errorf("onerror not neutralized, got %q", result)
				}
			},
		},
		{
			name:  "javascript URL neutralized",
			input: "<a href=\"javascript:alert('xss')\">Click</a>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "javascript:") {
					t.Errorf("javascript URL not neutralized, got %q", result)
				}
			},
		},
		{
			name:  "data URL neutralized",
			input: "<img src=\"data:text/html,<script>alert('xss')</script>\">",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "data:text/html") {
					t.Errorf("data URL not neutralized, got %q", result)
				}
			},
		},
		{
			name:  "SVG with event handler neutralized",
			input: "<svg onload=\"alert('xss')\"></svg>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "onload=") {
					t.Errorf("SVG onload not neutralized, got %q", result)
				}
			},
		},
		{
			name:  "safe tags preserved",
			input: "<p>Hello <strong>World</strong>!</p>",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "<p>") {
					t.Errorf("p tag not preserved, got %q", result)
				}
				if !strings.Contains(result, "<strong>") {
					t.Errorf("strong tag not preserved, got %q", result)
				}
			},
		},
		{
			name:  "code tag preserved",
			input: "<code>console.log('test')</code>",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "<code>") {
					t.Errorf("code tag not preserved, got %q", result)
				}
			},
		},
		{
			name:  "html entities preserved",
			input: "Use &lt;div&gt; for less-than comparison",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "&lt;") {
					t.Errorf("html entities not preserved, got %q", result)
				}
			},
		},
		{
			name:  "applet tag removed",
			input: "<applet code=\"evil.class\"></applet>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "applet") {
					t.Errorf("applet tag not removed, got %q", result)
				}
			},
		},
		{
			name:  "form tag removed",
			input: "<form action=\"evil\"><input name=\"x\"></form>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "<form") {
					t.Errorf("form tag not removed, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			tt.check(t, result)
		})
	}
}

func TestSanitizeContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "empty string",
			input: "",
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty, got %q", result)
				}
			},
		},
		{
			name:  "whitespace trimmed",
			input: "   Hello World   ",
			check: func(t *testing.T, result string) {
				if result != "Hello World" {
					t.Errorf("expected 'Hello World', got %q", result)
				}
			},
		},
		{
			name:  "script tag in content removed",
			input: "Try this: <script>alert(1)</script>",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "script") {
					t.Errorf("script not removed, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeContent(tt.input)
			tt.check(t, result)
		})
	}
}

func TestSanitizeThread(t *testing.T) {
	title, content := SanitizeThread("<script>bad()</script>Title", "Hello <img onerror='x' src='x'>")

	// Title should not contain script
	if strings.Contains(title, "script") {
		t.Errorf("title still contains script: %q", title)
	}

	// Content should not contain onerror
	if strings.Contains(content, "onerror=") {
		t.Errorf("content still contains onerror: %q", content)
	}
}

func TestSanitizeReply(t *testing.T) {
	content := SanitizeReply("Test <script>alert(1)</script>content")

	// Should not contain script
	if strings.Contains(content, "script") {
		t.Errorf("reply still contains script: %q", content)
	}
}

func TestSanitizeCategory(t *testing.T) {
	name, desc := SanitizeCategory("<script>bad()</script>Category", "Description <img onerror='x'>")

	// Name should not contain script
	if strings.Contains(name, "script") {
		t.Errorf("category name still contains script: %q", name)
	}

	// Description should not contain onerror
	if strings.Contains(desc, "onerror=") {
		t.Errorf("category desc still contains onerror: %q", desc)
	}
}

func TestSanitizeChallenge(t *testing.T) {
	title, desc := SanitizeChallenge("<script>bad()</script>Challenge", "Description <iframe src='evil'></iframe>")

	// Title should not contain script
	if strings.Contains(title, "script") {
		t.Errorf("challenge title still contains script: %q", title)
	}

	// Description should not contain iframe
	if strings.Contains(desc, "iframe") {
		t.Errorf("challenge desc still contains iframe: %q", desc)
	}
}
