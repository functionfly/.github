#!/usr/bin/env bash
# Install host tooling used by internal/bundler (Rust→WASM, Go WASI, WAT→WASM).
# Safe to re-run. From repo root: bash scripts/install-wasm-toolchain.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> FunctionFly WASM toolchain setup"

if ! command -v go >/dev/null 2>&1; then
  echo "WARN: go not found. Install Go 1.21+ for runtime go1.21 (https://go.dev/dl/)."
else
  echo "OK: $(go version)"
fi

if command -v apt-get >/dev/null 2>&1 && ! command -v wat2wasm >/dev/null 2>&1; then
  echo "==> Installing WABT (wat2wasm) via apt…"
  sudo apt-get update -qq
  sudo apt-get install -y wabt
elif command -v brew >/dev/null 2>&1 && ! command -v wat2wasm >/dev/null 2>&1; then
  echo "==> Installing WABT via Homebrew…"
  brew install wabt
elif command -v wat2wasm >/dev/null 2>&1; then
  echo "OK: wat2wasm ($(command -v wat2wasm))"
else
  echo "WARN: wat2wasm missing. Install WABT for browser-wasm .wat: https://github.com/WebAssembly/wabt"
fi

if ! command -v rustup >/dev/null 2>&1; then
  echo "WARN: rustup missing. Install https://rustup.rs for runtime rust."
else
  echo "==> Rust targets (default toolchain — used by fly bundle / cargo build)"
  rustup target add wasm32-wasip1 2>/dev/null || true
  rustup target add wasm32-wasi 2>/dev/null || true
  echo "==> Rust targets on stable (optional)"
  if ! rustup target add wasm32-wasip1 wasm32-wasi --toolchain stable 2>/dev/null; then
    echo "    (skipped: nightly-only wasi, rustup conflict, or offline — bundler prefers wasm32-wasip1 on default)"
    echo "    Repair: rustup update stable && rustup target add wasm32-wasip1 --toolchain stable"
  fi
fi

echo ""
echo "Optional for JS→WASM: javy + esbuild on PATH."
echo "Done."
