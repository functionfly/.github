//! API routes

use axum::{
    routing::{delete, get, patch, post},
    Router,
};

use super::handlers::*;

/// Create the API router
pub fn create_router() -> Router<AppState> {
    Router::new()
        // Health check
        .route("/health", get(health))

        // State routes
        .route("/v1/state/:state_id", get(get_state))
        .route("/v1/state/:state_id", delete(clear_state))
        .route("/v1/state/:state_id/key/:key", get(get_value))
        .route("/v1/state/:state_id/key/:key", post(set_value))
        .route("/v1/state/:state_id/key/:key", patch(merge_value))
        .route("/v1/state/:state_id/key/:key", delete(delete_value))
        .route("/v1/state/:state_id/keys", get(list_keys))
        .route("/v1/state/:state_id/hash", get(get_hash))

        // Snapshot routes
        .route("/v1/state/:state_id/snapshots", post(create_snapshot))
        .route("/v1/state/:state_id/snapshots", get(list_snapshots))
        .route("/v1/state/:state_id/snapshots/:version", get(get_snapshot))
        .route("/v1/state/:state_id/restore", post(restore_snapshot))

        // WASM routes
        .route("/v1/wasm/modules", post(load_wasm_module))
        .route("/v1/state/:state_id/execute", post(execute_wasm_function))
}
