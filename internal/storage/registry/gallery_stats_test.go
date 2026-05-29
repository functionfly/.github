package registry

import "testing"

func TestNormalizeGalleryRuntime(t *testing.T) {
	tests := map[string]string{
		"python3.12":      "python",
		"python-microvm":  "python",
		"nodejs":          "nodejs",
		"node20":          "nodejs",
		"javascript":      "nodejs",
		"deno":            "typescript",
		"bun":             "typescript",
		"typescript":      "typescript",
		"go1.22":          "go",
		"rust":            "rust",
		"java17":          "java",
		"c#":              "csharp",
		"csharp":          "csharp",
		"ruby3":           "ruby",
		"php8":            "php",
		"":                "",
		"   ":             "",
	}

	for input, want := range tests {
		if got := normalizeGalleryRuntime(input); got != want {
			t.Errorf("normalizeGalleryRuntime(%q) = %q, want %q", input, got, want)
		}
	}
}
