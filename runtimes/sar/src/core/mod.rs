use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;
use tracing::{info, instrument, warn};
use uuid::Uuid;

use crate::engine::{
    ExecutionStatus, Graph, GraphExecutionInput, GraphExecutionResult, NodeId, ExecutionContext, GraphExecutor, DefaultNodeExecutor,
};
use crate::events::{Event, NatsEventBus};
use crate::model::{ModelRouterConfig, ModelRouter};
use crate::scheduler::{AgentScheduler, SchedulerConfig};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct AgentId(pub Uuid);

impl std::fmt::Display for AgentId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "agent-{}", self.0)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct CellId(pub Uuid);

impl std::fmt::Display for CellId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "cell-{}", self.0)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentConfig {
    pub id: AgentId,
    pub name: String,
    pub graph: Graph,
    pub priority: u8,
    pub max_concurrent_cells: usize,
    pub isolation_enabled: bool,
    pub event_subscriptions: Vec<String>,
}

impl AgentConfig {
    pub fn new(name: String, graph: Graph) -> Self {
        Self {
            id: AgentId(Uuid::new_v4()),
            name,
            graph,
            priority: 2,
            max_concurrent_cells: 100,
            isolation_enabled: true,
            event_subscriptions: vec![],
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AgentStatus {
    Idle,
    Running,
    WaitingForEvent,
    Paused,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentState {
    pub agent_id: AgentId,
    pub status: AgentStatus,
    pub current_cell: Option<CellId>,
    pub memory_snapshot: HashMap<String, serde_json::Value>,
    pub metrics: AgentMetrics,
    pub isolation_enabled: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AgentMetrics {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub average_latency_ms: f64,
    pub total_cost_usd: f64,
    pub last_executed_at: Option<chrono::DateTime<chrono::Utc>>,
}

impl AgentMetrics {
    pub fn success_rate(&self) -> f64 {
        if self.total_executions == 0 {
            return 1.0;
        }
        self.successful_executions as f64 / self.total_executions as f64
    }
}

pub trait AgentCell: Send + Sync {
    fn id(&self) -> CellId;
    fn node_id(&self) -> NodeId;
}

#[derive(Debug, Clone)]
pub struct NodeExecutionError {
    pub node_id: NodeId,
    pub message: String,
    pub retryable: bool,
}

impl NodeExecutionError {
    pub fn new(node_id: NodeId, message: String) -> Self {
        Self { node_id, message, retryable: true }
    }
}

impl std::fmt::Display for NodeExecutionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Node {}: {}", self.node_id, self.message)
    }
}

impl std::error::Error for NodeExecutionError {}

pub struct StatefulAgentRuntime {
    scheduler: Arc<AgentScheduler>,
    event_bus: Arc<NatsEventBus>,
    model_router: Arc<ModelRouter>,
    agent_states: Arc<RwLock<HashMap<AgentId, AgentState>>>,
    shutdown_tx: Option<mpsc::Sender<()>>,
}

impl StatefulAgentRuntime {
    pub async fn new() -> anyhow::Result<Self> {
        Self::with_config(RuntimeConfig::default()).await
    }

    pub async fn with_config(config: RuntimeConfig) -> anyhow::Result<Self> {
        let scheduler = Arc::new(AgentScheduler::new(config.scheduler));

        let event_bus = if let Some(ref nats_url) = config.nats_url {
            match NatsEventBus::new(nats_url.clone()) {
                Ok(bus) => Arc::new(bus),
                Err(e) => {
                    tracing::warn!(error = %e, "Failed to connect to NATS, using default event bus");
                    Arc::new(NatsEventBus::default())
                }
            }
        } else {
            Arc::new(NatsEventBus::default())
        };

        let model_router = Arc::new(ModelRouter::new(config.model_router));
        let (shutdown_tx, _shutdown_rx) = mpsc::channel(1);

        Ok(Self {
            scheduler,
            event_bus,
            model_router,
            agent_states: Arc::new(RwLock::new(HashMap::new())),
            shutdown_tx: Some(shutdown_tx),
        })
    }

    #[instrument(skip_all, fields(agent_id = %agent.id))]
    pub async fn register_agent(&self, agent: AgentConfig) -> anyhow::Result<()> {
        let state = AgentState {
            agent_id: agent.id,
            status: AgentStatus::Idle,
            current_cell: None,
            memory_snapshot: HashMap::new(),
            metrics: AgentMetrics::default(),
            isolation_enabled: agent.isolation_enabled,
        };

        {
            let mut states = self.agent_states.write();
            states.insert(agent.id, state);
        }

        for subject in &agent.event_subscriptions {
            self.event_bus.subscribe(agent.id, subject.clone())?;
        }

        info!(agent_id = %agent.id, name = %agent.name, "Agent registered");
        Ok(())
    }

    pub async fn execute_agent(
        &self,
        agent_id: AgentId,
        input: HashMap<String, serde_json::Value>,
    ) -> anyhow::Result<GraphExecutionResult> {
        let agent_config = {
            let states = self.agent_states.read();
            states.get(&agent_id).cloned()
        };

        let agent_config = agent_config.ok_or_else(|| anyhow::anyhow!("Agent {} not found", agent_id))?;

        if agent_config.isolation_enabled {
            tracing::debug!(agent_id = %agent_id, "Executing agent with isolation ENABLED");
        }

        self.update_agent_status(agent_id, AgentStatus::Running);

        let executor = GraphExecutor::new(DefaultNodeExecutor);
        let isolated = if agent_config.isolation_enabled { Some(agent_id.0.to_string()) } else { None };
        let ctx = Arc::new(ExecutionContext::new(Uuid::new_v4(), None, isolated));

        let graph_input = GraphExecutionInput {
            graph_id: Uuid::new_v4(),
            initial_input: input,
            tenant_id: None,
        };

        let graph = Graph::new(Uuid::new_v4(), "execution".to_string());

        let result = executor.execute(&graph, graph_input, ctx).await;

        if result.status == ExecutionStatus::Completed {
            self.update_agent_status(agent_id, AgentStatus::Idle);
            self.record_success(agent_id, &result);
        } else {
            self.update_agent_status(agent_id, AgentStatus::Failed);
            self.record_failure(agent_id);
        }

        Ok(result)
    }

    pub async fn submit_event(&self, event: Event) -> anyhow::Result<()> {
        self.event_bus.publish(event)?;
        Ok(())
    }

    fn update_agent_status(&self, agent_id: AgentId, status: AgentStatus) {
        let mut states = self.agent_states.write();
        if let Some(state) = states.get_mut(&agent_id) {
            state.status = status;
        }
    }

    fn record_success(&self, agent_id: AgentId, result: &GraphExecutionResult) {
        let mut states = self.agent_states.write();
        if let Some(state) = states.get_mut(&agent_id) {
            state.metrics.total_executions += 1;
            state.metrics.successful_executions += 1;
            if let Some(duration) = result.total_duration_ms {
                let prev = state.metrics.average_latency_ms;
                let n = state.metrics.successful_executions as f64;
                state.metrics.average_latency_ms = prev + (duration as f64 - prev) / n;
            }
            state.metrics.last_executed_at = Some(chrono::Utc::now());
        }
    }

    fn record_failure(&self, agent_id: AgentId) {
        let mut states = self.agent_states.write();
        if let Some(state) = states.get_mut(&agent_id) {
            state.metrics.total_executions += 1;
            state.metrics.failed_executions += 1;
            state.metrics.last_executed_at = Some(chrono::Utc::now());
        }
    }

    pub fn get_agent_state(&self, agent_id: AgentId) -> Option<AgentState> {
        self.agent_states.read().get(&agent_id).cloned()
    }

    pub fn list_agents(&self) -> Vec<AgentId> {
        self.agent_states.read().keys().copied().collect()
    }

    pub async fn shutdown(&mut self) {
        if let Some(tx) = self.shutdown_tx.take() {
            let _ = tx.send(()).await;
        }
        info!("SAR runtime shutting down");
    }

    pub fn save_checkpoint(&self, agent_id: AgentId) -> Option<AgentState> {
        let states = self.agent_states.read();
        states.get(&agent_id).cloned()
    }

    pub fn restore_from_checkpoint(&self, state: AgentState) -> Option<AgentId> {
        let agent_id = state.agent_id;
        let mut states = self.agent_states.write();
        states.insert(agent_id, state);
        info!(agent_id = %agent_id, "Agent state restored from checkpoint");
        Some(agent_id)
    }

    pub async fn graceful_shutdown_agent(&self, agent_id: AgentId, grace_period: Duration) -> bool {
        let start = Instant::now();

        while start.elapsed() < grace_period {
            {
                let states = self.agent_states.read();
                if let Some(state) = states.get(&agent_id) {
                    if state.metrics.total_executions == 0 ||
                       state.status == AgentStatus::Idle {
                        info!(agent_id = %agent_id, "Agent shutdown complete");
                        return true;
                    }
                }
            }
            tokio::time::sleep(Duration::from_secs(1)).await;
        }

        warn!(agent_id = %agent_id, "Agent graceful shutdown timed out");
        false
    }
}

#[derive(Debug, Clone)]
pub struct RuntimeConfig {
    pub nats_url: Option<String>,
    pub memory_enabled: bool,
    pub wasm_enabled: bool,
    pub redis_url: Option<String>,
    pub postgres_url: Option<String>,
    pub vector_db_url: Option<String>,
    pub scheduler: SchedulerConfig,
    pub model_router: ModelRouterConfig,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            nats_url: std::env::var("NATS_URL").ok().or_else(|| Some("nats://localhost:4222".to_string())),
            memory_enabled: true,
            wasm_enabled: true,
            redis_url: std::env::var("REDIS_URL").ok(),
            postgres_url: std::env::var("DATABASE_URL").ok(),
            vector_db_url: std::env::var("VECTOR_DB_URL").ok(),
            scheduler: SchedulerConfig::default(),
            model_router: ModelRouterConfig::default(),
        }
    }
}
