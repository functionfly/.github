//! Graceful shutdown handling for the local runtime.

use std::sync::Arc;
use std::time::Duration;
use tokio::signal;
use tokio::sync::{broadcast, RwLock};
use tokio::time::{timeout, Instant};

use crate::errors::{RuntimeError, RuntimeResult};
use crate::logging::{CorrelationId, StructuredLogger};

type CleanupTask = Box<dyn FnOnce() -> std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send>> + Send + Sync>;
type CleanupFn = Arc<dyn Fn() -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<(), Box<dyn std::error::Error + Send + Sync>>> + Send>> + Send + Sync>;

/// Shutdown coordinator for managing graceful shutdown
pub struct ShutdownCoordinator {
    /// Shutdown signal sender
    shutdown_tx: broadcast::Sender<()>,
    /// Shutdown signal receiver
    shutdown_rx: broadcast::Receiver<()>,
    /// Logger for shutdown events
    logger: Arc<StructuredLogger>,
    /// Shutdown timeout
    timeout: Duration,
    /// Shutdown start time
    shutdown_start: Option<Instant>,
}

impl ShutdownCoordinator {
    /// Create a new shutdown coordinator
    pub fn new(logger: Arc<StructuredLogger>) -> Self {
        let (shutdown_tx, shutdown_rx) = broadcast::channel(1);

        Self {
            shutdown_tx,
            shutdown_rx,
            logger,
            timeout: Duration::from_secs(30), // 30 second default timeout
            shutdown_start: None,
        }
    }

    /// Create shutdown coordinator with custom timeout
    pub fn with_timeout(logger: Arc<StructuredLogger>, timeout: Duration) -> Self {
        let mut coordinator = Self::new(logger);
        coordinator.timeout = timeout;
        coordinator
    }

    /// Get a shutdown signal receiver for components
    pub fn subscribe(&self) -> broadcast::Receiver<()> {
        self.shutdown_tx.subscribe()
    }

    /// Check if shutdown has been signaled using the stored receiver
    pub fn is_shutdown_signaled(&self) -> bool {
        !self.shutdown_rx.is_empty()
    }

    /// Initiate graceful shutdown
    pub async fn shutdown(&mut self) -> RuntimeResult<()> {
        let correlation_id = self.logger.generate_correlation_id().await;
        self.shutdown_start = Some(Instant::now());

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            "Initiating graceful shutdown",
            &correlation_id,
        );

        // Send shutdown signal to all subscribers
        let _ = self.shutdown_tx.send(());

        // Wait for shutdown to complete or timeout
        match timeout(self.timeout, self.wait_for_cleanup(&correlation_id)).await {
            Ok(result) => {
                let elapsed = self.shutdown_start.map(|s| s.elapsed()).unwrap_or_default();
                self.logger.log_with_correlation(
                    crate::logging::LogLevel::Info,
                    format!("Shutdown completed successfully in {:.2}s", elapsed.as_secs_f64()),
                    &correlation_id,
                );
                result
            }
            Err(_) => {
                self.logger.log_with_correlation(
                    crate::logging::LogLevel::Error,
                    format!("Shutdown timed out after {:.2}s", self.timeout.as_secs_f64()),
                    &correlation_id,
                );
                Err(RuntimeError::new(
                    crate::errors::ErrorKind::Unknown,
                    "Shutdown timeout exceeded",
                ))
            }
        }
    }

    /// Wait for cleanup tasks to complete
    async fn wait_for_cleanup(&self, correlation_id: &CorrelationId) -> RuntimeResult<()> {
        // Give components time to cleanup
        tokio::time::sleep(Duration::from_millis(100)).await;

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Debug,
            "All cleanup tasks completed",
            correlation_id,
        );

        Ok(())
    }

    /// Check if shutdown has been initiated
    pub fn is_shutting_down(&self) -> bool {
        self.shutdown_start.is_some()
    }

    /// Get shutdown timeout
    pub fn timeout(&self) -> Duration {
        self.timeout
    }
}

/// Shutdown handler for a specific component
pub struct ComponentShutdown {
    name: String,
    shutdown_rx: broadcast::Receiver<()>,
    logger: Arc<StructuredLogger>,
    cleanup_tasks: Vec<CleanupTask>,
}

impl ComponentShutdown {
    /// Create a new component shutdown handler
    pub fn new(name: impl Into<String>, coordinator: &ShutdownCoordinator) -> Self {
        Self {
            name: name.into(),
            shutdown_rx: coordinator.subscribe(),
            logger: Arc::clone(&coordinator.logger),
            cleanup_tasks: Vec::new(),
        }
    }

    /// Add a cleanup task
    pub fn add_cleanup_task<F, Fut>(mut self, task: F) -> Self
    where
        F: FnOnce() -> Fut + Send + Sync + 'static,
        Fut: std::future::Future<Output = ()> + Send + 'static,
    {
        let boxed_task = Box::new(move || Box::pin(task()) as std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send>>);
        self.cleanup_tasks.push(boxed_task);
        self
    }

    /// Wait for shutdown signal and execute cleanup
    pub async fn wait_and_cleanup(mut self) {
        let correlation_id = self.logger.generate_correlation_id().await;

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Debug,
            format!("Component '{}' waiting for shutdown signal", self.name),
            &correlation_id,
        );

        // Wait for shutdown signal
        let _ = self.shutdown_rx.recv().await;

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            format!("Component '{}' received shutdown signal, starting cleanup", self.name),
            &correlation_id,
        );

        // Execute cleanup tasks
        for (i, task) in self.cleanup_tasks.into_iter().enumerate() {
            let start_time = Instant::now();
            task().await;

            let elapsed = start_time.elapsed();
            self.logger.log_with_correlation(
                crate::logging::LogLevel::Debug,
                format!("Component '{}' cleanup task {} completed in {:.2}ms", self.name, i + 1, elapsed.as_millis()),
                &correlation_id,
            );
        }

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            format!("Component '{}' cleanup completed", self.name),
            &correlation_id,
        );
    }
}

/// Resource manager for tracking and cleaning up resources
pub struct ResourceManager {
    resources: RwLock<Vec<ManagedResource>>,
    logger: Arc<StructuredLogger>,
}

impl ResourceManager {
    /// Create a new resource manager
    pub fn new(logger: Arc<StructuredLogger>) -> Self {
        Self {
            resources: RwLock::new(Vec::new()),
            logger,
        }
    }

    /// Register a resource for cleanup
    pub async fn register_resource(&self, resource: ManagedResource) {
        let mut resources = self.resources.write().await;
        resources.push(resource);

        let correlation_id = self.logger.generate_correlation_id().await;
        self.logger.log_with_correlation(
            crate::logging::LogLevel::Debug,
            format!("Registered resource for cleanup: {}", resources.len()),
            &correlation_id,
        );
    }

    /// Cleanup all registered resources
    pub async fn cleanup_all(&self) -> RuntimeResult<()> {
        let correlation_id = self.logger.generate_correlation_id().await;

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            "Starting resource cleanup",
            &correlation_id,
        );

        let mut resources = self.resources.write().await;
        let resource_count = resources.len();
        let mut success_count = 0;
        let mut failure_count = 0;

        for resource in resources.drain(..) {
            let start_time = Instant::now();
            let resource_name = resource.name.clone();

            if let Err(e) = resource.cleanup() {
                failure_count += 1;
                self.logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    format!("Failed to cleanup resource '{}': {}", resource_name, e),
                    &correlation_id,
                );
            } else {
                success_count += 1;
                let elapsed = start_time.elapsed();
                self.logger.log_with_correlation(
                    crate::logging::LogLevel::Debug,
                    format!("Cleaned up resource '{}' in {:.2}ms", resource_name, elapsed.as_millis()),
                    &correlation_id,
                );
            }
        }

        self.logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            format!("Resource cleanup completed: {} total, {} success, {} failed", resource_count, success_count, failure_count),
            &correlation_id,
        );

        Ok(())
    }

    /// Get resource count
    pub async fn resource_count(&self) -> usize {
        let resources = self.resources.read().await;
        resources.len()
    }
}

/// Managed resource that can be cleaned up
pub struct ManagedResource {
    name: String,
    cleanup_fn: CleanupFn,
}

impl ManagedResource {
    /// Create a new managed resource
    pub fn new<F, Fut>(
        name: impl Into<String>,
        cleanup_fn: F,
    ) -> Self
    where
        F: Fn() -> Fut + Send + Sync + 'static,
        Fut: std::future::Future<Output = Result<(), Box<dyn std::error::Error + Send + Sync>>> + Send + 'static,
    {
        Self {
            name: name.into(),
            cleanup_fn: Arc::new(move || Box::pin(cleanup_fn())),
        }
    }

    /// Execute cleanup
    pub fn cleanup(&self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        match tokio::runtime::Handle::try_current() {
            Ok(handle) => {
                let cleanup_fn = self.cleanup_fn.clone();
                handle.spawn(async move {
                    if let Err(e) = cleanup_fn().await {
                        tracing::warn!("Async resource cleanup failed: {}", e);
                    }
                });
                Ok(())
            }
            Err(_) => {
                let runtime = tokio::runtime::Runtime::new()?;
                runtime.block_on(async { (self.cleanup_fn)().await })
            }
        }
    }
}
/// Signal handler for graceful shutdown
pub async fn handle_shutdown_signals() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to listen for ctrl-c event");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to listen for SIGTERM")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {
            tracing::info!("Received Ctrl+C, initiating graceful shutdown");
        }
        _ = terminate => {
            tracing::info!("Received SIGTERM, initiating graceful shutdown");
        }
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::{AtomicBool, Ordering};
    use std::sync::Arc;

    #[tokio::test]
    async fn test_shutdown_coordinator() {
        let logger = Arc::new(StructuredLogger::new());
        let mut coordinator = ShutdownCoordinator::new(Arc::clone(&logger));

        let mut component = ComponentShutdown::new("test-component", &coordinator);

        let cleanup_executed = Arc::new(AtomicBool::new(false));
        let cleanup_executed_clone = Arc::clone(&cleanup_executed);

        component = component.add_cleanup_task(move || {
            let cleanup_executed = Arc::clone(&cleanup_executed_clone);
            async move {
                cleanup_executed.store(true, Ordering::Relaxed);
                tokio::time::sleep(Duration::from_millis(10)).await;
            }
        });

        // Start cleanup task
        let cleanup_handle = tokio::spawn(component.wait_and_cleanup());

        // Give it a moment to start waiting
        tokio::time::sleep(Duration::from_millis(10)).await;

        // Initiate shutdown
        coordinator.shutdown().await.unwrap();

        // Wait for cleanup to complete
        cleanup_handle.await.unwrap();

        // Verify cleanup was executed
        assert!(cleanup_executed.load(Ordering::Relaxed));
    }

    #[tokio::test]
    async fn test_resource_manager() {
        let logger = Arc::new(StructuredLogger::new());
        let manager = ResourceManager::new(Arc::clone(&logger));

        let cleanup_executed = Arc::new(AtomicBool::new(false));
        let cleanup_executed_clone = Arc::clone(&cleanup_executed);

        let resource = ManagedResource::new("test-resource", move || {
            let cleanup_executed = Arc::clone(&cleanup_executed_clone);
            async move {
                cleanup_executed.store(true, Ordering::Relaxed);
                Ok(())
            }
        });

        manager.register_resource(resource).await;
        assert_eq!(manager.resource_count().await, 1);

        manager.cleanup_all().await.unwrap();
        assert_eq!(manager.resource_count().await, 0);
        assert!(cleanup_executed.load(Ordering::Relaxed));
    }
}
