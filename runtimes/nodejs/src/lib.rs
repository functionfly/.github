//! FunctionFly Node.js Runtime Library
//! 
//! This module provides a high-performance JavaScript execution runtime
//! using QuickJS compiled to WebAssembly for secure, isolated function execution.

pub mod executor;
pub mod sandbox;
pub mod timeout;
pub mod memory;
pub mod native_modules;
pub mod host_functions;
pub mod config;
pub mod metrics;
pub mod wasm_entry;

use std::sync::Arc;
use std::future::Future;
use std::pin::Pin;
use std::time::Duration;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use uuid::Uuid;
use chrono::{DateTime, Utc};

pub use executor::NodeExecutor;
pub use sandbox::{Sandbox, SandboxConfig};
pub use timeout::TimeoutManager;
pub use memory::MemoryLimiter;
pub use config::{RuntimeConfig, RuntimeVersion};

// ============================================================================
// Error Types
// ============================================================================

#[derive(Error, Debug, Clone)]
pub enum RuntimeError {
    #[error("Compilation error: {0}")]
    Compilation(String),
    
    #[error("Execution error: {0}")]
    Execution(String),
    
    #[error("Timeout after {0}ms")]
    Timeout(u64),
    
    #[error("Memory limit exceeded: {0}")]
    MemoryLimit(String),
    
    #[error("Sandbox violation: {0}")]
    SecurityViolation(String),
    
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    
    #[error("Runtime not ready: {0}")]
    NotReady(String),
}

impl Serialize for RuntimeError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}

impl<'de> Deserialize<'de> for RuntimeError {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        // Parse the string to determine the error type
        if s.starts_with("Compilation error:") {
            Ok(RuntimeError::Compilation(s.trim_start_matches("Compilation error: ").to_string()))
        } else if s.starts_with("Execution error:") {
            Ok(RuntimeError::Execution(s.trim_start_matches("Execution error: ").to_string()))
        } else if s.starts_with("Timeout after") {
            let num = s.trim_start_matches("Timeout after ").trim_end_matches("ms");
            Ok(RuntimeError::Timeout(num.parse().unwrap_or(0)))
        } else if s.starts_with("Memory limit exceeded:") {
            Ok(RuntimeError::MemoryLimit(s.trim_start_matches("Memory limit exceeded: ").to_string()))
        } else if s.starts_with("Sandbox violation:") {
            Ok(RuntimeError::SecurityViolation(s.trim_start_matches("Sandbox violation: ").to_string()))
        } else if s.starts_with("Invalid input:") {
            Ok(RuntimeError::InvalidInput(s.trim_start_matches("Invalid input: ").to_string()))
        } else if s.starts_with("Runtime not ready:") {
            Ok(RuntimeError::NotReady(s.trim_start_matches("Runtime not ready: ").to_string()))
        } else {
            Ok(RuntimeError::Execution(s))
        }
    }
}

// ============================================================================
// Execution Types
// ============================================================================

/// Input for function execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionInput {
    pub data: serde_json::Value,
    pub metadata: ExecutionMetadata,
}

/// Metadata about the execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionMetadata {
    pub request_id: String,
    pub function_id: String,
    pub function_version: String,
    pub timestamp: DateTime<Utc>,
    pub headers: std::collections::HashMap<String, String>,
}

impl Default for ExecutionMetadata {
    fn default() -> Self {
        Self {
            request_id: Uuid::new_v4().to_string(),
            function_id: String::new(),
            function_version: "latest".to_string(),
            timestamp: Utc::now(),
            headers: std::collections::HashMap::new(),
        }
    }
}

/// Result of function execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionResult {
    pub request_id: String,
    pub success: bool,
    pub output: Option<serde_json::Value>,
    pub error: Option<RuntimeError>,
    pub execution_time_ms: u64,
    pub memory_used_bytes: u64,
    pub cold_start: bool,
    pub timestamp: DateTime<Utc>,
}

impl ExecutionResult {
    pub fn success(request_id: String, output: serde_json::Value, exec_time_ms: u64) -> Self {
        Self {
            request_id,
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms: exec_time_ms,
            memory_used_bytes: 0,
            cold_start: false,
            timestamp: Utc::now(),
        }
    }

    pub fn error(request_id: String, err: RuntimeError, exec_time_ms: u64) -> Self {
        Self {
            request_id,
            success: false,
            output: None,
            error: Some(err),
            execution_time_ms: exec_time_ms,
            memory_used_bytes: 0,
            cold_start: false,
            timestamp: Utc::now(),
        }
    }
}

// ============================================================================
// Runtime Trait
// ============================================================================

/// Trait for runtime executors
#[async_trait]
pub trait Runtime: Send + Sync {
    /// Execute a function with the given input
    async fn execute(
        &self,
        code: &str,
        input: ExecutionInput,
    ) -> ExecutionResult;

    /// Execute a function asynchronously (returns a Future)
    fn execute_async<'a>(
        &'a self,
        code: &'a str,
        input: ExecutionInput,
    ) -> Pin<Box<dyn Future<Output = ExecutionResult> + Send + 'a>>;

    /// Get runtime information
    fn info(&self) -> RuntimeInfo;

    /// Check if runtime is healthy
    async fn health_check(&self) -> bool;
}

/// Information about the runtime
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeInfo {
    pub name: String,
    pub version: String,
    pub supported_runtimes: Vec<RuntimeVersion>,
    pub max_memory_mb: u32,
    pub max_timeout_ms: u64,
    pub features: Vec<String>,
}

impl Default for RuntimeInfo {
    fn default() -> Self {
        Self {
            name: "functionfly-nodejs".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
            supported_runtimes: vec![
                RuntimeVersion::Node18,
                RuntimeVersion::Node20,
                RuntimeVersion::Deno,
            ],
            max_memory_mb: 512,
            max_timeout_ms: 30000,
            features: vec![
                "async_await".to_string(),
                "promises".to_string(),
                "fetch_api".to_string(),
                "streams".to_string(),
                "url_api".to_string(),
                "timing_api".to_string(),
                "structured_clone".to_string(),
            ],
        }
    }
}

// ============================================================================
// Convenience Functions
// ============================================================================

/// Create a new Node.js runtime with default configuration
pub fn create_runtime(config: RuntimeConfig) -> Result<NodeExecutor, RuntimeError> {
    NodeExecutor::new(config)
}

/// Create a new Node.js runtime with default settings
pub fn create_default_runtime() -> Result<NodeExecutor, RuntimeError> {
    NodeExecutor::new(RuntimeConfig::default())
}

/// Execute a simple function with string input/output
pub async fn execute_simple(
    code: &str,
    input: &str,
) -> Result<String, RuntimeError> {
    let runtime = create_default_runtime()?;
    
    let exec_input = ExecutionInput {
        data: serde_json::Value::String(input.to_string()),
        metadata: ExecutionMetadata::default(),
    };
    
    let result = runtime.execute(code, exec_input).await;
    
    if result.success {
        result
            .output
            .and_then(|v| v.as_str().map(String::from))
            .ok_or_else(|| RuntimeError::Execution("Invalid output format".to_string()))
    } else {
        Err(result.error.unwrap_or_else(|| {
            RuntimeError::Execution("Unknown error".to_string())
        }))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_runtime_info() {
        let runtime = create_default_runtime().unwrap();
        let info = runtime.info();
        
        assert_eq!(info.name, "functionfly-nodejs");
        assert!(!info.supported_runtimes.is_empty());
    }

    #[tokio::test]
    async fn test_simple_execution() {
        let code = r#"
            export function handler(input) {
                return "Hello, " + input + "!";
            }
        "#;
        
        let result = execute_simple(code, "World").await;
        assert!(result.is_ok());
    }
}
