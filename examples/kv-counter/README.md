# KV Counter Example

This example demonstrates how to use the KV (key-value) store capability in FunctionFly.

## Overview

The KV capability allows functions to store and retrieve key-value pairs with optional TTL (time-to-live) expiration. This is useful for:

- Caching frequently accessed data
- Storing user sessions or state
- Implementing rate limiting
- Temporary data storage between function invocations

## Capabilities Required

To use KV storage, your function must declare the `"kv"` capability in its `functionfly.jsonc`:

```json
{
  "capabilities": ["kv"]
}
```

## KV API

When the KV capability is enabled, the following functions are available in your WASM module:

### `kv_get(key_ptr, key_len, value_ptr, value_len_ptr) -> i32`

Retrieves a value from the KV store.

- **Parameters:**
  - `key_ptr`: Pointer to the key string in WASM memory
  - `key_len`: Length of the key string
  - `value_ptr`: Pointer to buffer for the value in WASM memory
  - `value_len_ptr`: Pointer to store the value length

- **Returns:**
  - `0`: Success
  - `-1`: Key not found
  - `-2`: Invalid key
  - `-3`: Memory write error

### `kv_set(key_ptr, key_len, value_ptr, value_len, ttl_seconds) -> i32`

Stores a value in the KV store.

- **Parameters:**
  - `key_ptr`: Pointer to the key string in WASM memory
  - `key_len`: Length of the key string
  - `value_ptr`: Pointer to the value string in WASM memory
  - `value_len`: Length of the value string
  - `ttl_seconds`: TTL in seconds (-1 = no TTL, 0 = delete)

- **Returns:**
  - `0`: Success
  - `-2`: Invalid key
  - `-3`: Invalid value
  - `-4`: Delete failed
  - `-5`: Invalid TTL
  - `-6`: Storage error

### `kv_exists(key_ptr, key_len) -> i32`

Checks if a key exists in the KV store.

- **Parameters:**
  - `key_ptr`: Pointer to the key string in WASM memory
  - `key_len`: Length of the key string

- **Returns:**
  - `1`: Key exists
  - `0`: Key does not exist
  - `-1`: Invalid key

## Usage Example

Here's how you might implement a visit counter in Rust:

```rust
#[no_mangle]
pub extern "C" fn handler() -> i32 {
    let key = "visit_count";

    // Get current count
    let mut count = 0;
    let mut value_buf = [0u8; 32];
    let mut value_len = 0;

    let result = unsafe {
        kv_get(
            key.as_ptr() as i32,
            key.len() as i32,
            value_buf.as_mut_ptr() as i32,
            &mut value_len as *mut i32,
        )
    };

    if result == 0 {
        // Parse existing count
        let value_str = std::str::from_utf8(&value_buf[..value_len as usize]).unwrap_or("0");
        count = value_str.parse().unwrap_or(0);
    }

    // Increment count
    count += 1;
    let new_value = count.to_string();

    // Store new count (no TTL)
    let result = unsafe {
        kv_set(
            key.as_ptr() as i32,
            key.len() as i32,
            new_value.as_ptr() as i32,
            new_value.len() as i32,
            -1, // no TTL
        )
    };

    if result == 0 {
        // Success - return the count
        count
    } else {
        // Error
        -1
    }
}
```

## Limitations

- Keys and values are stored as UTF-8 strings
- Maximum key/value size is limited by WASM memory constraints
- KV store is shared across all functions (use unique keys)
- Data is stored in memory only (lost on restart)
- Maximum 10,000 entries by default (configurable)