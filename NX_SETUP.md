# Nx Monorepo Setup - FunctionFly

## Overview

FunctionFly has been migrated to use Nx for improved monorepo management, build caching, and task orchestration across multiple languages (Go, Rust, TypeScript/JavaScript, Python).

## What is Nx?

Nx is a smart build system with:
- **Task caching**: Never rebuild the same code twice
- **Affected builds**: Only build what changed
- **Parallel execution**: Run tasks concurrently
- **Polyglot support**: Works with Go, Rust, Node.js, and more

## Project Structure

```
functionfly/
├── cmd/                    # Go CLI tools and services
│   ├── blog-api/          # NestJS blog API (detected ✓)
│   ├── fly/               # Main CLI (detected ✓)
│   ├── ffly/              # Function CLI (project.json exists)
│   ├── orchestrator-api/  # Main API service (project.json exists)
│   ├── health-monitor/    # Health monitoring (project.json exists)
│   └── migrate/           # Database migrations (project.json exists)
├── web/                   # Frontend applications
│   ├── dashboard/         # Main dashboard (detected ✓)
│   ├── docs/              # Documentation site (detected ✓)
│   └── admin-dashboard/   # Admin interface (detected ✓)
├── runtimes/              # Function runtimes
│   ├── local/             # Rust local runtime (detected ✓)
│   └── microvm/           # MicroVM runtime (detected ✓)
├── nx.json                # Nx configuration
├── package.json           # Root workspace config
└── Makefile               # Make + Nx integration
```

## Quick Start

### Using Nx Directly

```bash
# Show all projects
npx nx show projects

# Build a specific project
npx nx build fly-cli
npx nx build dashboard

# Build all projects
npx nx run-many --target=build --all

# Build only what changed
npx nx affected --target=build --base=main

# Test a project
npx nx test dashboard

# Run development server
npx nx dev dashboard

# Show project dependency graph
npx nx graph

# Reset Nx cache
npx nx reset
```

### Using Make (Integrated)

```bash
# Build all projects with Nx
make nx-build

# Build only affected projects
make nx-build-affected

# Test all projects
make nx-test

# Test only affected projects
make nx-test-affected

# Lint all projects
make nx-lint

# Show dependency graph
make nx-graph

# Reset Nx cache
make nx-reset

# Traditional Go-only builds still work
make build
make build-local-runtime
```

## Configured Projects

### Detected by Nx (7 projects)
1. **dashboard** - Main web dashboard (Vite + React)
2. **docs** - Documentation site (Vite + React)
3. **admin-dashboard** - Admin interface (Vite + React)
4. **blog-api** - Blog API (NestJS)
5. **fly-cli** - Main CLI tool (Go)
6. **runtime-local** - Local function runtime (Rust)
7. **runtime-microvm** - MicroVM runtime (Rust)

### Additional project.json Files Created
- **orchestrator-api** - Main orchestrator service (Go)
- **health-monitor** - System health monitoring (Go)
- **ffly-cli** - Function development CLI (Go)
- **migrate** - Database migration tool (Go)

Note: Some projects may not appear in `nx show projects` but have project.json files configured for future use.

## Key Features

### 1. Build Caching
Nx caches build outputs. If you rebuild without changes, it uses the cache:

```bash
$ npx nx build fly-cli
✓ Cached from previous build (instant!)
```

### 2. Affected Commands
Only rebuild what changed since a base commit:

```bash
# Build only projects affected by changes
npx nx affected --target=build --base=main

# Test only affected projects
npx nx affected --target=test --base=HEAD~1
```

### 3. Parallel Execution
Run multiple tasks concurrently:

```bash
# Build up to 4 projects in parallel
npx nx run-many --target=build --all --parallel=4
```

### 4. Project Graph
Visualize dependencies between projects:

```bash
npx nx graph
# Opens browser with interactive graph
```

## Configuration Files

### nx.json
Main Nx configuration with:
- Task defaults and caching rules
- Named inputs for fine-grained invalidation
- Parallel execution settings

### project.json (per project)
Each project has a `project.json` defining:
- **name**: Project identifier
- **targets**: Build, test, lint, dev tasks
- **executor**: How to run tasks
- **tags**: For filtering and organization

Example for Go service:
```json
{
  "name": "orchestrator-api",
  "targets": {
    "build": {
      "executor": "nx:run-commands",
      "command": "go build -o bin/orchestrator-api ./cmd/orchestrator-api",
      "cache": true
    }
  }
}
```

## Migration Notes

### Changed
- ✅ Added Nx task orchestration
- ✅ Build caching enabled
- ✅ Affected-based builds available
- ✅ Parallel execution supported
- ✅ Project dependency tracking

### Unchanged
- ✅ Makefile still works
- ✅ `make build` still builds Go services
- ✅ Can run `go build` directly
- ✅ Existing scripts and CI/CD compatible
- ✅ Docker configs unchanged

## Best Practices

### When to Use Nx
- **CI/CD**: Use `nx affected` to only build what changed
- **Development**: Use `nx` for parallel builds and caching
- **Large changes**: `nx run-many` for rebuilding everything

### When to Use Make
- **Quick Go builds**: `make build` is fine for simple Go builds
- **Existing workflows**: Keep using familiar commands
- **Deployment**: Existing Make targets work as before

### Hybrid Approach (Recommended)
```bash
# Local development
make build              # Quick Go compile
npx nx dev dashboard   # Frontend with HMR

# CI/CD
npx nx affected --target=build --base=main  # Smart builds
npx nx affected --target=test --base=main   # Smart tests

# Pre-commit
make nx-lint            # Lint everything
```

## Performance Tips

1. **Use affected commands** in CI to save time
2. **Enable remote caching** (Nx Cloud) for team sharing
3. **Run tests in parallel**: `nx affected --target=test --parallel=4`
4. **Check cache usage**: Look for "Cached" in output

## Troubleshooting

### Project not detected
```bash
# Reset cache
npx nx reset

# Verify project.json exists
ls -la cmd/*/project.json

# Check if it should have a project.json
# (Not all directories need to be Nx projects)
```

### Builds fail
```bash
# Run without cache to debug
npx nx build fly-cli --skip-nx-cache

# Check the underlying command works
cd cmd/fly && go build
```

### Clear everything
```bash
# Remove Nx cache
npx nx reset

# Remove node_modules
rm -rf node_modules package-lock.json
npm install
```

## Next Steps

1. **Add CI Integration**: Use `nx affected` in GitHub Actions
2. **Enable Nx Cloud**: Share cache across team
3. **Add more projects**: Create project.json for remaining services
4. **Configure remote cache**: Speed up CI with distributed caching

## Resources

- [Nx Documentation](https://nx.dev)
- [Nx Go Plugin](https://nx.dev/recipes/golang/go-plugin)
- [Nx Rust Support](https://nx.dev/recipes/other/customize-executor)
- [Affected Command](https://nx.dev/concepts/affected)

## Support

For issues or questions:
1. Check `npx nx graph` for project structure
2. Run `npx nx show project <name>` for project details
3. Review build logs with `--verbose` flag
4. See [Nx troubleshooting](https://nx.dev/troubleshooting/troubleshooting)
