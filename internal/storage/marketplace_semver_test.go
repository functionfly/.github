package storage

import "testing"

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"1.0.0", []int{1, 0, 0}},
		{"v1.0.0", []int{1, 0, 0}},
		{"1.2.3", []int{1, 2, 3}},
		{"1.0", []int{1, 0, 0}},
		{"10.20.30", []int{10, 20, 30}},
		{"1.0.0-beta", []int{1, 0, 0}},
		{"v2.1", []int{2, 1, 0}},
		{"0.0.1", []int{0, 0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSemver(tt.input)
			if len(got) != 3 {
				t.Fatalf("expected 3 elements, got %d", len(got))
			}
			if got[0] != tt.expected[0] || got[1] != tt.expected[1] || got[2] != tt.expected[2] {
				t.Errorf("parseSemver(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsNewerVersion_ValidSemver(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.1", false},
		{"1.1.0", "1.0.0", true},
		{"1.0.0", "1.1.0", false},
		{"2.0.0", "1.99.99", true},
		{"1.99.99", "2.0.0", false},
		{"1.0.0", "0.9.9", true},
		{"0.9.9", "1.0.0", false},
		{"v1.0.1", "v1.0.0", true},
		{"v2.0.0", "v1.0.0", true},
		{"1.0", "0.9", true},
		{"1.0", "1.0.1", false},
		{"1.0.1", "1.0", true},
		{"10.0.0", "9.99.99", true},
	}
	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewerVersion(tt.latest, tt.current)
			if got != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
					tt.latest, tt.current, got, tt.expected)
			}
		})
	}
}

func TestIsNewerVersion_InvalidSemver(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"abc", "def", false},
		{"xyz", "abc", true},
		{"1.0.0", "abc", true},
		{"abc", "1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewerVersion(tt.latest, tt.current)
			if got != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
					tt.latest, tt.current, got, tt.expected)
			}
		})
	}
}

func TestIsNewerVersion_PrereleaseHandling(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"1.0.0-beta", "1.0.0-alpha", true},
		{"1.0.0", "1.0.0-beta", true},
		{"1.0.0-rc.1", "1.0.0-rc.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewerVersion(tt.latest, tt.current)
			if got != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
					tt.latest, tt.current, got, tt.expected)
			}
		})
	}
}
