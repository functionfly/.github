//! FunctionFly Bun Runtime Library
//!
//! Secure production-ready JavaScript/TypeScript execution runtime with:
//! - WebAssembly sandbox isolation
//! - Resource limits (memory, CPU, time)
//! - Secure execution mode with seccomp/landlock
//! - Module restrictions and network controls
//! - Comprehensive metrics and observability

pub mod config;
pub mod execution;
pub mod sandbox;
pub mod security;
pub mod metrics;
pub mod http_server;
pub mod orchestrator_client;

pub use config::{RuntimeConfig, ExecutionLimits, SecurityPolicy};
pub use execution::{Executor, ExecutionRequest, ExecutionResponse};
pub use sandbox::{Sandbox, SandboxConfig, SandboxResult};
pub use security::{SecurityManager, Permission, PermissionSet};
pub use metrics::{MetricsCollector, RuntimeMetrics};
pub use http_server::{create_app, run_server, AppState};
pub use orchestrator_client::{
    OrchestratorClient, FunctionExecutionRequest, FunctionExecutionResponse,
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