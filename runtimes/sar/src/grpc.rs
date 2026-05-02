use std::sync::Arc;
use std::net::SocketAddr;
use std::time::Duration;
use parking_lot::RwLock;
use tokio::net::TcpListener;
use tonic::{Request, Response, Status};
use tonic::transport::Server;
use tower::Layer;
use tower_http::cors::{CorsLayer, Any};
use tracing::info;
use uuid::Uuid;
use chrono::Utc;

use crate::core::{AgentId, AgentConfig};
use crate::engine::Graph;
use crate::persistence::{AgentPersistence, CachedAgentRegistry, AgentUpdate};
use crate::nats_client::NatsOrchestratorClient;
use crate::StatefulAgentRuntime;

pub struct RegisterAgentRequest {
    pub name: String,
    pub graph: Graph,
    pub priority: u32,
    pub max_concurrent_cells: u32,
    pub isolation_enabled: bool,
    pub event_subscriptions: Vec<String>,
    pub metadata: std::collections::HashMap<String, String>,
}

pub struct RegisterAgentResponse {
    pub agent: AgentInfo,
}

pub struct UnregisterAgentRequest {
    pub agent_id: String,
}

pub struct UnregisterAgentResponse {
    pub success: bool,
}

pub struct GetAgentRequest {
    pub agent_id: String,
}

pub struct GetAgentResponse {
    pub agent: Option<AgentInfo>,
}

pub struct ListAgentsRequest {
    pub page_size: u32,
    pub page_token: String,
}

pub struct ListAgentsResponse {
    pub agents: Vec<AgentInfo>,
    pub next_page_token: String,
}

pub struct UpdateAgentRequest {
    pub agent_id: String,
    pub name: Option<String>,
    pub graph: Option<Graph>,
    pub priority: Option<u32>,
    pub max_concurrent_cells: Option<u32>,
    pub isolation_enabled: Option<bool>,
    pub event_subscriptions: Vec<String>,
    pub metadata: std::collections::HashMap<String, String>,
}

pub struct UpdateAgentResponse {
    pub agent: AgentInfo,
}

pub struct HeartbeatRequest {
    pub agent_id: String,
    pub status: String,
    pub metrics: std::collections::HashMap<String, String>,
}

pub struct HeartbeatResponse {
    pub next_heartbeat: chrono::DateTime<Utc>,
}

pub struct ExecuteAgentRequest {
    pub agent_id: String,
    pub input: std::collections::HashMap<String, serde_json::Value>,
    pub tenant_id: Option<String>,
}

pub struct ExecuteAgentResponse {
    pub execution_id: String,
    pub status: String,
    pub output: Option<std::collections::HashMap<String, serde_json::Value>>,
    pub error: Option<String>,
    pub node_results: std::collections::HashMap<String, NodeResult>,
    pub started_at: Option<chrono::DateTime<Utc>>,
    pub completed_at: Option<chrono::DateTime<Utc>>,
    pub total_duration_ms: Option<u64>,
}

#[derive(Debug, Clone)]
pub struct NodeResult {
    pub node_id: String,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub duration_ms: u64,
    pub attempts: u32,
    pub status: String,
}

#[derive(Debug, Clone)]
pub struct AgentInfo {
    pub id: String,
    pub name: String,
    pub graph: Graph,
    pub priority: u32,
    pub max_concurrent_cells: u32,
    pub isolation_enabled: bool,
    pub event_subscriptions: Vec<String>,
    pub metadata: std::collections::HashMap<String, String>,
    pub status: String,
    pub registered_at: chrono::DateTime<Utc>,
    pub last_heartbeat: Option<chrono::DateTime<Utc>>,
    pub metrics: AgentMetricsInfo,
}

#[derive(Debug, Clone)]
pub struct AgentMetricsInfo {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub average_latency_ms: f64,
    pub total_cost_usd: f64,
    pub last_executed_at: Option<chrono::DateTime<Utc>>,
}

fn persistence_to_agent_info(p: &AgentPersistence) -> AgentInfo {
    AgentInfo {
        id: p.id.0.to_string(),
        name: p.name.clone(),
        graph: p.graph.clone(),
        priority: p.priority as u32,
        max_concurrent_cells: p.max_concurrent_cells as u32,
        isolation_enabled: p.isolation_enabled,
        event_subscriptions: p.event_subscriptions.clone(),
        metadata: p.metadata.clone(),
        status: format!("{:?}", p.status),
        registered_at: p.registered_at,
        last_heartbeat: p.last_heartbeat,
        metrics: AgentMetricsInfo {
            total_executions: p.metrics.total_executions,
            successful_executions: p.metrics.successful_executions,
            failed_executions: p.metrics.failed_executions,
            average_latency_ms: p.metrics.average_latency_ms,
            total_cost_usd: p.metrics.total_cost_usd,
            last_executed_at: p.metrics.last_executed_at,
        },
    }
}

pub struct GrpcServer {
    runtime: Arc<StatefulAgentRuntime>,
    registry: Arc<CachedAgentRegistry>,
    nats_client: Arc<RwLock<NatsOrchestratorClient>>,
    port: u16,
}

impl GrpcServer {
    pub fn new(
        runtime: Arc<StatefulAgentRuntime>,
        registry: Arc<CachedAgentRegistry>,
        nats_client: Arc<RwLock<NatsOrchestratorClient>>,
        port: u16,
    ) -> Self {
        Self {
            runtime,
            registry,
            nats_client,
            port,
        }
    }

    pub async fn start(self) -> Result<(), String> {
        let addr: SocketAddr = format!("0.0.0.0:{}", self.port).parse()
            .map_err(|e| format!("Invalid address: {}", e))?;

        let listener = TcpListener::bind(addr).await
            .map_err(|e| format!("Failed to bind to address: {}", e))?;

        info!(port = self.port, "Starting gRPC server");

        let cors_layer = CorsLayer::new()
            .allow_origin(Any)
            .allow_methods(Any)
            .allow_headers(Any);

        let builder = Server::builder()
            .layer(cors_layer);

        let service = AgentSvc {
            runtime: self.runtime.clone(),
            registry: self.registry.clone(),
            nats_client: self.nats_client.clone(),
        };

        let health_service = HealthSvc;

        builder
            .add_service(service)
            .add_service(health_service)
            .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener))
            .await
            .map_err(|e| format!("gRPC server error: {}", e))
    }

    pub fn get_port(&self) -> u16 {
        self.port
    }
}

#[derive(Clone)]
struct AgentSvc {
    runtime: Arc<StatefulAgentRuntime>,
    registry: Arc<CachedAgentRegistry>,
    nats_client: Arc<RwLock<NatsOrchestratorClient>>,
}

#[derive(Clone)]
struct HealthSvc;

#[tonic::async_trait]
impl AgentRegistryRpc for AgentSvc {
    async fn register_agent(&self, request: Request<RegisterAgentRequest>) -> Result<Response<RegisterAgentResponse>, Status> {
        let req = request.into_inner();

        let agent_config = AgentConfig {
            id: AgentId(Uuid::new_v4()),
            name: req.name.clone(),
            graph: req.graph,
            priority: req.priority.max(1).min(4) as u8,
            max_concurrent_cells: req.max_concurrent_cells.max(1) as usize,
            isolation_enabled: req.isolation_enabled,
            event_subscriptions: req.event_subscriptions,
        };

        let _span = tracing::info_span!("register_agent", agent_name = %req.name);

        let persistence = self.registry.register(agent_config).await
            .map_err(|e| Status::internal(format!("Failed to register agent: {}", e)))?;

        let _ = self.nats_client.read().notify_agent_registered(persistence.id, &persistence.name);

        info!(agent_id = %persistence.id, "Agent registered");

        Ok(Response::new(RegisterAgentResponse {
            agent: persistence_to_agent_info(&persistence),
        }))
    }

    async fn unregister_agent(&self, request: Request<UnregisterAgentRequest>) -> Result<Response<UnregisterAgentResponse>, Status> {
        let req = request.into_inner();

        let agent_id = Uuid::parse_str(&req.agent_id)
            .map_err(|_| Status::invalid_argument("invalid agent_id"))?;
        let agent_id = AgentId(agent_id);

        let deleted = self.registry.unregister(&agent_id).await
            .map_err(|e| Status::internal(format!("Failed to unregister agent: {}", e)))?;

        if deleted {
            let _ = self.nats_client.read().notify_agent_unregistered(&agent_id);
            info!(agent_id = %agent_id, "Agent unregistered");
        }

        Ok(Response::new(UnregisterAgentResponse { success: deleted }))
    }

    async fn get_agent(&self, request: Request<GetAgentRequest>) -> Result<Response<GetAgentResponse>, Status> {
        let req = request.into_inner();

        let agent_id = Uuid::parse_str(&req.agent_id)
            .map_err(|_| Status::invalid_argument("invalid agent_id"))?;

        let agent = self.registry.get_cached(&AgentId(agent_id))
            .ok_or_else(|| Status::not_found("agent not found"))?;

        Ok(Response::new(GetAgentResponse {
            agent: Some(persistence_to_agent_info(&agent)),
        }))
    }

    async fn list_agents(&self, _request: Request<ListAgentsRequest>) -> Result<Response<ListAgentsResponse>, Status> {
        let agents = self.registry.list_cached();

        Ok(Response::new(ListAgentsResponse {
            agents: agents.iter().map(persistence_to_agent_info).collect(),
            next_page_token: String::new(),
        }))
    }

    async fn update_agent(&self, request: Request<UpdateAgentRequest>) -> Result<Response<UpdateAgentResponse>, Status> {
        let req = request.into_inner();

        let agent_id = Uuid::parse_str(&req.agent_id)
            .map_err(|_| Status::invalid_argument("invalid agent_id"))?;

        let update = AgentUpdate {
            name: req.name,
            graph: req.graph,
            priority: req.priority.map(|p| p.max(1).min(4) as u8),
            max_concurrent_cells: req.max_concurrent_cells.map(|c| c as usize),
            isolation_enabled: req.isolation_enabled,
            event_subscriptions: if req.event_subscriptions.is_empty() { None } else { Some(req.event_subscriptions) },
        };

        let updated = self.registry.update(&AgentId(agent_id), update).await
            .map_err(|e| Status::internal(format!("Failed to update agent: {}", e)))?
            .ok_or_else(|| Status::not_found("agent not found"))?;

        info!(agent_id = %agent_id, "Agent updated");

        Ok(Response::new(UpdateAgentResponse {
            agent: persistence_to_agent_info(&updated),
        }))
    }

    async fn heartbeat(&self, request: Request<HeartbeatRequest>) -> Result<Response<HeartbeatResponse>, Status> {
        let req = request.into_inner();

        let agent_id = Uuid::parse_str(&req.agent_id)
            .map_err(|_| Status::invalid_argument("invalid agent_id"))?;

        self.registry.update_heartbeat(&AgentId(agent_id)).await
            .map_err(|e| Status::internal(format!("Failed to update heartbeat: {}", e)))?;

        let _ = self.nats_client.read().notify_agent_heartbeat(&AgentId(agent_id), &req.status);

        let next_heartbeat = Utc::now() + chrono::Duration::seconds(30);
        Ok(Response::new(HeartbeatResponse { next_heartbeat }))
    }

    async fn execute_agent(&self, request: Request<ExecuteAgentRequest>) -> Result<Response<ExecuteAgentResponse>, Status> {
        let req = request.into_inner();

        let _span = tracing::info_span!("execute_agent", agent_id = %req.agent_id);

        let agent_id = Uuid::parse_str(&req.agent_id)
            .map_err(|_| Status::invalid_argument("invalid agent_id"))?;

        let result = self.runtime.execute_agent(AgentId(agent_id), req.input).await
            .map_err(|e| Status::internal(format!("Failed to execute agent: {}", e)))?;

        Ok(Response::new(ExecuteAgentResponse {
            execution_id: result.execution_id.to_string(),
            status: format!("{:?}", result.status),
            output: result.output,
            error: result.error,
            node_results: result.node_results
                .iter()
                .map(|(id, nr)| {
                    (id.0.to_string(), NodeResult {
                        node_id: id.0.to_string(),
                        output: nr.output.clone(),
                        error: nr.error.clone(),
                        duration_ms: nr.duration_ms,
                        attempts: nr.attempts,
                        status: format!("{:?}", nr.status),
                    })
                })
                .collect(),
            started_at: result.started_at,
            completed_at: result.completed_at,
            total_duration_ms: result.total_duration_ms,
        }))
    }
}

#[tonic::async_trait]
impl HealthRpc for HealthSvc {
    async fn check(&self, _request: Request<()>) -> Result<Response<HealthResponse>, Status> {
        Ok(Response::new(HealthResponse {
            status: HealthStatus::Healthy as i32,
        }))
    }
}

#[derive(Debug, Clone)]
pub enum HealthStatus {
    Unknown = 0,
    Healthy = 1,
    Unhealthy = 2,
    Degraded = 3,
}

pub struct HealthResponse {
    pub status: i32,
}

#[tonic::async_trait]
pub trait AgentRegistryRpc: Send + Sync {
    async fn register_agent(&self, request: Request<RegisterAgentRequest>) -> Result<Response<RegisterAgentResponse>, Status>;
    async fn unregister_agent(&self, request: Request<UnregisterAgentRequest>) -> Result<Response<UnregisterAgentResponse>, Status>;
    async fn get_agent(&self, request: Request<GetAgentRequest>) -> Result<Response<GetAgentResponse>, Status>;
    async fn list_agents(&self, request: Request<ListAgentsRequest>) -> Result<Response<ListAgentsResponse>, Status>;
    async fn update_agent(&self, request: Request<UpdateAgentRequest>) -> Result<Response<UpdateAgentResponse>, Status>;
    async fn heartbeat(&self, request: Request<HeartbeatRequest>) -> Result<Response<HeartbeatResponse>, Status>;
    async fn execute_agent(&self, request: Request<ExecuteAgentRequest>) -> Result<Response<ExecuteAgentResponse>, Status>;
}

#[tonic::async_trait]
pub trait HealthRpc: Send + Sync {
    async fn check(&self, request: Request<()>) -> Result<Response<HealthResponse>, Status>;
}