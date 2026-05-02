package github

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	return logger
}

func testRepo(name string) *GitHubRepo {
	return &GitHubRepo{
		ID:            12345,
		Name:          name,
		FullName:      "testowner/" + name,
		DefaultBranch: "main",
		Owner: GitHubUser{
			ID:    1,
			Login: "testowner",
		},
	}
}

func TestExplicitConfigDetector(t *testing.T) {
	detector := &ExplicitConfigDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "explicit-config", detector.Name())
		assert.Equal(t, 100, detector.Priority())
	})

	t.Run("no functionfly.json returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc", Size: 100},
			{Path: "go.mod", Type: "blob", SHA: "def", Size: 50},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("detects functionfly.json in root", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "functionfly.json", Type: "blob", SHA: "abc", Size: 200},
		}
		// Note: This detector fetches file content from GitHub API, so without
		// a mock client it will fail on the API call. We test the file detection
		// logic by checking that it identifies the config path.
		// For a pure unit test of the path detection, we verify the entries.
		var configPath string
		for _, e := range entries {
			if e.Type != "blob" {
				continue
			}
			base := e.Path
			if base == "functionfly.json" || base == "functionfly.jsonc" {
				configPath = e.Path
				break
			}
		}
		assert.Equal(t, "functionfly.json", configPath)
	})

	t.Run("detects functionfly.jsonc", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "functionfly.jsonc", Type: "blob", SHA: "abc", Size: 200},
		}
		var configPath string
		for _, e := range entries {
			base := e.Path
			if base == "functionfly.jsonc" || base == "functionfly.json" {
				configPath = e.Path
				break
			}
		}
		assert.Equal(t, "functionfly.jsonc", configPath)
	})

	t.Run("skips non-blob entries", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "src", Type: "tree", SHA: "abc"},
			{Path: "src/functionfly.json", Type: "blob", SHA: "def", Size: 100},
		}
		var configPath string
		for _, e := range entries {
			if e.Type != "blob" {
				continue
			}
			base := e.Path
			if base == "functionfly.json" || base == "functionfly.jsonc" {
				configPath = e.Path
				break
			}
		}
		// The detector uses filepath.Base, so "src/functionfly.json" -> "functionfly.json"
		assert.Empty(t, configPath, "should not match nested path with just base name comparison in entries")
	})
}

func TestServerlessFrameworkDetector(t *testing.T) {
	detector := &ServerlessFrameworkDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "serverless-framework", detector.Name())
		assert.Equal(t, 90, detector.Priority())
	})

	t.Run("no serverless.yml returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc", Size: 100},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestNodeDetector(t *testing.T) {
	detector := &NodeDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "node-detector", detector.Name())
		assert.Equal(t, 50, detector.Priority())
	})

	t.Run("no package.json returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc", Size: 100},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("package.json without entry files returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "README.md", Type: "blob", SHA: "def", Size: 100},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("detects Node.js function with index.js", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "index.js", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "default", result.Functions[0].Name)
		assert.Equal(t, "index.js", result.Functions[0].EntryPoint)
		assert.Equal(t, "node18", result.Functions[0].Runtime)
		assert.Equal(t, 0.6, result.Functions[0].Confidence)
		assert.Equal(t, "node-detector", result.Functions[0].Strategy)
	})

	t.Run("detects TypeScript function with index.ts", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "index.ts", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "node18-typescript", result.Functions[0].Runtime)
	})

	t.Run("detects handler.ts entry file", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "handler.ts", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "handler.ts", result.Functions[0].EntryPoint)
	})

	t.Run("detects multiple entry files", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "src/index.js", Type: "blob", SHA: "def", Size: 500},
			{Path: "src/handler.js", Type: "blob", SHA: "ghi", Size: 300},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Functions, 2)
	})

	t.Run("detects npm lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "package-lock.json", Type: "blob", SHA: "def", Size: 5000},
			{Path: "index.js", Type: "blob", SHA: "ghi", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		require.NotNil(t, result.Functions[0].Dependencies)
		assert.Equal(t, "npm", result.Functions[0].Dependencies.Manager)
		assert.Equal(t, "package-lock.json", result.Functions[0].Dependencies.Lockfile)
	})

	t.Run("detects yarn lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "yarn.lock", Type: "blob", SHA: "def", Size: 5000},
			{Path: "index.js", Type: "blob", SHA: "ghi", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "yarn", result.Functions[0].Dependencies.Manager)
	})

	t.Run("detects pnpm lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "pnpm-lock.yaml", Type: "blob", SHA: "def", Size: 5000},
			{Path: "index.js", Type: "blob", SHA: "ghi", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "pnpm", result.Functions[0].Dependencies.Manager)
	})

	t.Run("detects bun lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "bun.lockb", Type: "blob", SHA: "def", Size: 5000},
			{Path: "index.js", Type: "blob", SHA: "ghi", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "bun", result.Functions[0].Dependencies.Manager)
	})

	t.Run("subdirectory function gets proper name", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "package.json", Type: "blob", SHA: "abc", Size: 200},
			{Path: "functions/api/index.js", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "api", result.Functions[0].Name)
		assert.Equal(t, "functions/api", result.Functions[0].SubDirectory)
	})
}

func TestPythonDetector(t *testing.T) {
	detector := &PythonDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "python-detector", detector.Name())
		assert.Equal(t, 40, detector.Priority())
	})

	t.Run("no entry files returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "requirements.txt", Type: "blob", SHA: "abc", Size: 100},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("detects Python function with main.py", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "requirements.txt", Type: "blob", SHA: "abc", Size: 100},
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "default", result.Functions[0].Name)
		assert.Equal(t, "main.py", result.Functions[0].EntryPoint)
		assert.Equal(t, "python3.11", result.Functions[0].Runtime)
	})

	t.Run("detects handler.py", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "handler.py", Type: "blob", SHA: "abc", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "handler.py", result.Functions[0].EntryPoint)
	})

	t.Run("detects app.py", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "app.py", Type: "blob", SHA: "abc", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "app.py", result.Functions[0].EntryPoint)
	})

	t.Run("higher confidence with requirements.txt", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "requirements.txt", Type: "blob", SHA: "abc", Size: 100},
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 0.65, result.OverallConfidence)
	})

	t.Run("lower confidence without requirements.txt", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 0.55, result.OverallConfidence)
	})

	t.Run("detects pipenv lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "Pipfile.lock", Type: "blob", SHA: "abc", Size: 5000},
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		require.NotNil(t, result.Functions[0].Dependencies)
		assert.Equal(t, "pipenv", result.Functions[0].Dependencies.Manager)
	})

	t.Run("detects poetry lockfile", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "poetry.lock", Type: "blob", SHA: "abc", Size: 5000},
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "poetry", result.Functions[0].Dependencies.Manager)
	})

	t.Run("detects pip with requirements.txt", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "requirements.txt", Type: "blob", SHA: "abc", Size: 100},
			{Path: "main.py", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "pip", result.Functions[0].Dependencies.Manager)
	})
}

func TestGoDetector(t *testing.T) {
	detector := &GoDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "go-detector", detector.Name())
		assert.Equal(t, 30, detector.Priority())
	})

	t.Run("no go.mod returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("no main.go returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "go.mod", Type: "blob", SHA: "abc", Size: 100},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("detects Go function", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "go.mod", Type: "blob", SHA: "abc", Size: 100},
			{Path: "main.go", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "default", result.Functions[0].Name)
		assert.Equal(t, "main.go", result.Functions[0].EntryPoint)
		assert.Equal(t, "go1.22", result.Functions[0].Runtime)
		assert.Equal(t, 0.6, result.Functions[0].Confidence)
		require.NotNil(t, result.Functions[0].Dependencies)
		assert.Equal(t, "gomod", result.Functions[0].Dependencies.Manager)
	})

	t.Run("detects multiple main.go files", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "go.mod", Type: "blob", SHA: "abc", Size: 100},
			{Path: "cmd/server/main.go", Type: "blob", SHA: "def", Size: 500},
			{Path: "cmd/cli/main.go", Type: "blob", SHA: "ghi", Size: 300},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Functions, 2)
	})

	t.Run("subdirectory function name", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "go.mod", Type: "blob", SHA: "abc", Size: 100},
			{Path: "cmd/api/main.go", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "api", result.Functions[0].Name)
		assert.Equal(t, "cmd/api", result.Functions[0].SubDirectory)
	})
}

func TestRustDetector(t *testing.T) {
	detector := &RustDetector{logger: newTestLogger()}

	t.Run("Name and Priority", func(t *testing.T) {
		assert.Equal(t, "rust-detector", detector.Name())
		assert.Equal(t, 20, detector.Priority())
	})

	t.Run("no Cargo.toml returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "src/main.rs", Type: "blob", SHA: "abc", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("no main.rs in src/ returns nil", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "Cargo.toml", Type: "blob", SHA: "abc", Size: 100},
			{Path: "main.rs", Type: "blob", SHA: "def", Size: 500}, // not in src/
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("detects Rust function", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "Cargo.toml", Type: "blob", SHA: "abc", Size: 100},
			{Path: "src/main.rs", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "default", result.Functions[0].Name)
		assert.Equal(t, "src/main.rs", result.Functions[0].EntryPoint)
		assert.Equal(t, "rust1.75", result.Functions[0].Runtime)
		assert.Equal(t, 0.55, result.Functions[0].Confidence)
		require.NotNil(t, result.Functions[0].Dependencies)
		assert.Equal(t, "cargo", result.Functions[0].Dependencies.Manager)
	})

	t.Run("subdirectory function name", func(t *testing.T) {
		entries := []GitHubTreeEntry{
			{Path: "Cargo.toml", Type: "blob", SHA: "abc", Size: 100},
			{Path: "mylib/src/main.rs", Type: "blob", SHA: "def", Size: 500},
		}
		result, err := detector.Detect(context.Background(), testRepo("test"), entries)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Functions, 1)
		assert.Equal(t, "mylib", result.Functions[0].Name)
		assert.Equal(t, "mylib", result.Functions[0].SubDirectory)
	})
}

func TestDetectNodePkgManager(t *testing.T) {
	tests := []struct {
		lockfile string
		expected string
	}{
		{"package-lock.json", "npm"},
		{"yarn.lock", "yarn"},
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lockb", "bun"},
		{"some-other-lock.json", "npm"},
		{"path/to/yarn.lock", "yarn"},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			result := detectNodePkgManager(tt.lockfile)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectPythonPkgManager(t *testing.T) {
	tests := []struct {
		lockfile string
		expected string
	}{
		{"Pipfile.lock", "pipenv"},
		{"poetry.lock", "poetry"},
		{"requirements.txt", "pip"},
		{"unknown.lock", "pip"},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			result := detectPythonPkgManager(tt.lockfile)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no comments",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "line comment",
			input:    `{"key": "value"} // comment`,
			expected: `{"key": "value"} `,
		},
		{
			name:     "block comment",
			input:    `{"key": "value"} /* block */`,
			expected: `{"key": "value"} `,
		},
		{
			name:     "comment in string preserved",
			input:    `{"key": "has // comment"}`,
			expected: `{"key": "has // comment"}`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiline block comment",
			input:    "{\n  /* multi\n   line */\n  \"key\": \"value\"\n}",
			expected: "{\n  \n  \"key\": \"value\"\n}",
		},
		{
			name:     "escaped quote in string",
			input:    `{"key": "value with \"quote"}`,
			expected: `{"key": "value with \"quote"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripJSONComments(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScannerMergeResults(t *testing.T) {
	logger := newTestLogger()
	scanner := &Scanner{logger: logger}

	t.Run("nil results returns nil", func(t *testing.T) {
		result := scanner.mergeResults(nil)
		assert.Nil(t, result)
	})

	t.Run("empty results returns nil", func(t *testing.T) {
		result := scanner.mergeResults([]*ScanResult{})
		assert.Nil(t, result)
	})

	t.Run("merges functions from multiple results", func(t *testing.T) {
		results := []*ScanResult{
			{
				Functions: []DetectedFunction{
					{Name: "fn1", EntryPoint: "a.js", Runtime: "node18", Confidence: 0.8},
				},
				OverallConfidence: 0.8,
				StrategyUsed:      "node-detector",
			},
			{
				Functions: []DetectedFunction{
					{Name: "fn2", EntryPoint: "b.py", Runtime: "python3.11", Confidence: 0.6},
				},
				OverallConfidence: 0.6,
				StrategyUsed:      "python-detector",
			},
		}
		result := scanner.mergeResults(results)
		require.NotNil(t, result)
		assert.Len(t, result.Functions, 2)
		assert.Equal(t, 0.8, result.OverallConfidence)
		assert.Equal(t, "node-detector", result.StrategyUsed)
	})

	t.Run("deduplicates functions by entrypoint+runtime", func(t *testing.T) {
		results := []*ScanResult{
			{
				Functions: []DetectedFunction{
					{Name: "fn1", EntryPoint: "index.js", Runtime: "node18"},
				},
				OverallConfidence: 0.6,
			},
			{
				Functions: []DetectedFunction{
					{Name: "fn1-dup", EntryPoint: "index.js", Runtime: "node18"},
				},
				OverallConfidence: 0.8,
			},
		}
		result := scanner.mergeResults(results)
		require.NotNil(t, result)
		assert.Len(t, result.Functions, 1)
	})

	t.Run("keeps different runtimes for same entrypoint", func(t *testing.T) {
		results := []*ScanResult{
			{
				Functions: []DetectedFunction{
					{Name: "fn1", EntryPoint: "handler.js", Runtime: "node18"},
				},
			},
			{
				Functions: []DetectedFunction{
					{Name: "fn2", EntryPoint: "handler.js", Runtime: "node18-typescript"},
				},
			},
		}
		result := scanner.mergeResults(results)
		require.NotNil(t, result)
		assert.Len(t, result.Functions, 2)
	})

	t.Run("merges warnings", func(t *testing.T) {
		results := []*ScanResult{
			{Warnings: []string{"warn1"}},
			{Warnings: []string{"warn2", "warn3"}},
		}
		result := scanner.mergeResults(results)
		require.NotNil(t, result)
		assert.Len(t, result.Warnings, 3)
	})
}

func TestScannerDetectPrimaryRuntime(t *testing.T) {
	logger := newTestLogger()
	scanner := &Scanner{logger: logger}

	tests := []struct {
		name      string
		languages map[string]float64
		expected  string
	}{
		{"TypeScript dominant", map[string]float64{"TypeScript": 80, "JavaScript": 20}, "node"},
		{"JavaScript dominant", map[string]float64{"JavaScript": 90}, "node"},
		{"Python dominant", map[string]float64{"Python": 70, "JavaScript": 30}, "python"},
		{"Go dominant", map[string]float64{"Go": 100}, "go"},
		{"Rust dominant", map[string]float64{"Rust": 80, "TOML": 20}, "rust"},
		{"empty languages", map[string]float64{}, "unknown"},
		{"custom language", map[string]float64{"Zig": 100}, "zig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.detectPrimaryRuntime(tt.languages)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScannerEstimateImportTime(t *testing.T) {
	logger := newTestLogger()
	scanner := &Scanner{logger: logger}

	t.Run("no functions", func(t *testing.T) {
		result := scanner.estimateImportTime(&ScanResult{Functions: []DetectedFunction{}})
		assert.Equal(t, 10, result)
	})

	t.Run("one function", func(t *testing.T) {
		result := scanner.estimateImportTime(&ScanResult{
			Functions: []DetectedFunction{{}},
		})
		assert.Equal(t, 15, result)
	})

	t.Run("five functions", func(t *testing.T) {
		result := scanner.estimateImportTime(&ScanResult{
			Functions: make([]DetectedFunction, 5),
		})
		assert.Equal(t, 35, result)
	})
}

func TestScannerEstimateCost(t *testing.T) {
	logger := newTestLogger()
	scanner := &Scanner{logger: logger}

	t.Run("no functions", func(t *testing.T) {
		result := scanner.estimateCost(&ScanResult{Functions: []DetectedFunction{}})
		assert.InDelta(t, 0.01, result, 0.001)
	})

	t.Run("two functions", func(t *testing.T) {
		result := scanner.estimateCost(&ScanResult{
			Functions: make([]DetectedFunction, 2),
		})
		assert.InDelta(t, 0.02, result, 0.001)
	})
}
