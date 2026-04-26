package manifest

import (
	"testing"
)

// ─── Universal WASM Pipeline Manifest Tests ────────────────────────────────────

func TestValidate_AllNewRuntimes(t *testing.T) {
	t.Parallel()
	runtimes := []string{
		"rust", "go", "go1.21",
		"c", "c11", "cpp", "cpp17", "c++",
		"ruby", "ruby3.3",
		"kotlin", "kotlin1.9",
		"swift", "swift5.9",
	}

	for _, rt := range runtimes {
		t.Run(rt, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{
				Name:    "test-fn",
				Version: "1.0.0",
				Runtime: rt,
			}
			if err := m.Validate(); err != nil {
				t.Errorf("runtime %q should be valid: %v", rt, err)
			}
		})
	}
}

func TestValidate_EntryExtensions_AllRuntimes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		runtime string
		valid   []string
		invalid []string
	}{
		{"rust", []string{"main.rs"}, []string{"main.go", "main.c", "main.py"}},
		{"go", []string{"main.go"}, []string{"main.rs", "main.c", "main.py"}},
		{"go1.21", []string{"handler.go"}, []string{"handler.rs"}},
		{"c", []string{"main.c"}, []string{"main.cpp", "main.go"}},
		{"c11", []string{"test.c"}, []string{"test.rs"}},
		{"cpp", []string{"main.cpp", "main.cc", "main.cxx"}, []string{"main.c", "main.go"}},
		{"cpp17", []string{"test.cpp"}, []string{"test.c"}},
		{"c++", []string{"test.cpp"}, []string{"test.c"}},
		{"ruby", []string{"main.rb"}, []string{"main.py", "main.go"}},
		{"ruby3.3", []string{"test.rb"}, []string{"test.py"}},
		{"kotlin", []string{"Main.kt"}, []string{"Main.java", "Main.swift"}},
		{"kotlin1.9", []string{"test.kt"}, []string{"test.java"}},
		{"swift", []string{"main.swift"}, []string{"main.kt", "main.rs"}},
		{"swift5.9", []string{"test.swift"}, []string{"test.kt"}},
	}

	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			t.Parallel()
			for _, entry := range tt.valid {
				m := &Manifest{Name: "test", Version: "1.0.0", Runtime: tt.runtime, Entry: entry}
				if err := m.Validate(); err != nil {
					t.Errorf("runtime=%q entry=%q should be valid: %v", tt.runtime, entry, err)
				}
			}
			for _, entry := range tt.invalid {
				m := &Manifest{Name: "test", Version: "1.0.0", Runtime: tt.runtime, Entry: entry}
				if err := m.Validate(); err == nil {
					t.Errorf("runtime=%q entry=%q should be invalid", tt.runtime, entry)
				}
			}
		})
	}
}

func TestValidate_LegacyRuntimes_StillWork(t *testing.T) {
	t.Parallel()
	legacy := []string{"node18", "node20", "python3.11", "python3.12", "deno", "bun", "browser-wasm"}
	for _, rt := range legacy {
		t.Run(rt, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: "1.0.0", Runtime: rt}
			if err := m.Validate(); err != nil {
				t.Errorf("legacy runtime %q should still be valid: %v", rt, err)
			}
		})
	}
}

func TestValidate_InvalidRuntimes(t *testing.T) {
	t.Parallel()
	invalid := []string{
		"java", "scala", "haskell", "erlang", "elixir",
		"php", "perl", "lua", "r", "julia",
		"python3.10", "python3.9", "node16",
		"random_string", "",
	}
	for _, rt := range invalid {
		t.Run(rt, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: "1.0.0", Runtime: rt}
			if err := m.Validate(); err == nil {
				t.Errorf("runtime %q should be invalid", rt)
			}
		})
	}
}

func TestValidate_NameConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		valid bool
	}{
		{"hello", true},
		{"hello-world", true},
		{"test-fn-123", true},
		{"Hello", false},       // uppercase
		{"hello_world", false}, // underscore
		{"-hello", false},      // leading hyphen
		{"hello-", false},      // trailing hyphen
		{"", false},            // empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: tt.name, Version: "1.0.0", Runtime: "rust"}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("name %q should be valid: %v", tt.name, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("name %q should be invalid", tt.name)
			}
		})
	}
}

func TestValidate_VersionConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0", false},
		{"1", false},
		{"v1.0.0", false},
		{"1.0.0-beta", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: tt.version, Runtime: "rust"}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("version %q should be valid: %v", tt.version, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("version %q should be invalid", tt.version)
			}
		})
	}
}

func TestValidate_MemoryConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		memory int
		valid  bool
	}{
		{128, true},
		{256, true},
		{512, true},
		{1024, true},
		{64, false},
		{2048, false},
		{0, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: "1.0.0", Runtime: "rust", MemoryMB: &tt.memory}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("memory_mb=%d should be valid: %v", tt.memory, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("memory_mb=%d should be invalid", tt.memory)
			}
		})
	}
}

func TestValidate_TimeoutConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		timeout int
		valid   bool
	}{
		{1000, true},
		{5000, true},
		{30000, true},
		{500, false},
		{31000, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: "1.0.0", Runtime: "rust", TimeoutMS: &tt.timeout}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("timeout_ms=%d should be valid: %v", tt.timeout, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("timeout_ms=%d should be invalid", tt.timeout)
			}
		})
	}
}

func TestValidate_DescriptionLength(t *testing.T) {
	t.Parallel()
	short := &Manifest{Name: "test", Version: "1.0.0", Runtime: "rust", Description: "A short description"}
	if err := short.Validate(); err != nil {
		t.Errorf("short description should be valid: %v", err)
	}

	longDesc := make([]byte, 501)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	long := &Manifest{Name: "test", Version: "1.0.0", Runtime: "rust", Description: string(longDesc)}
	if err := long.Validate(); err == nil {
		t.Error("501-char description should be invalid")
	}
}

func TestValidate_PathTraversal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		entry string
		valid bool
	}{
		{"main.rs", true},
		{"../etc/passwd", false},
		{"src/main.rs", false},
		{"main\\test.rs", false},
	}

	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			t.Parallel()
			m := &Manifest{Name: "test", Version: "1.0.0", Runtime: "rust", Entry: tt.entry}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("entry %q should be valid: %v", tt.entry, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("entry %q should be invalid (path traversal)", tt.entry)
			}
		})
	}
}
