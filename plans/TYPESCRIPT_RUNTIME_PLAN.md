# TypeScript Runtime Implementation Plan

This document outlines the implementation plan for enhancing TypeScript support in FunctionFly, including adding the Bun runtime, implementing first-class type checking, and improving the overall TypeScript developer experience.

## Overview

FunctionFly currently has partial TypeScript support through esbuild transpilation, but lacks dedicated runtime support and first-class type checking. This plan addresses those gaps to provide a comprehensive TypeScript development experience.

## Current State Analysis

### Existing TypeScript Support

TypeScript is already partially supported via esbuild transpilation:

- **Transpilation**: `.ts` and `.tsx` files are transpiled to JavaScript using esbuild
- **Supported runtimes**: `node18`, `node20`, `deno` (in manifest validation)
- **Current execution flow**: `.ts` → esbuild → JavaScript → execution runtime
- **Edge targets**: Cloudflare Workers, Vercel Edge, and Deno Deploy already support TypeScript natively

### Current Implementation Files

| Component | Location | Notes |
|-----------|----------|-------|
| Manifest validation | [`internal/manifest/manifest.go`](internal/manifest/manifest.go:269) | Validates runtime strings |
| JavaScript bundler | [`internal/bundler/js_bundler.go`](internal/bundler/js_bundler.go:14) | Uses esbuild for transpilation |
| Local runtime | [`internal/localruntime/runtime.go`](internal/localruntime/runtime.go:1) | Function execution environment |

### Current Runtime Support

The manifest currently validates these runtimes:

```go
// internal/manifest/manifest.go:269-277
validRuntimes := map[string]bool{
    "node18":     true,
    "node20":     true,
    "python3.11": true,
    "deno":       true,
}
```

### Current Limitations

1. **No type checking**: esbuild only transpiles, doesn't perform type validation
2. **No Bun runtime**: Missing from valid runtimes list
3. **Limited npm support**: Dependencies are not automatically resolved
4. **QuickJS WASM sandbox**: Has limited async/await and module isolation support

---

## Implementation Plan Sections

### 1. Add Bun Runtime (Priority: High)

Bun provides significant performance advantages and native TypeScript support.

#### 1.1 Update Manifest Validation

**File**: [`internal/manifest/manifest.go`](internal/manifest/manifest.go:269)

Add "bun" to the valid runtimes map:

```go
validRuntimes := map[string]bool{
    "node18":     true,
    "node20":     true,
    "python3.11": true,
    "deno":       true,
    "bun":        true,  // NEW
}
```

Update extension validation:

```go
validExtensions := map[string][]string{
    "node18":     {".js", ".ts"},
    "node20":     {".js", ".ts"},
    "python3.11": {".py"},
    "deno":       {".js", ".ts"},
    "bun":        {".js", ".ts"},  // NEW
}
```

Update error messages and generateJSONCWithComments:

```go
// Line 142-143: Update comment
sb.WriteString(`  // Runtime: node18, node20, python3.11, python3.12, deno, bun
`)
```

#### 1.2 Create Bun Runtime Executor

Create a new directory structure for the Bun runtime:

```
internal/localruntime/bun/
├── executor.go      # Main execution logic
├── handler.go       # HTTP handler for function calls
├── sandbox.go       # Isolated execution environment
├── types.go         # Type definitions
└── test/
    └── executor_test.go
```

**Example executor implementation** ([`internal/localruntime/bun/executor.go`](internal/localruntime/bun/executor.go:1)):

```go
package bun

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
)

// Executor runs Bun functions in an isolated environment
type Executor struct {
    timeout time.Duration
    memory  int // MB
}

// NewExecutor creates a new Bun executor
func NewExecutor(timeout time.Duration, memory int) *Executor {
    return &Executor{
        timeout: timeout,
        memory:  memory,
    }
}

// Execute runs the Bun function with the given input
func (e *Executor) Execute(ctx context.Context, code string, input interface{}) (interface{}, error) {
    // Create temporary directory for function execution
    tempDir, err := e.createTempDir()
    if err != nil {
        return nil, fmt.Errorf("failed to create temp dir: %w", err)
    }
    defer e.cleanup(tempDir)

    // Write function code to temp file
    funcFile := filepath.Join(tempDir, "function.ts")
    if err := e.writeFunction(funcFile, code); err != nil {
        return nil, fmt.Errorf("failed to write function: %w", err)
    }

    // Write input to temp file
    inputFile := filepath.Join(tempDir, "input.json")
    inputBytes, _ := json.Marshal(input)
    if err := e.writeInput(inputFile, inputBytes); err != nil {
        return nil, fmt.Errorf("failed to write input: %w", err)
    }

    // Execute with Bun
    cmd := exec.CommandContext(ctx, "bun", "run", funcFile)
    cmd.Env = []string{
        "BUN_ENV=functionfly",
        "FUNCTION_INPUT_PATH=" + inputFile,
    }
    cmd.Dir = tempDir

    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("bun execution failed: %w, output: %s", err, string(output))
    }

    // Parse output
    var result interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse output: %w", err)
    }

    return result, nil
}

func (e *Executor) createTempDir() (string, error) { /* ... */ }
func (e *Executor) cleanup(dir string) { /* ... */ }
func (e *Executor) writeFunction(path, code string) error { /* ... */ }
func (e *Executor) writeInput(path string, data []byte) error { /* ... */ }
```

#### 1.3 Update Docker Compose

**File**: [`docker-compose.runtime.yml`](docker-compose.runtime.yml:1)

Add Bun runtime service:

```yaml
# Bun Runtime Service - Bun function execution
runtime-bun:
    image: functionfly/runtimes/bun:latest
    container_name: functionfly-runtime-bun
    ports:
      - "8085:8085"
    environment:
      - BUN_ENV=production
      - MAX_CONCURRENT=${MAX_CONCURRENT:-50}
      - PORT=8085
    networks:
      - functionfly-runtime
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8085/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 512M
```

#### 1.4 Bun Dockerfile

Create [`deploy/bun/Dockerfile`](deploy/bun/Dockerfile):

```dockerfile
FROM oven/bun:1.1 as base

WORKDIR /app

# Install production dependencies
FROM base AS dependencies
RUN bun install --frozen-lockfile --production

# Build runtime
FROM base AS builder
RUN bun add -d typescript @types/node

# Production image
FROM base AS runner
COPY --from=dependencies /app/node_modules ./node_modules
COPY --from=builder /app/node_modules/.bin ./node_modules/.bin

# Copy runtime code
COPY ./deploy/bun/ ./

# Expose port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8085/health || exit 1

# Run the runtime
CMD ["bun", "run", "index.ts"]
```

### 2. First-Class TypeScript Type Checking (Priority: High)

Add TypeScript compiler (tsc) integration to the bundler pipeline.

#### 2.1 Update JavaScript Bundler

**File**: [`internal/bundler/js_bundler.go`](internal/bundler/js_bundler.go:1)

Add type checking before bundling:

```go
// bundleJavaScript bundles JavaScript/TypeScript code using esbuild
// with optional type checking
func bundleJavaScript(manifest *manifest.Manifest, options *BundleOptions) ([]byte, error) {
    // Read and validate entry file
    entryFile, sourceCode, err := ReadEntryFile(manifest)
    if err != nil {
        return nil, NewBundlerErrorWithCause("javascript bundle", "failed to read entry file", err)
    }

    // Perform type checking for TypeScript files
    if options != nil && options.TypeCheck && (strings.HasSuffix(entryFile, ".ts") || strings.HasSuffix(entryFile, ".tsx")) {
        typeErrors, err := typeCheck(entryFile)
        if err != nil {
            return nil, NewTypeError("type checking failed", typeErrors)
        }
        if len(typeErrors) > 0 {
            return nil, NewTypeErrorWithDetails(typeErrors)
        }
    }

    // Check if esbuild is available
    if _, err := exec.LookPath("esbuild"); err != nil {
        fmt.Println("Warning: esbuild not found, using simple bundling")
        return sourceCode, nil
    }

    // ... rest of bundling logic
}

// BundleOptions contains options for the bundler
type BundleOptions struct {
    TypeCheck     bool
    SourceMap     bool
    Minify        bool
    Target        string
    ExternalDeps  []string
}

// typeCheck runs tsc --noEmit on the given file
func typeCheck(file string) ([]TypeError, error) {
    // Create temporary tsconfig if not present
    tsconfig, err := ensureTsconfig(file)
    if err != nil {
        return nil, err
    }
    defer os.Remove(tsconfig)

    cmd := exec.Command("tsc", "--noEmit", "--project", tsconfig)
    output, err := cmd.CombinedOutput()

    if err != nil {
        // Parse tsc output for type errors
        return parseTypeErrors(string(output)), err
    }

    return nil, nil
}
```

#### 2.2 Create Type Error Handling

**File**: [`internal/bundler/errors.go`](internal/bundler/errors.go:1) - Add TypeError type:

```go
// TypeError represents a TypeScript type checking error
type TypeError struct {
    File       string
    Line       int
    Column     int
    Message    string
    Code       string // TS error code like TS2307
}

// NewTypeError creates a new type error
func NewTypeError(message string, errors []TypeError) error {
    return &TypeErrorInfo{
        message: message,
        errors:  errors,
    }
}

// TypeErrorInfo contains detailed type error information
type TypeErrorInfo struct {
    message string
    errors  []TypeError
}

func (e *TypeErrorInfo) Error() string {
    var sb strings.Builder
    sb.WriteString(e.message + "\n")
    for _, err := range e.errors {
        sb.WriteString(fmt.Sprintf("%s:%d:%d: error %s: %s\n",
            err.File, err.Line, err.Column, err.Code, err.Message))
    }
    return sb.String()
}
```

#### 2.3 tsconfig.json Support

Add support for inheriting from project tsconfig.json:

```go
// ensureTsconfig creates or uses a tsconfig.json for type checking
func ensureTsconfig(entryFile string) (string, error) {
    // Check for existing tsconfig.json in the project
    dir := filepath.Dir(entryFile)
    for {
        tsconfig := filepath.Join(dir, "tsconfig.json")
        if _, err := os.Stat(tsconfig); err == nil {
            return tsconfig, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }

    // Create default tsconfig
    defaultConfig := `{
        "compilerOptions": {
            "target": "ES2020",
            "module": "ESNext",
            "strict": true,
            "esModuleInterop": true,
            "skipLibCheck": true,
            "forceConsistentCasingInFileNames": true,
            "moduleResolution": "bundler",
            "allowImportingTsExtensions": true,
            "noEmit": true,
            "lib": ["ES2020"]
        },
        "include": ["**/*.ts", "**/*.tsx"]
    }`

    tempFile := filepath.Join(os.TempDir(), "functionfly-tsconfig-"+fmt.Sprint(os.Getpid())+".json")
    if err := os.WriteFile(tempFile, []byte(defaultConfig), 0644); err != nil {
        return "", err
    }

    return tempFile, nil
}
```

### 3. Enhanced npm Package Support (Priority: Medium)

Implement proper dependency resolution for JavaScript/TypeScript functions.

#### 3.1 Package.json Parsing

**File**: [`internal/bundler/dependencies.go`](internal/bundler/dependencies.go:1) - Enhance existing implementation:

```go
// ParsePackageJson parses a package.json file
func ParsePackageJson(path string) (*PackageJson, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var pkg PackageJson
    if err := json.Unmarshal(data, &pkg); err != nil {
        return nil, err
    }

    return &pkg, nil
}

type PackageJson struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Dependencies map[string]string `json:"dependencies"`
    DevDeps      map[string]string `json:"devDependencies"`
    PeerDeps     map[string]string `json:"peerDependencies"`
}
```

#### 3.2 Dependency Resolution

Add npm registry integration:

```go
// Resolver handles npm package resolution
type Resolver struct {
    cache      *DependencyCache
    registry   string
    authToken  string
}

// ResolveDependencies fetches all required packages
func (r *Resolver) ResolveDependencies(deps map[string]string) (map[string][]byte, error) {
    results := make(map[string][]byte)

    for name, version := range deps {
        // Check cache first
        if cached, ok := r.cache.Get(name, version); ok {
            results[name] = cached
            continue
        }

        // Fetch from registry
        pkg, err := r.fetchPackage(name, version)
        if err != nil {
            return nil, fmt.Errorf("failed to resolve %s@%s: %w", name, version, err)
        }

        results[name] = pkg
        r.cache.Set(name, version, pkg)
    }

    return results, nil
}
```

#### 3.3 Dependency Caching

Implement Redis-based dependency caching:

```go
// DependencyCache caches resolved npm packages
type DependencyCache struct {
    redis  *redis.Client
    local  *lru.Cache
    ttl    time.Duration
}

func (c *DependencyCache) Get(name, version string) ([]byte, bool) {
    // Check local cache first
    key := fmt.Sprintf("dep:%s:%s", name, version)
    if val, ok := c.local.Get(key); ok {
        return val.([]byte), true
    }

    // Check Redis
    val, err := c.redis.Get(context.Background(), key).Bytes()
    if err == nil {
        c.local.Add(key, val)
        return val, true
    }

    return nil, false
}
```

### 4. TypeScript Developer Experience (Priority: Medium)

#### 4.1 Type Definitions for Function Context

Create TypeScript type definitions package:

**File**: [`web/dashboard/src/types/functionfly.d.ts`](web/dashboard/src/types/functionfly.d.ts)

```typescript
// FunctionFly Function Context Type Definitions

export interface Request {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: unknown;
  json: <T = unknown>() => Promise<T>;
  text: () => Promise<string>;
}

export interface Response {
  status: number;
  headers: Record<string, string>;
  body: unknown;
  send: (data: unknown) => void;
  json: (data: unknown) => void;
}

export interface Env {
  // Key-value store
  get(key: string): Promise<string | null>;
  put(key: string, value: string): Promise<void>;
  delete(key: string): Promise<void>;
  
  // Secrets (from Vault)
  secret(key: string): Promise<string>;
  
  // Cache
  cache: {
    get(key: string): Promise<string | null>;
    set(key: string, value: string, ttl?: number): Promise<void>;
    delete(key: string): Promise<void>;
  };
  
  // Database (if enabled)
  db?: {
    query<T>(sql: string, params?: unknown[]): Promise<T[]>;
    execute(sql: string, params?: unknown[]): Promise<{ rowsAffected: number }>;
  };
}

export interface Context {
  request: Request;
  response: Response;
  env: Env;
  params: Record<string, string>;
  logs: {
    debug(message: string, ...args: unknown[]): void;
    info(message: string, ...args: unknown[]): void;
    warn(message: string, ...args: unknown[]): void;
    error(message: string, ...args: unknown[]): void;
  };
}

// Handler function signature
export type Handler = (request: Request, context: Context) => Promise<Response> | Response;

// For simpler functions
export type SimpleHandler = (input: unknown, context: Context) => Promise<unknown> | unknown;
```

#### 4.2 IDE Integration

Create `.vscode/typescript.config.json` for the function project:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020"],
    "types": ["functionfly"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "moduleResolution": "bundler"
  }
}
```

### 5. Runtime Sandbox Improvements (Priority: Medium)

#### 5.1 Enhance JavaScript WASM Compiler

**File**: [`internal/bundler/js_wasm_compiler.go`](internal/bundler/js_wasm_compiler.go:1)

Improve QuickJS WASM for better TypeScript execution:

```go
// Enhanced QuickJS compilation with async support
func CompileToQuickJS(code string, opts *WASMOptions) ([]byte, error) {
    // Pre-process async/await to QuickJS-compatible promises
    processedCode := preprocessAsyncAwait(code)

    // Configure QuickJS runtime
    rt := quickjs.NewRuntime()
    defer rt.Close()

    ctx := rt.NewContext()
    defer ctx.Close()

    // Set up Promise handling
    ctx.Set("Promise", quickjs.Undefined()) // Use native Promise

    // Compile and return WASM
    return compileWASM(ctx, processedCode, opts)
}

// preprocessAsyncAwait converts async/await to Promise chains
func preprocessAsyncAwait(code string) string {
    // Use esbuild to transpile async/await for older targets
    // or rely on QuickJS's native Promise support
    return code
}
```

#### 5.2 Module Isolation

Implement proper ES module isolation:

```go
// ModuleIsolator provides sandboxed module execution
type ModuleIsolator struct {
    vm *quickjs.Runtime
}

// NewModuleIsolator creates a new isolated module environment
func NewModuleIsolator() *ModuleIsolator {
    rt := quickjs.NewRuntime()
    
    // Set up global limits
    rt.SetMemoryLimit(128 * 1024 * 1024) // 128MB
    rt.SetMaxStackSize(1024 * 1024)       // 1MB

    return &ModuleIsolator{vm: rt}
}

// ExecuteModule runs code in isolated context
func (m *ModuleIsolator) ExecuteModule(code string, moduleName string) (interface{}, error) {
    ctx := m.vm.NewContext()
    defer ctx.Close()

    // Create module scope
    module := ctx.Object()
    defer module.Free()

    // Set up exports
    exports := ctx.Object()
    module.Set("exports", exports)

    // Execute in module context
    result, err := ctx.Eval(code)
    if err != nil {
        return nil, err
    }

    return exports, nil
}
```

### 6. Testing Strategy

#### 6.1 Unit Tests for TypeScript Compilation

**File**: [`internal/bundler/js_bundler_test.go`](internal/bundler/js_bundler_test.go) - Create:

```go
func TestTypeScriptBundling(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr bool
        errType error
    }{
        {
            name: "valid typescript",
            code: `
                export function handler(input: string): string {
                    return input.toUpperCase();
                }
            `,
            wantErr: false,
        },
        {
            name: "type error",
            code: `
                export function handler(input: number): string {
                    return input.toUpperCase(); // Error: number doesn't have toUpperCase
                }
            `,
            wantErr: true,
            errType: &TypeErrorInfo{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 6.2 Integration Tests for Bun Runtime

**File**: [`internal/localruntime/bun/executor_test.go`](internal/localruntime/bun/executor_test.go):

```go
func TestBunExecutor(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Bun integration tests in short mode")
    }

    executor := NewExecutor(5*time.Second, 256)

    tests := []struct {
        name    string
        code    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        {
            name: "simple handler",
            code: `
                const input = JSON.parse(readFileSync(process.env.FUNCTION_INPUT_PATH, 'utf-8'));
                export default function() {
                    return { result: input.value * 2 };
                }
            `,
            input: map[string]int{"value": 21},
            want:  map[string]int{"result": 42},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result, err := executor.Execute(ctx, tt.code, tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

---

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)

| Task | Files | Effort |
|------|-------|--------|
| Add Bun to manifest validation | [`internal/manifest/manifest.go`](internal/manifest/manifest.go:269) | 1 day |
| Create Bun executor skeleton | [`internal/localruntime/bun/`](internal/localruntime/bun/) | 2 days |
| Add Bun to docker-compose | [`docker-compose.runtime.yml`](docker-compose.runtime.yml:1) | 1 day |
| Create Bun Dockerfile | [`deploy/bun/Dockerfile`](deploy/bun/Dockerfile) | 1 day |

**Milestone**: Users can select "bun" runtime and functions execute with Bun.

### Phase 2: Type Checking (Weeks 3-4)

| Task | Files | Effort |
|------|-------|--------|
| Integrate tsc in bundler | [`internal/bundler/js_bundler.go`](internal/bundler/js_bundler.go:1) | 2 days |
| Create type error handling | [`internal/bundler/errors.go`](internal/bundler/errors.go:1) | 1 day |
| Add tsconfig.json support | [`internal/bundler/tsconfig.go`](internal/bundler/tsconfig.go) (new) | 2 days |
| Return type errors to users | API handlers | 1 day |

**Milestone**: TypeScript type errors are caught at deploy time with clear messages.

### Phase 3: Package Support (Weeks 5-6)

| Task | Files | Effort |
|------|-------|--------|
| Enhance package.json parsing | [`internal/bundler/dependencies.go`](internal/bundler/dependencies.go:1) | 2 days |
| Implement npm registry resolver | [`internal/bundler/npm_resolver.go`](internal/bundler/npm_resolver.go) (new) | 3 days |
| Add dependency caching | Redis cache layer | 2 days |
| Handle peer dependencies | Resolution logic | 2 days |

**Milestone**: Functions can import npm packages that are automatically resolved and bundled.

### Phase 4: Polish (Weeks 7-8)

| Task | Files | Effort |
|------|-------|--------|
| Create TypeScript type definitions | [`web/dashboard/src/types/functionfly.d.ts`](web/dashboard/src/types/functionfly.d.ts) | 1 day |
| Add IDE integration docs | Docs | 1 day |
| Improve error messages | Error handlers | 2 days |
| Full test coverage | All new files | 3 days |
| Documentation | [`docs/TYPESCRIPT.md`](docs/TYPESCRIPT.md) (new) | 1 day |

**Milestone**: Complete TypeScript developer experience with types, docs, and tests.

---

## Success Metrics

- [ ] Bun runtime added to valid runtimes and functions execute correctly
- [ ] TypeScript type errors returned at deploy time with file/line/column
- [ ] npm dependencies automatically resolved and bundled
- [ ] Type definitions available for IDE autocompletion
- [ ] QuickJS sandbox handles async/await properly
- [ ] Test coverage > 80% for new code

---

## Appendix: Quick Reference

### Updated Manifest Runtime Options

```jsonc
{
  "name": "my-typescript-function",
  "version": "1.0.0",
  "runtime": "bun",  // Options: node18, node20, deno, bun
  "entry": "index.ts"
}
```

### TypeScript Function Example

```typescript
import { Request, Response, Context } from 'functionfly';

export async function handler(request: Request, context: Context): Promise<Response> {
  const data = await request.json<{ name: string }>();
  
  context.logs.info('Processing request', { name: data.name });
  
  return {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
    body: {
      message: `Hello, ${data.name}!`,
      timestamp: Date.now()
    }
  };
}
```

### Bun Advantages Summary

| Metric | Node.js 18 | Bun | Improvement |
|--------|-----------|-----|-------------|
| Cold Start | ~200ms | ~50ms | 4x faster |
| HTTP Throughput | 10k req/s | 50k req/s | 5x faster |
| TypeScript Support | Requires build | Native | Zero config |
| Package Manager | npm/yarn/pnpm | Built-in | Integrated |
