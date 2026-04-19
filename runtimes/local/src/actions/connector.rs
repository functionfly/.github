//! Action connector trait and shared infrastructure.
//!
//! Provides the `ActionConnector` trait that all external service connectors
//! implement. Includes idempotency caching, retry logic, and error classification.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use tokio::sync::RwLock;
use tracing::{debug, instrument, warn};

/// Idempotency key derived from action parameters.
//
// Format: SHA-256 of "(tenant_id)|(action)|(params_json)"
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct IdempotencyKey(pub String);

impl IdempotencyKey {
    /// Derive an idempotency key from tenant, action, and parameters.
    pub fn derive(tenant_id: Option<&str>, action: &str, params: &serde_json::Value) -> Self {
        let tenant = tenant_id.unwrap_or("anonymous");
        let params_str = serde_json::to_string(params).unwrap_or_default();
        let input = format!("{}|{}|{}", tenant, action, params_str);
        let hash = Sha256::digest(input.as_bytes());
        Self(hex::encode(hash))
    }

    pub fn as_str(&self) -> &str {
        &self.0
    }
}

/// Result of an action execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActionResult {
    /// Whether the action succeeded.
    pub success: bool,
    /// Action-specific output data.
    pub data: serde_json::Value,
    /// Provider reference (e.g., Stripe charge ID) for idempotency.
    pub provider_ref: Option<String>,
    /// Token usage for LLM/AI actions (None for connector actions).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub usage: Option<crate::router::flymind::Usage>,
    /// Latency in milliseconds.
    pub latency_ms: u64,
    /// Error message if failed.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

impl ActionResult {
    pub fn success(data: serde_json::Value, latency_ms: u64) -> Self {
        Self {
            success: true,
            data,
            provider_ref: None,
            usage: None,
            latency_ms,
            error: None,
        }
    }

    pub fn failure(error: String, latency_ms: u64) -> Self {
        Self {
            success: false,
            data: serde_json::Value::Null,
            provider_ref: None,
            usage: None,
            latency_ms,
            error: Some(error),
        }
    }

    pub fn with_provider_ref(mut self, ref_id: String) -> Self {
        self.provider_ref = Some(ref_id);
        self
    }
}

/// Error during action execution.
#[derive(Debug, Clone)]
pub struct ActionError {
    /// Human-readable message.
    pub message: String,
    /// Provider's error code (e.g., "card_declined").
    pub code: Option<String>,
    /// Whether this error should be retried.
    pub retryable: bool,
    /// HTTP status code if applicable.
    pub status_code: Option<u16>,
}

impl ActionError {
    pub fn retryable(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            code: None,
            retryable: true,
            status_code: None,
        }
    }

    pub fn fatal(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            code: None,
            retryable: false,
            status_code: None,
        }
    }

    pub fn with_code(mut self, code: &str) -> Self {
        self.code = Some(code.to_string());
        self
    }

    pub fn with_status(mut self, status: u16) -> Self {
        self.status_code = Some(status);
        self
    }

    pub fn from_reqwest(e: reqwest::Error) -> Self {
        if e.is_timeout() {
            Self::retryable(format!("Request timeout: {}", e))
                .with_status(e.status().map(|s| s.as_u16()).unwrap_or(408))
        } else if e.is_connect() {
            Self::retryable(format!("Connection error: {}", e))
        } else if let Some(status) = e.status() {
            let retryable = status.is_server_error() || status.as_u16() == 429;
            Self {
                message: format!("HTTP {}: {}", status.as_u16(), e),
                code: None,
                retryable,
                status_code: Some(status.as_u16()),
            }
        } else {
            Self::retryable(format!("Request failed: {}", e))
        }
    }
}

impl std::fmt::Display for ActionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.message)?;
        if let Some(ref code) = self.code {
            write!(f, " (code: {})", code)?;
        }
        write!(f, " [retryable={}]", self.retryable)
    }
}

impl std::error::Error for ActionError {}

/// Trait for external service action connectors.
pub trait ActionConnector: Send + Sync {
    /// The name of this connector (e.g., "stripe", "resend").
    fn name(&self) -> &'static str;

    /// Execute an action with the given parameters.
    //
    // The `params` JSON contains action-specific parameters (e.g., amount, currency).
    // The `idempotency_key` should be used to prevent duplicate executions.
    //
    // Returns `Ok(ActionResult)` on success, `Err(ActionError)` on failure.
    async fn execute(
        &self,
        tenant_id: Option<&str>,
        action: &str,
        params: serde_json::Value,
        idempotency_key: &IdempotencyKey,
    ) -> Result<ActionResult, ActionError>;

    /// Check if an action is supported by this connector.
    fn supports_action(&self, action: &str) -> bool;

    /// Validate parameters for a given action.
    fn validate_params(&self, action: &str, params: &serde_json::Value) -> Result<(), String> {
        let _ = (action, params);
        Ok(())
    }
}

/// Shared idempotency cache for action results.
//
// Prevents duplicate executions within the Rust runtime by caching results
// keyed on idempotency key. Results are cached with a TTL to allow for
// time-shifted retries within a reasonable window.
//
// Note: The authoritative idempotency layer is in the Go backend, which
// holds the definitive record. This cache is a local optimization only.
pub struct IdempotencyCache {
    store: Arc<RwLock<HashMap<String, (ActionResult, Instant)>>>,
    ttl: Duration,
}

impl IdempotencyCache {
    pub fn new(ttl_secs: u64) -> Self {
        Self {
            store: Arc::new(RwLock::new(HashMap::new())),
            ttl: Duration::from_secs(ttl_secs),
        }
    }

    /// Look up a cached result for an idempotency key.
    pub async fn get(&self, key: &IdempotencyKey) -> Option<ActionResult> {
        let store = self.store.read().await;
        store.get(key.as_str()).and_then(|(result, instant)| {
            if instant.elapsed() < self.ttl {
                Some(result.clone())
            } else {
                None
            }
        })
    }

    /// Store a result for an idempotency key.
    pub async fn set(&self, key: &IdempotencyKey, result: ActionResult) {
        let mut store = self.store.write().await;
        store.insert(key.0.clone(), (result, Instant::now()));
    }

    /// Clear expired entries.
    pub async fn prune(&self) {
        let mut store = self.store.write().await;
        store.retain(|_, (_, instant)| instant.elapsed() < self.ttl);
    }
}

impl Default for IdempotencyCache {
    fn default() -> Self {
        Self::new(300) // 5-minute TTL
    }
}

/// Execute an action with automatic idempotency checking and retry.
pub async fn execute_with_idempotency<C: ActionConnector>(
    connector: &C,
    cache: &IdempotencyCache,
    tenant_id: Option<&str>,
    action: &str,
    params: serde_json::Value,
    max_retries: u32,
) -> Result<ActionResult, ActionError> {
    let idempotency_key = IdempotencyKey::derive(tenant_id, action, &params);

    // Check cache first
    if let Some(cached) = cache.get(&idempotency_key).await {
        debug!(
            connector = %connector.name(),
            action = %action,
            key = %idempotency_key.as_str(),
            "Idempotency cache hit"
        );
        return Ok(cached);
    }

    // Validate params
    if let Err(e) = connector.validate_params(action, &params) {
        return Err(ActionError::fatal(format!("Invalid params for {}: {}", action, e)));
    }

    // Execute with retries
    let mut attempt = 0;
    let mut last_error: Option<ActionError> = None;

    loop {
        let start = Instant::now();

        match connector.execute(tenant_id, action, params.clone(), &idempotency_key).await {
            Ok(result) => {
                let latency = start.elapsed().as_millis() as u64;
                let mut result_with_latency = result.clone();
                result_with_latency.latency_ms = latency;
                cache.set(&idempotency_key, result.clone()).await;
                return Ok(result_with_latency);
            }
            Err(err) => {
                last_error = Some(err.clone());

                if !err.retryable || attempt >= max_retries {
                    break;
                }

                attempt += 1;
                let backoff = Duration::from_millis(100 * 2u64.saturating_pow(attempt.min(5)));
                debug!(
                    connector = %connector.name(),
                    action = %action,
                    attempt = attempt,
                    backoff_ms = backoff.as_millis(),
                    error = %err.message,
                    "Action failed, retrying"
                );
                tokio::time::sleep(backoff).await;
            }
        }
    }

    Err(last_error.unwrap_or_else(|| ActionError::fatal("Action failed after retries")))
}
