# FlyPy Implementation Roadmap

## Overview

This document breaks down the FlyPy architecture into concrete implementation tasks with dependencies and priorities.

---

## Phase 1: Core Infrastructure (Weeks 1-4)

### Task 1.1: Project Scaffolding
**Priority**: P0 | **Estimated**: 2 days

```
internal/flypy/
├── flypy.go              # Main package
├── parser/
│   ├── ast_parser.go     # Python AST parsing
│   ├── ast_visitor.go    # AST traversal
│   └── errors.go         # Compilation errors
├── restrictions/
│   ├── enforcer.go       # Subset enforcement
│   ├── allowed.go        # Whitelist of allowed features
│   └── patterns.go       # Forbidden patterns
└── ir/
    ├── node.go           # IR node definitions
    └── generator.go      # AST → IR conversion
```

**Deliverables**:
- [ ] `internal/flypy/flypy.go` - Package entry point
- [ ] `internal/flypy/parser/ast_parser.go` - Parse Python source to AST
- [ ] `internal/flypy/ir/node.go` - IR node types

### Task 1.2: Restricted Subset Enforcer
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 1.1

Implements compile-time rejection of non-deterministic code:

```go
// Check for forbidden imports
case *ast.Import:
    for _, alias := range n.Names {
        if !isAllowedModule(alias.Name) {
            errors = append(errors, CompileError{
                Type:    ForbiddenImport,
                Message: fmt.Sprintf("module '%s' not allowed in deterministic mode", alias.Name),
                Line:    n.Pos(),
            })
        }
    }

// Check for forbidden builtins
case *ast.Call:
    if ident, ok := n.Fun.(*ast.Name); ok {
        if !isAllowedBuiltin(ident.Name) {
            errors = append(errors, CompileError{
                Type:    ForbiddenBuiltin,
                Message: fmt.Sprintf("builtin '%s' not allowed", ident.Name),
                Line:    n.Pos(),
            })
        }
    }
```

**Deliverables**:
- [ ] Import whitelist enforcement
- [ ] Builtin function blacklist
- [ ] Forbidden pattern detection (eval, exec, open, etc.)
- [ ] Clear error messages with line numbers

### Task 1.3: IR Generator
**Priority**: P0 | **Estimated**: 4 days

**Blocked by**: 1.1, 1.2

Converts validated AST to deterministic IR:

```go
type Function struct {
    Name       string
    Parameters []Parameter
    Body       []IRNode
    Returns    IRNode
    Deterministic bool
    Pure       bool
}

type Operation struct {
    OpType   string    // "add", "multiply", "dict_get"
    Operands []IRNode
    Type     IRType    // "int", "float", "dict", "list"
}
```

**Deliverables**:
- [ ] AST → IR conversion for basic types
- [ ] Control flow IR generation (if, for, while)
- [ ] Function call IR generation
- [ ] Type inference

---

## Phase 2: Determinism Verification (Weeks 5-6)

### Task 2.1: Determinism Verifier
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 1.3

Validates that IR operations are truly deterministic:

```go
type DeterminismVerifier struct {
    IR     *ir.Module
    Errors []DeterminismError
}

func (v *DeterminismVerifier) Verify() error {
    for _, fn := range v.IR.Functions {
        if err := v.verifyFunction(fn); err != nil {
            return err
        }
    }
    return nil
}

func (v *DeterminismVerifier) verifyFunction(fn *ir.Function) error {
    for _, op := range fn.Body {
        if !op.Deterministic() {
            return DeterminismError{
                Function: fn.Name,
                Operation: op.Type(),
                Message:   fmt.Sprintf("operation %s is non-deterministic", op.Type()),
            }
        }
    }
    return nil
}
```

**Checks**:
- [ ] No hash() on mutable types
- [ ] No random operations
- [ ] No time-dependent operations
- [ ] Floating-point determinism

### Task 2.2: Side Effect Analyzer
**Priority**: P1 | **Estimated**: 2 days

**Blocked by**: 2.1

Detects and reports side effects:

```go
type SideEffectAnalyzer struct {
    IR      *ir.Module
    Effects []SideEffect
}

func (a *SideEffectAnalyzer) Analyze() []SideEffect {
    // Track mutations to:
    // - Global variables
    // - Closure variables
    // - Input parameters (if mutable)
    // - External state (files, network)
}
```

**Deliverables**:
- [ ] Side effect detection
- [ ] Effect classification (none, network, external_state)
- [ ] Idempotency verification

---

## Phase 3: Rust Backend & Wasm (Weeks 7-10)

### Task 3.1: Rust Backend Emitter
**Priority**: P0 | **Estimated**: 4 days

**Blocked by**: 2.1

Generates Rust code from IR:

```rust
// Generated output example
use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Input {
    pub value: i64,
}

#[derive(Serialize, Deserialize)]
pub struct Output {
    pub output: i64,
}

#[wasm_bindgen]
pub fn execute(input_ptr: i32, input_len: i32) -> i32 {
    let input = parse_input(input_ptr, input_len);
    let result = input.value * 2;
    let output = Output { output: result };
    serialize_output(&output)
}
```

**Deliverables**:
- [ ] IR → Rust code generation
- [ ] Wasm-bindgen interface
- [ ] JSON serialization helpers
- [ ] Error handling

### Task 3.2: Wasm Compilation Pipeline
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 3.1

Compiles Rust → Wasm:

```go
type WasmCompiler struct {
    wasmPackPath string
    outDir       string
}

func (c *WasmCompiler) Compile(rustSource string) ([]byte, error) {
    // 1. Create temp project
    // 2. Write Cargo.toml
    // 3. Write lib.rs
    // 4. Run wasm-pack build --target web --release
    // 5. Read and return .wasm file
    // 6. Cleanup
}
```

**Deliverables**:
- [ ] Cargo.toml generation
- [ ] wasm-pack invocation
- [ ] Output validation
- [ ] Cleanup handling

### Task 3.3: Artifact Bundle Builder
**Priority**: P1 | **Estimated**: 2 days

**Blocked by**: 3.2

Creates the signed artifact bundle:

```go
type ArtifactBuilder struct {
    signer crypto.Signer
}

type Artifact struct {
    WasmModule     []byte
    Manifest       Manifest
    CapabilityMap  CapabilityMap
    DeterminismHash string
    Signature      []byte
}

func (b *ArtifactBuilder) Build(irModule *ir.Module, wasmBytes []byte) (*Artifact, error) {
    manifest := GenerateManifest(irModule)
    capMap := GenerateCapabilityMap(irModule)
    detHash := ComputeDeterminismHash(irModule)
    sig, _ := b.signer.Sign([]byte(detHash))
    
    return &Artifact{
        WasmModule:     wasmBytes,
        Manifest:      manifest,
        CapabilityMap: capMap,
        DeterminismHash: detHash,
        Signature:     sig,
    }, nil
}
```

**Deliverables**:
- [ ] Manifest generation (JSON schema)
- [ ] Capability map generation
- [ ] Determinism hash computation
- [ ] Ed25519 signing

---

## Phase 4: CLI & SDK (Weeks 11-14)

### Task 4.1: FlyPy CLI
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 3.3

```bash
$ flypy --help
FlyPy - Deterministic Python Compiler

Usage:
  flypy build        Compile Python to deterministic Wasm
  flypy deploy       Deploy to FunctionFly registry
  flypy local        Run locally for testing
  flypy verify       Verify determinism of existing artifact

Flags:
  --mode string      deterministic|compatible (default: deterministic)
  --output string    Output directory (default: ./dist)
  --verbose          Verbose output
```

**Deliverables**:
- [ ] CLI commands (build, deploy, local, verify)
- [ ] Configuration file support (flypy.yaml)
- [ ] Progress output

### Task 4.2: Python SDK (Decorator API)
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 4.1

```python
import flypy

@flypy.function(
    name="process-order",
    deterministic=True,
    cache_ttl=3600
)
def handler(event):
    total = sum(item["price"] * item["quantity"] for item in    return {" event["items"])
total": total}
```

**Deliverables**:
- [ ] `@flypy.function` decorator
- [ ] `@flypy.input_schema` decorator
- [ ] `@flypy.output_schema` decorator
- [ ] Schema inference from type hints

### Task 4.3: Local Runtime
**Priority**: P1 | **Estimated**: 2 days

**Blocked by**: 4.2

Test functions locally before deployment:

```bash
$ flypy local
✓ Loaded function from ./handler.py
✓ Compiled to deterministic Wasm
✓ Verified determinism

Running on http://localhost:8080
POST / - Execute function
GET  /health - Health check
```

**Deliverables**:
- [ ] Local HTTP server
- [ ] Wasmtime integration
- [ ] Hot reload during development

---

## Phase 5: Registry Integration (Weeks 15-18)

### Task 5.1: Artifact Storage Service
**Priority**: P0 | **Estimated**: 3 days

**Blocked by**: 3.3

```go
type ArtifactStore interface {
    Store(ctx context.Context, artifact *Artifact) error
    Get(ctx context.Context, id string) (*Artifact, error)
    List(ctx context.Context, fn filters.ArtifactFilters) ([]*Artifact, error)
    Delete(ctx context.Context, id string) error
}
```

**Deliverables**:
- [ ] Storage interface
- [ ] Database schema
- [ ] S3/blob storage integration

### Task 5.2: Execution Engine Integration
**Priority**: P0 | **Estimated**: 4 days

**Blocked by**: 5.1

Modify registry execution to handle FlyPy artifacts:

```go
func (e *Executor) ExecuteFlyPy(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
    // 1. Load artifact from storage
    artifact, err := e.store.Get(req.FunctionID)
    
    // 2. Verify signature
    if !verifySignature(artifact) {
        return nil, ErrInvalidSignature
    }
    
    // 3. Verify determinism hash
    if !verifyDeterminismHash(artifact) {
        return nil, ErrDeterminismViolation
    }
    
    // 4. Instantiate Wasm
    instance, err := e.runtime.Instantiate(artifact.WasmModule)
    
    // 5. Execute
    result, err := instance.Execute(req.Input)
    
    // 6. Return with replay token
    return &ExecutionResult{
        Output:      result,
        ReplayToken: generateReplayToken(artifact, req.Input, result),
    }, nil
}
```

**Deliverables**:
- [ ] FlyPy runtime in registry executor
- [ ] Signature verification
- [ ] Determinism hash verification
- [ ] Replay token generation

### Task 5.3: Replay Verification
**Priority**: P1 | **Estimated**: 3 days

**Blocked by**: 5.2

```go
type ReplayVerifier struct {
    store ArtifactStore
}

func (v *ReplayVerifier) Verify(ctx context.Context, token ReplayToken) error {
    // 1. Get artifact from token
    artifact, _ := v.store.Get(token.FunctionID)
    
    // 2. Re-execute with same input
    result := executeDeterministic(artifact, token.Input)
    
    // 3. Compare outputs
    if !bytes.Equal(result.Output, token.Output) {
        return ErrReplayMismatch
    }
    
    return nil
}
```

**Deliverables**:
- [ ] Replay endpoint
- [ ] Output comparison
- [ ] Audit logging

---

## Phase 6: Production Hardening (Weeks 19-22)

### Task 6.1: Error Handling & Debugging
**Priority**: P1 | **Estimated**: 2 days

**Blocked by**: 5.2

- [ ] Source map generation
- [ ] Stack trace mapping
- [ ] Debug symbols in Wasm

### Task 6.2: Performance Optimization
**Priority**: P1 | **Estimated**: 3 days

**Blocked by**: 5.2

- [ ] Bundle size optimization (<50KB target)
- [ ] Cold start <5ms target
- [ ] Warm execution <1ms target

### Task 6.3: Documentation & Examples
**Priority**: P2 | **Estimated**: 2 days

**Blocked by**: All

- [ ] API documentation
- [ ] Tutorial examples
- [ ] Migration guide from standard Python

---

## Implementation Dependencies

```
Task Dependency Graph:

1.1 Project Scaffolding
    ↓
1.2 Restricted Subset Enforcer ← 1.1
    ↓
1.3 IR Generator ← 1.1, 1.2
    ↓
2.1 Determinism Verifier ← 1.3
    ↓
2.2 Side Effect Analyzer ← 2.1
    ↓
3.1 Rust Backend Emitter ← 2.1
    ↓
3.2 Wasm Compilation ← 3.1
    ↓
3.3 Artifact Builder ← 3.2
    ↓
4.1 CLI ← 3.3
    ↓
4.2 Python SDK ← 4.1
    ↓
4.3 Local Runtime ← 4.2
    ↓
5.1 Artifact Storage ← 3.3
    ↓
5.2 Execution Integration ← 5.1, 4.3
    ↓
5.3 Replay Verification ← 5.2
```

---

## Priority Matrix

| Priority | Tasks | Timeline |
|----------|-------|----------|
| P0 (Blocker) | 1.1, 1.2, 1.3, 2.1, 3.1, 3.2, 4.1, 4.2, 5.1, 5.2 | Weeks 1-18 |
| P1 (Important) | 2.2, 3.3, 4.3, 5.3, 6.1, 6.2 | Weeks 6-20 |
| P2 (Nice to have) | 6.3 | Weeks 19-22 |

---

## Parallelization Opportunities

Some tasks can run in parallel:

1. **Week 5-6**: Tasks 2.1 and 3.1 can start in parallel after 1.3
2. **Week 11**: Tasks 4.1 and 5.1 can progress in parallel after 3.3
3. **Week 15**: Tasks 5.1 and 5.2 can run alongside 4.3

---

## Estimated Total Effort

| Phase | Tasks | Days |
|-------|-------|------|
| Phase 1 | 3 | 9 |
| Phase 2 | 2 | 5 |
| Phase 3 | 3 | 9 |
| Phase 4 | 3 | 8 |
| Phase 5 | 3 | 10 |
| Phase 6 | 3 | 7 |
| **Total** | **17** | **48 days** |

---

## Next Steps

To begin implementation:

1. **Create the project structure** - Run `mkdir -p internal/flypy/{parser,restrictions,ir,verifier,backend,compiler,artifact,sdk}`
2. **Initialize Go module** - Add `github.com/functionfly/functionfly/internal/flypy` to go.mod
3. **Start with Task 1.1** - Implement AST parser

Ready to proceed with any specific task?
