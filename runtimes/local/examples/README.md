# WASI Examples for FunctionFly

This directory contains examples demonstrating WASI (WebAssembly System Interface) functionality in FunctionFly.

## Prerequisites

- Rust with `wasm32-wasi` target: `rustup target add wasm32-wasi`
- FunctionFly local runtime built with WASI support

## Building Examples

```bash
./build.sh
```

This will create `wasi_example.wasm` in the current directory.

## Running Examples

### Basic WASI Function

```bash
functionfly-local \
  --wasm wasi_example.wasm \
  --wasi-enabled
```

### With Environment Variables

```bash
functionfly-local \
  --wasm wasi_example.wasm \
  --wasi-enabled \
  --wasi-env NODE_ENV=production \
  --wasi-env API_URL=https://api.example.com
```

### With Filesystem Access

```bash
functionfly-local \
  --wasm wasi_example.wasm \
  --wasi-enabled \
  --wasi-dirs /tmp:/tmp:rw
```

### With Arguments

```bash
functionfly-local \
  --wasm wasi_example.wasm \
  --wasi-enabled \
  --wasi-args --verbose \
  --wasi-args config.json
```

## Example Output

When run successfully, the WASI example will output:

```
Hello from WASI-enabled WebAssembly!
Environment NODE_ENV: production
Arguments: ["wasi_example", "--verbose", "config.json"]
Contents of /tmp:
  (directory listing...)
Successfully wrote to /tmp/wasi-test.txt
WASI example completed!
```

## Troubleshooting

### "Directory not preopened" errors

- Add `--wasi-dirs /host/path:/wasm/path:rw` to allow filesystem access
- Use appropriate permissions (`r` for read-only, `rw` for read-write)

### Environment variables not available

- Use `--wasi-env KEY=VALUE` to pass environment variables
- Check that the variable names match exactly

### Permission denied errors

- Ensure preopened directories have the correct permissions
- For write access, use `:rw` instead of `:r`

## WASI Features Demonstrated

- **Environment Variables**: Accessing host environment variables
- **Command Line Arguments**: Receiving arguments passed to the module
- **Filesystem I/O**: Reading directories and writing files
- **Standard Output**: Using `println!` for logging
- **Error Handling**: Graceful handling of missing permissions/capabilities
