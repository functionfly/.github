//! External API host function implementation

use governor::Quota;
use once_cell::sync::Lazy;
use std::collections::HashMap;
use std::num::NonZeroU32;
use std::sync::Mutex;
use std::time::Duration;
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add the functionfly.external_api function for external API calls
pub fn add_external_api_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.external_api(method_ptr: i32, method_len: i32, url_ptr: i32, url_len: i32,
    //                          headers_ptr: i32, headers_len: i32, body_ptr: i32, body_len: i32,
    //                          response_ptr: i32, response_len_ptr: i32) -> i32
    // Returns 0 on success, negative values on error
    linker.func_wrap(
        "functionfly",
        "external_api",
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
            // Get HTTP method from WASM memory
            let method = match memory_utils::read_string_from_memory(&mut caller, method_ptr, method_len) {
                Ok(m) => m,
                Err(_) => return -1, // Invalid method
            };

            // Get URL from WASM memory
            let url = match memory_utils::read_string_from_memory(&mut caller, url_ptr, url_len) {
                Ok(u) => u,
                Err(_) => return -2, // Invalid URL
            };

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

            // Make external API request
            let result = make_external_api_request(&method, &url, &headers, body.as_deref(), &config);

            match result {
                Ok(response_body) => {
                    // Write response back to WASM memory
                    match memory_utils::write_string_to_memory(&mut caller, &response_body, response_ptr, response_len_ptr) {
                        Ok(_) => 0, // Success
                        Err(_) => -6, // Memory write error
                    }
                }
                Err(_) => -7, // External API request error
            }
        },
    )?;

    tracing::debug!("Added functionfly.external_api host function");
    Ok(())
}

/// Global rate limiter for external API calls
static EXTERNAL_API_RATE_LIMITER: Lazy<Mutex<governor::RateLimiter<governor::state::direct::NotKeyed, governor::state::InMemoryState, governor::clock::DefaultClock>>> = Lazy::new(|| {
    // Initialize with default rate limit, will be updated when first used
    Mutex::new(governor::RateLimiter::direct(Quota::per_minute(NonZeroU32::new(60).unwrap())))
});

/// Make an external API request with rate limiting
pub fn make_external_api_request(
    method: &str,
    url: &str,
    headers: &HashMap<String, String>,
    body: Option<&str>,
    config: &Config,
) -> anyhow::Result<String> {
    // Apply rate limiting
    {
        let limiter = EXTERNAL_API_RATE_LIMITER.lock().unwrap();
        limiter.check_n(NonZeroU32::new(1).unwrap())?;
    }

    // Create blocking HTTP client with timeout
    let client = reqwest::blocking::Client::builder()
        .timeout(Duration::from_secs(config.external_api_timeout_secs))
        .build()?;

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
        "External API call completed: {} {} -> {}",
        method, url, status
    );

    Ok(response_text)
}