# Bundler Architectural Suggestions

Based on analysis of [`internal/bundler/bundler.go`](internal/bundler/bundler.go), here are the key issues and improvement recommendations:

---

## 1. Code Duplication (High Priority)

**Issue:** The entry file detection logic is duplicated between:
- [`bundleJavaScript()`](internal/bundler/bundler.go:28) (lines 29-49)
- [`bundleJSToWasm()`](internal/bundler/bundler.go:125) (lines 126-146)

**Recommendation:** Extract to a shared helper:
```go
func findEntryFile(runtime, preferred string, alternatives []string) (string, error)
```

---

## 2. Unused Manifest Fields (Medium Priority)

**Issue:** The [`Manifest`](internal/manifest/manifest.go:13) struct has a `Dependencies` field that is **never used** in the bundler:
```go
Dependencies  map[string]string `json:"dependencies,omitempty"`
```

**Impact:** User-defined dependencies in `functionfly.jsonc` are ignored during bundling.

**Recommendation:** 
- Add `npm install` / `pip install` step for dependencies
- Or document that dependencies are handled at runtime by the edge provider

---

## 3. Hardcoded Entry Files (Medium Priority)

**Issue:** Entry files are hardcoded rather than configurable:
- [`bundleJavaScript()`](internal/bundler/bundler.go:30): `"index.js"` / `"main.ts"`
- [`bundlePython()`](internal/bundler/bundler.go:94): `"main.py"`

**Recommendation:** Add an optional `Entry` field to the Manifest:
```go
Entry string `json:"entry,omitempty"`
```

---

## 4. Misleading Wasm Bundling (High Priority - Naming)

**Issue:** Functions named `BundleToWasm`, `bundleJSToWasm`, `bundlePythonToWasm` do **not** actually produce WebAssembly. They produce:
- JavaScript wrappers containing source code as strings
- Python wrappers containing source code as strings

**Current behavior (lines 154-221, 238-319):** The "Wasm" output is just JavaScript/Python code that *would* run in a Wasm runtime, but it's not actually compiled to Wasm.

**Recommendation:** Rename for clarity:
- `BundleToWasm` → `BundleForLocalRuntime` or `BundleToWasmCompatible`
- Add TODO comment: `// TODO: Actually compile to Wasm using QuickJS/Pyodide`

---

## 5. Working Directory Assumption (Medium Priority)

**Issue:** The bundler assumes the current working directory contains entry files:
```go
if _, err := os.Stat(entryFile); os.IsNotExist(err) {
```

**Risk:** Will fail if CLI is run from a different directory.

**Recommendation:** Accept a `workingDir` parameter or resolve from manifest path.

---

## 6. No Dependency Resolution for Python (High Priority)

**Issue:** [`bundlePython()`](internal/bundler/bundler.go:93) just reads a single file:
```go
return simpleBundle(entryFile)
```

**Impact:** No `requirements.txt` or `pyproject.toml` handling.

**Recommendation:** 
- Read `requirements.txt` or `pyproject.toml` if present
- Include dependency list in bundle metadata
- Or document Python bundling limitation

---

## 7. Silent Fallback Without Warning (Low Priority)

**Issue:** When `esbuild` is not found, it falls back silently (line 52-56):
```go
if _, err := exec.LookPath("esbuild"); err != nil {
    fmt.Println("Warning: esbuild not found, using simple bundling")
    return simpleBundle(entryFile)
}
```

**Recommendation:** 
- Return an error instead of silently falling back
- Or at minimum, add a `--strict` flag to enforce esbuild requirement

---

## 8. Missing Error Types (Low Priority)

**Issue:** All errors use `fmt.Errorf` with generic messages.

**Recommendation:** Define custom error types:
```go
var (
    ErrEntryNotFound = errors.New("entry file not found")
    ErrUnsupportedRuntime = errors.New("unsupported runtime")
    ErrEsbuildNotFound = errors.New("esbuild not installed")
)
```

---

## Summary Todo List

```markdown
[ ] Extract findEntryFile() helper to reduce duplication
[ ] Add Entry field to Manifest and use it in bundler
[ ] Handle Dependencies from manifest (npm/pip install)
[ ] Rename Wasm functions to clarify they produce wrappers, not actual Wasm
[ ] Add workingDir parameter to Bundle functions
[ ] Improve Python bundling to handle requirements.txt
[ ] Consider strict mode for esbuild requirement
[ ] Add custom error types for better error handling
```

---

## Questions for Clarification

1. **Intent:** Is the Wasm bundling meant to be actual Wasm compilation, or is the current wrapper approach intentional for the local runtime?

2. **Dependencies:** Should the bundler handle `npm install`/`pip install` during publish, or is this delegated to the edge provider?

3. **Entry point:** Would you like to add an `entry` field to the manifest, or is the current auto-detection sufficient?
