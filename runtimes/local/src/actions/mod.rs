//! Action layer — external service connectors with idempotency.
//!
//! Provides a `ActionConnector` trait for executing external service operations
//! (Stripe, Resend, Shopify, etc.) as graph tool nodes. Each connector implements:
//! - **Idempotency**: deduplication via hash of (tenant_id, action, params)
//! - **Retry with backoff**: transient failures are retried with exponential backoff
//! - **Error classification**: distinguishes retryable vs fatal errors
//!
//! ## Architecture
//!
//! Each connector is a graph tool node that executes a specific action:
//! - `StripeConnector`: charge, create customer, update subscription
//! - `ResendConnector`: send email, send verification email
//! - `ShopifyConnector`: create order, manage customers, update inventory
//! - `HttpConnector`: generic REST API calls
//!
//! ## Idempotency
//!
//! Uses SHA-256 of `(tenant_id | action_name | normalized_params_json)` as the
//! idempotency key. On the Rust side, results are cached in-memory with TTL.
//! The Go backend provides the authoritative idempotency layer (same as existing
//! Stripe webhook handler). Rust acts as a cache layer to avoid redundant calls.
//!
//! ## Vault Integration
//!
//! Credentials are read from environment variables (or Vault in production).
//! The WASM cell sees only the action name and parameters — never the credentials.

pub mod connector;
pub mod stripe;
pub mod resend;
pub mod shopify;
pub mod http;

pub use connector::{ActionConnector, ActionError, ActionResult, IdempotencyKey};
pub use stripe::StripeConnector;
pub use resend::ResendConnector;
pub use shopify::ShopifyConnector;
pub use http::HttpConnector;
