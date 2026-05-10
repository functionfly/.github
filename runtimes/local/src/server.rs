//! HTTP server for the local runtime.

use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::RwLock;

use anyhow::Result;
use axum::{
    routing::{delete, get, post},
    Router,
};
use tower_http::cors::{Any, CorsLayer};
use tower_http::trace::TraceLayer;

use crate::config::Config;
use crate::engine::SharedState;
use crate::engine::sar_executor::SarNodeExecutor;
use crate::handlers::{
    analyze_graph, apply_optimization, budget_analysis, capability_introspection,
    create_wasi_context, engine_status, error_status, estimate_cost, execute_chunked,
    execute_function, execute_function_daemon, execute_graph, flush_costs,
    function_metadata, get_costs, get_mutation_log, get_suggestions, health_check,
    isolation_utils, kv_status, micropython_status, monitoring_stats,
    orchestrator_status, prometheus_metrics, python_cache_control, python_cache_status,
    ready_check, resource_status, rollback_mutation, scheduler_mark_healthy,
    scheduler_mark_unhealthy, scheduler_remove_node, scheduler_status, scheduling_simulate,
    security_status, set_function_quotas, shutdown_status, update_global_limits,
    webhook_status, AppState, execution_metrics, execution_result_examples,
    execution_result_info, runtime_config, runtime_control, runtime_metrics, runtime_status,
};
use crate::memory::layer::MemoryLayer;
use crate::router::{FlyMindClient, FlyMindConfig};
use crate::kv::SharedKVStore;
use crate::logging::{CorrelationId, StructuredLogger};
use crate::micropython::{ExecutorConfig, MicroPythonExecutor};
use crate::enterprise_security::EnterpriseSecurityEnforcer;
use crate::package::PackageManager;
use crate::resource_enforcer::ResourceEnforcer;
use crate::scheduler::{BinPackingScheduler, NodeCapacity};
use crate::agent_scheduler::agent_scheduler::{AgentScheduler, RateLimitConfig};
use crate::agent_scheduler::worker::{JobTracker, spawn_scheduler_workers};
use crate::security::SecurityMonitor;
use crate::pool::InstancePool;
use crate::engine::wasm_cell::WasmCellExecutor;
use crate::monitoring::ResourceMonitor;
use crate::observability::cost::{CostAttributor, CostAttributorConfig};
use crate::optimizer::optimizer::GraphOptimizer;
use crate::optimizer::config::ThresholdConfig;
use crate::shutdown::{ComponentShutdown, ResourceManager, ShutdownCoordinator};
use crate::yara_scanner::YaraScanner;
use crate::errors::ErrorRecovery;

/// Run the HTTP server with graceful shutdown support
pub async fn run_server(
    port: u16,
    config: Config,
    logger: Arc<StructuredLogger>,
    startup_correlation_id: CorrelationId,
) -> Result<()> {
    // Build CORS layer from config.cors_allow_origin (e.g. --cors-allow-origin on the CLI).
    // Empty or "*" => allow all origins (default, for local dev). In production, set to a
    // specific origin (e.g. "https://app.example.com") so only that origin is allowed.
    let cors = if config.cors_allow_origin.is_empty() || config.cors_allow_origin == "*" {
        CorsLayer::new()
            .allow_origin(Any)
            .allow_methods(Any)
            .allow_headers(Any)
    } else {
        // Parse the configured origin; fall back to Any if invalid.
        let origin_header = config.cors_allow_origin.parse::<axum::http::HeaderValue>()
            .unwrap_or_else(|_| {
                tracing::warn!("Invalid CORS origin '{}', falling back to Any (*)", config.cors_allow_origin);
                "*".parse().unwrap()
            });
        CorsLayer::new()
            .allow_origin(origin_header)
            .allow_methods(Any)
            .allow_headers(Any)
    };

    // Create KV store and start background cleanup task.
    // The background task removes expired entries every 30 seconds, replacing
    // the previous O(n) per-operation cleanup.
    let kv_store_for_cleanup: SharedKVStore = Arc::new(RwLock::new(crate::kv::KVStore::new(10000))); // Max 10k entries
    let _kv_cleanup_handle = crate::kv::start_background_cleanup(kv_store_for_cleanup);

    // Create security monitor (single instance shared across all components)
    let security_monitor = Arc::new(SecurityMonitor::new());


    // Register security profiles before creating the engine
    if config.hardened_security {
        security_monitor.register_profile(
            format!("{}@{}", config.function, config.version),
            SecurityMonitor::create_hardened_profile(),
        ).await;
    } else if config.enterprise_enabled {
        // Use enterprise profile with network whitelist for enterprise tier
        security_monitor.register_profile(
            format!("{}@{}", config.function, config.version),
            SecurityMonitor::create_enterprise_profile(config.network_whitelist.clone()),
        ).await;
    } else {
        security_monitor.register_profile(
            format!("{}@{}", config.function, config.version),
            SecurityMonitor::create_standard_profile(),
        ).await;
    }

    // Create shutdown coordinator for graceful shutdown management
    // Use custom timeout from config for graceful shutdown
    let shutdown_timeout = std::time::Duration::from_secs(config.shutdown_timeout_secs);
    let shutdown_coordinator = Arc::new(RwLock::new(ShutdownCoordinator::with_timeout(logger.clone(), shutdown_timeout)));

    // Create resource manager for tracking and cleaning up resources during shutdown
    let resource_manager = Arc::new(ResourceManager::new(logger.clone()));

    // Create instance pool
    let pool = InstancePool::new(10, 60);

    // Create shared state (includes Python engine if configured)
    // Pass the same security_monitor to ensure consistent violation tracking.
    let shared_state = SharedState::new(pool, config.clone(), (*logger).clone(), security_monitor.clone());

    // Create WASI linker for host function injection (kv, ai, fetch, etc.)
    // This is used by WasmCellExecutor for graph node WASM isolation.
    let wasi_linker = match crate::wasi::WasiLinker::new(
        shared_state.engine.engine(),
        &shared_state.config,
        Some(shared_state.kv.clone()),
        (*logger).clone(),
        security_monitor.clone(),
    ) {
        Ok(l) => Some(Arc::new(l)),
        Err(e) => {
            tracing::warn!("Failed to create WASI linker: {}. Graph node WASM isolation will be unavailable.", e);
            None
        }
    };

    // Create WASM cell executor for graph node isolation (Phase 2).
    // This executor is used by the graph engine to run Tool and Memory nodes
    // as isolated WASM cells with host function injection.
    let wasm_cell_executor: Option<Arc<WasmCellExecutor>> = if wasi_linker.is_some() {
        Some(Arc::new(WasmCellExecutor::new(
            shared_state.config.wasm_pool_max_concurrent.max(4),
            shared_state.config.wasm_pool_max_idle.max(2),
        )))
    } else {
        None
    };

    // Use the KV store from SharedState for AppState
    let kv_store = shared_state.kv.clone();

    // Start background pool pruning on the *shared* pool reference so the
    // pruning task operates on the same pool used by the server (fix for the
    // detached-clone bug in the old start_background_pruning() method).
    let _pool_pruning_handle = InstancePool::start_background_pruning_shared(shared_state.pool.clone());

    // Start background metrics cleanup to remove old entries (older than 1 hour)
    let _monitor_cleanup_handle = ResourceMonitor::start_background_cleanup(shared_state.monitor.clone());

    // Create package manager for Enterprise tier
    let package_manager = if config.enterprise_enabled && config.package_caching_enabled {
        let pm = PackageManager::new(
            shared_state.cache.clone(),
            config.package_cache_dir.clone().into(),
            config.package_cache_size_mb,
            config.network_whitelist.clone(),
            config.strict_network_whitelist,
        ).map_err(|e| {
            tracing::error!("Failed to create package manager: {}", e);
            anyhow::anyhow!("Package manager creation failed: {}", e)
        })?;
        Some(Arc::new(pm))
    } else {
        None
    };

    // Create resource enforcer for Enterprise tier
    let resource_enforcer = if config.enterprise_enabled {
        Some(Arc::new(ResourceEnforcer::new(
            shared_state.monitor.clone(),
            config.clone(),
        )))
    } else {
        None
    };

    // Create enterprise security enforcer for Enterprise tier
    let enterprise_security = if config.enterprise_enabled {
        Some(Arc::new(EnterpriseSecurityEnforcer::new(logger.clone())))
    } else {
        None
    };

    // Start background cleanup task for enterprise security (clears stale rate limit data)
    if let Some(ref es) = enterprise_security {
        let es_clone = es.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(300)); // Every 5 minutes
            loop {
                interval.tick().await;
                es_clone.cleanup_old_data().await;
            }
        });
    }

    // Create YARA scanner for WASM artifact validation (Phase 2)
    let yara_scanner = if config.enterprise_enabled {
        let scanner_config = crate::yara_scanner::YaraScannerConfig::default();
        match YaraScanner::new(scanner_config) {
            Ok(scanner) => Some(Arc::new(scanner)),
            Err(e) => {
                tracing::warn!("Failed to create YARA scanner: {}. YARA scanning will be unavailable.", e);
                None
            }
        }
    } else {
        None
    };

    // Create bin-packing scheduler for multi-node orchestration (Phase 4)
    let scheduler = if config.enterprise_enabled {
        Some(Arc::new(BinPackingScheduler::new(60))) // 60 second stale threshold
    } else {
        None
    };

    // Clone scheduler for app_state before we might move it
    let scheduler_for_state = scheduler.clone();

    // Register this node with the scheduler if enabled
    if let Some(ref sched) = scheduler {
        let node_capacity = NodeCapacity::new(
            "local-node",  // node_id
            4000,          // total_cpu_millicores (4 vCPUs)
            8192,          // total_memory_mb (8 GB)
            20,            // max_executions
        );
        sched.upsert_node(node_capacity).await;
    }

    // Create MicroPython executor for Python WASM execution (used by python-wasm runtime)
    let micropython_executor = if config.enterprise_enabled {
        let executor_config = ExecutorConfig::default();
        match MicroPythonExecutor::new(executor_config) {
            Ok(executor) => Some(Arc::new(executor)),
            Err(e) => {
                tracing::warn!("Failed to create MicroPython executor: {}", e);
                None
            }
        }
    } else {
        None
    };

    // Create error recovery manager for execution resilience
    let error_recovery = if config.enterprise_enabled {
        Some(Arc::new(ErrorRecovery::new()))
    } else {
        None
    };

    // Create Python runtime pool for interpreter reuse (Phase 3)
    let python_pool = if config.enterprise_enabled {
        let python_config = crate::python::runtime::PythonConfig::from(config.clone());
        let pool = crate::python_pool::PythonRuntimePool::new(
            config.python_pool_max_concurrent,
            config.python_pool_max_idle,
            config.python_pool_max_reuse,
            python_config,
        );
        tracing::info!("Python runtime pool created (max_concurrent={}, max_idle={})",
            config.python_pool_max_concurrent, config.python_pool_max_idle);
        Some(Arc::new(pool))
    } else {
        None
    };

    // Create multi-tier memory layer (Phase 3) — connects to Redis + PostgreSQL if configured
    let memory_layer = Arc::new(MemoryLayer::with_tiers(
        std::env::var("REDIS_URL").ok().as_deref(),
        std::env::var("DATABASE_URL").ok().as_deref(),
        10_000, // hot tier capacity per tenant
    ));
    // Connect to Redis and PostgreSQL (non-blocking — falls back to in-process if unavailable)
    let memory_layer_for_connect = memory_layer.clone();
    tokio::spawn(async move {
        memory_layer_for_connect.connect_tiers().await;
    });

    // Create FlyMind model router client (Phase 4) — routes LLM nodes to Python ai-service
    let flymind_config = FlyMindConfig::default();
    let flymind_client = Arc::new(FlyMindClient::new(flymind_config));

    // Create self-optimization engine (Phase 7) — analyzes execution history, emits suggestions
    let optimizer = Arc::new(GraphOptimizer::new(
        memory_layer.state().clone(),
        ThresholdConfig::default(),
    ));
    tracing::info!("Graph optimizer initialized — pattern detection active");

    // Create cost attributor (Phase 6) — tracks per-agent, per-execution-path cost
    let cost_attributor = Arc::new(CostAttributor::new(CostAttributorConfig::default()));
    tracing::info!("Cost attributor initialized — billing records will be emitted to Go backend");

    // Create the integrated SAR node executor (wires Memory + FlyMind + WASM cells)
    let sar_executor = match SarNodeExecutor::from_app_state(
        Some(memory_layer.clone()),
        Some(flymind_client.clone()),
        wasm_cell_executor.clone(),
    ) {
        Ok(executor) => {
            tracing::info!("SAR Node Executor initialized with real service implementations");
            Some(Arc::new(executor))
        }
        Err(e) => {
            tracing::warn!("Failed to create SAR Node Executor (falling back to stubs): {}", e);
            None
        }
    };

    // Create agent scheduler (Phase 5) — priority queues with rate limiting and backpressure
    let agent_scheduler = Arc::new(tokio::sync::RwLock::new(AgentScheduler::new(
        RateLimitConfig::default(),
    )));
    tracing::info!("Agent scheduler initialized with 4 priority queues");

    // Create job tracker for async execution status (retains last 10k jobs)
    let job_tracker = Arc::new(JobTracker::new(10_000));

    // Start real scheduler worker tasks to dequeue and execute graphs
    // Pass cost_attributor for Phase 6 observability in async executions
    spawn_scheduler_workers(
        agent_scheduler.clone(),
        sar_executor.clone(),
        job_tracker.clone(),
        Some(cost_attributor.clone()),
        4, // 4 worker threads
    );
    tracing::info!("Scheduler workers spawned (4 threads)");

    // Create app state
    let app_state = Arc::new(AppState {
        engine: shared_state.engine.clone(),
        pool: shared_state.pool.clone(),
        cache: shared_state.cache.clone(),
        kv: kv_store,
        config: shared_state.config.clone(),
        logger: shared_state.logger,
        monitor: shared_state.monitor.clone(),
        security_monitor,
        package_manager,
        resource_enforcer,
        enterprise_security,
        orchestrator_client: shared_state.orchestrator_client.clone(),
        yara_scanner,
        scheduler: scheduler_for_state,
        micropython_executor,
        shutdown_coordinator: Some(shutdown_coordinator.clone()),
        resource_manager: Some(resource_manager.clone()),
        error_recovery,
        python_pool,
        wasm_pool: shared_state.wasm_pool.clone(),
        wasm_cell_executor,
        memory_layer: Some(memory_layer),
        flymind_client: Some(flymind_client),
        sar_executor,
        agent_scheduler: Some(agent_scheduler),
        job_tracker: Some(job_tracker),
        optimizer: Some(optimizer),
        cost_attributor: Some(cost_attributor),
    });

    // Register resources with ResourceManager for cleanup on shutdown
    {
        let pool_for_cleanup = shared_state.pool.clone();
        let rm = resource_manager.clone();
        rm.register_resource(crate::shutdown::ManagedResource::new("instance_pool", move || {
            let pool = pool_for_cleanup.clone();
            async move {
                let mut p = pool.write().await;
                let stats = p.stats();
                p.clear();
                tracing::info!(
                    "Instance pool cleanup completed: cleared {} instances",
                    stats.total_instances
                );
                Ok(())
            }
        })).await;
    }

    // Register WASM pool for cleanup if enabled
    if let Some(ref wp) = shared_state.wasm_pool {
        let wp_clone = wp.clone();
        let rm = resource_manager.clone();
        rm.register_resource(crate::shutdown::ManagedResource::new("wasm_pool", move || {
            let pool = wp_clone.clone();
            async move {
                pool.clear().await;
                tracing::info!("WASM pool cleanup completed");
                Ok(())
            }
        })).await;
    }

    // Register scheduler for cleanup if enabled
    if let Some(ref sched) = scheduler {
        let sched_clone = sched.clone();
        let rm = resource_manager.clone();
        rm.register_resource(crate::shutdown::ManagedResource::new("scheduler", move || {
            let _scheduler = sched_clone.clone();
            async move {
                // Record scheduler cleanup
                tracing::info!("Scheduler cleanup completed");
                Ok(())
            }
        })).await;
    }

    // Build router
    let app = Router::new()
        .route("/", post(execute_function))
        .route("/execute/chunked", post(execute_chunked))
        .route("/execute/graph", post(execute_graph))
        .route("/execute/{function_id}/{version}", post(execute_function_daemon))
        .route("/health", get(health_check))
        .route("/ready", get(ready_check))
        .route("/monitoring", get(monitoring_stats))
        .route("/budget", get(budget_analysis))
        .route("/security", get(security_status))
        .route("/security/capabilities", get(capability_introspection))
        .route("/errors", get(error_status))
        .route("/scheduler", get(scheduler_status))
        .route("/resources", get(resource_status))
        .route("/kv", get(kv_status))
        .route("/webhook", get(webhook_status))
        .route("/orchestrator", get(orchestrator_status))
        // Prometheus metrics endpoint for unified observability with the Go backend
        .route("/metrics", get(prometheus_metrics))
        // Function metadata and execution result info endpoints
        .route("/function/metadata", get(function_metadata))
        .route("/function/execution-result", get(execution_result_info))
        // Scheduling and metrics endpoints
        .route("/scheduler/simulate", get(scheduling_simulate))
        .route("/scheduler/node/{node_id}/unhealthy", post(scheduler_mark_unhealthy))
        .route("/scheduler/node/{node_id}/healthy", post(scheduler_mark_healthy))
        .route("/scheduler/node/{node_id}", delete(scheduler_remove_node))
        .route("/metrics/execution", get(execution_metrics))
        // Optimizer endpoints (Phase 7: Self-Optimization)
        .route("/graphs/analyze", post(analyze_graph))
        .route("/graphs/{graph_id}/suggestions", get(get_suggestions))
        .route("/graphs/{graph_id}/optimize", post(apply_optimization))
        .route("/graphs/{graph_id}/mutations", get(get_mutation_log))
        .route("/mutations/{mutation_id}/rollback", post(rollback_mutation))
        // Billing endpoints (Phase 6: Cost Attribution)
        .route("/billing/costs", get(get_costs))
        .route("/billing/costs/flush", post(flush_costs))
        .route("/billing/estimate", get(estimate_cost))
        // Resource quota management endpoints
        .route("/resources/limits", post(update_global_limits))
        .route("/resources/quotas", post(set_function_quotas))
        // Runtime management endpoints
        .route("/runtime/config", get(runtime_config))
        .route("/runtime/status", get(runtime_status))
        .route("/runtime/control", post(runtime_control))
        .route("/runtime/metrics", get(runtime_metrics))
        .route("/runtime/shutdown", get(shutdown_status))
        .route("/runtime/engine", get(engine_status))
        .route("/runtime/micropython", get(micropython_status))
        .route("/function/result-examples", get(execution_result_examples))
        // Python cache management endpoints (Phase 3)
        .route("/runtime/python/cache", get(python_cache_status))
        .route("/runtime/python/cache/control", post(python_cache_control))
        // WASI context creation endpoint (for pre-warming and testing)
        .route("/runtime/wasi/create", post(create_wasi_context))
        // Isolation and security utilities
        .route("/security/isolation", get(isolation_utils))
        .layer(cors)
        .layer(TraceLayer::new_for_http())
        .with_state(app_state);

    // Create address
    let addr = SocketAddr::from(([127, 0, 0, 1], port));

    logger.log_with_correlation(
        crate::logging::LogLevel::Info,
        format!("Server listening on http://{}", addr),
        &startup_correlation_id,
    );

    // Start server
    let listener = tokio::net::TcpListener::bind(addr).await?;

    // Apply network namespace isolation after binding (inherited socket still works)
    if config.enable_net_ns {
        logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            "Applying network namespace isolation".to_string(),
            &startup_correlation_id,
        );
        if let Err(e) = crate::netns::apply_net_namespace() {
            // In strict mode, netns failure is fatal; in permissive mode, it's just a warning.
            let level = if config.netns_strict {
                crate::logging::LogLevel::Error
            } else {
                crate::logging::LogLevel::Warn
            };
            logger.log_with_correlation(
                level,
                format!("Failed to apply network namespace: {}", e),
                &startup_correlation_id,
            );
            if config.netns_strict {
                return Err(anyhow::anyhow!("Network namespace isolation is required but could not be applied: {}", e));
            }
        }
    }

    // Apply seccomp-BPF profile after initialization is complete but before
    // serving requests.  This limits the blast radius of any Wasmtime sandbox
    // escape by restricting which host syscalls the runtime process can make.
    if config.enable_seccomp {
        logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            "Applying seccomp-BPF syscall filter".to_string(),
            &startup_correlation_id,
        );
        if let Err(e) = crate::seccomp::apply_seccomp_profile(config.seccomp_strict) {
            // In strict mode, seccomp failure is fatal; in permissive mode, it's just a warning.
            let level = if config.seccomp_strict {
                crate::logging::LogLevel::Error
            } else {
                crate::logging::LogLevel::Warn
            };
            logger.log_with_correlation(
                level,
                format!("Failed to apply seccomp profile: {}", e),
                &startup_correlation_id,
            );
            if config.seccomp_strict {
                return Err(anyhow::anyhow!("Seccomp is required but could not be applied: {}", e));
            }
        }
    }

    // Spawn shutdown signal handler
    let shutdown_logger = logger.clone();
    let shutdown_coordinator_for_signal = shutdown_coordinator.clone();
    let shutdown_coordinator_for_cleanup = shutdown_coordinator.clone();
    let resource_manager_for_cleanup = resource_manager.clone();
    let _component_cleanup_task = tokio::spawn(async move {
        // Create a component shutdown handler to demonstrate its usage
        let coordinator = shutdown_coordinator_for_signal.read().await;
        let mut component = ComponentShutdown::new("server-main", &coordinator);
        drop(coordinator); // Release the read lock

        // Add cleanup tasks that demonstrate ComponentShutdown.wait_and_cleanup
        let cleanup_done = Arc::new(std::sync::atomic::AtomicBool::new(false));
        let cleanup_done_for_task = cleanup_done.clone();
        let cleanup_done_for_verify = cleanup_done.clone();

        component = component.add_cleanup_task(move || {
            let cleanup_done = cleanup_done_for_task;
            async move {
                // Simulate cleanup work
                tokio::time::sleep(std::time::Duration::from_millis(10)).await;
                cleanup_done.store(true, std::sync::atomic::Ordering::Relaxed);
                tracing::info!("Server main component cleanup executed");
            }
        });

        // Wait for shutdown signal and execute cleanup
        component.wait_and_cleanup().await;

        // Verify cleanup was executed
        if cleanup_done_for_verify.load(std::sync::atomic::Ordering::Relaxed) {
            tracing::debug!("Server main component cleanup verified");
        }
    });

    let shutdown_handle = tokio::spawn(async move {
        if let Err(e) = crate::shutdown::handle_shutdown_signals().await {
            tracing::error!("Shutdown signal handler error: {}", e);
        }
        shutdown_logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            "Shutdown signal received, initiating graceful shutdown".to_string(),
            &CorrelationId::new("shutdown-signal".to_string()),
        );

        // Initiate graceful shutdown via the coordinator
        let mut coordinator = shutdown_coordinator_for_cleanup.write().await;
        if let Err(e) = coordinator.shutdown().await {
            tracing::error!("Graceful shutdown failed: {}", e);
        }

        // Clean up all registered resources using ResourceManager
        if let Err(e) = resource_manager_for_cleanup.cleanup_all().await {
            tracing::error!("Resource cleanup failed: {}", e);
        }
    });

    // Run the server with graceful shutdown
    tokio::select! {
        result = axum::serve(listener, app) => {
            if let Err(e) = result {
                tracing::error!("Server error: {}", e);
            }
        }
        _ = shutdown_handle => {
            tracing::info!("Shutdown completed, server stopping...");
        }
    }

    // Resource cleanup is handled via the shutdown coordinator which manages
    // component shutdowns and resource cleanup through ResourceManager
    tracing::info!("Server shutdown complete");
    Ok(())
}

