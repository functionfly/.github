//! FunctionFly Ruby Runtime Binary
//!
//! Secure production-ready Ruby execution runtime with:
//! - HTTP server for the Go orchestrator
//! - Orchestrator communication via NATS
//! - Process isolation and seccomp enforcement
//! - Active code blocking and security auditing
//! - Circuit breaker and graceful shutdown

use ruby_runtime::{
    init_tracing, config::RuntimeConfig,
    OrchestratorClient, SecurityManager,
    MetricsCollector, ExecutionLimits, DefaultExecutor,
    Executor, http_server,
};
use parking_lot::RwLock;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::interval;
use tracing::{info, warn};

#[derive(Debug, Clone)]
struct Args {
    port: u16,
    max_concurrent: usize,
    max_memory_mb: u64,
    max_execution_time_secs: u64,
    sandbox_enabled: bool,
    nats_url: Option<String>,
}

impl Default for Args {
    fn default() -> Self {
        Self {
            port: 8092,
            max_concurrent: 100,
            max_memory_mb: 256,
            max_execution_time_secs: 30,
            sandbox_enabled: true,
            nats_url: std::env::var("NATS_URL").ok(),
        }
    }
}

impl Args {
    fn from_env() -> Self {
        let is_production = std::env::var("ENVIRONMENT")
            .map(|v| v.eq_ignore_ascii_case("production"))
            .unwrap_or(false);

        // In production, sandbox is always enabled regardless of env var
        let sandbox_enabled = if is_production {
            true
        } else {
            std::env::var("SANDBOX_ENABLED")
                .unwrap_or_else(|_| "true".to_string())
                .to_lowercase() != "false"
        };

        Self {
            port: std::env::var("PORT")
                .unwrap_or_else(|_| "8092".to_string())
                .parse()
                .unwrap_or(8092),
            max_concurrent: std::env::var("MAX_CONCURRENT")
                .unwrap_or_else(|_| "100".to_string())
                .parse()
                .unwrap_or(100),
            max_memory_mb: std::env::var("MAX_MEMORY_MB")
                .unwrap_or_else(|_| "256".to_string())
                .parse()
                .unwrap_or(256),
            max_execution_time_secs: std::env::var("MAX_EXECUTION_TIME_SECS")
                .unwrap_or_else(|_| "30".to_string())
                .parse()
                .unwrap_or(30),
            sandbox_enabled,
            nats_url: std::env::var("NATS_URL").ok(),
        }
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    init_tracing();

    let args = Args::from_env();

    // Initialize start time for uptime tracking
    http_server::init_start_time();

    info!(
        version = env!("CARGO_PKG_VERSION"),
        port = args.port,
        max_concurrent = args.max_concurrent,
        max_memory_mb = args.max_memory_mb,
        sandbox_enabled = args.sandbox_enabled,
        nats_url = ?args.nats_url,
        "Starting FunctionFly Ruby Runtime - Production Secure"
    );

    let api_token = std::env::var("RUNTIME_API_TOKEN").ok().filter(|t| !t.is_empty());
    let is_production = std::env::var("ENVIRONMENT")
        .map(|v| v.eq_ignore_ascii_case("production"))
        .unwrap_or(false);

    if api_token.is_none() {
        if is_production {
            tracing::error!(
                "RUNTIME_API_TOKEN is not set in production. \
                 The /execute endpoint is UNAUTHENTICATED. Set the token and restart."
            );
        } else {
            warn!("RUNTIME_API_TOKEN not set — /execute endpoint is unauthenticated (dev mode)");
        }
    }

    let config = RuntimeConfig {
        limits: ExecutionLimits {
            max_memory_mb: args.max_memory_mb,
            max_cpu_time_secs: args.max_execution_time_secs,
            max_wall_time_secs: args.max_execution_time_secs * 2,
            max_output_bytes: 1024 * 1024, // 1MB max output
            max_stack_depth: 1024,
            max_allocations: 10000,
        },
        security: ruby_runtime::config::SecurityPolicy {
            sandbox_enabled: args.sandbox_enabled,
            enable_seccomp: args.sandbox_enabled,
            enable_landlock: args.sandbox_enabled,
            allow_filesystem: false,
            allow_network: false,
            sanitize_code: true,
            blocked_syscalls: vec![
                "fork".to_string(),
                "vfork".to_string(),
                "clone".to_string(),
                "execve".to_string(),
                "mount".to_string(),
                "ptrace".to_string(),
            ],
            allowed_dirs: vec![],
            max_require_depth: 16,
        },
        ruby: ruby_runtime::config::RubyConfig::default(),
        use_sandbox: args.sandbox_enabled,
        max_concurrent: args.max_concurrent,
        default_timeout: Duration::from_secs(args.max_execution_time_secs),
    };

    config.validate().map_err(|e| anyhow::anyhow!("invalid config: {}", e))?;

    // Create security manager with auditing
    let security = SecurityManager::new(config.security.clone());
    let auditor = security.auditor();
    let metrics = Arc::new(MetricsCollector::new());

    // Create executor
    let executor = DefaultExecutor::new(config.clone(), security, metrics.clone());

    // Create orchestrator client
    let mut orchestrator = OrchestratorClient::new("ruby");
    if let Some(ref nats_url) = args.nats_url {
        orchestrator = orchestrator.with_nats_url(nats_url);
        if let Err(e) = orchestrator.connect() {
            warn!(error = %e, "Failed to connect to NATS, running in standalone mode");
        } else {
            if let Err(e) = orchestrator.register_runtime(vec!["ruby".to_string()]) {
                warn!(error = %e, "Failed to register with orchestrator");
            } else {
                info!(runtime_id = %orchestrator.runtime_id(), "Registered with orchestrator");
            }
        }
    }

    let orchestrator = Arc::new(RwLock::new(orchestrator));

    // Start background tasks
    let orchestrator_for_heartbeat = orchestrator.clone();
    let orchestrator_for_metrics = orchestrator.clone();

    // Heartbeat task
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(30));
        loop {
            interval.tick().await;

            let client = orchestrator_for_heartbeat.read();
            if client.is_registered() {
                if let Err(e) = client.send_heartbeat("healthy") {
                    warn!(error = %e, "Failed to send heartbeat");
                }
            }
        }
    });

    // Metrics reporting task
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(60));
        loop {
            interval.tick().await;

            let client = orchestrator_for_metrics.read();
            if client.is_registered() {
                if let Err(e) = client.report_metrics(0.0, 0, 0) {
                    warn!(error = %e, "Failed to report metrics");
                }
            }
        }
    });

    // Run the HTTP server with full security features
    http_server::run_server(
        config,
        args.port,
        executor,
        metrics,
        orchestrator,
        auditor,
        api_token,
    ).await?;

    Ok(())
}