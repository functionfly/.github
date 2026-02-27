//! FunctionFly MicroVM Orchestrator
//!
//! This module provides CPython execution inside Firecracker microVMs
//! for Enterprise tier customers.

mod executor;
mod firecracker;
mod orchestrator;
mod vsock;

use anyhow::Result;
use clap::Parser;
use orchestrator::MicroVMOrchestrator;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{error, info};

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

    /// Enable debug logging
    #[arg(long, short)]
    debug: bool,
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

    info!("Starting FunctionFly MicroVM Orchestrator");
    info!("Configuration: {} vCPUs, {}MB memory, max {} VMs",
          args.vcpus, args.memory_mb, args.max_vms);

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

    // Set up graceful shutdown
    let orchestrator_clone = Arc::clone(&orchestrator);
    tokio::signal::ctrl_c().await?;

    info!("Shutting down MicroVM Orchestrator...");

    // Cleanup all VMs
    let mut orch = orchestrator_clone.write().await;
    orch.shutdown().await?;

    info!("MicroVM Orchestrator stopped");

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
