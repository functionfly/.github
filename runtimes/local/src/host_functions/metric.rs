//! Custom metrics host function implementation.
//!
//! Provides `functionfly.emit_metric` — lets WASM guests emit named
//! floating-point gauge values that are:
//!   1. Recorded into the global Prometheus registry (scraped at `/metrics`).
//!   2. Logged via `tracing::debug!` for structured log-based observability.
//!
//! ## Host function signature
//! ```text
//! emit_metric(
//!   name_ptr:   i32,  // metric name (alphanumeric + underscores)
//!   name_len:   i32,
//!   value:      f64,  // gauge value
//!   labels_ptr: i32,  // optional JSON object {"key":"value",...} — pass 0 to omit
//!   labels_len: i32,
//! ) -> i32
//! ```
//! Returns:
//!   `0`  — success
//!   `-1` — invalid name (memory read error)
//!   `-2` — invalid name characters (not alphanumeric/_)
//!   `-3` — invalid labels JSON
//!   `-4` — labels memory read error

use std::collections::HashMap;

use once_cell::sync::Lazy;
use prometheus::{GaugeVec, Opts};
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Global Prometheus gauge for all custom metrics emitted from WASM functions.
///
/// Labels: `function` (name@version), `metric` (name declared by guest), `labels`
/// (flattened sorted "k=v,k=v" string from the optional JSON object).
static CUSTOM_METRICS: Lazy<GaugeVec> = Lazy::new(|| {
    let opts = Opts::new(
        "functionfly_custom_metric",
        "Custom gauge metric emitted from a WASM function via functionfly.emit_metric",
    );
    GaugeVec::new(opts, &["function", "metric", "labels"])
        .expect("Failed to register functionfly_custom_metric gauge")
});

/// Register the gauge with the default Prometheus registry once.
fn ensure_registered() {
    static REGISTERED: std::sync::Once = std::sync::Once::new();
    REGISTERED.call_once(|| {
        if let Err(e) = prometheus::register(Box::new(CUSTOM_METRICS.clone())) {
            // Already registered is fine; any other error is a programming bug.
            if !e.to_string().contains("already registered") {
                tracing::error!("Failed to register custom metrics gauge: {}", e);
            }
        }
    });
}

/// Add `functionfly.emit_metric` to the linker.
pub fn add_emit_metric_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    ensure_registered();

    let function_key = config.function_key();

    linker.func_wrap(
        "functionfly",
        "emit_metric",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              name_ptr: i32,
              name_len: i32,
              value: f64,
              labels_ptr: i32,
              labels_len: i32| -> i32 {
            let name =
                match memory_utils::read_string_from_memory(&mut caller, name_ptr, name_len) {
                    Ok(n) => n,
                    Err(_) => return -1,
                };

            // Validate: alphanumeric + underscores only
            if name.is_empty()
                || !name.chars().all(|c| c.is_alphanumeric() || c == '_')
            {
                tracing::warn!(
                    function = %function_key,
                    name = %name,
                    "emit_metric: invalid metric name"
                );
                return -2;
            }

            // Parse optional labels JSON object
            let label_tag = if labels_len > 0 {
                match memory_utils::read_string_from_memory(
                    &mut caller,
                    labels_ptr,
                    labels_len,
                ) {
                    Err(_) => return -4,
                    Ok(json) => {
                        match serde_json::from_str::<HashMap<String, String>>(&json) {
                            Err(_) => return -3,
                            Ok(map) => {
                                // Stable sorted "k=v,k=v" tag
                                let mut pairs: Vec<String> =
                                    map.iter().map(|(k, v)| format!("{}={}", k, v)).collect();
                                pairs.sort();
                                pairs.join(",")
                            }
                        }
                    }
                }
            } else {
                String::new()
            };

            // Record into Prometheus
            CUSTOM_METRICS
                .with_label_values(&[&function_key, &name, &label_tag])
                .set(value);

            tracing::debug!(
                function = %function_key,
                metric  = %name,
                value   = value,
                labels  = %label_tag,
                "functionfly.emit_metric"
            );

            0
        },
    )?;

    tracing::debug!("Added functionfly.emit_metric host function");
    Ok(())
}
