use std::sync::Arc;
use clap::Parser;
use parking_lot::RwLock;
use std::net::SocketAddr;
use std::time::{Duration, Instant};
use axum::{Router, Json, routing::get, routing::post, routing::delete, extract::State, middleware};
use serde::{Deserialize, Serialize};
use tokio::net::TcpListener;
use tracing::{info, warn, error, Level};
use tracing_subscriber::FmtSubscriber;
use chrono::Utc;

use sar_runtime::{
    StatefulAgentRuntime, RuntimeConfig,
    AgentRepository, CachedAgentRegistry, NatsOrchestratorClient,
    Graph, Node, NodeId, NodeType, Edge, AgentId, AgentConfig,
    LifecycleManager, LifecycleStats,
};

mod auth;
mod rate_limit;
mod validation;
mod tls;

use auth::ApiKeyAuth;
use rate_limit::{RateLimiter, RateLimitConfig, RateLimitResult};
use validation::InputValidator;
use tls::TlsConfig;

const MAX_GRACE_PERIOD_SECONDS: u64 = 3600;

#[derive(Clone)]
struct AppState {
    registry: Arc<CachedAgentRegistry>,
    nats_client: Arc<RwLock<NatsOrchestratorClient>>,
    lifecycle: Arc<LifecycleManager>,
    rate_limiter: Arc<RateLimiter>,
    auth: Arc<ApiKeyAuth>,
    require_auth: bool,
}

#[derive(Parser, Debug)]
#[command(name = "functionfly-sar")]
#[command(about = "FunctionFly Stateful Agent Runtime - Production Ready")]
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

    #[arg(long)]
    require_auth: bool,

    #[arg(long, default_value = "100")]
    rate_limit_rps: u64,

    #[arg(long, default_value = "20")]
    rate_limit_burst: u64,
}

#[derive(Debug, Deserialize)]
struct RegisterAgentRequest {
    name: String,
    priority: Option<u32>,
    max_concurrent_cells: Option<u32>,
    isolation_enabled: Option<bool>,
    #[serde(default)]
    metadata: std::collections::HashMap<String, String>,
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
    code: String,
}

#[derive(Debug, Serialize)]
struct UnregisterAgentResponse {
    success: bool,
}

#[derive(Debug, Serialize)]
struct HealthResponse {
    status: String,
    version: String,
    uptime_seconds: u64,
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

static START_TIME: std::sync::OnceLock<Instant> = std::sync::OnceLock::new();

fn get_uptime_seconds() -> u64 {
    START_TIME.get().map(|t| t.elapsed().as_secs()).unwrap_or(0)
}

fn validate_request(req: &RegisterAgentRequest) -> Result<(), ErrorResponse> {
    if let Err(e) = InputValidator::validate_agent_name(&req.name) {
        return Err(ErrorResponse {
            error: e.to_string(),
            code: "INVALID_NAME".to_string(),
        });
    }

    if let Some(p) = req.priority {
        if p < 1 || p > 4 {
            return Err(ErrorResponse {
                error: "priority must be 1-4".to_string(),
                code: "INVALID_PRIORITY".to_string(),
            });
        }
    }

    if let Some(cells) = req.max_concurrent_cells {
        if cells > 10000 {
            return Err(ErrorResponse {
                error: "max_concurrent_cells exceeds 10000".to_string(),
                code: "INVALID_MAX_CELLS".to_string(),
            });
        }
    }

    if let Err(e) = InputValidator::validate_metadata(&req.metadata) {
        return Err(ErrorResponse {
            error: e.to_string(),
            code: "INVALID_METADATA".to_string(),
        });
    }

    Ok(())
}

fn check_rate_limit(state: &AppState, agent_id: Option<&str>) -> Result<(), ErrorResponse> {
    let key = agent_id.unwrap_or("global");
    match state.rate_limiter.check(key, "default") {
        RateLimitResult::Allowed => Ok(()),
        RateLimitResult::RateLimited { retry_after_ms } => Err(ErrorResponse {
            error: format!("rate limited, retry after {}ms", retry_after_ms),
            code: "RATE_LIMITED".to_string(),
        }),
        RateLimitResult::QueueFull => Err(ErrorResponse {
            error: "queue full, backpressure triggered".to_string(),
            code: "QUEUE_FULL".to_string(),
        }),
    }
}

async fn auth_middleware(
    State(state): State<AppState>,
    request: axum::http::Request<axum::body::Body>,
    next: axum::middleware::Next,
) -> Result<axum::response::Response, axum::http::StatusCode> {
    let path = request.uri().path().to_string();

    if path == "/health" || path == "/metrics" || path == "/api/health" {
        return Ok(next.run(request).await);
    }

    if !state.require_auth {
        return Ok(next.run(request).await);
    }

    let key = request
        .headers()
        .get("X-API-Key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());

    match key {
        Some(k) if state.auth.validate_key(&k).is_some() => Ok(next.run(request).await),
        _ => Err(axum::http::StatusCode::UNAUTHORIZED),
    }
}

async fn rate_limit_middleware(
    State(state): State<AppState>,
    request: axum::http::Request<axum::body::Body>,
    next: axum::middleware::Next,
) -> Result<axum::response::Response, axum::http::StatusCode> {
    match state.rate_limiter.check_global("http") {
        RateLimitResult::Allowed => Ok(next.run(request).await),
        RateLimitResult::RateLimited { retry_after_ms } => {
            let mut response = axum::response::Response::new(axum::body::Body::empty());
            *response.status_mut() = axum::http::StatusCode::TOO_MANY_REQUESTS;
            response.headers_mut().insert(
                axum::http::header::RETRY_AFTER,
                axum::http::HeaderValue::from_str(&retry_after_ms.to_string()).unwrap_or_else(|_| {
                    axum::http::HeaderValue::from_bytes(b"1000").unwrap()
                }),
            );
            Ok(response)
        }
        RateLimitResult::QueueFull => {
            Err(axum::http::StatusCode::SERVICE_UNAVAILABLE)
        }
    }
}

async fn register_agent_handler(
    State(state): State<AppState>,
    Json(req): Json<RegisterAgentRequest>,
) -> Result<Json<AgentResponse>, (axum::http::StatusCode, Json<ErrorResponse>)> {
    if let Err(e) = check_rate_limit(&state, None) {
        return Err((axum::http::StatusCode::TOO_MANY_REQUESTS, Json(e)));
    }

    if let Err(e) = validate_request(&req) {
        return Err((axum::http::StatusCode::BAD_REQUEST, Json(e)));
    }

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
        .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string(), code: "INTERNAL_ERROR".to_string() })))?;

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
    if let Err(e) = check_rate_limit(&state, None) {
        return Err((axum::http::StatusCode::TOO_MANY_REQUESTS, Json(e)));
    }

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
    if let Err(e) = check_rate_limit(&state, None) {
        return Err((axum::http::StatusCode::TOO_MANY_REQUESTS, Json(e)));
    }

    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string(), code: "INVALID_ID".to_string() })))?;

    let deleted = state.registry.unregister(&AgentId(agent_uuid)).await
        .map_err(|e| (axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(ErrorResponse { error: e.to_string(), code: "INTERNAL_ERROR".to_string() })))?;

    if deleted {
        state.lifecycle.unregister_agent(&AgentId(agent_uuid));
        let _ = state.nats_client.read().notify_agent_unregistered(&AgentId(agent_uuid));
        info!(agent_id = %agent_id, "Agent unregistered via HTTP API");
    }

    Ok(Json(UnregisterAgentResponse { success: deleted }))
}

async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        uptime_seconds: get_uptime_seconds(),
    })
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
    if let Err(e) = check_rate_limit(&state, Some(&agent_id)) {
        return Err((axum::http::StatusCode::TOO_MANY_REQUESTS, Json(e)));
    }

    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string(), code: "INVALID_ID".to_string() })))?;

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
    if let Err(e) = check_rate_limit(&state, Some(&agent_id)) {
        return Err((axum::http::StatusCode::TOO_MANY_REQUESTS, Json(e)));
    }

    let agent_uuid = uuid::Uuid::parse_str(&agent_id)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string(), code: "INVALID_ID".to_string() })))?;

    let mut grace_period = req.grace_period_seconds.unwrap_or(30);

    if grace_period > MAX_GRACE_PERIOD_SECONDS {
        warn!(requested_grace_period = grace_period, capped = MAX_GRACE_PERIOD_SECONDS, "grace period capped");
        grace_period = MAX_GRACE_PERIOD_SECONDS;
    }

    let agent_id_obj = AgentId(agent_uuid);

    let requested = state.lifecycle.request_graceful_shutdown(&agent_id_obj, grace_period);
    if !requested {
        return Err((axum::http::StatusCode::NOT_FOUND, Json(ErrorResponse { error: "agent not found".to_string(), code: "NOT_FOUND".to_string() })));
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
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, Json(ErrorResponse { error: "invalid agent_id".to_string(), code: "INVALID_ID".to_string() })))?;

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

fn setup_api_key_auth() -> ApiKeyAuth {
    if let Some(api_key) = std::env::var("SAR_API_KEY").ok() {
        let auth = ApiKeyAuth::new();
        auth.add_key(api_key, "env-api-key".to_string(), false);
        info!("API key configured from SAR_API_KEY environment variable");
        return auth;
    }

    if let Some(admin_key) = std::env::var("SAR_ADMIN_API_KEY").ok() {
        let auth = ApiKeyAuth::new();
        auth.add_key(admin_key, "env-admin-key".to_string(), true);
        info!("Admin API key configured from SAR_ADMIN_API_KEY environment variable");
        return auth;
    }

    if std::env::var("SAR_REQUIRE_AUTH").is_err() {
        warn!("No API key configured and SAR_REQUIRE_AUTH not set - running WITHOUT AUTHENTICATION");
        return ApiKeyAuth::new();
    }

    ApiKeyAuth::from_env()
}

fn setup_tls() -> Option<TlsConfig> {
    TlsConfig::from_env().filter(|c| c.enabled()).map(|c| {
        info!(cert = %c.cert_path, "TLS enabled");
        c
    })
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let _ = START_TIME.set(Instant::now());

    let _subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .with_target(true)
        .with_thread_ids(true)
        .json()
        .init();

    let args = Args::parse();

    info!(version = env!("CARGO_PKG_VERSION"), "Starting FunctionFly SAR - Production Ready");

    let auth = Arc::new(setup_api_key_auth());
    let rate_limiter = Arc::new(RateLimiter::new(
        RateLimitConfig {
            requests_per_second: args.rate_limit_rps,
            burst_size: args.rate_limit_burst,
            max_queue_size: 10000,
        },
        RateLimitConfig {
            requests_per_second: 100,
            burst_size: 20,
            max_queue_size: 1000,
        },
    ));

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
            info!(nats_url = %url, "Connected to NATS");
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
        rate_limiter: rate_limiter.clone(),
        auth: auth.clone(),
        require_auth: args.require_auth,
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
        .layer(middleware::from_fn_with_state(state.clone(), |state, req, next| auth_middleware(state, req, next)))
        .layer(middleware::from_fn_with_state(state.clone(), |state, req, next| rate_limit_middleware(state, req, next)))
        .with_state(state);

    let addr: SocketAddr = format!("0.0.0.0:{}", args.api_port).parse()?;
    let listener = TcpListener::bind(addr).await?;

    info!(
        api_port = args.api_port,
        agents_loaded = active_agents,
        require_auth = args.require_auth,
        rate_limit_rps = args.rate_limit_rps,
        "SAR runtime initialized - PRODUCTION READY"
    );

    axum::serve(listener, app).await?;

    info!("SAR runtime shutdown complete");
    Ok(())
}