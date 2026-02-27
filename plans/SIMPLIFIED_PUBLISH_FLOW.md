# Simplified Publish & Execute Flow - Architectural Plan

## Problem Statement

The current system has two main issues:

1. **Slow Publish**: Functions are bundled to WASM during publish, which is slow because:
   - FlyPy compilation is expensive (requires Rust toolchain)
   - Verification runs synchronously during publish
   - WASM is generated even for functions that may never be executed

2. **Execute Problems**: 
   - FlyPy execution path has separate code paths
   - Runtime restarts for each function execution
   - No WASM caching between executions

## Proposed Solution

### 1. Simplified Publish Flow (Lazy Bundling)

```
┌─────────────────────────────────────────────────────────────────┐
│                        CURRENT FLOW                              │
├─────────────────────────────────────────────────────────────────┤
│  Publish Request                                                 │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌──────────────────┐    │
│  │   Validate  │───▶│  Compile/   │───▶│   Verify &      │    │
│  │   Manifest  │    │   Bundle    │    │   Store         │    │
│  └─────────────┘    └─────────────┘    └──────────────────┘    │
│                          │                                       │
│                     (SLOW)                                       │
│                     - FlyPy compile                             │
│                     - WASM bundle                               │
│                     - Verification sync                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        NEW FLOW                                  │
├─────────────────────────────────────────────────────────────────┤
│  Publish Request                                                 │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌──────────────────┐    │
│  │   Validate  │───▶│   Store     │───▶│   Return OK      │    │
│  │   Manifest  │    │   Source    │    │   (async verify) │    │
│  └─────────────┘    └─────────────┘    └──────────────────┘    │
│       │                 │                                         │
│       │            (FAST)                                        │
│       │            - No bundling                                 │
│       │            - No FlyPy                                    │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────┐           │
│  │  Background: lazy verify & optional bundle       │           │
│  │  (triggered on first execution, not publish)    │           │
│  └─────────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Disable FlyPy - Use Only MicroPython

Current supported Python runtimes:
- `python3.11` / `python3.12` → bundles to WASM with MicroPython
- `flypy-deterministic` → compiles to WASM via FlyPy (REMOVE)

New Python runtime approach:
- All Python functions use MicroPython runtime
- Remove FlyPy from supported runtimes
- Simplify bundler to always use MicroPython

### 3. Lazy Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     EXECUTION FLOW                               │
├─────────────────────────────────────────────────────────────────┤
│  Execute Request                                                │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │  Check if WASM cached in Redis                       │       │
│  └─────────────────────────────────────────────────────┘       │
│       │                                                         │
│       ▼                                                         │
│  ┌────────────────────┐    ┌────────────────────────────┐     │
│  │  WASM Cached?      │───▶│  YES: Execute cached       │     │
│  └────────────────────┘    └────────────────────────────┘     │
│       │ NO                                                        │
│       ▼                                                         │
│  ┌────────────────────┐    ┌────────────────────────────┐     │
│  │  Bundle to WASM    │───▶│  Cache & Execute           │     │
│  │  (MicroPython)     │    │  (then return result)     │     │
│  └────────────────────┘    └────────────────────────────┘     │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │  First time: also verify function (lazy verify)    │       │
│  └─────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

## Key Changes

### Files to Modify

1. **[`internal/api/handlers/registry/publish.go`](internal/api/handlers/registry/publish.go)**
   - Remove WASM bundling during publish
   - Store only source code
   - Return immediately after storing
   - Make verification async/optional

2. **[`internal/api/handlers/registry/execution/handlers.go`](internal/api/handlers/registry/execution/handlers.go)**
   - Add WASM caching check before execution
   - Bundle on-demand if not cached
   - Add lazy verification on first execution

3. **[`internal/bundler/wasm_bundler.go`](internal/bundler/wasm_bundler.go)**
   - Simplify to always use MicroPython for Python
   - Remove FlyPy fallback

4. **[`internal/plans/limits.go`](internal/plans/limits.go)**
   - Remove `RuntimeFlyPy` constant
   - Update `IsRuntimeAllowedForPlan()` 

5. **[`internal/functionregistry/types.go`](internal/functionregistry/types.go)**
   - Remove FlyPy runtime references

6. **[`internal/api/handlers/registry/execution/execution.go`](internal/api/handlers/registry/execution/execution.go)**
   - Remove `executeFlyPy()` function and its call
   - Simplify to single execution path

7. **[`internal/api/handlers/registry/execution/sandbox.go`](internal/api/handlers/registry/execution/sandbox.go)**
   - Add WASM caching layer
   - Reuse compiled WASM across executions

## Implementation Steps

### Phase 1: Disable FlyPy & Simplify Publish
1. Remove FlyPy runtime constant
2. Update bundler to not use FlyPy
3. Modify publish handler to not bundle during publish
4. Store source code only

### Phase 2: Lazy Execution
1. Add WASM cache (Redis-based)
2. Modify execution to bundle-on-demand
3. Add lazy verification

### Phase 3: Optimization
1. Add WASM caching in sandbox executor
2. Optimize runtime reuse
3. Add metrics for monitoring

## Backward Compatibility

- Existing published functions with FlyPy runtime will continue to work but should be re-published with MicroPython
- Pre-compiled WASM binaries (via `wasm` runtime) continue to work unchanged

## Security Considerations

- Lazy verification means some invalid functions may execute once before being blocked
- Mitigation: Aggressive rate limiting on first execution for new functions
- Consider: Async verification queue with publish-time basics + execute-time deep verification
