use std::sync::Arc;
use std::net::SocketAddr;
use axum::{Router, Json, routing::get, routing::post, routing::delete, extract::Path};
use serde::{Deserialize, Serialize};
use tokio::net::TcpListener;
use tracing::info;

use crate::{AgentId, AgentConfig, Graph, Node, NodeId, NodeType, Edge};
use crate::persistence::{CachedAgentRegistry, AgentPersistence};
use crate::nats_client::NatsOrchestratorClient;
use crate::StatefulAgentRuntime;

#[derive(Debug, Deserialize)]
pub struct RegisterAgentRequest {
    pub name: String,
    pub priority: Option<u32>,
    pub max_concurrent_cells: Option<u32>,
    pub isolation_enabled: Option<bool>,
}

#[derive(Debug, Serialize)]
pub struct AgentResponse {
    pub id: String,
    pub name: String,
    pub status: String,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

#[derive(Debug, Deserialize)]
pub struct UnregisterAgentRequest {
    pub agent_id: String,
}

#[derive(Debug, Serialize)]
pub struct UnregisterAgentResponse {
    pub success: bool,
}

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
}

pub struct HttpServer {
    registry: Arc<CachedAgentRegistry>,
    nats_client: Arc<parking_lot::RwLock<NatsOrchestratorClient>>,
    port: u16,
}

impl HttpServer {
    pub fn new(
        registry: Arc<CachedAgentRegistry>,
        nats_client: Arc<parking_lot::RwLock<NatsOrchestratorClient>>,
        port: u16,
    ) -> Self {
        Self {
            registry,
            nats_client,
            port,
        }
    }

    pub async fn start(self) -> Result<(), String> {
        let addr: SocketAddr = format!("0.0.0.0:{}", self.port).parse()
            .map_err(|e| format!("Invalid address: {}", e))?;

        let registry = self.registry.clone();
        let nats_client = self.nats_client.clone();

        async fn register_handler(
            registry: Arc<CachedAgentRegistry>,
            nats_client: Arc<parking_lot::RwLock<NatsOrchestratorClient>>,
            Json(req): Json<RegisterAgentRequest>,
        ) -> Result<Json<AgentResponse>, (axum::http::StatusCode, Json<ErrorResponse>)> {
            let graph_id = uuid::Uuid::new_v4();
            let input_node = NodeId(uuid::Uuid::new_v4());
            let llm_node = NodeId(uuid::Uuid::new_v4());
            let output_node = NodeId(uuid::Uuid::new_v4());

            let mut graph = Graph::new(graph_id, format!("{}-graph", req.name));
            graph.add_node(Node::new(input_node, "Input".to_string(), NodeType::Passthrough));
            graph.add_node(Node::new(llm_node, "LLM".to_string(), NodeType::llm("Process input".to_string())));
            graph.add_node(Node::new(output_node, "Output".to_string(), NodeType::Passthrough));
            graph.add_edge(Edge::dataflow(input_node, llm_node));
            graph.add_edge(Edge::dataflow(llm_node, output_node));

            let agent_config = AgentConfig {
                id: AgentId(uuid::Uuid::new_v4()),
                name: req.name.clone(),
                graph,
                priority: req.priority.unwrap_or(2).max(1).min(4) as u8,
                max_concurrent_cells: req.max_concurrent_cells.unwrap_or(100).max(1) as usize,
                isolation_enabled: req.isolation_enabled.unwrap_or(true),
                event_subscriptions: vec![],
            };

            let persistence = registry.register(agent_config).await
                .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string() })))?;

            let _ = nats_client.read().notify_agent_registered(persistence.id, &persistence.name);

            info!(agent_id = %persistence.id, name = %persistence.name, "Agent registered via HTTP API");

            Ok(Json(AgentResponse {
                id: persistence.id.0.to_string(),
                name: persistence.name,
                status: format!("{:?}", persistence.status),
            }))
        }

        async fn list_handler(
            registry: Arc<CachedAgentRegistry>,
        ) -> Result<Json<Vec<AgentResponse>>, (axum::http::StatusCode, Json<ErrorResponse>)> {
            let agents = registry.list_cached();
            Ok(Json(agents.iter().map(|a| AgentResponse {
                id: a.id.0.to_string(),
                name: a.name.clone(),
                status: format!("{:?}", a.status),
            }).collect()))
        }

        async fn unregister_handler(
            registry: Arc<CachedAgentRegistry>,
            nats_client: Arc<parking_lot::RwLock<NatsOrchestratorClient>>,
            Path(agent_id): Path<String>,
        ) -> Result<Json<UnregisterAgentResponse>, (axum::http::StatusCode, Json<ErrorResponse>)> {
            let agent_uuid = uuid::Uuid::parse_str(&agent_id)
                .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string() })))?;

            let deleted = registry.unregister(&AgentId(agent_uuid)).await
                .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string() })))?;

            if deleted {
                let _ = nats_client.read().notify_agent_unregistered(&AgentId(agent_uuid));
                info!(agent_id = %agent_id, "Agent unregistered via HTTP API");
            }

            Ok(Json(UnregisterAgentResponse { success: deleted }))
        }

        async fn health_handler() -> Json<HealthResponse> {
            Json(HealthResponse { status: "healthy".to_string() })
        }

        let app = Router::new()
            .route("/health", get(health_handler))
            .route("/api/agents", post(|req| register_handler(registry.clone(), nats_client.clone(), req)))
            .route("/api/agents", get(|()| list_handler(registry.clone())))
            .route("/api/agents/:id", delete(|path| unregister_handler(registry.clone(), nats_client.clone(), path)));

        let listener = TcpListener::bind(addr).await
            .map_err(|e| format!("Failed to bind: {}", e))?;

        info!(port = self.port, "Starting HTTP API server");
        axum::serve(listener, app).await
            .map_err(|e| format!("HTTP server error: {}", e))
    }
}

use uuid::Uuid;