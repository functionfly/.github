//! HTTP handlers for the local runtime, broken into focused modules.
//!
//! ## Module overview
//!
//! | Module | Responsibility |
//! |--------|----------------|
//! | [`types`] | Shared DTOs, `AppState`, and re-exports of all handler types |
//! | [`execution`] | Main function execution handler + `execute_with_error_handling` |
//! | [`health`] | `GET /health` and `GET /ready` handlers |
//! | [`monitoring`] | Stats, security, KV, webhook, orchestrator, and budget handlers |
//! | [`daemon`] | Daemon-mode execution handler + base64 decoder |
//! | [`scheduler_handlers`] | Scheduler status and scheduling simulation |
//! | [`metrics_handlers`] | Execution metrics, resource status, function metadata |
//! | [`runtime_handlers`] | WASI context, module linking, MicroPython, shutdown |
//! | [`prometheus`] | Prometheus metrics exporter |
//!
//! ## Re-exports
//!
//! All public items from sub-modules are re-exported at the `handlers` level so
//! callers can import everything from `crate::handlers` without knowing the
//! internal module layout.

pub mod billing_handlers;
pub mod daemon;
pub mod execution;
pub mod graph;
pub mod health;
pub mod metrics_handlers;
pub mod monitoring;
pub mod optimizer_handlers;
pub mod prometheus;
pub mod runtime_handlers;
pub mod scheduler_handlers;
pub mod types;

// ---------------------------------------------------------------------------
// Re-exports — public handlers and types
// ---------------------------------------------------------------------------

// Types & DTOs - re-exported for convenience
// Note: AppState is used by server.rs and other modules
// DaemonExecuteRequest, ErrorResponse, ExecuteRequest, ExecuteResponse, HealthResponse
// are used internally by handler modules via super::types
pub use types::{AppState};

// Execution
pub use execution::{execute_function, execute_chunked};
pub use graph::execute_graph;

// Health
pub use health::{health_check, ready_check};

// Monitoring
pub use monitoring::{
    budget_analysis, error_status, isolation_utils, kv_status, monitoring_stats,
    orchestrator_status, security_status, webhook_status,
};

// Daemon
pub use daemon::execute_function_daemon;

// Scheduler
pub use scheduler_handlers::{scheduler_mark_healthy, scheduler_mark_unhealthy, scheduler_remove_node, scheduler_status, scheduling_simulate};

// Optimizer (Phase 7)
pub use optimizer_handlers::{analyze_graph, get_suggestions, apply_optimization, get_mutation_log, rollback_mutation};

// Billing (Phase 6)
pub use billing_handlers::{get_costs, flush_costs, estimate_cost};

// Metrics
pub use metrics_handlers::{
    execution_metrics, execution_result_examples, execution_result_info,
    function_metadata, resource_status, update_global_limits, set_function_quotas,
    capability_introspection,
};

// Runtime / misc
pub use runtime_handlers::{
    micropython_status, runtime_config, runtime_control, runtime_metrics, runtime_status, shutdown_status,
    python_cache_status, python_cache_control, create_wasi_context, engine_status,
};

// Prometheus
pub use prometheus::prometheus_metrics;
