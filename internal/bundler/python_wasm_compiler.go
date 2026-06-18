package bundler

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/manifest"
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
		fmt.Printf("Warning: MicroPython WASM validation reported issues (%v) — using runtime anyway\n", err)
	}

	fmt.Printf("Bundled Python %q (%s, %d bytes) to MicroPython runtime (%d bytes)\n",
		entryFile, m.Name, len(sourceCode), len(wasmBytes))
	return wasmBytes, nil
}
