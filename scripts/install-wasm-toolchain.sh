#!/usr/bin/env bash
# Install host tooling used by the Universal WASM Runtime Pipeline.
# Covers: Rust→WASM, Go→WASM, C/C++→WASM, Ruby (mruby), Kotlin, Swift, JS→WASM.
# Safe to re-run. From repo root: bash scripts/install-wasm-toolchain.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

AVAILABLE=()
MISSING=()

echo "==> FunctionFly Universal WASM Toolchain Setup"
echo ""

# ── Go ──────────────────────────────────────────────────────────────
if ! command -v go >/dev/null 2>&1; then
  echo "WARN: go not found. Install Go 1.21+ for runtime go/go1.21 (https://go.dev/dl/)."
  MISSING+=("go")
else
  echo "OK: $(go version)"
  AVAILABLE+=("go")
fi

# ── WABT (wat2wasm) ────────────────────────────────────────────────
if command -v apt-get >/dev/null 2>&1 && ! command -v wat2wasm >/dev/null 2>&1; then
  echo "==> Installing WABT (wat2wasm) via apt…"
  sudo apt-get update -qq
  sudo apt-get install -y wabt
elif command -v brew >/dev/null 2>&1 && ! command -v wat2wasm >/dev/null 2>&1; then
  echo "==> Installing WABT via Homebrew…"
  brew install wabt
elif command -v wat2wasm >/dev/null 2>&1; then
  echo "OK: wat2wasm ($(command -v wat2wasm))"
  AVAILABLE+=("wabt")
else
  echo "WARN: wat2wasm missing. Install WABT for browser-wasm .wat: https://github.com/WebAssembly/wabt"
  MISSING+=("wabt")
fi

# ── Rust (cargo + wasm targets) ────────────────────────────────────
if ! command -v rustup >/dev/null 2>&1; then
  echo "WARN: rustup missing. Install https://rustup.rs for runtime rust."
  MISSING+=("rust")
else
  echo "==> Rust targets (default toolchain)"
  rustup target add wasm32-wasip1 2>/dev/null || true
  rustup target add wasm32-wasi 2>/dev/null || true
  echo "==> Rust targets on stable (optional)"
  if ! rustup target add wasm32-wasip1 wasm32-wasi --toolchain stable 2>/dev/null; then
    echo "    (skipped: nightly-only wasi, rustup conflict, or offline)"
  fi
  echo "OK: $(cargo --version)"
  AVAILABLE+=("rust")
fi

# ── C/C++ (Emscripten or WASI-SDK) ─────────────────────────────────
if [ -n "${WASI_SDK_PATH:-}" ] && [ -f "$WASI_SDK_PATH/bin/clang" ]; then
  echo "OK: WASI-SDK at $WASI_SDK_PATH"
  AVAILABLE+=("c", "cpp")
elif [ -f "/opt/wasi-sdk/bin/clang" ]; then
  echo "OK: WASI-SDK at /opt/wasi-sdk"
  AVAILABLE+=("c", "cpp")
elif command -v emcc >/dev/null 2>&1; then
  echo "OK: Emscripten ($(emcc --version | head -1))"
  AVAILABLE+=("c", "cpp", "emscripten")
elif [ -d "$ROOT/emsdk" ]; then
  echo "INFO: emsdk/ directory found. Activate it:"
  echo "      source emsdk/emsdk_env.sh"
  MISSING+=("c-emscripten")
else
  echo "WARN: No C/C++ WASM compiler found."
  echo "      Install WASI-SDK: https://github.com/aspect-build/aspect-cli"
  echo "      Or Emscripten:   cd emsdk && ./emsdk install latest && ./emsdk activate latest"
  MISSING+=("c", "cpp")
fi

# ── Ruby (mruby.wasm) ──────────────────────────────────────────────
if [ -f "$ROOT/bundler/ruby/mruby.wasm" ]; then
  echo "OK: mruby.wasm ($(du -h "$ROOT/bundler/ruby/mruby.wasm" | cut -f1))"
  AVAILABLE+=("ruby")
elif [ -n "${MRUBY_WASM_PATH:-}" ] && [ -f "$MRUBY_WASM_PATH" ]; then
  echo "OK: mruby.wasm via MRUBY_WASM_PATH"
  AVAILABLE+=("ruby")
else
  echo "WARN: mruby.wasm not found. Build it:"
  echo "      See bundler/ruby/build.sh (or download from releases)"
  MISSING+=("ruby")
fi

# ── Kotlin ─────────────────────────────────────────────────────────
if command -v kotlin >/dev/null 2>&1; then
  echo "OK: Kotlin ($(kotlin -version 2>&1 || echo 'version unknown'))"
  AVAILABLE+=("kotlin")
elif command -v kotlinc >/dev/null 2>&1; then
  echo "OK: kotlinc ($(kotlinc -version 2>&1 || echo 'version unknown'))"
  AVAILABLE+=("kotlin")
else
  echo "WARN: Kotlin not found. Install Kotlin 1.9+ for Kotlin/WASM support:"
  echo "      https://kotlinlang.org/docs/wasm-getting-started.html"
  MISSING+=("kotlin")
fi

# ── Swift (SwiftWasm) ─────────────────────────────────────────────
if command -v carton >/dev/null 2>&1; then
  echo "OK: carton (SwiftWasm build tool)"
  AVAILABLE+=("swift")
elif command -v swiftc >/dev/null 2>&1; then
  echo "OK: swiftc ($(swiftc -version 2>&1 | head -1 || echo 'version unknown'))"
  echo "    Note: SwiftWasm target support depends on toolchain version"
  AVAILABLE+=("swift")
else
  echo "WARN: Swift not found. Install SwiftWasm for runtime swift:"
  echo "      brew install swiftwasm/swiftwasm/carton"
  echo "      Or: https://swiftwasm.org"
  MISSING+=("swift")
fi

# ── JavaScript (Javy) ──────────────────────────────────────────────
if command -v javy >/dev/null 2>&1; then
  echo "OK: Javy ($(javy --version 2>&1 || echo 'version unknown'))"
  AVAILABLE+=("javascript")
else
  echo "INFO: Javy not on PATH. Install for JS→WASM compilation:"
  echo "      npm install -g @shopify/javy"
  MISSING+=("javy")
fi

# ── esbuild ────────────────────────────────────────────────────────
if command -v esbuild >/dev/null 2>&1; then
  echo "OK: esbuild ($(esbuild --version 2>&1 || echo 'version unknown'))"
else
  echo "INFO: esbuild not on PATH. Install for JS/TS bundling:"
  echo "      npm install -g esbuild"
fi

# ── Summary ────────────────────────────────────────────────────────
echo ""
echo "==> Toolchain Summary"
echo "    Available: ${AVAILABLE[*]:-none}"
if [ ${#MISSING[@]} -gt 0 ]; then
  echo "    Missing:   ${MISSING[*]}"
fi
echo ""
echo "Done."
