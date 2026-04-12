#!/bin/bash
# Install WASM3 from source for CGO bindings (user-local install, no sudo required)

set -e

WASM3_VERSION="v0.5.0"
INSTALL_PREFIX="${HOME}/.local"
BUILD_DIR="/tmp/wasm3-build-$$"

echo "=== Installing WASM3 ${WASM3_VERSION} to ${INSTALL_PREFIX} ==="

# Create directories
mkdir -p "${INSTALL_PREFIX}/include"
mkdir -p "${INSTALL_PREFIX}/lib"
mkdir -p "${INSTALL_PREFIX}/bin"

# Create build directory
mkdir -p "${BUILD_DIR}"
cd "${BUILD_DIR}"

# Clone WASM3
git clone --depth 1 --branch "${WASM3_VERSION}" https://github.com/wasm3/wasm3.git
cd wasm3

# Build the library
echo "Building WASM3 library..."
cd source

# Compile all m3_*.c files to object files
SOURCES="m3_api_libc.c m3_api_meta_wasi.c m3_api_tracer.c m3_api_uvwasi.c m3_api_wasi.c m3_bind.c m3_code.c m3_compile.c m3_core.c m3_emit.c m3_env.c m3_exec.c m3_function.c m3_info.c m3_module.c m3_optimize.c m3_parse.c"

for src in $SOURCES; do
    echo "  Compiling $src"
    gcc -c -O3 -I. "$src" -fPIC
done

# Create shared library
echo "Creating shared library..."
gcc -shared -o "${INSTALL_PREFIX}/lib/libwasm3.so" \
    m3_api_libc.o m3_api_meta_wasi.o m3_api_tracer.o m3_api_uvwasi.o m3_api_wasi.o \
    m3_bind.o m3_code.o m3_compile.o m3_core.o m3_emit.o m3_env.o m3_exec.o \
    m3_function.o m3_info.o m3_module.o m3_optimize.o m3_parse.o

# Create static library
ar rcs "${INSTALL_PREFIX}/lib/libwasm3.a" \
    m3_api_libc.o m3_api_meta_wasi.o m3_api_tracer.o m3_api_uvwasi.o m3_api_wasi.o \
    m3_bind.o m3_code.o m3_compile.o m3_core.o m3_emit.o m3_env.o m3_exec.o \
    m3_function.o m3_info.o m3_module.o m3_optimize.o m3_parse.o

# Copy headers (some may be optional, use || true to ignore missing)
echo "Installing headers..."
cp m3_api_defs.h "${INSTALL_PREFIX}/include/" || true
cp m3_api_libc.h "${INSTALL_PREFIX}/include/" || true
cp m3_api_llvm_bitcode.h "${INSTALL_PREFIX}/include/" 2>/dev/null || true
cp m3_api_tracer.h "${INSTALL_PREFIX}/include/" || true
cp m3_api_uvwasi.h "${INSTALL_PREFIX}/include/" || true
cp m3_api_wasi.h "${INSTALL_PREFIX}/include/" || true
cp m3_config.h "${INSTALL_PREFIX}/include/" || true
cp m3_config_defs.h "${INSTALL_PREFIX}/include/" 2>/dev/null || true
cp m3_core.h "${INSTALL_PREFIX}/include/" || true
cp m3_env.h "${INSTALL_PREFIX}/include/" || true
cp m3_exception.h "${INSTALL_PREFIX}/include/" 2>/dev/null || true
cp m3_exec.h "${INSTALL_PREFIX}/include/" 2>/dev/null || true
cp m3_exec_defs.h "${INSTALL_PREFIX}/include/" || true
cp m3_function.h "${INSTALL_PREFIX}/include/" || true
cp m3_compile.h "${INSTALL_PREFIX}/include/" 2>/dev/null || true

# Create main wasm3.h header that includes the others
cat > "${INSTALL_PREFIX}/include/wasm3.h" << 'EOF'
#ifndef WASM3_H
#define WASM3_H

#include "m3_env.h"
#include "m3_core.h"
#include "m3_function.h"

#ifdef __cplusplus
extern "C" {
#endif

// Main WASM3 API functions
typedef struct M3Runtime m3_runtime;
typedef struct M3Module m3_module;
typedef struct M3Function m3_function;
typedef struct M3Environment m3_environment;

m3_environment* m3_NewEnvironment(void);
void m3_FreeEnvironment(m3_environment* env);
m3_runtime* m3_NewRuntime(m3_environment* env, uint32_t stackSize);
void m3_FreeRuntime(m3_runtime* runtime);
int m3_ParseModule(m3_environment* env, m3_module** module, const uint8_t* wasm, uint32_t size);
void m3_FreeModule(m3_module* module);
int m3_LoadModule(m3_runtime* runtime, m3_module* module);
int m3_FindFunction(m3_function** func, m3_runtime* runtime, const char* name);
int m3_CallArgv(m3_function* func, int32_t argc, const char* argv[]);
const char* m3_GetErrorInfo(m3_runtime* runtime);
void* m3_GetMemory(m3_runtime* runtime, uint32_t* size, uint32_t memIndex);

#ifdef __cplusplus
}
#endif

#endif // WASM3_H
EOF

# Update user environment
cat >> ~/.bashrc << 'EOF'

# WASM3 paths for CGO
export WASM3_ROOT="${HOME}/.local"
export PKG_CONFIG_PATH="${WASM3_ROOT}/lib/pkgconfig:${PKG_CONFIG_PATH}"
export LD_LIBRARY_PATH="${WASM3_ROOT}/lib:${LD_LIBRARY_PATH}"
EOF

# Create pkgconfig file
mkdir -p "${INSTALL_PREFIX}/lib/pkgconfig"
cat > "${INSTALL_PREFIX}/lib/pkgconfig/wasm3.pc" << EOF
prefix=${INSTALL_PREFIX}
exec_prefix=\${prefix}
libdir=\${prefix}/lib
includedir=\${prefix}/include

Name: wasm3
Description: WebAssembly interpreter
Version: ${WASM3_VERSION}
Libs: -L\${libdir} -lwasm3
Cflags: -I\${includedir}
EOF

echo ""
echo "=== WASM3 installed successfully ==="
echo "Library: ${INSTALL_PREFIX}/lib/libwasm3.so"
echo "Headers: ${INSTALL_PREFIX}/include/wasm3.h"
echo ""
echo "Run: source ~/.bashrc"
echo "Then: CGO_ENABLED=1 go build ./internal/wasm/..."

# Cleanup
cd /
rm -rf "${BUILD_DIR}"
