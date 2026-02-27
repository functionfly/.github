# MicroPython WASM Module Linking - Implementation Handoff Checklist

## Overview

This checklist provides step-by-step instructions for implementing the Full MicroPython Execution feature using WASM module linking. The design document is at [`plans/MICROPYTHON_WASM_LINKING_DESIGN.md`](plans/MICROPYTHON_WASM_LINKING_DESIGN.md).

## Pre-Implementation Setup

- [ ] **Acquire MicroPython WASM binary**
  - [ ] Build `micropython-full.wasm` (1.1MB) from source OR
  - [ ] Download pre-built WASI-compatible binary
  - [ ] Verify WASM exports: `wasm2wat micropython-full.wasm | grep "(export"`
  - [ ] Place in `runtimes/local/assets/micropython-full.wasm`

- [ ] **Verify wasmtime version**
  - [ ] Ensure wasmtime >= 19.0 with component-model support
  - [ ] Update `Cargo.toml` if needed

## Phase 1: Core Module Structure

### 1.1 Create Directory Structure

```bash
mkdir -p runtimes/local/src/micropython
```

### 1.2 Create Module Files

Create the following files:

- [ ] `runtimes/local/src/micropython/mod.rs` - Module exports
- [ ] `runtimes/local/src/micropython/loader.rs` - MicroPython WASM loader
- [ ] `runtimes/local/src/micropython/wrapper.rs` - Wrapper module generator
- [ ] `runtimes/local/src/micropython/executor.rs` - Execution orchestrator
- [ ] `runtimes/local/src/micropython/memory.rs` - Shared memory management

### 1.3 Update Main Runtime Module

- [ ] Add `pub mod micropython;` to `runtimes/local/src/main.rs`

## Phase 2: MicroPython Loader Implementation

### 2.1 Implement Loader (`loader.rs`)

Key requirements:
- [ ] Load `micropython-full.wasm` at runtime
- [ ] Cache compiled `Module` in `Arc` for reuse
- [ ] Create `create_linked_instance()` method
- [ ] Handle import resolution for `env.*` functions

Reference interface:
```rust
pub struct MicroPythonLoader {
    mp_module: Arc<Module>,
    engine: Engine,
}

impl MicroPythonLoader {
    pub fn new(engine: &Engine, wasm_bytes: &[u8]) -> Result<Self>
    pub fn create_linked_instance(&self, store: &mut Store, wrapper: &Module) -> Result<Instance>
}
```

### 2.2 Required Imports to Define

The MicroPython module expects these `env.*` imports:

| Import Name | Signature | Purpose |
|-------------|-----------|---------|
| `memory` | `(memory 32)` | Shared linear memory |
| `mp_js_init` | `(func (param i32) (result i32))` | Initialize runtime |
| `mp_js_do_exec` | `(func (param i32 i32) (result i32))` | Execute Python code |
| `mp_js_exec_async` | `(func (param i32 i32) (result i32))` | Execute async (optional) |
| `mp_js_read` | `(func (param i32 i32) (result i32))` | Read input |
| `mp_js_write` | `(func (param i32 i32))` | Write output |
| `mp_js_readline` | `(func (param i32 i32) (result i32))` | Read line |
| `mp_js_fspath` | `(func (param i32 i32) (result i32))` | FS path conversion |
| `malloc` | `(func (param i32) (result i32))` | Allocate memory |
| `free` | `(func (param i32))` | Free memory |

## Phase 3: Wrapper Module Generator

### 3.1 WAT Template

The wrapper module must export:

```wat
(module
  ;; Memory exported for sharing
  (memory (export "memory") 32 128)

  ;; Required exports
  (func (export "mp_js_init") ...)
  (func (export "mp_js_do_exec") ...)
  (func (export "malloc") ...)
  (func (export "free") ...)
  (func (export "mp_js_write") ...)
  (func (export "mp_js_read") ...)
)
```

### 3.2 Implementation Steps

- [ ] Create `WrapperGenerator` struct
- [ ] Implement WAT generation with embedded Python code
- [ ] Use `wat` crate to compile WAT → WASM
- [ ] Handle string escaping for WAT embedding

### 3.3 Memory Layout Implementation

Implement the shared memory layout from the design doc:

```
0x00000 - 0x0FFFF: Wrapper static data
0x10000 - 0x1FFFF: MicroPython stack
0x20000 - 0x9FFFF: MicroPython heap
0xA0000 - 0xDFFFF: User code buffer
0xE0000 - 0xEFFFF: Output buffer
0xF0000+: Dynamic allocation
```

## Phase 4: Executor Implementation

### 4.1 Create Executor Struct

```rust
pub struct MicroPythonExecutor {
    loader: Arc<MicroPythonLoader>,
    wrapper_gen: WrapperGenerator,
    engine: Engine,
}
```

### 4.2 Implement Execute Method

Execution flow:
1. [ ] Generate wrapper module with embedded Python code
2. [ ] Create store with host state (input data)
3. [ ] Create linked instance via loader
4. [ ] Call `mp_js_init(heap_size)`
5. [ ] Call `mp_js_do_exec(code_ptr, code_len)`
6. [ ] Read output from shared memory
7. [ ] Return result

### 4.3 Host Functions

Implement these host function handlers:

- [ ] `host_log(ptr, len)` - Structured logging
- [ ] `get_input(ptr, max_len) -> actual_len` - Provide event input
- [ ] `set_output(ptr, len)` - Capture execution result

## Phase 5: Integration with Engine

### 5.1 Extend RuntimeType

In `runtimes/local/src/engine.rs`:

- [ ] Add `MicroPythonLinked` variant to `RuntimeType` enum
- [ ] Update `display_name()` method
- [ ] Add detection logic in `detect_runtime_type()`

### 5.2 Add Execution Path

In `WasmEngine::execute()`:

- [ ] Add match arm for `RuntimeType::MicroPythonLinked`
- [ ] Create `MicroPythonExecutor` instance
- [ ] Call executor with proper error handling

### 5.3 Configuration Options

Add to `Config` struct:

```rust
pub micropython_wasm_path: String,  // Path to micropython-full.wasm
pub micropython_heap_mb: u32,       // Heap size in MB (default: 512)
pub micropython_enable: bool,       // Enable linked execution
```

## Phase 6: Error Handling

### 6.1 Error Types

Create `runtimes/local/src/micropython/errors.rs`:

```rust
pub enum MicroPythonError {
    LoadError(String),
    LinkError(String),
    ExecutionError(i32),  // Return code from mp_js_do_exec
    MemoryError,
    TimeoutError,
}
```

### 6.2 Error Propagation

- [ ] Map MicroPython error codes to meaningful messages
- [ ] Capture stderr output for debugging
- [ ] Handle panics in WASM execution

## Phase 7: Testing

### 7.1 Unit Tests

- [ ] Test wrapper generation
- [ ] Test module linking
- [ ] Test memory layout
- [ ] Test host function calls

### 7.2 Integration Tests

Create `runtimes/local/tests/micropython_tests.rs`:

- [ ] Test simple `print()` function
- [ ] Test `handler(event)` entry point
- [ ] Test JSON input/output
- [ ] Test Python imports (stdlib)
- [ ] Test error handling
- [ ] Test memory limits
- [ ] Test timeout enforcement

### 7.3 Test Python Functions

```python
# tests/fixtures/simple_handler.py
def handler(event):
    return {"message": "Hello from MicroPython!"}

# tests/fixtures/json_processing.py
import json
def handler(event):
    data = json.loads(event)
    return json.dumps({"processed": data})
```

## Phase 8: Performance Optimization

### 8.1 Caching

- [ ] Cache loaded `micropython-full.wasm` `Module`
- [ ] Cache compiled wrapper modules by code hash
- [ ] Implement instance pooling

### 8.2 Memory Optimization

- [ ] Use `memory.grow` instead of large initial memory
- [ ] Implement proper `free()` (not bump allocator)
- [ ] Monitor memory usage during execution

## Phase 9: Security

### 9.1 Sandbox Configuration

- [ ] Limit memory to configured maximum
- [ ] Set fuel limits for CPU throttling
- [ ] Disable time access if not needed
- [ ] Restrict WASI capabilities

### 9.2 Input Validation

- [ ] Validate Python code size limits
- [ ] Check for forbidden imports
- [ ] Sanitize input JSON

## Phase 10: Documentation

- [ ] Add Rust doc comments to all public APIs
- [ ] Create example Python functions
- [ ] Update `WASI_README.md` with MicroPython info
- [ ] Document configuration options

## Implementation Order

Recommended implementation sequence:

1. **Week 1**: Phases 1-2 (Structure + Loader)
2. **Week 2**: Phase 3 (Wrapper Generator)
3. **Week 3**: Phases 4-5 (Executor + Integration)
4. **Week 4**: Phases 6-7 (Errors + Testing)
5. **Week 5**: Phases 8-10 (Optimization + Security + Docs)

## Code Review Checklist

Before submitting PR:

- [ ] All tests pass (`cargo test`)
- [ ] Clippy warnings resolved (`cargo clippy`)
- [ ] Code formatted (`cargo fmt`)
- [ ] No unwrap() in production code
- [ ] Proper error handling throughout
- [ ] Documentation complete
- [ ] Benchmarks show acceptable performance
- [ ] Security review completed

## Dependencies to Add

```toml
[dependencies]
# Add to runtimes/local/Cargo.toml
wat = "1.0"           # WAT parsing for wrapper generation

[dev-dependencies]
wasmtime = { version = "19.0", features = ["component-model", "async"] }
```

## Files to Modify

| File | Changes |
|------|---------|
| `runtimes/local/Cargo.toml` | Add `wat` dependency |
| `runtimes/local/src/main.rs` | Add `micropython` module |
| `runtimes/local/src/engine.rs` | Add `MicroPythonLinked` runtime type |
| `runtimes/local/src/config.rs` | Add micropython config options |
| `runtimes/local/src/wasi.rs` | Add micropython-specific imports |

## Verification Steps

After implementation, verify:

1. [ ] `cargo build` succeeds
2. [ ] `cargo test` passes
3. [ ] Example Python function executes correctly
4. [ ] Memory usage is within limits
5. [ ] Execution time is acceptable (< 100ms cold start)
6. [ ] Errors are properly propagated

## Rollback Plan

If issues arise:

1. Feature is behind `micropython_enable` config flag
2. Can fall back to existing RustPython runtime
3. Remove `MicroPythonLinked` from runtime type detection

## Contact

Questions about this implementation:
- Architecture: See [`plans/MICROPYTHON_WASM_LINKING_DESIGN.md`](plans/MICROPYTHON_WASM_LINKING_DESIGN.md)
- Current Python runtime: See `runtimes/local/src/python/`
- WASI integration: See `runtimes/local/src/wasi.rs`
