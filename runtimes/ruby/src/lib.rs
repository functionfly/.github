//! FunctionFly Ruby Runtime Library
//!
//! Secure production-ready Ruby execution runtime with:
//! - WebAssembly sandbox isolation
//! - Process isolation with seccomp and landlock
//! - Resource limits (memory, CPU, time)
//! - Active code blocking for dangerous patterns
//! - Security auditing and audit logging
//! - Rate limiting and circuit breaker
//! - TLS/HTTPS support
//! - NATS communication with orchestrator

pub mod config;
pub mod execution;
pub mod sandbox;
pub mod security;
pub mod metrics;
pub mod http_server;
pub mod orchestrator_client;

pub use config::{RuntimeConfig, ExecutionLimits, SecurityPolicy, RubyConfig};
pub use execution::{Executor, ExecutionRequest, ExecutionResponse, DefaultExecutor, TenantContext};
pub use sandbox::{Sandbox, SandboxConfig, SandboxResult};
pub use security::{
    SecurityManager, Permission, PermissionSet, SecurityAuditor,
    SecurityEvent, SecurityEventType, SecuritySeverity,
    CodeValidationResult, CodeViolation,
};
pub use metrics::{MetricsCollector, RuntimeMetrics};
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