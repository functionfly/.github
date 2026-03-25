//! FunctionFly MicroVM Orchestrator
//!
//! This module provides CPython execution inside Firecracker microVMs
//! for Enterprise tier customers.
//!
//! Exposes HTTP API (default :9091): /execute, /health, /stats, /metrics

mod executor;
mod firecracker;
mod firecracker_spawn;
mod http_server;
mod orchestrator;
mod vsock;

use anyhow::Result;
use clap::Parser;
use orchestrator::MicroVMOrchestrator;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use tokio::time::MissedTickBehavior;
use tracing::{error, info};
use tower_http::cors::{Any, CorsLayer};

/// MicroVM Orchestrator CLI
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Firecracker socket path
    #[arg(long, default_value = "/var/run/firecracker.sock")]
    socket_path: String,

    /// VM image path
    #[arg(long, default_value = "/var/lib/functionfly/vmimages")]
    image_path: String,

    /// Number of vCPUs per VM
    #[arg(long, default_value = "2")]
    vcpus: u32,

    /// Memory in MB per VM
    #[arg(long, default_value = "512")]
    memory_mb: u32,

    /// Maximum concurrent VMs
    #[arg(long, default_value = "100")]
    max_vms: u32,

    /// HTTP server port. Default 9091 (9090 is reserved for the main Prometheus metrics port).
    #[arg(long, default_value = "9091")]
    port: u16,

    /// Enable debug logging
    #[arg(long, short)]
    debug: bool,

    /// Remove warm-pool VMs idle longer than this (seconds). 0 disables periodic cleanup.
    #[arg(long, default_value = "600")]
    warm_idle_secs: u64,

    /// How often to run warm-pool idle cleanup (seconds).
    #[arg(long, default_value = "60")]
    cleanup_interval_secs: u64,
}

#[tokio::main]
async fn main() -> Result<()> {
    // Parse arguments
    let args = Args::parse();

    // Initialize logging
    let log_level = if args.debug {
        tracing::Level::DEBUG
    } else {
        tracing::Level::INFO
    };

    tracing_subscriber::fmt()
        .with_max_level(log_level)
        .with_target(false)
        .init();

    // Safety guard: refuse to start in production if dev mode is enabled.
    let dev_mode = std::env::var("FUNCTIONFLY_MICROVM_DEV_MODE")
        .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
        .unwrap_or(false);
    let is_production = std::env::var("ENVIRONMENT")
        .map(|v| v.eq_ignore_ascii_case("production"))
        .unwrap_or(false);
    if dev_mode && is_production {
        anyhow::bail!(
            "FUNCTIONFLY_MICROVM_DEV_MODE must NOT be set in a production environment \
             (ENVIRONMENT=production). Remove the variable and restart."
        );
    }

    // Resolve API bearer token from env (recommended) or absent (dev-only).
    let api_token: Option<Arc<str>> = std::env::var("FUNCTIONFLY_MICROVM_API_TOKEN")
        .ok()
        .filter(|t| !t.is_empty())
        .map(|t| Arc::from(t.as_str()));
    if api_token.is_none() && is_production {
        error!(
            "FUNCTIONFLY_MICROVM_API_TOKEN is not set in a production environment. \
             The /execute endpoint is unauthenticated — set the token and restart."
        );
        // Don't bail — operator may have network-level auth; log loudly instead.
    }

    info!("Starting FunctionFly MicroVM Orchestrator");
    info!("Configuration: {} vCPUs, {}MB memory, max {} VMs, port {}",
          args.vcpus, args.memory_mb, args.max_vms, args.port);
    info!("Auth: {}", if api_token.is_some() { "bearer token enabled" } else { "NO AUTH (dev)" });

    // Create orchestrator
    let orchestrator = Arc::new(RwLock::new(
        MicroVMOrchestrator::new(
            args.socket_path.clone(),
            args.image_path.clone(),
            args.vcpus,
            args.memory_mb,
            args.max_vms,
        ).await?
    ));

    let state = http_server::AppState {
        orchestrator: Arc::clone(&orchestrator),
        metrics: Arc::new(http_server::HttpMetrics::default()),
        api_token,
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = http_server::router(state).layer(cors);

    let addr = SocketAddr::from(([0, 0, 0, 0], args.port));
    let listener = tokio::net::TcpListener::bind(addr).await?;
    info!("MicroVM Orchestrator HTTP API listening on http://{}", addr);

    let server = axum::serve(listener, app);

    let mut cleanup_handle: Option<tokio::task::JoinHandle<()>> = None;
    if args.warm_idle_secs > 0 {
        let period = Duration::from_secs(args.cleanup_interval_secs.max(1));
        let idle = args.warm_idle_secs;
        let orch_cleanup = Arc::clone(&orchestrator);
        cleanup_handle = Some(tokio::spawn(async move {
            let mut interval = tokio::time::interval(period);
            interval.set_missed_tick_behavior(MissedTickBehavior::Skip);
            loop {
                interval.tick().await;
                let mut o = orch_cleanup.write().await;
                o.cleanup_idle_vms(idle).await;
            }
        }));
        info!(
            "Warm-pool idle cleanup: every {}s, evict after {}s idle",
            args.cleanup_interval_secs.max(1),
            idle
        );
    } else {
        info!("Warm-pool idle cleanup disabled (--warm-idle-secs 0)");
    }

    tokio::select! {
        r = server => {
            if let Some(h) = cleanup_handle.take() {
                h.abort();
                let _ = h.await;
            }
            if let Err(e) = r {
                error!("HTTP server error: {}", e);
            }
        }
        _ = tokio::signal::ctrl_c() => {
            info!("Shutting down MicroVM Orchestrator...");
            if let Some(h) = cleanup_handle.take() {
                h.abort();
                let _ = h.await;
            }
            let mut orch = orchestrator.write().await;
            orch.shutdown().await?;
            info!("MicroVM Orchestrator stopped");
        }
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_args_parsing() {
        let args = Args::parse_from(&[
            "microvm",
            "--vcpus", "4",
            "--memory-mb", "1024",
            "--max-vms", "50",
        ]);

        assert_eq!(args.vcpus, 4);
        assert_eq!(args.memory_mb, 1024);
        assert_eq!(args.max_vms, 50);
    }
}
