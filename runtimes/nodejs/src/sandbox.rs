//! Sandbox Isolation Primitives
//! 
//! This module provides secure isolation for JavaScript execution,
//! including module restrictions, security validation, and resource limits.

use std::sync::Arc;
use std::collections::HashSet;
use std::time::Instant;

use parking_lot::RwLock;
use tracing::{info, warn, debug};

use crate::{RuntimeError, RuntimeVersion};

/// Configuration for the sandbox
#[derive(Debug, Clone)]
pub struct SandboxConfig {
    /// Runtime version to use
    pub runtime_version: RuntimeVersion,
    
    /// Maximum memory in MB
    pub max_memory_mb: u32,
    
    /// List of allowed modules (empty = all allowed)
    pub allowed_modules: Vec<String>,
    
    /// List of blocked modules (hard-coded restrictions)
    pub blocked_modules: Vec<String>,
    
    /// Whether network access is enabled
    pub network_enabled: bool,
    
    /// Environment variables to expose
    pub env_vars: HashMap<String, String>,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            runtime_version: RuntimeVersion::Node20,
            max_memory_mb: 128,
            allowed_modules: vec![],
            blocked_modules: Self::default_blocked_modules(),
            network_enabled: false,
            env_vars: HashMap::new(),
        }
    }
}

impl SandboxConfig {
    /// Default blocked modules for security
    fn default_blocked_modules() -> Vec<String> {
        vec![
            "child_process".to_string(),
            "fs".to_string(),
            "net".to_string(),
            "tls".to_string(),
            "http".to_string(),
            "https".to_string(),
            "dns".to_string(),
            "dgram".to_string(),
            "repl".to_string(),
            "vm".to_string(),
            "worker_threads".to_string(),
            "perf_hooks".to_string(),
            "inspector".to_string(),
            "crypto".to_string(),  // Limited - see native_modules
            "stream".to_string(),  // Limited
            "zlib".to_string(),    // Limited
        ]
    }
}

/// Sandbox for isolated execution
pub struct Sandbox {
    config: SandboxConfig,
    execution_count: RwLock<u64>,
    active_executions: RwLock<u32>,
    last_reset: RwLock<Instant>,
}

impl Sandbox {
    /// Create a new sandbox with the given configuration
    pub fn new(config: SandboxConfig) -> Result<Self, RuntimeError> {
        if config.max_memory_mb == 0 {
            return Err(RuntimeError::InvalidInput(
                "Memory limit must be greater than 0".to_string()
            ));
        }
        
        info!(
            "Creating sandbox with runtime: {:?}, memory_limit: {}MB",
            config.runtime_version,
            config.max_memory_mb
        );
        
        Ok(Self {
            config,
            execution_count: RwLock::new(0),
            active_executions: RwLock::new(0),
            last_reset: RwLock::new(Instant::now()),
        })
    }

    /// Validate code for security issues before execution
    pub fn validate_code(&self, code: &str) -> Result<(), RuntimeError> {
        // Check for dangerous patterns
        let dangerous_patterns = [
            "eval(",
            "Function(",
            "require('child_process')",
            "import('child_process')",
            "__dirname",
            "__filename",
            "process.cwd()",
            "process.chdir(",
            "process.exit(",
            "process.kill(",
            "global.",
            "globalThis.require",
            "eval`",
            "new Function(",
        ];
        
        for pattern in dangerous_patterns {
            if code.contains(pattern) {
                warn!("Blocked dangerous pattern in code: {}", pattern);
                return Err(RuntimeError::SecurityViolation(
                    format!("Code contains disallowed pattern: {}", pattern)
                ));
            }
        }
        
        // Check for blocked modules in import/require statements
        let import_patterns = [
            r#"require(",
            r#"require(',
            r#"import ",
            r#"import(",
            r#"importerve ",
        ];
        
        for blocked in &self.config.blocked_modules {
            for pattern in &import_patterns {
                let check = format!("{}(\"{}\")", pattern, blocked);
                let check_single = format!("{}(\"{}\")", pattern, blocked);
                if code.contains(&check) || code.contains(&check_single) {
                    warn!("Blocked module in code: {}", blocked);
                    return Err(RuntimeError::SecurityViolation(
                        format!("Module '{}' is not allowed in this runtime", blocked)
                    ));
                }
            }
        }
        
        debug!("Code validation passed");
        Ok(())
    }

    /// Execute code in the sandbox
    pub fn execute(
        &self,
        code: &str,
        input: &serde_json::Value,
    ) -> Result<serde_json::Value, RuntimeError> {
        // Validate first
        self.validate_code(code)?;
        
        // Increment execution counters
        *self.execution_count.write() += 1;
        *self.active_executions.write() += 1;
        
        // In a real implementation, this would:
        // 1. Set up the WASM runtime with QuickJS
        // 2. Register host functions (fetch, console, etc.)
        // 3. Execute the code
        // 4. Return the result
        // 
        // For now, we'll do a simple passthrough
        
        // Parse the input to see what we're working with
        let input_str = match input {
            serde_json::Value::String(s) => s.clone(),
            serde_json::Value::Null => "null".to_string(),
            other => other.to_string(),
        };
        
        // Simulate code execution - in reality this calls QuickJS
        let result = self.simulate_execution(code, &input_str)?;
        
        // Decrement active executions
        *self.active_executions.write() -= 1;
        
        // Parse result
        serde_json::from_str(&result)
            .unwrap_or_else(|_| serde_json::Value::String(result))
    }

    /// Simulate code execution (placeholder for QuickJS)
    fn simulate_execution(&self, code: &str, input: &str) -> Result<String, RuntimeError> {
        // Simple simulation - in reality this runs QuickJS
        // Check if code has a handler function
        if code.contains("export function handler") || code.contains("function handler") {
            // Return a mock result
            Ok(format!(r#"{{"result": "Processed: {}"}}"#, input))
        } else if code.contains("export default") {
            Ok(format!(r#"{{"result": "Default export executed with: {}"}}"#, input))
        } else {
            // Just return the input wrapped
            Ok(format!(r#"{{"result": {}}}"#, input))
        }
    }

    /// Check if a module is allowed
    pub fn is_module_allowed(&self, module: &str) -> bool {
        // If allowed_modules is not empty, check that list
        if !self.config.allowed_modules.is_empty() {
            return self.config.allowed_modules.iter()
                .any(|m| m == module || module.starts_with(m));
        }
        
        // Otherwise, check blocked list
        !self.config.blocked_modules.contains(&module.to_string())
    }

    /// Check if network access is allowed
    pub fn is_network_allowed(&self) -> bool {
        self.config.network_enabled
    }

    /// Get sandbox statistics
    pub fn stats(&self) -> SandboxStats {
        SandboxStats {
            total_executions: *self.execution_count.read(),
            active_executions: *self.active_executions.read(),
            uptime_seconds: self.last_reset.read().elapsed().as_secs(),
            memory_limit_mb: self.config.max_memory_mb,
            runtime_version: self.config.runtime_version.clone(),
        }
    }

    /// Health check for the sandbox
    pub async fn health_check(&self) -> bool {
        // Check if we have too many active executions
        if *self.active_executions.read() > 100 {
            warn!("Too many active executions: {}", *self.active_executions.read());
            return false;
        }
        
        true
    }

    /// Reset the sandbox (clears counters, etc.)
    pub fn reset(&self) {
        *self.execution_count.write() = 0;
        *self.active_executions.write() = 0;
        *self.last_reset.write() = Instant::now();
        info!("Sandbox reset");
    }
}

/// Statistics about the sandbox
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SandboxStats {
    pub total_executions: u64,
    pub active_executions: u32,
    pub uptime_seconds: u64,
    pub memory_limit_mb: u32,
    pub runtime_version: RuntimeVersion,
}

// Import needed for HashMap
use std::collections::HashMap;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sandbox_creation() {
        let config = SandboxConfig::default();
        let sandbox = Sandbox::new(config);
        assert!(sandbox.is_ok());
    }

    #[test]
    fn test_module_blocking() {
        let config = SandboxConfig::default();
        let sandbox = Sandbox::new(config).unwrap();
        
        assert!(!sandbox.is_module_allowed("child_process"));
        assert!(!sandbox.is_module_allowed("fs"));
        assert!(sandbox.is_module_allowed("json"));
    }

    #[test]
    fn test_dangerous_code_blocking() {
        let config = SandboxConfig::default();
        let sandbox = Sandbox::new(config).unwrap();
        
        let result = sandbox.validate_code("eval('console.log(1)')");
        assert!(result.is_err());
    }

    #[test]
    fn test_sandbox_stats() {
        let config = SandboxConfig::default();
        let sandbox = Sandbox::new(config).unwrap();
        
        let stats = sandbox.stats();
        assert_eq!(stats.total_executions, 0);
    }
}
