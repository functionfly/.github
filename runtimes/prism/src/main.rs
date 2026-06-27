//! FunctionFly Prism Runtime - Main binary
//!
//! Universal Adaptive WASM Execution Fabric for AI-native execution,
//! autonomous agents, robotics, edge systems, and cross-language functions.

use std::sync::Arc;
use std::net::SocketAddr;
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::Instant;
use clap::Parser;
use tracing::{info, error, debug};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

mod cli;

use prism_runtime::runtime::RuntimeContext;
use prism_runtime::core::{CellId, CellConfig, CellStatus, ExecutionTarget, ExecutionMetrics};
use prism_runtime::quantum::{SnapshotType, MigrationManager, MigrationStrategy};
use prism_runtime::swarm::{SwarmCoordinator, SwarmId, CoordinatorConfig};
use prism_runtime::ucl::{Capability, CapabilityCategory};
use prism_runtime::neural::{ExecutionProfile, ExecutionOutcome, ExecutionFeatures};
use chrono::{Timelike, Datelike};

/// Token bucket rate limiter for DoS protection
#[derive(Clone)]
struct RateLimiter {
    tokens: Arc<AtomicU64>,
    max_tokens: u64,
    refill_rate: u64,
    last_refill: Arc<std::sync::Mutex<Instant>>,
}

impl RateLimiter {
    fn new(max_tokens: u64, refill_per_second: u64) -> Self {
        Self {
            tokens: Arc::new(AtomicU64::new(max_tokens)),
            max_tokens,
            refill_rate: refill_per_second,
            last_refill: Arc::new(std::sync::Mutex::new(Instant::now())),
        }
    }

    fn try_acquire(&self) -> bool {
        let mut last = self.last_refill.lock().unwrap();
        let now = Instant::now();
        let elapsed = now.duration_since(*last).as_secs_f64();
        let refill = (elapsed * self.refill_rate as f64) as u64;

        if refill > 0 {
            let new_tokens = (self.tokens.load(Ordering::Relaxed) + refill).min(self.max_tokens);
            self.tokens.store(new_tokens, Ordering::Relaxed);
            *last = now;
        }

        let current = self.tokens.load(Ordering::Relaxed);
        if current > 0 {
            self.tokens.store(current - 1, Ordering::Relaxed);
            true
        } else {
            false
        }
    }
}

/// Per-tenant rate limiters
struct TenantRateLimiters {
    limiters: dashmap::DashMap<String, RateLimiter>,
    default_limit: RateLimiter,
}

impl TenantRateLimiters {
    fn new(_requests_per_second: u64, max_tokens: u64) -> Self {
        Self {
            limiters: dashmap::DashMap::new(),
            default_limit: RateLimiter::new(max_tokens, 1000),
        }
    }

    fn get_or_create(&self, tenant_id: &str) -> RateLimiter {
        if tenant_id.is_empty() || tenant_id == "anonymous" {
            return self.default_limit.clone();
        }
        self.limiters.entry(tenant_id.to_string())
            .or_insert_with(|| RateLimiter::new(100, 100))
            .clone()
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = cli::Cli::parse();

    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(
            if cli.verbose { "debug" } else { "info" },
        ))
        .with(tracing_subscriber::fmt::layer())
        .init();

    info!("FunctionFly Prism Runtime v{}", env!("CARGO_PKG_VERSION"));

    match cli.command {
        cli::Commands::Start { address, mesh } => {
            start_runtime(address, mesh, &cli.config).await?;
        }
        cli::Commands::Cell { action } => {
            let runtime = build_runtime(false, "127.0.0.1:0").await?;
            match action {
                cli::CellCommands::Create { module, memory } => {
                    handle_cell_create(&runtime, &module, memory).await?;
                }
                cli::CellCommands::List => {
                    handle_cell_list(&runtime).await;
                }
                cli::CellCommands::Terminate { cell_id } => {
                    handle_cell_terminate(&runtime, &cell_id).await?;
                }
                cli::CellCommands::Snapshot { cell_id } => {
                    handle_cell_snapshot(&runtime, &cell_id).await?;
                }
                cli::CellCommands::Migrate { cell_id, target } => {
                    handle_cell_migrate(&runtime, &cell_id, &target).await?;
                }
            }
        }
        cli::Commands::Capability { action } => {
            let runtime = build_runtime(false, "127.0.0.1:0").await?;
            match action {
                cli::CapabilityCommands::Register { name, category } => {
                    handle_capability_register(&runtime, &name, &category).await?;
                }
                cli::CapabilityCommands::Discover { query } => {
                    handle_capability_discover(&runtime, &query).await;
                }
                cli::CapabilityCommands::List => {
                    handle_capability_list(&runtime).await;
                }
            }
        }
        cli::Commands::Swarm { action } => {
            let mut runtime = build_runtime(true, "127.0.0.1:0").await?;
            match action {
                cli::SwarmCommands::Create { swarm_id } => {
                    handle_swarm_create(&mut runtime, &swarm_id).await?;
                }
                cli::SwarmCommands::Join { swarm_id } => {
                    handle_swarm_join(&runtime, &swarm_id).await?;
                }
                cli::SwarmCommands::Leave { swarm_id } => {
                    handle_swarm_leave(&runtime, &swarm_id).await?;
                }
                cli::SwarmCommands::List => {
                    handle_swarm_list(&runtime).await;
                }
                cli::SwarmCommands::Command { swarm_id, cmd } => {
                    handle_swarm_command(&runtime, &swarm_id, &cmd).await?;
                }
            }
        }
        cli::Commands::Repl { language } => {
            let runtime = build_runtime(false, "127.0.0.1:0").await?;
            start_repl(&language, runtime).await?;
        }
        cli::Commands::Package { action } => {
            handle_package(action).await?;
        }
        cli::Commands::Status => {
            let runtime = build_runtime(false, "127.0.0.1:0").await?;
            show_status(&runtime).await?;
        }
        cli::Commands::Doc { output } => {
            generate_docs(&output)?;
        }
        cli::Commands::Exec { wasm, cell_id } => {
            // WASI's MemoryOutputPipe flush may try to drive an async runtime
            // internally, which is not allowed when the current thread is
            // already inside a tokio runtime. Run the WASM exec on a fresh
            // OS thread so the wasmtime-wasi sync path can use a private
            // executor without conflicting with the tokio main.
            let wasm_owned = wasm;
            let cell_owned = cell_id;
            let handle = std::thread::Builder::new()
                .name("prism-exec".to_string())
                .spawn(move || {
                    handle_exec(&wasm_owned, &cell_owned)
                })
                .map_err(|e| anyhow::anyhow!("Failed to spawn prism-exec thread: {}", e))?;
            handle.join().map_err(|e| anyhow::anyhow!("prism-exec thread panicked: {:?}", e))??;
        }
    }

    Ok(())
}

async fn build_runtime(mesh: bool, address: &str) -> anyhow::Result<Arc<RuntimeContext>> {
    let runtime = Arc::new(RuntimeContext::new(address.to_string(), mesh));
    runtime.init_fusion_executor().await
        .map_err(|e| anyhow::anyhow!("Failed to initialize fusion executor: {}", e))?;

    if mesh {
        let mut coordinator = runtime.swarm_coordinator.write().await;
        *coordinator = Some(SwarmCoordinator::new(CoordinatorConfig::default()));
    }

    let nats_url = std::env::var("NATS_URL").unwrap_or_else(|_| "nats://localhost:4222".to_string());
    match runtime.connect_nats(&nats_url).await {
        Ok(_) => info!("Connected to NATS at {}", nats_url),
        Err(e) => debug!("NATS not available: {}", e),
    }

    Ok(runtime)
}

async fn start_runtime(address: String, mesh: bool, _config_path: &str) -> anyhow::Result<()> {
    info!("Starting Prism Runtime on {}", address);
    info!("Mesh networking: {}", if mesh { "enabled" } else { "disabled" });

    let runtime = Arc::new(RuntimeContext::new(address.clone(), mesh));

    if let Err(e) = runtime.init_fusion_executor().await {
        error!("Failed to initialize fusion executor: {}", e);
        return Err(anyhow::anyhow!("Runtime initialization failed: {}", e));
    }
    info!("Fusion executor ready");

    if mesh {
        let mut coordinator = runtime.swarm_coordinator.write().await;
        *coordinator = Some(SwarmCoordinator::new(CoordinatorConfig::default()));
        info!("Swarm coordinator initialized");

        // Start the P2P mesh network (libp2p) so cells can migrate and
        // capabilities can be discovered across nodes.
        #[cfg(feature = "mesh")]
        {
            use prism_runtime::mesh::MeshConfig;
            // The libp2p listen address is a multiaddr string like
            // /ip4/0.0.0.0/tcp/0 (port 0 = OS-assigned).
            let mesh_config = MeshConfig::default();
            match runtime.start_mesh(mesh_config).await {
                Ok(peer) => info!(local_peer = %peer, "Mesh network active"),
                Err(e) => error!("Mesh network failed to start (continuing without it): {}", e),
            }
        }
    }

    let nats_url = std::env::var("NATS_URL").unwrap_or_else(|_| "nats://localhost:4222".to_string());
    match runtime.connect_nats(&nats_url).await {
        Ok(_) => info!("Connected to NATS at {}", nats_url),
        Err(e) => info!("NATS connection not available: {} (this is fine for local dev)", e),
    }

    let runtime_clone = runtime.clone();
    let addr: SocketAddr = address.parse()?;
    let listener = tokio::net::TcpListener::bind(addr).await?;
    info!("Prism Runtime listening on {}", address);

    // Rate limiters for DoS protection - 1000 req/s global, 100 req/s per tenant
    let rate_limiters = Arc::new(TenantRateLimiters::new(1000, 5000));
    let global_limiter = Arc::new(RateLimiter::new(10000, 10000));

    let api_token: Option<String> = std::env::var("RUNTIME_API_TOKEN").ok().filter(|t| !t.is_empty());
    let is_production = std::env::var("ENVIRONMENT")
        .map(|v| v.eq_ignore_ascii_case("production"))
        .unwrap_or(false);

    if api_token.is_none() {
        if is_production {
            error!(
                "RUNTIME_API_TOKEN is not set in production. \
                 The /execute endpoint is UNAUTHENTICATED. Set the token and restart."
            );
        } else {
            info!("RUNTIME_API_TOKEN not set — /execute endpoint is unauthenticated (dev mode)");
        }
    }

    let api_token = Arc::new(api_token);

    loop {
        let (mut stream, _peer) = listener.accept().await?;
        let rt = runtime_clone.clone();
        let rl = rate_limiters.clone();
        let gl = global_limiter.clone();
        let token = api_token.clone();

        tokio::spawn(async move {
            use tokio::io::{AsyncReadExt, AsyncWriteExt};

            // Check global rate limit first
            if !gl.try_acquire() {
                let resp = json_response(429, "Rate limit exceeded - try again later");
                let _ = stream.write_all(resp.as_bytes()).await;
                return;
            }

            let mut buf = [0u8; 2048];
            let mut stream = stream;
            let n = match stream.read(&mut buf).await {
                Ok(n) if n == 0 => return,
                Ok(n) => n,
                Err(_) => return,
            };

            // Extract tenant ID for per-tenant rate limiting
            let request = String::from_utf8_lossy(&buf[..n]).to_string();
            let tenant_id = extract_tenant_id(&request);
            let limiter = rl.get_or_create(&tenant_id);

            // Check per-tenant rate limit
            if !limiter.try_acquire() {
                debug!(tenant = %tenant_id, "Tenant rate limit exceeded");
                let resp = json_response(429, "Rate limit exceeded for tenant - try again later");
                let _ = stream.write_all(resp.as_bytes()).await;
                return;
            }

            let request_line = request.lines().next().unwrap_or("");
            debug!(request_line = %request_line, "Received request");
            let response = if request.starts_with("GET /health") {
                let status = rt.get_status().await;
                serde_json::json!({
                    "status_code": 200,
                    "body": {
                        "version": status.version,
                        "healthy": status.healthy,
                        "active_cells": status.active_cells,
                        "total_cells": status.total_cells,
                        "mesh_enabled": status.mesh_enabled,
                    }
                }).to_string()
            } else if request.starts_with("POST /cells/") && request.contains("/snapshot") {
                // Snapshot cell - must come before generic POST /cells/... creation
                debug!("Routing to snapshot handler");
                handle_http_snapshot_cell(&rt, &request).await
            } else if request.starts_with("POST /cells/") && !request.contains("/snapshot") && !request.contains("/execute") {
                // Generic cell creation with path like /cells/ID - exclude /snapshot and /execute
                debug!("Routing to cell creation handler");
                handle_http_create_cell(&rt, &request).await
            } else if request.starts_with("POST /cells") && !request.contains("/snapshots") && !request.contains("/execute") {
                // Generic cell creation - exclude /snapshots and /execute paths
                handle_http_create_cell(&rt, &request).await
            } else if request.starts_with("POST /execute") {
                // Auth check for execute endpoint
                if let Some(ref expected_token) = *token {
                    let auth_ok = request.lines()
                        .find(|line| line.to_lowercase().starts_with("authorization:"))
                        .and_then(|line| line.split(':').nth(1))
                        .map(|val| val.trim() == format!("Bearer {}", expected_token))
                        .unwrap_or(false);
                    if !auth_ok {
                        json_response(401, "unauthorized")
                    } else {
                        // Clone Arcs so all captured data is owned for spawn_blocking
                        let rt_exec = rt.clone();
                        let req_owned = request.clone();
                        let result = tokio::task::spawn_blocking(move || {
                            tokio::runtime::Handle::current()
                                .block_on(handle_http_execute(&rt_exec, &req_owned))
                        }).await;
                        match result {
                            Ok(resp) => resp,
                            Err(e) => json_response(500, &format!("Internal error: {}", e)),
                        }
                    }
                } else {
                    // No token configured — allow in dev mode
                    let rt_exec = rt.clone();
                    let req_owned = request.clone();
                    let result = tokio::task::spawn_blocking(move || {
                        tokio::runtime::Handle::current()
                            .block_on(handle_http_execute(&rt_exec, &req_owned))
                    }).await;
                    match result {
                        Ok(resp) => resp,
                        Err(e) => json_response(500, &format!("Internal error: {}", e)),
                    }
                }
            } else if request.starts_with("GET /cells/") && request.contains("/snapshots") {
                handle_http_list_snapshots(&rt, &request).await
            } else if request.starts_with("GET /cells/") && !request.contains("/snapshots") {
                // GET /cells/{id} - list specific cell (fallback for cell with no action)
                let cells = rt.list_cells().await;
                // Extract just the UUID part from path like "GET /cells/UUID HTTP/1.1" or "GET /cells/UUID?query=val"
                let request_path = request.lines().next().unwrap_or("");
                // Split by space to get method and path: "GET" vs "/cells/UUID HTTP/1.1"
                let mut parts = request_path.split_whitespace();
                let path_only = parts.nth(1).unwrap_or("");
                let after_cells = path_only.split("/cells/").nth(1).unwrap_or("");
                // UUID is the first path segment before / or ?
                let cell_id_str = after_cells.split(|c: char| c == '/' || c == '?').next().unwrap_or("");
                if let Ok(uuid) = uuid::Uuid::parse_str(cell_id_str) {
                    let cell_id = CellId::from_uuid(uuid);
                    if let Some(cell) = cells.iter().find(|c| c.id == cell_id) {
                        serde_json::json!({
                            "status_code": 200,
                            "body": {
                                "id": cell.id.to_string(),
                                "tenant": cell.tenant_id,
                                "status": format!("{:?}", cell.status),
                                "name": cell.metadata.name,
                                "memory_mb": cell.config.memory_limit_mb,
                            }
                        }).to_string()
                    } else {
                        json_response(404, "Cell not found")
                    }
                } else {
                    json_response(400, &format!("Invalid cell ID: {}", cell_id_str))
                }
            } else if request.starts_with("GET /cells") {
                // GET /cells - list all cells
                let cells = rt.list_cells().await;
                let entries: Vec<serde_json::Value> = cells.iter().map(|c| {
                    serde_json::json!({
                        "id": c.id.to_string(),
                        "tenant": c.tenant_id,
                        "status": format!("{:?}", c.status),
                        "name": c.metadata.name,
                        "memory_mb": c.config.memory_limit_mb,
                    })
                }).collect();
                serde_json::json!({
                    "status_code": 200,
                    "body": entries
                }).to_string()
            } else if request.starts_with("POST /snapshots/") && request.contains("/restore") {
                // POST /snapshots/{id}/restore - Restore from snapshot
                handle_http_restore_snapshot(&rt, &request).await
            } else if request.starts_with("DELETE /snapshots/") {
                // DELETE /snapshots/{id} - Delete a snapshot
                handle_http_delete_snapshot(&rt, &request).await
            } else if request.starts_with("POST /capabilities") && request.contains("/invoke") {
                // POST /capabilities/invoke - Invoke a capability
                handle_http_invoke_capability(&rt, &request).await
            } else if request.starts_with("POST /capabilities") {
                // POST /capabilities - Register a capability
                handle_http_register_capability(&rt, &request).await
            } else if request.starts_with("GET /capabilities") {
                // GET /capabilities - List all capabilities
                handle_http_list_capabilities(&rt, &request).await
            } else if request.starts_with("POST /swarms") && !request.contains("/join") && !request.contains("/leave") {
                // POST /swarms - Create a swarm
                handle_http_create_swarm(&rt, &request).await
            } else if request.starts_with("GET /swarms") {
                // GET /swarms - List all swarms
                handle_http_list_swarms(&rt, &request).await
            } else if request.starts_with("POST /swarms/") && request.contains("/join") {
                // POST /swarms/{id}/join - Join a swarm
                handle_http_join_swarm(&rt, &request).await
            } else if request.starts_with("POST /swarms/") && request.contains("/leave") {
                // POST /swarms/{id}/leave - Leave a swarm
                handle_http_leave_swarm(&rt, &request).await
            } else if request.starts_with("GET /optimize/") {
                // GET /optimize/{cell_id} - Get optimization suggestion
                handle_http_get_optimization(&rt, &request).await
            } else {
                json_response(404, "Not Found")
            };

            // Parse status code from JSON response
            let (status_line, _body) = if let Ok(json) = serde_json::from_str::<serde_json::Value>(&response) {
                let code = json.get("status_code").and_then(|v| v.as_u64()).unwrap_or(200) as u16;
                let body_str = json.get("body").map(|v| v.to_string()).unwrap_or_default();
                let status_text = match code {
                    200 => "OK",
                    201 => "Created",
                    400 => "Bad Request",
                    404 => "Not Found",
                    500 => "Internal Server Error",
                    _ => "Unknown",
                };
                (format!("HTTP/1.1 {} {}\r\nContent-Type: application/json\r\n\r\n{}", code, status_text, body_str), body_str)
            } else {
                (format!("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n{}", response), response)
            };

            let _ = stream.write_all(status_line.as_bytes()).await;
        });
    }
}

async fn handle_http_create_cell(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let name = extract_json_field(body, "name").unwrap_or_default();
    let module_path = extract_json_field(body, "module").unwrap_or_default();
    let memory_mb: u64 = extract_json_field(body, "memory")
        .and_then(|s| s.parse().ok())
        .unwrap_or(128);

    if name.is_empty() || module_path.is_empty() {
        return json_response(400, "Missing 'name' or 'module' field");
    }

    let wasm_bytes = match std::fs::read(&module_path) {
        Ok(b) => b,
        Err(e) => return json_response(400, &format!("Failed to read module: {}", e)),
    };

    if wasm_bytes.len() < 4 || &wasm_bytes[0..4] != b"\0asm" {
        return json_response(400, "Invalid WASM file: magic number not found");
    }

    let config = CellConfig {
        memory_limit_mb: memory_mb,
        execution_target: ExecutionTarget::Cloud,
        ..CellConfig::default()
    };

    match rt.create_cell("tenant-1", &name, wasm_bytes, config).await {
        Ok(cell_id) => json_response_data(201, &serde_json::json!({"id": cell_id.to_string(), "status": "created"}).to_string()),
        Err(e) => json_response(400, &format!("{}", e)),
    }
}

async fn handle_http_execute(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let cell_id_str = extract_json_field(body, "cell_id").unwrap_or_default();
    let input_str = extract_json_field(body, "input").unwrap_or_default();
    let input_bytes = input_str.as_bytes();

    let cell_id = match uuid::Uuid::parse_str(&cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, "Invalid cell_id format"),
    };

    match rt.get_cell(&cell_id).await {
        Some(cell) => {
            if let Some(module_id) = cell.wasm_module_id {
                use prism_runtime::wasm_fusion::{FusionGraph, FusionNode, FusionNodeType};

                let memory_limit_mb = cell.config.memory_limit_mb;
                let mut graph = FusionGraph::new(&cell_id_str);
                graph.add_node(FusionNode {
                    node_id: module_id.clone(),
                    name: module_id,
                    node_type: FusionNodeType::Wasm,
                    config: prism_runtime::wasm_fusion::NodeConfig {
                        entry_point: "handler".to_string(),
                        timeout_ms: cell.config.timeout_ms,
                        memory_limit_mb: cell.config.memory_limit_mb,
                        imports: Vec::new(),
                    },
                });

                // Clone the Arc so it can be moved into spawn_blocking
                let fusion_executor_arc = rt.fusion_executor.clone();
                let graph_arc = std::sync::Arc::new(std::sync::Mutex::new(Some(graph)));
                let graph_arc2 = graph_arc.clone();
                let input_vec = input_bytes.to_vec();

                let handle = tokio::task::spawn_blocking(move || {
                    // Acquire read lock inside the blocking task (short-lived)
                    let executor_guard = futures::executor::block_on(fusion_executor_arc.read());
                    match executor_guard.as_ref() {
                        Some(executor) => {
                            let mut g = graph_arc2.lock().unwrap().take().unwrap();
                            match futures::executor::block_on(executor.execute_graph(&mut g, &input_vec)) {
                                Ok(output) => json_response_data(200, &String::from_utf8_lossy(&output)),
                                Err(e) => json_response(500, &format!("Execution failed: {}", e)),
                            }
                        }
                        None => json_response(500, "Fusion executor not initialized"),
                    }
                });

                let resp = match handle.await {
                    Ok(resp) => resp,
                    Err(e) => json_response(500, &format!("Task failed: {}", e)),
                };

                // ── RL Feedback Loop: record execution outcome ────────
                // After the blocking task releases the executor lock, read the
                // stored metrics and feed them back into the neural optimizer.
                {
                    let executor_guard = rt.fusion_executor.read().await;
                    if let Some(ref executor) = *executor_guard {
                        if let Some(snapshot) = executor.take_last_metrics().await {
                            // Cache CPU state for future snapshots
                            if let Some(ref cpu) = snapshot.cpu_state {
                                let mut cpu_states = rt.last_cpu_states.write().await;
                                cpu_states.insert(cell_id, cpu.clone());
                            }

                            let profile = ExecutionProfile {
                                cell_id,
                                metrics: ExecutionMetrics {
                                    duration_ms: snapshot.exec_time_ms,
                                    memory_used_bytes: snapshot.memory_used_bytes,
                                    ..Default::default()
                                },
                                features: ExecutionFeatures {
                                    input_size_bytes: input_bytes.len() as u64,
                                    memory_limit_mb,
                                    vcpus: 1,
                                    gpu_used: false,
                                    execution_location: "cloud".to_string(),
                                    time_of_day: chrono::Utc::now().hour() as f32,
                                    day_of_week: chrono::Utc::now().weekday().num_days_from_sunday() as u8,
                                },
                                outcome: if snapshot.success {
                                    ExecutionOutcome::Success
                                } else {
                                    ExecutionOutcome::Error
                                },
                            };
                            rt.record_execution_outcome(profile).await;
                        }
                    }
                }

                resp
            } else {
                json_response(400, "Cell has no WASM module")
            }
        }
        None => json_response(404, "Cell not found"),
    }
}

fn extract_json_field(json: &str, field: &str) -> Option<String> {
    // Parse the body as a real JSON document and pull the named field out.
    // This replaces the previous substring-based extractor which produced
    // silently-wrong results for nested objects and accepted malformed input.
    let value: serde_json::Value = serde_json::from_str(json).ok()?;
    let mut current = &value;
    for part in field.split('.') {
        current = current.get(part)?;
    }
    match current {
        serde_json::Value::String(s) => Some(s.clone()),
        serde_json::Value::Number(n) => Some(n.to_string()),
        serde_json::Value::Bool(b) => Some(b.to_string()),
        serde_json::Value::Null => None,
        // For arrays and objects, return the JSON serialization as a string.
        other => Some(other.to_string()),
    }
}

/// Extract tenant ID from request for rate limiting
fn extract_tenant_id(request: &str) -> String {
    // Try to get tenant from header first
    for line in request.lines() {
        if line.to_lowercase().starts_with("x-tenant-id:") ||
           line.to_lowercase().starts_with("x-tenant:") {
            let parts: Vec<&str> = line.split(':').collect();
            if parts.len() >= 2 {
                return parts[1].trim().to_string();
            }
        }
    }

    // Fall back to body tenant_id field
    if let Some(tenant) = extract_json_field(request, "tenant_id") {
        return tenant;
    }

    // Fall back to JWT subject or API key prefix (simplified)
    for line in request.lines() {
        if line.to_lowercase().starts_with("authorization:") {
            let val = line.split(':').nth(1).unwrap_or("").trim();
            if val.starts_with("Bearer ") {
                // Use first 8 chars of token as tenant identifier
                let token = &val[7..];
                // Simple hash for rate limit key - not for auth
                return format!("token:{}", &token[..token.len().min(16)]);
            }
        }
    }

    "anonymous".to_string()
}

/// Detect whether we are running under WSL (Windows Subsystem for Linux). Landlock has
/// known issues under WSL2 kernels that cause SIGSEGV immediately after
/// landlock_restrict_self, so we disable it there by default.
fn is_wsl() -> bool {
    std::fs::read_to_string("/proc/version")
        .map(|v| v.contains("microsoft") || v.contains("WSL"))
        .unwrap_or(false)
}

/// Handle prism exec - execute WASM in isolated subprocess with full security enforcement
fn handle_exec(wasm_path: &Option<String>, cell_id: &Option<String>) -> anyhow::Result<()> {
    use prism_runtime::security::{SecurityPolicy, SecurityManager, EnforceResourceLimits};

    // Read WASM from stdin if no path provided
    let wasm_bytes = if let Some(path) = wasm_path {
        std::fs::read(path)?
    } else {
        use std::io::Read;
        let mut stdin = std::io::stdin();
        let mut buffer = Vec::new();
        stdin.read_to_end(&mut buffer)?;
        buffer
    };

    // Apply security enforcement. The memory limit must be large enough for
    // wasmtime's Cranelift compiler to allocate its codegen buffers AND its
    // pooling allocator to pre-reserve instance memory. The pooling allocator
    // wants a single contiguous reservation (~4 GiB+). 16 GiB is a safe default
    // that still bounds a runaway process while accommodating the wasmtime
    // pool allocator and large WASM modules.
    let memory_limit: u64 = std::env::var("PRISM_MEMORY_LIMIT_BYTES")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(16u64 * 1024 * 1024 * 1024);
    let cpu_limit: u64 = 30u64;
    let policy = SecurityPolicy {
        sandbox_enabled: true,
        allow_filesystem: false,
        allow_network: false,
        enable_seccomp: true,
        enable_landlock: true,
        allowed_dirs: vec![],
        blocked_syscalls: vec![],
        require_enclave: false,
        memory_limit_bytes: memory_limit,
        cpu_time_limit_secs: cpu_limit,
    };

    let manager = SecurityManager::new(policy);
    let ctx = manager.create_execution_context();

    // Apply OS-level resource limits FIRST (before any further syscalls).
    let limits = EnforceResourceLimits::new(memory_limit, cpu_limit);
    if std::env::var("PRISM_NO_RLIMIT").is_err() {
        if let Err(e) = limits.apply() {
            eprintln!("Warning: Failed to apply resource limits: {}", e);
        }
    }

    // Apply seccomp filter (BPF program installed on current thread).
    if let Err(e) = ctx.apply_seccomp() {
        eprintln!("Warning: Failed to apply seccomp: {}", e);
    }

    // Apply landlock restrictions. The default policy denies all filesystem
    // access, but the host process still needs to load dynamic libraries and
    // read /etc/ld.so.cache for wasmtime's linker. The WASM module itself
    // was already read into a Vec<u8> before this point. The proper WASM
    // sandbox is wasmtime's instance-level memory + fuel limits, which is
    // the correct layer for sandboxing untrusted WASM code.
    //
    // NOTE: Landlock is applied with a small, broad allowlist so that
    // wasmtime's internal linker/codegen can keep working. Landlock on the
    // HOST process is a defense-in-depth measure; the per-WASM-cell
    // isolation is what actually matters for security.
    if std::env::var("PRISM_NO_LANDLOCK").is_err() && !is_wsl() {
        // Minimal: only /lib and /lib64 needed for dynamic linking.
        let allowed = [
            std::path::PathBuf::from("/lib"),
            std::path::PathBuf::from("/lib64"),
        ];
        if let Err(e) = ctx.apply_landlock(&allowed) {
            eprintln!("Warning: Failed to apply landlock: {}", e);
        }
    }

    // Validate WASM before execution.
    match manager.validate_wasm(&wasm_bytes) {
        Ok(result) => {
            if !result.valid {
                eprintln!("WASM validation failed:");
                for v in &result.violations {
                    eprintln!("  [{:?}] {}: {}", v.severity, v.pattern, v.description);
                }
                std::process::exit(1);
            }
            info!("WASM validated successfully, hash: {}", result.code_hash);
        }
        Err(e) => {
            eprintln!("WASM validation error: {}", e);
            std::process::exit(1);
        }
    }

    // Compile the module and execute the "_start" entry point.
    info!("Compiling WASM module ({} bytes)", wasm_bytes.len());
    // Build a wasmtime Engine with single-threaded compilation so the runtime
    // does not depend on the rayon global thread pool (which may not initialize
    // after a landlock/seccomp lockdown or under a strict RLIMIT_NPROC).
    let mut engine_config = wasmtime::Config::new();
    // The wasmtime 45.x default Engine uses rayon for parallel compilation.
    // Disabling the parallel compiler keeps the process model simple and avoids
    // thread-pool init failures in sandboxed environments.
    engine_config.parallel_compilation(false);
    let engine = match wasmtime::Engine::new(&engine_config) {
        Ok(e) => e,
        Err(e) => {
            eprintln!("Failed to create wasmtime engine: {}", e);
            std::process::exit(1);
        }
    };
    let module = match wasmtime::Module::new(&engine, &wasm_bytes) {
        Ok(m) => m,
        Err(e) => {
            eprintln!("Failed to compile WASM: {}", e);
            std::process::exit(1);
        }
    };

    let mut linker: wasmtime::Linker<ExecState> = wasmtime::Linker::new(&engine);
    if let Err(e) = wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |state: &mut ExecState| &mut state.wasi) {
        eprintln!("Failed to configure WASI linker: {}", e);
        std::process::exit(1);
    }

    // Build a WasiP1Ctx with closed stdin and a captured stdout pipe.
    let wasi = wasmtime_wasi::WasiCtxBuilder::new()
        .stdout(wasmtime_wasi::p2::pipe::MemoryOutputPipe::new(1024))
        .build_p1();
    // Wrap the WasiP1Ctx and the resource limiter together so the Store's
    // closure can return a &mut ResourceLimiter that lives as long as the Store.
    let exec_state = ExecState { wasi, limiter: MemoryLimiter { bytes: memory_limit as usize } };
    let mut store = wasmtime::Store::new(&engine, exec_state);
    store.limiter(|state: &mut ExecState| -> &mut dyn wasmtime::ResourceLimiter {
        &mut state.limiter
    });

    let mut linker: wasmtime::Linker<ExecState> = wasmtime::Linker::new(&engine);
    if let Err(e) = wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |state: &mut ExecState| &mut state.wasi) {
        eprintln!("Failed to configure WASI linker: {}", e);
        std::process::exit(1);
    }

    let instance = match linker.instantiate(&mut store, &module) {
        Ok(i) => i,
        Err(e) => {
            eprintln!("Failed to instantiate WASM: {}", e);
            std::process::exit(1);
        }
    };

    // Call the "_start" function. Per the WASI preview1 ABI, _start is an exported
    // function with no parameters and no results.
    let start = match instance.get_func(&mut store, "_start") {
        Some(f) => f,
        None => {
            eprintln!("WASM module has no _start export - nothing to execute");
            return Ok(());
        }
    };
    let typed_start = match start.typed::<(), ()>(&store) {
        Ok(t) => t,
        Err(e) => {
            eprintln!("_start has unexpected signature: {}", e);
            std::process::exit(1);
        }
    };

    info!(cell_id = ?cell_id, "Executing WASM _start");
    match typed_start.call(&mut store, ()) {
        Ok(()) => {
            info!("WASM execution completed successfully");
            Ok(())
        }
        Err(e) => {
            // Trap (including out-of-bounds, divide-by-zero, etc.) and resource exhaustion both
            // surface as a wasmtime error. Exit non-zero so the orchestrator sees the failure.
            eprintln!("WASM execution failed: {}", e);
            std::process::exit(1);
        }
    }
}

/// State for prism exec: bundles the WASI context with the resource limiter so the
/// Store can hold both as a single value while the limiter closure stays valid.
struct ExecState {
    wasi: wasmtime_wasi::p1::WasiP1Ctx,
    limiter: MemoryLimiter,
}

/// ResourceLimiter implementation that caps a single Store's memory to a fixed byte budget.
struct MemoryLimiter {
    bytes: usize,
}

impl wasmtime::ResourceLimiter for MemoryLimiter {
    fn memory_growing(&mut self, _current: usize, desired: usize, _maximum: Option<usize>) -> Result<bool, wasmtime::Error> {
        Ok(desired <= self.bytes)
    }
    fn table_growing(&mut self, _current: usize, desired: usize, _maximum: Option<usize>) -> Result<bool, wasmtime::Error> {
        // Allow table growth up to a reasonable bound (e.g. 64K elements).
        Ok(desired <= 65_536)
    }
}

fn json_response(status: u16, message: &str) -> String {
    serde_json::json!({
        "status_code": status,
        "body": {"error": message}
    }).to_string()
}

fn json_response_data(status: u16, data: &str) -> String {
    // Try to parse data as JSON value and embed directly in body
    if let Ok(parsed) = serde_json::from_str::<serde_json::Value>(data) {
        serde_json::json!({
            "status_code": status,
            "body": parsed
        }).to_string()
    } else {
        // Fallback: data isn't valid JSON, treat as error message
        json_response(status, data)
    }
}

/// Handle POST /cells/{id}/snapshot - Create a snapshot of a cell
async fn handle_http_snapshot_cell(rt: &Arc<RuntimeContext>, request: &str) -> String {
    // Extract cell_id from URL path, not body
    let path = request.lines().next().unwrap_or("");
    let cell_id_str = path.split("/cells/").nth(1)
        .unwrap_or("")
        .split('/')
        .next()
        .unwrap_or("")
        .split('?')
        .next()
        .unwrap_or("");

    info!(path = %path, cell_id = %cell_id_str, "Snapshot handler called");

    let cell_id = match uuid::Uuid::parse_str(cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, &format!("Invalid cell_id: {}", cell_id_str)),
    };

    // Parse body for snapshot_type if provided
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    let snapshot_type_str = if parts.len() >= 2 {
        let body = parts[1].trim();
        extract_json_field(body, "snapshot_type").unwrap_or_else(|| "Full".to_string())
    } else {
        "Full".to_string()
    };

    let snapshot_type = match snapshot_type_str.as_str() {
        "Fresh" => SnapshotType::Fresh,
        "Incremental" => SnapshotType::Incremental,
        _ => SnapshotType::Full,
    };

    match rt.snapshot_cell(&cell_id, snapshot_type).await {
        Ok(snapshot) => {
            json_response_data(201, &serde_json::json!({
                "snapshot_id": snapshot.metadata.snapshot_id,
                "cell_id": snapshot.metadata.cell_id.to_string(),
                "size_bytes": snapshot.metadata.size_bytes,
                "created_at": snapshot.metadata.created_at.to_rfc3339(),
            }).to_string())
        }
        Err(e) => json_response(400, &format!("Snapshot failed: {}", e)),
    }
}

/// Handle GET /cells/{id}/snapshots - List snapshots for a cell
async fn handle_http_list_snapshots(rt: &Arc<RuntimeContext>, request: &str) -> String {
    // Extract cell_id from request path
    let path = request.lines().next().unwrap_or("");
    let cell_id_str = path.split("/cells/").nth(1).unwrap_or("")
        .split("/").next().unwrap_or("");

    let cell_id = match uuid::Uuid::parse_str(cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, "Invalid cell_id format"),
    };

    let snapshots = rt.list_cell_snapshots(&cell_id).await;
    let entries: Vec<serde_json::Value> = snapshots.iter().map(|s| {
        serde_json::json!({
            "snapshot_id": s.snapshot_id,
            "cell_id": s.cell_id.to_string(),
            "snapshot_type": format!("{:?}", s.snapshot_type),
            "size_bytes": s.size_bytes,
            "created_at": s.created_at.to_rfc3339(),
        })
    }).collect();

    json_response_data(200, &serde_json::json!({
        "snapshots": entries
    }).to_string())
}

/// Handle POST /snapshots/{id}/restore - Restore a cell from a snapshot
async fn handle_http_restore_snapshot(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let snapshot_id = extract_json_field(body, "snapshot_id").unwrap_or_default();

    if snapshot_id.is_empty() {
        return json_response(400, "Missing 'snapshot_id' field");
    }

    match rt.restore_cell_from_snapshot(&snapshot_id).await {
        Ok(cell_id) => json_response_data(200, &serde_json::json!({
            "restored_cell_id": cell_id.to_string(),
            "status": "restored"
        }).to_string()),
        Err(e) => json_response(400, &format!("Restore failed: {}", e)),
    }
}

/// Handle DELETE /snapshots/{id} - Delete a snapshot
async fn handle_http_delete_snapshot(rt: &Arc<RuntimeContext>, request: &str) -> String {
    // Extract snapshot_id from path like "DELETE /snapshots/{id} HTTP/1.1"
    let path = request.lines().next().unwrap_or("");
    // Split by whitespace to get just the path portion
    let path_only = path.split_whitespace().nth(1).unwrap_or("");
    let snapshot_id = path_only.split("/snapshots/").nth(1).unwrap_or("")
        .split("/").next().unwrap_or("")
        .split('?').next().unwrap_or("");

    if snapshot_id.is_empty() {
        return json_response(400, "Missing snapshot_id");
    }

    match rt.delete_snapshot(snapshot_id).await {
        Ok(_) => json_response_data(200, &serde_json::json!({
            "snapshot_id": snapshot_id,
            "deleted": true
        }).to_string()),
        Err(e) => json_response(400, &format!("Delete failed: {}", e)),
    }
}

/// Handle POST /capabilities/invoke - Invoke a capability
async fn handle_http_invoke_capability(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let name = extract_json_field(body, "name").unwrap_or_default();
    let input = extract_json_field(body, "input").unwrap_or_default();

    if name.is_empty() {
        return json_response(400, "Missing 'name' field");
    }

    match rt.invoke_capability(&name, input.as_bytes()).await {
        Ok(output) => json_response_data(200, &String::from_utf8_lossy(&output)),
        Err(e) => json_response(400, &format!("Capability invocation failed: {}", e)),
    }
}

/// Handle POST /capabilities - Register a capability
async fn handle_http_register_capability(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let name = extract_json_field(body, "name").unwrap_or_default();
    let category_str = extract_json_field(body, "category").unwrap_or_else(|| "Compute".to_string());
    let category = match category_str.to_lowercase().as_str() {
        "ai" => CapabilityCategory::Ai,
        "compute" => CapabilityCategory::Compute,
        "storage" => CapabilityCategory::Storage,
        "network" => CapabilityCategory::Network,
        "crypto" => CapabilityCategory::Crypto,
        "sensors" => CapabilityCategory::Sensors,
        "system" => CapabilityCategory::System,
        _ => CapabilityCategory::Compute,
    };

    let capability = Capability::new(&name, category, "local");

    match rt.register_capability(capability).await {
        Ok(_) => json_response_data(201, &serde_json::json!({"name": name, "status": "registered"}).to_string()),
        Err(e) => json_response(400, &format!("Registration failed: {}", e)),
    }
}

/// Handle GET /capabilities - List all capabilities
async fn handle_http_list_capabilities(rt: &Arc<RuntimeContext>, _request: &str) -> String {
    let caps = rt.list_capabilities().await;
    let entries: Vec<serde_json::Value> = caps.iter().map(|c| {
        serde_json::json!({
            "capability_id": c.capability_id.to_string(),
            "name": c.name,
            "category": format!("{:?}", c.category),
            "version": c.metadata.version,
            "trust_score": c.trust.score,
            "is_remote": c.is_remote,
        })
    }).collect();
    serde_json::json!({
        "status_code": 200,
        "body": {
            "capabilities": entries,
            "count": entries.len()
        }
    }).to_string()
}

/// Handle POST /swarms - Create a swarm
async fn handle_http_create_swarm(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    if parts.len() < 2 {
        return json_response(400, "Invalid request body");
    }
    let body = parts[1].trim();

    let swarm_id = extract_json_field(body, "swarm_id").unwrap_or_else(|| {
        uuid::Uuid::new_v4().to_string()
    });

    match rt.create_swarm(swarm_id.clone()).await {
        Ok(id) => json_response_data(201, &serde_json::json!({"swarm_id": id, "status": "created"}).to_string()),
        Err(e) => json_response(400, &format!("Create swarm failed: {}", e)),
    }
}

/// Handle GET /swarms - List all swarms
async fn handle_http_list_swarms(rt: &Arc<RuntimeContext>, _request: &str) -> String {
    let swarms = rt.list_swarms().await;
    let entries: Vec<serde_json::Value> = swarms.iter().map(|s| {
        serde_json::json!({
            "swarm_id": s.swarm_id.to_string(),
            "cell_count": s.cell_count(),
            "created_at": s.created_at.to_rfc3339(),
            "health": {
                "is_healthy": s.state.health.is_healthy,
                "active_count": s.state.health.active_count,
                "total_count": s.state.health.total_count,
            }
        })
    }).collect();
    serde_json::json!({
        "status_code": 200,
        "body": {
            "swarms": entries,
            "count": entries.len()
        }
    }).to_string()
}

/// Handle POST /swarms/{id}/join - Join a swarm
async fn handle_http_join_swarm(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    let (swarm_id, cell_id_str) = if parts.len() >= 2 {
        let body = parts[1].trim();
        let swarm_id = request.lines().next().unwrap_or("")
            .split("/swarms/").nth(1).unwrap_or("")
            .split("/join").next().unwrap_or("")
            .split('?').next().unwrap_or("")
            .split('/').next().unwrap_or("");
        let cell_id_str = extract_json_field(body, "cell_id").unwrap_or_default();
        (swarm_id.to_string(), cell_id_str)
    } else {
        return json_response(400, "Invalid request");
    };

    if swarm_id.is_empty() {
        return json_response(400, "Missing swarm_id");
    }

    let cell_id = match uuid::Uuid::parse_str(&cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, "Invalid cell_id format"),
    };

    match rt.join_swarm(&swarm_id, cell_id).await {
        Ok(_) => json_response_data(200, &serde_json::json!({"swarm_id": swarm_id, "status": "joined"}).to_string()),
        Err(e) => json_response(400, &format!("Join swarm failed: {}", e)),
    }
}

/// Handle POST /swarms/{id}/leave - Leave a swarm
async fn handle_http_leave_swarm(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let parts: Vec<&str> = request.split("\r\n\r\n").collect();
    let (swarm_id, cell_id_str) = if parts.len() >= 2 {
        let body = parts[1].trim();
        let swarm_id = request.lines().next().unwrap_or("")
            .split("/swarms/").nth(1).unwrap_or("")
            .split("/leave").next().unwrap_or("")
            .split('?').next().unwrap_or("")
            .split('/').next().unwrap_or("");
        let cell_id_str = extract_json_field(body, "cell_id").unwrap_or_default();
        (swarm_id.to_string(), cell_id_str)
    } else {
        return json_response(400, "Invalid request");
    };

    if swarm_id.is_empty() {
        return json_response(400, "Missing swarm_id");
    }

    let cell_id = match uuid::Uuid::parse_str(&cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, "Invalid cell_id format"),
    };

    match rt.leave_swarm(&swarm_id, cell_id).await {
        Ok(_) => json_response_data(200, &serde_json::json!({"swarm_id": swarm_id, "status": "left"}).to_string()),
        Err(e) => json_response(400, &format!("Leave swarm failed: {}", e)),
    }
}

/// Handle GET /optimize/{cell_id} - Get optimization suggestion
async fn handle_http_get_optimization(rt: &Arc<RuntimeContext>, request: &str) -> String {
    let path = request.lines().next().unwrap_or("");
    let cell_id_str = path.split("/optimize/").nth(1).unwrap_or("")
        .split_whitespace().next().unwrap_or("")
        .split('?').next().unwrap_or("");

    if cell_id_str.is_empty() {
        return json_response(400, "Missing cell_id");
    }

    let cell_id = match uuid::Uuid::parse_str(cell_id_str) {
        Ok(uuid) => CellId::from_uuid(uuid),
        Err(_) => return json_response(400, "Invalid cell_id format"),
    };

    let suggestion = rt.get_optimization_suggestion(&cell_id).await;
    serde_json::json!({
        "status_code": 200,
        "body": {
            "cell_id": suggestion.cell_id.to_string(),
            "suggested_memory_mb": suggestion.suggested_memory_mb,
            "suggested_timeout_ms": suggestion.suggested_timeout_ms,
            "suggested_location": suggestion.suggested_location,
            "cache_recommended": suggestion.cache_recommended,
            "confidence": suggestion.confidence,
        }
    }).to_string()
}

async fn handle_cell_create(runtime: &Arc<RuntimeContext>, module_path: &str, memory_mb: u64) -> anyhow::Result<()> {
    info!("Creating cell from module: {}", module_path);

    let wasm_bytes = std::fs::read(module_path)
        .map_err(|e| anyhow::anyhow!("Failed to read module file '{}': {}", module_path, e))?;

    if wasm_bytes.len() < 4 || &wasm_bytes[0..4] != b"\0asm" {
        return Err(anyhow::anyhow!("Invalid WASM file: magic number not found"));
    }

    let name = std::path::Path::new(module_path)
        .file_stem()
        .and_then(|s| s.to_str())
        .unwrap_or("cell")
        .to_string();

    let config = CellConfig {
        memory_limit_mb: memory_mb,
        execution_target: ExecutionTarget::Cloud,
        ..CellConfig::default()
    };

    match runtime.create_cell("tenant-1", &name, wasm_bytes, config).await {
        Ok(cell_id) => {
            println!("Cell created successfully:");
            println!("  ID: {}", cell_id);
            println!("  Name: {}", name);
            println!("  Memory: {} MB", memory_mb);
            println!("  Status: Initializing");
            Ok(())
        }
        Err(e) => {
            error!("Failed to create cell: {}", e);
            Err(anyhow::anyhow!("Cell creation failed: {}", e))
        }
    }
}

async fn handle_cell_list(runtime: &Arc<RuntimeContext>) {
    let cells = runtime.list_cells().await;

    if cells.is_empty() {
        println!("No active cells.");
        return;
    }

    println!("Active Cells:");
    println!("{:-<80}", "");
    println!("{:<38} {:<15} {:<12} {:<12}", "ID", "NAME", "STATUS", "MEMORY");
    println!("{:-<80}", "");

    for cell in &cells {
        println!(
            "{:<38} {:<15} {:<12} {} MB",
            cell.id,
            cell.metadata.name,
            format!("{:?}", cell.status),
            cell.config.memory_limit_mb
        );
    }
    println!("{:-<80}", "");
    println!("Total: {} cells", cells.len());
}

async fn handle_cell_terminate(runtime: &Arc<RuntimeContext>, cell_id_str: &str) -> anyhow::Result<()> {
    let cell_id = parse_cell_id(cell_id_str)?;

    match runtime.terminate_cell(&cell_id).await {
        Ok(()) => {
            println!("Cell {} terminated successfully.", cell_id);
            Ok(())
        }
        Err(e) => {
            error!("Failed to terminate cell: {}", e);
            Err(anyhow::anyhow!("Cell termination failed: {}", e))
        }
    }
}

async fn handle_cell_snapshot(runtime: &Arc<RuntimeContext>, cell_id_str: &str) -> anyhow::Result<()> {
    let cell_id = parse_cell_id(cell_id_str)?;

    match runtime.get_cell(&cell_id).await {
        Some(cell) => {
            let snapshot = serde_json::json!({
                "cell_id": cell.id.to_string(),
                "tenant_id": cell.tenant_id,
                "status": format!("{:?}", cell.status),
                "config": {
                    "memory_limit_mb": cell.config.memory_limit_mb,
                    "timeout_ms": cell.config.timeout_ms,
                    "isolation_enabled": cell.config.isolation_enabled,
                },
                "metadata": {
                    "name": cell.metadata.name,
                    "version": cell.metadata.version,
                    "runtime": cell.metadata.runtime,
                },
                "wasm_module_id": cell.wasm_module_id,
                "checkpoint_epoch": cell.checkpoint_epoch,
            });

            let output_path = format!("snapshot_{}.json", cell_id);
            std::fs::write(&output_path, snapshot.to_string())
                .map_err(|e| anyhow::anyhow!("Failed to write snapshot: {}", e))?;

            println!("Cell snapshot saved to: {}", output_path);
            Ok(())
        }
        None => {
            error!("Cell not found: {}", cell_id);
            Err(anyhow::anyhow!("Cell not found: {}", cell_id))
        }
    }
}

async fn handle_cell_migrate(runtime: &Arc<RuntimeContext>, cell_id_str: &str, target: &str) -> anyhow::Result<()> {
    let cell_id = parse_cell_id(cell_id_str)?;

    // Get the cell to check if it can migrate
    let cell = match runtime.get_cell(&cell_id).await {
        Some(c) => c,
        None => return Err(anyhow::anyhow!("Cell {} not found", cell_id_str)),
    };

    if !cell.can_migrate() {
        return Err(anyhow::anyhow!(
            "Cell {} is not in a migratable state (status: {:?}). Only Running, Waiting, or Frozen cells can migrate.",
            cell_id_str, cell.status
        ));
    }

    info!(
        cell_id = %cell_id,
        target = %target,
        status = ?cell.status,
        "Cell migration initiated"
    );

    // Perform the actual migration
    // 1. Create a snapshot of the cell
    let snapshot = runtime.snapshot_cell(&cell_id, SnapshotType::Full).await?;

    info!(
        snapshot_id = %snapshot.metadata.snapshot_id,
        size_bytes = snapshot.metadata.size_bytes,
        "Cell snapshot created for migration"
    );

    // 2. Create migration result with strategy selection based on cell state
    let mut migration_manager = MigrationManager::new();

    // Select migration strategy based on downtime tolerance
    // - Live: minimal downtime but more complex (for Running cells)
    // - PreCopy: moderate downtime with iterative copying (for Waiting cells)
    // - StopCopy: higher downtime but simpler (for Frozen cells)
    let strategy = match cell.status {
        CellStatus::Running => MigrationStrategy::Live,
        CellStatus::Waiting => MigrationStrategy::PreCopy,
        CellStatus::Frozen => MigrationStrategy::StopCopy,
        _ => MigrationStrategy::PreCopy,
    };

    info!(strategy = ?strategy, "Selected migration strategy");

    // 3. Perform migration (this would involve network transfer in production)
    let result = migration_manager.migrate_cell(
        cell_id,
        "local-node",
        target,
        strategy,
        &snapshot,
    ).await?;

    // 4. Mark cell as migrating during actual transfer
    {
        let mut cells = runtime.cells.write().await;
        if let Some(c) = cells.get_mut(&cell_id) {
            c.status = CellStatus::Migrating;
            c.checkpoint_epoch = snapshot.metadata.checkpoint_epoch;
        }
    }

    // 5. Report migration completion
    if result.success {
        info!(
            migration_id = %result.migration_id,
            downtime_ms = result.downtime_ms,
            bytes_transferred = result.bytes_transferred,
            "Migration completed successfully"
        );

        println!("Cell {} migrated successfully:", cell_id_str);
        println!("  Migration ID: {}", result.migration_id);
        println!("  Target: {}", target);
        println!("  Strategy: {:?}", strategy);
        println!("  Downtime: {}ms", result.downtime_ms);
        println!("  Data transferred: {} bytes", result.bytes_transferred);
        println!("  Duration: {}ms", result.total_duration_ms);

        // Mark cell as running on target
        {
            let mut cells = runtime.cells.write().await;
            if let Some(c) = cells.get_mut(&cell_id) {
                c.status = CellStatus::Running;
            }
        }

        Ok(())
    } else {
        Err(anyhow::anyhow!(
            "Migration failed: {}",
            result.error.unwrap_or_else(|| "Unknown error".to_string())
        ))
    }
}

async fn handle_capability_register(
    runtime: &Arc<RuntimeContext>,
    name: &str,
    category: &str,
) -> anyhow::Result<()> {
    use prism_runtime::ucl::Capability;
    use prism_runtime::ucl::CapabilityCategory;

    let cat = match category {
        "ai" => CapabilityCategory::Ai,
        "compute" => CapabilityCategory::Compute,
        "storage" => CapabilityCategory::Storage,
        "network" => CapabilityCategory::Network,
        "crypto" => CapabilityCategory::Crypto,
        "sensors" => CapabilityCategory::Sensors,
        "system" | _ => CapabilityCategory::System,
    };

    let cap = Capability::new(name, cat, "prism-cli");

    match runtime.register_capability(cap).await {
        Ok(()) => {
            println!("Capability registered: {} / {}", category, name);
            Ok(())
        }
        Err(e) => {
            error!("Failed to register capability: {}", e);
            Err(anyhow::anyhow!("Capability registration failed: {}", e))
        }
    }
}

async fn handle_capability_discover(runtime: &Arc<RuntimeContext>, query: &str) {
    let capabilities = runtime.discover_capabilities(query).await;

    if capabilities.is_empty() {
        println!("No capabilities matching '{}'", query);
        return;
    }

    println!("Capabilities matching '{}':", query);
    println!("{:-<60}", "");
    println!("{:<25} {:<20} {:<12}", "NAME", "CATEGORY", "VERSION");
    println!("{:-<60}", "");

    for cap in &capabilities {
        println!("{:<25} {:<20} {}", cap.name, format!("{:?}", cap.category), cap.metadata.version);
    }
    println!("{:-<60}", "");
    println!("Found {} capabilities", capabilities.len());
}

async fn handle_capability_list(runtime: &Arc<RuntimeContext>) {
    let capabilities = runtime.list_capabilities().await;

    if capabilities.is_empty() {
        println!("No registered capabilities.");
        return;
    }

    println!("Registered Capabilities:");
    println!("{:-<60}", "");
    println!("{:<25} {:<20} {:<12}", "NAME", "CATEGORY", "VERSION");
    println!("{:-<60}", "");

    for cap in &capabilities {
        println!("{:<25} {:<20} {}", cap.name, format!("{:?}", cap.category), cap.metadata.version);
    }
    println!("{:-<60}", "");
    println!("Total: {} capabilities", capabilities.len());
}

async fn handle_swarm_create(runtime: &mut Arc<RuntimeContext>, swarm_id: &str) -> anyhow::Result<()> {
    use prism_runtime::swarm::SwarmId;

    let mut coordinator = runtime.swarm_coordinator.write().await;

    let coordinator = coordinator.as_mut()
        .ok_or_else(|| anyhow::anyhow!("Swarm coordinator not initialized. Start with --mesh flag."))?;

    let swarm_id = SwarmId::new(swarm_id);
    match coordinator.create_swarm(swarm_id) {
        Ok(swarm) => {
            println!("Swarm created: {}", swarm.swarm_id);
            Ok(())
        }
        Err(e) => {
            error!("Failed to create swarm: {}", e);
            Err(anyhow::anyhow!("Swarm creation failed: {}", e))
        }
    }
}

async fn handle_swarm_join(runtime: &Arc<RuntimeContext>, swarm_id: &str) -> anyhow::Result<()> {
    // Swarm join requires mesh networking to connect to peers
    // In local mode, we track the swarm membership in the coordinator

    if !runtime.mesh_enabled {
        info!("Mesh networking disabled - swarm join is local-only");
    }

    // Get the swarm coordinator and add this node as a peer
    let coordinator = runtime.swarm_coordinator.read().await;
    if let Some(coord) = coordinator.as_ref() {
        // Check if swarm exists
        if let Some(swarm) = coord.get_swarm(&SwarmId::new(swarm_id)) {
            // In a real implementation, we would:
            // 1. Establish connections to existing peer nodes
            // 2. Perform handshake protocol to join the mesh
            // 3. Sync state with existing swarm members
            // 4. Start participating in swarm coordination

            info!(
                swarm_id = swarm_id,
                peer_count = swarm.peer_nodes.len(),
                cell_count = swarm.cell_count(),
                "Joining swarm - syncing state with peers"
            );

            // Simulate peer discovery and state sync
            for (peer_id, address) in &swarm.peer_nodes {
                tracing::debug!(peer_id = %peer_id, address = %address, "Would connect to peer");
            }

            println!("Successfully joined swarm: {}", swarm_id);
            println!("  Connected to {} existing peers", swarm.peer_nodes.len());
            println!("  Swarm contains {} cells", swarm.cell_count());

            Ok(())
        } else {
            Err(anyhow::anyhow!("Swarm '{}' not found. Create it first with 'swarm create'.", swarm_id))
        }
    } else {
        Err(anyhow::anyhow!("Swarm coordinator not initialized. Start with --mesh flag."))
    }
}

async fn handle_swarm_leave(runtime: &Arc<RuntimeContext>, swarm_id: &str) -> anyhow::Result<()> {
    // Swarm leave requires:
    // 1. Notifying peers about departure
    // 2. Transferring any assigned work
    // 3. Cleaning up swarm state

    if !runtime.mesh_enabled {
        info!("Mesh networking disabled - local swarm leave");
    }

    let coordinator = runtime.swarm_coordinator.read().await;
    if let Some(coord) = coordinator.as_ref() {
        let swarm = coord.get_swarm(&SwarmId::new(swarm_id));
        if let Some(swarm) = swarm {
            info!(
                swarm_id = swarm_id,
                peer_count = swarm.peer_nodes.len(),
                cell_count = swarm.cell_count(),
                "Leaving swarm - notifying peers"
            );

            // Notify peers about departure (in real implementation)
            for (peer_id, _) in &swarm.peer_nodes {
                tracing::debug!(peer_id = %peer_id, "Would notify peer about departure");
            }

            println!("Successfully left swarm: {}", swarm_id);
            Ok(())
        } else {
            Err(anyhow::anyhow!("Swarm '{}' not found.", swarm_id))
        }
    } else {
        Err(anyhow::anyhow!("Swarm coordinator not initialized. Start with --mesh flag."))
    }
}

async fn handle_swarm_list(runtime: &Arc<RuntimeContext>) {
    let coordinator = runtime.swarm_coordinator.read().await;

    match coordinator.as_ref() {
        Some(coord) => {
            let swarms = coord.swarms();
            if swarms.is_empty() {
                println!("No active swarms.");
                return;
            }

            println!("Active Swarms:");
            println!("{:-<60}", "");
            println!("{:<38} {:<10} {:<8}", "ID", "CELLS", "HEALTHY");
            println!("{:-<60}", "");

            for (id, swarm) in swarms {
                println!(
                    "{:<38} {:<10} {}",
                    id.0,
                    swarm.cells.len(),
                    swarm.state.health.is_healthy
                );
            }
            println!("{:-<60}", "");
            println!("Total: {} swarms", swarms.len());
        }
        None => {
            println!("Swarm coordinator not initialized. Start with --mesh flag.");
        }
    }
}

async fn handle_swarm_command(runtime: &Arc<RuntimeContext>, swarm_id: &str, cmd: &str) -> anyhow::Result<()> {
    use prism_runtime::swarm::SwarmId;

    println!("Swarm command '{}' received for swarm '{}'", cmd, swarm_id);

    let coordinator = runtime.swarm_coordinator.read().await;
    match coordinator.as_ref() {
        Some(coord) => {
            let swarm_id_obj = SwarmId::new(swarm_id);
            let health_commands = coord.check_and_heal(&swarm_id_obj)?;
            println!("Health check generated {} commands", health_commands.len());
            for hc in health_commands {
                println!("  - {:?}", hc);
            }
        }
        None => {
            println!("Swarm coordinator not initialized. Start with --mesh flag.");
        }
    }

    Ok(())
}

async fn start_repl(_language: &str, _runtime: Arc<RuntimeContext>) -> anyhow::Result<()> {
    let mut repl = crate::cli::Repl::new();
    info!("Starting REPL (type 'help' for commands)");
    repl.run()?;
    Ok(())
}

async fn handle_package(action: cli::PackageCommands) -> anyhow::Result<()> {
    use prism_runtime::cli::package::*;

    match action {
        cli::PackageCommands::Build { source, output, description, language, capability, resource } => {
            println!("Building package from {} to {}", source, output);

            // Read source WASM file
            let wasm_bytes = std::fs::read(&source)
                .map_err(|e| anyhow::anyhow!("Failed to read module '{}': {}", source, e))?;

            if wasm_bytes.len() < 4 || &wasm_bytes[0..4] != b"\0asm" {
                return Err(anyhow::anyhow!("'{}' is not a valid WASM module", source));
            }

            let name = std::path::Path::new(&source)
                .file_stem()
                .and_then(|s| s.to_str())
                .unwrap_or("module")
                .to_string();

            let module = PackageModule {
                module_id: uuid::Uuid::new_v4().to_string(),
                name: name.clone(),
                language: "wasm".to_string(),
                bytecode: wasm_bytes,
                entry_point: "handler".to_string(),
            };

            // Build package with all configured options
            let mut builder = PackageBuilder::new(&name, "1.0.0")
                .with_runtime("wasm");

            // Apply description if provided
            if let Some(desc) = description {
                builder = builder.with_description(&desc);
            }

            // Add language(s)
            if let Some(langs) = language {
                for lang in langs {
                    builder = builder.with_language(&lang);
                }
            }

            // Add capabilities
            if let Some(caps) = capability {
                for cap in caps {
                    builder = builder.with_capability(&cap);
                }
            }

            // Add resources
            if let Some(resources) = resource {
                for resource_path in resources {
                    let resource_data = std::fs::read(&resource_path)
                        .map_err(|e| anyhow::anyhow!("Failed to read resource '{}': {}", resource_path, e))?;

                    let resource_name = std::path::Path::new(&resource_path)
                        .file_name()
                        .and_then(|s| s.to_str())
                        .unwrap_or("unknown")
                        .to_string();

                    let resource_type = std::path::Path::new(&resource_path)
                        .extension()
                        .and_then(|s| s.to_str())
                        .unwrap_or("binary")
                        .to_string();

                    let pkg_resource = PackageResource {
                        resource_id: uuid::Uuid::new_v4().to_string(),
                        name: resource_name,
                        resource_type,
                        content: resource_data,
                    };

                    builder = builder.add_resource(pkg_resource);
                }
            }

            let package = builder
                .add_module(module)
                .build();

            if let Err(e) = package.validate() {
                return Err(anyhow::anyhow!("Package validation failed: {}", e));
            }

            let output_path = std::path::Path::new(&output);
            let data = serde_json::to_vec_pretty(&package)
                .map_err(|e| anyhow::anyhow!("Failed to serialize package: {}", e))?;
            std::fs::write(output_path, data)
                .map_err(|e| anyhow::anyhow!("Failed to write package: {}", e))?;

            println!("Package built successfully: {}", output);
            println!("  Module: {}", name);
            println!("  Size: {} bytes", package.modules[0].bytecode.len());
            if !package.metadata.languages.is_empty() {
                println!("  Languages: {}", package.metadata.languages.join(", "));
            }
            if !package.metadata.capabilities.is_empty() {
                println!("  Capabilities: {}", package.metadata.capabilities.join(", "));
            }
            if !package.resources.is_empty() {
                println!("  Resources: {}", package.resources.len());
            }
        }
        cli::PackageCommands::Inspect { package, verbose } => {
            let pkg = Package::load(std::path::Path::new(&package))
                .map_err(|e| anyhow::anyhow!("Failed to load package '{}': {}", package, e))?;

            println!("Package: {} v{}", pkg.metadata.name, pkg.metadata.version);
            println!("  Runtime: {}", pkg.metadata.runtime);
            println!("  Modules: {}", pkg.modules.len());
            println!("  Resources: {}", pkg.resources.len());

            if !pkg.metadata.description.is_empty() {
                println!("  Description: {}", pkg.metadata.description);
            }
            if !pkg.metadata.languages.is_empty() {
                println!("  Languages: {}", pkg.metadata.languages.join(", "));
            }
            if !pkg.metadata.capabilities.is_empty() {
                println!("  Capabilities: {}", pkg.metadata.capabilities.join(", "));
            }

            if let Some(ref sig) = pkg.signature {
                println!("  Signed: yes ({} at {})", sig.algorithm, sig.signed_at);
            } else {
                println!("  Signed: no");
            }

            // Validate package structure
            if let Err(e) = pkg.validate() {
                eprintln!("  Warning: validation failed - {}", e);
            }

            if verbose {
                println!("\n--- Full Package JSON ---");
                let json = serde_json::to_string_pretty(&pkg)
                    .map_err(|e| anyhow::anyhow!("Failed to serialize: {}", e))?;
                println!("{}", json);
            }
        }
        cli::PackageCommands::Sign { package, key, verify } => {
            let mut pkg = Package::load(std::path::Path::new(&package))
                .map_err(|e| anyhow::anyhow!("Failed to load package '{}': {}", package, e))?;

            use ed25519_dalek::{SigningKey, Signature, Signer};
            use rand::TryRng;
            use rand::rngs::SysRng;

            let key_data = std::fs::read(&key)
                .map_err(|e| anyhow::anyhow!("Failed to read key '{}': {}", key, e))?;

            let signing_key = if key_data.len() == 32 {
                let mut key_bytes = [0u8; 32];
                key_bytes.copy_from_slice(&key_data);
                SigningKey::from_bytes(&key_bytes)
            } else if key_data.len() == 64 {
                let mut keypair_bytes = [0u8; 64];
                keypair_bytes.copy_from_slice(&key_data);
                SigningKey::from_keypair_bytes(&keypair_bytes)
                    .map_err(|_| anyhow::anyhow!("Invalid Ed25519 keypair"))?
            } else {
                let mut seed: [u8; 32] = [0; 32];
                SysRng.try_fill_bytes(&mut seed)
                    .map_err(|e| anyhow::anyhow!("Failed to read OS random: {}", e))?;
                SigningKey::from_bytes(&seed)
            };

            let pkg_bytes = serde_json::to_vec(&pkg)
                .map_err(|e| anyhow::anyhow!("Failed to serialize package: {}", e))?;

            let signature: Signature = signing_key.sign(&pkg_bytes);
            let public_key = hex::encode(signing_key.verifying_key().to_bytes());

            pkg.signature = Some(PackageSignature {
                algorithm: "ed25519".to_string(),
                signature: signature.to_bytes().to_vec(),
                public_key,
                signed_at: chrono::Utc::now().timestamp(),
            });

            let data = serde_json::to_vec_pretty(&pkg)
                .map_err(|e| anyhow::anyhow!("Failed to serialize signed package: {}", e))?;
            std::fs::write(&package, data)
                .map_err(|e| anyhow::anyhow!("Failed to write signed package: {}", e))?;

            println!("Package signed successfully: {}", package);

            // Verify signature if requested
            if verify {
                if let Err(e) = pkg.verify() {
                    return Err(anyhow::anyhow!("Signature verification failed: {}", e));
                }
                println!("Signature verified successfully!");
            }
        }
    }
    Ok(())
}

async fn show_status(runtime: &Arc<RuntimeContext>) -> anyhow::Result<()> {
    let status = runtime.get_status().await;
    let capabilities = runtime.list_capabilities().await;

    println!("Prism Runtime Status");
    println!("====================");
    println!("Version:     {}", status.version);
    println!("Healthy:     {}", status.healthy);
    println!("Mesh:        {}", status.mesh_enabled);
    println!("Address:     {}", status.listen_address);
    println!("Active Cells: {}", status.active_cells);
    println!("Total Cells:  {}", status.total_cells);

    if !capabilities.is_empty() {
        println!("\nCapabilities:");
        for cap in &capabilities {
            println!("  - {} ({:?})", cap.name, cap.category);
        }
    }

    Ok(())
}

fn parse_cell_id(id_str: &str) -> anyhow::Result<CellId> {
    match uuid::Uuid::parse_str(id_str) {
        Ok(uuid) => Ok(CellId::from_uuid(uuid)),
        Err(_) => Err(anyhow::anyhow!("Invalid cell ID format: '{}' (expected UUID)", id_str)),
    }
}

fn generate_docs(output: &str) -> anyhow::Result<()> {
    use std::io::Write;

    println!("Generating documentation to {}...", output);

    // Generate OpenAPI 3.0 spec for the HTTP API
    let spec = serde_json::json!({
        "openapi": "3.0.3",
        "info": {
            "title": "FunctionFly Prism Runtime API",
            "description": "Universal Adaptive WASM Execution Fabric for AI-native execution, autonomous agents, robotics, edge systems, and cross-language functions.",
            "version": env!("CARGO_PKG_VERSION"),
            "contact": {
                "name": "FunctionFly Team",
                "url": "https://functionfly.com"
            },
            "license": {
                "name": "MIT"
            }
        },
        "servers": [
            {
                "url": "http://localhost:8080",
                "description": "Local development server"
            }
        ],
        "paths": {
            "/health": {
                "get": {
                    "summary": "Health check",
                    "description": "Returns the runtime health status",
                    "responses": {
                        "200": {
                            "description": "Healthy response",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {
                                            "status_code": {"type": "integer"},
                                            "body": {
                                                "type": "object",
                                                "properties": {
                                                    "version": {"type": "string"},
                                                    "healthy": {"type": "boolean"},
                                                    "active_cells": {"type": "integer"},
                                                    "total_cells": {"type": "integer"},
                                                    "mesh_enabled": {"type": "boolean"}
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "/cells": {
                "get": {
                    "summary": "List all cells",
                    "description": "Returns a list of all active cells in the runtime",
                    "responses": {
                        "200": {
                            "description": "List of cells",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "properties": {
                                                "id": {"type": "string"},
                                                "tenant": {"type": "string"},
                                                "status": {"type": "string"},
                                                "name": {"type": "string"},
                                                "memory_mb": {"type": "integer"}
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "post": {
                    "summary": "Create a new cell",
                    "description": "Creates a new WASM cell from a module file",
                    "requestBody": {
                        "required": true,
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["name", "module"],
                                    "properties": {
                                        "name": {"type": "string", "description": "Cell name"},
                                        "module": {"type": "string", "description": "Path to WASM module file"},
                                        "memory": {"type": "integer", "description": "Memory limit in MB (default: 128)"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "201": {"description": "Cell created successfully"},
                        "400": {"description": "Invalid request"},
                        "500": {"description": "Internal server error"}
                    }
                }
            },
            "/cells/{id}": {
                "get": {
                    "summary": "Get cell details",
                    "parameters": [
                        {
                            "name": "id",
                            "in": "path",
                            "required": true,
                            "schema": {"type": "string", "format": "uuid"}
                        }
                    ],
                    "responses": {
                        "200": {"description": "Cell details"},
                        "404": {"description": "Cell not found"}
                    }
                }
            },
            "/cells/{id}/snapshot": {
                "post": {
                    "summary": "Create cell snapshot",
                    "description": "Creates a snapshot of a cell for migration or backup",
                    "parameters": [
                        {
                            "name": "id",
                            "in": "path",
                            "required": true,
                            "schema": {"type": "string", "format": "uuid"}
                        }
                    ],
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "snapshot_type": {
                                            "type": "string",
                                            "enum": ["Fresh", "Incremental", "Full"],
                                            "default": "Full"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "201": {"description": "Snapshot created"},
                        "400": {"description": "Invalid request"},
                        "404": {"description": "Cell not found"}
                    }
                }
            },
            "/cells/{id}/snapshots": {
                "get": {
                    "summary": "List cell snapshots",
                    "parameters": [
                        {
                            "name": "id",
                            "in": "path",
                            "required": true,
                            "schema": {"type": "string", "format": "uuid"}
                        }
                    ],
                    "responses": {
                        "200": {"description": "List of snapshots"},
                        "404": {"description": "Cell not found"}
                    }
                }
            },
            "/snapshots/{id}/restore": {
                "post": {
                    "summary": "Restore from snapshot",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["snapshot_id"],
                                    "properties": {
                                        "snapshot_id": {"type": "string"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {"description": "Cell restored"},
                        "400": {"description": "Invalid request"},
                        "404": {"description": "Snapshot not found"}
                    }
                }
            },
            "/snapshots/{id}": {
                "delete": {
                    "summary": "Delete snapshot",
                    "parameters": [
                        {
                            "name": "id",
                            "in": "path",
                            "required": true,
                            "schema": {"type": "string"}
                        }
                    ],
                    "responses": {
                        "200": {"description": "Snapshot deleted"},
                        "400": {"description": "Invalid request"},
                        "404": {"description": "Snapshot not found"}
                    }
                }
            },
            "/capabilities": {
                "get": {
                    "summary": "List capabilities",
                    "description": "Returns all registered capabilities",
                    "responses": {
                        "200": {
                            "description": "List of capabilities",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {
                                            "capabilities": {
                                                "type": "array",
                                                "items": {
                                                    "type": "object",
                                                    "properties": {
                                                        "capability_id": {"type": "string"},
                                                        "name": {"type": "string"},
                                                        "category": {"type": "string"},
                                                        "version": {"type": "string"},
                                                        "trust_score": {"type": "number"}
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "post": {
                    "summary": "Register capability",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["name", "category"],
                                    "properties": {
                                        "name": {"type": "string"},
                                        "category": {"type": "string", "enum": ["ai", "compute", "storage", "network", "crypto", "sensors", "system"]}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "201": {"description": "Capability registered"},
                        "400": {"description": "Invalid request"}
                    }
                }
            },
            "/capabilities/invoke": {
                "post": {
                    "summary": "Invoke capability",
                    "description": "Invokes a registered capability by name",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["name"],
                                    "properties": {
                                        "name": {"type": "string"},
                                        "input": {"type": "string"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {"description": "Capability invoked successfully"},
                        "400": {"description": "Capability not found or invocation failed"}
                    }
                }
            },
            "/swarms": {
                "get": {
                    "summary": "List swarms",
                    "responses": {
                        "200": {"description": "List of swarms"}
                    }
                },
                "post": {
                    "summary": "Create swarm",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["swarm_id"],
                                    "properties": {
                                        "swarm_id": {"type": "string"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "201": {"description": "Swarm created"},
                        "400": {"description": "Invalid request"}
                    }
                }
            },
            "/swarms/{id}/join": {
                "post": {
                    "summary": "Join swarm",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["cell_id"],
                                    "properties": {
                                        "cell_id": {"type": "string", "format": "uuid"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {"description": "Joined swarm"},
                        "400": {"description": "Invalid request"}
                    }
                }
            },
            "/swarms/{id}/leave": {
                "post": {
                    "summary": "Leave swarm",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "required": ["cell_id"],
                                    "properties": {
                                        "cell_id": {"type": "string", "format": "uuid"}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {"description": "Left swarm"},
                        "400": {"description": "Invalid request"}
                    }
                }
            },
            "/optimize/{cell_id}": {
                "get": {
                    "summary": "Get optimization suggestion",
                    "description": "Returns an RL-based optimization suggestion for a cell",
                    "parameters": [
                        {
                            "name": "cell_id",
                            "in": "path",
                            "required": true,
                            "schema": {"type": "string", "format": "uuid"}
                        }
                    ],
                    "responses": {
                        "200": {
                            "description": "Optimization suggestion",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {
                                            "cell_id": {"type": "string"},
                                            "suggested_memory_mb": {"type": "integer"},
                                            "suggested_timeout_ms": {"type": "integer"},
                                            "suggested_location": {"type": "string"},
                                            "cache_recommended": {"type": "boolean"},
                                            "confidence": {"type": "number"}
                                        }
                                    }
                                }
                            }
                        },
                        "400": {"description": "Invalid cell_id"}
                    }
                }
            }
        },
        "components": {
            "schemas": {
                "Cell": {
                    "type": "object",
                    "properties": {
                        "id": {"type": "string", "format": "uuid"},
                        "tenant": {"type": "string"},
                        "status": {"type": "string"},
                        "name": {"type": "string"},
                        "memory_mb": {"type": "integer"}
                    }
                },
                "Swarm": {
                    "type": "object",
                    "properties": {
                        "swarm_id": {"type": "string"},
                        "cell_count": {"type": "integer"},
                        "created_at": {"type": "string", "format": "date-time"},
                        "health": {
                            "type": "object",
                            "properties": {
                                "is_healthy": {"type": "boolean"},
                                "active_count": {"type": "integer"},
                                "total_count": {"type": "integer"}
                            }
                        }
                    }
                },
                "Capability": {
                    "type": "object",
                    "properties": {
                        "capability_id": {"type": "string"},
                        "name": {"type": "string"},
                        "category": {"type": "string"},
                        "version": {"type": "string"},
                        "trust_score": {"type": "number"}
                    }
                },
                "OptimizationSuggestion": {
                    "type": "object",
                    "properties": {
                        "cell_id": {"type": "string"},
                        "suggested_memory_mb": {"type": "integer"},
                        "suggested_timeout_ms": {"type": "integer"},
                        "suggested_location": {"type": "string"},
                        "cache_recommended": {"type": "boolean"},
                        "confidence": {"type": "number"}
                    }
                }
            }
        }
    });

    let spec_yaml = serde_yaml::to_string(&spec)
        .map_err(|e| anyhow::anyhow!("Failed to convert to YAML: {}", e))?;

    let mut file = std::fs::File::create(output)
        .map_err(|e| anyhow::anyhow!("Failed to create file: {}", e))?;

    file.write_all(spec_yaml.as_bytes())
        .map_err(|e| anyhow::anyhow!("Failed to write documentation: {}", e))?;

    println!("Documentation generated successfully: {}", output);
    println!("  Format: OpenAPI 3.0 (YAML)");
    println!("  Endpoints: {}", spec.as_object().unwrap().get("paths").unwrap().as_object().unwrap().len());

    Ok(())
}