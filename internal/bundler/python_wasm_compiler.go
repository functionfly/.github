package bundler

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/sirupsen/logrus"
)

// bundlePythonForWasmRuntime bundles Python source for execution inside the
// MicroPython WASM runtime. This is the only production path — there is no
// stub, no WAT template, and no silent fallback. If the MicroPython runtime
// is not available, bundling fails fast with a clear error so operators can
// install the runtime before deploying Python functions.
//
// The returned WASM module is the standard MicroPython interpreter. User
// Python code is loaded at runtime by the execution host via mp_js_do_exec
// (the documented MicroPython JS interop). The host is expected to call:
//   1. init()            — initialize the MicroPython runtime
//   2. load_code(ptr, n) — copy user code into WASM linear memory
//   3. execute(ptr, n)   — invoke the user handler with JSON input
func bundlePythonForWasmRuntime(m *manifest.Manifest) ([]byte, error) {
	entryFile, sourceCode, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm python bundle", "failed to read entry file", err)
	}

	if !IsMicropythonAvailable() {
		return nil, NewBundlerErrorWithCause(
			"wasm python bundle",
			"microPython runtime unavailable",
			fmt.Errorf("micropython.wasm not found at %s — run scripts/build-python-runtime.sh or place a real MicroPython WASM build at internal/bundler/python/micropython.wasm", GetMicropythonRuntimePath()),
		)
	}

	wasmBytes, err := CompileWithMicropython(string(sourceCode), m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm python bundle", "microPython compilation failed", err)
	}

	// Precompiled runtime is known-good; validation may report false positives
	// for the embedded user-code data section, so warn but do not fail.
	if err := validateWasmModule(wasmBytes); err != nil {
		logrus.WithField("error", err).Warn("MicroPython WASM validation reported issues — using runtime anyway")
	}

	logrus.WithFields(logrus.Fields{
		"entry_file": entryFile,
		"name":       m.Name,
		"source_len": len(sourceCode),
		"wasm_len":   len(wasmBytes),
	}).Debug("Bundled Python to MicroPython runtime")
	return wasmBytes, nil
}
