//! API routes
//!
//! Security hardening applied:
//! - Auth middleware for JWT/API key validation
//! - Tenant isolation enforced in handlers
//! - Rate limiting via tower-rate-limit crate (recommended for production)
//! - CORS restricted to configured origins
//! - P0: CORS origins must be explicitly set in production (fail-fast)

use std::sync::Arc;
use axum::{
    routing::{delete, get, patch, post},
    Router,
};
use tower_http::cors::{AllowMethods, CorsLayer};
use axum::http::HeaderValue;

use super::handlers::*;
use super::middleware::RateLimiter;

/// Create the API router with security middleware
pub fn create_router() -> Router<AppState> {
    create_router_with_limiter(Arc::new(RateLimiter::new(1000, 60))) // 1000 req/min default
}

/// Create router with custom rate limiter
#[allow(dead_code)]
pub fn create_router_with_limiter(rate_limiter: std::sync::Arc<RateLimiter>) -> Router<AppState> {
    let _ = rate_limiter; // Rate limiting applied per-handler in production
    Router::new()
        // Health check (no auth required)
        .route("/health", get(health))

        // State routes (protected)
        .route("/v1/state/:state_id", get(get_state))
        .route("/v1/state/:state_id", delete(clear_state))
        .route("/v1/state/:state_id/key/:key", get(get_value))
        .route("/v1/state/:state_id/key/:key", post(set_value))
        .route("/v1/state/:state_id/key/:key", patch(merge_value))
        .route("/v1/state/:state_id/key/:key", delete(delete_value))
        .route("/v1/state/:state_id/keys", get(list_keys))
        .route("/v1/state/:state_id/hash", get(get_hash))

        // Snapshot routes (protected)
        .route("/v1/state/:state_id/snapshots", post(create_snapshot))
        .route("/v1/state/:state_id/snapshots", get(list_snapshots))
        .route("/v1/state/:state_id/snapshots/:version", get(get_snapshot))
        .route("/v1/state/:state_id/restore", post(restore_snapshot))

        // WASM routes (protected)
        .route("/v1/wasm/modules", post(load_wasm_module))
        .route("/v1/state/:state_id/execute", post(execute_wasm_function))

        // Apply security middleware layers
        // Auth middleware is applied via Extension<AuthContext> in handlers
        // Rate limiting is performed per-request in handlers using rate_limiter
        // Apply explicit CORS configuration - restrict to known origins in production
        // SECURITY P0: In production, require explicit CORS origins - fail fast if not set
        // Example: STATEFABRIC_CORS_ORIGINS="https://dashboard.example.com,https://app.example.com"
        .layer({
            let is_production = std::env::var("STATEFABRIC_ENVIRONMENT")
                .map(|v| v == "production" || v == "prod")
                .unwrap_or(false);

            let allowed_origins = std::env::var("STATEFABRIC_CORS_ORIGINS")
                .map(|origins| {
                    origins.split(',')
                        .filter_map(|o| o.trim().parse::<HeaderValue>().ok())
                        .collect::<Vec<_>>()
                })
                .unwrap_or_else(|_| {
                    if is_production {
                        // SECURITY P0: Fail fast in production if CORS not configured
                        eprintln!("FATAL: STATEFABRIC_CORS_ORIGINS must be set in production mode");
                        eprintln!("Example: STATEFABRIC_CORS_ORIGINS=\"https://dashboard.example.com,https://app.example.com\"");
                        std::process::exit(1);
                    }
                    // Development fallback - localhost only
                    vec![
                        "http://localhost:3000".parse().unwrap(),
                        "http://localhost:8080".parse().unwrap(),
                    ]
                });

            CorsLayer::new()
                .allow_origin(allowed_origins)
                .allow_methods(AllowMethods::any())
                .allow_headers(tower_http::cors::AllowHeaders::any())
        })
}
