use std::sync::Arc;
use clap::Parser;
use parking_lot::RwLock;
use std::net::SocketAddr;
use axum::{Router, Json, routing::get, routing::post, routing::delete, extract::State};
use serde::{Deserialize, Serialize};
use tokio::net::TcpListener;
use tracing::{info, warn, Level};
use tracing_subscriber::FmtSubscriber;
use chrono::Utc;

use sar_runtime::{
    StatefulAgentRuntime, RuntimeConfig,
    AgentRepository, CachedAgentRegistry, NatsOrchestratorClient,
    Graph, Node, NodeId, NodeType, Edge, AgentId, AgentConfig,
    LifecycleManager, LifecycleStats,
};

#[derive(Clone)]
struct AppState {
    registry: Arc<CachedAgentRegistry>,
    nats_client: Arc<RwLock<NatsOrchestratorClient>>,
    lifecycle: Arc<LifecycleManager>,
}

#[derive(Parser, Debug)]
#[command(name = "functionfly-sar")]
#[command(about = "FunctionFly Stateful Agent Runtime")]
struct Args {
    #[arg(long, env = "NATS_URL")]
    nats_url: Option<String>,

    #[arg(long, env = "REDIS_URL")]
    redis_url: Option<String>,

    #[arg(long, env = "DATABASE_URL")]
    database_url: Option<String>,

    #[arg(long, default_value = "10000")]
    max_concurrent: usize,

    #[arg(long, default_value = "8082")]
    api_port: u16,
}

#[derive(Debug, Deserialize)]
struct RegisterAgentRequest {
    name: String,
    priority: Option<u32>,
    max_concurrent_cells: Option<u32>,
    isolation_enabled: Option<bool>,
}

#[derive(Debug, Serialize)]
struct AgentResponse {
    id: String,
    name: String,
    status: String,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Debug, Serialize)]
struct UnregisterAgentResponse {
    success: bool,
}

#[derive(Debug, Serialize)]
struct HealthResponse {
    status: String,
}

#[derive(Debug, Serialize)]
struct LifecycleStatsResponse {
    total_agents: usize,
    alive_agents: usize,
    orphaned_agents: usize,
    shutting_down_agents: usize,
}

#[derive(Debug, Deserialize)]
struct HeartbeatRequest {
    status: Option<String>,
    state_snapshot: Option<std::collections::HashMap<String, String>>,
}

#[derive(Debug, Deserialize)]
struct ShutdownRequest {
    grace_period_seconds: Option<u64>,
}

async fn register_agent_handler(
    State(state): State<AppState>,
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

    let persistence = state.registry.register(agent_config).await
        .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string() })))?;

    state.lifecycle.register_agent(persistence.id);

    let _ = state.nats_client.read().notify_agent_registered(persistence.id, &persistence.name);

    info!(agent_id = %persistence.id, name = %persistence.name, "Agent registered via HTTP API");

    Ok(Json(AgentResponse {
        id: persistence.id.0.to_string(),
        name: persistence.name,
        status: format!("{:?}", persistence.status),
    }))
}

async fn list_agents_handler(
    State(state): State<AppState>,
) -> Result<Json<Vec<AgentResponse>>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    let agents = state.registry.list_cached();
    Ok(Json(agents.iter().map(|a| AgentResponse {
        id: a.id.0.to_string(),
        name: a.name.clone(),
        status: format!("{:?}", a.status),
    }).collect()))
}

async fn unregister_agent_handler(
    State(state): State<AppState>,
    axum::extract::Path(agent_id): axum::extract::Path<String>,
) -> Result<Json<UnregisterAgentResponse>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string() })))?;

    let deleted = state.registry.unregister(&AgentId(agent_uuid)).await
        .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string() })))?;

    if deleted {
        state.lifecycle.unregister_agent(&AgentId(agent_uuid));
        let _ = state.nats_client.read().notify_agent_unregistered(&AgentId(agent_uuid));
        info!(agent_id = %agent_id, "Agent unregistered via HTTP API");
    }

    Ok(Json(UnregisterAgentResponse { success: deleted }))
}

async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse { status: "healthy".to_string() })
}

async fn lifecycle_stats_handler(
    State(state): State<AppState>,
) -> Json<LifecycleStatsResponse> {
    let stats = state.lifecycle.get_stats();
    Json(LifecycleStatsResponse {
        total_agents: stats.total_agents,
        alive_agents: stats.alive_agents,
        orphaned_agents: stats.orphaned_agents,
        shutting_down_agents: stats.shutting_down_agents,
    })
}

async fn heartbeat_handler(
    State(state): State<AppState>,
    axum::extract::Path(agent_id): axum::extract::Path<String>,
    Json(req): Json<HeartbeatRequest>,
) -> Result<Json<serde_json::Value>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string() })))?;

    let agent_id_obj = AgentId(agent_uuid);
    state.lifecycle.update_heartbeat(&agent_id_obj);

    if let Some(snapshot) = req.state_snapshot {
        state.lifecycle.save_state_snapshot(&agent_id_obj, snapshot);
    }

    let next_heartbeat = Utc::now() + chrono::Duration::seconds(30);

    Ok(Json(serde_json::json!({
        "ok": true,
        "next_heartbeat": next_heartbeat.to_rfc3339()
    })))
}

async fn shutdown_handler(
    State(state): State<AppState>,
    axum::extract::Path(agent_id): axum::extract::Path<String>,
    Json(req): Json<ShutdownRequest>,
) -> Result<Json<serde_json::Value>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string() })))?;

    let grace_period = req.grace_period_seconds.unwrap_or(30);
    let agent_id_obj = AgentId(agent_uuid);

    let requested = state.lifecycle.request_graceful_shutdown(&agent_id_obj, grace_period);
    if !requested {
        return Err((axum::http::StatusCode::NOT_FOUND, Json(ErrorResponse { error: "agent not found".to_string() })));
    }

    Ok(Json(serde_json::json!({
        "ok": true,
        "message": "graceful shutdown initiated",
        "grace_period_seconds": grace_period
    })))
}

async fn check_shutdown_handler(
    State(state): State<AppState>,
    axum::extract::Path(agent_id): axum::extract::Path<String>,
) -> Result<Json<serde_json::Value>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string() })))?;

    let agent_id_obj = AgentId(agent_uuid);
    let complete = state.lifecycle.check_shutdown_complete(&agent_id_obj);

    Ok(Json(serde_json::json!({
        "ok": true,
        "shutdown_complete": complete
    })))
}

fn create_in_memory_repository() -> Arc<AgentRepository> {
    Arc::new(AgentRepository::new_in_memory())
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let _subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .with_target(true)
        .with_thread_ids(true)
        .json()
        .init();

    let args = Args::parse();

    info!(version = env!("CARGO_PKG_VERSION"), "Starting FunctionFly SAR");

    let mut config = RuntimeConfig::default();

    if let Some(ref url) = args.nats_url {
        config.nats_url = Some(url.clone());
    }
    if let Some(ref url) = args.redis_url {
        config.redis_url = Some(url.clone());
    }
    if let Some(ref url) = args.database_url {
        config.postgres_url = Some(url.clone());
    }
    config.scheduler.max_concurrent = args.max_concurrent;

    let runtime = Arc::new(StatefulAgentRuntime::with_config(config).await?);
    let _ = runtime;

    let repository: Arc<AgentRepository> = if let Some(ref db_url) = args.database_url {
        match AgentRepository::new(db_url).await {
            Ok(repo) => {
                if let Err(e) = repo.init_schema().await {
                    warn!(error = %e, "Failed to initialize persistence schema");
                }
                Arc::new(repo)
            }
            Err(e) => {
                warn!(error = %e, "Failed to connect to database, using in-memory store");
                create_in_memory_repository()
            }
        }
    } else {
        warn!("No DATABASE_URL provided, using in-memory store");
        create_in_memory_repository()
    };

    let registry = Arc::new(CachedAgentRegistry::new(repository));
    if let Err(e) = registry.load_all().await {
        warn!(error = %e, "Failed to load agents from persistence");
    }

    let nats_client = Arc::new(RwLock::new(NatsOrchestratorClient::new()));
    if let Some(ref url) = args.nats_url {
        if let Err(e) = nats_client.write().connect(url) {
            warn!(error = %e, "Failed to connect to NATS");
        } else {
            let _ = nats_client.write().start_event_listener(|_event| {});
        }
    }

    let active_agents = registry.list_cached().len() as u32;
    let _ = nats_client.read().report_runtime_status(true, active_agents);

    let lifecycle_manager = Arc::new(LifecycleManager::new(registry.clone()));
    for agent in registry.list_cached() {
        lifecycle_manager.register_agent(agent.id);
    }

    let state = AppState {
        registry,
        nats_client,
        lifecycle: lifecycle_manager,
    };

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/api/agents", post(register_agent_handler))
        .route("/api/agents", get(list_agents_handler))
        .route("/api/agents/{id}", delete(unregister_agent_handler))
        .route("/api/lifecycle/stats", get(lifecycle_stats_handler))
        .route("/api/agents/{id}/heartbeat", post(heartbeat_handler))
        .route("/api/agents/{id}/shutdown", post(shutdown_handler))
        .route("/api/agents/{id}/shutdown/check", get(check_shutdown_handler))
        .with_state(state);

    let addr: SocketAddr = format!("0.0.0.0:{}", args.api_port).parse()?;
    let listener = TcpListener::bind(addr).await?;

    info!(
        api_port = args.api_port,
        agents_loaded = active_agents,
        "SAR runtime initialized with full agent management"
    );

    axum::serve(listener, app).await?;

    info!("SAR runtime shutdown complete");
    Ok(())
}