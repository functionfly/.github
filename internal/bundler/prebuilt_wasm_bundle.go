package bundler

import (
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/manifest"
)

// bundlePrebuiltWasm returns WASM bytes for runtime "browser-wasm": reads a .wasm entry or
// compiles .wat using wat2wasm (see compileWATToWasm).
func bundlePrebuiltWasm(m *manifest.Manifest) ([]byte, error) {
	entryPath, data, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("browser-wasm bundle", "failed to read entry file", err)
	}

	ext := strings.ToLower(filepath.Ext(entryPath))
	switch ext {
	case ".wasm":
		return data, nil
	case ".wat":
		wasm, err := compileWATToWasm(string(data))
		if err != nil {
			return nil, NewBundlerErrorWithCause("browser-wasm bundle", "wat2wasm failed (install WABT: https://github.com/WebAssembly/wabt)", err)
		}
		return wasm, nil
	default:
		return nil, NewBundlerError("browser-wasm bundle", "entry must be .wasm or .wat for runtime browser-wasm")
	}
}
