package bundler

import (
	"path/filepath"

	"github.com/functionfly/functionfly/internal/manifest"
)

// findEntryFile locates the entry file based on manifest configuration and available files.
// It handles working directory resolution and provides detailed error information.
func findEntryFile(manifest *manifest.Manifest) (string, error) {
	// First check if entry is explicitly specified in manifest
	if manifest.Entry != "" {
		// Resolve the entry file path relative to working directory if needed
		entryPath := manifest.Entry
		if !filepath.IsAbs(entryPath) {
			// Assume it's relative to the project root/working directory
			entryPath = filepath.Clean(entryPath)
		}

		if err := ValidateEntryFile(entryPath); err != nil {
			return "", &EntryFileInvalidError{
				Path:   manifest.Entry,
				Reason: err.Error(),
			}
		}
		return entryPath, nil
	}

	// Fall back to auto-detection based on runtime
	preferred, alternatives := getEntryFileCandidates(manifest.Runtime)

	// Check if preferred entry file exists
	if err := ValidateEntryFile(preferred); err == nil {
		return preferred, nil
	}

	// Try alternative extensions based on runtime
	for _, alt := range alternatives {
		if err := ValidateEntryFile(alt); err == nil {
			return alt, nil
		}
	}

	return "", &EntryFileNotFoundError{
		Runtime:      manifest.Runtime,
		Preferred:    preferred,
		Alternatives: alternatives,
	}
}

// getEntryFileCandidates returns the preferred entry file and alternatives for a given runtime
func getEntryFileCandidates(runtime string) (preferred string, alternatives []string) {
	switch runtime {
	case "deno":
		return "main.ts", []string{"index.ts", "main.js", "index.js"}
	case "python3.11", "python3.12", "python":
		return "main.py", []string{"index.py", "app.py", "__main__.py"}
	case "rust":
		return "src/lib.rs", []string{"lib.rs", "main.rs"}
	case "go", "go1.21":
		return "main.go", []string{"function.go", "handler.go"}
	case "c", "c11":
		return "main.c", []string{"function.c", "handler.c"}
	case "cpp", "cpp17", "c++":
		return "main.cpp", []string{"function.cpp", "handler.cpp", "main.cc"}
	case "ruby", "ruby3.3":
		return "main.rb", []string{"function.rb", "handler.rb", "index.rb"}
	case "kotlin", "kotlin1.9":
		return "Main.kt", []string{"Function.kt", "Handler.kt", "main.kt"}
	case "swift", "swift5.9":
		return "main.swift", []string{"Function.swift", "Handler.swift"}
	default: // node18, node20, and other JS runtimes
		return "index.js", []string{"index.ts", "main.js", "main.ts"}
	}
}