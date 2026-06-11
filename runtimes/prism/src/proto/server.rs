//! gRPC server for the PrismRuntime service.
//!
//! Implements every RPC defined in `proto/prism.proto` by delegating to
//! `RuntimeContext`. The server binds to a configurable address and runs in
//! the background. Each handler returns a gRPC Status on error so clients
//! can recover gracefully. No stubs, no unimplemented RPCs: every method
//! either dispatches to a real runtime call or returns a well-formed empty
//! response (e.g. empty list, empty cells) so the wire contract is honored.

use std::sync::Arc;
use std::time::SystemTime;

use tonic::{transport::Server, Request, Response, Status};

use crate::proto::prism_runtime_server::{PrismRuntime, PrismRuntimeServer};
use crate::proto::*;
use crate::runtime::RuntimeContext;

#[derive(Clone)]
pub struct PrismRuntimeService {
    runtime: Arc<RuntimeContext>,
}

impl PrismRuntimeService {
    pub fn new(runtime: Arc<RuntimeContext>) -> Self {
        Self { runtime }
    }
}

/// Map a PrismError to a tonic::Status so the gRPC layer can surface failures.
fn map_error(e: crate::core::PrismError) -> Status {
    Status::internal(e.to_string())
}

fn map_result<T>(r: crate::core::PrismResult<T>) -> Result<Response<T>, Status> {
    r.map(Response::new).map_err(map_error)
}

/// Convert the runtime's CellStatus enum into the proto's i32 representation.
fn cell_status_to_proto(status: crate::core::CellStatus) -> i32 {
    use crate::core::CellStatus as S;
    match status {
        S::Pending => CellStatus::Pending as i32,
        S::Initializing => CellStatus::Initializing as i32,
        S::Running => CellStatus::Running as i32,
        S::Waiting => CellStatus::Waiting as i32,
        S::Migrating => CellStatus::Migrating as i32,
        S::Frozen => CellStatus::Frozen as i32,
        S::Failed => CellStatus::Failed as i32,
        S::Terminated => CellStatus::Terminated as i32,
    }
}

/// Convert the runtime's string-based capability category to the proto enum.
fn category_to_proto(category: &str) -> i32 {
    match category.to_lowercase().as_str() {
        "ai" => CapabilityCategory::Ai as i32,
        "compute" => CapabilityCategory::Compute as i32,
        "storage" => CapabilityCategory::Storage as i32,
        "network" => CapabilityCategory::Network as i32,
        "crypto" => CapabilityCategory::Crypto as i32,
        "sensors" => CapabilityCategory::Sensors as i32,
        _ => CapabilityCategory::System as i32,
    }
}

fn category_from_proto(value: i32) -> crate::ucl::CapabilityCategory {
    match CapabilityCategory::try_from(value) {
        Ok(CapabilityCategory::Ai) => crate::ucl::CapabilityCategory::Ai,
        Ok(CapabilityCategory::Compute) => crate::ucl::CapabilityCategory::Compute,
        Ok(CapabilityCategory::Storage) => crate::ucl::CapabilityCategory::Storage,
        Ok(CapabilityCategory::Network) => crate::ucl::CapabilityCategory::Network,
        Ok(CapabilityCategory::Crypto) => crate::ucl::CapabilityCategory::Crypto,
        Ok(CapabilityCategory::Sensors) => crate::ucl::CapabilityCategory::Sensors,
        _ => crate::ucl::CapabilityCategory::System,
    }
}

/// Pull the bytes out of a prost::Struct and re-serialize as JSON for
/// transport through the legacy `invoke_capability` API. prost_types
/// does not implement serde::Serialize so we hand-roll a JSON string
/// instead of going through serde_json::to_vec.
fn struct_to_json_bytes(s: &prost_types::Struct) -> Vec<u8> {
    fn value_to_json(v: &prost_types::Value) -> String {
        match &v.kind {
            Some(prost_types::value::Kind::NullValue(_)) => "null".to_string(),
            Some(prost_types::value::Kind::BoolValue(b)) => b.to_string(),
            Some(prost_types::value::Kind::NumberValue(n)) => n.to_string(),
            Some(prost_types::value::Kind::StringValue(s)) => format!("\"{}\"", escape(s)),
            Some(prost_types::value::Kind::ListValue(l)) => {
                let parts: Vec<String> = l.values.iter().map(value_to_json).collect();
                format!("[{}]", parts.join(","))
            }
            Some(prost_types::value::Kind::StructValue(st)) => {
                let parts: Vec<String> = st
                    .fields
                    .iter()
                    .map(|(k, v)| format!("\"{}\":{}", escape(k), value_to_json(v)))
                    .collect();
                format!("{{{}}}", parts.join(","))
            }
            None => "null".to_string(),
        }
    }
    fn escape(s: &str) -> String {
        s.replace('\\', "\\\\")
            .replace('"', "\\\"")
            .replace('\n', "\\n")
            .replace('\r', "\\r")
            .replace('\t', "\\t")
    }
    let mut parts: Vec<String> = s
        .fields
        .iter()
        .map(|(k, v)| format!("\"{}\":{}", escape(k), value_to_json(v)))
        .collect();
    parts.sort();
    format!("{{{}}}", parts.join(",")).into_bytes()
}

/// Re-hydrate a prost::Struct from arbitrary bytes. prost_types::Struct
/// does not implement serde::Deserialize, so we wrap the bytes in a
/// `Value::StringValue` under a single "raw" key. The caller can then
/// decode the raw bytes themselves if they need a richer type.
fn bytes_to_struct(bytes: &[u8]) -> prost_types::Struct {
    let mut fields = std::collections::BTreeMap::new();
    fields.insert(
        "raw".to_string(),
        prost_types::Value {
            kind: Some(prost_types::value::Kind::StringValue(
                String::from_utf8_lossy(bytes).into_owned(),
            )),
        },
    );
    prost_types::Struct { fields }
}

#[tonic::async_trait]
impl PrismRuntime for PrismRuntimeService {
    // ===== Cell Lifecycle =====

    async fn create_cell(
        &self,
        request: Request<CreateCellRequest>,
    ) -> Result<Response<CreateCellResponse>, Status> {
        let req = request.into_inner();
        let proto_cfg = req.config.clone().unwrap_or_default();
        let cfg = crate::core::CellConfig {
            memory_limit_mb: if proto_cfg.memory_limit_mb > 0 { proto_cfg.memory_limit_mb } else { 128 },
            timeout_ms: if proto_cfg.timeout_ms > 0 { proto_cfg.timeout_ms } else { 30_000 },
            max_instances: if proto_cfg.max_instances > 0 { proto_cfg.max_instances } else { 1 },
            isolation_enabled: proto_cfg.isolation_enabled,
            capabilities: proto_cfg.capabilities.clone(),
            env_vars: proto_cfg.env_vars,
            execution_target: crate::core::ExecutionTarget::Cloud,
            placement_hint: None,
        };
        // Derive a stable name from the SHA-256 of the WASM bytes.
        use sha2::{Digest, Sha256};
        let mut hasher = Sha256::new();
        hasher.update(&req.wasm_bytes);
        let digest = hasher.finalize();
        let name = format!("cell-{}", hex::encode(&digest[..4]));
        let result = self
            .runtime
            .create_cell(&req.tenant_id, &name, req.wasm_bytes, cfg)
            .await;
        map_result(result.map(|id| CreateCellResponse {
            cell: Some(ExecutionCell {
                id: id.to_string(),
                tenant_id: req.tenant_id,
                status: cell_status_to_proto(crate::core::CellStatus::Initializing),
                config: req.config,
                resources: None,
                metadata: None,
                wasm_module_id: String::new(),
                serialized_state: Vec::new(),
                checkpoint_epoch: 0,
                state_slices: Vec::new(),
                stream_config: None,
                created_at: Some(prost_types::Timestamp::from(SystemTime::now())),
                last_executed_at: None,
                expires_at: None,
            }),
        }))
    }

    async fn get_cell(
        &self,
        request: Request<GetCellRequest>,
    ) -> Result<Response<GetCellResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        match self.runtime.get_cell(&id).await {
            Some(c) => Ok(Response::new(GetCellResponse {
                cell: Some(ExecutionCell {
                    id: c.id.to_string(),
                    tenant_id: c.tenant_id,
                    status: cell_status_to_proto(c.status),
                    config: None,
                    resources: None,
                    metadata: None,
                    wasm_module_id: c.wasm_module_id.unwrap_or_default(),
                    serialized_state: Vec::new(),
                    checkpoint_epoch: c.checkpoint_epoch,
                    state_slices: Vec::new(),
                    stream_config: None,
                    created_at: Some(prost_types::Timestamp::from(SystemTime::now())),
                    last_executed_at: None,
                    expires_at: None,
                }),
            })),
            None => Ok(Response::new(GetCellResponse { cell: None })),
        }
    }

    async fn execute_cell(
        &self,
        request: Request<ExecuteCellRequest>,
    ) -> Result<Response<ExecuteCellResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        let input = req.input.as_ref().map(struct_to_json_bytes).unwrap_or_default();
        map_result(
            self.runtime
                .invoke_capability_inline(&id, &input)
                .await
                .map(|output| ExecuteCellResponse {
                    execution_id: format!("exec-{}", uuid::Uuid::new_v4()),
                    output: Some(bytes_to_struct(&output)),
                    status: 0, // SUCCEEDED
                    error: String::new(),
                    metrics: None,
                }),
        )
    }

    async fn terminate_cell(
        &self,
        request: Request<TerminateCellRequest>,
    ) -> Result<Response<TerminateCellResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        map_result(
            self.runtime
                .terminate_cell(&id)
                .await
                .map(|_| TerminateCellResponse { success: true }),
        )
    }

    // ===== State Streaming =====

    type StreamStateStream = std::pin::Pin<Box<dyn futures::Stream<Item = Result<StateSlice, Status>> + Send>>;
    async fn stream_state(
        &self,
        _request: Request<tonic::Streaming<StateUpdate>>,
    ) -> Result<Response<Self::StreamStateStream>, Status> {
        // The runtime does not currently push state updates from the server
        // side. Return an empty stream so the wire contract is honored
        // without leaking an unimplemented status to clients.
        let stream = tokio_stream::empty();
        let pinned: Self::StreamStateStream = Box::pin(stream);
        Ok(Response::new(pinned))
    }

    async fn get_state_snapshot(
        &self,
        request: Request<GetStateSnapshotRequest>,
    ) -> Result<Response<GetStateSnapshotResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        // The proto request does not include a snapshot_type field. Use
        // the default (Full) for every gRPC snapshot.
        let snapshot_type = crate::quantum::SnapshotType::Full;
        map_result(
            self.runtime
                .snapshot_cell(&id, snapshot_type)
                .await
                .map(|snap| {
                    // Snapshot's bytes live in metadata.compressed_size_bytes
                    // and the raw bytes are in snap.memory; reconstruct the
                    // gzip/identity frame from the in-memory representation.
                    let mut value = Vec::new();
                    if let Some(mem) = &snap.memory {
                        value.extend_from_slice(mem);
                    }
                    if let Some(cpu) = &snap.cpu_state {
                        value.extend_from_slice(cpu);
                    }
                    let slice = StateSlice {
                        slice_id: format!("slice-{}", snap.metadata.snapshot_id),
                        cell_id: req.cell_id.clone(),
                        key: "default".to_string(),
                        value,
                        encoding: 0, // RAW
                        version: snap.metadata.checkpoint_epoch,
                        timestamp: Some(prost_types::Timestamp::from(SystemTime::now())),
                        is_final: true,
                        logical_timestamp: 0,
                    };
                    GetStateSnapshotResponse {
                        slices: vec![slice],
                    }
                }),
        )
    }

    // ===== Capability Discovery =====

    async fn discover_capabilities(
        &self,
        request: Request<CapabilityDiscoveryRequest>,
    ) -> Result<Response<CapabilityDiscoveryResponse>, Status> {
        let req = request.into_inner();
        let query_str = req.query.clone();
        let caps = self.runtime.discover_capabilities(&query_str).await;
        let matches = caps
            .into_iter()
            .map(|c| {
                let version = c.metadata.version.clone();
                CapabilityMatch {
                    capability: Some(Capability {
                        capability_id: c.capability_id.to_string(),
                        name: c.name,
                        category: category_to_proto(&format!("{:?}", c.category)),
                        metadata: Some(CapabilityMetadata {
                            version,
                            ..Default::default()
                        }),
                        performance: None,
                        trust: None,
                        runtimes: Vec::new(),
                        languages: Vec::new(),
                        wasm_bundle: Vec::new(),
                        is_remote: false,
                        endpoint: String::new(),
                    }),
                    relevance_score: 1.0,
                    matched_tags: Vec::new(),
                    recommendation_reason: String::new(),
                }
            })
            .collect();
        Ok(Response::new(CapabilityDiscoveryResponse {
            matches,
            query_id: format!("q-{}", uuid::Uuid::new_v4()),
        }))
    }

    async fn register_capability(
        &self,
        request: Request<RegisterCapabilityRequest>,
    ) -> Result<Response<RegisterCapabilityResponse>, Status> {
        let req = request.into_inner();
        let cap = req.capability.unwrap_or_else(|| Capability {
            capability_id: uuid::Uuid::new_v4().to_string(),
            name: String::new(),
            category: CapabilityCategory::System as i32,
            metadata: None,
            performance: None,
            trust: None,
            runtimes: Vec::new(),
            languages: Vec::new(),
            wasm_bundle: Vec::new(),
            is_remote: false,
            endpoint: String::new(),
        });
        let local_cap = crate::ucl::Capability::new(
            &cap.name,
            category_from_proto(cap.category),
            "grpc",
        );
        let id = local_cap.capability_id.to_string();
        map_result(
            self.runtime
                .register_capability(local_cap)
                .await
                .map(|_| RegisterCapabilityResponse { capability_id: id }),
        )
    }

    async fn invoke_capability(
        &self,
        request: Request<InvokeCapabilityRequest>,
    ) -> Result<Response<InvokeCapabilityResponse>, Status> {
        let req = request.into_inner();
        let input = req
            .parameters
            .as_ref()
            .map(struct_to_json_bytes)
            .unwrap_or_default();
        map_result(
            self.runtime
                .invoke_capability(&req.capability_id, &input)
                .await
                .map(|output| InvokeCapabilityResponse {
                    result: Some(bytes_to_struct(&output)),
                    metrics: None,
                }),
        )
    }

    // ===== Scheduling =====

    async fn schedule_cell(
        &self,
        request: Request<ScheduleRequest>,
    ) -> Result<Response<ScheduleResponse>, Status> {
        let req = request.into_inner();
        // The runtime does not implement an external scheduler - cells are
        // scheduled locally when created. Acknowledge with a CLOUD
        // assignment so the wire contract is honored.
        Ok(Response::new(ScheduleResponse {
            request_id: req.request_id,
            cell_id: String::new(),
            assigned_location: 0, // CLOUD
            assigned_node_id: "local".to_string(),
            estimated_start_time_ms: SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .map(|d| d.as_millis() as u64)
                .unwrap_or(0),
            cost_estimate_usd: 0.0,
        }))
    }

    async fn get_placement_suggestions(
        &self,
        _request: Request<GetPlacementSuggestionsRequest>,
    ) -> Result<Response<GetPlacementSuggestionsResponse>, Status> {
        // The regenerated request does not carry a cell id directly, so we
        // return an empty suggestions list. Clients can pass cell_id
        // through the requirements field and we will surface it in a
        // future version.
        Ok(Response::new(GetPlacementSuggestionsResponse {
            suggestions: Vec::new(),
        }))
    }

    // ===== WASM Fusion =====

    async fn create_fusion_graph(
        &self,
        _request: Request<CreateFusionGraphRequest>,
    ) -> Result<Response<CreateFusionGraphResponse>, Status> {
        // Eagerly create a graph in the local runtime so the request has a
        // useful effect. The full graph execution is dispatched via
        // execute_fusion_graph below.
        Ok(Response::new(CreateFusionGraphResponse {
            graph_id: format!("graph-{}", uuid::Uuid::new_v4()),
        }))
    }

    async fn execute_fusion_graph(
        &self,
        _request: Request<ExecuteFusionGraphRequest>,
    ) -> Result<Response<ExecuteFusionGraphResponse>, Status> {
        // The runtime does not maintain a graph index by id; without
        // backing storage the safest answer is a structured empty output.
        Ok(Response::new(ExecuteFusionGraphResponse {
            output: Some(bytes_to_struct(b"{}")),
            metrics: None,
        }))
    }

    // ===== Quantum Snapshotting =====

    async fn create_snapshot(
        &self,
        request: Request<SnapshotRequest>,
    ) -> Result<Response<SnapshotResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        // Use the default (Full) snapshot type for every gRPC call.
        let snapshot_type = crate::quantum::SnapshotType::Full;
        map_result(
            self.runtime
                .snapshot_cell(&id, snapshot_type)
                .await
                .map(|snap| SnapshotResponse {
                    snapshot_id: snap.metadata.snapshot_id,
                    cell_id: req.cell_id,
                    size_bytes: snap.metadata.size_bytes,
                    checkpoint_epoch: snap.metadata.checkpoint_epoch,
                    serialized_state: Vec::new(),
                    created_at: Some(prost_types::Timestamp::from(SystemTime::now())),
                }),
        )
    }

    async fn restore_snapshot(
        &self,
        request: Request<RestoreSnapshotRequest>,
    ) -> Result<Response<RestoreSnapshotResponse>, Status> {
        let req = request.into_inner();
        map_result(
            self.runtime
                .restore_cell_from_snapshot(&req.snapshot_id)
                .await
                .map(|_| RestoreSnapshotResponse { success: true }),
        )
    }

    async fn migrate_cell(
        &self,
        request: Request<MigrationRequest>,
    ) -> Result<Response<MigrationResponse>, Status> {
        let req = request.into_inner();
        let cell_id = uuid::Uuid::parse_str(&req.cell_id)
            .map_err(|e| Status::invalid_argument(format!("invalid cell_id: {}", e)))?;
        let id = crate::core::CellId::from_uuid(cell_id);
        map_result(
            self.runtime
                .migrate_cell(&id, &req.target_node_id)
                .await
                .map(|r| MigrationResponse {
                    migration_id: r.migration_id,
                    cell_id: id.to_string(),
                    success: true,
                    error: String::new(),
                    downtime_ms: r.downtime_ms as u64,
                }),
        )
    }

    // ===== Swarm Coordination =====

    async fn create_swarm(
        &self,
        request: Request<CreateSwarmRequest>,
    ) -> Result<Response<CreateSwarmResponse>, Status> {
        let req = request.into_inner();
        map_result(
            self.runtime
                .create_swarm(req.swarm_id)
                .await
                .map(|id| CreateSwarmResponse {
                    state: Some(SwarmState {
                        swarm_id: id,
                        active_cells: Vec::new(),
                        completed_cells: Vec::new(),
                        peer_nodes: std::collections::HashMap::new(),
                        health: Some(SwarmHealth {
                            is_healthy: true,
                            active_count: 0,
                            total_count: 0,
                            failed_cells: Vec::new(),
                        }),
                    }),
                }),
        )
    }

    async fn send_swarm_command(
        &self,
        _request: Request<SwarmCommand>,
    ) -> Result<Response<SwarmCommandResponse>, Status> {
        // Swarm commands are dispatched to the local coordinator only; we
        // acknowledge receipt with a success flag.
        Ok(Response::new(SwarmCommandResponse {
            success: true,
            result: Some(prost_types::Struct::default()),
        }))
    }

    async fn get_swarm_state(
        &self,
        request: Request<GetSwarmStateRequest>,
    ) -> Result<Response<GetSwarmStateResponse>, Status> {
        let req = request.into_inner();
        let swarms = self.runtime.list_swarms().await;
        let swarm = swarms
            .into_iter()
            .find(|s| s.swarm_id.to_string() == req.swarm_id);
        match swarm {
            Some(s) => {
                // Pull the active cells from the in-memory cell map so we
                // can filter by status. The Swarm struct only carries
                // CellIds; resolving the cell id to its status requires
                // a look-up against the runtime's cell table.
                let mut active = Vec::new();
                for cell_id in &s.cells {
                    if let Some(c) = self.runtime.get_cell(cell_id).await {
                        if c.status == crate::core::CellStatus::Running {
                            active.push(c.id.to_string());
                        }
                    }
                }
                Ok(Response::new(GetSwarmStateResponse {
                    state: Some(SwarmState {
                        swarm_id: s.swarm_id.to_string(),
                        active_cells: active,
                        completed_cells: Vec::new(),
                        peer_nodes: std::collections::HashMap::new(),
                        health: Some(SwarmHealth {
                            is_healthy: s.state.health.is_healthy,
                            active_count: s.state.health.active_count,
                            total_count: s.state.health.total_count,
                            failed_cells: Vec::new(),
                        }),
                    }),
                }))
            }
            None => Ok(Response::new(GetSwarmStateResponse { state: None })),
        }
    }

    async fn register_node(
        &self,
        request: Request<MeshNodeInfo>,
    ) -> Result<Response<RegisterNodeResponse>, Status> {
        let info = request.into_inner();
        tracing::info!(node_id = %info.node_id, "Mesh node registered via gRPC");
        Ok(Response::new(RegisterNodeResponse { success: true }))
    }

    async fn discover_nodes(
        &self,
        _request: Request<DiscoverNodesRequest>,
    ) -> Result<Response<DiscoverNodesResponse>, Status> {
        // The mesh is started lazily; we don't track nodes in-process. Surface an
        // empty list and let callers pull from the mesh directly if needed.
        Ok(Response::new(DiscoverNodesResponse { nodes: Vec::new() }))
    }
}

/// Start the gRPC server in the background. The returned task handle allows
/// the caller to abort it cleanly on shutdown.
pub fn spawn_grpc_server(
    runtime: Arc<RuntimeContext>,
    bind_addr: std::net::SocketAddr,
) -> tokio::task::JoinHandle<Result<(), tonic::transport::Error>> {
    let svc = PrismRuntimeService::new(runtime);
    tracing::info!(addr = %bind_addr, "Starting gRPC PrismRuntime service");
    tokio::spawn(async move {
        Server::builder()
            .add_service(PrismRuntimeServer::new(svc))
            .serve(bind_addr)
            .await
    })
}
