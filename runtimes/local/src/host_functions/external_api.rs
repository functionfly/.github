//! External API host function implementation.
//!
//! Provides `functionfly.external_api` for WASM guests to call external HTTP
//! endpoints.  The following hardening improvements are applied over the original:
//!
//! - **Network policy**: the same whitelist / private-range checks that
//!   `functionfly.fetch` enforces are applied here.  Private/loopback addresses
//!   and unrecognised URL schemes are always blocked.  If a non-empty
//!   `strict_network_whitelist` is configured, the destination host must match.
//!
//! - **Per-function rate limiting**: the rate limiter is created *per linker
//!   registration* (i.e. per function instance) instead of globally.  This
//!   means separate functions each get their own quota, and one noisy function
//!   cannot starve another.

use std::collections::HashMap;
use std::collections::HashSet;
use std::num::NonZeroU32;
use std::sync::Arc;
use std::time::Duration;
use wasmtime_wasi::p1::WasiP1Ctx;

use governor::Quota;

use crate::config::Config;

use super::fetch::is_network_request_allowed;
use super::memory_utils;

/// Add `functionfly.external_api` to the linker.
///
/// Signature:
/// ```text
/// external_api(
///   method_ptr, method_len,       // HTTP verb
///   url_ptr,    url_len,          // target URL
///   headers_ptr, headers_len,     // JSON object of extra headers (pass 0,0 to omit)
///   body_ptr,    body_len,        // request body (pass 0,0 to omit)
///   response_ptr, response_len_ptr // output: response body
/// ) -> i32
/// ```
/// Returns:
///   `0`  — success
///   `-1` — method memory read error
///   `-2` — URL memory read error
///   `-3` — headers memory read error
///   `-4` — body memory read error
///   `-5` — invalid headers JSON
///   `-6` — response write error
///   `-7` — HTTP request error
///   `-8` — network blocked by policy
///   `-9` — rate limit exceeded
pub fn add_external_api_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // Per-function rate limiter — created once at registration, so quota is
    // isolated per function rather than shared globally.
    let rpm = NonZeroU32::new(config.external_api_rate_limit.max(1)).unwrap();
    let rate_limiter = Arc::new(governor::RateLimiter::direct(Quota::per_minute(rpm)));

    // Build whitelist once at registration time.
    let whitelist: HashSet<String> = config.network_whitelist.iter().cloned().collect();
    let strict_whitelist = config.strict_network_whitelist;

    // Create a persistent HTTP client for connection pooling.
    let http_client = Arc::new(
        reqwest::blocking::Client::builder()
            .timeout(Duration::from_secs(config.external_api_timeout_secs))
            .build()
            .unwrap_or_else(|_| reqwest::blocking::Client::new()),
    );

    let function_key = config.function_key();

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
            // Rate limit check (per-function quota)
            if rate_limiter
                .check_n(NonZeroU32::new(1).unwrap())
                .is_err()
            {
                tracing::warn!(
                    function = %function_key,
                    "external_api: rate limit exceeded ({} rpm)",
                    rpm
                );
                return -9;
            }

            let method =
                match memory_utils::read_string_from_memory(&mut caller, method_ptr, method_len) {
                    Ok(m) => m,
                    Err(_) => return -1,
                };

            let url =
                match memory_utils::read_string_from_memory(&mut caller, url_ptr, url_len) {
                    Ok(u) => u,
                    Err(_) => return -2,
                };

            // Network policy — same rules as functionfly.fetch
            if !is_network_request_allowed(&url, &whitelist, strict_whitelist) {
                tracing::warn!(
                    function = %function_key,
                    url = %url,
                    "external_api: network request blocked by policy"
                );
                return -8;
            }

            let headers_json = if headers_len > 0 {
                match memory_utils::read_string_from_memory(
                    &mut caller,
                    headers_ptr,
                    headers_len,
                ) {
                    Ok(h) => Some(h),
                    Err(_) => return -3,
                }
            } else {
                None
            };

            let body = if body_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, body_ptr, body_len) {
                    Ok(b) => Some(b),
                    Err(_) => return -4,
                }
            } else {
                None
            };

            let headers: HashMap<String, String> = match headers_json {
                Some(json) => match serde_json::from_str(&json) {
                    Ok(h) => h,
                    Err(_) => return -5,
                },
                None => HashMap::new(),
            };

            let result =
                make_external_api_request(&http_client, &method, &url, &headers, body.as_deref());

            match result {
                Ok(response_body) => {
                    match memory_utils::write_string_to_memory(
                        &mut caller,
                        &response_body,
                        response_ptr,
                        response_len_ptr,
                    ) {
                        Ok(_) => 0,
                        Err(_) => -6,
                    }
                }
                Err(_) => -7,
            }
        },
    )?;

    tracing::debug!("Added functionfly.external_api host function");
    Ok(())
}

/// Execute the external API HTTP request.
pub fn make_external_api_request(
    client: &reqwest::blocking::Client,
    method: &str,
    url: &str,
    headers: &HashMap<String, String>,
    body: Option<&str>,
) -> anyhow::Result<String> {
    let mut request_builder = match method.to_uppercase().as_str() {
        "GET" => client.get(url),
        "POST" => client.post(url),
        "PUT" => client.put(url),
        "DELETE" => client.delete(url),
        "PATCH" => client.patch(url),
        "HEAD" => client.head(url),
        "OPTIONS" => client.request(reqwest::Method::OPTIONS, url),
        other => return Err(anyhow::anyhow!("Unsupported HTTP method: {}", other)),
    };

    for (key, value) in headers {
        request_builder = request_builder.header(key, value);
    }

    if let Some(body_content) = body {
        request_builder = request_builder.body(body_content.to_string());
    }

    let response = request_builder.send()?;
    let status = response.status();
    let response_text = response.text()?;

    tracing::info!("external_api: {} {} -> {}", method, url, status);

    Ok(response_text)
}
