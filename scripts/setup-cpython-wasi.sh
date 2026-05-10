#!/usr/bin/env bash
# scripts/setup-cpython-wasi.sh
# Download or build CPython-WASI for use with RuntimeType::PythonWasm in runtimes/local.
set -euo pipefail

VERSION="${1:-3.13.0}"
DEST_DIR="./runtimes/cpython-wasi"

# Check if already present
if [[ -f "${DEST_DIR}/python.wasm" ]]; then
    echo "CPython WASI already present at ${DEST_DIR}/python.wasm"
    echo "Run with --force to rebuild/replace."
    exit 0
fi

# Allow force rebuild
if [[ "${1:-}" == "--force" ]]; then
    rm -rf "${DEST_DIR}"
    shift || true
    VERSION="${1:-3.13.0}"
fi

URL="https://www.python.org/ftp/python/${VERSION}/python-${VERSION}-wasm32-wasi.tar.gz"

mkdir -p "${DEST_DIR}"

# Try to download pre-built binary first
echo "Attempting to download CPython ${VERSION} WASI build..."
if curl -fsSL -I "${URL}" >/dev/null 2>&1; then
    echo "Downloading from ${URL}..."
    curl -fsSL "${URL}" | tar -xz -C "${DEST_DIR}" --strip-components=1
else
    echo "No pre-built binary available. Building CPython ${VERSION} WASI from source..."

    # Check prerequisites
    if ! command -v wasmtime >/dev/null 2>&1; then
        echo "ERROR: wasmtime is required. Install from https://wasmtime.dev/"
        exit 1
    fi

    # Check for wasi-sdk or download it
    WASI_SDK_PATH="${WASI_SDK_PATH:-}"
    if [[ -z "${WASI_SDK_PATH}" ]] || [[ ! -x "${WASI_SDK_PATH}/bin/clang" ]]; then
        if [[ -x "/opt/wasi-sdk/bin/clang" ]]; then
            WASI_SDK_PATH="/opt/wasi-sdk"
        else
            echo "WASI SDK not found. Downloading..."
            WASI_SDK_VERSION="25"
            WASI_SDK_URL="https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-${WASI_SDK_VERSION}/wasi-sdk-${WASI_SDK_VERSION}.0-x86_64-linux.tar.gz"
            WASI_SDK_TMP="/tmp/wasi-sdk-setup-$$"
            mkdir -p "${WASI_SDK_TMP}"
            curl -fsSL -L "${WASI_SDK_URL}" | tar -xz -C "${WASI_SDK_TMP}" --strip-components=1
            WASI_SDK_PATH="${WASI_SDK_TMP}"
        fi
    fi

    echo "Using WASI SDK at ${WASI_SDK_PATH}"

    # Clone CPython source (shallow)
    CPYTHON_TMP="/tmp/cpython-wasi-build-$$"
    echo "Cloning CPython ${VERSION} source..."
    git clone --depth=1 --branch="v${VERSION}" https://github.com/python/cpython.git "${CPYTHON_TMP}"

    # Build CPython WASI
    echo "Building CPython WASI (this may take several minutes)..."
    cd "${CPYTHON_TMP}"
    export WASI_SDK_PATH
    bash Tools/wasm/build_wasi.sh 2>&1 | tail -20

    # Copy artifacts
    cp "builddir/wasi/python.wasm" "${DEST_DIR}/"

    # Copy stdlib: build lib + source Lib
    mkdir -p "${DEST_DIR}/lib"
    if [[ -d "builddir/wasi/build/lib.wasi-wasm32-${VERSION%.*}" ]]; then
        cp -r "builddir/wasi/build/lib.wasi-wasm32-${VERSION%.*}"/* "${DEST_DIR}/lib/" 2>/dev/null || true
    fi
    cp -r "Lib/"* "${DEST_DIR}/lib/"

    # Cleanup
    rm -rf "${CPYTHON_TMP}" "${WASI_SDK_TMP}"

    echo "Build complete."
fi

echo "Verifying extraction..."
if [[ -f "${DEST_DIR}/python.wasm" ]]; then
    echo "OK: ${DEST_DIR}/python.wasm ($(stat -c%s "${DEST_DIR}/python.wasm" | numfmt --to=iec))"
else
    echo "ERROR: python.wasm not found in ${DEST_DIR}"
    exit 1
fi

if [[ -d "${DEST_DIR}/lib" ]]; then
    echo "OK: ${DEST_DIR}/lib/ (stdlib)"
else
    echo "WARNING: ${DEST_DIR}/lib/ not found — imports will fail until stdlib is present"
fi

# Quick smoke test
echo "Running smoke test..."
if wasmtime --dir "${DEST_DIR}/lib::/lib/python${VERSION%.*}" "${DEST_DIR}/python.wasm" -c "print('WASI OK')" >/dev/null 2>&1; then
    echo "OK: wasmtime smoke test passed"
else
    echo "WARNING: wasmtime smoke test failed — runtime may still work with correct config"
fi

echo ""
echo "Update config.rs defaults if you want these to be the CLI defaults:"
echo "  cpython_wasm_path:     \"${DEST_DIR}/python.wasm\""
echo "  cpython_stdlib_path:   \"${DEST_DIR}/lib\""
echo ""
echo "Done."
