# StateFabric WASM Integration

This document describes the Wasmtime integration for StateFabric, enabling WebAssembly functions to execute with shared memory and commit events to the event log.

## Overview

The WASM integration provides:

- **Shared Memory**: 1MB buffer for host-WASM communication
- **CommitEvent API**: WASM functions can commit events to the event log
- **State Access**: Read/write access to state data
- **Deterministic Execution**: Configurable deterministic mode
- **Gas Metering**: Optional execution cost tracking

## Host Functions Available to WASM

### CommitEvent API

```rust
// Commit an event to the event log
fn commit_event(
    event_type: i32,    // 0=SET, 1=DELETE, 2=MERGE
    key_ptr: i32,       // Pointer to key string in memory
    key_len: i32,       // Length of key string
    value_ptr: i32,     // Pointer to JSON value string
    value_len: i32      // Length of value string
) -> i32; // Returns 0 on success, negative on error
```

### State Access Functions

```rust
// Get state value
fn get_state(
    key_ptr: i32,       // Pointer to key string
    key_len: i32,       // Length of key string
    output_ptr: i32,    // Output buffer pointer
    max_len: i32        // Maximum output length
) -> i32; // Returns bytes written, negative on error

// Set state value (temporary, doesn't create event)
fn set_state(
    key_ptr: i32,       // Pointer to key string
    key_len: i32,       // Length of key string
    value_ptr: i32,     // Pointer to JSON value string
    value_len: i32      // Length of value string
) -> i32; // Returns 0 on success, negative on error
```

## API Endpoints

### Load WASM Module

```http
POST /v1/wasm/modules
Content-Type: application/json

{
  "name": "my_module",
  "wasm_bytes": [/* base64 encoded WASM bytes */]
}
```

### Execute WASM Function

```http
POST /v1/state/{state_id}/execute
Content-Type: application/json

{
  "module_name": "my_module",
  "function_name": "process_data",
  "input": [/* input bytes */]
}
```

Response:
```json
{
  "success": true,
  "output": [/* output bytes */],
  "committed_events": [
    {
      "id": "event-uuid",
      "state_id": "state-uuid",
      "event_type": "set",
      "sequence": 123,
      "timestamp": "2026-02-20T..."
    }
  ],
  "gas_used": null,
  "execution_time_ms": 5
}
```

## Example WASM Module

See `example.wat` for a complete WebAssembly Text example that demonstrates:

- Reading state values
- Committing SET events
- Committing MERGE events
- Multiple operations in sequence

## Building WASM Modules

### From Rust

```rust
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
extern "C" {
    fn commit_event(event_type: i32, key_ptr: i32, key_len: i32, value_ptr: i32, value_len: i32) -> i32;
    fn get_state(key_ptr: i32, key_len: i32, output_ptr: i32, max_len: i32) -> i32;
}

#[wasm_bindgen]
pub fn my_function() -> i32 {
    // Your logic here
    unsafe {
        commit_event(0, /* ... */);
    }
    0
}
```

### From WAT (WebAssembly Text)

Use the `wat2wasm` tool:

```bash
wat2wasm example.wat -o example.wasm
```

## Configuration

Configure WASM runtime in your StateManager:

```rust
use statefabric::{StateManager, WasmConfig};

let wasm_config = WasmConfig {
    max_memory_pages: 256,     // 16MB
    deterministic: true,       // Deterministic execution
    enable_gas: false,         // Gas metering disabled
    max_execution_time_ms: 5000, // 5 second timeout
};

let state_manager = StateManager::with_wasm(
    object_store,
    snapshot_repo,
    event_repo,
    wasm_config
)?;
```

## Event Types

- `0`: SET - Set a key-value pair
- `1`: DELETE - Delete a key
- `2`: MERGE - Merge object values

## Error Codes

Host functions return negative values on error:

- `-1`: Invalid state context
- `-2`: Memory read error
- `-3`: Invalid JSON
- `-4`: Serialization error
- `-5`: Invalid event type
- `-6`: Event commit failed

## Security Considerations

- WASM modules run in isolated memory space
- Execution is sandboxed with configurable limits
- Deterministic mode ensures reproducible results
- Gas metering available for resource control
- Timeout protection prevents infinite loops

## Performance

- JIT compilation for fast execution
- Shared memory avoids serialization overhead
- Event batching for atomic operations
- Configurable memory and execution limits