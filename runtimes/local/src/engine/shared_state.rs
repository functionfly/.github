//! Shared state across requests.

use std::sync::Arc;
use tokio::sync::RwLock;

use crate::cache::ResultCache;
use crate::config::Config;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::monitoring::ResourceMonitor;
use crate::orchestrator_client::OrchestratorClient;
use crate::pool::{InstancePool, PoolManager};
use crate::python::engine::PythonSharedState;

use super::WasmEngine;

/// Shared state across requests
pub struct SharedState {
    pub engine: Arc<WasmEngine>,
    pub pool: Arc<RwLock<InstancePool>>,
    pub cache: Arc<RwLock<ResultCache>>,
    pub kv: SharedKVStore,
    pub config: Config,
    /// Structured logger
    pub logger: StructuredLogger,
    /// Resource monitor
    pub monitor: Arc<ResourceMonitor>,
    /// MicroVM orchestrator client (for Enterprise tier)
    pub orchestrator_client: Option<Arc<OrchestratorClient>>,
    /// WASM instance pool manager for warm-instance reuse
    pub wasm_pool: Option<Arc<PoolManager>>,
}

impl SharedState {
    pub fn new(pool: InstancePool, config: Config, logger: StructuredLogger, security_monitor: Arc<crate::security::SecurityMonitor>) -> Self {
        // Create orchestrator client for Enterprise tier first
        let orchestrator_client = if config.enterprise_enabled {
            Some(Arc::new(OrchestratorClient::new(
                config.orchestrator_url.clone(),
                config.orchestrator_timeout_secs,
            )))
        } else {
            None
        };

        // Create shared Python state for RustPython fallback execution (before WasmEngine)
        let python_shared_state = match PythonSharedState::new(config.clone().into()) {
            Ok(state) => {
                tracing::info!("Shared Python engine initialized");
                Some(Arc::new(state))
            }
            Err(e) => {
                tracing::warn!("Failed to create shared Python engine: {}. \
                    RustPython fallback will create engines per-request.", e);
                None
            }
        };

        // Create WASM engine with logger, orchestrator client, and python shared state
        let engine = match WasmEngine::with_config(
            config.clone(),
            None,
            logger.clone(),
            orchestrator_client.clone(),
            security_monitor,
            python_shared_state.clone(),
        ) {
            Ok(e) => e,
            Err(e) => {
                tracing::error!("Failed to create WASM engine: {}", e);
                panic!("Failed to create WASM engine: {}", e);
            }
        };

        // Create WASM pool manager if enabled
        let wasm_pool: Option<Arc<PoolManager>> = if config.wasm_pool_enabled {
            Some(Arc::new(PoolManager::new(
                config.wasm_pool_max_concurrent,
                config.wasm_pool_max_idle,
            )))
        } else {
            None
        };

        // Attach pool manager to the engine for warm instance reuse
        if let Some(ref pool_mgr) = wasm_pool {
            engine.set_pool_manager(Arc::clone(pool_mgr));
        }

        Self {
            engine: Arc::new(engine),
            pool: Arc::new(RwLock::new(pool)),
            cache: Arc::new(RwLock::new(ResultCache::new(config.cache_ttl))),
            kv: Arc::new(RwLock::new(crate::kv::KVStore::new(10000))), // Max 10k entries
            config: config.clone(),
            logger: logger.clone(),
            monitor: Arc::new(ResourceMonitor::new(Some(Arc::new(logger)))),
            orchestrator_client,
            wasm_pool,
        }
    }
}
