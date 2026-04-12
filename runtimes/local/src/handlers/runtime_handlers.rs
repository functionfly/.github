//! Runtime management handlers: configuration, status, and runtime control.

use axum::{extract::State, response::IntoResponse, Json};
use std::sync::Arc;
use serde::Deserialize;

use super::types::AppState;
use crate::micropython;

// ---------------------------------------------------------------------------
// Runtime Configuration
// ---------------------------------------------------------------------------

/// Runtime configuration status handler.
pub async fn runtime_config(State(state): State<Arc<AppState>>) -> axum::response::Response {
    use crate::package::{MAX_PACKAGE_DOWNLOAD_BYTES, PACKAGE_DOWNLOAD_TIMEOUT};

    let package_manager_info = if let Some(ref pm) = state.package_manager {
        Some(serde_json::json!({
            "available": true,
            "cache_stats": pm.get_cache_stats().await,
            "download_limits": {
                "max_package_download_bytes": MAX_PACKAGE_DOWNLOAD_BYTES,
                "max_package_download_mb": MAX_PACKAGE_DOWNLOAD_BYTES / (1024 * 1024),
                "download_timeout_secs": PACKAGE_DOWNLOAD_TIMEOUT.as_secs(),
            }
        }))
    } else {
        Some(serde_json::json!({
            "available": false,
            "download_limits": {
                "max_package_download_bytes": MAX_PACKAGE_DOWNLOAD_BYTES,
                "max_package_download_mb": MAX_PACKAGE_DOWNLOAD_BYTES / (1024 * 1024),
                "download_timeout_secs": PACKAGE_DOWNLOAD_TIMEOUT.as_secs(),
            }
        }))
    };

    Json(serde_json::json!({
        "runtime": {
            "function": state.config.function,
            "version": state.config.version,
            "memory_mb": state.config.memory_mb,
            "timeout_ms": state.config.timeout_ms,
        },
        "features": {
            "micropython_enabled": state.micropython_executor.is_some(),
            "package_manager": package_manager_info
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Runtime Status
// ---------------------------------------------------------------------------

/// Runtime status handler - provides operational status without sensitive details.
pub async fn runtime_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let pool_stats = state.pool.read().await.stats();
    let cache_stats = state.cache.read().await.stats();

    let wasm_pool_stats = if let Some(ref wasm_pool) = state.wasm_pool {
        let stats = wasm_pool.stats().await;
        Some(serde_json::json!({
            "enabled": true,
            "pools": stats.iter().map(|s| {
                serde_json::json!({
                    "function_key": s.function_key,
                    "idle_count": s.idle_count,
                    "max_idle": s.max_idle,
                    "available_permits": s.available_permits,
                    "max_concurrent": s.max_concurrent,
                })
            }).collect::<Vec<_>>(),
            "warmed_functions": stats.len(),
        }))
    } else {
        Some(serde_json::json!({
            "enabled": false
        }))
    };

    let python_pool_stats = if let Some(ref python_pool) = state.python_pool {
        let stats = python_pool.stats().await;
        Some(serde_json::json!({
            "enabled": true,
            "idle_count": stats.idle_count,
            "max_idle": stats.max_idle,
            "active_count": stats.active_count,
            "max_concurrent": stats.max_concurrent,
        }))
    } else {
        Some(serde_json::json!({
            "enabled": false
        }))
    };

    Json(serde_json::json!({
        "status": "operational",
        "uptime_seconds": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs(),
        "pool": {
            "total_instances": pool_stats.total_instances,
            "functions_in_pool": pool_stats.functions_in_pool,
            "current_memory_usage_mb": pool_stats.current_memory_usage_mb,
            "max_memory_usage_mb": pool_stats.max_memory_usage_mb,
        },
        "wasm_pool": wasm_pool_stats,
        "python_pool": python_pool_stats,
        "cache": {
            "enabled": state.cache.read().await.is_enabled(),
            "entries": cache_stats.entries,
            "binary_entries": cache_stats.binary_entries,
            "ttl_secs": cache_stats.ttl_secs,
        },
        "health": {
            "pool_healthy": pool_stats.current_memory_usage_mb <= pool_stats.max_memory_usage_mb,
            "cache_healthy": state.cache.read().await.is_enabled(),
        }
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Engine Status (WasmEngine info)
// ---------------------------------------------------------------------------

/// Engine status handler - provides WasmEngine operational information.
pub async fn engine_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let engine_ptr = format!("{:p}", state.engine.engine() as *const _);

    let wasi_linker_info = if let Some(_linker) = state.engine.wasi_linker() {
        serde_json::json!({
            "available": true,
            "function_key": format!("{}@{}", state.config.function, state.config.version),
        })
    } else {
        serde_json::json!({
            "available": false
        })
    };

    Json(serde_json::json!({
        "engine": {
            "engine_ptr": engine_ptr,
            "wasi_enabled": state.engine.wasi_linker().is_some(),
            "wasi_linker": wasi_linker_info,
        },
        "constructors": {
            "new": "WasmEngine::new(logger, security_monitor) -> anyhow::Result<Self>",
            "with_config": "WasmEngine::with_config(config, kv_store, logger, orchestrator_client, security_monitor) -> anyhow::Result<Self>",
        },
        "config": {
            "function": state.config.function,
            "version": state.config.version,
            "runtime": state.config.runtime,
            "aot_cache_enabled": state.config.aot_cache_enabled,
            "wasm_pool_enabled": state.config.wasm_pool_enabled,
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Runtime Control (Admin only)
// ---------------------------------------------------------------------------

#[derive(Deserialize)]
pub struct RuntimeControlRequest {
    pub action: RuntimeControlAction,
}

#[derive(Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RuntimeControlAction {
    ClearModuleCache,
    ResetPool,
    ClearKVStore,
    SetDebugMode { enabled: bool },
}

/// Runtime control handler for administrative operations.
pub async fn runtime_control(
    State(state): State<Arc<AppState>>,
    Json(request): Json<RuntimeControlRequest>,
) -> axum::response::Response {
    match request.action {
        RuntimeControlAction::ClearModuleCache => {
            let mut result_cache = state.cache.write().await;
            result_cache.clear();

            Json(serde_json::json!({
                "action": "clear_module_cache",
                "status": "completed",
                "result_cache_cleared": true,
            })).into_response()
        }

        RuntimeControlAction::ResetPool => {
            let mut pool = state.pool.write().await;
            let stats_before = pool.stats();
            pool.clear();
            let stats_after = pool.stats();

            Json(serde_json::json!({
                "action": "reset_pool",
                "status": "completed",
                "instances_cleared": stats_before.total_instances,
                "pool_reset": true,
                "stats_after": {
                    "total_instances": stats_after.total_instances,
                    "functions_in_pool": stats_after.functions_in_pool
                }
            })).into_response()
        }

        RuntimeControlAction::ClearKVStore => {
            let mut kv_store = state.kv.write().await;
            kv_store.clear();

            Json(serde_json::json!({
                "action": "clear_kv_store",
                "status": "completed",
                "kv_store_cleared": true
            })).into_response()
        }

        RuntimeControlAction::SetDebugMode { enabled } => {
            if let Some(ref pm) = state.package_manager {
                if let Err(e) = pm.cleanup_cache().await {
                    return Json(serde_json::json!({
                        "error": format!("Failed to cleanup package cache: {}", e)
                    })).into_response();
                }
            }

            Json(serde_json::json!({
                "action": "set_debug_mode",
                "enabled": enabled,
                "status": "completed",
                "debug_enabled": enabled
            })).into_response()
        }
    }
}

// ---------------------------------------------------------------------------
// Runtime Metrics (Read-only)
// ---------------------------------------------------------------------------

/// Runtime metrics handler - provides operational metrics.
pub async fn runtime_metrics(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let memory_stats = state.monitor.get_memory_stats();
    let pool = state.pool.read().await;
    let has_capacity = pool.has_capacity(64);
    let has_capacity_legacy = pool.has_capacity_legacy();

    Json(serde_json::json!({
        "metrics": {
            "used_memory_mb": memory_stats.used_mb,
            "limit_memory_mb": memory_stats.limit_mb,
        },
        "pool_capacity": {
            "has_capacity_for_64mb": has_capacity,
            "has_legacy_capacity": has_capacity_legacy,
        },
        "thresholds": {
            "max_memory_mb": state.config.memory_mb,
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Shutdown Status
// ---------------------------------------------------------------------------

/// Shutdown status handler - provides graceful shutdown coordination status.
pub async fn shutdown_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let shutdown_info = if let Some(ref sc) = state.shutdown_coordinator {
        let coordinator = sc.read().await;
        Some(serde_json::json!({
            "is_shutting_down": coordinator.is_shutting_down(),
            "timeout_seconds": coordinator.timeout().as_secs(),
            "shutdown_signaled": coordinator.is_shutdown_signaled(),
        }))
    } else {
        None
    };

    let resource_info = if let Some(ref rm) = state.resource_manager {
        let count = rm.resource_count().await;
        Some(serde_json::json!({
            "managed_resources_count": count,
            "resource_manager_active": true
        }))
    } else {
        Some(serde_json::json!({
            "managed_resources_count": 0,
            "resource_manager_active": false
        }))
    };

    Json(serde_json::json!({
        "shutdown_coordinator": shutdown_info,
        "resource_management": resource_info,
        "status": if shutdown_info.as_ref().map(|s| s["is_shutting_down"].as_bool().unwrap_or(false)).unwrap_or(false) {
            "shutting_down"
        } else {
            "operational"
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Python Cache Status
// ---------------------------------------------------------------------------

/// Python cache statistics endpoint.
pub async fn python_cache_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let cache = state.cache.read().await;
    let cache_stats = cache.stats();

    Json(serde_json::json!({
        "python_cache": {
            "enabled": cache.is_enabled(),
            "ttl_seconds": cache_stats.ttl_secs,
            "entries": cache_stats.entries,
            "binary_entries": cache_stats.binary_entries,
        },
        "enterprise_enabled": state.config.enterprise_enabled,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Python Cache Operations (Admin only - write operations)
// ---------------------------------------------------------------------------

#[derive(Deserialize)]
pub struct PythonCacheRequest {
    pub action: PythonCacheAction,
}

#[derive(Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PythonCacheAction {
    GetPythonWasm { source_hash: String },
    SetPythonWasm { source_hash: String, wasm_bytes: String },
    GetPackage { package_name: String, version: String },
    SetPackage { package_name: String, version: String, package_data: String },
    GetDependencyResolution { requirements_hash: String },
    SetDependencyResolution { requirements_hash: String, resolution: String },
    GetRustPythonRuntime,
    SetRustPythonRuntime { runtime_bytes: String },
    HashPythonSource { source_code: String },
    HashRequirements { requirements: Vec<String> },
}

const MAX_CACHE_ENTRY_SIZE: usize = 50 * 1024 * 1024; // 50MB max for any cache entry

/// Python cache management endpoint with input validation.
pub async fn python_cache_control(
    State(state): State<Arc<AppState>>,
    Json(request): Json<PythonCacheRequest>,
) -> axum::response::Response {
    let mut cache = state.cache.write().await;

    match request.action {
        PythonCacheAction::GetPythonWasm { source_hash } => {
            if source_hash.len() > 128 {
                return Json(serde_json::json!({
                    "action": "get_python_wasm",
                    "error": "Invalid hash format"
                })).into_response();
            }
            match cache.get_python_wasm(&source_hash) {
                Some(wasm_bytes) => Json(serde_json::json!({
                    "action": "get_python_wasm",
                    "found": true,
                    "wasm_bytes_size": wasm_bytes.len(),
                })).into_response(),
                None => Json(serde_json::json!({
                    "action": "get_python_wasm",
                    "found": false
                })).into_response(),
            }
        }
        PythonCacheAction::SetPythonWasm { source_hash, wasm_bytes } => {
            if source_hash.len() > 128 {
                return Json(serde_json::json!({
                    "action": "set_python_wasm",
                    "error": "Invalid hash format"
                })).into_response();
            }
            use base64::Engine;
            let wasm_data = match base64::engine::general_purpose::STANDARD.decode(&wasm_bytes) {
                Ok(d) => d,
                Err(e) => return Json(serde_json::json!({
                    "action": "set_python_wasm",
                    "error": format!("Invalid base64: {}", e)
                })).into_response(),
            };
            if wasm_data.len() > MAX_CACHE_ENTRY_SIZE {
                return Json(serde_json::json!({
                    "action": "set_python_wasm",
                    "error": "Data exceeds maximum size limit"
                })).into_response();
            }
            cache.set_python_wasm(&source_hash, &wasm_data);
            Json(serde_json::json!({
                "action": "set_python_wasm",
                "success": true,
                "source_hash": source_hash,
                "bytes_stored": wasm_data.len()
            })).into_response()
        }
        PythonCacheAction::GetPackage { package_name, version } => {
            if package_name.len() > 256 || version.len() > 64 {
                return Json(serde_json::json!({
                    "action": "get_package",
                    "error": "Invalid package name or version format"
                })).into_response();
            }
            match cache.get_package(&package_name, &version) {
                Some(data) => Json(serde_json::json!({
                    "action": "get_package",
                    "found": true,
                    "package_name": package_name,
                    "version": version,
                    "size": data.len()
                })).into_response(),
                None => Json(serde_json::json!({
                    "action": "get_package",
                    "found": false,
                    "package_name": package_name,
                    "version": version
                })).into_response(),
            }
        }
        PythonCacheAction::SetPackage { package_name, version, package_data } => {
            if package_name.len() > 256 || version.len() > 64 {
                return Json(serde_json::json!({
                    "action": "set_package",
                    "error": "Invalid package name or version format"
                })).into_response();
            }
            use base64::Engine;
            let pkg_data = match base64::engine::general_purpose::STANDARD.decode(&package_data) {
                Ok(d) => d,
                Err(e) => return Json(serde_json::json!({
                    "action": "set_package",
                    "error": format!("Invalid base64: {}", e)
                })).into_response(),
            };
            if pkg_data.len() > MAX_CACHE_ENTRY_SIZE {
                return Json(serde_json::json!({
                    "action": "set_package",
                    "error": "Data exceeds maximum size limit"
                })).into_response();
            }
            cache.set_package(&package_name, &version, &pkg_data);
            Json(serde_json::json!({
                "action": "set_package",
                "success": true,
                "package_name": package_name,
                "version": version,
                "bytes_stored": pkg_data.len()
            })).into_response()
        }
        PythonCacheAction::GetDependencyResolution { requirements_hash } => {
            if requirements_hash.len() > 128 {
                return Json(serde_json::json!({
                    "action": "get_dependency_resolution",
                    "error": "Invalid hash format"
                })).into_response();
            }
            match cache.get_dependency_resolution(&requirements_hash) {
                Some(resolution) => Json(serde_json::json!({
                    "action": "get_dependency_resolution",
                    "found": true,
                    "requirements_hash": requirements_hash,
                    "resolution": resolution
                })).into_response(),
                None => Json(serde_json::json!({
                    "action": "get_dependency_resolution",
                    "found": false
                })).into_response(),
            }
        }
        PythonCacheAction::SetDependencyResolution { requirements_hash, resolution } => {
            if requirements_hash.len() > 128 {
                return Json(serde_json::json!({
                    "action": "set_dependency_resolution",
                    "error": "Invalid hash format"
                })).into_response();
            }
            if resolution.len() > MAX_CACHE_ENTRY_SIZE {
                return Json(serde_json::json!({
                    "action": "set_dependency_resolution",
                    "error": "Resolution exceeds maximum size limit"
                })).into_response();
            }
            cache.set_dependency_resolution(&requirements_hash, resolution.clone());
            Json(serde_json::json!({
                "action": "set_dependency_resolution",
                "success": true,
                "requirements_hash": requirements_hash,
                "resolution_size": resolution.len()
            })).into_response()
        }
        PythonCacheAction::GetRustPythonRuntime => {
            match cache.get_rustpython_runtime() {
                Some(runtime) => Json(serde_json::json!({
                    "action": "get_rustpython_runtime",
                    "found": true,
                    "size": runtime.len()
                })).into_response(),
                None => Json(serde_json::json!({
                    "action": "get_rustpython_runtime",
                    "found": false
                })).into_response(),
            }
        }
        PythonCacheAction::SetRustPythonRuntime { runtime_bytes } => {
            use base64::Engine;
            let runtime_data = match base64::engine::general_purpose::STANDARD.decode(&runtime_bytes) {
                Ok(d) => d,
                Err(e) => return Json(serde_json::json!({
                    "action": "set_rustpython_runtime",
                    "error": format!("Invalid base64: {}", e)
                })).into_response(),
            };
            if runtime_data.len() > MAX_CACHE_ENTRY_SIZE {
                return Json(serde_json::json!({
                    "action": "set_rustpython_runtime",
                    "error": "Data exceeds maximum size limit"
                })).into_response();
            }
            cache.set_rustpython_runtime(&runtime_data);
            Json(serde_json::json!({
                "action": "set_rustpython_runtime",
                "success": true,
                "bytes_stored": runtime_data.len()
            })).into_response()
        }
        PythonCacheAction::HashPythonSource { source_code } => {
            if source_code.len() > 1_000_000 {
                return Json(serde_json::json!({
                    "action": "hash_python_source",
                    "error": "Source code exceeds maximum size"
                })).into_response();
            }
            let hash = crate::cache::ResultCache::hash_python_source(&source_code);
            Json(serde_json::json!({
                "action": "hash_python_source",
                "hash": hash
            })).into_response()
        }
        PythonCacheAction::HashRequirements { requirements } => {
            if requirements.len() > 1000 {
                return Json(serde_json::json!({
                    "action": "hash_requirements",
                    "error": "Too many requirements"
                })).into_response();
            }
            let hash = crate::cache::ResultCache::hash_requirements(&requirements);
            Json(serde_json::json!({
                "action": "hash_requirements",
                "hash": hash
            })).into_response()
        }
    }
}

// ---------------------------------------------------------------------------
// WASI Context Creation
// ---------------------------------------------------------------------------

#[derive(Deserialize)]
pub struct CreateWasiContextRequest {
    pub function_key: String,
    pub input: Option<String>,
}

/// Create a WASI context for pre-warming or testing.
pub async fn create_wasi_context(
    State(state): State<Arc<AppState>>,
    Json(request): Json<CreateWasiContextRequest>,
) -> axum::response::Response {
    if request.function_key.len() > 256 {
        return Json(serde_json::json!({
            "success": false,
            "error": "Invalid function key format"
        })).into_response();
    }

    let input = request.input.unwrap_or_default();

    match state.engine.create_wasi_context(&request.function_key, &input) {
        Ok(wasi_ctx) => {
            Json(serde_json::json!({
                "success": true,
                "function_key": wasi_ctx.function_key,
                "time_access_allowed": wasi_ctx.time_access_allowed,
                "wasi_ctx_type": "WasiContext",
                "timestamp": std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs()
            })).into_response()
        }
        Err(e) => {
            Json(serde_json::json!({
                "success": false,
                "error": e.to_string()
            })).into_response()
        }
    }
}

// ---------------------------------------------------------------------------
// MicroPython Execution Status
// ---------------------------------------------------------------------------

/// MicroPython execution status - shows availability without executing code.
pub async fn micropython_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let executor_available = state.micropython_executor.is_some();
    let micropython_available = micropython::is_micropython_available();
    let wasm_path = micropython::find_micropython_wasm();

    let executor_info = if let Some(ref executor) = state.micropython_executor {
        Some(serde_json::json!({
            "available": true,
            "ready": executor.is_ready(),
            "config": {
                "timeout_ms": executor.config().timeout_ms,
            }
        }))
    } else {
        None
    };

    Json(serde_json::json!({
        "micropython": {
            "executor_available": executor_available,
            "runtime_available": micropython_available,
            "wasm_path": wasm_path.as_ref(),
            "interface_version": micropython::MP_INTERFACE_VERSION,
        },
        "executor": executor_info,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}
