//! StateFabric - Main entry point

use axum::serve;
use std::net::SocketAddr;
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use statefabric::api::{create_router, AppState};

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

    // Create app state
    let app_state = AppState::new();

    // Create router
    let app = create_router()
        .layer(TraceLayer::new_for_http())
        .with_state(app_state);

    // Get address from environment or use default
    // SECURITY P0: Bind to localhost by default to prevent unintended exposure
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
