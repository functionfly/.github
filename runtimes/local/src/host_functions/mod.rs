//! FunctionFly Host Functions
//!
//! This module implements the FunctionFly-specific host functions that are
//! available to WebAssembly modules. These functions provide access to:
//! - Logging (functionfly.log)
//! - HTTP requests (functionfly.fetch)
//! - Key-value storage (functionfly.kv_get, functionfly.kv_set)
//! - Environment variables (functionfly.get_env)
//! - Email sending (functionfly.email)
//! - File storage (functionfly.storage_read, functionfly.storage_write)
//! - AI/ML inference (functionfly.ai)
//! - External API calls (functionfly.external_api)
//! - Webhook sending (functionfly.webhook_send)
//!
//! These functions are imported by WASM modules under the "functionfly" namespace
//! and provide a standardized interface for FunctionFly capabilities.

use anyhow::Context;
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::capability::Capabilities;
use crate::config::Config;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::security::SecurityMonitor;
use std::sync::Arc;

// Module declarations
pub mod ai;
pub mod email;
pub mod env;
pub mod external_api;
pub mod fetch;
pub mod kv;
pub mod logging;
pub mod memory_utils;
pub mod micropython;
pub mod storage;
pub mod webhook;

/// FunctionFly host functions linker
pub struct HostFunctionsLinker {
    kv_store: Option<SharedKVStore>,
    logger: StructuredLogger,
    config: Config,
    security_monitor: Arc<SecurityMonitor>,
}

impl HostFunctionsLinker {
    /// Create a new host functions linker
    pub fn new(
        kv_store: Option<SharedKVStore>,
        logger: StructuredLogger,
        config: Config,
        security_monitor: Arc<SecurityMonitor>,
    ) -> Self {
        Self {
            kv_store,
            logger,
            config,
            security_monitor,
        }
    }

    /// Add all FunctionFly host functions to the linker
    pub fn add_to_linker(
        &self,
        linker: &mut wasmtime::Linker<WasiP1Ctx>,
    ) -> anyhow::Result<()> {
        let capabilities = Capabilities::from_string(&self.config.capabilities);

        // Always add logging function (basic capability)
        logging::add_log_function(self.logger.clone(), linker)?;

        // Add fetch function if fetch capability is declared.
        // Pass the full config so the whitelist / strict-whitelist settings are
        // enforced at the host-function level.
        if capabilities.can_fetch() {
            fetch::add_fetch_function(linker, self.security_monitor.clone(), self.config.clone())?;
        }

        // Add KV functions if KV capability is declared
        if capabilities.can_kv() {
            let kv_store = self.kv_store.as_ref()
                .context("KV store required for KV capability")?
                .clone();
            kv::add_kv_functions(kv_store, linker)?;
        }

        // Add email function if email capability is declared
        if capabilities.can_email() {
            email::add_email_function(self.config.clone(), linker)?;
        }

        // Add storage functions if storage capability is declared
        if capabilities.can_storage() {
            storage::add_storage_functions(self.config.clone(), linker)?;
        }

        // Add AI function if AI capability is declared
        if capabilities.can_ai() {
            ai::add_ai_function(linker)?;
        }

        // Add external API function if external_api capability is declared
        if capabilities.can_external_api() {
            external_api::add_external_api_function(self.config.clone(), linker)?;
        }

        // Always add webhook function (WASM modules may import it, but it will return
        // an error at runtime if the webhook capability is not declared)
        webhook::add_webhook_function(linker, capabilities.can_webhook())?;

        // Add environment function (always available)
        env::add_get_env_function(self.config.clone(), linker)?;

        // Add MicroPython wrapper env stubs so Python WASM modules can instantiate
        micropython::add_micropython_env_stubs(linker)?;

        tracing::debug!("Added FunctionFly host functions to WASM linker");
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_storage_path() {
        // Valid path
        let result = storage::validate_storage_path("test.txt", "/tmp/storage");
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), std::path::PathBuf::from("/tmp/storage/test.txt"));

        // Path with directory traversal should fail
        let result = storage::validate_storage_path("../outside.txt", "/tmp/storage");
        assert!(result.is_err());

        // Absolute path should fail
        let result = storage::validate_storage_path("/etc/passwd", "/tmp/storage");
        assert!(result.is_err());
    }

    #[test]
    fn test_send_email() {
        let config = Config::default();
        let result = email::send_email(
            "test@example.com",
            "Test Subject",
            "Test Body",
            Some("sender@example.com"),
            &config,
        );
        // Should succeed (currently just logs)
        assert!(result.is_ok());
    }

    #[test]
    fn test_run_ai_inference() {
        // Test sentiment analysis
        let result = ai::run_ai_inference("sentiment", "This is great!");
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), "positive");

        // Test unknown model
        let result = ai::run_ai_inference("unknown", "test");
        assert!(result.is_err());
    }

    #[test]
    fn test_make_external_api_request() {
        let config = Config::default();
        let headers = std::collections::HashMap::new();
        let result = external_api::make_external_api_request(
            "GET",
            "https://api.example.com/test",
            &headers,
            None,
            &config,
        );
        // Should succeed (currently returns simulated response)
        assert!(result.is_ok());
    }
}
