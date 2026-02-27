package bundler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/manifest"
)

// bundleJavaScript bundles JavaScript/TypeScript code using esbuild
func bundleJavaScript(manifest *manifest.Manifest) ([]byte, error) {
	// Read and validate entry file using shared helper
	entryFile, sourceCode, err := ReadEntryFile(manifest)
	if err != nil {
		return nil, NewBundlerErrorWithCause("javascript bundle", "failed to read entry file", err)
	}

	// Check if esbuild is available
	if _, err := exec.LookPath("esbuild"); err != nil {
		// Fallback to simple file reading for development
		fmt.Println("Warning: esbuild not found, using simple bundling")
		return sourceCode, nil // simpleBundle just returns the source code
	}

	// Create temporary output file with unique name to avoid conflicts
	tempOut := filepath.Join(os.TempDir(), fmt.Sprintf("functionfly-js-bundle-%d.js", os.Getpid()))
	defer os.Remove(tempOut)

	// Build esbuild command with optimized settings
	args := []string{
		entryFile,
		"--bundle",
		"--minify",
		"--format=esm",
		"--target=node18",
		"--platform=node",
		fmt.Sprintf("--outfile=%s", tempOut),
		"--sourcemap", // Include sourcemaps for better debugging
	}

	// Add TypeScript support if needed
	if strings.HasSuffix(entryFile, ".ts") || strings.HasSuffix(entryFile, ".tsx") {
		args = append(args, "--loader:.ts=ts", "--loader:.tsx=tsx")
	}

	cmd := exec.Command("esbuild", args...)

	// Execute compilation
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("esbuild", entryFile, string(output), err)
	}

	// Read the bundled output
	bundle, err := os.ReadFile(tempOut)
	if err != nil {
		return nil, NewBundlerErrorWithCause("javascript bundle", "failed to read bundled output", err)
	}

	if len(bundle) == 0 {
		return nil, NewBundlerError("javascript bundle", "bundled output is empty")
	}

	return bundle, nil
}