#!/usr/bin/env bash
# Fast test runner script for functionfly-local runtime
# Usage: ./run-tests.sh [option]
#   fast      - Run with fast-test profile (optimized, parallel)
#   release   - Run with test-release profile (fully optimized)
#   quick     - Run only unit tests (fastest)
#   nextest   - Run with cargo nextest (if installed)
#   compile   - Compile only, don't run
#   watch     - Run tests in watch mode (requires cargo-watch)

set -e

cd "$(dirname "$0")"
RUNTIME_DIR="/home/micro/projects/functionfly/runtimes/local"
cd "$RUNTIME_DIR"

echo "Running tests in: $(pwd)"

case "${1:-fast}" in
    fast|f)
        echo "Running with fast-test profile (optimized, all cores)..."
        cargo test --profile fast-test -- --test-threads=0 "$@"
        ;;
    
    release|r)
        echo "Running with test-release profile (fully optimized)..."
        cargo test --profile test-release -- --test-threads=0 "$@"
        ;;
    
    quick|q)
        echo "Running unit tests only (excludes integration tests)..."
        cargo test --profile fast-test --lib -- --test-threads=0 "$@"
        ;;
    
    nextest|n)
        if ! command -v cargo-nextest &> /dev/null; then
            echo "cargo-nextest not found. Installing..."
            cargo install cargo-nextest --locked
        fi
        echo "Running with cargo nextest (highly parallel)..."
        cargo nextest run --profile fast -- --test-threads=0
        ;;
    
    compile|c)
        echo "Compiling tests only (no-run)..."
        cargo test --profile fast-test --no-run
        ;;
    
    watch|w)
        if ! command -v cargo-watch &> /dev/null; then
            echo "cargo-watch not found. Installing..."
            cargo install cargo-watch
        fi
        echo "Running tests in watch mode..."
        cargo watch -x "test --profile fast-test -- --test-threads=0"
        ;;
    
    single|s)
        echo "Running single-threaded (for debugging race conditions)..."
        cargo test -- --test-threads=1
        ;;
    
    help|h|--help|-h)
        cat << 'EOF'
Usage: ./run-tests.sh [option] [test-filter]

Options:
  fast      Run with fast-test profile (balanced, default)
  release   Run with test-release profile (fully optimized)
  quick     Run only lib tests (fastest compilation)
  nextest   Run with cargo nextest (best parallelization)
  compile   Compile only, don't run tests
  watch     Run in watch mode (auto-run on file changes)
  single    Run single-threaded (for debugging)
  help      Show this help

Examples:
  ./run-tests.sh fast test_aot          # Run AOT tests with fast profile
  ./run-tests.sh nextest pool           # Run pool tests with nextest
  ./run-tests.sh quick                  # Quick lib tests only
EOF
        ;;
    
    *)
        # Pass through to cargo test with fast profile
        echo "Running with fast-test profile and filter: $1"
        cargo test --profile fast-test -- --test-threads=0 "$@"
        ;;
esac
