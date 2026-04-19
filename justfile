# Justfile for FunctionFly - Test optimization commands
# Install just: cargo install just
# Run: just test-fast, just test-quick, etc.

# Default recipe - list available commands
default:
    @echo "Available test commands:"
    @echo "  just test           - Run all tests with fast profile"
    @echo "  just test-quick     - Run only lib tests (fastest)"
    @echo "  just test-release   - Run tests with full optimizations"
    @echo "  just test-nextest   - Run with cargo nextest (parallel)"
    @echo "  just test-watch     - Run tests in watch mode"
    @echo "  just test-local     - Run local runtime tests only"
    @echo "  just test-nodejs    - Run Node.js runtime tests only"
    @echo "  just test-microvm   - Run MicroVM runtime tests only"
    @echo ""
    @echo "Add test filter as argument: just test-local aot_cache"

# Variables
LOCAL_RUNTIME := "runtimes/local"
NODEJS_RUNTIME := "runtimes/nodejs"
MICROVM_RUNTIME := "runtimes/microvm"

# === Main Test Commands ===

# Run all tests with fast profile
test FILTER="":
    cd {{LOCAL_RUNTIME}} && cargo test --profile fast-test -- --test-threads=0 {{FILTER}}

# Run only lib tests (fastest compilation)
test-quick FILTER="":
    cd {{LOCAL_RUNTIME}} && cargo test --profile fast-test --lib -- --test-threads=0 {{FILTER}}

# Run tests with release-level optimizations
test-release FILTER="":
    cd {{LOCAL_RUNTIME}} && cargo test --profile test-release -- --test-threads=0 {{FILTER}}

# Run with cargo nextest (best parallelization, install with: cargo install cargo-nextest)
test-nextest:
    cd {{LOCAL_RUNTIME}} && cargo nextest run --profile fast

# Run tests in watch mode (requires cargo-watch)
test-watch:
    cd {{LOCAL_RUNTIME}} && cargo watch -x "test --profile fast-test --lib -- --test-threads=0"

# Compile tests only (no run)
test-compile:
    cd {{LOCAL_RUNTIME}} && cargo test --profile fast-test --no-run

# === Runtime-Specific Tests ===

# Run local runtime tests only
test-local FILTER="":
    cd {{LOCAL_RUNTIME}} && cargo test --profile fast-test -- --test-threads=0 {{FILTER}}

# Run Node.js runtime tests
test-nodejs FILTER="":
    cd {{NODEJS_RUNTIME}} && cargo test --profile fast-test -- --test-threads=0 {{FILTER}}

# Run MicroVM runtime tests
test-microvm FILTER="":
    cd {{MICROVM_RUNTIME}} && cargo test --profile fast-test -- --test-threads=0 {{FILTER}}

# === CI / Automation Commands ===

# Run all tests for CI (continues on failure)
test-ci:
    cd {{LOCAL_RUNTIME}} && cargo nextest run --profile ci || cargo test --profile fast-test -- --test-threads=0

# Run tests with timing info
test-timed:
    cd {{LOCAL_RUNTIME}} && time cargo test --profile fast-test -- --test-threads=0 --nocapture

# Single-threaded tests (for debugging race conditions)
test-single FILTER="":
    cd {{LOCAL_RUNTIME}} && cargo test -- --test-threads=1 {{FILTER}}

# === Utility Commands ===

# Clean build artifacts
clean:
    cd {{LOCAL_RUNTIME}} && cargo clean
    cd {{NODEJS_RUNTIME}} && cargo clean
    cd {{MICROVM_RUNTIME}} && cargo clean

# Check code without building
check:
    cd {{LOCAL_RUNTIME}} && cargo check

# Run clippy
lint:
    cd {{LOCAL_RUNTIME}} && cargo clippy --all-targets --all-features

# Format code
fmt:
    cd {{LOCAL_RUNTIME}} && cargo fmt
    cd {{NODEJS_RUNTIME}} && cargo fmt
    cd {{MICROVM_RUNTIME}} && cargo fmt
