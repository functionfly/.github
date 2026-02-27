# Migration Plan: RustPython → Perfect (FlyPy)

## Overview

This document outlines the complete migration from RustPython interpreter to Perfect (FlyPy) - a build-time Python-to-WASM compiler.

## Current State

### RustPython Implementation
- **Location**: `runtimes/local/src/python/`
- **Type**: Runtime interpreter using `rustpython_vm`
- **Execution**: Dynamic interpretation at runtime
- **Artifacts**: Python source code stored and interpreted on each execution

### Perfect/FlyPy Target
- **Type**: Build-time compiler (ahead-of-time)
- **Output**: Pre-compiled Wasm modules
- **Execution**: Native Wasm in existing Wasmtime runtime
- **Benefits**: Determinism, density, performance, security

---

## Phase 1: Build Real Python-to-WASM Compiler

### 1.1 Replace Stub Compiler
**Current**: `cmd/flypy-go/compile.go` - creates fake minimal WASM files
**Target**: Real Python-to-WASM compilation

**Implementation Options**:
1. **Pyodide-based**: Use Pyodide's Python-to-WASM compilation pipeline
2. **MicroPython**: Compile MicroPython to WASM (lighter weight)
3. **Custom AST → WASM**: Build custom compiler using Python AST

**Recommended**: Use Pyodide's compilation approach with restrictions

### 1.2 Compiler Requirements

```
Inputs:
├── Python source code (.py)
├── Function metadata (name, version, entry_point)
├── Capabilities configuration (kv, webhook, etc.)
└── Determinism settings

Outputs:
├── function.wasm (compiled WASM module)
├── manifest.json (metadata + hashes)
├── capabilities.json (permissions)
└── determinism.hash (cryptographic proof)
```

### 1.3 Compiler Features

```go
// Required compiler flags
type CompileConfig struct {
    Input             string  // Python source file
    Metadata          string  // Function metadata JSON
    Output            string  // Output directory
    Mode              string  // "deterministic" | "compatible"
    OptimizationLevel string  // "minimal" | "balanced" | "aggressive"
    
    // FlyPy-specific restrictions
    DisableDynamicImport bool
    DisableEval         bool  
    DisableExec         bool
    MaxMemoryMB         int
    AllowedImports      []string
}
```

### 1.4 Restrictions Validation

The compiler MUST enforce:
- ✗ No `eval()`, `exec()`, `compile()`
- ✗ No `__import__`, `importlib`
- ✗ No file system access
- ✗ No network access (except via capability system)
- ✗ No subprocess/threading
- ✓ Pure functions preferred
- ✓ Deterministic operations only

---

## Phase 2: Replace Runtime Execution Model

### 2.1 Current Execution Flow (RustPython)
```
User uploads Python → Stored as source → Interpreter loads → Executes line-by-line
```

### 2.2 New Execution Flow (Perfect)
```
User uploads Python → FlyPy compiles → WASM artifact stored → Wasmtime executes
```

### 2.3 Runtime Changes Required

**File**: `runtimes/local/src/engine.rs`

Current config:
```rust
pub struct Config {
    // ... existing fields
    pub python_runtime: String,  // REMOVE
    // WASM becomes primary
}
```

New config:
```rust
pub struct Config {
    // ... existing fields
    pub compiled_runtime: String,  // "perfect-1.0"
    pub manifest_path: Option<String>,
}
```

### 2.4 Update Function Execution

**File**: `runtimes/local/src/handlers.rs`

Current:
```rust
// Executes Python via RustPython interpreter
let result = python_runtime.execute(python_code, input).await;
```

New:
```rust
// Executes pre-compiled WASM (no change needed - already supports WASM!)
let result = state.engine.execute(&wasm_bytes, input, &state.config).await;
```

**Key Insight**: The existing WasmEngine already supports executing WASM modules. The only change is that WASM comes from pre-compiled artifacts instead of Rust/Go.

---

## Phase 3: Update Function Registration & Storage

### 3.1 API Changes

**Current**: Accept Python source code
```json
POST /v1/functions
{
  "name": "my-function",
  "runtime": "python",
  "source": "def handler(input): return input"
}
```

**New**: Accept compiled WASM artifacts
```json
POST /v1/functions
{
  "name": "my-function", 
  "runtime": "perfect",
  "artifacts": {
    "wasm": "base64-encoded-wasm...",
    "manifest": { ... },
    "capabilities": { ... },
    "determinism_hash": "sha256..."
  }
}
```

### 3.2 Storage Schema Changes

**Current table**: `functions`
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| name | String | Function name |
| runtime | String | "python" (RustPython) |
| source_code | Text | Python source |
| version | String | Semantic version |

**New table**: `functions`
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| name | String | Function name |
| runtime | String | "perfect" (FlyPy) |
| wasm_path | String | Path to compiled WASM |
| manifest_json | JSONB | Compilation metadata |
| capabilities_json | JSONB | Permission config |
| determinism_hash | String | Cryptographic proof |
| source_hash | String | Original source hash |

### 3.3 Registry Updates

**File**: `plans/REPO_SCAFFOLD.md`

Add new runtime type:
```json
{
  "runtimes": [
    { "id": "rust", "name": "Rust", "type": "compiled" },
    { "id": "go", "name": "Go", "type": "compiled" },
    { "id": "javascript", "name": "JavaScript", "type": "compiled" },
    { "id": "perfect", "name": "Perfect (Python)", "type": "compiled" }
  ]
}
```

---

## Phase 4: Remove RustPython Code

### 4.1 Files to Remove

```bash
# Remove Python runtime module
rm -rf runtimes/local/src/python/

# Remove RustPython dependencies from Cargo.toml
# Edit runtimes/local/Cargo.toml
```

### 4.2 Cargo.toml Changes

**Current**:
```toml
[dependencies]
rustpython-vm = { version = "0.4", features = ["freeze-stdlib"] }
```

**Remove**:
- `rustpython-vm` dependency
- All Python-specific runtime handling

### 4.3 Config Changes

**File**: `runtimes/local/src/config.rs`

```rust
// REMOVE
// pub python_runtime: String,

// Keep for Perfect
pub compiled_runtime: String,
```

### 4.4 Handler Updates

Remove Python-specific paths and handlers:
- Remove `/v1/execute/python` endpoint (if exists)
- Remove Python runtime health checks
- Remove Python-specific error handling

---

## Phase 5: Testing & Verification

### 5.1 Compiler Tests

```python
# test_compiler_restrictions.py
def test_deny_eval():
    """Must reject code with eval()"""
    code = "result = eval('1 + 1')"
    result = compile_python(code)
    assert result.error contains "eval not allowed"

def test_deny_import():
    """Must reject dynamic imports"""
    code = "import os"
    result = compile_python(code)
    assert result.error contains "import not allowed"

def test_allow_basic():
    """Must accept pure Python functions"""
    code = """
def handler(input_data):
    return input_data.upper()
"""
    result = compile_python(code)
    assert result.success
    assert result.wasm_bytes > 0
```

### 5.2 Integration Tests

```python
# test_perfect_runtime.py
def test_end_toend_execution():
    """Full flow: compile → deploy → execute"""
    # 1. Compile Python to WASM
    wasm = compile_python(python_code)
    
    # 2. Store artifact
    func_id = deploy_function(wasm)
    
    # 3. Execute via runtime
    result = execute_function(func_id, "hello")
    
    assert result == "HELLO"
```

### 5.3 Performance Benchmarks

Compare RustPython vs Perfect:
| Metric | RustPython | Perfect | Improvement |
|--------|------------|---------|-------------|
| Cold start | ~200ms | ~5ms | 40x |
| Memory usage | 128MB | 1MB | 128x |
| Execution time | Variable | Predictable | +determinism |
| Density (funcs/vm) | 1-5 | 100+ | 20x+ |

---

## Phase 6: Documentation & Migration Guide

### 6.1 User Migration Guide

For existing RustPython users:

```
┌─────────────────────────────────────────────────────────┐
│  MIGRATION: Python → Perfect                            │
├─────────────────────────────────────────────────────────┤
│  OLD (RustPython):                                      │
│  1. Write Python function                               │
│  2. Upload source code                                 │
│  3. Interpreter executes at runtime                     │
│                                                         │
│  NEW (Perfect):                                         │
│  1. Write Python function                               │
│  2. Run: flypy-go compile --input handler.py \          │
│            --metadata meta.json --output ./build        │
│  3. Upload: function.wasm + manifest                   │
│  4. Pre-compiled WASM executes                          │
└─────────────────────────────────────────────────────────┘
```

### 6.2 Breaking Changes

```markdown
## Breaking Changes in v2.0.0

### Removed
- `runtime: "python"` - RustPython interpreter deprecated
- Python source code upload endpoint

### Added
- `runtime: "perfect"` - FlyPy compiled Python
- WASM artifact upload endpoint
- Determinism hash verification

### Migration Required
All Python functions must be recompiled using Perfect CLI
before redeployment.
```

---

## Implementation Order

```
Week 1-2: Perfect Compiler Implementation
├── Set up Pyodide/MicroPython WASM build
├── Implement restriction validation
├── Generate manifest + determinism hash
└── CLI integration tests

Week 3-4: Runtime Integration
├── Update function storage schema
├── Modify API to accept WASM artifacts
├── Update registry for "perfect" runtime
└── Integration tests

Week 5: Deprecation & Cleanup
├── Remove RustPython code
├── Update Cargo.toml
├── Clean up Python handlers
└── Update configuration

Week 6: Testing & Release
├── Full integration tests
├── Performance benchmarks
├── Documentation
└── Release v2.0.0
```

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Compiler bugs | Runtime errors | Extensive test coverage |
| Breaking user workflow | Adoption friction | Migration tooling |
| Feature parity | Limited use cases | Document supported features |
| Compilation time | Slower deploys | Caching, incremental builds |

---

## Success Criteria

- [ ] Perfect compiler produces valid WASM
- [ ] All existing Python functions can be recompiled
- [ ] Execution produces identical results to RustPython
- [ ] Determinism hash verifiable
- [ ] Performance benchmarks meet targets
- [ ] Documentation complete
- [ ] No regression in Rust/Go/JS runtimes
