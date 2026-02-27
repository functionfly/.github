# Webhook Notifier Example

This example demonstrates how to use the webhook capability in FunctionFly to send HTTP requests to external services.

## Overview

The webhook capability allows functions to make outbound HTTP requests to trigger webhooks, send notifications, or integrate with external APIs. This is useful for:

- Sending notifications to external services
- Triggering workflows in other systems
- Integrating with webhooks from services like Slack, Discord, or custom APIs
- Sending data to monitoring or logging services

## Capabilities Required

To use webhook functionality, your function must declare the `"webhook"` capability in its `functionfly.jsonc`:

```json
{
  "capabilities": ["webhook"]
}
```

## Webhook API

When the webhook capability is enabled, the following function is available in your WASM module:

### `webhook_send(url_ptr, url_len, method_ptr, method_len, payload_ptr, payload_len, headers_ptr, headers_len) -> i32`

Sends an HTTP request to a webhook endpoint.

- **Parameters:**
  - `url_ptr`: Pointer to the URL string in WASM memory
  - `url_len`: Length of the URL string
  - `method_ptr`: Pointer to the HTTP method string (GET, POST, PUT, PATCH, DELETE)
  - `method_len`: Length of the method string
  - `payload_ptr`: Pointer to the request payload string (can be 0 for GET requests)
  - `payload_len`: Length of the payload string (can be 0)
  - `headers_ptr`: Pointer to JSON string with headers (can be 0 for no custom headers)
  - `headers_len`: Length of the headers JSON string (can be 0)

- **Returns:**
  - `0`: Success
  - `-1`: Invalid URL
  - `-2`: Invalid HTTP method
  - `-3`: Invalid payload
  - `-4`: Invalid headers
  - `-5`: Failed to parse headers JSON
  - `-6`: HTTP request failed

## Usage Example

Here's how to send a webhook notification:

```rust
#[no_mangle]
pub extern "C" fn handler() -> i32 {
    let url = "https://api.example.com/webhook";
    let method = "POST";
    let payload = r#"{"message": "Hello from FunctionFly!"}"#;
    let headers = r#"{"Content-Type": "application/json"}"#;

    let result = unsafe {
        webhook_send(
            url.as_ptr() as i32,
            url.len() as i32,
            method.as_ptr() as i32,
            method.len() as i32,
            payload.as_ptr() as i32,
            payload.len() as i32,
            headers.as_ptr() as i32,
            headers.len() as i32,
        )
    };

    if result == 0 {
        // Success
        0
    } else {
        // Error
        -1
    }
}
```

## Headers Format

Headers should be provided as a JSON object:

```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer token123",
  "X-Custom-Header": "value"
}
```

## Supported HTTP Methods

- GET
- POST
- PUT
- PATCH
- DELETE

## Limitations

- Request timeout is fixed at 30 seconds
- Maximum payload size is 1MB
- Only HTTPS URLs are allowed (HTTP may be blocked by runtime configuration)
- Custom headers are supported but may be filtered for security

## Testing

To test this example:

1. Build the function:
   ```bash
   cargo build --release --target wasm32-wasi
   ```

2. Deploy to FunctionFly:
   ```bash
   fly deploy
   ```

3. Invoke the function:
   ```bash
   curl -X POST https://your-function-url
   ```

The function will send a POST request to https://httpbin.org/post with a JSON payload.