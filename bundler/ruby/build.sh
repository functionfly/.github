#!/usr/bin/env bash
# Build mruby as a WASM module for use with FunctionFly.
# Requires: emcc (Emscripten) or WASI-SDK clang + ruby
#
# Usage: bash bundler/ruby/build.sh
#
# The emsdk is bundled at <repo>/emsdk/. If emcc is not on PATH the script
# activates it automatically.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
MRUBY_VERSION="3.2.0"

# ── Activate emsdk if emcc is not on PATH ──────────────────────────────────
if ! command -v emcc >/dev/null 2>&1; then
    EMSDK_DIR="$ROOT/emsdk"
    if [ -f "$EMSDK_DIR/emsdk_env.sh" ]; then
        echo "==> Activating bundled emsdk"
        source "$EMSDK_DIR/emsdk_env.sh" 2>/dev/null || true
    fi
fi

# ── Ensure ruby is on PATH (linuxbrew) ─────────────────────────────────────
if ! command -v ruby >/dev/null 2>&1; then
    if [ -x /home/linuxbrew/.linuxbrew/bin/ruby ]; then
        export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
    fi
fi

echo "==> Building mruby $MRUBY_VERSION for WASM"

# Check for compilers
if command -v emcc >/dev/null 2>&1; then
    CC_BIN="emcc"
    echo "Using Emscripten (emcc) — $(emcc --version | head -1)"
    USE_EMSCRIPTEN=1
elif [ -n "${WASI_SDK_PATH:-}" ] && [ -f "$WASI_SDK_PATH/bin/clang" ]; then
    CC_BIN="$WASI_SDK_PATH/bin/clang"
    echo "Using WASI-SDK clang"
    USE_EMSCRIPTEN=0
else
    echo "ERROR: No WASM C compiler found."
    echo "Install Emscripten (bundled at emsdk/) or set WASI_SDK_PATH."
    exit 1
fi

if ! command -v ruby >/dev/null 2>&1; then
    echo "ERROR: ruby is required for mruby's build system."
    exit 1
fi

# ── Clone mruby if not present ─────────────────────────────────────────────
if [ ! -d "$BUILD_DIR/mruby" ]; then
    echo "==> Cloning mruby $MRUBY_VERSION"
    mkdir -p "$BUILD_DIR"
    git clone --depth 1 --branch "$MRUBY_VERSION" \
        https://github.com/mruby/mruby.git "$BUILD_DIR/mruby"
fi

# ── Write the cross-build config ───────────────────────────────────────────
CROSSCONF="$BUILD_DIR/mruby_build_config.rb"

if [ "$USE_EMSCRIPTEN" -eq 1 ]; then
    EMCC_DIR="$(dirname "$(command -v emcc)")"
    cat > "$CROSSCONF" <<RUBY
MRuby::Build.new do |conf|
  toolchain :gcc
  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-bin-mrbc"
end

MRuby::CrossBuild.new("functionfly-wasm") do |conf|
  toolchain :clang

  conf.cc.command = "emcc"
  conf.cc.flags   = %w(-O2 -DMRB_NO_STDIO -DMRB_USE_FLOAT32 -w)

  conf.linker.command     = "emcc"
  conf.linker.flags       = %w(-O2 -s STANDALONE_WASM=1 -s ALLOW_MEMORY_GROWTH=1 --no-entry)
  conf.archiver.command   = "emar"

  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-error"
  conf.gem core: "mruby-math"
  conf.gem core: "mruby-time"
  conf.gem core: "mruby-string-ext"
  conf.gem core: "mruby-numeric-ext"
  conf.gem core: "mruby-array-ext"
  conf.gem core: "mruby-hash-ext"
  conf.gem core: "mruby-range-ext"
  conf.gem core: "mruby-proc-ext"
  conf.gem core: "mruby-symbol-ext"
  conf.gem core: "mruby-kernel-ext"
  conf.gem core: "mruby-object-ext"
  conf.gem core: "mruby-fiber"
  conf.gem core: "mruby-enumerator"
  conf.gem core: "mruby-enum-lazy"
  conf.gem core: "mruby-toplevel-ext"
  conf.gem core: "mruby-method"
end
RUBY
else
    cat > "$CROSSCONF" <<RUBY
MRuby::Build.new do |conf|
  toolchain :gcc
  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-bin-mrbc"
end

MRuby::CrossBuild.new("functionfly-wasm") do |conf|
  toolchain :clang

  conf.cc.command = "$WASI_SDK_PATH/bin/clang"
  conf.cc.flags   = %w(--target=wasm32-wasi -O2 -DMRB_NO_STDIO -DMRB_USE_FLOAT32 -w)

  conf.linker.command     = "$WASI_SDK_PATH/bin/clang"
  conf.linker.flags       = %w(--target=wasm32-wasi -O2 -nostartfiles -Wl,--no-entry -Wl,--allow-undefined)
  conf.archiver.command   = "$WASI_SDK_PATH/bin/llvm-ar"

  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-error"
  conf.gem core: "mruby-math"
  conf.gem core: "mruby-time"
  conf.gem core: "mruby-string-ext"
  conf.gem core: "mruby-numeric-ext"
  conf.gem core: "mruby-array-ext"
  conf.gem core: "mruby-hash-ext"
  conf.gem core: "mruby-range-ext"
  conf.gem core: "mruby-proc-ext"
  conf.gem core: "mruby-symbol-ext"
  conf.gem core: "mruby-kernel-ext"
  conf.gem core: "mruby-object-ext"
  conf.gem core: "mruby-fiber"
  conf.gem core: "mruby-enumerator"
  conf.gem core: "mruby-enum-lazy"
  conf.gem core: "mruby-toplevel-ext"
  conf.gem core: "mruby-method"
end
RUBY
fi

# ── Run mruby's rake build ─────────────────────────────────────────────────
cd "$BUILD_DIR/mruby"
echo "==> Running mruby rake build (host + cross)"
ruby minirake MRUBY_CONFIG="$CROSSCONF" 2>&1 | tail -20

# ── Collect the cross-build artifacts ──────────────────────────────────────
CROSS_BUILD="$BUILD_DIR/mruby/build/functionfly-wasm"
CROSS_LIB="$CROSS_BUILD/lib/libmruby.a"

if [ ! -f "$CROSS_LIB" ]; then
    echo "ERROR: Cross-build libmruby.a not found at $CROSS_LIB"
    echo "Build artifacts:"
    find "$CROSS_BUILD" -type f 2>/dev/null | head -20
    exit 1
fi

echo "  Cross-build lib: $CROSS_LIB ($(du -h "$CROSS_LIB" | cut -f1))"

# ── Compile the C wrapper and link into mruby.wasm ─────────────────────────
echo "==> Compiling C wrapper and linking mruby.wasm"
OUT="$SCRIPT_DIR/mruby.wasm"
WRAPPER="$SCRIPT_DIR/mruby_wasm.c"
MRUBY_INCLUDE="$BUILD_DIR/mruby/include"

if [ "$USE_EMSCRIPTEN" -eq 1 ]; then
    emcc "$WRAPPER" \
        "$CROSS_LIB" \
        -I"$MRUBY_INCLUDE" \
        -o "$OUT" \
        -s WASM=1 \
        -s STANDALONE_WASM=1 \
        -s EXPORTED_FUNCTIONS='["_mruby_init","_mruby_exec","_mruby_exec_string","_mruby_result_ptr","_mruby_result_len","_mruby_error_ptr","_mruby_error_len","_mruby_cleanup","_malloc","_free"]' \
        -s ALLOW_MEMORY_GROWTH=1 \
        -s INITIAL_MEMORY=2097152 \
        -s TOTAL_STACK=1048576 \
        --no-entry \
        -O2 \
        -DMRB_NO_STDIO \
        -DMRB_USE_FLOAT32 \
        -w
else
    "$WASI_SDK_PATH/bin/clang" \
        "$WRAPPER" \
        "$CROSS_LIB" \
        -I"$MRUBY_INCLUDE" \
        -o "$OUT" \
        --target=wasm32-wasi \
        -O2 \
        -nostartfiles \
        -Wl,--no-entry \
        -Wl,--export=mruby_init \
        -Wl,--export=mruby_exec \
        -Wl,--export=mruby_exec_string \
        -Wl,--export=mruby_result_ptr \
        -Wl,--export=mruby_result_len \
        -Wl,--export=mruby_error_ptr \
        -Wl,--export=mruby_error_len \
        -Wl,--export=mruby_cleanup \
        -Wl,--export=malloc \
        -Wl,--export=free \
        -Wl,--allow-undefined \
        -DMRB_NO_STDIO \
        -DMRB_USE_FLOAT32 \
        -w
fi

SIZE=$(du -h "$OUT" | cut -f1)
echo "==> Built $OUT ($SIZE)"

# ── Verify the WASM module ─────────────────────────────────────────────────
if command -v wasm-objdump >/dev/null 2>&1; then
    echo "==> Exports:"
    wasm-objdump -x "$OUT" 2>/dev/null | grep "export" | head -15
elif command -v wasm2wat >/dev/null 2>&1; then
    echo "==> Exports:"
    wasm2wat "$OUT" 2>/dev/null | grep "(export" | head -15
fi

echo "Done."
