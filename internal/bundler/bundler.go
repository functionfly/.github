package bundler

import (
	"github.com/functionfly/functionfly/internal/manifest"
)

// Bundle creates a bundled version of the function code.
// It handles different runtime types and delegates to appropriate bundlers.
// Dependencies are automatically installed if specified in the manifest.
func Bundle(manifest *manifest.Manifest) ([]byte, error) {
	if manifest == nil {
		return nil, NewBundlerError("bundle", "manifest cannot be nil")
	}

	// Install dependencies first
	if err := InstallDependencies(manifest); err != nil {
		return nil, NewBundlerErrorWithCause("bundle", "dependency installation failed", err)
	}

	switch manifest.Runtime {
	case "node18", "node20", "deno":
		return bundleJavaScript(manifest)
	case "python3.11":
		return bundlePython(manifest)
	default:
		return nil, &RuntimeNotSupportedError{
			Runtime:   manifest.Runtime,
			Supported: []string{"node18", "node20", "deno", "python3.11"},
		}
	}
}

// BundleWithWorkingDirectory creates a bundled version of the function code
// with explicit working directory support for consistent path resolution.
// Dependencies are automatically installed if specified in the manifest.
func BundleWithWorkingDirectory(manifest *manifest.Manifest, workingDir string) ([]byte, error) {
	if manifest == nil {
		return nil, NewBundlerError("bundle", "manifest cannot be nil")
	}

	// Resolve and validate working directory
	resolvedDir, err := ResolveWorkingDirectory(workingDir)
	if err != nil {
		return nil, NewBundlerErrorWithCause("bundle", "failed to resolve working directory", err)
	}

	// Execute bundling within the specified working directory
	var result []byte
	err = WithWorkingDirectory(resolvedDir, func() error {
		var bundleErr error
		result, bundleErr = Bundle(manifest)
		return bundleErr
	})

	return result, err
}
