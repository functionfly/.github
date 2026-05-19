//! FunctionFly Kotlin Runtime Library
//!
//! Secure production-ready Kotlin/JVM execution runtime with:
//! - WASM sandbox isolation
//! - Resource limits (memory, CPU, time)
//! - Secure execution mode with seccomp/landlock
//! - Package/class restrictions and network controls
//! - Comprehensive metrics and observability
//! - NATS integration with orchestrator

pub mod config;
pub mod execution;
pub mod sandbox;
pub mod security;
pub mod metrics;
pub mod http_server;
pub mod orchestrator_client;

pub use config::{RuntimeConfig, ExecutionLimits, SecurityPolicy, JvmConfig};
pub use execution::{Executor, ExecutionRequest, ExecutionResponse, ValidationResult};
pub use sandbox::{Sandbox, SandboxConfig, SandboxResult};
pub use security::{SecurityManager, Permission, PermissionSet, SecurityViolation, BytecodeValidator};
pub use metrics::{MetricsCollector, RuntimeMetrics};
pub use http_server::{create_app, run_server, run_server_with_shutdown, AppState};
pub use orchestrator_client::{
    OrchestratorClient, NatsConfig, OrchestratorMessage, RuntimeStatus,
    start_heartbeat_loop, start_metrics_loop,
};

use once_cell::sync::Lazy;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

/// Initialize tracing for the runtime
pub fn init_tracing() {
    static TRACING: Lazy<()> = Lazy::new(|| {
        let env_filter = EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| EnvFilter::new("info"));

        tracing_subscriber::registry()
            .with(env_filter)
            .with(tracing_subscriber::fmt::layer().with_target(true))
            .init();
    });
    Lazy::force(&TRACING);
}