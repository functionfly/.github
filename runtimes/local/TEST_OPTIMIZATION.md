# Optimizing Rust Test Speed

This guide covers how to run Rust tests faster for the FunctionFly local runtime.

## Quick Start

Use the provided test runner script:

```bash
cd runtimes/local

# Fast optimized tests (default - best balance)
./run-tests.sh fast

# Or use the cargo alias
cargo tfast
```

## Test Profiles

We've configured several test profiles optimized for different use cases:

### 1. Default Test Profile (Debug)
```bash
cargo test
```
- Unoptimized code
- Fastest compilation
- Slowest execution

### 2. Fast Test Profile (Recommended)
```bash
cargo test --profile fast-test
# or
cargo tf
```
- Basic optimizations (`opt-level = 1`)
- Faster linking (no LTO)
- Parallel codegen units
- **~2-5x faster execution** than debug
- Good for CI and local development

### 3. Test Release Profile
```bash
cargo test --profile test-release
```
- Full release optimizations
- Thin LTO enabled
- Fastest execution
- Slowest compilation
- Best for final verification before deploy

### 4. Release Mode Tests
```bash
cargo tr  # alias for: cargo test --release
```
- Full release build
- Use when tests are too slow in debug

## Parallel Execution

### Using All CPU Cores
```bash
# Auto-detect threads
cargo test -- --test-threads=0

# Specific number
cargo test -- --test-threads=8
```

### Using Nextest (Best Parallelization)
```bash
# Install nextest once
cargo install cargo-nextest --locked

# Run with nextest (highly recommended)
cargo nextest run

# Fast profile with nextest
cargo nextest run --profile fast

# CI profile (retries, all tests even on failure)
cargo nextest run --profile ci
```

Nextest provides:
- Better parallel scheduling
- Test grouping by resource requirements
- Retry support
- Clearer output

## Running Specific Test Categories

### Unit Tests Only (Fastest)
```bash
cargo test --lib --profile fast-test -- --test-threads=0
```

### Exclude Slow Tests
```bash
# Skip tests with "integration" or "slow" in name
cargo test --profile fast-test -- --skip integration --skip slow
```

### Single Module Tests
```bash
cargo test --profile fast-test aot_cache
```

## Build Optimizations

### Faster Linker (Optional)
Install and configure a faster linker for even faster builds:

**With lld (LLVM linker):**
```bash
# Install lld
sudo apt install lld  # Ubuntu/Debian

# Uncomment in .cargo/config.toml:
# [target.x86_64-unknown-linux-gnu]
# linker = "clang"
# rustflags = ["-C", "link-arg=-fuse-ld=lld"]
```

**With mold (fastest):**
```bash
# Install mold
sudo apt install mold  # Ubuntu/Debian

# Uncomment in .cargo/config.toml:
# [target.x86_64-unknown-linux-gnu]
# rustflags = ["-C", "link-arg=-fuse-ld=mold"]
```

### Parallel Compilation
Already configured in `.cargo/config.toml`:
```toml
[build]
jobs = 0  # Use all cores
```

## Runtime Optimizations

### Test Code Changes Made
1. **Replaced sleeps with yields**: Changed `tokio::time::sleep(Duration::from_millis(10))` to `tokio::task::yield_now()` - saves ~9ms per async test
2. **Test profiles**: Added optimized profiles that balance compile time and runtime
3. **Parallel execution**: Configured to use all CPU cores by default

### Environment Variables
Set these for faster test execution:
```bash
# Disable backtraces (faster, but less debug info)
export RUST_BACKTRACE=0

# Or in .cargo/config.toml (already set)
```

## Watch Mode (Continuous Testing)

```bash
# Install cargo-watch
cargo install cargo-watch

# Run tests on every file change
cargo watch -x "test --profile fast-test --lib -- --test-threads=0"

# Or use the script
./run-tests.sh watch
```

## CI Optimizations

For CI/CD pipelines, use the CI profile with nextest:

```bash
# Run all tests, retry flaky ones once, continue on failure
cargo nextest run --profile ci -- --test-threads=0

# Generate JUnit report for test analysis
cargo nextest run --profile ci --results-junit test-results.xml
```

## Benchmarking Test Performance

```bash
# Time the test run
time cargo test --profile fast-test -- --test-threads=0

# Compare with release
time cargo test --profile test-release -- --test-threads=0

# With nextest (usually fastest)
time cargo nextest run --profile fast
```

## Troubleshooting

### Tests Too Slow in Debug
Use `cargo tf` (fast-test profile) - includes basic optimizations.

### Compilation Too Slow
Use `cargo tq` (quick) for lib tests only, or increase codegen units.

### Linking Too Slow
Install and configure lld or mold linker (see above).

### Test Timeouts
Some tests may need more time in optimized builds due to different timing characteristics. If tests fail with timeouts:
```bash
# Run single-threaded for debugging
cargo test -- --test-threads=1
```

## Summary Table

| Command | Compile | Run | Best For |
|---------|---------|-----|----------|
| `cargo test` | Fast | Slow | Quick iteration |
| `cargo tf` | Medium | Fast | **Default (recommended)** |
| `cargo tr` | Slow | Fastest | Pre-deployment |
| `cargo nextest run` | Medium | Fastest | Parallel test runs |

## Recommended Workflow

1. **During development**: Use `cargo tq` (quick lib tests)
2. **Before commit**: Use `cargo tf` (all tests with fast profile)
3. **In CI**: Use `cargo nextest run --profile ci`
4. **Final verification**: Use `cargo tr` (test-release profile)
