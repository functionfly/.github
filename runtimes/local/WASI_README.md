# WASI Support in FunctionFly Local Runtime

This document describes the WebAssembly System Interface (WASI) support implemented in the FunctionFly local runtime.

## Overview

WASI provides a standardized interface for WebAssembly modules to interact with the host system in a portable way. The FunctionFly local runtime now supports WASI, allowing WebAssembly functions to:

- Access environment variables
- Read from and write to the filesystem (with configurable permissions)
- Receive command-line arguments
- Access standard I/O streams (stdin/stdout/stderr)
- Use system time and random number generation

## Configuration

WASI support can be enabled and configured through command-line flags:

### Basic WASI Enablement

```bash
functionfly-local --wasi-enabled
```

### Environment Variables

Pass environment variables to the WebAssembly module:

```bash
functionfly-local --wasi-env NODE_ENV=production --wasi-env API_URL=https://api.example.com
```

### Filesystem Access

Configure preopened directories for filesystem access:

```bash
functionfly-local --wasi-dirs /host/input:/input:r --wasi-dirs /host/output:/output:rw
```

Format: `host_path:wasm_path:permissions`

- `host_path`: Path on the host filesystem
- `wasm_path`: Path visible to the WebAssembly module
- `permissions`: `r` for read-only, `rw` for read-write

### Command Line Arguments

Pass arguments to the WebAssembly module:

```bash
functionfly-local --wasi-args config.json --wasi-args --verbose
```

### Network and Time Access

Control additional capabilities:

```bash
functionfly-local --wasi-allow-network --wasi-allow-time
```

## Implementation Details

### WASI Context

The runtime creates a WASI context (`WasiP1Ctx`) that provides the system interface to WebAssembly modules. This includes:

- **Environment Variables**: Custom variables plus standard ones (PATH, PWD, HOME)
- **Filesystem**: Preopened directories with configurable permissions
- **Arguments**: Program name plus configured arguments
- **I/O Streams**: stdin/stdout/stderr (currently simplified)

### Linker

A WASI linker connects WASI imports to their implementations, allowing WebAssembly modules to call system functions.

### Execution Flow

1. Check if WASI is enabled
2. Create WASI context with configuration
3. Set up linker with WASI functions
4. Instantiate WebAssembly module with WASI imports
5. Execute the module's entry point (`_start` or `main`)

## Example Usage

### Running a WASI-enabled Function

```bash
# Compile a Rust function to WASI
cargo build --target wasm32-wasi --release

# Run with WASI support
functionfly-local \
  --wasm target/wasm32-wasi/release/my_function.wasm \
  --wasi-enabled \
  --wasi-env DATABASE_URL=postgres://localhost \
  --wasi-dirs ./data:/data:rw \
  --wasi-args --config /data/config.json
```

### Node.js WASI Example

```javascript
// Compile JavaScript to WebAssembly with WASI support
// (Using tools like Javy or wasm-pack)

// The WebAssembly module can now:
// - Access environment variables: process.env.DATABASE_URL
// - Read/write files in preopened directories
// - Receive command line arguments
// - Use console.log (redirected to stdout)
```

## Security Considerations

- **Filesystem Access**: Only explicitly preopened directories are accessible
- **Environment Variables**: Only configured variables are exposed
- **Network Access**: Disabled by default for security
- **System Calls**: Limited to WASI-standard interfaces

## Limitations

- **Output Capture**: stdout/stderr capture is currently simplified
- **Advanced Permissions**: Complex permission schemes not fully implemented
- **Networking**: Network access control is basic
- **Real-time**: Some WASI features may have limitations in the local runtime

## Future Enhancements

- Full stdout/stderr capture and streaming
- Advanced filesystem permissions (per-file control)
- Network access controls and filtering
- Real-time monitoring of WASI calls
- Integration with FunctionFly's security features
