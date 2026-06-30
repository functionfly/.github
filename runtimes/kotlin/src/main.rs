//! FunctionFly Kotlin/JVM Runtime
//!
//! A secure production-ready runtime for executing untrusted Kotlin/JVM code
//! with WASM sandbox isolation, resource limits, and comprehensive security.

use anyhow::Result;
use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::signal;
use tokio::sync::RwLock;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use kotlin_runtime::{
    config::RuntimeConfig,
    execution::Executor,
    metrics::MetricsCollector,
    orchestrator_client::{start_heartbeat_loop, start_metrics_loop, NatsConfig, OrchestratorClient},
    AppState,
};

/// Kotlin runtime command-line arguments
#[derive(Parser, Debug)]
#[command(name = "functionfly-kotlin")]
#[command(about = "FunctionFly Kotlin/JVM Runtime - Secure production-ready code execution")]
struct Args {
    /// Server address to listen on
    #[arg(short, long, default_value = "127.0.0.1:8091")]
    addr: SocketAddr,

    /// Enable debug logging
    #[arg(short, long)]
    debug: bool,

    /// Maximum concurrent executions
    #[arg(long, default_value = "100")]
    max_concurrent: usize,

    /// Maximum memory per execution (MB)
    #[arg(long, default_value = "256")]
    max_memory_mb: u64,

    /// Maximum CPU time per execution (seconds)
    #[arg(long, default_value = "10")]
    max_cpu_time_secs: u64,

    /// Maximum wall time per execution (seconds)
    #[arg(long, default_value = "30")]
    max_wall_time_secs: u64,

    /// Enable sandbox mode
    #[arg(long, default_value = "true")]
    sandbox: bool,

    /// NATS URL (optional)
    #[arg(long)]
    nats_url: Option<String>,

    /// Runtime ID for NATS (optional, auto-generated if not provided)
    #[arg(long)]
    runtime_id: Option<String>,
}

impl Args {
    /// Build runtime configuration from arguments
    fn to_config(&self) -> RuntimeConfig {
        RuntimeConfig {
            limits: kotlin_runtime::config::ExecutionLimits {
                max_memory_mb: self.max_memory_mb,
                max_cpu_time_secs: self.max_cpu_time_secs,
                max_wall_time_secs: self.max_wall_time_secs,
                ..Default::default()
            },
            ..Default::default()
        }
    }

    /// Get NATS configuration
    fn nats_config(&self) -> Option<NatsConfig> {
        self.nats_url.as_ref().map(|url| {
            let mut config = NatsConfig::default();
            config.url = url.clone();
            if let Some(ref id) = self.runtime_id {
                config.runtime_id = id.clone();
            }
            config
        })
    }
}

/// Initialize logging
fn init_logging(debug: bool) {
    let env_filter = if debug {
        EnvFilter::new("debug")
    } else {
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info"))
    };

    tracing_subscriber::registry()
        .with(env_filter)
        .with(tracing_subscriber::fmt::layer().with_target(true))
        .init();
}

#[tokio::main]
async fn main() -> Result<()> {
    // Parse command line arguments
    let args = Args::parse();

    // Initialize logging
    init_logging(args.debug);

    tracing::info!(
        "Starting FunctionFly Kotlin Runtime v{}",
        env!("CARGO_PKG_VERSION")
    );
    tracing::info!("Listening on {}", args.addr);

    let api_token = std::env::var("RUNTIME_API_TOKEN").ok().filter(|t| !t.is_empty());
    let is_production = std::env::var("ENVIRONMENT")
        .map(|v| v.eq_ignore_ascii_case("production"))
        .unwrap_or(false);

    if api_token.is_none() {
        if is_production {
            anyhow::bail!(
                "RUNTIME_API_TOKEN is required in production (ENVIRONMENT=production). \
                 Set the token and restart."
            );
        } else {
            tracing::warn!("RUNTIME_API_TOKEN not set — /execute endpoint is unauthenticated (dev mode)");
        }
    }

    // Build configuration
    let config = args.to_config();

    // Create metrics collector
    let metrics = MetricsCollector::new("kotlin-runtime");

    // Create executor
    let executor = Executor::new(config.clone(), Arc::new(metrics))?;

    tracing::info!("Executor initialized with max_concurrent={}", config.max_concurrent);

    // Optionally connect to NATS
    let nats_client = if let Some(nats_config) = args.nats_config() {
        let client = Arc::new(RwLock::new(OrchestratorClient::new(nats_config)));

        if let Err(e) = client.write().await.connect().await {
            tracing::warn!("Failed to connect to NATS: {}. Continuing without NATS.", e);
        } else {
            // Register with orchestrator
            if let Err(e) = client.read().await.register().await {
                tracing::warn!("Failed to register with orchestrator: {}", e);
            } else {
                // Start heartbeat loop
                let client_clone = client.clone();
                tokio::spawn(async move {
                    if let Err(e) = start_heartbeat_loop(client_clone, 30).await {
                        tracing::error!("Heartbeat loop error: {}", e);
                    }
                });

                // Start metrics reporting loop
                let client_clone = client.clone();
                let metrics_clone = Arc::new(MetricsCollector::new("kotlin-runtime"));
                tokio::spawn(async move {
                    if let Err(e) = start_metrics_loop(client_clone, metrics_clone, 60).await {
                        tracing::error!("Metrics loop error: {}", e);
                    }
                });

                tracing::info!("Connected to NATS orchestrator");
            }
        }

        Some(client)
    } else {
        tracing::info!("NATS not configured, running in standalone mode");
        None
    };

    // Run the server
    tracing::info!("Server listening on {}", args.addr);
    let listener = tokio::net::TcpListener::bind(args.addr).await?;

    // Graceful shutdown
    let shutdown = async {
        signal::ctrl_c()
            .await
            .expect("failed to install CTRL+C handler");
        tracing::info!("Shutdown signal received");
    };

    // Clone config for server
    let config_clone = config.clone();
    let metrics_for_server = MetricsCollector::new("kotlin-runtime");

    axum::serve(listener, kotlin_runtime::http_server::create_app(
        AppState::new(executor, metrics_for_server, config_clone).with_auth(api_token)
    ))
    .with_graceful_shutdown(shutdown)
    .await?;

    // Cleanup: deregister from NATS
    if let Some(ref client) = nats_client {
        if let Err(e) = client.read().await.deregister("shutdown".to_string()).await {
            tracing::warn!("Failed to deregister from orchestrator: {}", e);
        }
    }

    tracing::info!("Kotlin runtime stopped");
    Ok(())
}