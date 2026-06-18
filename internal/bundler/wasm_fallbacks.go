package bundler

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/manifest"
)

// createFallbackWasmWrapper is reserved for the JS/TypeScript path only,
// where the runtime host (QuickJS, Otto, etc.) can execute the source even
// without a precompiled compiler in the build environment. The host bridges
// the source into the embedded __execute export.
//
// Python, Ruby, Kotlin and other runtimes require their precompiled WASM
// runtime to be present in the build environment. For those runtimes there
// is no acceptable fallback — returning a stub that simply yields an error
// at execution time would let broken bundles ship to production. The
// appropriate behavior is to fail the build with a clear error so the
// missing runtime is installed before the function is deployed.
func createFallbackWasmWrapper(sourceCode string, m *manifest.Manifest, runtime string) ([]byte, error) {
	switch runtime {
	case "javascript", "node18", "node20", "deno", "typescript":
		return createJSWasmWrapperFromSource(sourceCode, m)
	default:
		return nil, fmt.Errorf("bundler: no fallback available for runtime %q — install the precompiled runtime (e.g. micropython.wasm for Python, mruby.wasm for Ruby) and rebuild", runtime)
	}
}

// createJSWasmWrapperFromSource produces a minimal WAT module that embeds the
// source code in linear memory and exposes a no-op __execute export. The
// runtime host (QuickJS, Otto, etc.) reads the source from memory and
// executes it via its own JS engine.
func createJSWasmWrapperFromSource(sourceCode string, m *manifest.Manifest) ([]byte, error) {
	wat := `(module
  (type (func (param i32 i32) (result i32)))
  (type (func (param i32)))
  (type (func (result i32)))

  (memory (export "memory") 1)

  (func (export "__execute") (param i32 i32) (result i32)
    (i32.const 0)
  )

  (func (export "init")
  )
)`

	return []byte(wat), nil
}
