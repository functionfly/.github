//! Structured logging with correlation IDs and contextual information.

use std::collections::HashMap;
use std::fmt;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{span, Level};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

/// Global logger instance for managing correlation IDs and structured logging
#[derive(Clone)]
pub struct StructuredLogger {
    correlation_id_generator: Arc<RwLock<CorrelationIdGenerator>>,
}

/// Correlation ID generator for request tracing
pub struct CorrelationIdGenerator {
    next_id: u64,
    prefix: String,
}

/// Correlation ID for tracing requests across the system
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct CorrelationId(String);

/// Structured log entry with context
#[derive(Debug, Clone)]
pub struct LogEntry {
    pub level: LogLevel,
    pub message: String,
    pub correlation_id: Option<CorrelationId>,
    pub fields: HashMap<String, serde_json::Value>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

/// Log level enum
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LogLevel {
    Trace,
    Debug,
    Info,
    Warn,
    Error,
}

impl From<LogLevel> for Level {
    fn from(level: LogLevel) -> Level {
        match level {
            LogLevel::Trace => Level::TRACE,
            LogLevel::Debug => Level::DEBUG,
            LogLevel::Info => Level::INFO,
            LogLevel::Warn => Level::WARN,
            LogLevel::Error => Level::ERROR,
        }
    }
}

impl CorrelationId {
    /// Create a new correlation ID
    pub fn new(id: String) -> Self {
        Self(id)
    }

    /// Generate a new unique correlation ID
    pub fn generate() -> Self {
        use std::time::{SystemTime, UNIX_EPOCH};
        let timestamp = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_nanos();

        Self(format!("req_{:016x}", timestamp))
    }

    /// Get the string representation
    pub fn as_str(&self) -> &str {
        &self.0
    }
}

impl fmt::Display for CorrelationId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl CorrelationIdGenerator {
    /// Create a new generator
    pub fn new(prefix: impl Into<String>) -> Self {
        Self {
            next_id: 0,
            prefix: prefix.into(),
        }
    }

    /// Generate next correlation ID
    pub fn next(&mut self) -> CorrelationId {
        self.next_id += 1;
        CorrelationId::new(format!("{}_{}", self.prefix, self.next_id))
    }
}

// Explicitly implement Send + Sync for CorrelationIdGenerator
// This is safe because u64 and String are Send + Sync
unsafe impl Send for CorrelationIdGenerator {}
unsafe impl Sync for CorrelationIdGenerator {}

impl StructuredLogger {
    /// Create a new structured logger
    pub fn new() -> Self {
        Self {
            correlation_id_generator: Arc::new(RwLock::new(CorrelationIdGenerator::new("ffly"))),
        }
    }

    /// Generate a new correlation ID
    pub async fn generate_correlation_id(&self) -> CorrelationId {
        let mut generator = self.correlation_id_generator.write().await;
        generator.next()
    }


    /// Log a structured entry
    pub fn log(&self, level: LogLevel, message: impl Into<String>) {
        let message = message.into();
        match level {
            LogLevel::Trace => tracing::trace!(message, log_type = "structured"),
            LogLevel::Debug => tracing::debug!(message, log_type = "structured"),
            LogLevel::Info => tracing::info!(message, log_type = "structured"),
            LogLevel::Warn => tracing::warn!(message, log_type = "structured"),
            LogLevel::Error => tracing::error!(message, log_type = "structured"),
        }
    }

    /// Log with correlation ID
    pub fn log_with_correlation(
        &self,
        level: LogLevel,
        message: impl Into<String>,
        correlation_id: &CorrelationId,
    ) {
        let message = message.into();
        let correlation_id = correlation_id.as_str();
        match level {
            LogLevel::Trace => tracing::trace!(message, correlation_id, log_type = "structured"),
            LogLevel::Debug => tracing::debug!(message, correlation_id, log_type = "structured"),
            LogLevel::Info => tracing::info!(message, correlation_id, log_type = "structured"),
            LogLevel::Warn => tracing::warn!(message, correlation_id, log_type = "structured"),
            LogLevel::Error => tracing::error!(message, correlation_id, log_type = "structured"),
        }
    }

    /// Log function execution with timing
    pub fn log_function_execution(
        &self,
        correlation_id: &CorrelationId,
        function_name: &str,
        execution_time_ms: u64,
        success: bool,
        cache_hit: bool,
    ) {
        tracing::info!(
            function_name,
            execution_time_ms,
            success,
            cache_hit,
            correlation_id = correlation_id.as_str(),
            log_type = "function_execution"
        );
    }

    /// Log error with context
    pub fn log_error(
        &self,
        correlation_id: &CorrelationId,
        error: &crate::errors::RuntimeError,
    ) {
        let mut context_info = Vec::new();

        if let Some(ref context) = error.context {
            if let Some(ref fn_name) = context.function_name {
                context_info.push(format!("function: {}", fn_name));
            }
            if let Some(ref fn_version) = context.function_version {
                context_info.push(format!("version: {}", fn_version));
            }
            if let Some(exec_time) = context.execution_time {
                context_info.push(format!("exec_time: {}ms", exec_time.as_millis()));
            }
        }

        let context_str = if context_info.is_empty() {
            String::new()
        } else {
            format!(" ({})", context_info.join(", "))
        };

        tracing::error!(
            correlation_id = correlation_id.as_str(),
            error_kind = format!("{:?}", error.kind),
            recoverable = error.recoverable,
            recovery_suggestions = format!("{:?}", error.recovery_suggestions),
            log_type = "error",
            "{}{}",
            error.message,
            context_str
        );
    }

    /// Log pool statistics
    pub fn log_pool_stats(
        &self,
        correlation_id: &CorrelationId,
        total_instances: usize,
        functions_in_pool: usize,
        pruned_count: usize,
    ) {
        tracing::info!(
            correlation_id = correlation_id.as_str(),
            log_type = "pool_stats",
            "Pool statistics: {} total instances, {} functions, {} pruned",
            total_instances,
            functions_in_pool,
            pruned_count
        );
    }

    /// Log cache statistics
    pub fn log_cache_stats(
        &self,
        correlation_id: &CorrelationId,
        cache_entries: usize,
        cache_hits: u64,
        cache_misses: u64,
    ) {
        let hit_rate = if cache_hits + cache_misses > 0 {
            (cache_hits as f64) / ((cache_hits + cache_misses) as f64) * 100.0
        } else {
            0.0
        };

        tracing::info!(
            correlation_id = correlation_id.as_str(),
            log_type = "cache_stats",
            "Cache statistics: {} entries, {} hits, {} misses, {:.2}% hit rate",
            cache_entries,
            cache_hits,
            cache_misses,
            hit_rate
        );
    }

    /// Log memory usage
    pub fn log_memory_usage(
        &self,
        correlation_id: &CorrelationId,
        memory_used_mb: f64,
        memory_limit_mb: u32,
    ) {
        let usage_percent = (memory_used_mb / memory_limit_mb as f64) * 100.0;

        tracing::info!(
            correlation_id = correlation_id.as_str(),
            log_type = "memory_usage",
            "Memory usage: {:.2}MB / {}MB ({:.1}%)",
            memory_used_mb,
            memory_limit_mb,
            usage_percent
        );
    }
}

impl Default for StructuredLogger {
    fn default() -> Self {
        Self::new()
    }
}

// Explicitly implement Send + Sync for StructuredLogger
// This is safe because Arc<RwLock<T>> is Send + Sync when T is Send + Sync
unsafe impl Send for StructuredLogger {}
unsafe impl Sync for StructuredLogger {}

/// Initialize structured logging with JSON output
pub fn init_structured_logging(verbose: bool) -> StructuredLogger {
    let logger = StructuredLogger::new();

    // Only initialize the global subscriber if it hasn't been set yet
    // This prevents issues in tests where multiple initializations occur
    static INIT: std::sync::Once = std::sync::Once::new();
    INIT.call_once(|| {
        let filter = if verbose {
            tracing_subscriber::EnvFilter::new("functionfly=debug")
        } else {
            tracing_subscriber::EnvFilter::new("functionfly=info")
        };

        // Create human-readable layer for console output with structured fields
        let console_layer = tracing_subscriber::fmt::layer()
            .with_target(false)
            .with_thread_ids(false)
            .with_thread_names(false)
            .with_file(false)
            .with_line_number(false);

        tracing_subscriber::registry()
            .with(filter)
            .with(console_layer)
            .init();
    });

    logger
}

/// Request context for carrying correlation ID through request lifecycle
#[derive(Clone)]
pub struct RequestContext {
    pub correlation_id: CorrelationId,
    pub start_time: std::time::Instant,
    pub function_name: Option<String>,
    pub function_version: Option<String>,
}

impl RequestContext {
    /// Create new request context
    pub fn new(correlation_id: CorrelationId) -> Self {
        Self {
            correlation_id,
            start_time: std::time::Instant::now(),
            function_name: None,
            function_version: None,
        }
    }

    /// Set function information
    pub fn with_function(mut self, name: impl Into<String>, version: impl Into<String>) -> Self {
        self.function_name = Some(name.into());
        self.function_version = Some(version.into());
        self
    }

    /// Get elapsed time since request start
    pub fn elapsed(&self) -> std::time::Duration {
        self.start_time.elapsed()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_correlation_id_generation() {
        let id1 = CorrelationId::generate();
        let id2 = CorrelationId::generate();

        assert_ne!(id1, id2);
        assert!(id1.as_str().starts_with("req_"));
        assert!(id2.as_str().starts_with("req_"));
    }

    #[test]
    fn test_correlation_id_display() {
        let id = CorrelationId::new("test_123".to_string());
        assert_eq!(id.as_str(), "test_123");
        assert_eq!(format!("{}", id), "test_123");
    }

    #[tokio::test]
    async fn test_logger_correlation_id_generation() {
        let logger = StructuredLogger::new();

        let id1 = logger.generate_correlation_id().await;
        let id2 = logger.generate_correlation_id().await;

        assert_ne!(id1, id2);
        assert!(id1.as_str().starts_with("ffly_"));
        assert!(id2.as_str().starts_with("ffly_"));
    }

    #[test]
    fn test_request_context() {
        let correlation_id = CorrelationId::generate();
        let context = RequestContext::new(correlation_id.clone())
            .with_function("test-function", "1.0.0");

        assert_eq!(context.correlation_id, correlation_id);
        assert_eq!(context.function_name, Some("test-function".to_string()));
        assert_eq!(context.function_version, Some("1.0.0".to_string()));

        // Test elapsed time
        std::thread::sleep(std::time::Duration::from_millis(10));
        let elapsed = context.elapsed();
        assert!(elapsed.as_millis() >= 10);
    }
}