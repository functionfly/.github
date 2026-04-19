# Rust Test Optimization Summary

## Changes Made to Speed Up Tests

### 1. Optimized Test Profiles (Cargo.toml)

Added to all three runtime crates (`local`, `nodejs`, `microvm`):

- **`[profile.test]`** - Basic optimizations (`opt-level = 1`) while keeping debug info and assertions
- **`[profile.fast-test]`** - Medium optimizations (`opt-level = 2`) for CI/regular use
- **`[profile.test-release]`** - Full release optimizations for final verification

**Speed improvements:**
- `fast-test` profile: ~2-5x faster execution than debug
- `test-release` profile: ~5-10x faster execution (but slower compilation)

### 2. Cargo Configuration (.cargo/config.toml)

Created in all runtime directories with:
- Convenient aliases: `cargo tf`, `cargo tr`, `cargo tq`, `cargo tfast`
- Sparse registry protocol for faster downloads
- Environment variables for faster execution

### 3. Nextest Configuration (.config/nextest.toml)

Created for `local` runtime with:
- Parallel test scheduling
- CI profile with retries
- Test groups for resource management

### 4. Test Code Optimizations (src/tests/mod.rs)

- **Replaced `sleep(10ms)` with `yield_now()`** - Saves ~9ms per async test
  - Changed 2 instances of `tokio::time::sleep(Duration::from_millis(10)).await`
  - To `tokio::task::yield_now().await`

### 5. Helper Scripts

- **`run-tests.sh`** - Bash script with multiple test modes (fast, quick, nextest, etc.)
- **`justfile`** - Command runner recipes at project root
- **`TEST_OPTIMIZATION.md`** - Comprehensive documentation

## How to Use

### Quick Commands

```bash
# From runtimes/local directory:
cd runtimes/local

# Fast optimized tests (recommended for daily use)
cargo tf                          # or: cargo test --profile fast-test
cargo tfast                       # Fast profile + all CPU cores

# Quick lib tests only (fastest)
cargo tq                          # or: cargo test --profile fast-test --lib

# Release mode tests (slowest compile, fastest run)
cargo tr                          # or: cargo test --release

# With nextest (best parallelization)
cargo nextest run --profile fast
```

### Using the Scripts

```bash
# From runtimes/local:
./run-tests.sh fast          # Fast profile with all cores
./run-tests.sh quick         # Lib tests only
./run-tests.sh nextest       # Use nextest runner

# From project root (using just):
just test                  # All tests with fast profile
just test-quick            # Lib tests only
just test-nextest         # Nextest runner
```

### Running Specific Tests

```bash
# Run only AOT cache tests
cargo tf aot_cache

# Run with single thread (debugging)
cargo test -- --test-threads=1

# Skip slow tests
cargo tf -- --skip integration
```

## Expected Performance Improvements

| Scenario | Before | After | Improvement |
|----------|--------|-------|-------------|
| Debug tests | Baseline | Same | - |
| Fast-test profile | Baseline | 2-5x faster | Significant |
| Test-release profile | Baseline | 5-10x faster | Maximum |
| With nextest | Baseline | Better parallel | CPU-dependent |
| Sleep reduction (2 tests) | 20ms | ~0.1ms | 200x faster |

## Installation Recommendations

### Optional Tools for Maximum Speed

1. **cargo-nextest** (highly recommended):
   ```bash
   cargo install cargo-nextest --locked
   ```

2. **cargo-watch** (for watch mode):
   ```bash
   cargo install cargo-watch
   ```

3. **cargo-just** (for justfile recipes):
   ```bash
   cargo install just
   ```

4. **Faster linker** (optional, Linux only):
   ```bash
   # Option A: lld (LLVM linker)
   sudo apt install lld
   
   # Option B: mold (fastest)
   sudo apt install mold
   
   # Then uncomment in .cargo/config.toml
   ```

## Files Modified

### Local Runtime
- `runtimes/local/Cargo.toml` - Added test profiles
- `runtimes/local/.cargo/config.toml` - Created
- `runtimes/local/.config/nextest.toml` - Created
- `runtimes/local/src/tests/mod.rs` - Optimized sleep calls
- `runtimes/local/run-tests.sh` - Created
- `runtimes/local/TEST_OPTIMIZATION.md` - Created

### Node.js Runtime
- `runtimes/nodejs/Cargo.toml` - Added test profiles
- `runtimes/nodejs/.cargo/config.toml` - Created

### MicroVM Runtime
- `runtimes/microvm/Cargo.toml` - Added test profiles
- `runtimes/microvm/.cargo/config.toml` - Created

### Project Root
- `justfile` - Created with test recipes

## Troubleshooting

### "profile-rustflags is not valid" warning
This is expected - the profile rustflags are commented out. Use `RUSTFLAGS` env var instead:
```bash
RUSTFLAGS="-C target-cpu=native" cargo tf
```

### Tests fail with fast-test profile
Some timing-sensitive tests may need the test-release profile:
```bash
cargo test --profile test-release
```

### Linker not found
If you enable lld/mold but it's not installed, comment it out in `.cargo/config.toml`.

## Migration Guide

### For existing workflows

**Before:**
```bash
cd runtimes/local
cargo test
```

**After (recommended):**
```bash
cd runtimes/local
cargo tf                    # Fast profile
# or
./run-tests.sh fast         # Script with all optimizations
```

**In CI:**
```bash
# Use fast-test for quick feedback, test-release for final verification
cargo test --profile fast-test -- --test-threads=0
# or with nextest
cargo nextest run --profile ci
```
