//! FunctionFly Host Functions
//!
//! This module implements the FunctionFly-specific host functions that are
//! available to WebAssembly modules. These functions provide access to:
//!
//! | Function                                   | Capability      |
//! |--------------------------------------------|-----------------|
//! | `functionfly.log`                          | (always)        |
//! | `functionfly.get_env`                      | (always)        |
//! | `functionfly.fetch`                        | `fetch:read/write` |
//! | `functionfly.kv_get`, `kv_set`             | `kv`            |
//! | `functionfly.email`                        | `email`         |
//! | `functionfly.storage_read`, `storage_write`| `storage`       |
//! | `functionfly.ai`                           | `ai`            |
//! | `functionfly.external_api`                 | `external_api`  |
//! | `functionfly.webhook_send`                 | `webhook`       |
//! | `functionfly.get_secret`                   | `secret`        |
//! | `functionfly.queue_push`, `queue_pop`      | `queue`         |
//! | `functionfly.crypto_hmac`, `crypto_hash`, `crypto_random` | `crypto` |
//! | `functionfly.emit_metric`                  | `metric`        |

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
pub mod crypto;
pub mod email;
pub mod env;
pub mod external_api;
pub mod fetch;
pub mod kv;
pub mod logging;
pub mod memory_utils;
pub mod metric;
pub mod micropython;
pub mod queue;
pub mod secret;
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

    /// Add all FunctionFly host functions to the linker, gated by the
    /// capabilities declared in `config.capabilities`.
    pub fn add_to_linker(
        &self,
        linker: &mut wasmtime::Linker<WasiP1Ctx>,
    ) -> anyhow::Result<()> {
        let capabilities = Capabilities::from_string(&self.config.capabilities);

        // --- Always-available functions ---

        logging::add_log_function(self.logger.clone(), linker)?;
        env::add_get_env_function(self.config.clone(), linker)?;

        // MicroPython wrapper env stubs — always needed so Python WASM modules can instantiate.
        micropython::add_micropython_env_stubs(linker)?;

        // --- Capability-gated functions ---

        if capabilities.can_fetch() {
            fetch::add_fetch_function(linker, self.security_monitor.clone(), self.config.clone())?;
        }

        if capabilities.can_kv() {
            let kv_store = self
                .kv_store
                .as_ref()
                .context("KV store required for kv capability")?
                .clone();
            // Use tenant_id:function_key as namespace for cross-tenant isolation.
            // If tenant_id is not set (e.g. local dev), fall back to function_key only.
            let namespace = if let Some(ref tenant_id) = self.config.tenant_id {
                format!("{}:{}", tenant_id, self.config.function_key())
            } else {
                self.config.function_key()
            };
            kv::add_kv_functions_namespaced(kv_store, namespace, linker)?;
        }

        if capabilities.can_email() {
            email::add_email_function(self.config.clone(), linker)?;
        }

        if capabilities.can_storage() {
            storage::add_storage_functions(self.config.clone(), linker)?;
        }

        if capabilities.can_ai() {
            ai::add_ai_function(linker)?;
        }

        if capabilities.can_external_api() {
            external_api::add_external_api_function(self.config.clone(), linker)?;
        }

        // webhook is always registered (WASM modules may import it), but returns
        // -7 at runtime when the capability is not declared.
        webhook::add_webhook_function(linker, capabilities.can_webhook(), self.config.clone())?;

        if capabilities.can_secret() {
            secret::add_get_secret_function(self.config.clone(), linker)?;
        }

        if capabilities.can_queue() {
            let queue_store = queue::new_queue_store();
            let namespace = self.config.function_key();
            queue::add_queue_functions(
                queue_store,
                namespace,
                self.config.queue_max_len,
                self.config.queue_max_queues,
                linker,
            )?;
        }

        if capabilities.can_crypto() {
            crypto::add_crypto_functions(linker)?;
        }

        if capabilities.can_metric() {
            metric::add_emit_metric_function(self.config.clone(), linker)?;
        }

        tracing::debug!("Added FunctionFly host functions to WASM linker");
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_storage_path_existing_rejects_traversal() {
        let result = storage::validate_storage_path("../outside.txt", "/tmp");
        assert!(result.is_err());

        let result = storage::validate_storage_path("/etc/passwd", "/tmp");
        assert!(result.is_err());
    }

    #[test]
    fn test_validate_storage_path_non_existent_returns_full_path() {
        // Non-existent file falls back to string-level check (no canonicalize error).
        let result = storage::validate_storage_path("subdir/new_file.txt", "/tmp/storage");
        assert!(result.is_ok());
    }

    #[test]
    fn test_send_email_no_server() {
        // This test verifies that email sending attempts to connect to a server
        // and fails gracefully when no server is available
        let config = Config::default();
        let result = email::send_email(
            "test@example.com",
            "Test Subject",
            "Test Body",
            Some("sender@example.com"),
            &config,
        );
        // Should fail because there's no SMTP server running
        assert!(result.is_err());
    }

    #[test]
    fn test_run_ai_inference_sentiment() {
        let result = ai::run_ai_inference("sentiment", "This is great!");
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), "positive");

        let result = ai::run_ai_inference("unknown", "test");
        assert!(result.is_err());
    }

    #[test]
    fn test_crypto_hmac_sha256() {
        let result = crypto::compute_hmac("sha256", b"key", b"message");
        assert!(result.is_ok());
        assert_eq!(result.unwrap().len(), 64);
    }

    #[test]
    fn test_crypto_hash_sha256_known_vector() {
        let result = crypto::compute_hash("sha256", b"hello");
        assert!(result.is_ok());
        assert_eq!(
            result.unwrap(),
            "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
        );
    }

    #[test]
    fn test_crypto_random_bytes() {
        let bytes = crypto::read_random_bytes(32);
        assert!(bytes.is_ok());
        assert_eq!(bytes.unwrap().len(), 32);
    }

    #[test]
    fn test_secret_map_from_env() {
        std::env::set_var("SECRET_TEST_KEY_42", "test_value");
        let config = Config::default();
        let secrets = secret::build_secret_map(&config);
        assert_eq!(secrets.get("TEST_KEY_42").map(|s| s.as_str()), Some("test_value"));
        std::env::remove_var("SECRET_TEST_KEY_42");
    }
}
