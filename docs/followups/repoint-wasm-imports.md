# Follow-up: Repoint remaining `internal/wasm` imports

Phase 0 of `.kilo/plans/externalize-wasm-pool-service.md` created
`github.com/functionfly/wasm` as a standalone module containing the 8 core
pool files plus 4 dependencies and 2 type-definition files
(`audit.go`, `runtime_type.go`). The plan called for repointing the
orchestrator's 4 import sites in Phase 0, but the repointing was deferred
because 3 files in `internal/wasm/` (`router.go`, `typescript_runtime.go`,
`state_fabric_handler.go`) have orchestrator-specific imports
(`internal/tracing`, `internal/bundler`, `internal/storage/statefabric`)
that block a clean move.

## Status

| File | Repointed? | Notes |
|------|------------|-------|
| `internal/api/handlers/functions/functions.go` | ✅ | Single use of `NewPythonRuntimeWithDebug`; repointed to `wasmpool` alias; builds clean |
| `internal/api/routes.go` | ⏳ | Uses `PerTenantPools` and `InitPoolsWithConfig` (both in new module) + `InstancePool`, `PythonRuntime`, `NewPythonRuntime` (in new module). Can repoint to `wasmpool` if the remaining `internal/wasm/` files also import `PerTenantPools` from the new module |
| `internal/api/handlers/registry/execution/wasm_integration.go` | ⏳ | Uses 15+ types. Most are in the new module (`InstancePool`, `AuditLogger`, `RuntimeType*`, `MetricsRecorder`, `StatusSuccess`, `ExecutionAudit`); 4 stay in `internal/wasm/` (`RuntimeRouter`, `DeterministicConfig`, `DefaultDeterministicConfig`, `NewDeterministicExecutor`) |
| `internal/api/handlers/registry/execution/engines.go` | ⏳ | Uses 8 types. All except `NewTypeScriptRuntime` are in the new module |

## Pre-repoint checklist

Before repointing any of the 3 remaining files, do this:

1. **Move `deterministic.go`** to the new module if it has no orchestrator-specific imports. The new module already has `audit.go` and `runtime_type.go` for similar reason.

   ```bash
   # Check imports:
   head -15 internal/wasm/deterministic.go
   # If clean (only stdlib + the wasm module), copy it to the new module.
   ```

2. **Update the remaining `internal/wasm/` files** that reference `PerTenantPools` (router.go, etc.) to import it from the new module:

   ```go
   import wasmpool "github.com/functionfly/wasm"
   // ...
   pool := wasmpool.PerTenantPools  // was: wasm.PerTenantPools
   ```

3. **Remove duplicate type definitions** from `internal/wasm/` for the
   types that are now in the new module (`InstancePool`, `PythonRuntime`,
   `MetricsRecorder`, `AuditLogger`, `RuntimeType`, `StatusSuccess`,
   `ExecutionAudit`, etc.). The `internal/wasm/` package will fail to
   compile if both definitions exist in different packages and the
   importing file references the wrong one.

   **Strategy:** delete the moved files from `internal/wasm/` after
   confirming the new module has them. The 15 files copied in Phase 0
   are:
   - `pool.go`, `runtime_pool.go`, `runtime.go`, `config.go`, `metrics.go`, `iot3_runtime.go`, `security.go`, `types.go`
   - `host_functions.go`, `micropython_host.go`, `streaming_state.go`, `streaming.go`
   - `audit.go`, `runtime_type.go` (created during Phase 0)

4. **For `router.go`, `typescript_runtime.go`, `state_fabric_handler.go`** (the 3 files that can't move cleanly), refactor to take their orchestrator-specific dependencies as interfaces:

   ```go
   // Before:
   import "github.com/functionfly/functionfly/internal/tracing"
   ctx, _ = tracing.StartSpan(ctx, ...)

   // After:
   type Tracer interface {
       StartSpan(ctx context.Context, name string) (context.Context, func())
       SetAttribute(ctx context.Context, key string, val interface{})
   }
   // Inject from the orchestrator at construction time.
   ```

   This is the largest piece of the follow-up; each file needs its own
   thin interface layer.

## Step-by-step for `routes.go`

1. Add `wasmpool "github.com/functionfly/wasm"` import (keep `wasm` for any types that stay in `internal/wasm/`).
2. Replace `wasm.PerTenantPools` → `wasmpool.PerTenantPools`.
3. Replace `wasm.InstancePool` → `wasmpool.InstancePool`.
4. Replace `wasm.NewPythonRuntime` → `wasmpool.NewPythonRuntime`.
5. Replace `wasm.InitPoolsWithConfig` → `wasmpool.InitPoolsWithConfig`.
6. Build and verify.

## Step-by-step for `wasm_integration.go`

1. Add `wasmpool` import.
2. Replace types in the new module:
   - `wasm.InstancePool` → `wasmpool.InstancePool`
   - `wasm.AuditLogger` → `wasmpool.AuditLogger`
   - `wasm.RuntimeTypeFromString` → `wasmpool.RuntimeTypeFromString`
   - `wasm.RuntimeUnknown` → `wasmpool.RuntimeUnknown`
   - `wasm.RuntimePythonWASM` → `wasmpool.RuntimePythonWASM`
   - `wasm.MetricsRecorder` → `wasmpool.MetricsRecorder`
   - `wasm.NewMetricsRecorder` → `wasmpool.NewMetricsRecorder`
   - `wasm.StatusSuccess` → `wasmpool.StatusSuccess`
   - `wasm.ExecutionAudit` → `wasmpool.ExecutionAudit`
3. Keep types in `internal/wasm/`:
   - `wasm.RuntimeRouter`
   - `wasm.DeterministicConfig`
   - `wasm.DefaultDeterministicConfig`
   - `wasm.NewDeterministicExecutor`
4. Build and verify.

## Step-by-step for `engines.go`

1. Add `wasmpool` import.
2. Replace types in the new module:
   - `wasm.HostFunctionHandler` → `wasmpool.HostFunctionHandler`
   - `wasm.PythonRuntime` → `wasmpool.PythonRuntime`
   - `wasm.NewPythonRuntime` → `wasmpool.NewPythonRuntime`
   - `wasm.NewPythonRuntimeWithConfig` → `wasmpool.NewPythonRuntimeWithConfig`
   - `wasm.NewDefaultSecurityConfig` → `wasmpool.NewDefaultSecurityConfig`
   - `wasm.InstancePool` → `wasmpool.InstancePool`
   - `wasm.PooledInstance` → `wasmpool.PooledInstance`
3. Keep types in `internal/wasm/`:
   - `wasm.NewTypeScriptRuntime`
4. Build and verify.

## Validation

After each file is repointed:

```bash
CGO_ENABLED=1 go build ./internal/...
CGO_ENABLED=1 go test -short -count=1 ./internal/api/handlers/functions/...
CGO_ENABLED=1 go test -short -count=1 ./internal/api/handlers/registry/execution/...
```

The end-to-end check:

```bash
CGO_ENABLED=1 go test -short -count=1 ./internal/...
```

No new tests should be needed; the repointing is a pure rename. If a test
fails, it's because a type moved between packages and the test's import
needs updating too.

## Rollback

If a repointing causes regressions:

```bash
git checkout -- internal/api/handlers/functions/functions.go
# or whichever file was just repointed
```

The new module's `replace` directive in `go.mod` still points to `../wasm`,
so a rollback to a working state is just a file-level revert.
