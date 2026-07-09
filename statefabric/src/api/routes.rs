//! API routes
//!
//! Security hardening applied:
//! - Auth middleware for JWT/API key validation (wired into the protected router)
//! - Tenant isolation enforced in handlers + state manager
//! - Rate limiting per IP / per tenant
//! - CORS restricted to configured origins with explicit methods/headers
//! - Request body size limits (10 MB) to prevent memory exhaustion
//! - Security headers (HSTS, CSP, X-Frame-Options, nosniff, Referrer-Policy, Permissions-Policy)
//! - Hard request timeout
//! - P0: CORS origins must be explicitly set in production (fail-fast)

use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::Request,
    http::{HeaderValue, Method},
    middleware::{self, Next},
    response::Response,
    routing::{delete, get, patch, post},
    Extension, Router,
};
use tower_http::{
    cors::{AllowHeaders, AllowMethods, CorsLayer},
    limit::RequestBodyLimitLayer,
    timeout::TimeoutLayer,
    trace::TraceLayer,
};

use super::handlers::*;
use super::middleware::{
    auth_middleware_with_repo, rate_limit_middleware, security_headers_middleware, AuthContext,
    RateLimiter,
};

/// Default per-IP rate limit window in seconds.
const RATE_LIMIT_WINDOW_SECS: u64 = 60;
/// Default max requests per IP per window. Override via `STATEFABRIC_RATE_LIMIT_MAX`.
const DEFAULT_RATE_LIMIT_MAX: u32 = 1000;
/// Default request body limit: 10 MB. JSON state payloads and WASM modules share this.
const DEFAULT_BODY_LIMIT_BYTES: usize = 10 * 1024 * 1024;
/// Hard request timeout (across all layers, including WASM).
const REQUEST_TIMEOUT_SECS: u64 = 30;

/// Create the API router with security middleware and app state.
///
/// `api_key_repo` is the database-backed API key repo used by `auth_middleware_with_repo`
/// to validate `X-API-Key` requests. Pass `None` only in dev mode (env-var fallback).
pub fn create_router(app_state: AppState) -> Router {
    create_router_with_repo(Arc::new(app_state), None)
}

/// Create router with an explicit Postgres connection for API key lookups.
pub fn create_router_with_repo(
    app_state: Arc<AppState>,
    api_key_repo: Option<crate::storage::ApiKeyRepository>,
) -> Router {
    let max_requests: u32 = std::env::var("STATEFABRIC_RATE_LIMIT_MAX")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(DEFAULT_RATE_LIMIT_MAX);
    let rate_limiter = Arc::new(RateLimiter::new(max_requests, RATE_LIMIT_WINDOW_SECS));

    let cors_layer = build_cors_layer();

    // Public (no auth): health and metrics.
    let public = Router::new()
        .route("/health", get(health))
        .route("/health/detailed", get(health_detailed))
        .route("/metrics", get(metrics))
        .layer(Extension(app_state.clone()))
        .layer(middleware::from_fn(security_headers_middleware));

    // Protected: state, snapshots, WASM.
    let repo_for_layer = api_key_repo.clone();
    let protected = Router::new()
        // axum 0.8 requires {capture} syntax (was :capture in axum 0.7).
        .route("/v1/state/{state_id}", get(get_state))
        .route("/v1/state/{state_id}", delete(clear_state))
        .route("/v1/state/{state_id}/key/{key}", get(get_value))
        .route("/v1/state/{state_id}/key/{key}", post(set_value))
        .route("/v1/state/{state_id}/key/{key}", patch(merge_value))
        .route("/v1/state/{state_id}/key/{key}", delete(delete_value))
        .route("/v1/state/{state_id}/keys", get(list_keys))
        .route("/v1/state/{state_id}/hash", get(get_hash))
        .route("/v1/state/{state_id}/snapshots", post(create_snapshot))
        .route("/v1/state/{state_id}/snapshots", get(list_snapshots))
        .route("/v1/state/{state_id}/snapshots/{version}", get(get_snapshot))
        .route("/v1/state/{state_id}/restore", post(restore_snapshot))
        .route("/v1/wasm/modules", post(load_wasm_module))
        .route("/v1/state/{state_id}/execute", post(execute_wasm_function))
        .layer(Extension(app_state.clone()))
        .layer(middleware::from_fn(move |req, next| {
            let repo = repo_for_layer.clone();
            async move { auth_middleware_with_repo(req, next, repo).await }
        }));

    // Outer layers applied in reverse order: bottom-most runs FIRST.
    // Order on the wire (top to bottom of this list) ends up:
    //   Trace (logging) -> SecurityHeaders -> RateLimit -> BodyLimit ->
    //   Timeout -> CORS -> router
    // SECURITY: `TraceLayer::new_for_http()` does NOT capture request or
    // response bodies - it only records method, URI, status, latency. We
    // intentionally avoid the `make_span_with`/`on_request` builders that
    // could be wired to log bodies. Authorization headers and JWT/API-key
    // values are NEVER logged.
    Router::new()
        .merge(public)
        .merge(protected)
        .layer(TraceLayer::new_for_http())
        .layer(middleware::from_fn(security_headers_middleware))
        .layer(middleware::from_fn(rate_limit_middleware(
            rate_limiter,
            rate_limit_key_from_request,
        )))
        .layer(RequestBodyLimitLayer::new(DEFAULT_BODY_LIMIT_BYTES))
        // SECURITY: 30s hard timeout on every request - prevents slow-loris
        // and runaway WASM executions. The deprecated `TimeoutLayer::new`
        // returns 500 on timeout (the newer `with_status_code` returns 408);
        // behaviour is acceptable for our hard cap.
        .layer(deprecated_timeout_layer(Duration::from_secs(REQUEST_TIMEOUT_SECS)))
        .layer(cors_layer)
}

/// Wrapper to keep the deprecated `TimeoutLayer::new` call in one place so
/// we can upgrade it later without touching the router builder.
#[allow(deprecated)]
fn deprecated_timeout_layer(d: Duration) -> TimeoutLayer {
    TimeoutLayer::new(d)
}

/// Build the CORS layer with hard-coded safe defaults.
fn build_cors_layer() -> CorsLayer {
    let is_production = is_production_env();

    let allowed_origins: Vec<HeaderValue> = match std::env::var("STATEFABRIC_CORS_ORIGINS") {
        Ok(origins) => origins
            .split(',')
            .filter_map(|o| o.trim().parse::<HeaderValue>().ok())
            .collect(),
        Err(_) => {
            if is_production {
                eprintln!("FATAL: STATEFABRIC_CORS_ORIGINS must be set in production mode");
                eprintln!("Example: STATEFABRIC_CORS_ORIGINS=\"https://dashboard.example.com,https://app.example.com\"");
                std::process::exit(1);
            }
            vec![
                "http://localhost:3000".parse().unwrap(),
                "http://localhost:8080".parse().unwrap(),
            ]
        }
    };

    CorsLayer::new()
        .allow_origin(allowed_origins)
        // Explicit methods - no wildcard.
        .allow_methods(AllowMethods::list([
            Method::GET,
            Method::POST,
            Method::PATCH,
            Method::DELETE,
            Method::OPTIONS,
        ]))
        // Explicit headers - no wildcard. Only what the API actually uses.
        .allow_headers(AllowHeaders::list([
            axum::http::header::AUTHORIZATION,
            axum::http::header::CONTENT_TYPE,
            axum::http::header::HeaderName::from_static("x-api-key"),
            axum::http::header::HeaderName::from_static("x-request-id"),
        ]))
        .max_age(Duration::from_secs(600))
}

/// Extract a stable per-client key for rate limiting.
///
/// Order of precedence:
///   1. `X-Forwarded-For` first hop (when behind a trusted proxy).
///   2. Authenticated tenant_id (when the auth middleware ran first and
///      attached an `AuthContext`).
///   3. Fallback "anonymous" bucket.
fn rate_limit_key_from_request(req: &Request) -> String {
    if let Some(xff) = req
        .headers()
        .get("x-forwarded-for")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.split(',').next())
        .map(|s| s.trim().to_string())
    {
        return format!("ip:{}", xff);
    }

    if let Some(auth) = req.extensions().get::<AuthContext>() {
        return format!("tenant:{}", auth.tenant_id);
    }

    "anonymous".to_string()
}

fn is_production_env() -> bool {
    std::env::var("STATEFABRIC_ENVIRONMENT")
        .map(|v| v == "production" || v == "prod")
        .unwrap_or(false)
}

// Re-export for callers that import from `routes`.
#[allow(dead_code)]
async fn _security_headers_typed(req: Request, next: Next) -> Response {
    security_headers_middleware(req, next).await
}
