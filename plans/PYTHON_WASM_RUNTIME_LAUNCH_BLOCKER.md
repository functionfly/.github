# Python WASM Runtime — Launch Blocker Resolution Plan

> **Goal:** Obtain a working Python WASM runtime for FunctionFly launch.
> **Date:** 2026-05-09
> **Status:** Critical blocker — immediate action required

---

## Context

The SAR daemon (`runtimes/local`) has a `WasmCellExecutor` that can run WASI modules. The config expects `cpython.wasm` at `./runtimes/cpython.wasm`, but that file does not exist. The `RuntimeType::PythonWasm` path in `runtimes/local/src/engine/mod.rs` is also broken: it passes user code as a JSON stdin blob (`{"__python_source__":...}`) that CPython-WASI does not understand, and it never preopens the stdlib directory.

The current fallback — `RuntimeType::Python` (RustPython) — is initialized with `Interpreter::without_stdlib()`, so it lacks `json`, `os`, `sys`, `math`, `re`, `collections`, and other modules required by real user functions.

**Strategy:** Dual-track unblock.

---

## Track 1 — Immediate Unblocker (Hours)

Enable RustPython stdlib so the default fallback path actually works for real functions.

- [ ] **1.1** Add `rustpython-stdlib = "0.5.0"` to `runtimes/local/Cargo.toml`
- [ ] **1.2** Replace `Interpreter::without_stdlib(Default::default())` with `Interpreter::with_init(...)` in `runtimes/local/src/python/runtime.rs` (line ~91)
- [ ] **1.3** Verify `rustpython-stdlib` provides: `json`, `xml`, `os`, `sys`, `math`, `re`, `collections`, `datetime`, `typing`
- [ ] **1.4** Run `cargo test -p functionfly-local` to confirm no compilation or runtime regressions
- [ ] **1.5** Execute a real Python function through the HTTP daemon (`/execute`) that uses `import json` and `import os`
- [ ] **1.6** Benchmark cold-start latency and memory usage vs. the no-stdlib baseline
- [ ] **1.7** Document which stdlib modules are available in the RustPython path (update `docs/RUNTIME_SPEC.md` or equivalent)

---

## Track 2 — Proper CPython-WASI Integration (1–2 Weeks)

Fix the CPython-WASI path so it becomes the primary runtime for Free/Pro tiers.

### 2a. Binary & Stdlib Distribution

- [ ] **2.1** Create `scripts/setup-cpython-wasi.sh` to download and extract the official CPython WASI build:
  ```bash
  VERSION="3.13.0"
  URL="https://www.python.org/ftp/python/${VERSION}/python-${VERSION}-wasm32-wasi.tar.gz"
  mkdir -p runtimes/cpython-wasi
  curl -L "$URL" | tar -xz -C runtimes/cpython-wasi --strip-components=1
  ```
- [ ] **2.2** Add `runtimes/cpython-wasi/` to `.gitignore` (binary is ~8 MB + stdlib)
- [ ] **2.3** Add `cpython_stdlib_path: String` to `Config` in `runtimes/local/src/config.rs` with default `"./runtimes/cpython-wasi/lib"`
- [ ] **2.4** Update `cpython_wasm_path` default to `"./runtimes/cpython-wasi/python.wasm"`
- [ ] **2.5** Add a startup check: if `use_cpython_wasm: true` but `cpython_wasm_path` is missing, emit a clear error with a link to `setup-cpython-wasi.sh`

### 2b. Engine Integration Fixes

- [ ] **2.6** Rewrite the `RuntimeType::PythonWasm` match arm in `runtimes/local/src/engine/mod.rs` (lines 280–342)
  - Stop injecting JSON via stdin; CPython-WASI does not parse it
  - Write user code to a temp `.py` file inside a temp directory
  - Preopen that temp directory as `/tmp` in the WASI context
  - Preopen `cpython_stdlib_path` as `/lib` in the WASI context
  - Set `wasi_args` to `[
"python.wasm", "/tmp/handler.py"]`
  - Pass input JSON via a second env var (e.g., `FUNCTIONFLY_INPUT`) instead of stdin, or via a second preopened file `/tmp/input.json`
- [ ] **2.7** Update `WasiContext::new_with_input` or add a new `WasiContext::new_for_cpython` helper that accepts preopened stdlib and script directories
- [ ] **2.8** Ensure stdout/stderr capture pipes are correctly wired — CPython-WASI writes results to `fd_write` (stdout), not a memory return pointer
- [ ] **2.9** Test with a simple handler:
  ```python
  import json, sys
  def handler(event):
      return {"result": json.dumps(event)}
  ```
- [ ] **2.10** Test stdlib imports that RustPython lacks: `urllib`, `http.client`, `xml.etree`, `sqlite3`, `zlib`

### 2c. AOT Cache & Pooling

- [ ] **2.11** Wire the CPython-WASM binary into the existing AOT cache — it is a single immutable binary, so it should compile once and deserialize in microseconds
- [ ] **2.12** Wire CPython-WASM instances into `WasmCellPool` / `PooledWasmInstance` if the pooling model applies (CPython is stateful; pooling may require process-level reuse rather than module-level reuse)
- [ ] **2.13** Decide whether to pool CPython-WASM instances or spawn fresh ones per request (trade-off: memory vs. cold-start)

### 2d. Fallback & Tier Logic

- [ ] **2.14** Update `detect_runtime_type` in `engine/mod.rs` so that:
  - Free/Pro tier with `use_cpython_wasm: true` → `RuntimeType::PythonWasm`
  - UltraLow tier or `use_cpython_wasm: false` → `RuntimeType::Python` (RustPython)
- [ ] **2.15** Ensure graceful fallback: if CPython-WASI binary is missing or execution fails with a stdlib-not-found error, fall back to RustPython with a logged warning

---

## Track 3 — Post-Launch Optimization (Optional, 2–4 Weeks)

- [ ] **3.1** Evaluate MicroPython-WASI for UltraLow tier (~300 KB, fast cold start, limited stdlib)
- [ ] **3.2** Build a true WASI port of MicroPython from source (remove Emscripten JS dependencies)
- [ ] **3.3** Add `RuntimeType::MicroPython` variant if MicroPython proves viable for json-only functions
- [ ] **3.4** Bundle frozen stdlib modules into `micropython-full.wasm` to reduce import overhead

---

## Files to Modify

| File | Purpose |
|------|---------|
| `runtimes/local/Cargo.toml` | Add `rustpython-stdlib` dependency |
| `runtimes/local/src/python/runtime.rs` | Enable stdlib in RustPython interpreter |
| `runtimes/local/src/config.rs` | Add `cpython_stdlib_path`, update defaults |
| `runtimes/local/src/engine/mod.rs` | Fix CPython-WASI execution path |
| `runtimes/local/src/wasi.rs` | Add helper for CPython directory preopening |
| `scripts/setup-cpython-wasi.sh` | Download/untar official CPython WASI build |
| `docs/RUNTIME_SPEC.md` (or equivalent) | Document which stdlib modules are available per runtime |

---

## Acceptance Criteria

- [ ] A Python function using `import json` and `import os` executes successfully via the RustPython fallback path (Track 1)
- [ ] A Python function using `import urllib.request` executes successfully via the CPython-WASI path (Track 2)
- [ ] Cold-start latency for CPython-WASI is < 100 ms on a warm node (AOT cache hit)
- [ ] The runtime gracefully falls back to RustPython if CPython-WASI binary is missing
- [ ] No new critical panics or memory leaks in `cargo test -p functionfly-local`

---

*Plan created by FunctionFly System Architect — 2026-05-09*
