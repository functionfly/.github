//! FunctionFly Local Runtime
//!
//! A local development runtime that mirrors production behavior.

use anyhow::Result;
use clap::Parser;
use std::sync::Arc;

mod budget;
mod cache;
mod capability;
mod config;
pub mod engine;
mod errors;
mod handlers;
mod host_functions;
mod kv;
mod logging;
mod monitoring;
mod netns;
mod orchestrator_client;
mod package;
pub mod pool;
mod python;
mod python_pool;
mod micropython;
mod enterprise_security;
mod resource_enforcer;
mod scheduler;
mod seccomp;
mod security;
mod server;
mod shutdown;
mod tests;
mod wasi;
mod wasm_interface;
mod yara_scanner;

use config::Config;
use logging::init_structured_logging;
use server::run_server;
use shutdown::{handle_shutdown_signals, ShutdownCoordinator};

/// Initialize structured logging
fn init_structured_logger(config: &Config) -> logging::StructuredLogger {
    init_structured_logging(config.verbose)
}

#[tokio::main]
async fn main() -> Result<()> {
    // Parse command line arguments first
    let config = Config::parse();

    // Initialize structured logging
    let logger = Arc::new(init_structured_logger(&config));

    // Generate correlation ID for startup
    let startup_correlation_id = logger.generate_correlation_id().await;

    // Validate configuration
    if let Err(e) = config.validate() {
        logger.log_with_correlation(
            crate::logging::LogLevel::Error,
            format!("Configuration validation failed: {}", e),
            &startup_correlation_id,
        );
        return Err(anyhow::anyhow!("Configuration validation failed: {}", e));
    }

    logger.log_with_correlation(
        crate::logging::LogLevel::Info,
        format!("Starting FunctionFly local runtime on port {}", config.port),
        &startup_correlation_id,
    );

    // Create shutdown coordinator
    let mut shutdown_coordinator = ShutdownCoordinator::new(Arc::clone(&logger));

    // Spawn server task
    let server_logger = Arc::clone(&logger);
    let server_config = config.clone();
    let server_correlation_id = startup_correlation_id.clone();

    let server_handle = tokio::spawn(async move {
        run_server(
            server_config.port,
            server_config,
            server_logger,
            server_correlation_id,
        ).await
    });

    // Wait for shutdown signal or server error
    tokio::select! {
        result = server_handle => {
            match result {
                Ok(Ok(())) => {
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Info,
                        "Server shut down gracefully",
                        &startup_correlation_id,
                    );
                }
                Ok(Err(e)) => {
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Error,
                        format!("Server error: {}", e),
                        &startup_correlation_id,
                    );
                    return Err(e);
                }
                Err(e) => {
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Error,
                        format!("Server task panicked: {}", e),
                        &startup_correlation_id,
                    );
                    return Err(anyhow::anyhow!("Server task panicked: {}", e));
                }
            }
        }
        _ = handle_shutdown_signals() => {
            logger.log_with_correlation(
                crate::logging::LogLevel::Info,
                "Received shutdown signal, initiating graceful shutdown",
                &startup_correlation_id,
            );

            // Initiate graceful shutdown
            if let Err(e) = shutdown_coordinator.shutdown().await {
                logger.log_with_correlation(
                    crate::logging::LogLevel::Error,
                    format!("Shutdown error: {}", e),
                    &startup_correlation_id,
                );
                return Err(anyhow::anyhow!("Shutdown failed: {}", e));
            }
        }
    }

    logger.log_with_correlation(
        crate::logging::LogLevel::Info,
        "FunctionFly local runtime shutdown complete",
        &startup_correlation_id,
    );

    Ok(())
}
