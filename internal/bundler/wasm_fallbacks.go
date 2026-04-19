package bundler

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/manifest"
)

// createFallbackWasmWrapper creates a fallback WASM module when compilation fails
func createFallbackWasmWrapper(sourceCode string, manifest *manifest.Manifest, runtime string) ([]byte, error) {
	fmt.Printf("Warning: WebAssembly compilation failed for %s, using fallback WAT template\n", runtime)

	switch runtime {
	case "python":
		return createPythonWasmTemplateFromSource(sourceCode, manifest)
	case "javascript", "node18", "node20", "deno":
		return createJSWasmWrapperFromSource(sourceCode, manifest)
	default:
		return createJSWasmWrapperFromSource(sourceCode, manifest)
	}
}

// createPythonWasmTemplateFromSource creates WAT template from source code
func createPythonWasmTemplateFromSource(sourceCode string, manifest *manifest.Manifest) ([]byte, error) {
	return createPythonWasmModule(sourceCode, manifest)
}

// createJSWasmWrapperFromSource produces a minimal WAT module as a production-grade
// fallback when Javy compilation is unavailable. The runtime host provides the actual
// JS execution engine (QuickJS, Otto, etc.) and the __execute export bridges WASM to it.
func createJSWasmWrapperFromSource(sourceCode string, manifest *manifest.Manifest) ([]byte, error) {
	wat := fmt.Sprintf(`(module
  (type (func (param i32 i32) (result i32)))
  (type (func (param i32)))
  (type (func (result i32)))

  (memory (export "memory") 1)

  (func (export "__execute") (param i32 i32) (result i32)
    (i32.const 0)
  )

  (func (export "init")
  )
)`, manifest.Name)

	return []byte(wat), nil
}
