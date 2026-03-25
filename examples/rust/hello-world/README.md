# Hello World - Rust Example

This is a simple Hello World example demonstrating how to write Rust functions for the FunctionFly platform.

## Prerequisites

1. **Rust toolchain** - Install from <https://rustup.rs/>
2. **wasm32-wasi target** - Add the WASI target:

   ```bash
   rustup target add wasm32-wasi
   ```

## Compilation

To compile this function to WebAssembly:

```bash
# Navigate to the example directory
cd examples/rust/hello-world

# Build for wasm32-wasi
cargo build --target wasm32-wasi --release

# The output will be in target/wasm32-wasi/release/hello_world.wasm
```

## Deployment

After compilation, you can deploy the WASM file to FunctionFly:

```bash
# Using the fly CLI (when Rust compile command is available)
fly deploy --wasm ./target/wasm32-wasi/release/hello_world.wasm

# Or manually upload the WASM file through the dashboard
```

## How It Works

### Handler Function

The main handler is the `handler()` function marked with `#[no_mangle]` and `pub extern "C"`:

```rust
#[no_mangle]
pub extern "C" fn handler() -> i32 {
    // Your logic here
    println!("{}", response);
    0 // Return 0 for success
}
```

### Input/Output

- **Input**: The FunctionFly runtime passes JSON input via stdin
- **Output**: Write JSON response to stdout
- **Errors**: Write error messages to stderr and return non-zero

### Return Codes

| Code | Description |
|------|-------------|
| 0    | Success |
| -1   | General error |
| -2   | Invalid input |
| -3   | Timeout |

## Example Input/Output

**Input** (via stdin):

```json
{"name": "FunctionFly"}
```

**Output** (via stdout):

```json
{"message": "Hello, FunctionFly! Welcome to FunctionFly Rust runtime."}
```

## Adding Dependencies

If you need external crates, add them to `Cargo.toml`:

```toml
[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
```

Note: Only WASI-compatible crates will work in the FunctionFly runtime.

## More Examples

See also:

- [`examples/webhook-notifier/`](../../webhook-notifier/) - More complex example with HTTP calls
