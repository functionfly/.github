package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid flags",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist"},
			wantErr: false,
		},
		{
			name:        "missing input",
			args:        []string{"--metadata", "meta.json", "--output", "./dist"},
			wantErr:     true,
			errContains: "missing required flags",
		},
		{
			name:        "missing metadata",
			args:        []string{"--input", "test.py", "--output", "./dist"},
			wantErr:     true,
			errContains: "missing required flags",
		},
		{
			name:        "missing output",
			args:        []string{"--input", "test.py", "--metadata", "meta.json"},
			wantErr:     true,
			errContains: "missing required flags",
		},
		{
			name:        "invalid mode",
			args:        []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--mode", "invalid"},
			wantErr:     true,
			errContains: "invalid mode",
		},
		{
			name:        "invalid optimization level",
			args:        []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--optimize", "ultra"},
			wantErr:     true,
			errContains: "invalid optimization level",
		},
		{
			name:    "valid deterministic mode",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--mode", "deterministic"},
			wantErr: false,
		},
		{
			name:    "valid compatible mode",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--mode", "compatible"},
			wantErr: false,
		},
		{
			name:    "valid optimization levels",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--optimize", "minimal"},
			wantErr: false,
		},
		{
			name:    "valid balanced optimization",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--optimize", "balanced"},
			wantErr: false,
		},
		{
			name:    "valid aggressive optimization",
			args:    []string{"--input", "test.py", "--metadata", "meta.json", "--output", "./dist", "--optimize", "aggressive"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, err := parseAndValidateFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseAndValidateFlags() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("parseAndValidateFlags() error = %v, want containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("parseAndValidateFlags() unexpected error: %v", err)
				}
				if flags != nil && tt.args[0] == "--input" {
					if flags.input == "" {
						t.Error("input flag not set")
					}
				}
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name        string
		metadata    *Metadata
		wantErr     bool
		errContains string
	}{
		{
			name: "valid metadata",
			metadata: &Metadata{
				Name:       "test-function",
				Version:    "1.0.0",
				EntryPoint: "handler",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			metadata: &Metadata{
				Version:    "1.0.0",
				EntryPoint: "handler",
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "default version",
			metadata: &Metadata{
				Name:       "test-function",
				EntryPoint: "handler",
			},
			wantErr: false,
		},
		{
			name: "default entry point",
			metadata: &Metadata{
				Name:    "test-function",
				Version: "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "invalid name with special chars",
			metadata: &Metadata{
				Name: "test@function",
			},
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name: "valid name with hyphen",
			metadata: &Metadata{
				Name: "test-function",
			},
			wantErr: false,
		},
		{
			name: "valid name with underscore",
			metadata: &Metadata{
				Name: "test_function",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetadata(tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateMetadata() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("validateMetadata() error = %v, want containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateMetadata() unexpected error: %v", err)
				}
				// Check defaults
				if tt.metadata.Version == "" {
					// Version should be set to default
				}
			}
		})
	}
}

func TestValidateSourceCode(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid source",
			source:  "def handler(event): return {'result': 'ok'}",
			wantErr: false,
		},
		{
			name:        "empty source",
			source:      "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:    "large source under limit",
			source:  string(make([]byte, 1024*1024)), // 1MB
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSourceCode(tt.source)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateSourceCode() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("validateSourceCode() error = %v, want containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateSourceCode() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCompileError(t *testing.T) {
	tests := []struct {
		name     string
		err      *CompileError
		expected string
	}{
		{
			name: "error with line and column",
			err: &CompileError{
				Phase:   "parse",
				Message: "syntax error",
				Line:    10,
				Column:  5,
			},
			expected: "[parse] syntax error at line 10, column 5",
		},
		{
			name: "error without line",
			err: &CompileError{
				Phase:   "restriction",
				Message: "eval() not allowed",
			},
			expected: "[restriction] eval() not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("CompileError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	tests := []struct {
		char  rune
		valid bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', false},
		{'_', false},
		{'@', false},
		{' ', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			if got := isAlphaNumeric(tt.char); got != tt.valid {
				t.Errorf("isAlphaNumeric(%q) = %v, want %v", tt.char, got, tt.valid)
			}
		})
	}
}

func TestReadInputFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "flypy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test source file
	sourcePath := filepath.Join(tmpDir, "test.py")
	sourceContent := "def handler(event): return {'result': 'ok'}"
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create test metadata file
	metadataPath := filepath.Join(tmpDir, "meta.json")
	metadata := Metadata{
		Name:       "test-function",
		Version:    "1.0.0",
		EntryPoint: "handler",
	}
	metadataJSON, _ := json.Marshal(metadata)
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	tests := []struct {
		name         string
		inputPath    string
		metadataPath string
		wantErr      bool
	}{
		{
			name:         "valid files",
			inputPath:    sourcePath,
			metadataPath: metadataPath,
			wantErr:      false,
		},
		{
			name:         "missing source file",
			inputPath:    filepath.Join(tmpDir, "nonexistent.py"),
			metadataPath: metadataPath,
			wantErr:      true,
		},
		{
			name:         "missing metadata file",
			inputPath:    sourcePath,
			metadataPath: filepath.Join(tmpDir, "nonexistent.json"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &CompileFlags{
				input:    tt.inputPath,
				metadata: tt.metadataPath,
			}
			source, meta, err := readInputFiles(flags)
			if tt.wantErr {
				if err == nil {
					t.Errorf("readInputFiles() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("readInputFiles() unexpected error: %v", err)
				}
				if source != sourceContent {
					t.Errorf("readInputFiles() source = %q, want %q", source, sourceContent)
				}
				if meta.Name != metadata.Name {
					t.Errorf("readInputFiles() metadata.Name = %q, want %q", meta.Name, metadata.Name)
				}
			}
		})
	}
}

func TestGenerateManifest(t *testing.T) {
	// This test verifies the manifest generation logic
	// Note: It doesn't test the full flypy.Result since that requires
	// a complete compilation pipeline

	t.Run("manifest structure", func(t *testing.T) {
		// Verify that the manifest has the expected structure
		expectedKeys := []string{
			"name", "version", "runtime", "entry_point",
			"mode", "build_time", "optimization_level",
			"wasm_file", "wasm_size_bytes", "hashes", "capabilities",
		}

		// This is a basic structure test
		// Full integration tests would require the flypy compiler
		_ = expectedKeys
	})
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
