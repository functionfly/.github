//! HTTP server for the local runtime.

use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::RwLock;

use anyhow::Result;
use axum::{
    routing::{get, post},
    Router,
};
use tower_http::cors::{Any, CorsLayer};
use tower_http::trace::TraceLayer;

use crate::cache::ResultCache;
use crate::config::Config;
use crate::engine::{SharedState, WasmEngine};
use crate::handlers::{execute_function, health_check, ready_check, monitoring_stats, budget_analysis, security_status, kv_status, webhook_status, orchestrator_status, AppState};
use crate::kv::SharedKVStore;
use crate::logging::{CorrelationId, StructuredLogger};
use crate::monitoring::ResourceMonitor;
use crate::enterprise_security::EnterpriseSecurityEnforcer;
use crate::package::PackageManager;
use crate::resource_enforcer::ResourceEnforcer;
use crate::security::SecurityMonitor;
use crate::pool::InstancePool;
use crate::shutdown::{handle_shutdown_signals, ShutdownCoordinator};

/// Run the HTTP server with graceful shutdown support
pub async fn run_server(
    port: u16,
    config: Config,
    logger: Arc<StructuredLogger>,
    startup_correlation_id: CorrelationId,
) -> Result<()> {
    // Build CORS layer
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    // Create KV store
    let kv_store: SharedKVStore = Arc::new(RwLock::new(crate::kv::KVStore::new(10000))); // Max 10k entries

    // Create orchestrator client for Enterprise tier
    let orchestrator_client = if config.enterprise_enabled {
        Some(Arc::new(crate::orchestrator_client::OrchestratorClient::new(
            config.orchestrator_url.clone(),
            config.orchestrator_timeout_secs,
        )))
    } else {
        None
    };

    // Create security monitor
    let security_monitor = Arc::new(SecurityMonitor::new());

    // Create Wasm engine with KV store and orchestrator client
    let engine = WasmEngine::with_config(
        config.clone(),
        Some(kv_store.clone()),
        (*logger).clone(),
        orchestrator_client.clone(),
        security_monitor.clone(),
    )?;

    // Create instance pool
    let pool = InstancePool::new(10, 60);

    // Create shared state (includes Python engine if configured)
    let shared_state = SharedState::new(pool, config.clone(), (*logger).clone(), security_monitor.clone());

    // Update shared state with orchestrator client if needed
    // Note: SharedState::new already handles orchestrator client initialization

    // Create security monitor
    let security_monitor = Arc::new(SecurityMonitor::new());

    // Register security profiles
    if config.hardened_security {
        security_monitor.register_profile(
            format!("{}@{}", config.function, config.version),
            SecurityMonitor::create_hardened_profile(),
        ).await;
    } else {
        security_monitor.register_profile(
            format!("{}@{}", config.function, config.version),
            SecurityMonitor::create_standard_profile(),
        ).await;
    }

    // Create package manager for Enterprise tier
    let package_manager = if config.enterprise_enabled && config.package_caching_enabled {
        Some(Arc::new(PackageManager::new(
            shared_state.cache.clone(),
            config.package_cache_dir.clone().into(),
            config.package_cache_size_mb,
        ).unwrap_or_else(|e| {
            tracing::error!("Failed to create package manager: {}", e);
            panic!("Package manager creation failed");
        })))
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
    });

    // Build router
    let app = Router::new()
        .route("/", post(execute_function))
        .route("/health", get(health_check))
        .route("/ready", get(ready_check))
        .route("/monitoring", get(monitoring_stats))
        .route("/budget", get(budget_analysis))
        .route("/security", get(security_status))
        .route("/kv", get(kv_status))
        .route("/webhook", get(webhook_status))
        .route("/orchestrator", get(orchestrator_status))
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
    axum::serve(listener, app).await?;

    Ok(())
}

