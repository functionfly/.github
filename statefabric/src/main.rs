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
    let addr = std::env::var("ADDR")
        .unwrap_or_else(|_| "127.0.0.1:8080".to_string())
        .parse::<SocketAddr>()
        .unwrap_or_else(|_| SocketAddr::from(([127, 0, 0, 1], 8080)));

    // Start server
    let listener = tokio::net::TcpListener::bind(&addr)
        .await
        .expect("Failed to bind to address");

    tracing::info!("Listening on {}", addr);

    serve(listener, app)
        .await
        .expect("Failed to start server");
}
