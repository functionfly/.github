//! Monitoring and status handlers.

use axum::{extract::State, response::IntoResponse, Json};
use std::sync::Arc;

use super::types::AppState;

// ---------------------------------------------------------------------------
// Monitoring stats
// ---------------------------------------------------------------------------

/// Monitoring statistics handler.
pub async fn monitoring_stats(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let report = state.monitor.generate_report().await;
    let correlation_id = state.logger.generate_correlation_id().await;
    let cache_stats = state.monitor.get_cache_stats().await;
    let memory_stats = state.monitor.get_memory_stats();

    state.logger.log_cache_stats(
        &correlation_id,
        cache_stats.entries,
        cache_stats.hits,
        cache_stats.misses,
    );

    state.logger.log_memory_usage(
        &correlation_id,
        memory_stats.used_mb,
        memory_stats.limit_mb,
    );

    state.logger.log(
        crate::logging::LogLevel::Info,
        format!("Monitoring stats requested: {} total executions", report.stats.total_executions),
    );

    state.logger.log_entry_with_fields(
        crate::logging::LogLevel::Info,
        "Monitoring stats requested",
        &correlation_id,
        Default::default(),
    );

    let orchestrator_stats = if let Some(client) = &state.orchestrator_client {
        match client.get_stats().await {
            Ok(stats) => Some(serde_json::json!({
                "active_vms": stats.active_vms,
                "warm_vms": stats.warm_vms,
                "max_vms": stats.max_vms
            })),
            Err(_) => None,
        }
    } else {
        None
    };

    let mut response = serde_json::json!({
        "stats": report.stats,
        "cache_stats": cache_stats,
        "memory_stats": memory_stats,
        "performance": {
            "p50_execution_time_ms": report.p50_execution_time_ms,
            "p95_execution_time_ms": report.p95_execution_time_ms,
            "p99_execution_time_ms": report.p99_execution_time_ms,
            "recent_executions": report.recent_executions
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    });

    if let Some(stats) = orchestrator_stats {
        response["orchestrator"] = stats;
    }

    Json(response).into_response()
}

// ---------------------------------------------------------------------------
// Security status
// ---------------------------------------------------------------------------

/// Security status handler.
pub async fn security_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let report = state.security_monitor.get_security_report().await;

    let enterprise_report = if let Some(ref enterprise) = state.enterprise_security {
        Some(enterprise.get_security_report().await)
    } else {
        None
    };

    let function_key = format!("{}@{}", state.config.function, state.config.version);

    let (syscall_allowed, network_allowed, filesystem_allowed, fd_limit, env_limit) = if state.config.enterprise_enabled {
        let syscall_check = state
            .security_monitor
            .is_syscall_allowed(&function_key, "fd_write")
            .await;
        let network_check = state
            .security_monitor
            .is_network_allowed(&function_key, "https://example.com")
            .await;
        let filesystem_check = state
            .security_monitor
            .is_filesystem_allowed(&function_key)
            .await;
        let fd_check = state
            .security_monitor
            .check_file_descriptor_limit(&function_key, 5)
            .await;
        let env_check = state
            .security_monitor
            .check_env_vars_limit(&function_key, 10)
            .await;
        (Some(syscall_check), Some(network_check), Some(filesystem_check), Some(fd_check), Some(env_check))
    } else {
        (None, None, None, None, None)
    };

    let attack_patterns = state
        .security_monitor
        .get_attack_patterns(&function_key)
        .await;

    let profile = state
        .security_monitor
        .get_profile(&function_key)
        .await;

    let attack_pattern_with_history = state
        .security_monitor
        .get_attack_pattern_with_history(&function_key, crate::security::AttackPatternType::SyscallFlood)
        .await;

    let attack_pattern_memory_history = state
        .security_monitor
        .get_attack_pattern_with_history(&function_key, crate::security::AttackPatternType::MemoryExhaustion)
        .await;

    let shutdown_info = if let Some(ref sc) = state.shutdown_coordinator {
        let coordinator = sc.read().await;
        Some(serde_json::json!({
            "is_shutting_down": coordinator.is_shutting_down(),
            "timeout_secs": coordinator.timeout().as_secs()
        }))
    } else {
        None
    };

    Json(serde_json::json!({
        "security_report": report,
        "enterprise_security_report": enterprise_report,
        "hardened_mode": state.config.hardened_security,
        "enterprise_enabled": state.config.enterprise_enabled,
        "isolation_level": if state.config.hardened_security {
            "maximum"
        } else if state.config.enterprise_enabled {
            "enterprise"
        } else {
            "standard"
        },
        "syscall_allowed": syscall_allowed,
        "network_allowed": network_allowed,
        "filesystem_allowed": filesystem_allowed,
        "file_descriptor_limit_ok": fd_limit,
        "env_vars_limit_ok": env_limit,
        "shutdown_coordinator": shutdown_info,
        "attack_patterns": attack_patterns.iter().map(|p| {
            serde_json::json!({
                "type": p.pattern_type.debug_name(),
                "occurrences": p.occurrences,
                "first_seen": p.first_seen,
                "last_seen": p.last_seen
            })
        }).collect::<Vec<_>>(),
        "attack_pattern_history": {
            "syscall_flood": attack_pattern_with_history.as_ref().map(|p| {
                serde_json::json!({
                    "type": p.pattern_type.debug_name(),
                    "occurrences": p.occurrences,
                    "first_seen": p.first_seen,
                    "last_seen": p.last_seen,
                    "age_seconds": std::time::SystemTime::now()
                        .duration_since(std::time::UNIX_EPOCH)
                        .unwrap_or_default()
                        .as_secs()
                        .saturating_sub(p.first_seen)
                })
            }),
            "memory_exhaustion": attack_pattern_memory_history.as_ref().map(|p| {
                serde_json::json!({
                    "type": p.pattern_type.debug_name(),
                    "occurrences": p.occurrences,
                    "first_seen": p.first_seen,
                    "last_seen": p.last_seen,
                    "age_seconds": std::time::SystemTime::now()
                        .duration_since(std::time::UNIX_EPOCH)
                        .unwrap_or_default()
                        .as_secs()
                        .saturating_sub(p.first_seen)
                })
            })
        },
        "security_profile": {
            "max_file_descriptors": profile.max_file_descriptors,
            "allow_network": profile.allow_network,
            "allow_filesystem": profile.allow_filesystem,
            "max_env_vars": profile.max_env_vars,
            "disable_dangerous_syscalls": profile.disable_dangerous_syscalls,
            "audit_syscalls": profile.audit_syscalls
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// KV store status
// ---------------------------------------------------------------------------

/// KV store status handler.
pub async fn kv_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let (stats, keys_sample) = {
        let store = state.kv.read().await;
        let all_keys = store.keys();
        let keys_sample: Vec<String> = all_keys.iter().take(10).cloned().collect();
        let store_stats = store.stats();
        (store_stats, keys_sample)
    };

    let kv_exists = {
        let has_content = stats.total_entries > 0;
        let capacity_percent = (stats.total_entries as f64 / stats.max_entries as f64) * 100.0;
        serde_json::json!({
            "has_content": has_content,
            "capacity_percent": capacity_percent,
            "sample_keys": keys_sample,
        })
    };

    Json(serde_json::json!({
        "kv_stats": {
            "total_entries": stats.total_entries,
            "expired_entries": stats.expired_entries,
            "max_entries": stats.max_entries,
            "has_kv_capability": crate::capability::Capabilities::from_string(&state.config.capabilities).can_kv()
        },
        "kv_exists": kv_exists,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Webhook status
// ---------------------------------------------------------------------------

/// Webhook capability status handler.
pub async fn webhook_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let capabilities = crate::capability::Capabilities::from_string(&state.config.capabilities);

    Json(serde_json::json!({
        "webhook_capability": {
            "enabled": capabilities.can_webhook(),
            "supported_methods": ["GET", "POST", "PUT", "PATCH", "DELETE"],
            "timeout_seconds": 30,
            "max_payload_size_kb": 1024
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Orchestrator status
// ---------------------------------------------------------------------------

/// Orchestrator status handler.
pub async fn orchestrator_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let orchestrator_info = if let Some(client) = &state.orchestrator_client {
        let health = client.ping().await;
        let stats = if health {
            match client.get_stats().await {
                Ok(stats) => Some(serde_json::json!({
                    "active_vms": stats.active_vms,
                    "warm_vms": stats.warm_vms,
                    "max_vms": stats.max_vms
                })),
                Err(_) => None,
            }
        } else {
            None
        };

        serde_json::json!({
            "available": health,
            "url": client.url(),
            "stats": stats,
            "enterprise_enabled": state.config.enterprise_enabled,
            "tier": state.config.tier,
            "supports_microvm": state.config.supports_microvm()
        })
    } else {
        serde_json::json!({
            "available": false,
            "enterprise_enabled": state.config.enterprise_enabled,
            "tier": state.config.tier,
            "supports_microvm": false
        })
    };

    Json(serde_json::json!({
        "orchestrator": orchestrator_info,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Budget analysis
// ---------------------------------------------------------------------------

use crate::budget::{BudgetOptimizer, BudgetTier, FunctionProfile};

/// Budget analysis handler.
pub async fn budget_analysis(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let stats = state.monitor.get_stats().await;
    let avg_execution_time = if stats.total_executions > 0 {
        stats.average_execution_time_ms as u64
    } else {
        100
    };

    let function_profile = FunctionProfile {
        runtime: state.config.runtime.clone(),
        avg_execution_time_ms: avg_execution_time,
        avg_memory_mb: state.config.memory_mb as f64,
        cold_start_time_ms: 50,
        requests_per_minute: 10.0,
        cache_hit_rate: if state.config.deterministic { 0.7 } else { 0.1 },
    };

    let optimizer = BudgetOptimizer::for_tier(&BudgetTier::UltraLow);
    let recommendations = optimizer.generate_recommendations(&[function_profile]);

    Json(serde_json::json!({
        "budget_analysis": recommendations,
        "current_config": {
            "memory_mb": state.config.memory_mb,
            "timeout_ms": state.config.timeout_ms,
            "deterministic": state.config.deterministic,
            "cpu_fuel_limit": state.config.cpu_fuel_limit,
            "max_concurrent_per_function": state.config.max_concurrent_per_function
        },
        "optimization_suggestions": recommendations.suggestions,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

// ---------------------------------------------------------------------------
// Isolation utilities status
// ---------------------------------------------------------------------------

/// Isolation utilities status - shows available security utilities.
pub async fn isolation_utils(State(state): State<Arc<AppState>>) -> axum::response::Response {
    use crate::security::IsolationUtils;

    let function_key = format!("{}@{}", state.config.function, state.config.version);

    let isolated_dir = IsolationUtils::create_isolated_temp_dir(&function_key);
    let dir_created = isolated_dir.is_ok();
    let dir_path = isolated_dir.as_ref().ok().map(|p| p.to_string_lossy().to_string());

    if let Ok(ref path) = isolated_dir {
        let _ = IsolationUtils::cleanup_isolated_dir(path);
    }

    let path_validation = IsolationUtils::is_safe_path(
        std::path::Path::new("/tmp"),
        "/tmp/test.txt"
    );

    let unsafe_path_validation = IsolationUtils::is_safe_path(
        std::path::Path::new("/tmp"),
        "/etc/passwd"
    );

    let env_vars = vec![
        ("SAFE_VAR".to_string(), "safe_value".to_string()),
        ("LD_LIBRARY_PATH".to_string(), "/dangerous/path".to_string()),
        ("PATH".to_string(), "/bin:/usr/bin".to_string()),
    ];
    let sanitized_vars = IsolationUtils::sanitize_env_vars(&env_vars);

    Json(serde_json::json!({
        "isolation_utils": {
            "available": true,
            "description": "Security utilities for sandbox isolation",
            "function_key": function_key,
            "methods": {
                "create_isolated_temp_dir": "Creates isolated temp directory for function execution",
                "cleanup_isolated_dir": "Cleans up isolated directory after execution",
                "is_safe_path": "Validates paths for path traversal prevention",
                "sanitize_env_vars": "Removes dangerous environment variables"
            }
        },
        "temp_directory": {
            "create_test": {
                "success": dir_created,
                "path": dir_path
            }
        },
        "path_validation": {
            "safe_path_test": {
                "path": "/tmp/test.txt",
                "base": "/tmp",
                "is_safe": path_validation
            },
            "unsafe_path_test": {
                "path": "/etc/passwd",
                "base": "/tmp",
                "is_safe": unsafe_path_validation
            }
        },
        "env_var_sanitization": {
            "original_count": env_vars.len(),
            "sanitized_count": sanitized_vars.len(),
            "removed_vars": env_vars.len() - sanitized_vars.len()
        },
        "security_profiles": {
            "hardened_mode": state.config.hardened_security,
            "enterprise_enabled": state.config.enterprise_enabled,
            "security_profile_fields": {
                "max_file_descriptors": 10,
                "allow_filesystem": !state.config.hardened_security,
                "max_env_vars": if state.config.hardened_security { 5 } else { 20 },
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
// Error status
// ---------------------------------------------------------------------------

use crate::errors::{RuntimeError, ErrorKind, ErrorRecovery};

/// Error status handler - shows available error types and recovery strategies.
pub async fn error_status(State(_state): State<Arc<AppState>>) -> axum::response::Response {
    let error_recovery = ErrorRecovery::new();
    
    let test_timeout_error = RuntimeError::new(ErrorKind::TimeoutExceeded, "Execution timeout");
    let test_memory_error = RuntimeError::new(ErrorKind::MemoryLimitExceeded, "Memory limit exceeded");
    
    let timeout_strategy = error_recovery.get_recovery_strategy(&test_timeout_error);
    let memory_strategy = error_recovery.get_recovery_strategy(&test_memory_error);

    Json(serde_json::json!({
        "error_kinds": {
            "execution_errors": ["WasmCompilation", "WasmInstantiation", "WasmExecution", "FunctionNotFound", "TimeoutExceeded"],
            "resource_errors": ["MemoryLimitExceeded", "FuelLimitExceeded", "InstancePoolExhausted", "RateLimitExceeded"],
            "python_errors": ["PythonEngineUnavailable", "PythonExecutionFailed", "PythonModuleNotFound"],
            "pool_errors": ["PoolPruningFailed", "PoolCapacityExceeded"],
            "cache_errors": ["CacheCorruption", "CacheSizeExceeded"],
            "security_errors": ["SecurityViolation", "ResourceLimitExceeded"]
        },
        "recovery_strategies": {
            "timeout_strategy": format!("{:?}", timeout_strategy),
            "memory_strategy": format!("{:?}", memory_strategy),
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}
