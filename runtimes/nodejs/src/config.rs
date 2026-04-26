//! Runtime Configuration
//!
//! This module provides configuration types for the Node.js runtime.

use std::collections::HashMap;

use serde::{Deserialize, Serialize};

#[cfg(feature = "wasmtime")]
use crate::engine::WasmEngineConfig;

/// Runtime version
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum RuntimeVersion {
    /// Node.js 18.x
    Node18,
    /// Node.js 20.x
    Node20,
    /// Deno
    Deno,
}

#[cfg(feature = "wasmtime")]
impl From<&RuntimeConfig> for WasmEngineConfig {
    fn from(runtime_config: &RuntimeConfig) -> Self {
        WasmEngineConfig {
            max_memory_mb: runtime_config.max_memory_mb,
            max_timeout_ms: runtime_config.max_timeout_ms,
            max_wasm_stack: runtime_config.max_wasm_stack,
            fuel_per_ms: runtime_config.fuel_per_ms,
            consume_fuel: true,
            epoch_interruption: true,
        }
    }
}

impl Default for RuntimeVersion {
    fn default() -> Self {
        Self::Node20
    }
}

impl std::fmt::Display for RuntimeVersion {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Node18 => write!(f, "nodejs18"),
            Self::Node20 => write!(f, "nodejs20"),
            Self::Deno => write!(f, "deno"),
        }
    }
}

/// Configuration for the Node.js runtime
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeConfig {
    /// Runtime version to use
    pub version: RuntimeVersion,

    /// Maximum memory in MB (default: 128)
    pub max_memory_mb: u32,

    /// Maximum execution timeout in milliseconds (default: 30000)
    pub max_timeout_ms: u64,

    /// Whether to enable code caching
    pub enable_cache: bool,

    /// Maximum cached functions
    pub max_cached_functions: usize,

    /// List of allowed npm packages (empty = allow all bundled)
    pub allowed_modules: Vec<String>,

    /// List of blocked npm packages (explicitly forbidden)
    pub blocked_modules: Vec<String>,

    /// Whether network access is enabled
    pub network_enabled: bool,

    /// Environment variables to expose to functions
    pub environment: HashMap<String, String>,

    /// Whether to enable verbose logging
    pub verbose_logging: bool,

    /// Custom handler function name (default: "handler")
    pub handler_name: String,

    #[cfg(feature = "wasmtime")]
    /// Maximum WASM stack size in bytes (default: 512KB)
    pub max_wasm_stack: usize,

    #[cfg(feature = "wasmtime")]
    /// Fuel per millisecond for CPU limit enforcement
    pub fuel_per_ms: u64,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            version: RuntimeVersion::Node20,
            max_memory_mb: 128,
            max_timeout_ms: 30000,
            enable_cache: true,
            max_cached_functions: 100,
            allowed_modules: vec![],
            blocked_modules: vec![],
            network_enabled: false,
            environment: HashMap::new(),
            verbose_logging: false,
            handler_name: "handler".to_string(),
            #[cfg(feature = "wasmtime")]
            max_wasm_stack: 512 * 1024,
            #[cfg(feature = "wasmtime")]
            fuel_per_ms: 10_000,
        }
    }
}

impl RuntimeConfig {
    /// Validate the configuration
    pub fn validate(&self) -> Result<(), crate::RuntimeError> {
        if self.max_memory_mb == 0 {
            return Err(crate::RuntimeError::InvalidInput(
                "max_memory_mb must be greater than 0".to_string(),
            ));
        }

        if self.max_memory_mb > 2048 {
            return Err(crate::RuntimeError::InvalidInput(
                "max_memory_mb cannot exceed 2048MB".to_string(),
            ));
        }

        if self.max_timeout_ms == 0 {
            return Err(crate::RuntimeError::InvalidInput(
                "max_timeout_ms must be greater than 0".to_string(),
            ));
        }

        if self.max_timeout_ms > 300000 {
            return Err(crate::RuntimeError::InvalidInput(
                "max_timeout_ms cannot exceed 300000ms (5 minutes)".to_string(),
            ));
        }

        if self.max_cached_functions == 0 {
            return Err(crate::RuntimeError::InvalidInput(
                "max_cached_functions must be greater than 0".to_string(),
            ));
        }

        Ok(())
    }

    /// Create a config for development
    pub fn development() -> Self {
        Self {
            version: RuntimeVersion::Node20,
            max_memory_mb: 256,
            max_timeout_ms: 60000,
            enable_cache: true,
            max_cached_functions: 50,
            allowed_modules: vec![],
            blocked_modules: vec![],
            network_enabled: true,
            environment: HashMap::from([("NODE_ENV".to_string(), "development".to_string())]),
            verbose_logging: true,
            handler_name: "handler".to_string(),
            #[cfg(feature = "wasmtime")]
            max_wasm_stack: 512 * 1024,
            #[cfg(feature = "wasmtime")]
            fuel_per_ms: 10_000,
        }
    }

    /// Create a config for production
    pub fn production() -> Self {
        Self {
            version: RuntimeVersion::Node20,
            max_memory_mb: 128,
            max_timeout_ms: 30000,
            enable_cache: true,
            max_cached_functions: 100,
            allowed_modules: vec![],
            blocked_modules: vec![],
            network_enabled: false,
            environment: HashMap::from([("NODE_ENV".to_string(), "production".to_string())]),
            verbose_logging: false,
            handler_name: "handler".to_string(),
            #[cfg(feature = "wasmtime")]
            max_wasm_stack: 512 * 1024,
            #[cfg(feature = "wasmtime")]
            fuel_per_ms: 10_000,
        }
    }

    /// Create a config with custom settings
    pub fn custom(version: RuntimeVersion, memory_mb: u32, timeout_ms: u64) -> Self {
        Self {
            version,
            max_memory_mb: memory_mb,
            max_timeout_ms: timeout_ms,
            ..Default::default()
        }
    }
}
