//! Comprehensive error types and recovery mechanisms for the local runtime.

use std::fmt;
use std::time::Duration;

/// Main error type for the local runtime
#[derive(Debug, Clone)]
pub struct RuntimeError {
    /// Error kind for categorization
    pub kind: ErrorKind,
    /// Human-readable error message
    pub message: String,
    /// Correlation ID for tracing
    pub correlation_id: Option<String>,
    /// Recovery suggestions
    pub recovery_suggestions: Vec<String>,
    /// Whether this error is recoverable
    pub recoverable: bool,
    /// Context information
    pub context: Option<ErrorContext>,
}

/// Error categorization for better handling
#[allow(dead_code)]
#[derive(Debug, Clone, PartialEq)]
pub enum ErrorKind {
    // Execution errors
    WasmCompilation,
    WasmInstantiation,
    WasmExecution,
    FunctionNotFound,
    TimeoutExceeded,

    // Resource errors
    MemoryLimitExceeded,
    FuelLimitExceeded,
    InstancePoolExhausted,
    RateLimitExceeded,

    // Configuration errors
    InvalidConfig,
    WasmFileNotFound,
    WasmFileCorrupt,

    // WASI errors
    WasiNotSupported,
    WasiSyscallFailed,

    // Python runtime errors
    PythonEngineUnavailable,
    PythonExecutionFailed,
    PythonModuleNotFound,

    // Pool management errors
    PoolPruningFailed,
    PoolCapacityExceeded,

    // Cache errors
    CacheCorruption,
    CacheSizeExceeded,

    // Security errors
    SecurityViolation,
    ResourceLimitExceeded,

    // Network/IO errors
    ConnectionFailed,
    IoError,

    // Unknown/other errors
    Unknown,
}

/// Additional context for errors
#[derive(Debug, Clone, Default)]
#[allow(dead_code)]
pub struct ErrorContext {
    pub function_name: Option<String>,
    pub function_version: Option<String>,
    pub input_size: Option<usize>,
    pub execution_time: Option<Duration>,
    pub memory_used: Option<usize>,
    pub instance_id: Option<String>,
    pub request_id: Option<String>,
}

#[allow(dead_code)]
impl RuntimeError {
    /// Create a new runtime error
    pub fn new(kind: ErrorKind, message: impl Into<String>) -> Self {
        let recoverable = Self::is_recoverable(&kind);
        let recovery_suggestions = Self::get_recovery_suggestions(&kind);

        Self {
            kind,
            message: message.into(),
            correlation_id: None,
            recovery_suggestions,
            recoverable,
            context: None,
        }
    }

    /// Create an error with correlation ID
    pub fn with_correlation_id(mut self, correlation_id: impl Into<String>) -> Self {
        self.correlation_id = Some(correlation_id.into());
        self
    }

    /// Create an error with context
    pub fn with_context(mut self, context: ErrorContext) -> Self {
        self.context = Some(context);
        self
    }

    /// Add recovery suggestion
    pub fn with_recovery_suggestion(mut self, suggestion: impl Into<String>) -> Self {
        self.recovery_suggestions.push(suggestion.into());
        self
    }

    /// Check if an error kind is recoverable
    fn is_recoverable(kind: &ErrorKind) -> bool {
        matches!(
            kind,
            ErrorKind::TimeoutExceeded
                | ErrorKind::MemoryLimitExceeded
                | ErrorKind::FuelLimitExceeded
                | ErrorKind::InstancePoolExhausted
                | ErrorKind::PoolCapacityExceeded
                | ErrorKind::ConnectionFailed
                | ErrorKind::IoError
        )
    }

    /// Get recovery suggestions for error kinds
    fn get_recovery_suggestions(kind: &ErrorKind) -> Vec<String> {
        match kind {
            ErrorKind::TimeoutExceeded => vec![
                "Increase timeout limit with --timeout-ms flag".to_string(),
                "Optimize function performance".to_string(),
                "Check for infinite loops in code".to_string(),
            ],
            ErrorKind::MemoryLimitExceeded => vec![
                "Increase memory limit with --memory-mb flag".to_string(),
                "Optimize memory usage in function".to_string(),
                "Check for memory leaks".to_string(),
            ],
            ErrorKind::FuelLimitExceeded => vec![
                "Increase fuel limit in engine configuration".to_string(),
                "Optimize function execution efficiency".to_string(),
            ],
            ErrorKind::InstancePoolExhausted => vec![
                "Increase pool size or reduce concurrent requests".to_string(),
                "Check for instance leaks".to_string(),
            ],
            ErrorKind::WasmCompilation => vec![
                "Verify WASM module is valid".to_string(),
                "Check WASM target architecture".to_string(),
                "Recompile WASM module".to_string(),
            ],
            ErrorKind::WasmInstantiation => vec![
                "Check WASM module imports".to_string(),
                "Verify WASI configuration".to_string(),
                "Check memory requirements".to_string(),
            ],
            ErrorKind::PythonEngineUnavailable => vec![
                "Ensure Python runtime is properly configured".to_string(),
                "Check Python engine dependencies".to_string(),
                "Verify Python WASM module compatibility".to_string(),
            ],
            ErrorKind::PoolPruningFailed => vec![
                "Check instance pool configuration".to_string(),
                "Monitor pool statistics".to_string(),
                "Restart runtime if necessary".to_string(),
            ],
            _ => vec!["Check logs for more details".to_string()],
        }
    }

    /// Create a timeout error
    pub fn timeout(timeout_ms: u64) -> Self {
        Self::new(
            ErrorKind::TimeoutExceeded,
            format!("Execution timeout exceeded {}ms", timeout_ms),
        )
    }

    /// Create a WASM compilation error
    pub fn wasm_compilation(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::WasmCompilation, message)
    }

    /// Create a WASM execution error
    pub fn wasm_execution(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::WasmExecution, message)
    }

    /// Create a memory limit error
    pub fn memory_limit(limit_mb: u32) -> Self {
        Self::new(
            ErrorKind::MemoryLimitExceeded,
            format!("Memory limit of {}MB exceeded", limit_mb),
        )
    }

    /// Create an instance pool exhausted error
    pub fn pool_exhausted() -> Self {
        Self::new(
            ErrorKind::InstancePoolExhausted,
            "Instance pool capacity exceeded",
        )
    }

    /// Create a file not found error
    pub fn file_not_found(path: impl Into<String>) -> Self {
        Self::new(
            ErrorKind::WasmFileNotFound,
            format!("WASM file not found: {}", path.into()),
        )
    }

    /// Create a Python execution error
    pub fn python_execution(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::PythonExecutionFailed, message)
    }

    /// Create a configuration error
    pub fn config_error(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::InvalidConfig, message)
    }

    /// Create a security violation error
    pub fn security_violation(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::SecurityViolation, message)
    }

    /// Create a resource limit exceeded error
    pub fn resource_limit(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::ResourceLimitExceeded, message)
    }

    /// Create a WASM instantiation error
    pub fn wasm_instantiation(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::WasmInstantiation, message)
    }

    /// Create a function not found error
    pub fn function_not_found(function_name: &str) -> Self {
        Self::new(
            ErrorKind::FunctionNotFound,
            format!("Function '{}' not found in WASM module", function_name),
        )
    }

    /// Create a rate limit exceeded error
    pub fn rate_limit_exceeded() -> Self {
        Self::new(
            ErrorKind::RateLimitExceeded,
            "Rate limit exceeded",
        )
    }

    /// Create a cache corruption error
    pub fn cache_corruption(description: impl Into<String>) -> Self {
        Self::new(ErrorKind::CacheCorruption, description)
    }

    /// Create a pool capacity exceeded error
    pub fn pool_capacity_exceeded() -> Self {
        Self::new(
            ErrorKind::PoolCapacityExceeded,
            "Pool capacity exceeded",
        )
    }

    /// Create a WASI not supported error
    pub fn wasi_not_supported(feature: impl Into<String>) -> Self {
        Self::new(ErrorKind::WasiNotSupported, feature)
    }

    /// Create a WASI syscall failed error
    pub fn wasi_syscall_failed(syscall: impl Into<String>, message: impl Into<String>) -> Self {
        Self::new(
            ErrorKind::WasiSyscallFailed,
            format!("WASI syscall '{}' failed: {}", syscall.into(), message.into()),
        )
    }

    /// Create a Python engine unavailable error
    pub fn python_engine_unavailable(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::PythonEngineUnavailable, message)
    }

    /// Create a Python module not found error
    pub fn python_module_not_found(module_name: impl Into<String>) -> Self {
        Self::new(
            ErrorKind::PythonModuleNotFound,
            format!("Python module '{}' not found", module_name.into()),
        )
    }

    /// Create a pool pruning failed error
    pub fn pool_pruning_failed(message: impl Into<String>) -> Self {
        Self::new(ErrorKind::PoolPruningFailed, message)
    }

    /// Create a cache size exceeded error
    pub fn cache_size_exceeded(max_size: usize) -> Self {
        Self::new(
            ErrorKind::CacheSizeExceeded,
            format!("Cache size exceeded maximum of {} bytes", max_size),
        )
    }

    /// Create a fuel limit exceeded error
    pub fn fuel_limit_exceeded() -> Self {
        Self::new(
            ErrorKind::FuelLimitExceeded,
            "CPU fuel limit exceeded",
        )
    }

    /// Create a wasm file corrupt error
    pub fn wasm_file_corrupt(path: impl Into<String>) -> Self {
        Self::new(
            ErrorKind::WasmFileCorrupt,
            format!("WASM file is corrupt: {}", path.into()),
        )
    }
}

impl fmt::Display for RuntimeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "[{:?}] {}", self.kind, self.message)?;

        if let Some(ref id) = self.correlation_id {
            write!(f, " (correlation_id: {})", id)?;
        }

        Ok(())
    }
}

impl std::error::Error for RuntimeError {}

/// Result type alias for runtime operations
pub type RuntimeResult<T> = Result<T, RuntimeError>;

/// Recovery strategy for handling errors
#[derive(Debug, Clone)]
pub enum RecoveryStrategy {
    /// Retry the operation after a delay
    Retry { attempts: u32, delay_ms: u64 },
    /// Fallback to alternative implementation
    Fallback { alternative: String },
    /// Degrade gracefully (reduced functionality)
    Degrade { message: String },
    /// Fail fast (immediate termination)
    FailFast,
    /// Custom recovery logic
    Custom { description: String },
}

impl RecoveryStrategy {
    /// Create a retry strategy
    pub fn retry(attempts: u32, delay_ms: u64) -> Self {
        Self::Retry { attempts, delay_ms }
    }

    /// Create a fallback strategy
    pub fn fallback(alternative: impl Into<String>) -> Self {
        Self::Fallback {
            alternative: alternative.into(),
        }
    }

    /// Create a degrade strategy
    pub fn degrade(message: impl Into<String>) -> Self {
        Self::Degrade {
            message: message.into(),
        }
    }
}

/// Error recovery manager
pub struct ErrorRecovery {
    max_retry_attempts: u32,
    base_retry_delay_ms: u64,
}

impl Default for ErrorRecovery {
    fn default() -> Self {
        Self {
            max_retry_attempts: 3,
            base_retry_delay_ms: 100,
        }
    }
}

impl ErrorRecovery {
    /// Create new error recovery manager
    pub fn new() -> Self {
        Self::default()
    }

    /// Determine recovery strategy for an error
    pub fn get_recovery_strategy(&self, error: &RuntimeError) -> RecoveryStrategy {
        match error.kind {
            ErrorKind::TimeoutExceeded => {
                RecoveryStrategy::retry(2, 500)
            }
            ErrorKind::ConnectionFailed | ErrorKind::IoError => {
                RecoveryStrategy::retry(3, 200)
            }
            ErrorKind::InstancePoolExhausted => {
                RecoveryStrategy::degrade("Using direct execution without pooling".to_string())
            }
            ErrorKind::MemoryLimitExceeded => {
                Self::failover("Switching to memory-efficient execution mode")
            }
            ErrorKind::PythonEngineUnavailable => {
                RecoveryStrategy::fallback("wasm".to_string())
            }
            _ if error.recoverable => {
                RecoveryStrategy::retry(self.max_retry_attempts, self.base_retry_delay_ms)
            }
            _ => RecoveryStrategy::FailFast,
        }
    }

    /// Execute recovery strategy
    pub async fn execute_recovery<F, Fut, T>(
        &self,
        strategy: &RecoveryStrategy,
        operation: F,
    ) -> RuntimeResult<T>
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = RuntimeResult<T>>,
    {
        match strategy {
            RecoveryStrategy::Retry { attempts, delay_ms } => {
                self.execute_retry(*attempts, *delay_ms, operation).await
            }
            RecoveryStrategy::Fallback { alternative } => {
                Err(RuntimeError::new(
                    ErrorKind::Unknown,
                    format!("Fallback to {} required", alternative),
                ))
            }
            RecoveryStrategy::Degrade { message } => {
                Err(RuntimeError::new(
                    ErrorKind::Unknown,
                    format!("Degraded mode: {}", message),
                ))
            }
            RecoveryStrategy::FailFast => {
                Err(RuntimeError::new(
                    ErrorKind::Unknown,
                    "Operation failed and cannot be recovered",
                ))
            }
            RecoveryStrategy::Custom { description } => {
                Err(RuntimeError::new(
                    ErrorKind::Unknown,
                    format!("Custom recovery required: {}", description),
                ))
            }
        }
    }

    /// Execute retry strategy
    async fn execute_retry<F, Fut, T>(
        &self,
        attempts: u32,
        delay_ms: u64,
        operation: F,
    ) -> RuntimeResult<T>
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = RuntimeResult<T>>,
    {
        let mut last_error = None;

        for attempt in 1..=attempts {
            match operation().await {
                Ok(result) => return Ok(result),
                Err(error) => {
                    last_error = Some(error);
                    if attempt < attempts {
                        tracing::warn!(
                            "Operation failed (attempt {}/{}), retrying in {}ms",
                            attempt,
                            attempts,
                            delay_ms
                        );
                        tokio::time::sleep(Duration::from_millis(delay_ms)).await;
                    }
                }
            }
        }

        Err(last_error.unwrap_or_else(|| {
            RuntimeError::new(ErrorKind::Unknown, "All retry attempts failed")
        }))
    }

    /// Create a failover strategy (helper method)
    fn failover(description: impl Into<String>) -> RecoveryStrategy {
        RecoveryStrategy::Custom {
            description: description.into(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_error_creation() {
        let error = RuntimeError::new(ErrorKind::WasmCompilation, "Test error");
        assert_eq!(error.kind, ErrorKind::WasmCompilation);
        assert_eq!(error.message, "Test error");
        assert!(!error.recoverable); // WasmCompilation errors are not recoverable
    }

    #[test]
    fn test_timeout_error() {
        let error = RuntimeError::timeout(5000);
        assert_eq!(error.kind, ErrorKind::TimeoutExceeded);
        assert!(error.message.contains("5000"));
    }

    #[test]
    fn test_recovery_strategies() {
        let recovery = ErrorRecovery::new();

        // Test timeout error recovery
        let timeout_error = RuntimeError::timeout(1000);
        let strategy = recovery.get_recovery_strategy(&timeout_error);
        match strategy {
            RecoveryStrategy::Retry { attempts, delay_ms } => {
                assert_eq!(attempts, 2);
                assert_eq!(delay_ms, 500);
            }
            _ => panic!("Expected retry strategy"),
        }

        // Test non-recoverable error
        let config_error = RuntimeError::config_error("Invalid config");
        let strategy = recovery.get_recovery_strategy(&config_error);
        assert!(matches!(strategy, RecoveryStrategy::FailFast));
    }

    #[test]
    fn test_error_kind_recoverability() {
        // Test that recoverable errors are correctly identified
        let recoverable_kinds = vec![
            ErrorKind::TimeoutExceeded,
            ErrorKind::MemoryLimitExceeded,
            ErrorKind::FuelLimitExceeded,
            ErrorKind::InstancePoolExhausted,
            ErrorKind::PoolCapacityExceeded,
            ErrorKind::ConnectionFailed,
            ErrorKind::IoError,
        ];
        
        for kind in recoverable_kinds {
            let error = RuntimeError::new(kind.clone(), "test");
            assert!(error.recoverable, "Expected {:?} to be recoverable", kind);
        }
        
        // Test non-recoverable errors
        let non_recoverable_kinds = vec![
            ErrorKind::WasmCompilation,
            ErrorKind::WasmInstantiation,
            ErrorKind::WasmExecution,
            ErrorKind::FunctionNotFound,
            ErrorKind::InvalidConfig,
            ErrorKind::SecurityViolation,
        ];
        
        for kind in non_recoverable_kinds {
            let error = RuntimeError::new(kind.clone(), "test");
            assert!(!error.recoverable, "Expected {:?} to be non-recoverable", kind);
        }
    }

    #[test]
    fn test_error_factory_methods() {
        // Test all error factory methods
        let errors = vec![
            RuntimeError::timeout(1000),
            RuntimeError::wasm_compilation("compilation failed"),
            RuntimeError::wasm_execution("execution failed"),
            RuntimeError::memory_limit(128),
            RuntimeError::pool_exhausted(),
            RuntimeError::file_not_found("/path/to/file.wasm"),
            RuntimeError::python_execution("python error"),
            RuntimeError::config_error("invalid config"),
            RuntimeError::security_violation("security issue"),
            RuntimeError::resource_limit("resource exceeded"),
            RuntimeError::wasm_instantiation("instantiation failed"),
            RuntimeError::function_not_found("handler"),
            RuntimeError::rate_limit_exceeded(),
            RuntimeError::cache_corruption("cache corrupted"),
            RuntimeError::pool_capacity_exceeded(),
            RuntimeError::wasi_not_supported("wasi feature"),
            RuntimeError::wasi_syscall_failed("syscall", "failed"),
            RuntimeError::python_engine_unavailable("engine down"),
            RuntimeError::python_module_not_found("module"),
            RuntimeError::pool_pruning_failed("pruning error"),
            RuntimeError::cache_size_exceeded(1000),
            RuntimeError::fuel_limit_exceeded(),
            RuntimeError::wasm_file_corrupt("/path/to/file.wasm"),
        ];
        
        for error in errors {
            assert!(!error.message.is_empty());
            assert!(!error.recovery_suggestions.is_empty() || error.recoverable);
        }
    }

    #[test]
    fn test_error_with_correlation_id() {
        let error = RuntimeError::new(ErrorKind::TimeoutExceeded, "test");
        let error_with_id = error.with_correlation_id("corr-123");
        
        assert_eq!(error_with_id.correlation_id, Some("corr-123".to_string()));
    }

    #[test]
    fn test_error_with_context() {
        let error = RuntimeError::new(ErrorKind::TimeoutExceeded, "test");
        let context = ErrorContext {
            function_name: Some("test-fn".to_string()),
            function_version: Some("1.0.0".to_string()),
            input_size: Some(1024),
            execution_time: Some(Duration::from_millis(5000)),
            memory_used: Some(128 * 1024 * 1024),
            instance_id: Some("inst-123".to_string()),
            request_id: Some("req-456".to_string()),
        };
        
        let error_with_ctx = error.with_context(context);
        assert!(error_with_ctx.context.is_some());
        let ctx = error_with_ctx.context.unwrap();
        assert_eq!(ctx.function_name, Some("test-fn".to_string()));
        assert_eq!(ctx.function_version, Some("1.0.0".to_string()));
    }

    #[test]
    fn test_error_with_recovery_suggestion() {
        let error = RuntimeError::new(ErrorKind::TimeoutExceeded, "test");
        let error_with_suggestion = error.with_recovery_suggestion("Try increasing the timeout");
        
        assert!(error_with_suggestion.recovery_suggestions.contains(&"Try increasing the timeout".to_string()));
    }

    #[test]
    fn test_error_display() {
        let error = RuntimeError::new(ErrorKind::TimeoutExceeded, "timeout exceeded");
        let display = format!("{}", error);
        
        assert!(display.contains("TimeoutExceeded"));
        assert!(display.contains("timeout exceeded"));
    }

    #[test]
    fn test_error_display_with_correlation_id() {
        let error = RuntimeError::new(ErrorKind::WasmCompilation, "compilation failed")
            .with_correlation_id("corr-789");
        
        let display = format!("{}", error);
        assert!(display.contains("corr-789"));
    }

    #[test]
    fn test_recovery_strategy_factories() {
        // Test Retry factory
        let retry = RecoveryStrategy::retry(3, 200);
        assert!(matches!(retry, RecoveryStrategy::Retry { attempts: 3, delay_ms: 200 }));
        
        // Test Fallback factory
        let fallback = RecoveryStrategy::fallback("alternative");
        assert!(matches!(fallback, RecoveryStrategy::Fallback { alternative } if alternative == "alternative"));
        
        // Test Degrade factory
        let degrade = RecoveryStrategy::degrade("degraded mode");
        assert!(matches!(degrade, RecoveryStrategy::Degrade { message } if message == "degraded mode"));
    }

    #[test]
    fn test_error_recovery_default() {
        let recovery = ErrorRecovery::default();
        assert_eq!(recovery.max_retry_attempts, 3);
        assert_eq!(recovery.base_retry_delay_ms, 100);
    }

    #[test]
    fn test_error_recovery_panic_without_attempts() {
        let recovery = ErrorRecovery::new();
        
        // Create a non-recoverable error
        let error = RuntimeError::config_error("invalid");
        let strategy = recovery.get_recovery_strategy(&error);
        
        // Should return FailFast for non-recoverable errors
        assert!(matches!(strategy, RecoveryStrategy::FailFast));
    }
}
