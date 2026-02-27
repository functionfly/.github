# FlyPy Tech Stack Recommendations

## Current Dependencies (Leverage Existing)

| Package | Purpose | Status |
|---------|---------|--------|
| `spf13/cobra` | CLI framework | ✅ Already in use |
| `golang.org/x/crypto/ed25519` | Artifact signing | ✅ Already in use |
| `github.com/bytecodealliance/wasmtime-go` | Wasm runtime | ✅ Already available |
| `gorm.io/gorm` | Database | ✅ Already in use |
| `github.com/redis/go-redis/v9` | Caching | ✅ Already in use |

---

## Recommended New Dependencies

### 1. Python Parsing & AST

```go
// Option A: Pure Go Python parser (recommended)
import (
    "github.com/go-python/parser"  // Python AST parser in Go
)

// Option B: Use Python subprocess with AST extraction
// (simpler but requires Python installed)
```

**Recommendation**: Start with **Option B** (subprocess) for faster initial development, migrate to Option A later.

**Why**: Building a full Python parser in Go is complex. Using Python's native `ast` module via subprocess is pragmatic and reliable.

### 2. Wasm Compilation

```go
import (
    "github.com/bytecodealliance/wasmtime-go/v19"  // Already present
    // or
    "github.com/wa-lang/wa"  // WA language - alternative Wasm backend
)
```

**Recommendation**: Use **wasm-pack + cargo** for compilation (external process).

**Why**: The Rust backend approach generates Rust code that gets compiled with wasm-pack. This is more maintainable than embedding a Go Wasm compiler.

### 3. Schema Validation

```go
import (
    "github.com/invopop/jsonschema"  // JSON Schema generation
    "github.com/xeipuuv/gojsonschema" // JSON Schema validation
)
```

**Recommendation**: Use **gojsonschema** for validation.

**Why**: Need to validate input/output schemas and ensure compiled functions match declared interfaces.

### 4. Templating

```go
import (
    "text/template"  // Standard library - sufficient
    // or
    "github.com/yuin/goldmark"  // For documentation generation
)
```

**Recommendation**: Use **standard library `text/template`**.

**Why**: Simple and sufficient for Rust code generation templates.

### 5. Type System Enhancement

```go
import (
    "github.com/wI2L/go-pkgd"  // Type reflection utilities
)
```

**Recommendation**: Skip for now, use standard reflection.

---

## External Tools Required

### 1. Rust & wasm-pack

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install wasm-pack
cargo install wasm-pack
```

**Purpose**: Compile generated Rust → Wasm

### 2. Python 3.12+ (Development)

```bash
# For subprocess-based AST parsing
# No additional dependencies needed - uses stdlib
```

---

## Project Structure

```
internal/flypy/
├── flypy.go                 # Main entry point
├── config.go                # Configuration
├── compiler.go              # Main compiler orchestration
│
├── parser/                  # Python source parsing
│   ├── ast_parser.go       # Uses Python subprocess for AST
│   ├── converter.go        # Python AST → Go AST
│   └── errors.go           # Parse error types
│
├── restrictions/            # Subset enforcement
│   ├── enforcer.go         # Main restriction checker
│   ├── allowed_modules.go  # Whitelist of imports
│   ├── forbidden.go        # List of disallowed constructs
│   └── errors.go           # Restriction error types
│
├── ir/                      # Intermediate Representation
│   ├── node.go             # IR node definitions
│   ├── module.go           # Module (collection of functions)
│   ├── generator.go        # AST → IR conversion
│   └── canonicalizer.go    # IR normalization
│
├── verifier/               # Determinism verification
│   ├── determinism.go      # Core determinism checks
│   ├── side_effects.go     # Side effect analysis
│   ├── capability.go       # Capability requirements
│   └── errors.go           # Verification error types
│
├── backend/                # Code generation
│   ├── rust_emitter.go     # IR → Rust generation
│   ├── templates.go        # Rust code templates
│   ├── wasm_interface.go   # Wasm-bindgen interface
│   └── serializer.go       # JSON serialization helpers
│
├── compiler/               # Wasm compilation
│   ├── wasm_compiler.go    # Rust → Wasm pipeline
│   ├── wasm_pack.go        # wasm-pack invocation
│   └── validator.go        # Output Wasm validation
│
├── artifact/               # Artifact bundle
│   ├── builder.go          # Bundle creation
│   ├── manifest.go         # Manifest generation
│   ├── capability_map.go   # Capability mapping
│   ├── signer.go           # Ed25519 signing
│   └── hasher.go           # Determinism hash
│
├── sdk/                    # Python SDK (separate module)
│   ├── __init__.py         # Main SDK
│   ├── decorators.py       # @flypy.function decorator
│   ├── schema.py           # Type hints → JSON Schema
│   └── cli.py              # CLI wrapper
│
└── cli/                    # CLI commands
    ├── cmd_build.go        # flypy build
    ├── cmd_deploy.go       # flypy deploy
    ├── cmd_local.go        # flypy local
    └── cmd_verify.go       # flypy verify
```

---

## Key Technical Decisions

### Decision 1: Python AST Extraction

**Choice**: Python subprocess → JSON

```python
# extract_ast.py (helper script)
import ast
import json
import sys

source = sys.stdin.read()
tree = ast.parse(source)
print(json.dumps(ast.dump(tree)))
```

**Rationale**: 
- Reliable (uses Python's own parser)
- Fast enough for development
- Easy to migrate later

### Decision 2: Rust Code Generation

**Choice**: Template-based generation

```go
const rustTemplate = `
use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Input {
{{range .InputFields}}
    pub {{.Name}}: {{.Type}},
{{end}}
}

#[wasm_bindgen]
pub fn execute(input_ptr: i32, input_len: i32) -> i32 {
    let input = parse_input(input_ptr, input_len);
    let result = {{.Body}};
    serialize_output(&result)
}
`
```

**Rationale**:
- Predictable output
- Easy to debug
- Full control over generated code

### Decision 3: Wasm Compilation

**Choice**: External wasm-pack process

```go
func (c *Compiler) Compile(rustSource string) ([]byte, error) {
    // 1. Create temp directory
    // 2. Write Cargo.toml
    // 3. Write src/lib.rs
    // 4. Run: wasm-pack build --target web --release
    // 5. Read pkg/*.wasm
    // 6. Cleanup
}
```

**Rationale**:
- Mature toolchain
- Optimized output
- Well-tested

---

## Dependencies to Add to go.mod

```go
require (
    // Python AST parsing (via subprocess, no new deps)
    // Schema validation
    github.com/invopop/jsonschema v0.12.0
    github.com/xeipuuv/gojsonschema v1.2.0
    
    // Additional utilities
    github.com/google/go-cmp v0.6.0  // For testing
    github.com/agnivade/levenshtein v1.1.1  // Fuzzy matching for errors
)
```

---

## Build Requirements

### Development Environment

```bash
# Required
go 1.24+
rust 1.75+
wasm-pack

# Optional but recommended
python 3.12+  # For subprocess AST extraction
```

### Runtime Environment

```bash
# For compilation (build-time only)
rust
wasm-pack

# For execution (run-time)
wasmtime  # OR
wasmer   # OR
browser (Wasm)
```

---

## Testing Strategy

```go
// Test pyramid for FlyPy

// 1. Unit Tests
func TestRestrictionEnforcer(t *testing.T) {
    // Test individual restriction rules
    testCases := []struct {
        code     string
        wantErr  bool
        errType  ErrorType
    }{
        {"import os", true, ForbiddenImport},
        {"open('file.txt')", true, ForbiddenBuiltin},
        {"x = 1 + 2", false, 0},
    }
    // ...
}

// 2. Integration Tests  
func TestCompilerPipeline(t *testing.T) {
    // Test full: Python source → Artifact
    source := `
        def handler(event):
            return {"result": event["x"] * 2}
    `
    artifact, err := Compile(source)
    require.NoError(t, err)
    require.NotNil(t, artifact.WasmModule)
}

// 3. End-to-End Tests
func TestExecution(t *testing.T) {
    // Test: Compile → Deploy → Execute
    artifact := compileTestFunction()
    result := executeWasm(artifact, `{"x": 5}`)
    assert.Equal(t, `{"result": 10}`, result)
}
```

---

## CI/CD Integration

```yaml
# .github/workflows/flypy.yml
name: FlyPy Build

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
      - name: Setup Rust
        uses: dtolnay/rust-action@stable
      - name: Install wasm-pack
        run: cargo install wasm-pack
      - name: Test
        run: go test ./internal/flypy/...
      
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build CLI
        run: go build -o flypy ./cmd/flypy
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: flypy
          path: flypy
```

---

## Summary

| Category | Recommendation | Priority |
|----------|---------------|----------|
| Python parsing | Subprocess → JSON | P0 |
| Wasm compilation | wasm-pack external | P0 |
| Schema validation | gojsonschema | P1 |
| Code generation | text/template | P0 |
| Signing | Ed25519 (existing) | P0 |

The stack is intentionally **pragmatic** - use what's proven over what's novel. The core insight remains: compile Python to deterministic Wasm, not interpret it.
