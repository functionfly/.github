//! Webhook host function implementation

use std::collections::HashMap;
use wasmtime_wasi::p1::WasiP1Ctx;

use super::memory_utils;

/// Add the webhook_send function to the linker
/// This function is available in both "env" and "functionfly" namespaces for compatibility
///
/// # Arguments
/// * `linker` - The WASM linker to add the function to
/// * `capability_granted` - Whether the webhook capability is granted. If false, the function
///   will still be added (so WASM modules can instantiate), but will return an error at runtime.
pub fn add_webhook_function(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
    capability_granted: bool,
) -> anyhow::Result<()> {
    // functionfly.webhook_send(url_ptr: i32, url_len: i32, method_ptr: i32, method_len: i32,
    //                          payload_ptr: i32, payload_len: i32, headers_ptr: i32, headers_len: i32) -> i32
    // Returns 0 on success, negative values on error

    // Add to "env" namespace for WASM modules that import from "env"
    linker.func_wrap(
        "env",
        "webhook_send",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              url_ptr: i32,
              url_len: i32,
              method_ptr: i32,
              method_len: i32,
              payload_ptr: i32,
              payload_len: i32,
              headers_ptr: i32,
              headers_len: i32| -> i32 {
            if !capability_granted {
                tracing::warn!("webhook_send called but webhook capability not granted");
                return -7; // Capability not granted
            }
            execute_webhook(&mut caller, url_ptr, url_len, method_ptr, method_len,
                           payload_ptr, payload_len, headers_ptr, headers_len)
        },
    )?;

    // Also add to "functionfly" namespace for newer modules
    linker.func_wrap(
        "functionfly",
        "webhook_send",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              url_ptr: i32,
              url_len: i32,
              method_ptr: i32,
              method_len: i32,
              payload_ptr: i32,
              payload_len: i32,
              headers_ptr: i32,
              headers_len: i32| -> i32 {
            if !capability_granted {
                tracing::warn!("webhook_send called but webhook capability not granted");
                return -7; // Capability not granted
            }
            execute_webhook(&mut caller, url_ptr, url_len, method_ptr, method_len,
                           payload_ptr, payload_len, headers_ptr, headers_len)
        },
    )?;

    tracing::debug!("Added webhook_send function to WASM linker (env and functionfly namespaces), capability_granted={}", capability_granted);
    Ok(())
}

/// Execute the webhook request
fn execute_webhook(
    caller: &mut wasmtime::Caller<WasiP1Ctx>,
    url_ptr: i32,
    url_len: i32,
    method_ptr: i32,
    method_len: i32,
    payload_ptr: i32,
    payload_len: i32,
    headers_ptr: i32,
    headers_len: i32,
) -> i32 {
    // Get the URL from WASM memory
    let url = match memory_utils::read_string_from_memory(caller, url_ptr, url_len) {
        Ok(u) => u,
        Err(_) => return -1, // Invalid URL
    };

    // Get the HTTP method from WASM memory
    let method = match memory_utils::read_string_from_memory(caller, method_ptr, method_len) {
        Ok(m) => m,
        Err(_) => return -2, // Invalid method
    };

    // Get the payload from WASM memory (can be empty for GET requests)
    let payload = if payload_len > 0 {
        match memory_utils::read_string_from_memory(caller, payload_ptr, payload_len) {
            Ok(p) => Some(p),
            Err(_) => return -3, // Invalid payload
        }
    } else {
        None
    };

    // Get the headers from WASM memory (JSON string)
    let headers_json = if headers_len > 0 {
        match memory_utils::read_string_from_memory(caller, headers_ptr, headers_len) {
            Ok(h) => Some(h),
            Err(_) => return -4, // Invalid headers
        }
    } else {
        None
    };

    // Parse headers if provided
    let headers: HashMap<String, String> = if let Some(json) = headers_json {
        match serde_json::from_str(&json) {
            Ok(h) => h,
            Err(_) => return -5, // Invalid headers JSON
        }
    } else {
        HashMap::new()
    };

    // Send webhook using blocking operation
    let result = tokio::task::block_in_place(|| {
        tokio::runtime::Handle::current().block_on(async {
            send_webhook_request(&url, &method, payload.as_deref(), &headers).await
        })
    });

    match result {
        Ok(_) => 0, // Success
        Err(_) => -6, // Request failed
    }
}

/// Send an HTTP request for webhook functionality
async fn send_webhook_request(
    url: &str,
    method: &str,
    payload: Option<&str>,
    headers: &HashMap<String, String>,
) -> anyhow::Result<()> {
    // Create HTTP client
    let client = reqwest::Client::new();

    // Build request
    let mut request_builder = match method.to_uppercase().as_str() {
        "GET" => client.get(url),
        "POST" => client.post(url),
        "PUT" => client.put(url),
        "PATCH" => client.patch(url),
        "DELETE" => client.delete(url),
        _ => return Err(anyhow::anyhow!("Unsupported HTTP method: {}", method)),
    };

    // Add headers
    for (key, value) in headers {
        request_builder = request_builder.header(key, value);
    }

    // Add payload if provided
    if let Some(payload) = payload {
        request_builder = request_builder
            .header("Content-Type", "application/json")
            .body(payload.to_string());
    }

    // Send request with timeout
    let response = request_builder
        .timeout(std::time::Duration::from_secs(30))
        .send()
        .await?;

    // Check if response is successful
    if !response.status().is_success() {
        return Err(anyhow::anyhow!("Webhook request failed with status: {}", response.status()));
    }

    tracing::debug!("Webhook sent successfully to {} with method {}", url, method);
    Ok(())
}
