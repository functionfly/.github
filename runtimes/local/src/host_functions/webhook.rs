//! Webhook host function implementation.
//!
//! Provides `functionfly.webhook_send` (and the legacy `env.webhook_send` alias)
//! for WASM guests to deliver HTTP webhook payloads.
//!
//! ## Security
//! The same network policy enforced by `functionfly.fetch` is applied here:
//! - Only `http`/`https` schemes are permitted.
//! - Requests to localhost, loopback, and RFC 1918 private ranges are blocked.
//! - When `strict_network_whitelist` is enabled the destination host must appear
//!   in `network_whitelist`.

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::fetch::is_network_request_allowed;
use super::memory_utils;

/// Add `functionfly.webhook_send` (and the legacy `env.webhook_send` alias) to
/// the linker.
///
/// `capability_granted` — if `false` the function is still registered (so WASM
/// modules can instantiate) but returns `-7` at runtime.
pub fn add_webhook_function(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
    capability_granted: bool,
    config: Config,
) -> anyhow::Result<()> {
    let whitelist: Arc<HashSet<String>> =
        Arc::new(config.network_whitelist.iter().cloned().collect());
    let strict_whitelist = config.strict_network_whitelist;

    // `env` namespace — legacy alias
    {
        let wl = whitelist.clone();
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
                    return -7;
                }
                execute_webhook(
                    &mut caller,
                    url_ptr,
                    url_len,
                    method_ptr,
                    method_len,
                    payload_ptr,
                    payload_len,
                    headers_ptr,
                    headers_len,
                    &wl,
                    strict_whitelist,
                )
            },
        )?;
    }

    // `functionfly` namespace — preferred
    {
        let wl = whitelist.clone();
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
                    return -7;
                }
                execute_webhook(
                    &mut caller,
                    url_ptr,
                    url_len,
                    method_ptr,
                    method_len,
                    payload_ptr,
                    payload_len,
                    headers_ptr,
                    headers_len,
                    &wl,
                    strict_whitelist,
                )
            },
        )?;
    }

    tracing::debug!(
        "Added webhook_send host functions (env + functionfly), capability_granted={}",
        capability_granted
    );
    Ok(())
}

#[allow(clippy::too_many_arguments)]
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
    whitelist: &HashSet<String>,
    strict_whitelist: bool,
) -> i32 {
    let url = match memory_utils::read_string_from_memory(caller, url_ptr, url_len) {
        Ok(u) => u,
        Err(_) => return -1,
    };

    // Network policy — same rules as functionfly.fetch
    if !is_network_request_allowed(&url, whitelist, strict_whitelist) {
        tracing::warn!("webhook_send: destination blocked by network policy: {}", url);
        return -8;
    }

    let method = match memory_utils::read_string_from_memory(caller, method_ptr, method_len) {
        Ok(m) => m,
        Err(_) => return -2,
    };

    let payload = if payload_len > 0 {
        match memory_utils::read_string_from_memory(caller, payload_ptr, payload_len) {
            Ok(p) => Some(p),
            Err(_) => return -3,
        }
    } else {
        None
    };

    let headers_json = if headers_len > 0 {
        match memory_utils::read_string_from_memory(caller, headers_ptr, headers_len) {
            Ok(h) => Some(h),
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

    let result = tokio::task::block_in_place(|| {
        tokio::runtime::Handle::current()
            .block_on(async { send_webhook_request(&url, &method, payload.as_deref(), &headers).await })
    });

    match result {
        Ok(_) => 0,
        Err(_) => -6,
    }
}

async fn send_webhook_request(
    url: &str,
    method: &str,
    payload: Option<&str>,
    headers: &HashMap<String, String>,
) -> anyhow::Result<()> {
    let client = reqwest::Client::new();

    let mut request_builder = match method.to_uppercase().as_str() {
        "GET" => client.get(url),
        "POST" => client.post(url),
        "PUT" => client.put(url),
        "PATCH" => client.patch(url),
        "DELETE" => client.delete(url),
        other => return Err(anyhow::anyhow!("Unsupported HTTP method: {}", other)),
    };

    for (key, value) in headers {
        request_builder = request_builder.header(key, value);
    }

    if let Some(payload) = payload {
        request_builder = request_builder
            .header("Content-Type", "application/json")
            .body(payload.to_string());
    }

    let response = request_builder
        .timeout(std::time::Duration::from_secs(30))
        .send()
        .await?;

    if !response.status().is_success() {
        return Err(anyhow::anyhow!(
            "Webhook request failed with status: {}",
            response.status()
        ));
    }

    tracing::debug!("webhook_send: delivered to {} [{}]", url, method);
    Ok(())
}
