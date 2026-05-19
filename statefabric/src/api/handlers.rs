//! API handlers

use axum::{
    extract::{Path, State},
    response::Json,
    http::StatusCode,
};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::models::{CreateSnapshotRequest, RestoreSnapshotRequest};
use crate::state::StateManager;
use crate::wasm::WasmConfig;

/// App state for Axum
#[derive(Clone)]
pub struct AppState {
    pub state_manager: std::sync::Arc<StateManager>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            state_manager: std::sync::Arc::new(StateManager::new()),
        }
    }

    pub fn with_storage(
        object_store: Box<dyn crate::storage::ObjectStore + Send + Sync>,
        snapshot_repo: crate::storage::PostgresSnapshotRepository,
        event_repo: crate::storage::PostgresEventRepository,
    ) -> Self {
        Self {
            state_manager: std::sync::Arc::new(StateManager::with_storage(object_store, snapshot_repo, event_repo)),
        }
    }

    pub fn with_wasm(
        object_store: Box<dyn crate::storage::ObjectStore + Send + Sync>,
        snapshot_repo: crate::storage::PostgresSnapshotRepository,
        event_repo: crate::storage::PostgresEventRepository,
        wasm_config: WasmConfig,
    ) -> Result<Self, Box<dyn std::error::Error>> {
        let state_manager = StateManager::with_wasm(object_store, snapshot_repo, event_repo, wasm_config)?;
        Ok(Self {
            state_manager: std::sync::Arc::new(state_manager),
        })
    }
}

// ==================== Request/Response Types ====================

#[derive(Debug, Deserialize)]
pub struct SetValueRequest {
    pub value: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct SetValueResponse {
    pub key: String,
    pub value: serde_json::Value,
    pub state_hash: String,
}

#[derive(Debug, Serialize)]
pub struct GetValueResponse {
    pub key: String,
    pub value: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct StateResponse {
    pub state_id: String,
    pub data: serde_json::Value,
    pub key_count: i32,
    pub size_bytes: i64,
    pub hash: String,
}

#[derive(Debug, Serialize)]
pub struct EventResponse {
    pub id: String,
    pub state_id: String,
    pub event_type: String,
    pub sequence: i64,
    pub timestamp: String,
}

#[derive(Debug, Serialize)]
pub struct SnapshotResponse {
    pub id: String,
    pub state_id: String,
    pub snapshot_version: i64,
    pub label: Option<String>,
    pub key_count: i32,
    pub created_at: String,
}

#[derive(Debug, Deserialize)]
pub struct LoadWasmRequest {
    pub name: String,
    pub wasm_bytes: Vec<u8>,
}

#[derive(Debug, Serialize)]
pub struct LoadWasmResponse {
    pub name: String,
    pub loaded: bool,
}

#[derive(Debug, Deserialize)]
pub struct ExecuteWasmRequest {
    pub module_name: String,
    pub function_name: String,
    pub input: Vec<u8>,
}

#[derive(Debug, Serialize)]
pub struct ExecuteWasmResponse {
    pub success: bool,
    pub output: Vec<u8>,
    pub committed_events: Vec<EventResponse>,
    pub gas_used: Option<u64>,
    pub execution_time_ms: u64,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

// ==================== State Handlers ====================

/// Get entire state
pub async fn get_state(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
) -> Result<Json<StateResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let data = state.state_manager.get(uuid)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;

    let key_count = data.as_object()
        .map(|o| o.len() as i32)
        .unwrap_or(0);

    let size_bytes = serde_json::to_vec(&data)
        .map(|v| v.len() as i64)
        .unwrap_or(0);

    let hash = state.state_manager.hash(uuid)
        .await
        .unwrap_or_default();

    Ok(Json(StateResponse {
        state_id,
        data,
        key_count,
        size_bytes,
        hash,
    }))
}

/// Set a value in state
pub async fn set_value(
    State(state): State<AppState>,
    Path((state_id, key)): Path<(String, String)>,
    Json(req): Json<SetValueRequest>,
) -> Result<Json<SetValueResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    state.state_manager.set(uuid, key.clone(), req.value.clone())
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let state_hash = state.state_manager.hash(uuid)
        .await
        .unwrap_or_default();

    Ok(Json(SetValueResponse {
        key,
        value: req.value,
        state_hash,
    }))
}

/// Get a specific value
pub async fn get_value(
    State(state): State<AppState>,
    Path((state_id, key)): Path<(String, String)>,
) -> Result<Json<GetValueResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let value = state.state_manager.get_key(uuid, &key)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;

    Ok(Json(GetValueResponse {
        key,
        value,
    }))
}

/// Delete a value
pub async fn delete_value(
    State(state): State<AppState>,
    Path((state_id, key)): Path<(String, String)>,
) -> Result<StatusCode, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    state.state_manager.delete(uuid, &key)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(StatusCode::NO_CONTENT)
}

/// Merge a value into an existing key
pub async fn merge_value(
    State(state): State<AppState>,
    Path((state_id, key)): Path<(String, String)>,
    Json(req): Json<SetValueRequest>,
) -> Result<Json<SetValueResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let merged_value = state.state_manager.merge(uuid, key.clone(), req.value)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let state_hash = state.state_manager.hash(uuid)
        .await
        .unwrap_or_default();

    Ok(Json(SetValueResponse {
        key,
        value: merged_value,
        state_hash,
    }))
}

/// Clear all state
pub async fn clear_state(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
) -> Result<StatusCode, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    state.state_manager.clear(uuid)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(StatusCode::NO_CONTENT)
}

/// Get list of keys
pub async fn list_keys(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
) -> Result<Json<Vec<String>>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let keys = state.state_manager.keys(uuid)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;

    Ok(Json(keys))
}

/// Get state hash
pub async fn get_hash(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
) -> Result<Json<String>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let hash = state.state_manager.hash(uuid)
        .await
        .unwrap_or_default();

    Ok(Json(hash))
}

// ==================== Health Handler ====================

pub async fn health() -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "statefabric",
        "version": env!("CARGO_PKG_VERSION")
    }))
}

// ==================== Snapshot Handlers ====================

/// Create a snapshot of the current state
pub async fn create_snapshot(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
    Json(req): Json<CreateSnapshotRequest>,
) -> Result<Json<SnapshotResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let snapshot = state.state_manager.create_snapshot(uuid, req)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(SnapshotResponse {
        id: snapshot.id.to_string(),
        state_id: snapshot.state_id.to_string(),
        snapshot_version: snapshot.snapshot_version,
        label: snapshot.label,
        key_count: snapshot.key_count,
        created_at: snapshot.created_at.to_rfc3339(),
    }))
}

/// Restore state from a snapshot
pub async fn restore_snapshot(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
    Json(req): Json<RestoreSnapshotRequest>,
) -> Result<StatusCode, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    state.state_manager.restore_snapshot(uuid, req)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(StatusCode::OK)
}

/// List snapshots for a state
pub async fn list_snapshots(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
) -> Result<Json<Vec<SnapshotResponse>>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let snapshots = state.state_manager.list_snapshots(uuid)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let responses = snapshots.into_iter().map(|s| SnapshotResponse {
        id: s.id.to_string(),
        state_id: s.state_id.to_string(),
        snapshot_version: s.snapshot_version,
        label: s.label,
        key_count: s.key_count,
        created_at: s.created_at.to_rfc3339(),
    }).collect();

    Ok(Json(responses))
}

// ==================== WASM Handlers ====================

/// Load a WASM module
pub async fn load_wasm_module(
    State(state): State<AppState>,
    Json(req): Json<LoadWasmRequest>,
) -> Result<Json<LoadWasmResponse>, StatusCode> {
    if !state.state_manager.has_wasm_runtime() {
        return Err(StatusCode::SERVICE_UNAVAILABLE);
    }

    state.state_manager.load_wasm_module(&req.name, &req.wasm_bytes)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(LoadWasmResponse {
        name: req.name,
        loaded: true,
    }))
}

/// Execute a WASM function
pub async fn execute_wasm_function(
    State(state): State<AppState>,
    Path(state_id): Path<String>,
    Json(req): Json<ExecuteWasmRequest>,
) -> Result<Json<ExecuteWasmResponse>, StatusCode> {
    if !state.state_manager.has_wasm_runtime() {
        return Err(StatusCode::SERVICE_UNAVAILABLE);
    }

    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let result = state.state_manager.clone().execute_wasm_function(
        &req.module_name,
        &req.function_name,
        uuid,
        &req.input,
    )
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let committed_events = result.committed_events.into_iter().map(|e| EventResponse {
        id: e.id.to_string(),
        state_id: e.state_id.to_string(),
        event_type: e.event_type.as_str().to_string(),
        sequence: e.sequence_num,
        timestamp: e.timestamp.to_rfc3339(),
    }).collect();

    Ok(Json(ExecuteWasmResponse {
        success: result.success,
        output: result.output,
        committed_events,
        gas_used: result.gas_used,
        execution_time_ms: result.execution_time_ms,
    }))
}

/// Get a specific snapshot by version
pub async fn get_snapshot(
    State(state): State<AppState>,
    Path((state_id, version)): Path<(String, i64)>,
) -> Result<Json<SnapshotResponse>, StatusCode> {
    let uuid = Uuid::parse_str(&state_id)
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let snapshot = state.state_manager.get_snapshot(uuid, version)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
        .ok_or(StatusCode::NOT_FOUND)?;

    Ok(Json(SnapshotResponse {
        id: snapshot.id.to_string(),
        state_id: snapshot.state_id.to_string(),
        snapshot_version: snapshot.snapshot_version,
        label: snapshot.label,
        key_count: snapshot.key_count,
        created_at: snapshot.created_at.to_rfc3339(),
    }))
}
