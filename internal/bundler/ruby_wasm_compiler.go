// ruby_wasm_compiler.go implements Ruby → WASM via mruby interpreter (interpreter-embedded pattern).
package bundler

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/functionfly/functionfly/internal/manifest"
)

const (
	mrubyWASMPath    = "bundler/ruby/mruby.wasm"
	rubyCodeOffset   = 1024
	rubyOutputOffset = 4096
)

// bundleRubyForWasmRuntime bundles Ruby source for WASM execution using the mruby interpreter.
// Follows the interpreter-embedded pattern: a pre-compiled mruby.wasm receives user code
// at runtime via WASI stdin or memory injection.
func bundleRubyForWasmRuntime(m *manifest.Manifest) ([]byte, error) {
	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("ruby bundle", "failed to read entry file", err)
	}

	mrubyWasm, mrubyErr := loadMrubyRuntime()
	if mrubyErr != nil {
		fmt.Printf("Warning: mruby runtime unavailable (%v), using fallback wrapper\n", mrubyErr)
		return createRubyFallbackWasmWrapper(string(src), m)
	}

	return createRubyWasmWithCode(mrubyWasm, string(src), m)
}

// loadMrubyRuntime loads the pre-compiled mruby WASM binary.
func loadMrubyRuntime() ([]byte, error) {
	candidates := []string{
		mrubyWASMPath,
		filepath.Join("..", mrubyWASMPath),
		filepath.Join("..", "..", mrubyWASMPath),
	}

	if envPath := os.Getenv("MRUBY_WASM_PATH"); envPath != "" {
		candidates = append([]string{envPath}, candidates...)
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if len(data) >= 8 && data[0] == 0x00 && data[1] == 0x61 && data[2] == 0x73 && data[3] == 0x6D {
			return data, nil
		}
		return nil, fmt.Errorf("invalid WASM file at %s: bad magic bytes", path)
	}

	return nil, fmt.Errorf("mruby.wasm not found; place it in bundler/ruby/ or set MRUBY_WASM_PATH")
}

// createRubyWasmWithCode generates a WASM wrapper module that imports functions from the
// mruby runtime module. The wrapper embeds user Ruby source in a data segment and, at
// init time, copies it into mruby's WASM memory for execution.
//
// The generated wrapper imports from the "mruby" namespace so that the runtime
// (wasmtime) can link it with the pre-compiled mruby.wasm at instantiation time.
func createRubyWasmWithCode(mrubyWasm []byte, sourceCode string, m *manifest.Manifest) ([]byte, error) {
	escaped := escapeWatString(sourceCode)
	codeLen := len(sourceCode)

	// Metadata JSON stored in a data section so the runtime can introspect the module.
	metadata := escapeWatString(fmt.Sprintf(
		`{"name":%q,"runtime":"mruby","runtime_version":"3.2.0","version":%q,"entry_point":"handler","embedded_code_size":%d}`,
		m.Name, m.Version, codeLen))

	// The wrapper module:
	//   - imports mruby_init/exec/cleanup + malloc from the "mruby" module
	//   - embeds the user's Ruby source at rubyCodeOffset
	//   - embeds metadata at metadataOffset
	//   - exports init/__execute/alloc matching the FunctionFly WASM interface
	//
	// At runtime the host (wasmtime) instantiates mruby.wasm under the "mruby"
	// namespace, then instantiates this wrapper which links against it.

	metadataOffset := rubyCodeOffset + codeLen + 256 // leave padding after code

	wat := fmt.Sprintf(`(module
  ;; ── mruby runtime imports ──────────────────────────────────────────────
  (import "mruby" "mruby_init"        (func $mruby_init   (result i32)))
  (import "mruby" "mruby_exec"        (func $mruby_exec   (param i32 i32) (result i32)))
  (import "mruby" "mruby_result_ptr"  (func $mruby_res_ptr (result i32)))
  (import "mruby" "mruby_result_len"  (func $mruby_res_len (result i32)))
  (import "mruby" "mruby_error_ptr"   (func $mruby_err_ptr (result i32)))
  (import "mruby" "mruby_error_len"   (func $mruby_err_len (result i32)))
  (import "mruby" "mruby_cleanup"     (func $mruby_cleanup))
  (import "mruby" "malloc"            (func $mruby_malloc  (param i32) (result i32)))
  (import "mruby" "free"              (func $mruby_free    (param i32)))

  ;; ── memory ─────────────────────────────────────────────────────────────
  (memory (export "memory") 4)

  ;; ── data: embedded Ruby source at offset %d ────────────────────────────
  (data (i32.const %d) "%s\00")

  ;; ── data: metadata JSON ────────────────────────────────────────────────
  (data (i32.const %d) "%s\00")

  ;; ── globals ────────────────────────────────────────────────────────────
  (global $initialized (mut i32) (i32.const 0))
  (global $code_ptr    i32         (i32.const %d))
  (global $code_len    i32         (i32.const %d))
  (global $heap_next   (mut i32)   (i32.const %d))

  ;; ── init ───────────────────────────────────────────────────────────────
  (func (export "init")
    (if (i32.eqz (global.get $initialized)) (then
      (drop (call $mruby_init))
      (global.set $initialized (i32.const 1))
    ))
  )

  ;; ── __execute(ptr, len) → result_ptr ───────────────────────────────────
  ;; Ignores host-supplied args; uses the embedded source instead.
  ;; Returns pointer to result string in mruby's memory, or 0 on error.
  (func $__execute (export "__execute") (param $ptr i32) (param $len i32) (result i32)
    ;; Ensure initialisation
    (if (i32.eqz (global.get $initialized)) (then
      (drop (call $mruby_init))
      (global.set $initialized (i32.const 1))
    ))

    ;; Execute the embedded Ruby source
    (drop (call $mruby_exec (global.get $code_ptr) (global.get $code_len)))

    ;; Return result pointer (0-length result means error)
    (if (result i32) (i32.gt_s (call $mruby_res_len) (i32.const 0))
      (then (call $mruby_res_ptr))
      (else (i32.const 0))
    )
  )

  ;; ── alloc(size) → ptr ──────────────────────────────────────────────────
  (func (export "alloc") (param $size i32) (result i32)
    (local $ret i32)
    (global.set $heap_next
      (i32.add (local.tee $ret (global.get $heap_next)) (local.get $size)))
    (local.get $ret)
  )

  ;; ── handler(ptr, len) → result_ptr ─────────────────────────────────────
  ;; Alias used by some runtimes.
  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (call $__execute (local.get $ptr) (local.get $len))
  )
)`,
		rubyCodeOffset,
		rubyCodeOffset, escaped,
		metadataOffset, metadata,
		rubyCodeOffset,
		codeLen,
		rubyOutputOffset+1024,
	)

	wasmBytes, err := watToWasm(wat)
	if err != nil {
		// If WAT compilation fails (e.g. wat2wasm not installed), fall back to the
		// simple data-carrier wrapper so the daemon can still read the source.
		fmt.Printf("Warning: mruby wrapper WAT compilation failed (%v), falling back\n", err)
		return createRubyFallbackWasmWrapper(sourceCode, m)
	}

	return wasmBytes, nil
}

// createRubyFallbackWasmWrapper creates a minimal WASM wrapper that embeds Ruby source
// for execution by the daemon runtime. The source code is stored in WASM data sections.
func createRubyFallbackWasmWrapper(sourceCode string, m *manifest.Manifest) ([]byte, error) {
	escaped := escapeWatString(sourceCode)

	wat := fmt.Sprintf(`(module
  (type $exec_t (func (param i32 i32) (result i32)))
  (type $init_t (func))
  (type $alloc_t (func (param i32) (result i32)))

  (memory (export "memory") 2)

  ;; Ruby source code stored at offset %d
  (data (i32.const %d) "%s\00")

  ;; Source code length stored at offset 0 as 4-byte LE int
  (data (i32.const 0) "%s")

  ;; Fallback metadata at offset 64
  (data (i32.const 64) "{\"runtime\":\"mruby\",\"fallback\":true}\00")

  (func (export "__execute") (param $ptr i32) (param $len i32) (result i32)
    (i32.const %d)
  )

  (func (export "init")
    ;; No-op: Ruby source is loaded at compile time
  )

  (func (export "alloc") (param $size i32) (result i32)
    ;; Simple bump allocator starting at offset %d
    (i32.const %d)
  )
)`, rubyCodeOffset, rubyCodeOffset, escaped, encodeSourceLength(len(sourceCode)),
		rubyOutputOffset, rubyOutputOffset+1024, rubyOutputOffset+1024)

	wasmBytes, err := watToWasm(wat)
	if err != nil {
		return nil, NewBundlerErrorWithCause("ruby fallback", "WAT compilation failed", err)
	}

	return wasmBytes, nil
}

// escapeWatString escapes a string for inclusion in a WAT data segment.
func escapeWatString(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		case '"':
			result = append(result, '\\', '"')
		case '\\':
			result = append(result, '\\', '\\')
		default:
			if b >= 32 && b < 127 {
				result = append(result, b)
			} else {
				result = append(result, fmt.Sprintf("\\%02x", b)...)
			}
		}
	}
	return string(result)
}

// encodeSourceLength encodes a source code length as a little-endian 4-byte hex string for WAT.
func encodeSourceLength(length int) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(length))
	result := ""
	for _, b := range buf {
		result += fmt.Sprintf("\\%02x", b)
	}
	return result
}

// watToWasm compiles WAT text format to WASM binary using wabt's wat2wasm if available.
func watToWasm(wat string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "functionfly-wat-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	watPath := filepath.Join(tmpDir, "module.wat")
	wasmPath := filepath.Join(tmpDir, "module.wasm")

	if err := os.WriteFile(watPath, []byte(wat), 0644); err != nil {
		return nil, fmt.Errorf("failed to write WAT: %w", err)
	}

	// Try wat2wasm from wabt
	wat2wasm, lookErr := findWat2Wasm()
	if lookErr != nil {
		// Fall back to returning the WAT as-is (some runtimes accept WAT)
		return nil, fmt.Errorf("wat2wasm not found: %w", lookErr)
	}

	cmd := exec.Command(wat2wasm, watPath, "-o", wasmPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("wat2wasm failed: %s: %w", string(out), err)
	}

	return os.ReadFile(wasmPath)
}

// findWat2Wasm locates the wat2wasm binary.
func findWat2Wasm() (string, error) {
	if p, err := exec.LookPath("wat2wasm"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("wat2wasm not found; install wabt (apt-get install wabt)")
}
