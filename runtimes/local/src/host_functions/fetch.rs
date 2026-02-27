//! HTTP fetch host function implementation

use std::collections::HashMap;
use std::sync::Arc;
use wasmtime_wasi::p1::WasiP1Ctx;

use super::memory_utils;
use crate::security::SecurityMonitor;

/// Add the functionfly.fetch function for HTTP requests
pub fn add_fetch_function(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
    security_monitor: Arc<SecurityMonitor>,
) -> anyhow::Result<()> {
    // functionfly.fetch(method_ptr: i32, method_len: i32, url_ptr: i32, url_len: i32,
    //                   headers_ptr: i32, headers_len: i32, body_ptr: i32, body_len: i32,
    //                   response_ptr: i32, response_len_ptr: i32) -> i32
    // Returns 0 on success, negative values on error
    let security_monitor_clone = security_monitor.clone();
    linker.func_wrap(
        "functionfly",
        "fetch",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              method_ptr: i32,
              method_len: i32,
              url_ptr: i32,
              url_len: i32,
              headers_ptr: i32,
              headers_len: i32,
              body_ptr: i32,
              body_len: i32,
              response_ptr: i32,
              response_len_ptr: i32| -> i32 {
            // Get the HTTP method from WASM memory
            let method = match memory_utils::read_string_from_memory(&mut caller, method_ptr, method_len) {
                Ok(m) => m,
                Err(_) => return -1, // Invalid method
            };

            // Get the URL from WASM memory
            let url = match memory_utils::read_string_from_memory(&mut caller, url_ptr, url_len) {
                Ok(u) => u,
                Err(_) => return -2, // Invalid URL
            };

            // For now, we'll implement a basic network check
            // TODO: Pass function key through execution context for proper per-function whitelisting
            // Check if the URL is allowed (basic validation)
            if !is_network_request_allowed(&url) {
                tracing::warn!("Network request blocked: {}", url);
                return -8; // Network access denied
            }

            // Get headers (JSON string, optional)
            let headers_json = if headers_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, headers_ptr, headers_len) {
                    Ok(h) => Some(h),
                    Err(_) => return -3, // Invalid headers
                }
            } else {
                None
            };

            // Get request body (optional)
            let body = if body_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, body_ptr, body_len) {
                    Ok(b) => Some(b),
                    Err(_) => return -4, // Invalid body
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

            // Make HTTP request
            let result = make_http_request(&method, &url, &headers, body.as_deref());

            match result {
                Ok(response_body) => {
                    // Write response back to WASM memory
                    match memory_utils::write_string_to_memory(&mut caller, &response_body, response_ptr, response_len_ptr) {
                        Ok(_) => 0, // Success
                        Err(_) => -6, // Memory write error
                    }
                }
                Err(_) => -7, // HTTP request error
            }
        },
    )?;

    tracing::debug!("Added functionfly.fetch host function");
    Ok(())
}

/// Check if a network request is allowed
fn is_network_request_allowed(url: &str) -> bool {
    // Parse the URL to extract domain
    if let Ok(parsed_url) = url::Url::parse(url) {
        let host = parsed_url.host_str().unwrap_or("");

        // Basic security checks
        // - No localhost/private IPs in production
        // - No suspicious schemes
        match parsed_url.scheme() {
            "http" | "https" => {
                // Allow common public domains
                // TODO: Make this configurable via enterprise config
                let allowed_domains = [
                    "api.github.com",
                    "httpbin.org",
                    "jsonplaceholder.typicode.com",
                    "api.example.com", // Example for testing
                ];

                // Check if host matches allowed domains or is a reasonable public domain
                allowed_domains.contains(&host) ||
                (!host.contains("localhost") &&
                 !host.contains("127.0.0.1") &&
                 !host.contains("0.0.0.0") &&
                 !host.starts_with("10.") &&
                 !host.starts_with("192.168.") &&
                 !host.starts_with("172."))
            }
            _ => false, // Only allow HTTP/HTTPS
        }
    } else {
        false // Invalid URL
    }
}

/// Make an HTTP request
fn make_http_request(
    method: &str,
    url: &str,
    headers: &HashMap<String, String>,
    body: Option<&str>,
) -> anyhow::Result<String> {
    // Create blocking HTTP client
    let client = reqwest::blocking::Client::new();

    // Build the request
    let mut request_builder = match method.to_uppercase().as_str() {
        "GET" => client.get(url),
        "POST" => client.post(url),
        "PUT" => client.put(url),
        "DELETE" => client.delete(url),
        "PATCH" => client.patch(url),
        "HEAD" => client.head(url),
        "OPTIONS" => client.request(reqwest::Method::OPTIONS, url),
        _ => return Err(anyhow::anyhow!("Unsupported HTTP method: {}", method)),
    };

    // Add headers
    for (key, value) in headers {
        request_builder = request_builder.header(key, value);
    }

    // Add body if provided
    if let Some(body_content) = body {
        request_builder = request_builder.body(body_content.to_string());
    }

    // Send the request
    let response = request_builder.send()?;
    let status = response.status();
    let response_text = response.text()?;

    tracing::info!(
        "HTTP request completed: {} {} -> {}",
        method, url, status
    );

    Ok(response_text)
}