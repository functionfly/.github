//! StateFabric - Main entry point
//!
//! Production configuration requires:
//! - STATEFABRIC_JWT_SECRET (required in all environments)
//! - DATABASE_URL (for PostgreSQL)
//! - STATEFABRIC_CORS_ORIGINS (required in production)
//! - Object storage (R2, S3, or local) for event/snapshot persistence

use axum::{
    extract::Request,
    http::header::HeaderName,
    middleware::Next,
    response::Response,
    serve,
};
use std::net::SocketAddr;
use std::sync::Arc;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use statefabric::api::{create_router_with_repo, AppState, setup_prometheus_exporter, setup_metrics_handle};
use statefabric::api::middleware::{set_audit_logger, PostgresAuditLogger};

/// Correlation ID middleware - adds X-Request-ID to responses
#[allow(dead_code)]
async fn correlation_id_middleware(
    req: Request,
    next: Next,
) -> Response {
    let request_id = req.headers()
        .get("X-Request-ID")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string())
        .unwrap_or_else(|| uuid::Uuid::new_v4().to_string());

    let mut response = next.run(req).await;

    response.headers_mut().insert(
        HeaderName::from_static("x-request-id"),
        request_id.parse().unwrap_or_else(|_| "unknown".parse().unwrap()),
    );

    response
}

/// Validate required environment variables for production
fn validate_environment() {
    let is_production = std::env::var("STATEFABRIC_ENVIRONMENT")
        .map(|v| v == "production" || v == "prod")
        .unwrap_or(false);

    // JWT secret is ALWAYS required
    let jwt_secret = std::env::var("STATEFABRIC_JWT_SECRET");
    if jwt_secret.is_err() {
        eprintln!("FATAL: STATEFABRIC_JWT_SECRET environment variable is required");
        if is_production {
            std::process::exit(1);
        }
        eprintln!("WARNING: Running without JWT secret is insecure - DO NOT use in production");
    } else if is_production {
        // SECURITY: Enforce minimum secret length in production (256 bits = 32 bytes)
        let secret = jwt_secret.unwrap();
        if secret.len() < 32 {
            eprintln!("FATAL: STATEFABRIC_JWT_SECRET must be at least 32 bytes (256 bits) in production");
            eprintln!("Current length: {} bytes", secret.len());
            eprintln!("Generate a secure secret: openssl rand -hex 32");
            std::process::exit(1);
        }
    }

    // CORS origins are required in production
    if is_production && std::env::var("STATEFABRIC_CORS_ORIGINS").is_err() {
        eprintln!("FATAL: STATEFABRIC_CORS_ORIGINS must be set in production mode");
        eprintln!("Example: STATEFABRIC_CORS_ORIGINS=\"https://dashboard.example.com\"");
        std::process::exit(1);
    }

    // SECURITY P0: Encryption is REQUIRED in production
    if is_production && std::env::var("STATEFABRIC_ENCRYPTION_KEY").is_err() {
        eprintln!("FATAL: STATEFABRIC_ENCRYPTION_KEY must be set in production mode");
        eprintln!("Encryption at rest is mandatory - data will not be stored unencrypted");
        eprintln!("Generate a secure key: openssl rand -hex 32");
        std::process::exit(1);
    }

    // Warn about in-memory storage in production
    if is_production && std::env::var("DATABASE_URL").is_err() {
        eprintln!("WARNING: DATABASE_URL not set - using in-memory storage (data will be lost on restart)");
        eprintln!("WARNING: This is INSECURE and data-loss prone for production!");
    }
}

/// Create fully configured AppState with all storage backends
async fn create_production_app_state() -> AppState {
    let database_url = std::env::var("DATABASE_URL")
        .expect("DATABASE_URL must be set in production");

    tracing::info!("Connecting to PostgreSQL...");

    let pg_pool = sqlx::postgres::PgPoolOptions::new()
        .max_connections(20)
        .min_connections(5)
        .acquire_timeout(std::time::Duration::from_secs(30))
        .idle_timeout(std::time::Duration::from_secs(600))
        .connect(&database_url)
        .await
        .expect("Failed to connect to PostgreSQL");

    tracing::info!("PostgreSQL connected successfully");

    // Install the Postgres-backed audit logger so security events are
    // persisted alongside the existing tracing logs. set_audit_logger is a
    // no-op if a logger was already installed.
    set_audit_logger(std::sync::Arc::new(PostgresAuditLogger::new(pg_pool.clone())));

    // Create repositories
    let state_repo = statefabric::storage::PostgresStateRepository::new(pg_pool.clone());
    let event_repo = statefabric::storage::PostgresEventRepository::new(pg_pool.clone());
    let snapshot_repo = statefabric::storage::PostgresSnapshotRepository::new(pg_pool.clone());

    // Setup object storage (R2, S3, or local).
    //
    // We construct the storage directly as `Arc<dyn ObjectStore + Send + Sync>`
    // to avoid the unsafe `Box::into_raw` -> `Arc::from_raw` reinterpretation
    // pattern. If we later need to wrap the store in `EncryptedObjectStore`,
    // we use `Arc::new(EncryptedObjectStore::new(arc_store))` so the inner
    // Arc is moved into the wrapper without any pointer casts.
    let object_store: Option<Arc<dyn statefabric::storage::ObjectStore + Send + Sync>> =
        if let Ok(storage_config) = statefabric::storage::StorageConfig::from_env() {
            tracing::info!("Configuring object storage: {:?}", storage_config.backend);
            let arc_store: Arc<dyn statefabric::storage::ObjectStore + Send + Sync> =
                statefabric::storage::create_object_store(&storage_config)
                    .await
                    .expect("Failed to create object store");

            if std::env::var("STATEFABRIC_ENCRYPTION_KEY").is_ok() {
                // Build the encrypted wrapper around the Arc directly.
                let encrypted = statefabric::storage::EncryptedObjectStore::new(arc_store);
                Some(Arc::new(encrypted) as Arc<dyn statefabric::storage::ObjectStore + Send + Sync>)
            } else {
                Some(arc_store)
            }
        } else {
            tracing::warn!("No object storage configured - events/snapshots will not be persisted");
            None
        };

    // Setup Redis cache
    let redis_pool = if let Ok(redis_url) = std::env::var("REDIS_URL") {
        tracing::info!("Connecting to Redis...");
        let redis = redis::Client::open(redis_url.as_str())
            .expect("Failed to create Redis client");
        let conn = redis::aio::ConnectionManager::new(redis)
            .await
            .expect("Failed to connect to Redis");
        tracing::info!("Redis connected successfully");
        Some(conn)
    } else {
        tracing::warn!("REDIS_URL not set - caching disabled");
        None
    };

    // Build app state with all backends
    if let (Some(redis), Some(store)) = (redis_pool, object_store) {
        AppState::with_storage(pg_pool, redis, store, state_repo, event_repo, snapshot_repo)
    } else if let Some(pg_pool) = Option::Some(pg_pool) {
        AppState::with_postgres(pg_pool, state_repo, event_repo, snapshot_repo)
    } else {
        panic!("DATABASE_URL must be set - no storage available");
    }
}

#[tokio::main]
async fn main() {
    // Initialize logging
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "statefabric=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("Starting StateFabric server...");
    tracing::info!("Version: {}", env!("CARGO_PKG_VERSION"));

    // SECURITY P0: Validate environment before starting
    validate_environment();

    // Setup Prometheus metrics handle for /metrics endpoint
    if let Err(e) = setup_metrics_handle() {
        tracing::warn!("Failed to setup metrics handle: {}", e);
    }

    // Setup Prometheus metrics server on separate port 9090 (if enabled)
    if std::env::var("STATEFABRIC_METRICS_ENABLED").unwrap_or_default() == "true" {
        let metrics_port: u16 = std::env::var("STATEFABRIC_METRICS_PORT")
            .unwrap_or_else(|_| "9090".to_string())
            .parse()
            .unwrap_or(9090);

        if let Err(e) = setup_prometheus_exporter(metrics_port) {
            tracing::error!("Failed to setup Prometheus exporter: {}", e);
        }
    }

    // Create app state - production or development
    let app_state = if std::env::var("STATEFABRIC_ENVIRONMENT")
        .map(|v| v == "production" || v == "prod")
        .unwrap_or(false)
    {
        create_production_app_state().await
    } else if std::env::var("DATABASE_URL").is_ok() {
        create_production_app_state().await
    } else {
        // Development mode - in-memory state
        tracing::warn!("No DATABASE_URL set, running in development mode with in-memory state");
        tracing::warn!("DO NOT use in-memory mode in production!");
        AppState::new()
    };

    // Pull the API key repo out of the app state so we can pass it into the
    // router (auth middleware uses it to validate X-API-Key requests).
    let api_key_repo = app_state.api_key_repo.clone();
    let app_state = Arc::new(app_state);

    // Create Axum router (app_state is stored in Router via State extractor)
    let app = statefabric::api::create_router_with_repo(app_state, api_key_repo);

    let addr = std::env::var("ADDR")
        .unwrap_or_else(|_| "127.0.0.1:8080".to_string())
        .parse::<SocketAddr>()
        .map_err(|e| {
            eprintln!("FATAL: Invalid ADDR format '{}': {}", std::env::var("ADDR").unwrap_or_default(), e);
            e
        })
        .unwrap_or_else(|_| SocketAddr::from(([127, 0, 0, 1], 8080)));

    // SECURITY P0: Warn if binding to all interfaces in production
    if addr.ip().is_unspecified() {
        eprintln!("WARNING: Binding to 0.0.0.0 - service will be publicly accessible!");
        eprintln!("Set ADDR=127.0.0.1:8080 for localhost-only or use a firewall.");
    }

    // Start server
    let listener = tokio::net::TcpListener::bind(&addr)
        .await
        .map_err(|e| {
            eprintln!("FATAL: Failed to bind to address {}: {}", addr, e);
            e
        })
        .expect("listener setup failed"); // Only panics on programmer error, not runtime

    tracing::info!("Listening on {}", addr);

    serve(listener, app)
        .await
        .expect("server future panicked"); // Only on programmer error
}
