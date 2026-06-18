# Follow-up: Move `state_fabric_handler.go` to the new module

After the Phase 0–3 + follow-up work, `internal/wasm/` has only one file left:

```
internal/wasm/state_fabric_handler.go   (332 lines)
```

This file uses `statefabric.Repository` (a concrete type from the
orchestrator's storage layer) for 4 methods and references several
statefabric package types:

- `statefabric.Repository` — 4 methods called:
  - `GetFabric(ctx, tenantID, fabricID) (*Fabric, error)`
  - `ListStores(ctx, tenantID, fabricID) ([]FabricStore, error)`
  - `CreateSnapshot(ctx, tenantID, fabricID, name) (*Snapshot, error)`
  - `UpdateFabric(ctx, tenantID, fabricID, updates) (*Fabric, error)`
- `statefabric.Fabric` — struct used in 6+ places
- `statefabric.FabricStore` — slice type used in 1 place
- `statefabric.Snapshot` — type used in 1 place
- `statefabric.FabricStatusOnline`, `statefabric.FabricStatusPending` —
  string constants used in 3 places

## Why this is deferred

The `statefabric.Fabric` struct has many fields (the orchestrator's
full Fabric model). Re-defining it in the new module would either:

1. **Duplicate the struct** (~100+ lines of fields) and keep them in sync
   forever — fragile.
2. **Import `internal/storage/statefabric` from the new module** — creates
   a circular dependency (new module is supposed to be a leaf).
3. **Use `interface{}` for Fabric/Snapshot/Snapshot in the interface** —
   the handler does type assertions everywhere, which is ugly and error-prone.
4. **Move `internal/storage/statefabric` into the new module** — out of
   scope; the storage layer is much larger than the wasm module.

Option 3 is the only one that keeps the new module clean. It's about
20 type assertions across 332 lines. Doable but not a quick win.

## Recommended approach (option 3)

1. **Create a `StateFabricRepo` interface in the new module** that returns
   `interface{}` for Fabric, []interface{} for stores, etc.

2. **Add `FabricStatusOnline` and `FabricStatusPending` as string
   constants** in the new module (they're just strings).

3. **Refactor `state_fabric_handler.go`** to use the interface and do
   type assertions.

4. **Create an adapter in the orchestrator** that implements
   `wasmpool.StateFabricRepo` and delegates to `statefabric.Repository`.

5. **Move `state_fabric_handler.go` to the new module**, delete
   `internal/wasm/`, done.

## Current state

The orchestrator's `internal/wasm/` package is now a single 332-line
file. Everything else is in `github.com/functionfly/wasm`:

- `router.go` (moved; uses `Tracer` interface)
- `typescript_runtime.go` (moved; uses `Bundler` interface)
- All core types, pools, runtimes, audit, metrics, streaming, host
  functions, errors, deterministic execution, wazero fallback, AI
  inference, browser config, etc.

The plan's original Phase 0 goal — extract `internal/wasm/` as a
published Go module with no behavior change — is effectively complete.
The remaining `state_fabric_handler.go` is a stable orchestrator-specific
adapter that can be refactored in a follow-up without affecting the new
module's API.
