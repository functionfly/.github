//! Metrics, metadata, and execution-result handlers.

use axum::{extract::{State, Json}, response::IntoResponse};
use std::sync::Arc;

use super::types::AppState;
use crate::wasm_interface::{ExecutionResult, FunctionMetadata};
use crate::monitoring::ExecutionMetrics;
use crate::resource_enforcer::{GlobalResourceLimits, QuotaType, ResourceQuota};
use crate::host_functions::storage::{validate_storage_path, validate_storage_path_for_write};
use crate::package::{parse_requirement_name, PackageCacheStats};

// ---------------------------------------------------------------------------
// Execution metrics
// ---------------------------------------------------------------------------

/// Execution metrics handler — shows ExecutionMetrics usage.
pub async fn execution_metrics(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let stats = state.monitor.get_stats().await;

    let example_metrics = ExecutionMetrics {
        function_name: state.config.function.clone(),
        function_version: state.config.version.clone(),
        execution_time_ms: stats.average_execution_time_ms as u64,
        cpu_fuel_used: state.config.cpu_fuel_limit,
        memory_used_mb: stats.average_execution_time_ms as f64 / 1000.0,
        peak_memory_mb: stats.peak_memory_usage_mb,
        cache_hit: false,
        cold_start: false,
        error_occurred: stats.error_rate > 0.0,
        timestamp: std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs(),
    };

    Json(serde_json::json!({
        "execution_metrics": {
            "example_record": example_metrics,
            "validate_storage_path_demo": {
                "description": "Validates storage paths for security",
                "test_path_valid": validate_storage_path("test.txt", "/tmp").is_ok(),
                "test_path_for_write": validate_storage_path_for_write("output.txt", "/tmp").is_ok()
            },
            "parse_requirement_name_demo": {
                "description": "Parse Python package requirement names",
                "requests_parsed": parse_requirement_name("requests>=2.0.0"),
                "flask_parsed": parse_requirement_name("flask")
            },
            "package_cache_stats_demo": {
                "description": "Package cache statistics structure",
                "example_stats": PackageCacheStats {
                    result_cache_entries: stats.total_executions as usize,
                    filesystem_cache_size_mb: 64,
                    max_cache_size_mb: 256,
                    cache_hit_ratio: stats.cache_hit_rate
                }
            },
            "aggregate_stats": {
                "total_executions": stats.total_executions,
                "total_execution_time_ms": stats.total_execution_time_ms,
                "average_execution_time_ms": stats.average_execution_time_ms,
                "peak_memory_usage_mb": stats.peak_memory_usage_mb,
                "total_memory_used_mb": stats.total_memory_used_mb,
                "cache_hit_rate": stats.cache_hit_rate,
                "error_rate": stats.error_rate,
                "functions_served": stats.functions_served,
                "uptime_seconds": stats.uptime_seconds
            }
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Resource usage / status
// ---------------------------------------------------------------------------

/// Resource usage handler (for enterprise resource enforcement).
pub async fn resource_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let resource_info = if let Some(ref enforcer) = state.resource_enforcer {
        let report = enforcer.get_resource_report().await;
        let policies = enforcer.policies().await;
        let default_policy = enforcer.get_policy("default").await;
        let enforcer_config = enforcer.config();

        let limiter_info = serde_json::json!({
            "wasm_resource_limiter": {
                "memory_limit_bytes": enforcer_config.memory_mb as u64 * 1024 * 1024,
                "max_memory_mb": enforcer_config.memory_mb
            }
        });

        Some(serde_json::json!({
            "resource_report": {
                "global_limits": report.global_limits,
                "function_quotas": report.function_quotas,
                "current_stats": report.current_stats,
                "timestamp": report.timestamp
            },
            "enforcement_policies": policies,
            "default_policy": default_policy,
            "config": {
                "enterprise_enabled": enforcer_config.enterprise_enabled,
                "memory_mb": enforcer_config.memory_mb,
                "timeout_ms": enforcer_config.timeout_ms,
            },
            "wasm_limiter": limiter_info
        }))
    } else {
        None
    };

    Json(serde_json::json!({
        "resource_enforcement": resource_info,
        "enterprise_enabled": state.config.enterprise_enabled,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Function metadata
// ---------------------------------------------------------------------------

/// Function metadata handler — returns standardized function metadata.
pub async fn function_metadata(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let capabilities = crate::capability::Capabilities::from_string(&state.config.capabilities);

    let metadata = FunctionMetadata {
        name: state.config.function.clone(),
        runtime: state.config.runtime.clone(),
        runtime_version: "1.0.0".to_string(),
        version: state.config.version.clone(),
        entry_point: "handler".to_string(),
        dependencies: vec![],
        memory_mb: state.config.memory_mb,
        timeout_ms: state.config.timeout_ms,
        uses_network: capabilities.can_fetch(),
        uses_filesystem: capabilities.can_storage(),
        runtime_metadata: serde_json::json!({
            "deterministic": state.config.deterministic,
            "cpu_fuel_limit": state.config.cpu_fuel_limit,
            "max_concurrent": state.config.max_concurrent_per_function
        }),
    };

    Json(serde_json::json!({
        "function_metadata": metadata,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Execution result info
// ---------------------------------------------------------------------------

/// Execution result format info handler — shows what ExecutionResult contains.
pub async fn execution_result_info(State(_state): State<Arc<AppState>>) -> axum::response::Response {
    let example_result = ExecutionResult::success(
        "example output".to_string(),
        100,
        1024 * 1024,
        5000,
    );

    // Demonstrate track_memory_usage with a real WASM module
    let engine = wasmtime::Engine::default();
    let mut store = wasmtime::Store::new(&engine, ());

    // Create a module with 2 pages of memory (128KB)
    let module = match wasmtime::Module::new(&engine, r#"
        (module
            (memory (export "memory") 2)
            (func (export "test"))
        )
    "#) {
        Ok(m) => m,
        Err(_) => {
            return Json(serde_json::json!({
                "execution_result_format": {
                    "description": "Standardized execution result returned by WASM functions",
                    "fields": {
                        "output": "Function execution output as string",
                        "success": "Whether execution completed successfully",
                        "error": "Error message if execution failed",
                        "exec_time_ms": "Execution time in milliseconds",
                        "memory_used": "Memory used during execution in bytes",
                        "fuel_used": "Fuel consumed during execution"
                    },
                    "example": example_result
                },
                "available_methods": {
                    "success": "Creates a successful ExecutionResult",
                    "failure": "Creates a failed ExecutionResult",
                    "track_memory_usage": "Tracks memory usage from WASM instance"
                },
                "error": "Failed to create test WASM module",
                "timestamp": std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs()
            }))
            .into_response();
        }
    };

    let instance = match wasmtime::Instance::new(&mut store, &module, &[]) {
        Ok(i) => i,
        Err(_) => {
            return Json(serde_json::json!({
                "execution_result_format": {
                    "description": "Standardized execution result returned by WASM functions",
                    "example": example_result
                },
                "available_methods": {
                    "success": "Creates a successful ExecutionResult",
                    "failure": "Creates a failed ExecutionResult",
                    "track_memory_usage": "Tracks memory usage from WASM instance"
                },
                "error": "Failed to instantiate test WASM module",
                "timestamp": std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs()
            }))
            .into_response();
        }
    };

    // Actually use track_memory_usage
    let memory_used = ExecutionResult::track_memory_usage(&mut store, &instance);
    let memory_pages = memory_used / (64 * 1024);
    let memory_mb = memory_used / (1024 * 1024);

    Json(serde_json::json!({
        "execution_result_format": {
            "description": "Standardized execution result returned by WASM functions",
            "fields": {
                "output": "Function execution output as string",
                "success": "Whether execution completed successfully",
                "error": "Error message if execution failed",
                "exec_time_ms": "Execution time in milliseconds",
                "memory_used": "Memory used during execution in bytes",
                "fuel_used": "Fuel consumed during execution"
            },
            "example": example_result
        },
        "available_methods": {
            "success": "Creates a successful ExecutionResult",
            "failure": "Creates a failed ExecutionResult",
            "track_memory_usage": "Tracks memory usage from WASM instance - demonstrated below"
        },
        "track_memory_usage_demo": {
            "memory_used_bytes": memory_used,
            "memory_used_pages": memory_pages,
            "memory_used_mb": memory_mb,
            "description": "Memory tracking from WASM instance with 2 pages (128KB)"
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Capability introspection
// ---------------------------------------------------------------------------

/// Full capability introspection for production runtime.
/// Returns comprehensive information about all capabilities, their status,
/// and production-ready configuration.
pub async fn capability_introspection(State(state): State<Arc<AppState>>) -> axum::response::Response {
    use crate::capability::{
        ALLOWED_CAPABILITIES, describe_capability, validate_capabilities,
        Capabilities
    };

    let capabilities = Capabilities::from_string(&state.config.capabilities);
    let declared_caps: Vec<String> = capabilities.all().iter().cloned().collect();
    let validation_result = validate_capabilities(&capabilities);

    // Categorize capabilities by risk level
    let (low_risk, medium_risk, high_risk) = {
        let mut low = Vec::new();
        let mut medium = Vec::new();
        let mut high = Vec::new();

        for cap in ALLOWED_CAPABILITIES {
            let risk = match *cap {
                // Low risk - read-only, local
                "cache:read" | "kv" | "metric" => "low",
                // Medium risk - read-write but scoped
                "fetch:read" | "cache:write" | "queue" => "medium",
                // High risk - write access, external communication
                "fetch:write" | "crypto" | "webhook" | "email" | "storage"
                | "ai" | "external_api" | "secret" => "high",
                _ => "unknown",
            };

            let info = serde_json::json!({
                "name": cap,
                "description": describe_capability(cap),
                "declared": capabilities.has(cap),
            });

            match risk {
                "low" => low.push(info),
                "medium" => medium.push(info),
                "high" => high.push(info),
                _ => {}
            }
        }
        (low, medium, high)
    };

    // Production recommendations
    let recommendations = {
        let mut recs = Vec::new();

        if capabilities.is_empty() {
            recs.push("No capabilities declared - function runs in fully sandboxed mode".to_string());
        }

        if capabilities.can_fetch_write() && !capabilities.can_fetch_read() {
            recs.push("Consider adding fetch:read for GET requests if needed".to_string());
        }

        if capabilities.can_storage() && !capabilities.can_kv() {
            recs.push("Storage access granted - consider using KV for structured data".to_string());
        }

        if capabilities.can_email() {
            recs.push("Email capability enabled - ensure SMTP is properly configured".to_string());
        }

        if capabilities.can_secret() {
            recs.push("Secret access enabled - ensure secrets are properly scoped".to_string());
        }

        if capabilities.can_external_api() {
            recs.push("External API access enabled - consider rate limiting".to_string());
        }

        recs
    };

    Json(serde_json::json!({
        "capability_introspection": {
            "description": "Production-ready capability introspection for FunctionFly runtime",
            "declared_capabilities": declared_caps,
            "validation": {
                "valid": validation_result.is_ok(),
                "error": validation_result.err(),
            },
            "capability_matrix": {
                "low_risk": low_risk,
                "medium_risk": medium_risk,
                "high_risk": high_risk,
            },
            "capability_checks": {
                "can_network": capabilities.can_fetch(),
                "can_network_read": capabilities.can_fetch_read(),
                "can_network_write": capabilities.can_fetch_write(),
                "can_kv": capabilities.can_kv(),
                "can_cache": capabilities.can_cache(),
                "can_cache_read": capabilities.can_cache_read(),
                "can_cache_write": capabilities.can_cache_write(),
                "can_storage": capabilities.can_storage(),
                "can_email": capabilities.can_email(),
                "can_webhook": capabilities.can_webhook(),
                "can_ai": capabilities.can_ai(),
                "can_crypto": capabilities.can_crypto(),
                "can_queue": capabilities.can_queue(),
                "can_metric": capabilities.can_metric(),
                "can_secret": capabilities.can_secret(),
                "can_external_api": capabilities.can_external_api(),
            },
            "production_recommendations": recommendations,
            "allowed_capabilities": ALLOWED_CAPABILITIES,
        },
        "runtime_config": {
            "function": state.config.function,
            "version": state.config.version,
            "runtime": state.config.runtime,
            "memory_mb": state.config.memory_mb,
            "timeout_ms": state.config.timeout_ms,
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

/// Execution result examples handler — demonstrates ExecutionResult usage.
pub async fn execution_result_examples(State(_state): State<Arc<AppState>>) -> axum::response::Response {
    use crate::python::engine::{PythonExecutionResult, PythonFunctionMetadata, PythonEngine};

    let success_result = ExecutionResult::success(
        "Function executed successfully".to_string(),
        150,
        2097152,
        5000,
    );

    let failure_result = ExecutionResult::failure(
        "Execution failed: out of memory".to_string(),
        50,
        1024,
        1000,
    );

    // Demonstrate PythonEngine::is_python_code
    let test_code_samples = [
        ("def hello(): return 'world'", true),
        ("import numpy as np\nnp.array([1,2,3])", true),
        ("console.log('hello')", false),
        ("function test() { return 1; }", false),
    ];
    let python_detection_results: Vec<serde_json::Value> = test_code_samples
        .iter()
        .map(|(code, expected)| serde_json::json!({
            "code": code.chars().take(30).collect::<String>(),
            "is_python": PythonEngine::is_python_code(code),
            "expected": expected
        }))
        .collect();

    // Demonstrate PythonExecutionResult
    let python_success = PythonExecutionResult::success("Python result: hello".to_string(), 75);
    let python_failure = PythonExecutionResult::failure("Python error: invalid syntax".to_string(), 10);

    // Demonstrate PythonFunctionMetadata
    let python_meta = PythonFunctionMetadata {
        name: "my_python_function".to_string(),
        python_version: "3.11".to_string(),
        runtime_version: "rustpython-0.4".to_string(),
        entry_point: "handler".to_string(),
        dependencies: vec!["numpy".to_string(), "requests".to_string()],
        memory_mb: 256,
        uses_network: true,
        uses_filesystem: true,
    };

    let example_pages: usize = 100;
    let example_memory_bytes = example_pages * 64 * 1024;

    Json(serde_json::json!({
        "execution_result_examples": {
            "wasm": {
                "success": success_result,
                "failure": failure_result
            },
            "python": {
                "success": python_success,
                "failure": python_failure,
                "metadata": python_meta,
                "is_python_code_detection": python_detection_results
            },
            "memory_calculation": {
                "example_pages": example_pages,
                "bytes_per_page": 64 * 1024,
                "total_bytes": example_memory_bytes
            },
            "methods": {
                "success": "Creates a successful ExecutionResult with output",
                "failure": "Creates a failed ExecutionResult with error message",
                "track_memory_usage": "Static method to track memory from WASM instance"
            }
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Resource quota management
// ---------------------------------------------------------------------------

/// Update global resource limits (admin endpoint for resource enforcement).
pub async fn update_global_limits(
    State(state): State<Arc<AppState>>,
    Json(limits): Json<GlobalResourceLimits>,
) -> axum::response::Response {
    if let Some(ref enforcer) = state.resource_enforcer {
        enforcer.update_global_limits(limits.clone()).await;
        Json(serde_json::json!({
            "success": true,
            "message": "Global resource limits updated",
            "limits": limits
        }))
        .into_response()
    } else {
        Json(serde_json::json!({
            "success": false,
            "error": "Resource enforcer not enabled (requires enterprise_enabled)"
        }))
        .into_response()
    }
}

/// Set resource quotas for a specific function.
pub async fn set_function_quotas(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<serde_json::Value>,
) -> axum::response::Response {
    if let Some(ref enforcer) = state.resource_enforcer {
        let function_key = match payload.get("function_key").and_then(|v| v.as_str()) {
            Some(key) => key.to_string(),
            None => {
                return Json(serde_json::json!({
                    "success": false,
                    "error": "function_key is required"
                }))
                .into_response()
            }
        };

        let quotas: Vec<ResourceQuota> = payload
            .get("quotas")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter().filter_map(|q| {
                    let quota_type = match q.get("quota_type")?.as_str()? {
                        "cpu_time_per_minute" => QuotaType::CpuTimePerMinute,
                        "memory_usage" => QuotaType::MemoryUsage,
                        "concurrent_executions" => QuotaType::ConcurrentExecutions,
                        "executions_per_hour" => QuotaType::ExecutionsPerHour,
                        "bandwidth_per_minute" => QuotaType::BandwidthPerMinute,
                        _ => return None,
                    };
                    Some(ResourceQuota {
                        quota_type,
                        limit: q.get("limit")?.as_f64().unwrap_or(0.0),
                        window_seconds: q.get("window_seconds")?.as_u64().unwrap_or(60),
                        current_usage: 0.0,
                        last_reset: std::time::Instant::now(),
                    })
                }).collect()
            })
            .unwrap_or_default();

        enforcer.set_function_quotas(function_key.clone(), quotas.clone()).await;

        Json(serde_json::json!({
            "success": true,
            "message": format!("Quotas updated for function {}", function_key),
            "function_key": function_key,
            "quotas": quotas
        }))
        .into_response()
    } else {
        Json(serde_json::json!({
            "success": false,
            "error": "Resource enforcer not enabled (requires enterprise_enabled)"
        }))
        .into_response()
    }
}
